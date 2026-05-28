package ftpsync

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	pathpkg "path"
	"strings"
	"time"
	"unicode/utf8"

	ftp "github.com/jlaffaye/ftp"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

type ftpDialFunc func(addr string, options ...ftp.DialOption) (ftpConnection, error)

type ftpConnection interface {
	Login(user, password string) error
	Quit() error
	Type(transferType ftp.TransferType) error
	Walk(root string) ftpWalker
	Stor(path string, r io.Reader) error
	Retr(path string) (io.ReadCloser, error)
	MakeDir(path string) error
	Delete(path string) error
	RemoveDirRecur(path string) error
	GetEntry(path string) (*ftp.Entry, error)
	IsTimePreciseInList() bool
}

type ftpWalker interface {
	Next() bool
	Err() error
	Path() string
	Stat() *ftp.Entry
}

type ftpClient struct {
	conn            ftpConnection
	pathCodec       *ftpPathCodec
	listTimePrecise bool
}

var newFTPClient = connectFTPClient

func connectFTPClient(ctx context.Context, opts FTPOptions) (*ftpClient, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	if !opts.PassiveMode {
		return nil, fmt.Errorf("ftp: active mode is not supported in v1; set passive mode")
	}
	pathCodec, err := newFTPPathCodec(opts)
	if err != nil {
		return nil, err
	}

	dialOpts := make([]ftp.DialOption, 0, 2)
	if opts.Timeout > 0 {
		dialOpts = append(dialOpts, ftp.DialWithTimeout(opts.Timeout))
	}
	if pathCodec.disableUTF8Feature() {
		dialOpts = append(dialOpts, ftp.DialWithDisabledUTF8(true))
	}

	addr := fmt.Sprintf("%s:%d", opts.Host, opts.Port)
	conn, err := defaultFTPDial(addr, dialOpts...)
	if err != nil {
		return nil, err
	}
	if err := conn.Login(opts.Username, opts.Password); err != nil {
		_ = conn.Quit()
		return nil, err
	}
	if err := conn.Type(ftp.TransferTypeBinary); err != nil {
		_ = conn.Quit()
		return nil, err
	}
	return &ftpClient{conn: conn, pathCodec: pathCodec, listTimePrecise: conn.IsTimePreciseInList()}, nil
}

func defaultFTPDial(addr string, options ...ftp.DialOption) (ftpConnection, error) {
	conn, err := ftp.Dial(addr, options...)
	if err != nil {
		return nil, err
	}
	return ftpServerConn{ServerConn: conn}, nil
}

type ftpServerConn struct {
	*ftp.ServerConn
}

func (c ftpServerConn) Walk(root string) ftpWalker {
	return ftpWalkerAdapter{Walker: c.ServerConn.Walk(root)}
}

func (c ftpServerConn) Retr(path string) (io.ReadCloser, error) {
	return c.ServerConn.Retr(path)
}

type ftpWalkerAdapter struct {
	*ftp.Walker
}

func (c *ftpClient) mkdirAll(remotePath string) error {
	cleanPath, err := c.encodePath(remotePath)
	if err != nil {
		return err
	}
	if cleanPath == "." || cleanPath == "/" {
		return nil
	}
	parts := strings.Split(strings.TrimPrefix(cleanPath, "/"), "/")
	current := ""
	if strings.HasPrefix(cleanPath, "/") {
		current = "/"
	}
	for _, part := range parts {
		if part == "" || part == "." {
			continue
		}
		current = joinFTPPath(current, part)
		err := c.conn.MakeDir(current)
		if err == nil || isFTPAlreadyExists(err) {
			continue
		}
		return err
	}
	return nil
}

func (c *ftpClient) writeFile(remotePath string, localPath string) error {
	encodedPath, err := c.encodePath(remotePath)
	if err != nil {
		return err
	}
	localFile, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer localFile.Close()
	return c.conn.Stor(encodedPath, localFile)
}

func (c *ftpClient) remove(remotePath string) error {
	cleanPath, err := c.encodePath(remotePath)
	if err != nil {
		return err
	}
	entry, err := c.conn.GetEntry(cleanPath)
	if err != nil {
		if isFTPNotExist(err) {
			return nil
		}
		return err
	}
	if entry != nil && entry.Type == ftp.EntryTypeFolder {
		return c.conn.RemoveDirRecur(cleanPath)
	}
	if err := c.conn.Delete(cleanPath); err != nil && !isFTPNotExist(err) {
		return err
	}
	return nil
}

func (c *ftpClient) walk(root string, fn fs.WalkDirFunc) error {
	encodedRoot, err := c.encodePath(root)
	if err != nil {
		return err
	}
	walker := c.conn.Walk(encodedRoot)
	for {
		next := walker.Next()
		if err := walker.Err(); err != nil {
			return err
		}
		if !next {
			return nil
		}
		entry := walker.Stat()
		if entry == nil {
			continue
		}
		decodedPath := c.decodePath(walker.Path())
		decodedEntry := c.decodeEntry(entry)
		info := newFTPFileInfo(decodedEntry, decodedPath, c.listTimePrecise)
		if err := fn(decodedPath, fs.FileInfoToDirEntry(info), nil); err != nil {
			return err
		}
	}
}

func (c *ftpClient) readFile(remotePath string, localPath string) error {
	encodedPath, err := c.encodePath(remotePath)
	if err != nil {
		return err
	}
	response, err := c.conn.Retr(encodedPath)
	if err != nil {
		return err
	}
	defer response.Close()
	if err := os.MkdirAll(filepathDir(localPath), 0o755); err != nil {
		return err
	}
	localFile, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer localFile.Close()
	_, err = io.Copy(localFile, response)
	return err
}

func (c *ftpClient) close() error {
	if c == nil || c.conn == nil {
		return nil
	}
	return c.conn.Quit()
}

func (c *ftpClient) encodePath(value string) (string, error) {
	if c.pathCodec == nil {
		return cleanFTPPath(value), nil
	}
	return c.pathCodec.encodePath(value)
}

func (c *ftpClient) decodePath(value string) string {
	if c.pathCodec == nil {
		return cleanFTPPath(value)
	}
	return c.pathCodec.decodePath(value)
}

func (c *ftpClient) decodeEntry(entry *ftp.Entry) *ftp.Entry {
	if c.pathCodec == nil {
		return entry
	}
	return c.pathCodec.decodeEntry(entry)
}

type ftpPathCodec struct {
	mode string
}

const (
	ftpEncodingAuto = "auto"
	ftpEncodingUTF8 = "utf8"
	ftpEncodingGBK  = "gbk"
)

func newFTPPathCodec(opts FTPOptions) (*ftpPathCodec, error) {
	mode := strings.ToLower(strings.TrimSpace(opts.PathEncoding))
	if mode == "" || mode == "utf-8" {
		mode = ftpEncodingUTF8
	}
	if mode == "gb2312" {
		mode = ftpEncodingGBK
	}
	switch mode {
	case ftpEncodingAuto, ftpEncodingUTF8, ftpEncodingGBK:
		return &ftpPathCodec{mode: mode}, nil
	default:
		return nil, fmt.Errorf("ftp: unsupported encoding %q", opts.PathEncoding)
	}
}

func (c *ftpPathCodec) disableUTF8Feature() bool {
	return c.mode == ftpEncodingGBK
}

func (c *ftpPathCodec) encodePath(value string) (string, error) {
	clean := cleanFTPPath(value)
	if c.mode != ftpEncodingGBK {
		return clean, nil
	}
	encoded, _, err := transform.String(simplifiedchinese.GBK.NewEncoder(), clean)
	return encoded, err
}

func (c *ftpPathCodec) decodePath(value string) string {
	if value == "" {
		return value
	}
	decoded, ok := c.decodeString(value)
	if !ok {
		return value
	}
	return cleanFTPPath(decoded)
}

func (c *ftpPathCodec) decodeEntry(entry *ftp.Entry) *ftp.Entry {
	if entry == nil {
		return nil
	}
	decoded := *entry
	decoded.Name = c.decodeName(entry.Name)
	decoded.Target = c.decodeName(entry.Target)
	return &decoded
}

func (c *ftpPathCodec) decodeName(value string) string {
	decoded, ok := c.decodeString(value)
	if !ok {
		return value
	}
	return decoded
}

func (c *ftpPathCodec) decodeString(value string) (string, bool) {
	if value == "" {
		return value, true
	}
	switch c.mode {
	case ftpEncodingUTF8:
		return value, true
	case ftpEncodingGBK:
		decoded, _, err := transform.String(simplifiedchinese.GBK.NewDecoder(), value)
		return decoded, err == nil
	case ftpEncodingAuto:
		if utf8.ValidString(value) {
			return value, true
		}
		decoded, _, err := transform.String(simplifiedchinese.GBK.NewDecoder(), value)
		return decoded, err == nil
	default:
		return value, true
	}
}

type ftpFileInfo struct {
	entry           *ftp.Entry
	name            string
	listTimePrecise bool
}

func newFTPFileInfo(entry *ftp.Entry, name string, listTimePrecise bool) fs.FileInfo {
	return ftpFileInfo{entry: entry, name: name, listTimePrecise: listTimePrecise}
}

func (i ftpFileInfo) Name() string {
	if i.entry != nil && i.entry.Name != "" {
		return i.entry.Name
	}
	return pathpkg.Base(i.name)
}

func (i ftpFileInfo) Size() int64 {
	if i.entry == nil {
		return 0
	}
	return int64(i.entry.Size)
}

func (i ftpFileInfo) Mode() fs.FileMode {
	if i.entry == nil {
		return 0
	}
	switch i.entry.Type {
	case ftp.EntryTypeFolder:
		return fs.ModeDir | 0o755
	case ftp.EntryTypeLink:
		return fs.ModeSymlink | 0o777
	default:
		return 0o644
	}
}

func (i ftpFileInfo) ModTime() time.Time {
	if i.entry == nil {
		return time.Time{}
	}
	if i.listTimePrecise {
		return i.entry.Time
	}
	return i.entry.Time.Truncate(time.Second)
}

func (i ftpFileInfo) IsDir() bool { return i.Mode().IsDir() }
func (i ftpFileInfo) Sys() any    { return i.entry }

func cleanFTPPath(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "/"
	}
	clean := pathpkg.Clean(strings.ReplaceAll(trimmed, `\`, "/"))
	if clean == "." {
		return "/"
	}
	return clean
}

func joinFTPPath(base string, elem string) string {
	cleanBase := cleanFTPPath(base)
	cleanElem := strings.ReplaceAll(strings.TrimSpace(elem), `\`, "/")
	if cleanElem == "" || cleanElem == "." {
		return cleanBase
	}
	if cleanBase == "/" {
		return cleanFTPPath("/" + strings.TrimPrefix(cleanElem, "/"))
	}
	return cleanFTPPath(pathpkg.Join(cleanBase, cleanElem))
}

func isFTPAlreadyExists(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "file exists") || strings.Contains(message, "directory already exists") || strings.Contains(message, "550")
}

func isFTPNotExist(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "not exist") || strings.Contains(message, "not found") || strings.Contains(message, "550")
}

func filepathDir(value string) string {
	idx := strings.LastIndexAny(value, `/\\`)
	if idx < 0 {
		return "."
	}
	if idx == 0 {
		return value[:1]
	}
	return value[:idx]
}
