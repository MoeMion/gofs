package ftpsync

import (
	"context"
	"errors"
	"io/fs"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/no-src/gofs/core"
	legacysync "github.com/no-src/gofs/sync"
)

func TestSyncOnceResultReturnsCompactSummary(t *testing.T) {
	defer withSyncOnceExecutor(runSyncOnceScaffold)()
	runSyncOnce = func(ctx context.Context, svc *FTPSyncService, adapter syncOnceAdapter, result *Result) error {
		result.PathsVisited = 3
		result.FilesAttempted = 2
		result.DirectoriesAttempted = 1
		svc.reportProgress(Progress{Path: "alpha.txt", FilesTransferred: 2, FilesTotal: 2})
		return nil
	}

	svc, err := NewFTPSyncService(completeLocalToFTPOptionsForOneShot())
	if err != nil {
		t.Fatalf("construct service: %v", err)
	}

	result, err := svc.SyncOnce(context.Background())
	if err != nil {
		t.Fatalf("SyncOnce returned error: %v", err)
	}
	if result.Direction != DirectionLocalToFTP {
		t.Fatalf("expect direction in result, got %#v", result)
	}
	if result.SourceRoot != "/data/source" || result.DestinationRoot != "/incoming" {
		t.Fatalf("expect compact roots in result, got %#v", result)
	}
	if result.PathsVisited != 3 || result.FilesAttempted != 2 || result.DirectoriesAttempted != 1 {
		t.Fatalf("expect compact summary counts, got %#v", result)
	}
	if result.Partial {
		t.Fatalf("expect successful summary to be non-partial")
	}
	if result.StartedAt.IsZero() || result.CompletedAt.IsZero() || result.CompletedAt.Before(result.StartedAt) {
		t.Fatalf("expect timestamps to be populated, got %#v", result)
	}
}

func TestSyncOncePartialFailureReturnsResultAndTypedError(t *testing.T) {
	defer withSyncOnceExecutor(runSyncOnceScaffold)()
	runSyncOnce = func(ctx context.Context, svc *FTPSyncService, adapter syncOnceAdapter, result *Result) error {
		result.PathsVisited = 4
		result.FilesAttempted = 3
		result.DirectoriesAttempted = 1
		result.FailureCount = 1
		result.Partial = true
		return errors.New("write failed for partial.txt")
	}

	opts := completeLocalToFTPOptionsForOneShot()
	const secret = "partial-secret-password"
	opts.Destination.FTP.Password = secret
	svc, err := NewFTPSyncService(opts)
	if err != nil {
		t.Fatalf("construct service: %v", err)
	}

	result, err := svc.SyncOnce(context.Background())
	if err == nil {
		t.Fatalf("expect partial failure error")
	}
	if !IsKind(err, ErrTransfer) {
		t.Fatalf("expect transfer error kind, got %v", err)
	}
	if !result.Partial || result.FailureCount != 1 || result.PathsVisited != 4 {
		t.Fatalf("expect partial summary to be preserved, got %#v", result)
	}
	if strings.Contains(err.Error(), secret) {
		t.Fatalf("partial failure error leaked FTP password")
	}
}

func TestSyncOnceHooksReceiveCompactSignals(t *testing.T) {
	defer withSyncOnceExecutor(runSyncOnceScaffold)()
	var logs []string
	var progress []Progress
	var events []SyncEvent
	runSyncOnce = func(ctx context.Context, svc *FTPSyncService, adapter syncOnceAdapter, result *Result) error {
		result.PathsVisited = 1
		result.FilesAttempted = 1
		svc.log("custom oneshot log")
		svc.reportProgress(Progress{Path: "alpha.txt", FilesTransferred: 1, FilesTotal: 1})
		svc.reportEvent(SyncEvent{Operation: "sync_once", Path: "alpha.txt", Status: "complete"})
		return nil
	}

	opts := completeLocalToFTPOptionsForOneShot()
	opts.Hooks = HookSet{
		Logger: LoggerFunc(func(message string) { logs = append(logs, message) }),
		Progress: func(snapshot Progress) {
			progress = append(progress, snapshot)
		},
		Event: func(event SyncEvent) {
			events = append(events, event)
		},
	}
	svc, err := NewFTPSyncService(opts)
	if err != nil {
		t.Fatalf("construct service: %v", err)
	}

	_, err = svc.SyncOnce(context.Background())
	if err != nil {
		t.Fatalf("SyncOnce returned error: %v", err)
	}
	if len(logs) == 0 || len(progress) == 0 || len(events) == 0 {
		t.Fatalf("expect hooks to receive compact signals, logs=%d progress=%d events=%d", len(logs), len(progress), len(events))
	}
}

func TestSyncOnceResultRemainsCompact(t *testing.T) {
	resultType := reflect.TypeOf(Result{})
	for i := 0; i < resultType.NumField(); i++ {
		field := resultType.Field(i)
		if field.Type.Kind() == reflect.Slice {
			t.Fatalf("result should stay compact without per-file slices, found %s", field.Name)
		}
	}
}

func TestSyncOnceBuildsLegacyAdapterLocalToFTP(t *testing.T) {
	adapter, err := newSyncOnceAdapter(completeLocalToFTPOptionsForOneShot())
	if err != nil {
		t.Fatalf("newSyncOnceAdapter: %v", err)
	}
	if !adapter.option.Source.IsDisk() {
		t.Fatalf("expect local source disk VFS")
	}
	if !adapter.option.Dest.Is(core.FTP) {
		t.Fatalf("expect FTP destination VFS")
	}
	assertFTPVFS(t, adapter.option.Dest, completeLocalToFTPOptionsForOneShot().Destination.FTP)
	if adapter.option.Retry.Count() != completeLocalToFTPOptionsForOneShot().Retry.Count {
		t.Fatalf("expect retry count to round-trip")
	}
	if !adapter.option.PathIgnore.MatchPath("tmp/cache", "test", "literal") {
		t.Fatalf("expect typed ignore adapter to match literal rule")
	}
	if !adapter.option.PathIgnore.MatchPath("file.part", "test", "glob") {
		t.Fatalf("expect typed ignore adapter to match glob rule")
	}
	if !adapter.option.PathIgnore.MatchPath("debug/output.log", "test", "regexp") {
		t.Fatalf("expect typed ignore adapter to match regexp rule")
	}
}

func TestSyncOnceBuildsLegacyAdapterFTPToLocal(t *testing.T) {
	adapter, err := newSyncOnceAdapter(completeFTPToLocalOptionsForOneShot())
	if err != nil {
		t.Fatalf("newSyncOnceAdapter: %v", err)
	}
	if !adapter.option.Source.Is(core.FTP) {
		t.Fatalf("expect FTP source VFS")
	}
	if !adapter.option.Dest.IsDisk() {
		t.Fatalf("expect local destination disk VFS")
	}
	assertFTPVFS(t, adapter.option.Source, completeFTPToLocalOptionsForOneShot().Source.FTP)
	if adapter.option.Dest.Path().Base() != "/data/destination" {
		t.Fatalf("expect local destination root to round-trip, got %q", adapter.option.Dest.Path().Base())
	}
}

func TestNewFTPSyncServiceSyncOnceBuildsLegacyAdapterViaSeam(t *testing.T) {
	defer withSyncBuilder(buildLegacySync)()
	defer withSyncOnceExecutor(runSyncOnceScaffold)()
	var captured legacysync.Option
	buildLegacySync = func(opt legacysync.Option) (legacysync.Sync, error) {
		captured = opt
		return noopLegacySync{source: opt.Source, dest: opt.Dest}, nil
	}
	runSyncOnce = func(ctx context.Context, svc *FTPSyncService, adapter syncOnceAdapter, result *Result) error {
		syncer, err := adapter.newSync()
		if err != nil {
			return err
		}
		defer syncer.Close()
		result.PathsVisited = 1
		return nil
	}

	svc, err := NewFTPSyncService(completeLocalToFTPOptionsForOneShot())
	if err != nil {
		t.Fatalf("construct service: %v", err)
	}

	_, err = svc.SyncOnce(context.Background())
	if err != nil {
		t.Fatalf("SyncOnce returned error: %v", err)
	}
	if !captured.Source.IsDisk() || !captured.Dest.Is(core.FTP) {
		t.Fatalf("expect SyncOnce seam to build legacy local→FTP adapter")
	}
}

func TestSyncOnceLocalToFTP(t *testing.T) {
	defer withSyncBuilder(buildLegacySync)()
	tempDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tempDir, "nested"), 0o755); err != nil {
		t.Fatalf("mkdir nested: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "nested", "child.txt"), []byte("child"), 0o644); err != nil {
		t.Fatalf("write child file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "root.txt"), []byte("root"), 0o644); err != nil {
		t.Fatalf("write root file: %v", err)
	}
	if runtime.GOOS != "windows" {
		if err := os.Symlink("root.txt", filepath.Join(tempDir, "root-link")); err != nil {
			t.Fatalf("create symlink: %v", err)
		}
	}

	var captured *recordingLegacySync
	buildLegacySync = func(opt legacysync.Option) (legacysync.Sync, error) {
		captured = &recordingLegacySync{source: opt.Source, dest: opt.Dest}
		return captured, nil
	}

	opts := completeLocalToFTPOptionsForOneShot()
	opts.Source.LocalPath = tempDir
	svc, err := NewFTPSyncService(opts)
	if err != nil {
		t.Fatalf("construct service: %v", err)
	}

	result, err := svc.SyncOnce(context.Background())
	if err != nil {
		t.Fatalf("SyncOnce returned error: %v", err)
	}
	if captured == nil {
		t.Fatalf("expected legacy sync to be constructed")
	}
	if len(captured.creates) < 3 {
		t.Fatalf("expected create operations for root, nested dir, and files, got %#v", captured.creates)
	}
	if len(captured.writes) != 2 {
		t.Fatalf("expected writes for two regular files, got %#v", captured.writes)
	}
	if runtime.GOOS != "windows" && len(captured.symlinks) != 1 {
		t.Fatalf("expected one symlink operation, got %#v", captured.symlinks)
	}
	if result.PathsVisited < 4 {
		t.Fatalf("expected visited paths to reflect walked tree, got %#v", result)
	}
	if result.FilesAttempted < 2 || result.DirectoriesAttempted < 2 {
		t.Fatalf("expected compact counts to be updated, got %#v", result)
	}
	if result.Partial || result.FailureCount != 0 {
		t.Fatalf("expected successful run, got %#v", result)
	}
	if !captured.closed {
		t.Fatalf("expected legacy sync to be closed")
	}
	for _, created := range captured.creates {
		if !filepath.IsAbs(created) {
			t.Fatalf("expected absolute create path, got %q", created)
		}
	}
	for _, written := range captured.writes {
		if !filepath.IsAbs(written) {
			t.Fatalf("expected absolute write path, got %q", written)
		}
	}
	if runtime.GOOS != "windows" {
		if got := filepath.Base(captured.symlinks[0].oldname); got != "root.txt" {
			t.Fatalf("expected symlink target root.txt, got %q", captured.symlinks[0].oldname)
		}
	}
}

func TestSyncOnceLocalToFTPPartialFailure(t *testing.T) {
	defer withSyncBuilder(buildLegacySync)()
	tempDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tempDir, "nested"), 0o755); err != nil {
		t.Fatalf("mkdir nested: %v", err)
	}
	failingFile := filepath.Join(tempDir, "nested", "fail.txt")
	if err := os.WriteFile(failingFile, []byte("fail"), 0o644); err != nil {
		t.Fatalf("write fail file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "after.txt"), []byte("after"), 0o644); err != nil {
		t.Fatalf("write after file: %v", err)
	}

	var captured *recordingLegacySync
	buildLegacySync = func(opt legacysync.Option) (legacysync.Sync, error) {
		captured = &recordingLegacySync{
			source:        opt.Source,
			dest:          opt.Dest,
			failWritePath: failingFile,
		}
		return captured, nil
	}

	var events []SyncEvent
	var progress []Progress
	opts := completeLocalToFTPOptionsForOneShot()
	opts.Source.LocalPath = tempDir
	opts.Hooks = HookSet{
		Logger: noopLogger{},
		Progress: func(snapshot Progress) {
			progress = append(progress, snapshot)
		},
		Event: func(event SyncEvent) {
			events = append(events, event)
		},
	}
	svc, err := NewFTPSyncService(opts)
	if err != nil {
		t.Fatalf("construct service: %v", err)
	}

	result, err := svc.SyncOnce(context.Background())
	if err == nil {
		t.Fatalf("expected partial failure error")
	}
	if !IsKind(err, ErrTransfer) {
		t.Fatalf("expected transfer error kind, got %v", err)
	}
	if !result.Partial || result.FailureCount == 0 {
		t.Fatalf("expected partial result counts, got %#v", result)
	}
	if result.PathsVisited < 3 {
		t.Fatalf("expected walker to continue after failure, got %#v", result)
	}
	if captured == nil || !containsString(captured.writes, filepath.Join(tempDir, "after.txt")) {
		t.Fatalf("expected walker to continue after failure, writes=%#v", captured.writes)
	}
	if len(progress) == 0 {
		t.Fatalf("expected progress callbacks")
	}
	if !hasFailedEvent(events, failingFile) {
		t.Fatalf("expected failed sync event for %q, events=%#v", failingFile, events)
	}
	if strings.Contains(err.Error(), opts.Destination.FTP.Password) {
		t.Fatalf("expected partial failure error to hide password")
	}
	if !captured.closed {
		t.Fatalf("expected legacy sync to be closed")
	}
}

func TestSyncOnceFTPToLocalAutoCreateRoot(t *testing.T) {
	defer withSyncBuilder(buildLegacySync)()
	destinationRoot := filepath.Join(t.TempDir(), "created-root")

	var captured *recordingLegacySync
	buildLegacySync = func(opt legacysync.Option) (legacysync.Sync, error) {
		captured = &recordingLegacySync{source: opt.Source, dest: opt.Dest}
		return captured, nil
	}

	remoteRoot := "/outgoing"
	remoteFile := path.Join(remoteRoot, "nested", "child.txt")
	runFTPToLocalSyncOnceWalk := withSyncOnceExecutor(func(ctx context.Context, svc *FTPSyncService, adapter syncOnceAdapter, result *Result) error {
		return runSyncOnceScaffold(ctx, svc, syncOnceAdapter{option: adapter.option}, result)
	})
	defer runFTPToLocalSyncOnceWalk()

	previousBuilder := buildLegacySync
	buildLegacySync = func(opt legacysync.Option) (legacysync.Sync, error) {
		captured = &recordingLegacySync{source: opt.Source, dest: opt.Dest}
		return &walkingLegacySync{
			recordingLegacySync: captured,
			walk:                []walkEntry{{path: remoteRoot, isDir: true}, {path: path.Join(remoteRoot, "nested"), isDir: true}, {path: remoteFile, isDir: false}},
		}, nil
	}
	defer func() { buildLegacySync = previousBuilder }()

	opts := completeFTPToLocalOptionsForOneShot()
	opts.Destination.LocalPath = destinationRoot
	svc, err := NewFTPSyncService(opts)
	if err != nil {
		t.Fatalf("construct service: %v", err)
	}

	result, err := svc.SyncOnce(context.Background())
	if err != nil {
		t.Fatalf("SyncOnce returned error: %v", err)
	}
	if _, err := os.Stat(destinationRoot); err != nil {
		t.Fatalf("expected destination root to be auto-created: %v", err)
	}
	if captured == nil {
		t.Fatalf("expected legacy sync to be constructed")
	}
	if !containsString(captured.creates, remoteRoot) || !containsString(captured.creates, path.Join(remoteRoot, "nested")) {
		t.Fatalf("expected directory create operations for remote tree, got %#v", captured.creates)
	}
	if !containsString(captured.writes, remoteFile) {
		t.Fatalf("expected file write operation for remote file, got %#v", captured.writes)
	}
	if result.PathsVisited != 3 || result.FilesAttempted != 1 || result.DirectoriesAttempted != 2 {
		t.Fatalf("expected compact FTP→local summary counts, got %#v", result)
	}
	if result.Partial || result.FailureCount != 0 {
		t.Fatalf("expected successful FTP→local run, got %#v", result)
	}
	if !captured.closed {
		t.Fatalf("expected legacy sync to be closed")
	}
}

func TestSyncOnceFTPToLocalSuccess(t *testing.T) {
	defer withSyncBuilder(buildLegacySync)()
	destinationRoot := t.TempDir()
	remoteRoot := "/outgoing"
	remoteFile := path.Join(remoteRoot, "nested", "child.txt")

	var captured *recordingLegacySync
	buildLegacySync = func(opt legacysync.Option) (legacysync.Sync, error) {
		captured = &recordingLegacySync{source: opt.Source, dest: opt.Dest}
		return &walkingLegacySync{
			recordingLegacySync: captured,
			walk:                []walkEntry{{path: remoteRoot, isDir: true}, {path: path.Join(remoteRoot, "nested"), isDir: true}, {path: remoteFile, isDir: false}},
		}, nil
	}

	svc, err := NewFTPSyncService(func() Options {
		opts := completeFTPToLocalOptionsForOneShot()
		opts.Destination.LocalPath = destinationRoot
		return opts
	}())
	if err != nil {
		t.Fatalf("construct service: %v", err)
	}

	result, err := svc.SyncOnce(context.Background())
	if err != nil {
		t.Fatalf("SyncOnce returned error: %v", err)
	}
	if captured == nil {
		t.Fatalf("expected legacy sync to be constructed")
	}
	if !containsString(captured.writes, remoteFile) {
		t.Fatalf("expected write for remote file, got %#v", captured.writes)
	}
	if result.Direction != DirectionFTPToLocal || result.DestinationRoot != destinationRoot {
		t.Fatalf("expected FTP→local result roots, got %#v", result)
	}
}

func TestSyncOnceFTPToLocalNeverWritesToCWD(t *testing.T) {
	defer withSyncBuilder(buildLegacySync)()
	workDir := t.TempDir()
	destinationRoot := filepath.Join(t.TempDir(), "dest-root")
	remoteRoot := "/outgoing"
	remoteFile := path.Join(remoteRoot, "nested", "child.txt")

	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(workDir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(originalWD); err != nil {
			t.Errorf("restore cwd: %v", err)
		}
	})

	buildLegacySync = func(opt legacysync.Option) (legacysync.Sync, error) {
		return &walkingLegacySync{
			recordingLegacySync: &recordingLegacySync{source: opt.Source, dest: opt.Dest},
			walk:                []walkEntry{{path: remoteRoot, isDir: true}, {path: path.Join(remoteRoot, "nested"), isDir: true}, {path: remoteFile, isDir: false}},
		}, nil
	}

	svc, err := NewFTPSyncService(func() Options {
		opts := completeFTPToLocalOptionsForOneShot()
		opts.Destination.LocalPath = destinationRoot
		return opts
	}())
	if err != nil {
		t.Fatalf("construct service: %v", err)
	}

	_, err = svc.SyncOnce(context.Background())
	if err != nil {
		t.Fatalf("SyncOnce returned error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(workDir, "nested", "child.txt")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected no synced file in cwd, stat err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(destinationRoot, "nested", "child.txt")); err != nil {
		t.Fatalf("expected synced file under configured destination root: %v", err)
	}
}

func TestSyncOnceFTPToLocalPartialFailure(t *testing.T) {
	defer withSyncBuilder(buildLegacySync)()
	destinationRoot := t.TempDir()
	remoteRoot := "/outgoing"
	failingFile := path.Join(remoteRoot, "nested", "fail.txt")
	afterFile := path.Join(remoteRoot, "after.txt")

	buildLegacySync = func(opt legacysync.Option) (legacysync.Sync, error) {
		return &walkingLegacySync{
			recordingLegacySync: &recordingLegacySync{source: opt.Source, dest: opt.Dest, failWritePath: failingFile},
			walk:                []walkEntry{{path: remoteRoot, isDir: true}, {path: path.Join(remoteRoot, "nested"), isDir: true}, {path: failingFile, isDir: false}, {path: afterFile, isDir: false}},
		}, nil
	}

	svc, err := NewFTPSyncService(func() Options {
		opts := completeFTPToLocalOptionsForOneShot()
		opts.Destination.LocalPath = destinationRoot
		return opts
	}())
	if err != nil {
		t.Fatalf("construct service: %v", err)
	}

	result, err := svc.SyncOnce(context.Background())
	if err == nil {
		t.Fatalf("expected partial failure error")
	}
	if !IsKind(err, ErrTransfer) {
		t.Fatalf("expected transfer error kind, got %v", err)
	}
	if !result.Partial || result.FailureCount == 0 || result.PathsVisited < 3 || result.FilesAttempted < 1 || result.DirectoriesAttempted < 2 {
		t.Fatalf("expected partial FTP→local summary counts, got %#v", result)
	}
	if _, err := os.Stat(filepath.Join(destinationRoot, "after.txt")); err != nil {
		t.Fatalf("expected later file to still be written after failure: %v", err)
	}
}

func assertFTPVFS(t *testing.T, vfs core.VFS, expected FTPOptions) {
	t.Helper()
	if vfs.Host() != expected.Host || vfs.Port() != expected.Port {
		t.Fatalf("unexpected FTP host/port => %s:%d", vfs.Host(), vfs.Port())
	}
	if vfs.RemotePath().Base() != expected.RemotePath {
		t.Fatalf("unexpected remote path => %q", vfs.RemotePath().Base())
	}
	if vfs.FTPUsername() != expected.Username || vfs.FTPPassword() != expected.Password {
		t.Fatalf("unexpected FTP credentials in VFS")
	}
	if vfs.FTPPassiveMode() != expected.PassiveMode {
		t.Fatalf("unexpected passive mode => %t", vfs.FTPPassiveMode())
	}
	if expected.Timeout > 0 && vfs.FTPTimeout() != expected.Timeout.String() {
		t.Fatalf("unexpected FTP timeout => %q", vfs.FTPTimeout())
	}
	if expected.PathEncoding != "" && vfs.FTPEncoding() != expected.PathEncoding {
		t.Fatalf("unexpected FTP encoding => %q", vfs.FTPEncoding())
	}

	parsed, err := url.Parse(vfs.Addr())
	if err == nil && parsed != nil {
		_ = parsed
	}
}

func completeLocalToFTPOptionsForOneShot() Options {
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
		Retry: RetryOptions{Count: 3, Wait: time.Second, Async: true},
		IgnoreRules: []IgnoreRule{
			{Kind: IgnoreKindLiteral, Pattern: "tmp/cache"},
			{Kind: IgnoreKindGlob, Pattern: "*.part"},
			{Kind: IgnoreKindRegexp, Pattern: `^debug/.*`},
		},
	}
}

func completeFTPToLocalOptionsForOneShot() Options {
	return Options{
		Direction: DirectionFTPToLocal,
		Source: Endpoint{FTP: FTPOptions{
			Host:         "ftp.example.test",
			Port:         21,
			Username:     "ftp-user",
			Password:     "secret-password",
			RemotePath:   "/outgoing",
			PassiveMode:  false,
			Timeout:      20 * time.Second,
			PathEncoding: "gbk",
		}},
		Destination: Endpoint{LocalPath: "/data/destination"},
		Retry:       RetryOptions{Count: 2, Wait: time.Second},
	}
}

func withSyncOnceExecutor(executor syncOnceExecutor) func() {
	previous := runSyncOnce
	runSyncOnce = executor
	return func() {
		runSyncOnce = previous
	}
}

func withSyncBuilder(builder syncBuilder) func() {
	previous := buildLegacySync
	buildLegacySync = builder
	return func() {
		buildLegacySync = previous
	}
}

type noopLegacySync struct {
	source core.VFS
	dest   core.VFS
}

func (n noopLegacySync) Create(path string) error              { return nil }
func (n noopLegacySync) Symlink(oldname, newname string) error { return nil }
func (n noopLegacySync) Write(path string) error               { return nil }
func (n noopLegacySync) Remove(path string) error              { return nil }
func (n noopLegacySync) Rename(path string) error              { return nil }
func (n noopLegacySync) Chmod(path string) error               { return nil }
func (n noopLegacySync) IsDir(path string) (bool, error)       { return false, nil }
func (n noopLegacySync) SyncOnce(path string) error            { return nil }
func (n noopLegacySync) Source() core.VFS                      { return n.source }
func (n noopLegacySync) Dest() core.VFS                        { return n.dest }
func (n noopLegacySync) Close()                                {}

type recordingLegacySync struct {
	source        core.VFS
	dest          core.VFS
	creates       []string
	writes        []string
	symlinks      []recordedSymlink
	failWritePath string
	closed        bool
}

type recordedSymlink struct {
	oldname string
	newname string
}

type walkingLegacySync struct {
	*recordingLegacySync
	walk []walkEntry
}

type walkEntry struct {
	path  string
	isDir bool
}

func (r *recordingLegacySync) Create(path string) error {
	r.creates = append(r.creates, path)
	return nil
}

func (r *recordingLegacySync) Symlink(oldname, newname string) error {
	r.symlinks = append(r.symlinks, recordedSymlink{oldname: oldname, newname: newname})
	return nil
}

func (r *recordingLegacySync) Write(path string) error {
	r.writes = append(r.writes, path)
	if path == r.failWritePath {
		return errors.New("simulated write failure")
	}
	return nil
}

func (r *recordingLegacySync) Remove(path string) error        { return nil }
func (r *recordingLegacySync) Rename(path string) error        { return nil }
func (r *recordingLegacySync) Chmod(path string) error         { return nil }
func (r *recordingLegacySync) IsDir(path string) (bool, error) { return false, nil }
func (r *recordingLegacySync) SyncOnce(path string) error      { return nil }
func (r *recordingLegacySync) Source() core.VFS                { return r.source }
func (r *recordingLegacySync) Dest() core.VFS                  { return r.dest }
func (r *recordingLegacySync) Close()                          { r.closed = true }

func (r *recordingLegacySync) WalkSourceDir(root string, fn fs.WalkDirFunc) error {
	return nil
}

func (r *recordingLegacySync) ReadSourceLink(path string) (string, error) {
	return "", nil
}

func (w *walkingLegacySync) Create(path string) error {
	if err := w.recordingLegacySync.Create(path); err != nil {
		return err
	}
	isDir, _ := w.IsDir(path)
	return w.materialize(path, isDir)
}

func (w *walkingLegacySync) Write(path string) error {
	if err := w.recordingLegacySync.Write(path); err != nil {
		return err
	}
	return w.materialize(path, false)
}

func (w *walkingLegacySync) IsDir(path string) (bool, error) {
	for _, entry := range w.walk {
		if entry.path == path {
			return entry.isDir, nil
		}
	}
	return false, nil
}

func (w *walkingLegacySync) SyncOnce(root string) error {
	for _, entry := range w.walk {
		if entry.isDir {
			if err := w.Create(entry.path); err != nil {
				return err
			}
			continue
		}
		if err := w.Create(entry.path); err != nil {
			return err
		}
		if err := w.Write(entry.path); err != nil {
			return err
		}
	}
	return nil
}

func (w *walkingLegacySync) WalkSourceDir(root string, fn fs.WalkDirFunc) error {
	for _, entry := range w.walk {
		if err := fn(entry.path, stubDirEntry{entry: entry}, nil); err != nil {
			return err
		}
	}
	return nil
}

func (w *walkingLegacySync) ReadSourceLink(path string) (string, error) {
	return "", nil
}

func (w *walkingLegacySync) materialize(remotePath string, isDir bool) error {
	if w == nil || w.dest.IsEmpty() {
		return nil
	}
	relative := strings.TrimPrefix(path.Clean(remotePath), path.Clean(w.source.RemotePath().Base()))
	relative = strings.TrimPrefix(relative, "/")
	target := filepath.Join(w.dest.Path().Base(), filepath.FromSlash(relative))
	if isDir {
		return os.MkdirAll(target, 0o755)
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	return os.WriteFile(target, []byte("synced"), 0o644)
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func hasFailedEvent(events []SyncEvent, target string) bool {
	for _, event := range events {
		if event.Path == target && event.Status == "failed" && event.ErrorKind == ErrTransfer {
			return true
		}
	}
	return false
}

type stubDirEntry struct {
	entry walkEntry
}

func (s stubDirEntry) Name() string {
	return path.Base(s.entry.path)
}

func (s stubDirEntry) IsDir() bool {
	return s.entry.isDir
}

func (s stubDirEntry) Type() fs.FileMode {
	if s.entry.isDir {
		return fs.ModeDir
	}
	return 0
}

func (s stubDirEntry) Info() (fs.FileInfo, error) {
	return stubFileInfo{entry: s.entry}, nil
}

type stubFileInfo struct {
	entry walkEntry
}

func (s stubFileInfo) Name() string       { return path.Base(s.entry.path) }
func (s stubFileInfo) Size() int64        { return 0 }
func (s stubFileInfo) Mode() fs.FileMode  { return stubDirEntry{entry: s.entry}.Type() }
func (s stubFileInfo) ModTime() time.Time { return time.Time{} }
func (s stubFileInfo) IsDir() bool        { return s.entry.isDir }
func (s stubFileInfo) Sys() any           { return nil }
