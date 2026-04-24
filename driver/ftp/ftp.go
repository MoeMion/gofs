package ftp

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	pathpkg "path"
	"strings"
	"sync"
	"syscall"
	"time"

	ftp "github.com/jlaffaye/ftp"
	"github.com/no-src/gofs/core"
	"github.com/no-src/gofs/driver"
	"github.com/no-src/gofs/internal/rate"
	"github.com/no-src/gofs/logger"
	"github.com/no-src/gofs/retry"
)

var (
	errFTPActiveModeUnsupported = errors.New("ftp: active mode is not supported in v1; set ftp_passive=true")
	errFTPSymlinkUnsupported    = errors.New("ftp: symlink is not supported")
	errFTPReadLinkUnsupported   = errors.New("ftp: readlink is not supported")
	errFTPChtimesUnsupported    = errors.New("ftp: server does not support setting file times")
	errFTPOffline               = errors.New("ftp offline")
)

type dialFn func(addr string, options ...ftp.DialOption) (ftpConn, error)

type ftpConn interface {
	Login(user, password string) error
	NoOp() error
	Quit() error
	Type(transferType ftp.TransferType) error
	Walk(root string) ftpWalker
	CurrentDir() (string, error)
	ChangeDir(path string) error
	List(path string) ([]*ftp.Entry, error)
	NameList(path string) ([]string, error)
	Retr(path string) (*ftp.Response, error)
	Stor(path string, r io.Reader) error
	Delete(path string) error
	MakeDir(path string) error
	RemoveDirRecur(path string) error
	Rename(from, to string) error
	GetEntry(path string) (*ftp.Entry, error)
	GetTime(path string) (time.Time, error)
	SetTime(path string, t time.Time) error
	IsGetTimeSupported() bool
	IsSetTimeSupported() bool
	IsTimePreciseInList() bool
}

type ftpWalker interface {
	Next() bool
	Err() error
	Path() string
	Stat() *ftp.Entry
}

type serverConn struct {
	*ftp.ServerConn
}

func (c *serverConn) Walk(root string) ftpWalker {
	return &ftpWalkerAdapter{Walker: c.ServerConn.Walk(root)}
}

type ftpWalkerAdapter struct {
	*ftp.Walker
}

func (w *ftpWalkerAdapter) Next() bool {
	return w.Walker.Next()
}

func (w *ftpWalkerAdapter) Err() error {
	return w.Walker.Err()
}

func (w *ftpWalkerAdapter) Path() string {
	return w.Walker.Path()
}

func (w *ftpWalkerAdapter) Stat() *ftp.Entry {
	return w.Walker.Stat()
}

type ftpDriver struct {
	client          ftpConn
	driverName      string
	remoteAddr      string
	ftpConfig       core.FTPConfig
	r               retry.Retry
	mu              sync.Mutex
	online          bool
	autoReconnect   bool
	maxTranRate     int64
	logger          *logger.Logger
	dial            dialFn
	listTimePrecise bool
}

// NewFTPDriver get a ftp driver.
func NewFTPDriver(remoteAddr string, ftpConfig core.FTPConfig, autoReconnect bool, r retry.Retry, maxTranRate int64, logger *logger.Logger) driver.Driver {
	return newFTPDriver(remoteAddr, ftpConfig, autoReconnect, r, maxTranRate, logger, defaultDial)
}

func newFTPDriver(remoteAddr string, ftpConfig core.FTPConfig, autoReconnect bool, r retry.Retry, maxTranRate int64, logger *logger.Logger, dial dialFn) *ftpDriver {
	if logger == nil {
		logger = loggerPkg()
	}
	if dial == nil {
		dial = defaultDial
	}
	return &ftpDriver{
		driverName:    "ftp",
		remoteAddr:    remoteAddr,
		ftpConfig:     ftpConfig,
		r:             r,
		autoReconnect: autoReconnect,
		maxTranRate:   maxTranRate,
		logger:        logger,
		dial:          dial,
	}
}

func loggerPkg() *logger.Logger {
	return logger.InnerLogger()
}

func defaultDial(addr string, options ...ftp.DialOption) (ftpConn, error) {
	conn, err := ftp.Dial(addr, options...)
	if err != nil {
		return nil, err
	}
	return &serverConn{ServerConn: conn}, nil
}

func (d *ftpDriver) DriverName() string {
	return d.driverName
}

func (d *ftpDriver) Connect() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.connectLocked()
}

func (d *ftpDriver) connectLocked() error {
	if d.online && d.client != nil {
		return nil
	}
	if strings.TrimSpace(d.ftpConfig.Username) == "" {
		return errors.New("ftp: the username is required")
	}
	if !d.ftpConfig.PassiveMode {
		return errFTPActiveModeUnsupported
	}

	options := make([]ftp.DialOption, 0, 1)
	if strings.TrimSpace(d.ftpConfig.Timeout) != "" {
		timeout, err := time.ParseDuration(d.ftpConfig.Timeout)
		if err != nil {
			return fmt.Errorf("ftp: invalid timeout %q: %w", d.ftpConfig.Timeout, err)
		}
		options = append(options, ftp.DialWithTimeout(timeout))
	}

	client, err := d.dial(d.remoteAddr, options...)
	if err != nil {
		return err
	}
	if err := client.Login(d.ftpConfig.Username, d.ftpConfig.Password); err != nil {
		_ = client.Quit()
		return err
	}
	if err := client.Type(ftp.TransferTypeBinary); err != nil {
		_ = client.Quit()
		return err
	}

	d.client = client
	d.online = true
	d.listTimePrecise = client.IsTimePreciseInList()
	d.logger.Debug("connect to ftp server success => %s", d.remoteAddr)
	return nil
}

func (d *ftpDriver) reconnectLocked() error {
	d.logger.Debug("reconnect to ftp server => %s", d.remoteAddr)
	if d.r == nil {
		return d.connectLocked()
	}
	return d.r.Do(func() error {
		return d.connectLocked()
	}, "ftp reconnect").Wait()
}

func (d *ftpDriver) reconnectIfLost(f func(client ftpConn) error) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.online || d.client == nil {
		return errFTPOffline
	}

	err := f(d.client)
	if !d.autoReconnect {
		return err
	}
	if !d.isTransportLost(err) {
		return err
	}

	d.logger.Error(err, "connect to ftp server failed")
	d.closeClientLocked()
	if reconnectErr := d.reconnectLocked(); reconnectErr != nil {
		return fmt.Errorf("ftp: reconnect failed after transport loss: %w", reconnectErr)
	}
	if err = f(d.client); err != nil {
		return fmt.Errorf("ftp: operation failed after reconnect: %w", err)
	}
	return nil
}

func (d *ftpDriver) closeClientLocked() {
	if d.client != nil {
		_ = d.client.Quit()
	}
	d.client = nil
	d.online = false
}

func (d *ftpDriver) isTransportLost(err error) bool {
	if err == nil {
		if d.client == nil {
			return false
		}
		return d.client.NoOp() != nil
	}
	if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) || errors.Is(err, syscall.EPIPE) || errors.Is(err, syscall.ECONNRESET) {
		return true
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "connection reset") || strings.Contains(message, "broken pipe") || strings.Contains(message, "use of closed network connection") || strings.Contains(message, "not connected") || strings.Contains(message, "connection lost") || strings.Contains(message, "control connection")
}

func (d *ftpDriver) MkdirAll(path string) error {
	cleanPath := cleanFTPPath(path)
	if cleanPath == "." || cleanPath == "/" {
		return nil
	}
	return d.reconnectIfLost(func(client ftpConn) error {
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
			err := client.MakeDir(current)
			if err == nil || isFTPAlreadyExists(err) {
				continue
			}
			return err
		}
		return nil
	})
}

func (d *ftpDriver) Create(path string) error {
	return d.reconnectIfLost(func(client ftpConn) error {
		return client.Stor(cleanFTPPath(path), strings.NewReader(""))
	})
}

func (d *ftpDriver) Symlink(oldname, newname string) error {
	return errFTPSymlinkUnsupported
}

func (d *ftpDriver) Remove(path string) error {
	cleanPath := cleanFTPPath(path)
	return d.reconnectIfLost(func(client ftpConn) error {
		entry, err := client.GetEntry(cleanPath)
		if err != nil {
			if isFTPNotExist(err) {
				return nil
			}
			return err
		}
		if entry != nil && entry.Type == ftp.EntryTypeFolder {
			return client.RemoveDirRecur(cleanPath)
		}
		if err := client.Delete(cleanPath); err != nil && !isFTPNotExist(err) {
			return err
		}
		return nil
	})
}

func (d *ftpDriver) Rename(oldPath, newPath string) error {
	return d.reconnectIfLost(func(client ftpConn) error {
		return client.Rename(cleanFTPPath(oldPath), cleanFTPPath(newPath))
	})
}

func (d *ftpDriver) Chtimes(path string, aTime time.Time, mTime time.Time) error {
	return d.reconnectIfLost(func(client ftpConn) error {
		if !client.IsSetTimeSupported() {
			return errFTPChtimesUnsupported
		}
		if err := client.SetTime(cleanFTPPath(path), mTime); err != nil {
			if isFTPUnsupported(err) {
				return errFTPChtimesUnsupported
			}
			return err
		}
		return nil
	})
}

func (d *ftpDriver) WalkDir(root string, fn fs.WalkDirFunc) error {
	return d.reconnectIfLost(func(client ftpConn) error {
		walker := client.Walk(cleanFTPPath(root))
		for {
			next := walker.Next()
			err := walker.Err()
			if err != nil {
				return err
			}
			if !next {
				return nil
			}
			entry := walker.Stat()
			if entry == nil {
				continue
			}
			fi := newFTPFileInfo(entry, walker.Path(), d.listTimePrecise)
			if err := fn(walker.Path(), fs.FileInfoToDirEntry(fi), nil); err != nil {
				return err
			}
		}
	})
}

func (d *ftpDriver) Open(path string) (http.File, error) {
	var file http.File
	err := d.reconnectIfLost(func(client ftpConn) error {
		entry, err := client.GetEntry(cleanFTPPath(path))
		if err != nil {
			return err
		}
		if entry != nil && entry.Type == ftp.EntryTypeFolder {
			file = newFTPDirFile(client, cleanFTPPath(path), d.listTimePrecise)
			return nil
		}
		resp, err := client.Retr(cleanFTPPath(path))
		if err != nil {
			return err
		}
		file = rate.NewFile(newFTPFile(resp, client, cleanFTPPath(path), d.listTimePrecise), d.maxTranRate, d.logger)
		return nil
	})
	return file, err
}

func (d *ftpDriver) Stat(path string) (fs.FileInfo, error) {
	return d.stat(path)
}

func (d *ftpDriver) Lstat(path string) (fs.FileInfo, error) {
	return d.stat(path)
}

func (d *ftpDriver) stat(path string) (fs.FileInfo, error) {
	var fi fs.FileInfo
	err := d.reconnectIfLost(func(client ftpConn) error {
		entry, err := client.GetEntry(cleanFTPPath(path))
		if err != nil {
			return err
		}
		fi = newFTPFileInfo(entry, cleanFTPPath(path), d.listTimePrecise)
		return nil
	})
	return fi, err
}

func (d *ftpDriver) GetFileTime(path string) (cTime time.Time, aTime time.Time, mTime time.Time, err error) {
	err = d.reconnectIfLost(func(client ftpConn) error {
		entry, entryErr := client.GetEntry(cleanFTPPath(path))
		if entryErr != nil {
			return entryErr
		}
		mtime := time.Time{}
		if client.IsGetTimeSupported() {
			mtime, entryErr = client.GetTime(cleanFTPPath(path))
			if entryErr != nil && !isFTPUnsupported(entryErr) {
				return entryErr
			}
		}
		if mtime.IsZero() && entry != nil {
			mtime = entry.Time
			if !d.listTimePrecise {
				mtime = mtime.Truncate(time.Second)
			}
		}
		cTime = mtime
		aTime = mtime
		mTime = mtime
		return nil
	})
	return
}

func (d *ftpDriver) Write(src string, dest string) error {
	return d.reconnectIfLost(func(client ftpConn) error {
		srcFile, err := os.Open(src)
		if err != nil {
			return err
		}
		defer srcFile.Close()
		return client.Stor(cleanFTPPath(dest), rate.NewReader(srcFile, d.maxTranRate, d.logger))
	})
}

func (d *ftpDriver) ReadLink(path string) (string, error) {
	return "", errFTPReadLinkUnsupported
}

func cleanFTPPath(path string) string {
	if strings.TrimSpace(path) == "" {
		return "/"
	}
	clean := pathpkg.Clean(path)
	if clean == "." {
		return "/"
	}
	return clean
}

func joinFTPPath(base string, elem string) string {
	if base == "" || base == "/" {
		return "/" + strings.TrimPrefix(elem, "/")
	}
	return pathpkg.Join(base, elem)
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

func isFTPUnsupported(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "not implemented") || strings.Contains(message, "not supported") || strings.Contains(message, "502") || strings.Contains(message, "504")
}
