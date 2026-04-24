package monitor

import (
	"errors"
	"strings"
	"testing"

	"github.com/no-src/gofs/core"
	"github.com/no-src/gofs/logger"
	"github.com/no-src/gofs/result"
)

type testSyncer struct {
	source core.VFS
	dest   core.VFS
}

func (s testSyncer) Create(path string) error              { return nil }
func (s testSyncer) Symlink(oldname, newname string) error { return nil }
func (s testSyncer) Write(path string) error               { return nil }
func (s testSyncer) Remove(path string) error              { return nil }
func (s testSyncer) Rename(path string) error              { return nil }
func (s testSyncer) Chmod(path string) error               { return nil }
func (s testSyncer) IsDir(path string) (bool, error)       { return false, nil }
func (s testSyncer) SyncOnce(path string) error            { return nil }
func (s testSyncer) Source() core.VFS                      { return s.source }
func (s testSyncer) Dest() core.VFS                        { return s.dest }
func (s testSyncer) Close()                                {}

type trackingSyncer struct {
	testSyncer
	syncOnceCalls int
	closeCalls    int
}

func (s *trackingSyncer) SyncOnce(path string) error {
	s.syncOnceCalls++
	return nil
}

func (s *trackingSyncer) Close() {
	s.closeCalls++
}

func TestNewMonitor_RoutesFTPSourceToFTPPullClientMonitor(t *testing.T) {
	opt := Option{
		Syncer: testSyncer{
			source: core.NewVFS("ftp://127.0.0.1:21?path=./dest&remote_path=/remote/source&ftp_user=user&ftp_pass=pass&ftp_passive=true"),
			dest:   core.NewDiskVFS("./testdata/dest"),
		},
	}

	m, err := NewMonitor(opt, func(content string, ext string) result.Result { return nil })
	if err != nil {
		t.Fatalf("expected FTP route without error, got %v", err)
	}

	if _, ok := m.(*ftpPullClientMonitor); !ok {
		t.Fatalf("expected *ftpPullClientMonitor, got %T", m)
	}
}

func TestFTPPullClientMonitorStart_SyncOnce(t *testing.T) {
	syncer := &trackingSyncer{
		testSyncer: testSyncer{
			source: core.NewVFS("ftp://127.0.0.1:21?path=./dest&remote_path=/remote/source&ftp_user=user&ftp_pass=pass&ftp_passive=true"),
			dest:   core.NewDiskVFS("./testdata/dest"),
		},
	}
	m, err := NewFTPPullClientMonitor(Option{
		SyncOnce: true,
		Syncer:   syncer,
		Logger:   logger.NewTestLogger(),
	})
	if err != nil {
		t.Fatalf("create ftp pull monitor error => %v", err)
	}

	wd, err := m.Start()
	if err != nil {
		t.Fatalf("start ftp pull monitor error => %v", err)
	}
	if wd == nil {
		t.Fatal("expected wait handle for sync_once mode")
	}
	if err := wd.Wait(); err != nil {
		t.Fatalf("wait ftp pull monitor error => %v", err)
	}
	if syncer.syncOnceCalls != 1 {
		t.Fatalf("expected sync_once call count 1, got %d", syncer.syncOnceCalls)
	}
	if syncer.closeCalls != 1 {
		t.Fatalf("expected close call count 1, got %d", syncer.closeCalls)
	}
}

func TestFTPPullClientMonitorStart_SyncCron(t *testing.T) {
	syncer := &trackingSyncer{
		testSyncer: testSyncer{
			source: core.NewVFS("ftp://127.0.0.1:21?path=./dest&remote_path=/remote/source&ftp_user=user&ftp_pass=pass&ftp_passive=true"),
			dest:   core.NewDiskVFS("./testdata/dest"),
		},
	}
	m, err := NewFTPPullClientMonitor(Option{
		Syncer: syncer,
		Logger: logger.NewTestLogger(),
	})
	if err != nil {
		t.Fatalf("create ftp pull monitor error => %v", err)
	}
	if err := m.SyncCron("*/5 * * * * *"); err != nil {
		t.Fatalf("register ftp sync cron error => %v", err)
	}

	wd, err := m.Start()
	if err != nil {
		t.Fatalf("start ftp pull monitor error => %v", err)
	}
	if wd == nil {
		t.Fatal("expected wait handle for sync_cron mode")
	}
	if err := m.Shutdown(); err != nil {
		t.Fatalf("shutdown ftp pull monitor error => %v", err)
	}
	if err := wd.Wait(); err != nil {
		t.Fatalf("wait ftp pull monitor error => %v", err)
	}
	if syncer.closeCalls != 1 {
		t.Fatalf("expected close call count 1, got %d", syncer.closeCalls)
	}
}

func TestFTPPullClientMonitorStart_RequiresSyncOnceOrSyncCron(t *testing.T) {
	m, err := NewFTPPullClientMonitor(Option{
		Syncer: testSyncer{
			source: core.NewVFS("ftp://127.0.0.1:21?path=./dest&remote_path=/remote/source&ftp_user=user&ftp_pass=pass&ftp_passive=true"),
			dest:   core.NewDiskVFS("./testdata/dest"),
		},
		Logger: logger.NewTestLogger(),
	})
	if err != nil {
		t.Fatalf("create ftp pull monitor error => %v", err)
	}

	wd, err := m.Start()
	if wd != nil {
		t.Fatal("expected nil wait handle when startup is rejected")
	}
	if !errors.Is(err, errFTPMonitorRequiresPolling) {
		t.Fatalf("expected ftp polling requirement error, got %v", err)
	}
	if !strings.Contains(err.Error(), "ftp source monitor requires -sync_once or -sync_cron because FTP has no event stream") {
		t.Fatalf("expected explicit FTP polling requirement error, got %v", err)
	}
}
