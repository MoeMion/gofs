package ftpsync

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestBackgroundHandleWait(t *testing.T) {
	defer withSyncOnceExecutor(runSyncOnceScaffold)()
	runSyncOnce = func(ctx context.Context, svc *FTPSyncService, adapter syncOnceAdapter, result *Result) error {
		return nil
	}

	svc, err := NewFTPSyncService(completeLocalToFTPOptionsForBackground())
	if err != nil {
		t.Fatalf("construct service: %v", err)
	}
	handle, err := svc.StartBackground(context.Background())
	if err != nil {
		t.Fatalf("StartBackground returned error: %v", err)
	}
	if handle == nil {
		t.Fatalf("expect non-nil background handle")
	}
	if handle.Done() == nil {
		t.Fatalf("expect non-nil Done channel")
	}
	if err := handle.Stop(context.Background()); err != nil {
		t.Fatalf("Stop returned error: %v", err)
	}
	if err := handle.Wait(); err != nil {
		t.Fatalf("Wait returned error: %v", err)
	}
	select {
	case <-handle.Done():
	case <-time.After(time.Second):
		t.Fatalf("expect Done to close after Stop")
	}
}

func TestStartBackgroundInitialSyncBeforeReady(t *testing.T) {
	defer withSyncOnceExecutor(runSyncOnceScaffold)()
	initialSyncDone := make(chan struct{})
	runSyncOnce = func(ctx context.Context, svc *FTPSyncService, adapter syncOnceAdapter, result *Result) error {
		close(initialSyncDone)
		return errors.New("initial catch-up failed")
	}

	svc, err := NewFTPSyncService(completeLocalToFTPOptionsForBackground())
	if err != nil {
		t.Fatalf("construct service: %v", err)
	}
	handle, err := svc.StartBackground(context.Background())
	if err != nil {
		t.Fatalf("StartBackground returned error: %v", err)
	}
	defer func() { _ = handle.Stop(context.Background()) }()

	readyHandle, ok := handle.(interface{ Ready() <-chan struct{} })
	if !ok {
		t.Fatalf("background handle should expose internal readiness for tests")
	}

	select {
	case <-readyHandle.Ready():
		select {
		case <-initialSyncDone:
		default:
			t.Fatalf("background handle became ready before initial sync completed")
		}
	case <-time.After(time.Second):
		t.Fatalf("expect background handle readiness after initial sync")
	}
	if handle.Err() == nil {
		t.Fatalf("expect initial sync failure to be observable on handle")
	}
}

func completeLocalToFTPOptionsForBackground() Options {
	return Options{
		Direction: DirectionLocalToFTP,
		Source:    Endpoint{LocalPath: "/data/source"},
		Destination: Endpoint{FTP: FTPOptions{
			Host:         "ftp.example.test",
			Port:         21,
			Username:     "ftp-user",
			Password:     "secret-password",
			RemotePath:   "/incoming",
			PassiveMode:  true,
			Timeout:      15 * time.Second,
			PathEncoding: "utf-8",
		}},
	}
}
