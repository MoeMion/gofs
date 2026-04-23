package sync

import (
	"errors"
	"testing"

	"github.com/no-src/gofs/core"
)

func TestNewSync_RoutesDiskToFTPToFTPPushClientSync(t *testing.T) {
	opt := Option{
		Source: core.NewDiskVFS("./testdata/source"),
		Dest:   core.NewVFS("ftp://127.0.0.1:21?path=./source&remote_path=/remote/dest&ftp_user=user&ftp_pass=pass&ftp_passive=true"),
	}

	_, err := NewSync(opt)
	if !errors.Is(err, errFTPBackendDeferred) {
		t.Fatalf("expected ftp deferred error, got %v", err)
	}

	if errors.Is(err, errFileSystemUnsupported) {
		t.Fatalf("expected FTP route, got unsupported error: %v", err)
	}
	}

func TestNewSync_RoutesFTPToDiskToFTPPullClientSync(t *testing.T) {
	opt := Option{
		Source: core.NewVFS("ftp://127.0.0.1:21?path=./dest&remote_path=/remote/source&ftp_user=user&ftp_pass=pass&ftp_passive=true"),
		Dest:   core.NewDiskVFS("./testdata/dest"),
	}

	_, err := NewSync(opt)
	if !errors.Is(err, errFTPBackendDeferred) {
		t.Fatalf("expected ftp deferred error, got %v", err)
	}

	if errors.Is(err, errFileSystemUnsupported) {
		t.Fatalf("expected FTP route, got unsupported error: %v", err)
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
