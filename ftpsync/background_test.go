package ftpsync

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestBackgroundHandleWait(t *testing.T) {
	defer withSyncOnceExecutor(runSyncOnceScaffold)()
	runSyncOnce = func(ctx context.Context, svc *FTPSyncService, adapter syncOnceAdapter, result *Result) error {
		return nil
	}

	svc, err := NewFTPSyncService(completeLocalToFTPOptionsForBackgroundRoot(t.TempDir()))
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

	svc, err := NewFTPSyncService(completeLocalToFTPOptionsForBackgroundRoot(t.TempDir()))
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

func TestBackgroundDebouncesBurstEvents(t *testing.T) {
	defer withSyncOnceExecutor(runSyncOnceScaffold)()
	var syncPasses int32
	runSyncOnce = func(ctx context.Context, svc *FTPSyncService, adapter syncOnceAdapter, result *Result) error {
		atomic.AddInt32(&syncPasses, 1)
		return nil
	}

	sourceRoot := t.TempDir()
	svc, err := NewFTPSyncService(completeLocalToFTPOptionsForBackgroundRoot(sourceRoot))
	if err != nil {
		t.Fatalf("construct service: %v", err)
	}
	handle, err := svc.StartBackground(context.Background())
	if err != nil {
		t.Fatalf("StartBackground returned error: %v", err)
	}
	defer func() { _ = handle.Stop(context.Background()) }()
	waitReady(t, handle)
	waitForSyncPasses(t, &syncPasses, 1)

	for i := 0; i < 5; i++ {
		if err := os.WriteFile(filepath.Join(sourceRoot, "burst.txt"), []byte{byte(i)}, 0o644); err != nil {
			t.Fatalf("write burst file: %v", err)
		}
	}
	waitForSyncPasses(t, &syncPasses, 2)
	time.Sleep(backgroundDebounceDelay * 3)
	if got := atomic.LoadInt32(&syncPasses); got != 2 {
		t.Fatalf("expected burst to coalesce into one follow-up sync pass, got %d", got)
	}
}

func TestBackgroundWatchesNewDirectories(t *testing.T) {
	defer withSyncOnceExecutor(runSyncOnceScaffold)()
	var syncPasses int32
	runSyncOnce = func(ctx context.Context, svc *FTPSyncService, adapter syncOnceAdapter, result *Result) error {
		atomic.AddInt32(&syncPasses, 1)
		return nil
	}

	sourceRoot := t.TempDir()
	svc, err := NewFTPSyncService(completeLocalToFTPOptionsForBackgroundRoot(sourceRoot))
	if err != nil {
		t.Fatalf("construct service: %v", err)
	}
	handle, err := svc.StartBackground(context.Background())
	if err != nil {
		t.Fatalf("StartBackground returned error: %v", err)
	}
	defer func() { _ = handle.Stop(context.Background()) }()
	waitReady(t, handle)
	waitForSyncPasses(t, &syncPasses, 1)

	nested := filepath.Join(sourceRoot, "nested")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("create nested directory: %v", err)
	}
	waitForSyncPasses(t, &syncPasses, 2)
	if err := os.WriteFile(filepath.Join(nested, "child.txt"), []byte("child"), 0o644); err != nil {
		t.Fatalf("write nested child: %v", err)
	}
	waitForSyncPasses(t, &syncPasses, 3)
}

func TestBackgroundSyncFailureIsObservableAndNonTerminal(t *testing.T) {
	defer withSyncOnceExecutor(runSyncOnceScaffold)()
	var syncPasses int32
	runSyncOnce = func(ctx context.Context, svc *FTPSyncService, adapter syncOnceAdapter, result *Result) error {
		pass := atomic.AddInt32(&syncPasses, 1)
		if pass == 2 {
			return newError(ErrTransfer, "simulated steady-state failure", errors.New("upload failed"))
		}
		return nil
	}

	sourceRoot := t.TempDir()
	svc, err := NewFTPSyncService(completeLocalToFTPOptionsForBackgroundRoot(sourceRoot))
	if err != nil {
		t.Fatalf("construct service: %v", err)
	}
	handle, err := svc.StartBackground(context.Background())
	if err != nil {
		t.Fatalf("StartBackground returned error: %v", err)
	}
	defer func() { _ = handle.Stop(context.Background()) }()
	waitReady(t, handle)
	waitForSyncPasses(t, &syncPasses, 1)

	if err := os.WriteFile(filepath.Join(sourceRoot, "fail-once.txt"), []byte("fail"), 0o644); err != nil {
		t.Fatalf("write failure trigger: %v", err)
	}
	waitForSyncPasses(t, &syncPasses, 2)
	if err := handle.Err(); err == nil || !IsKind(err, ErrTransfer) {
		t.Fatalf("expected steady-state transfer failure to be observable, got %v", err)
	}
	select {
	case <-handle.Done():
		t.Fatalf("background run should not terminate after one sync failure")
	default:
	}

	if err := os.WriteFile(filepath.Join(sourceRoot, "after-failure.txt"), []byte("after"), 0o644); err != nil {
		t.Fatalf("write after failure trigger: %v", err)
	}
	waitForSyncPasses(t, &syncPasses, 3)
	select {
	case <-handle.Done():
		t.Fatalf("background run should stay alive after processing later sync pass")
	default:
	}
}

func TestBackgroundRunsFollowUpSyncForEventsDuringActiveSync(t *testing.T) {
	defer withSyncOnceExecutor(runSyncOnceScaffold)()
	var syncPasses int32
	secondStarted := make(chan struct{})
	releaseSecond := make(chan struct{})
	runSyncOnce = func(ctx context.Context, svc *FTPSyncService, adapter syncOnceAdapter, result *Result) error {
		pass := atomic.AddInt32(&syncPasses, 1)
		if pass == 2 {
			close(secondStarted)
			<-releaseSecond
		}
		return nil
	}

	sourceRoot := t.TempDir()
	svc, err := NewFTPSyncService(completeLocalToFTPOptionsForBackgroundRoot(sourceRoot))
	if err != nil {
		t.Fatalf("construct service: %v", err)
	}
	handle, err := svc.StartBackground(context.Background())
	if err != nil {
		t.Fatalf("StartBackground returned error: %v", err)
	}
	defer func() { _ = handle.Stop(context.Background()) }()
	waitReady(t, handle)
	waitForSyncPasses(t, &syncPasses, 1)

	if err := os.WriteFile(filepath.Join(sourceRoot, "during-sync-1.txt"), []byte("first"), 0o644); err != nil {
		t.Fatalf("write first trigger: %v", err)
	}
	select {
	case <-secondStarted:
	case <-time.After(time.Second):
		t.Fatalf("expect second sync to start")
	}
	if err := os.WriteFile(filepath.Join(sourceRoot, "during-sync-2.txt"), []byte("second"), 0o644); err != nil {
		t.Fatalf("write trigger during active sync: %v", err)
	}
	time.Sleep(backgroundDebounceDelay * 2)
	close(releaseSecond)
	waitForSyncPasses(t, &syncPasses, 3)
}

func TestBackgroundStopShutsDownDeterministically(t *testing.T) {
	defer withSyncOnceExecutor(runSyncOnceScaffold)()
	var syncPasses int32
	secondStarted := make(chan struct{})
	releaseSecond := make(chan struct{})
	runSyncOnce = func(ctx context.Context, svc *FTPSyncService, adapter syncOnceAdapter, result *Result) error {
		pass := atomic.AddInt32(&syncPasses, 1)
		if pass == 2 {
			close(secondStarted)
			<-ctx.Done()
			<-releaseSecond
		}
		return nil
	}

	sourceRoot := t.TempDir()
	svc, err := NewFTPSyncService(completeLocalToFTPOptionsForBackgroundRoot(sourceRoot))
	if err != nil {
		t.Fatalf("construct service: %v", err)
	}
	handle, err := svc.StartBackground(context.Background())
	if err != nil {
		t.Fatalf("StartBackground returned error: %v", err)
	}
	waitReady(t, handle)
	waitForSyncPasses(t, &syncPasses, 1)
	if err := os.WriteFile(filepath.Join(sourceRoot, "stop-during-sync.txt"), []byte("stop"), 0o644); err != nil {
		t.Fatalf("write stop trigger: %v", err)
	}
	select {
	case <-secondStarted:
	case <-time.After(time.Second):
		t.Fatalf("expect second sync to start")
	}
	stopReturned := make(chan error, 1)
	go func() {
		stopReturned <- handle.Stop(context.Background())
	}()
	select {
	case err := <-stopReturned:
		t.Fatalf("Stop returned before in-flight sync worker exited: %v", err)
	case <-time.After(100 * time.Millisecond):
	}
	close(releaseSecond)
	select {
	case err := <-stopReturned:
		if err != nil {
			t.Fatalf("Stop returned error: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatalf("expect Stop to return after worker exits")
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

func TestBackgroundContextCancelStopsRunner(t *testing.T) {
	defer withSyncOnceExecutor(runSyncOnceScaffold)()
	runSyncOnce = func(ctx context.Context, svc *FTPSyncService, adapter syncOnceAdapter, result *Result) error {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	svc, err := NewFTPSyncService(completeLocalToFTPOptionsForBackgroundRoot(t.TempDir()))
	if err != nil {
		t.Fatalf("construct service: %v", err)
	}
	handle, err := svc.StartBackground(ctx)
	if err != nil {
		t.Fatalf("StartBackground returned error: %v", err)
	}
	waitReady(t, handle)
	cancel()
	if err := waitBackground(handle, time.Second); err != nil {
		t.Fatalf("Wait returned error after cancel: %v", err)
	}
}

func TestBackgroundStopAndCancelRaceIdempotent(t *testing.T) {
	defer withSyncOnceExecutor(runSyncOnceScaffold)()
	runSyncOnce = func(ctx context.Context, svc *FTPSyncService, adapter syncOnceAdapter, result *Result) error {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	svc, err := NewFTPSyncService(completeLocalToFTPOptionsForBackgroundRoot(t.TempDir()))
	if err != nil {
		t.Fatalf("construct service: %v", err)
	}
	handle, err := svc.StartBackground(ctx)
	if err != nil {
		t.Fatalf("StartBackground returned error: %v", err)
	}
	waitReady(t, handle)

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = handle.Stop(context.Background())
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		cancel()
	}()
	wg.Wait()
	if err := waitBackground(handle, time.Second); err != nil {
		t.Fatalf("Wait returned error after stop/cancel race: %v", err)
	}
}

func TestBackgroundWaitReturnsFinalError(t *testing.T) {
	defer withSyncOnceExecutor(runSyncOnceScaffold)()
	var syncPasses int32
	runSyncOnce = func(ctx context.Context, svc *FTPSyncService, adapter syncOnceAdapter, result *Result) error {
		pass := atomic.AddInt32(&syncPasses, 1)
		if pass == 2 {
			return newError(ErrTransfer, "simulated steady-state failure", errors.New("upload failed"))
		}
		return nil
	}

	sourceRoot := t.TempDir()
	svc, err := NewFTPSyncService(completeLocalToFTPOptionsForBackgroundRoot(sourceRoot))
	if err != nil {
		t.Fatalf("construct service: %v", err)
	}
	handle, err := svc.StartBackground(context.Background())
	if err != nil {
		t.Fatalf("StartBackground returned error: %v", err)
	}
	waitReady(t, handle)
	waitForSyncPasses(t, &syncPasses, 1)
	if err := os.WriteFile(filepath.Join(sourceRoot, "fail-before-stop.txt"), []byte("fail"), 0o644); err != nil {
		t.Fatalf("write failure trigger: %v", err)
	}
	waitForSyncPasses(t, &syncPasses, 2)
	if err := handle.Err(); err == nil || !IsKind(err, ErrTransfer) {
		t.Fatalf("expected latest runtime transfer failure, got %v", err)
	}
	if err := handle.Stop(context.Background()); err != nil {
		t.Fatalf("Stop returned final error instead of clean shutdown: %v", err)
	}
	if err := handle.Wait(); err != nil {
		t.Fatalf("Wait returned latest runtime error instead of final result: %v", err)
	}
}

func completeLocalToFTPOptionsForBackground() Options {
	return completeLocalToFTPOptionsForBackgroundRoot("/data/source")
}

func completeLocalToFTPOptionsForBackgroundRoot(sourceRoot string) Options {
	return Options{
		Direction: DirectionLocalToFTP,
		Source:    Endpoint{LocalPath: sourceRoot},
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

func waitReady(t *testing.T, handle Handle) {
	t.Helper()
	readyHandle, ok := handle.(interface{ Ready() <-chan struct{} })
	if !ok {
		t.Fatalf("background handle should expose internal readiness for tests")
	}
	select {
	case <-readyHandle.Ready():
	case <-time.After(time.Second):
		t.Fatalf("expect background handle readiness")
	}
}

func waitForSyncPasses(t *testing.T, syncPasses *int32, want int32) {
	t.Helper()
	deadline := time.After(2 * time.Second)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-deadline:
			t.Fatalf("expected at least %d sync passes, got %d", want, atomic.LoadInt32(syncPasses))
		case <-ticker.C:
			if atomic.LoadInt32(syncPasses) >= want {
				return
			}
		}
	}
}

func waitBackground(handle Handle, timeout time.Duration) error {
	done := make(chan error, 1)
	go func() {
		done <- handle.Wait()
	}()
	select {
	case err := <-done:
		return err
	case <-time.After(timeout):
		return errors.New("background wait timed out")
	}
}
