package sync

import (
	"errors"
	"io/fs"
	"net/http"
	"testing"
	"time"

	"github.com/no-src/gofs/core"
	"github.com/no-src/gofs/driver"
	"github.com/no-src/gofs/logger"
	"github.com/no-src/gofs/retry"
	"github.com/no-src/nsgo/hashutil"
)

type fakeFTPDriver struct{}

func (d *fakeFTPDriver) DriverName() string                                          { return "ftp" }
func (d *fakeFTPDriver) Connect() error                                              { return nil }
func (d *fakeFTPDriver) MkdirAll(path string) error                                  { return nil }
func (d *fakeFTPDriver) Create(path string) error                                    { return nil }
func (d *fakeFTPDriver) Symlink(oldname, newname string) error                       { return nil }
func (d *fakeFTPDriver) Remove(path string) error                                    { return nil }
func (d *fakeFTPDriver) Rename(oldPath, newPath string) error                        { return nil }
func (d *fakeFTPDriver) Chtimes(path string, aTime time.Time, mTime time.Time) error { return nil }
func (d *fakeFTPDriver) WalkDir(root string, fn fs.WalkDirFunc) error                { return nil }
func (d *fakeFTPDriver) Open(path string) (http.File, error)                         { return nil, nil }
func (d *fakeFTPDriver) Stat(path string) (fs.FileInfo, error)                       { return nil, nil }
func (d *fakeFTPDriver) Lstat(path string) (fs.FileInfo, error)                      { return nil, nil }
func (d *fakeFTPDriver) GetFileTime(path string) (time.Time, time.Time, time.Time, error) {
	return time.Time{}, time.Time{}, time.Time{}, nil
}
func (d *fakeFTPDriver) Write(src string, dest string) error  { return nil }
func (d *fakeFTPDriver) ReadLink(path string) (string, error) { return "", nil }

func TestNewSync_RoutesDiskToFTPToFTPPushClientSync(t *testing.T) {
	sourceDir := t.TempDir()
	destDir := t.TempDir()

	originalFactory := newFTPPushDriver
	t.Cleanup(func() {
		newFTPPushDriver = originalFactory
	})

	newFTPPushDriver = func(remoteAddr string, ftpConfig core.FTPConfig, autoReconnect bool, r retry.Retry, maxTranRate int64, logger *logger.Logger) driver.Driver {
		return &fakeFTPDriver{}
	}

	opt := Option{
		Source:            core.NewDiskVFS(sourceDir),
		Dest:              core.NewVFS("ftp://127.0.0.1:21?path=" + destDir + "&remote_path=/remote/dest&ftp_user=user&ftp_pass=pass&ftp_passive=true"),
		ChunkSize:         1,
		ChecksumAlgorithm: hashutil.DefaultHash,
	}

	s, err := NewSync(opt)
	if err != nil {
		t.Fatalf("expected FTP push route without deferred error, got %v", err)
	}

	ftpSync, ok := s.(*ftpPushClientSync)
	if !ok {
		t.Fatalf("expected *ftpPushClientSync, got %T", s)
	}

	if _, ok := ftpSync.driver.(*fakeFTPDriver); !ok {
		t.Fatalf("expected fake FTP driver wiring, got %T", ftpSync.driver)
	}
	if ftpSync.basePath != "/remote/dest" {
		t.Fatalf("expected FTP base path /remote/dest, got %q", ftpSync.basePath)
	}
}

func TestNewSync_RoutesFTPToDiskToFTPPullClientSync(t *testing.T) {
	destDir := t.TempDir()

	originalFactory := newFTPPullDriver
	t.Cleanup(func() {
		newFTPPullDriver = originalFactory
	})

	newFTPPullDriver = func(remoteAddr string, ftpConfig core.FTPConfig, autoReconnect bool, r retry.Retry, maxTranRate int64, logger *logger.Logger) driver.Driver {
		return &fakeFTPDriver{}
	}

	opt := Option{
		Source:            core.NewVFS("ftp://127.0.0.1:21?path=./dest&remote_path=/remote/source&ftp_user=user&ftp_pass=pass&ftp_passive=true"),
		Dest:              core.NewDiskVFS(destDir),
		ChunkSize:         1,
		ChecksumAlgorithm: hashutil.DefaultHash,
	}

	s, err := NewSync(opt)
	if err != nil {
		t.Fatalf("expected FTP pull route without deferred error, got %v", err)
	}

	ftpSync, ok := s.(*ftpPullClientSync)
	if !ok {
		t.Fatalf("expected *ftpPullClientSync, got %T", s)
	}

	if _, ok := ftpSync.driver.(*fakeFTPDriver); !ok {
		t.Fatalf("expected fake FTP driver wiring, got %T", ftpSync.driver)
	}
	if ftpSync.diskSync.sourceAbsPath != "/remote/source" {
		t.Fatalf("expected sourceAbsPath reset to /remote/source, got %q", ftpSync.diskSync.sourceAbsPath)
	}
	if ftpSync.diskSync.statFn == nil {
		t.Fatal("expected FTP pull statFn to be wired")
	}
	if ftpSync.diskSync.getFileTimeFn == nil {
		t.Fatal("expected FTP pull getFileTimeFn to be wired")
	}
}

func TestNewSync_UnsupportedNonFTPFallsBackToUnsupportedError(t *testing.T) {
	opt := Option{
		Source: core.NewVFS("minio://127.0.0.1:9000?path=./dest&remote_path=/bucket/source&secure=false"),
		Dest:   core.NewVFS("ftp://127.0.0.1:21?path=./source&remote_path=/remote/dest&ftp_user=user&ftp_pass=pass&ftp_passive=true"),
	}

	_, err := NewSync(opt)
	if !errors.Is(err, errFileSystemUnsupported) {
		t.Fatalf("expected unsupported error, got %v", err)
	}
}
