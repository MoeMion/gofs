package ftp

import (
	"errors"
	"io"
	"io/fs"
	"net/http"
	"path"
	"strings"

	ftp "github.com/jlaffaye/ftp"
)

type ftpFile struct {
	resp            *ftp.Response
	client          ftpConn
	name            string
	listTimePrecise bool
	closed          bool
}

func newFTPFile(resp *ftp.Response, client ftpConn, name string, listTimePrecise bool) http.File {
	return &ftpFile{
		resp:            resp,
		client:          client,
		name:            name,
		listTimePrecise: listTimePrecise,
	}
}

func (f *ftpFile) Close() error {
	f.closed = true
	err := f.resp.Close()
	if isFTPSuccessReply(err) {
		return nil
	}
	return err
}

func (f *ftpFile) Read(p []byte) (n int, err error) {
	n, err = f.resp.Read(p)
	if err != nil && isFTPSuccessReply(err) {
		if n > 0 {
			return n, nil
		}
		return 0, io.EOF
	}
	return n, err
}

func (f *ftpFile) Seek(offset int64, whence int) (int64, error) {
	return 0, fs.ErrInvalid
}

func (f *ftpFile) Readdir(count int) ([]fs.FileInfo, error) {
	entries, err := f.client.List(f.name)
	if err != nil {
		return nil, err
	}
	fis := make([]fs.FileInfo, 0, len(entries))
	for _, entry := range entries {
		fis = append(fis, newFTPFileInfo(entry, path.Join(f.name, entry.Name), f.listTimePrecise))
	}
	if count > 0 && len(fis) > count {
		fis = fis[:count]
	}
	return fis, nil
}

func (f *ftpFile) Stat() (fs.FileInfo, error) {
	entry, err := f.client.GetEntry(f.name)
	if err != nil {
		return nil, err
	}
	return newFTPFileInfo(entry, f.name, f.listTimePrecise), nil
}

type ftpDirFile struct {
	io.Reader
	io.Seeker

	client          ftpConn
	name            string
	listTimePrecise bool
}

func newFTPDirFile(client ftpConn, name string, listTimePrecise bool) http.File {
	return &ftpDirFile{
		client:          client,
		name:            name,
		listTimePrecise: listTimePrecise,
	}
}

func (f *ftpDirFile) Close() error {
	return nil
}

func (f *ftpDirFile) Readdir(count int) ([]fs.FileInfo, error) {
	entries, err := f.client.List(f.name)
	if err != nil {
		return nil, err
	}
	fis := make([]fs.FileInfo, 0, len(entries))
	for _, entry := range entries {
		fis = append(fis, newFTPFileInfo(entry, path.Join(f.name, entry.Name), f.listTimePrecise))
	}
	if count > 0 && len(fis) > count {
		fis = fis[:count]
	}
	return fis, nil
}

func (f *ftpDirFile) Stat() (fs.FileInfo, error) {
	entry, err := f.client.GetEntry(f.name)
	if err != nil {
		return nil, err
	}
	return newFTPFileInfo(entry, f.name, f.listTimePrecise), nil
}

func isFTPSuccessReply(err error) bool {
	if err == nil || errors.Is(err, io.EOF) {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "226 transfer complete") || strings.Contains(message, "250 listing") || strings.Contains(message, "end mlsd") || strings.Contains(message, "end mlst")
}
