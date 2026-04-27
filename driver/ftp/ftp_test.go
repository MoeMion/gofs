package ftp

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	pathpkg "path"
	"strings"
	"testing"
	"time"

	ftpclient "github.com/jlaffaye/ftp"
	"github.com/no-src/gofs/core"
	"github.com/no-src/gofs/logger"
	"github.com/no-src/gofs/wait"
)

type fakeRetry struct {
	called int
	do     func(func() error) error
}

func (r *fakeRetry) Do(f func() error, desc string) wait.Wait {
	r.called++
	err := error(nil)
	if r.do != nil {
		err = r.do(f)
	} else {
		err = f()
	}
	return fakeWait{err: err}
}

func (r *fakeRetry) DoWithContext(_ context.Context, f func() error, desc string) wait.Wait {
	return r.Do(f, desc)
}

func (r *fakeRetry) Count() int { return 1 }

func (r *fakeRetry) WaitTime() time.Duration { return 0 }

type fakeWait struct{ err error }

func (w fakeWait) Wait() error { return w.err }

type fakeConn struct {
	loginUser         string
	loginPassword     string
	loginErr          error
	typeCalled        ftpclient.TransferType
	typeErr           error
	quitCalled        bool
	noOpErr           error
	entries           map[string]*ftpclient.Entry
	list              map[string][]*ftpclient.Entry
	walk              *fakeWalker
	retrPath          string
	retrErr           error
	storPath          string
	storBody          string
	storErr           error
	deletePath        string
	deleteErr         error
	makeDirs          []string
	makeDirErrs       map[string]error
	removeDirRecur    string
	removeDirRecurErr error
	renameFrom        string
	renameTo          string
	renameErr         error
	getEntryErr       error
	getTime           time.Time
	getTimeErr        error
	setTimePath       string
	setTimeValue      time.Time
	setTimeErr        error
	getTimeSupported  bool
	setTimeSupported  bool
	listTimePrecise   bool
}

func (c *fakeConn) Login(user, password string) error {
	c.loginUser = user
	c.loginPassword = password
	return c.loginErr
}
func (c *fakeConn) NoOp() error { return c.noOpErr }
func (c *fakeConn) Quit() error { c.quitCalled = true; return nil }
func (c *fakeConn) Type(transferType ftpclient.TransferType) error {
	c.typeCalled = transferType
	return c.typeErr
}
func (c *fakeConn) Walk(root string) ftpWalker {
	if c.walk == nil {
		return &fakeWalker{}
	}
	return c.walk
}
func (c *fakeConn) CurrentDir() (string, error)                  { return "/", nil }
func (c *fakeConn) ChangeDir(path string) error                  { return nil }
func (c *fakeConn) List(path string) ([]*ftpclient.Entry, error) { return c.list[path], nil }
func (c *fakeConn) NameList(path string) ([]string, error)       { return nil, nil }
func (c *fakeConn) Retr(path string) (*ftpclient.Response, error) {
	c.retrPath = path
	if c.retrErr != nil {
		return nil, c.retrErr
	}
	return &ftpclient.Response{}, nil
}
func (c *fakeConn) Stor(path string, r io.Reader) error {
	c.storPath = path
	if r != nil {
		body, _ := io.ReadAll(r)
		c.storBody = string(body)
	}
	return c.storErr
}
func (c *fakeConn) Delete(path string) error { c.deletePath = path; return c.deleteErr }
func (c *fakeConn) MakeDir(path string) error {
	c.makeDirs = append(c.makeDirs, path)
	if c.makeDirErrs != nil {
		return c.makeDirErrs[path]
	}
	return nil
}
func (c *fakeConn) RemoveDirRecur(path string) error {
	c.removeDirRecur = path
	return c.removeDirRecurErr
}
func (c *fakeConn) Rename(from, to string) error {
	c.renameFrom = from
	c.renameTo = to
	return c.renameErr
}
func (c *fakeConn) GetEntry(path string) (*ftpclient.Entry, error) {
	if c.getEntryErr != nil {
		return nil, c.getEntryErr
	}
	entry, ok := c.entries[path]
	if !ok {
		return nil, errors.New("550 not found")
	}
	return entry, nil
}
func (c *fakeConn) GetTime(path string) (time.Time, error) { return c.getTime, c.getTimeErr }
func (c *fakeConn) SetTime(path string, t time.Time) error {
	c.setTimePath = path
	c.setTimeValue = t
	return c.setTimeErr
}
func (c *fakeConn) IsGetTimeSupported() bool  { return c.getTimeSupported }
func (c *fakeConn) IsSetTimeSupported() bool  { return c.setTimeSupported }
func (c *fakeConn) IsTimePreciseInList() bool { return c.listTimePrecise }

type fakeWalker struct {
	items []*fakeWalkerItem
	idx   int
}

type fakeWalkerItem struct {
	path  string
	entry *ftpclient.Entry
	err   error
}

func (w *fakeWalker) Next() bool {
	if w.idx >= len(w.items) {
		return false
	}
	w.idx++
	return true
}
func (w *fakeWalker) Err() error {
	if w.idx == 0 || w.idx > len(w.items) {
		return nil
	}
	return w.items[w.idx-1].err
}
func (w *fakeWalker) Path() string           { return w.items[w.idx-1].path }
func (w *fakeWalker) Stat() *ftpclient.Entry { return w.items[w.idx-1].entry }

func newTestDriver(conn ftpConn) *ftpDriver {
	return newFTPDriver("127.0.0.1:21", core.FTPConfig{Username: "demo", Password: "pw", Timeout: "5s", PassiveMode: true}, true, nil, 0, logger.NewTestLogger(), func(addr string, options ...ftpclient.DialOption) (ftpConn, error) {
		return conn, nil
	})
}

func TestFTPDriverConnect(t *testing.T) {
	t.Run("connects and switches to binary mode", func(t *testing.T) {
		conn := &fakeConn{}
		d := newTestDriver(conn)

		if err := d.Connect(); err != nil {
			t.Fatalf("expect connect success, but got error => %v", err)
		}
		if conn.loginUser != "demo" || conn.loginPassword != "pw" {
			t.Fatalf("expect login credentials recorded, but got %q/%q", conn.loginUser, conn.loginPassword)
		}
		if conn.typeCalled != ftpclient.TransferTypeBinary {
			t.Fatalf("expect binary transfer mode, but got => %s", conn.typeCalled)
		}
		if !d.online {
			t.Fatal("expect driver to be online after connect")
		}
	})

	t.Run("rejects active mode", func(t *testing.T) {
		d := newFTPDriver("127.0.0.1:21", core.FTPConfig{Username: "demo", Password: "pw", PassiveMode: false}, true, nil, 0, logger.NewTestLogger(), nil)
		err := d.Connect()
		if err == nil {
			t.Fatal("expect active mode error, but got nil")
		}
		if !strings.Contains(err.Error(), "active mode") {
			t.Fatalf("expect active mode error, but got => %v", err)
		}
	})
}

func TestFTPPathCodec(t *testing.T) {
	t.Run("default config uses auto encoding", func(t *testing.T) {
		codec, err := newFTPPathCodec(core.FTPConfig{})
		if err != nil {
			t.Fatalf("expect no error, but got => %v", err)
		}
		if codec.mode != ftpEncodingAuto {
			t.Fatalf("expect mode %q, but got %q", ftpEncodingAuto, codec.mode)
		}
	})

	t.Run("explicit gbk disables utf8 feature and round trips chinese path", func(t *testing.T) {
		codec, err := newFTPPathCodec(core.FTPConfig{Encoding: "gbk"})
		if err != nil {
			t.Fatalf("expect no error, but got => %v", err)
		}
		if !codec.disableUTF8Feature() {
			t.Fatal("expect gbk mode to disable utf8 feature")
		}
		encoded, err := codec.encodePath("/中文/目录.txt")
		if err != nil {
			t.Fatalf("expect no encode error, but got => %v", err)
		}
		decoded := codec.decodePath(encoded)
		if decoded != pathpkg.Clean("/中文/目录.txt") {
			t.Fatalf("expect decoded path %q, but got %q", pathpkg.Clean("/中文/目录.txt"), decoded)
		}
	})

	t.Run("auto mode decodes gbk bytes and leaves utf8 untouched", func(t *testing.T) {
		codec, err := newFTPPathCodec(core.FTPConfig{Encoding: "auto"})
		if err != nil {
			t.Fatalf("expect no error, but got => %v", err)
		}
		encoded, err := (&ftpPathCodec{mode: ftpEncodingGBK}).encodePath("中文")
		if err != nil {
			t.Fatalf("expect no encode error, but got => %v", err)
		}
		decoded := codec.decodeName(encoded)
		if decoded != "中文" {
			t.Fatalf("expect decoded name %q, but got %q", "中文", decoded)
		}
		if codec.decodeName("utf8-name") != "utf8-name" {
			t.Fatal("expect utf8 name to remain unchanged in auto mode")
		}
	})

	t.Run("invalid explicit encoding returns error", func(t *testing.T) {
		_, err := newFTPPathCodec(core.FTPConfig{Encoding: "shift-jis"})
		if err == nil {
			t.Fatal("expect invalid encoding error, but got nil")
		}
	})
}

func TestFTPDriverWalkDir(t *testing.T) {
	conn := &fakeConn{walk: &fakeWalker{items: []*fakeWalkerItem{{path: "/root", entry: &ftpclient.Entry{Name: "root", Type: ftpclient.EntryTypeFolder}}, {path: "/root/child", entry: &ftpclient.Entry{Name: "child", Type: ftpclient.EntryTypeFolder}}, {path: "/root/child/file.txt", entry: &ftpclient.Entry{Name: "file.txt", Type: ftpclient.EntryTypeFile, Size: 3}}}}}
	d := newTestDriver(conn)
	d.client = conn
	d.online = true

	var visited []string
	err := d.WalkDir("/root", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		kind := "file"
		if entry.IsDir() {
			kind = "dir"
		}
		visited = append(visited, path+":"+kind)
		return nil
	})
	if err != nil {
		t.Fatalf("expect walk success, but got error => %v", err)
	}
	expect := []string{"/root:dir", "/root/child:dir", "/root/child/file.txt:file"}
	if strings.Join(visited, ",") != strings.Join(expect, ",") {
		t.Fatalf("expect visited %v, but got %v", expect, visited)
	}
}

func TestFTPDriverFileOperations(t *testing.T) {
	now := time.Date(2026, 4, 24, 8, 0, 0, 0, time.UTC)
	conn := &fakeConn{
		entries: map[string]*ftpclient.Entry{
			"/dir":           {Name: "dir", Type: ftpclient.EntryTypeFolder, Time: now},
			"/file.txt":      {Name: "file.txt", Type: ftpclient.EntryTypeFile, Size: 7, Time: now},
			"/renamed.txt":   {Name: "renamed.txt", Type: ftpclient.EntryTypeFile, Size: 7, Time: now},
			"/time-fallback": {Name: "time-fallback", Type: ftpclient.EntryTypeFile, Size: 5, Time: now},
		},
		list: map[string][]*ftpclient.Entry{
			"/dir": {{Name: "child.txt", Type: ftpclient.EntryTypeFile, Size: 4, Time: now}},
		},
		getTimeSupported: false,
	}
	d := newTestDriver(conn)
	d.client = conn
	d.online = true

	t.Run("write delegates to stor", func(t *testing.T) {
		tmpDir := t.TempDir()
		src := filepath.Join(tmpDir, "src.txt")
		if err := os.WriteFile(src, []byte("content"), 0o644); err != nil {
			t.Fatalf("write temp source error => %v", err)
		}
		if err := d.Write(src, "/remote.txt"); err != nil {
			t.Fatalf("expect write success, but got error => %v", err)
		}
		if conn.storPath != "/remote.txt" || conn.storBody != "content" {
			t.Fatalf("expect stor delegate to /remote.txt with content, but got path=%q body=%q", conn.storPath, conn.storBody)
		}
	})

	t.Run("open delegates to retr", func(t *testing.T) {
		f, err := d.Open("/file.txt")
		if err != nil {
			t.Fatalf("expect open success, but got error => %v", err)
		}
		if conn.retrPath != "/file.txt" {
			t.Fatalf("expect retr path /file.txt, but got => %s", conn.retrPath)
		}
		if _, ok := f.(http.File); !ok {
			t.Fatal("expect returned value to satisfy http.File")
		}
	})

	t.Run("remove and rename delegate", func(t *testing.T) {
		if err := d.Remove("/file.txt"); err != nil {
			t.Fatalf("expect remove success, but got error => %v", err)
		}
		if conn.deletePath != "/file.txt" {
			t.Fatalf("expect delete path /file.txt, but got => %s", conn.deletePath)
		}
		if err := d.Remove("/dir"); err != nil {
			t.Fatalf("expect remove dir success, but got error => %v", err)
		}
		if conn.removeDirRecur != "/dir" {
			t.Fatalf("expect recursive remove /dir, but got => %s", conn.removeDirRecur)
		}
		if err := d.Rename("/file.txt", "/renamed.txt"); err != nil {
			t.Fatalf("expect rename success, but got error => %v", err)
		}
		if conn.renameFrom != "/file.txt" || conn.renameTo != "/renamed.txt" {
			t.Fatalf("expect rename /file.txt -> /renamed.txt, but got %q -> %q", conn.renameFrom, conn.renameTo)
		}
	})

	t.Run("file time falls back conservatively", func(t *testing.T) {
		cTime, aTime, mTime, err := d.GetFileTime("/time-fallback")
		if err != nil {
			t.Fatalf("expect get file time success, but got error => %v", err)
		}
		if !cTime.Equal(now) || !aTime.Equal(now) || !mTime.Equal(now) {
			t.Fatalf("expect fallback times to use entry time => %v, %v, %v", cTime, aTime, mTime)
		}
	})

	t.Run("unsupported operations return explicit errors", func(t *testing.T) {
		if _, err := d.ReadLink("/file.txt"); !errors.Is(err, errFTPReadLinkUnsupported) {
			t.Fatalf("expect readlink unsupported error, but got => %v", err)
		}
		if err := d.Symlink("/a", "/b"); !errors.Is(err, errFTPSymlinkUnsupported) {
			t.Fatalf("expect symlink unsupported error, but got => %v", err)
		}
		if err := d.Chtimes("/file.txt", now, now); !errors.Is(err, errFTPChtimesUnsupported) {
			t.Fatalf("expect chtimes unsupported error, but got => %v", err)
		}
	})
}

func TestFTPDriverReconnect(t *testing.T) {
	transportErr := io.EOF
	firstConn := &fakeConn{noOpErr: transportErr}
	secondConn := &fakeConn{entries: map[string]*ftpclient.Entry{"/after": {Name: "after", Type: ftpclient.EntryTypeFile, Time: time.Now()}}}
	dials := 0
	r := &fakeRetry{do: func(f func() error) error { return f() }}
	d := newFTPDriver("127.0.0.1:21", core.FTPConfig{Username: "demo", Password: "pw", PassiveMode: true}, true, r, 0, logger.NewTestLogger(), func(addr string, options ...ftpclient.DialOption) (ftpConn, error) {
		dials++
		if dials == 1 {
			return firstConn, nil
		}
		return secondConn, nil
	})
	if err := d.Connect(); err != nil {
		t.Fatalf("initial connect error => %v", err)
	}
	if err := d.Create("/after"); err != nil {
		t.Fatalf("expect reconnect and retry success, but got error => %v", err)
	}
	if dials != 2 {
		t.Fatalf("expect reconnect dial count 2, but got => %d", dials)
	}
	if r.called != 1 {
		t.Fatalf("expect retry helper called once, but got => %d", r.called)
	}

	t.Run("returns explicit reconnect failure", func(t *testing.T) {
		rf := &fakeRetry{do: func(f func() error) error { return errors.New("reconnect failed") }}
		dials := 0
		d := newFTPDriver("127.0.0.1:21", core.FTPConfig{Username: "demo", Password: "pw", PassiveMode: true}, true, rf, 0, logger.NewTestLogger(), func(addr string, options ...ftpclient.DialOption) (ftpConn, error) {
			dials++
			return &fakeConn{noOpErr: io.EOF}, nil
		})
		if err := d.Connect(); err != nil {
			t.Fatalf("connect error => %v", err)
		}
		err := d.Create("/blocked")
		if err == nil || !strings.Contains(err.Error(), "reconnect failed") {
			t.Fatalf("expect explicit reconnect failure, but got => %v", err)
		}
	})
}
