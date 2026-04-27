package sync

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/no-src/gofs/contract"
	"github.com/no-src/gofs/core"
	"github.com/no-src/gofs/driver"
	"github.com/no-src/gofs/ignore"
	"github.com/no-src/gofs/logger"
	"github.com/no-src/gofs/retry"
	"github.com/no-src/nsgo/hashutil"
)

type fakeFTPDriver struct {
	createCalls   []string
	removeCalls   []string
	renameCalls   [][2]string
	writeCalls    [][2]string
	connectCount  int
	walkDirFn     func(root string, fn fs.WalkDirFunc) error
	openFn        func(path string) (http.File, error)
	statFn        func(path string) (fs.FileInfo, error)
	getFileTimeFn func(path string) (time.Time, time.Time, time.Time, error)
	readLinkFn    func(path string) (string, error)
}

func (d *fakeFTPDriver) DriverName() string { return "ftp" }
func (d *fakeFTPDriver) Connect() error {
	d.connectCount++
	return nil
}
func (d *fakeFTPDriver) MkdirAll(path string) error {
	d.createCalls = append(d.createCalls, path)
	return nil
}
func (d *fakeFTPDriver) Create(path string) error {
	d.createCalls = append(d.createCalls, path)
	return nil
}
func (d *fakeFTPDriver) Symlink(oldname, newname string) error { return nil }
func (d *fakeFTPDriver) Remove(path string) error {
	d.removeCalls = append(d.removeCalls, path)
	return nil
}
func (d *fakeFTPDriver) Rename(oldPath, newPath string) error {
	d.renameCalls = append(d.renameCalls, [2]string{oldPath, newPath})
	return nil
}
func (d *fakeFTPDriver) Chtimes(path string, aTime time.Time, mTime time.Time) error { return nil }
func (d *fakeFTPDriver) WalkDir(root string, fn fs.WalkDirFunc) error {
	if d.walkDirFn != nil {
		return d.walkDirFn(root, fn)
	}
	return nil
}
func (d *fakeFTPDriver) Open(path string) (http.File, error) {
	if d.openFn != nil {
		return d.openFn(path)
	}
	return nil, fmt.Errorf("unexpected Open call: %s", path)
}
func (d *fakeFTPDriver) Stat(path string) (fs.FileInfo, error) {
	if d.statFn != nil {
		return d.statFn(path)
	}
	return nil, fmt.Errorf("unexpected Stat call: %s", path)
}
func (d *fakeFTPDriver) Lstat(path string) (fs.FileInfo, error) {
	return d.Stat(path)
}
func (d *fakeFTPDriver) GetFileTime(path string) (time.Time, time.Time, time.Time, error) {
	if d.getFileTimeFn != nil {
		return d.getFileTimeFn(path)
	}
	return time.Time{}, time.Time{}, time.Time{}, fmt.Errorf("unexpected GetFileTime call: %s", path)
}
func (d *fakeFTPDriver) Write(src string, dest string) error {
	d.writeCalls = append(d.writeCalls, [2]string{src, dest})
	return nil
}
func (d *fakeFTPDriver) ReadLink(path string) (string, error) {
	if d.readLinkFn != nil {
		return d.readLinkFn(path)
	}
	return "", nil
}

type fakeHTTPFile struct {
	*bytes.Reader
	stat fs.FileInfo
}

func newFakeHTTPFile(content string, stat fs.FileInfo) *fakeHTTPFile {
	return &fakeHTTPFile{
		Reader: bytes.NewReader([]byte(content)),
		stat:   stat,
	}
}

func (f *fakeHTTPFile) Close() error               { return nil }
func (f *fakeHTTPFile) Stat() (fs.FileInfo, error) { return f.stat, nil }
func (f *fakeHTTPFile) Readdir(count int) ([]fs.FileInfo, error) {
	return nil, nil
}

type fakeFileInfo struct {
	name    string
	size    int64
	mode    fs.FileMode
	modTime time.Time
	isDir   bool
}

func (fi fakeFileInfo) Name() string       { return fi.name }
func (fi fakeFileInfo) Size() int64        { return fi.size }
func (fi fakeFileInfo) Mode() fs.FileMode  { return fi.mode }
func (fi fakeFileInfo) ModTime() time.Time { return fi.modTime }
func (fi fakeFileInfo) IsDir() bool        { return fi.isDir }
func (fi fakeFileInfo) Sys() any           { return nil }

type fakePathIgnore struct{}

func (fakePathIgnore) MatchPath(path, caller, desc string) bool { return false }

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

func TestFTPPushClientSync_SkipsImplicitCWDLocalMirrorWhenPathOmitted(t *testing.T) {
	sourceDir := t.TempDir()
	workDir := t.TempDir()
	remoteBase := "/remote/dest"
	ftpDriver := &fakeFTPDriver{}

	originalWorkDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	if err := os.Chdir(workDir); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(originalWorkDir); err != nil {
			t.Errorf("restore Chdir() error = %v", err)
		}
	})

	sourceFile := filepath.Join(sourceDir, "leaked.txt")
	if err := os.WriteFile(sourceFile, []byte("content"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	opt := testSyncOption(t, core.NewDiskVFS(sourceDir), core.NewVFS("ftp://127.0.0.1:21?remote_path="+remoteBase+"&ftp_user=user&ftp_pass=pass&ftp_passive=true"))
	syncer, err := newDiskSync(opt)
	if err != nil {
		t.Fatalf("newDiskSync() error = %v", err)
	}

	s := &ftpPushClientSync{
		driverPushClientSync: newDriverPushClientSync(*syncer, remoteBase),
	}
	s.driver = ftpDriver

	if err := s.Create(sourceFile); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(workDir, "leaked.txt")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected omitted FTP path not to create a local CWD mirror, stat err=%v", err)
	}
	if len(ftpDriver.createCalls) != 1 || ftpDriver.createCalls[0] != "/remote/dest/leaked.txt" {
		t.Fatalf("expected remote create for /remote/dest/leaked.txt, got %#v", ftpDriver.createCalls)
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

func TestFTPPushClientSync_PreservesDriverBackedDeleteAndRenameSemantics(t *testing.T) {
	sourceDir := t.TempDir()
	remoteBase := "/remote/dest"
	ftpDriver := &fakeFTPDriver{}

	opt := testSyncOption(t, core.NewDiskVFS(sourceDir), core.NewVFS("ftp://127.0.0.1:21?path="+sourceDir+"&remote_path="+remoteBase+"&ftp_user=user&ftp_pass=pass&ftp_passive=true"))
	opt.SyncOnce = true

	syncer, err := newDiskSync(opt)
	if err != nil {
		t.Fatalf("newDiskSync() error = %v", err)
	}

	s := &ftpPushClientSync{
		driverPushClientSync: newDriverPushClientSync(*syncer, remoteBase),
	}
	s.driver = ftpDriver

	s.enableLogicallyDelete = false

	removedPath := filepath.Join(sourceDir, "nested", "file.txt")
	if err := s.Remove(removedPath); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}
	if len(ftpDriver.removeCalls) != 1 || ftpDriver.removeCalls[0] != "/remote/dest/nested/file.txt" {
		t.Fatalf("expected remote delete for /remote/dest/nested/file.txt, got %#v", ftpDriver.removeCalls)
	}

	renamedPath := filepath.Join(sourceDir, "renamed.txt")
	if err := os.WriteFile(renamedPath, []byte("rename target"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := s.Rename(renamedPath); err != nil {
		t.Fatalf("Rename() error = %v", err)
	}
	if len(ftpDriver.removeCalls) != 2 || ftpDriver.removeCalls[1] != "/remote/dest/renamed.txt" {
		t.Fatalf("expected rename to force remote delete for /remote/dest/renamed.txt, got %#v", ftpDriver.removeCalls)
	}
}

func TestFTPPullClientSync_PreservesDriverBackedRemoveAndRenameSemantics(t *testing.T) {
	destDir := t.TempDir()

	opt := testSyncOption(t, core.NewVFS("ftp://127.0.0.1:21?path=./dest&remote_path=/remote/source&ftp_user=user&ftp_pass=pass&ftp_passive=true"), core.NewDiskVFS(destDir))
	syncer, err := newDiskSync(opt)
	if err != nil {
		t.Fatalf("newDiskSync() error = %v", err)
	}

	s := &ftpPullClientSync{
		driverPullClientSync: newDriverPullClientSync(*syncer),
	}
	s.diskSync.sourceAbsPath = "/remote/source"

	obsolete := filepath.Join(destDir, "nested", "obsolete.txt")
	if err := os.MkdirAll(filepath.Dir(obsolete), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(obsolete, []byte("obsolete"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if err := s.Remove("/remote/source/nested/obsolete.txt"); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}
	if _, err := os.Stat(obsolete); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected local remove to delete %s, got err=%v", obsolete, err)
	}

	renamed := filepath.Join(destDir, "nested", "renamed.txt")
	if err := os.WriteFile(renamed, []byte("stale"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := s.Rename("/remote/source/nested/renamed.txt"); err != nil {
		t.Fatalf("Rename() error = %v", err)
	}
	if _, err := os.Stat(renamed); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected local rename handling to remove stale file %s, got err=%v", renamed, err)
	}
}

func TestFTPPullClientSync_WriteSkipsSecondRunWhenSizeAndPreciseModTimeMatch(t *testing.T) {
	ftpModTime := time.Unix(1710000000, 0)
	ftpDriver := &fakeFTPDriver{
		openFn: func(path string) (http.File, error) {
			return newFakeHTTPFile("same-content", fakeFileInfo{name: filepath.Base(path), size: int64(len("same-content")), mode: 0o644, modTime: ftpModTime}), nil
		},
		statFn: func(path string) (fs.FileInfo, error) {
			return fakeFileInfo{name: filepath.Base(path), size: int64(len("same-content")), mode: 0o644, modTime: ftpModTime}, nil
		},
		getFileTimeFn: func(path string) (time.Time, time.Time, time.Time, error) {
			return ftpModTime, ftpModTime, ftpModTime, nil
		},
	}

	s, remotePath, localDest := newTestFTPPullClientSync(t, ftpDriver)
	if err := os.WriteFile(localDest, []byte("same-content"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := os.Chtimes(localDest, ftpModTime, ftpModTime); err != nil {
		t.Fatalf("Chtimes() error = %v", err)
	}

	before, err := os.ReadFile(localDest)
	if err != nil {
		t.Fatalf("ReadFile() before error = %v", err)
	}

	if err := s.Write(remotePath); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	after, err := os.ReadFile(localDest)
	if err != nil {
		t.Fatalf("ReadFile() after error = %v", err)
	}
	if string(after) != string(before) {
		t.Fatalf("expected second run no-op to preserve file content, before=%q after=%q", before, after)
	}
}

func TestFTPPullClientSync_WriteDoesNotSkipWhenMetadataIsAmbiguous(t *testing.T) {
	coarseListTime := time.Unix(1710000000, 0)
	preciseFTPTime := coarseListTime.Add(45 * time.Second)
	ftpDriver := &fakeFTPDriver{
		openFn: func(path string) (http.File, error) {
			return newFakeHTTPFile("new-content", fakeFileInfo{name: filepath.Base(path), size: int64(len("new-content")), mode: 0o644, modTime: coarseListTime}), nil
		},
		statFn: func(path string) (fs.FileInfo, error) {
			return fakeFileInfo{name: filepath.Base(path), size: int64(len("new-content")), mode: 0o644, modTime: coarseListTime}, nil
		},
		getFileTimeFn: func(path string) (time.Time, time.Time, time.Time, error) {
			return preciseFTPTime, preciseFTPTime, preciseFTPTime, nil
		},
	}

	s, remotePath, localDest := newTestFTPPullClientSync(t, ftpDriver)
	if err := os.WriteFile(localDest, []byte("old-content"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := os.Chtimes(localDest, coarseListTime, coarseListTime); err != nil {
		t.Fatalf("Chtimes() error = %v", err)
	}

	if err := s.Write(remotePath); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	content, err := os.ReadFile(localDest)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(content) != "new-content" {
		t.Fatalf("expected ambiguous FTP metadata to force safe rewrite, got %q", content)
	}

	stat, err := os.Stat(localDest)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if !stat.ModTime().Equal(preciseFTPTime) {
		t.Fatalf("expected rewritten file mod time %v, got %v", preciseFTPTime, stat.ModTime())
	}
}

func testSyncOption(t *testing.T, source, dest core.VFS) Option {
	t.Helper()

	pi, err := ignore.NewPathIgnore("", false, logger.NewTestLogger())
	if err != nil {
		t.Fatalf("NewPathIgnore() error = %v", err)
	}

	return Option{
		Source:            source,
		Dest:              dest,
		ChunkSize:         1,
		ChecksumAlgorithm: hashutil.DefaultHash,
		PathIgnore:        pi,
		Logger:            logger.NewTestLogger(),
	}
}

func newTestFTPPullClientSync(t *testing.T, ftpDriver *fakeFTPDriver) (*ftpPullClientSync, string, string) {
	t.Helper()

	destDir := t.TempDir()
	remoteRoot := "/remote/source"
	remotePath := remoteRoot + "/file.txt"
	localDest := filepath.Join(destDir, "file.txt")

	opt := testSyncOption(t, core.NewVFS("ftp://127.0.0.1:21?path=./dest&remote_path="+remoteRoot+"&ftp_user=user&ftp_pass=pass&ftp_passive=true"), core.NewDiskVFS(destDir))
	syncer, err := newDiskSync(opt)
	if err != nil {
		t.Fatalf("newDiskSync() error = %v", err)
	}

	s := &ftpPullClientSync{
		driverPullClientSync: newDriverPullClientSync(*syncer),
	}
	s.driver = ftpDriver
	s.diskSync.sourceAbsPath = remoteRoot
	s.diskSync.statFn = ftpDriver.Stat
	s.diskSync.getFileTimeFn = ftpDriver.GetFileTime
	s.diskSync.isDirFn = func(path string) (bool, error) {
		return false, nil
	}

	if err := os.MkdirAll(filepath.Dir(localDest), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	file, err := os.Create(localDest)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	return s, remotePath, localDest
}

var _ driver.Driver = (*fakeFTPDriver)(nil)
var _ http.File = (*fakeHTTPFile)(nil)
var _ fs.FileInfo = (*fakeFileInfo)(nil)
var _ ignore.PathIgnore = (*fakePathIgnore)(nil)
var _ = contract.FileInfo{}
