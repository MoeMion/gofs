package ftp

import (
	"io/fs"
	pathpkg "path"
	"time"

	ftp "github.com/jlaffaye/ftp"
)

type ftpFileInfo struct {
	entry           *ftp.Entry
	fullPath        string
	listTimePrecise bool
}

func newFTPFileInfo(entry *ftp.Entry, fullPath string, listTimePrecise bool) fs.FileInfo {
	return &ftpFileInfo{entry: entry, fullPath: fullPath, listTimePrecise: listTimePrecise}
}

func (fi *ftpFileInfo) Name() string {
	if fi.entry != nil && fi.entry.Name != "" {
		return fi.entry.Name
	}
	return pathpkg.Base(fi.fullPath)
}

func (fi *ftpFileInfo) Size() int64 {
	if fi.entry == nil {
		return 0
	}
	return int64(fi.entry.Size)
}

func (fi *ftpFileInfo) Mode() fs.FileMode {
	if fi.IsDir() {
		return fs.ModeDir | 0o755
	}
	if fi.entry != nil && fi.entry.Type == ftp.EntryTypeLink {
		return fs.ModeSymlink | 0o666
	}
	return 0o666
}

func (fi *ftpFileInfo) ModTime() time.Time {
	if fi.entry == nil {
		return time.Time{}
	}
	if fi.listTimePrecise {
		return fi.entry.Time
	}
	return fi.entry.Time.Truncate(time.Second)
}

func (fi *ftpFileInfo) IsDir() bool {
	return fi.entry != nil && fi.entry.Type == ftp.EntryTypeFolder
}

func (fi *ftpFileInfo) Sys() any {
	return fi.entry
}
