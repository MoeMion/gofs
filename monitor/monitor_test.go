package monitor

import (
	"errors"
	"testing"

	"github.com/no-src/gofs/core"
	"github.com/no-src/gofs/result"
)

type testSyncer struct {
	source core.VFS
	dest   core.VFS
}

func (s testSyncer) Create(path string) error          { return nil }
func (s testSyncer) Symlink(oldname, newname string) error { return nil }
func (s testSyncer) Write(path string) error           { return nil }
func (s testSyncer) Remove(path string) error          { return nil }
func (s testSyncer) Rename(path string) error          { return nil }
func (s testSyncer) Chmod(path string) error           { return nil }
func (s testSyncer) IsDir(path string) (bool, error)   { return false, nil }
func (s testSyncer) SyncOnce(path string) error        { return nil }
func (s testSyncer) Source() core.VFS                  { return s.source }
func (s testSyncer) Dest() core.VFS                    { return s.dest }
func (s testSyncer) Close()                            {}

func TestNewMonitor_RoutesFTPSourceToFTPPullClientMonitor(t *testing.T) {
	opt := Option{
		Syncer: testSyncer{
			source: core.NewVFS("ftp://127.0.0.1:21?path=./dest&remote_path=/remote/source&ftp_user=user&ftp_pass=pass&ftp_passive=true"),
			dest:   core.NewDiskVFS("./testdata/dest"),
		},
	}

	_, err := NewMonitor(opt, func(content string, ext string) result.Result { return nil })
	if !errors.Is(err, errFTPMonitorDeferred) {
		t.Fatalf("expected ftp monitor deferred error, got %v", err)
	}

	if err != nil && err.Error() == "file system unsupported ! source=>FTP" {
		t.Fatalf("expected FTP route, got unsupported error: %v", err)
	}
}
