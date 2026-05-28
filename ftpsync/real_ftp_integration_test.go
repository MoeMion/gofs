package ftpsync

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	ftpserver "github.com/fclairamb/ftpserverlib"
	"github.com/spf13/afero"
)

const (
	ftpFixtureUser     = "ftp-user"
	ftpFixturePassword = "secret-password"
)

type ftpTestServer struct {
	Host     string
	Port     int
	User     string
	Password string
	Root     string
}

func TestIntegrationRealFTPLocalToFTP(t *testing.T) {
	server := startLoopbackFTPServer(t)
	sourceRoot := t.TempDir()
	writeTestFile(t, filepath.Join(sourceRoot, "root.txt"), "root")
	writeTestFile(t, filepath.Join(sourceRoot, "nested", "child.txt"), "child")
	writeTestFile(t, filepath.Join(server.Root, "incoming", "stale.txt"), "stale")

	svc, err := NewFTPSyncService(Options{
		Direction: DirectionLocalToFTP,
		Source:    Endpoint{LocalPath: sourceRoot},
		Destination: Endpoint{FTP: FTPOptions{
			Host:         server.Host,
			Port:         server.Port,
			Username:     server.User,
			Password:     server.Password,
			RemotePath:   "/incoming",
			PassiveMode:  true,
			Timeout:      5 * time.Second,
			PathEncoding: "utf8",
		}},
	})
	if err != nil {
		t.Fatalf("construct service: %v", err)
	}

	result, err := svc.SyncOnce(context.Background())
	if err != nil {
		t.Fatalf("SyncOnce local to FTP: %v", err)
	}
	if result.Direction != DirectionLocalToFTP || result.FilesAttempted != 2 || result.FailureCount != 0 {
		t.Fatalf("unexpected local to FTP result: %#v", result)
	}
	assertFileContent(t, filepath.Join(server.Root, "incoming", "root.txt"), "root")
	assertFileContent(t, filepath.Join(server.Root, "incoming", "nested", "child.txt"), "child")
	assertPathMissing(t, filepath.Join(server.Root, "incoming", "stale.txt"))
}

func TestIntegrationRealFTPToLocal(t *testing.T) {
	server := startLoopbackFTPServer(t)
	writeTestFile(t, filepath.Join(server.Root, "outgoing", "root.txt"), "root")
	writeTestFile(t, filepath.Join(server.Root, "outgoing", "nested", "child.txt"), "child")
	destinationRoot := filepath.Join(t.TempDir(), "dest")

	svc, err := NewFTPSyncService(Options{
		Direction: DirectionFTPToLocal,
		Source: Endpoint{FTP: FTPOptions{
			Host:         server.Host,
			Port:         server.Port,
			Username:     server.User,
			Password:     server.Password,
			RemotePath:   "/outgoing",
			PassiveMode:  true,
			Timeout:      5 * time.Second,
			PathEncoding: "utf8",
		}},
		Destination: Endpoint{LocalPath: destinationRoot},
	})
	if err != nil {
		t.Fatalf("construct service: %v", err)
	}

	result, err := svc.SyncOnce(context.Background())
	if err != nil {
		t.Fatalf("SyncOnce FTP to local: %v", err)
	}
	if result.Direction != DirectionFTPToLocal || result.FilesAttempted != 2 || result.FailureCount != 0 {
		t.Fatalf("unexpected FTP to local result: %#v", result)
	}
	assertFileContent(t, filepath.Join(destinationRoot, "root.txt"), "root")
	assertFileContent(t, filepath.Join(destinationRoot, "nested", "child.txt"), "child")
}

type loopbackFTPDriver struct {
	root string
	fs   afero.Fs
}

func (d *loopbackFTPDriver) GetSettings() (*ftpserver.Settings, error) {
	return &ftpserver.Settings{
		ListenAddr:          "127.0.0.1:0",
		PublicHost:          "127.0.0.1",
		DisableActiveMode:   true,
		DefaultTransferType: ftpserver.TransferTypeBinary,
	}, nil
}

func (d *loopbackFTPDriver) ClientConnected(cc ftpserver.ClientContext) (string, error) {
	cc.SetDebug(false)
	return "ftpsync integration fixture", nil
}

func (d *loopbackFTPDriver) ClientDisconnected(cc ftpserver.ClientContext) {}

func (d *loopbackFTPDriver) AuthUser(cc ftpserver.ClientContext, user, pass string) (ftpserver.ClientDriver, error) {
	if user != ftpFixtureUser || pass != ftpFixturePassword {
		return nil, errors.New("bad test FTP credentials")
	}
	return &loopbackFTPClientDriver{Fs: d.fs}, nil
}

func (d *loopbackFTPDriver) GetTLSConfig() (*tls.Config, error) {
	return nil, errors.New("TLS disabled for plain FTP integration fixture")
}

type loopbackFTPClientDriver struct {
	afero.Fs
}

func startLoopbackFTPServer(t *testing.T) ftpTestServer {
	t.Helper()
	root := t.TempDir()
	driver := &loopbackFTPDriver{root: root, fs: afero.NewBasePathFs(afero.NewOsFs(), root)}
	server := ftpserver.NewFtpServer(driver)
	server.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	if err := server.Listen(); err != nil {
		t.Fatalf("listen on 127.0.0.1 FTP fixture: %v", err)
	}
	done := make(chan error, 1)
	go func() { done <- server.Serve() }()
	t.Cleanup(func() {
		if err := server.Stop(); err != nil {
			t.Fatalf("stop FTP fixture: %v", err)
		}
		select {
		case err := <-done:
			if err != nil && !errors.Is(err, net.ErrClosed) && !errors.Is(err, io.EOF) {
				t.Fatalf("serve FTP fixture: %v", err)
			}
		case <-time.After(time.Second):
			t.Fatalf("FTP fixture did not stop")
		}
	})
	addr := server.Addr()
	host, portValue, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("parse FTP fixture address %q: %v", addr, err)
	}
	port, err := net.LookupPort("tcp", portValue)
	if err != nil {
		t.Fatalf("parse FTP fixture port %q: %v", portValue, err)
	}
	if host != "127.0.0.1" {
		t.Fatalf("FTP fixture must bind to loopback, got %q", host)
	}
	return ftpTestServer{Host: host, Port: port, User: ftpFixtureUser, Password: ftpFixturePassword, Root: root}
}

func writeTestFile(t *testing.T, filePath string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(filePath), err)
	}
	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", filePath, err)
	}
}

func assertFileContent(t *testing.T, filePath string, want string) {
	t.Helper()
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("read %s: %v", filePath, err)
	}
	if string(content) != want {
		t.Fatalf("%s content = %q, want %q", filePath, content, want)
	}
}

func assertPathMissing(t *testing.T, filePath string) {
	t.Helper()
	if _, err := os.Stat(filePath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected %s to be missing, stat err=%v", filePath, err)
	}
}
