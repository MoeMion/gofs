package ftpsync

import (
	"context"
	"errors"
	"net/url"
	"reflect"
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
