package ftpsync

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestSyncOnceResultReturnsCompactSummary(t *testing.T) {
	defer withSyncOnceExecutor(runSyncOnceScaffold)()
	runSyncOnce = func(ctx context.Context, svc *FTPSyncService, result *Result) error {
		result.PathsVisited = 3
		result.FilesAttempted = 2
		result.DirectoriesAttempted = 1
		svc.reportProgress(Progress{Path: "alpha.txt", FilesTransferred: 2, FilesTotal: 2})
		return nil
	}

	opts := completeLocalToFTPOptionsForOneShot()
	opts.Source.LocalPath = t.TempDir()
	svc, err := NewFTPSyncService(opts)
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
	if result.SourceRoot != opts.Source.LocalPath || result.DestinationRoot != "/incoming" {
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
	runSyncOnce = func(ctx context.Context, svc *FTPSyncService, result *Result) error {
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
	runSyncOnce = func(ctx context.Context, svc *FTPSyncService, result *Result) error {
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

func TestSyncOnceBuildsPackageLocalFTPClientOptions(t *testing.T) {
	defer withFTPClientFactory(openFTPClient)()
	var captured FTPOptions
	openFTPClient = func(ctx context.Context, opts FTPOptions) (ftpCore, error) {
		captured = opts
		return newRecordingFTPClient(), nil
	}

	opts := completeLocalToFTPOptionsForOneShot()
	opts.Source.LocalPath = t.TempDir()
	svc, err := NewFTPSyncService(opts)
	if err != nil {
		t.Fatalf("construct service: %v", err)
	}
	_, err = svc.SyncOnce(context.Background())
	if err != nil {
		t.Fatalf("SyncOnce returned error: %v", err)
	}
	if captured.Host != "ftp.example.test" || captured.Port != 21 || captured.Username != "ftp-user" || captured.Password != "secret-password" {
		t.Fatalf("expected typed FTP options to reach package-local client, got %#v", captured)
	}
	if captured.RemotePath != "/incoming" || !captured.PassiveMode || captured.Timeout != 15*time.Second || captured.PathEncoding != "utf-8" {
		t.Fatalf("expected FTP path and compatibility options to round-trip, got %#v", captured)
	}
}

func TestSyncOnceLocalToFTP(t *testing.T) {
	defer withFTPClientFactory(openFTPClient)()
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

	var captured *recordingFTPClient
	openFTPClient = func(ctx context.Context, opts FTPOptions) (ftpCore, error) {
		captured = newRecordingFTPClient()
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
	if captured == nil || !containsFTPWrite(captured.writes, filepath.Join(tempDir, "root.txt")) || !containsFTPWrite(captured.writes, filepath.Join(tempDir, "nested", "child.txt")) {
		t.Fatalf("expected package-local FTP writes, got %#v", captured)
	}
	if result.PathsVisited < 4 || result.FilesAttempted != 2 || result.DirectoriesAttempted < 2 || result.Partial || result.FailureCount != 0 {
		t.Fatalf("expected successful compact counts, got %#v", result)
	}
	if !captured.closed {
		t.Fatalf("expected FTP client to be closed")
	}
}

func TestSyncOnceLocalToFTPRemovesRemoteMissingLocally(t *testing.T) {
	defer withFTPClientFactory(openFTPClient)()
	tempDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tempDir, "keep.txt"), []byte("keep"), 0o644); err != nil {
		t.Fatalf("write keep file: %v", err)
	}

	var captured *recordingFTPClient
	openFTPClient = func(ctx context.Context, opts FTPOptions) (ftpCore, error) {
		captured = &recordingFTPClient{walkEntries: []walkEntry{
			{path: "/incoming", isDir: true},
			{path: "/incoming/keep.txt", isDir: false},
			{path: "/incoming/deleted.txt", isDir: false},
			{path: "/incoming/deleted-dir", isDir: true},
		}}
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
	if !containsString(captured.removes, "/incoming/deleted.txt") || !containsString(captured.removes, "/incoming/deleted-dir") {
		t.Fatalf("expected stale remote paths to be removed, got %#v", captured.removes)
	}
	if containsString(captured.removes, "/incoming/keep.txt") || containsString(captured.removes, "/incoming") {
		t.Fatalf("expected existing file and remote root to be kept, got %#v", captured.removes)
	}
	if result.FailureCount != 0 || result.Partial {
		t.Fatalf("expected successful removal pass, got %#v", result)
	}
}

func TestSyncOnceLocalToFTPNormalizesRemoteRootSeparators(t *testing.T) {
	defer withFTPClientFactory(openFTPClient)()
	tempDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tempDir, "keep.txt"), []byte("keep"), 0o644); err != nil {
		t.Fatalf("write keep file: %v", err)
	}

	var captured *recordingFTPClient
	openFTPClient = func(ctx context.Context, opts FTPOptions) (ftpCore, error) {
		captured = &recordingFTPClient{walkEntries: []walkEntry{
			{path: `/incoming\nested`, isDir: true},
			{path: `/incoming\keep.txt`, isDir: false},
		}}
		return captured, nil
	}

	opts := completeLocalToFTPOptionsForOneShot()
	opts.Source.LocalPath = tempDir
	opts.Destination.FTP.RemotePath = `\incoming\`
	svc, err := NewFTPSyncService(opts)
	if err != nil {
		t.Fatalf("construct service: %v", err)
	}

	_, err = svc.SyncOnce(context.Background())
	if err != nil {
		t.Fatalf("SyncOnce returned error: %v", err)
	}
	if !containsFTPWriteRemote(captured.writes, "/incoming/keep.txt") {
		t.Fatalf("expected slash-style FTP write path, got %#v", captured.writes)
	}
	if containsString(captured.removes, `/incoming\keep.txt`) {
		t.Fatalf("expected remote keep path not to be removed when separators differ, got %#v", captured.removes)
	}
	if !containsString(captured.removes, `/incoming\nested`) {
		t.Fatalf("expected stale normalized remote directory to be removed, got %#v", captured.removes)
	}
}

func TestSyncOnceLocalToFTPPartialFailure(t *testing.T) {
	defer withFTPClientFactory(openFTPClient)()
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

	var captured *recordingFTPClient
	openFTPClient = func(ctx context.Context, opts FTPOptions) (ftpCore, error) {
		captured = newRecordingFTPClient()
		captured.failWritePath = failingFile
		return captured, nil
	}
	var events []SyncEvent
	opts := completeLocalToFTPOptionsForOneShot()
	opts.Source.LocalPath = tempDir
	opts.Hooks.Event = func(event SyncEvent) { events = append(events, event) }
	svc, err := NewFTPSyncService(opts)
	if err != nil {
		t.Fatalf("construct service: %v", err)
	}

	result, err := svc.SyncOnce(context.Background())
	if err == nil || !IsKind(err, ErrTransfer) {
		t.Fatalf("expected transfer error, got %v", err)
	}
	if !result.Partial || result.FailureCount == 0 || !containsFTPWrite(captured.writes, filepath.Join(tempDir, "after.txt")) {
		t.Fatalf("expected partial run to continue, result=%#v writes=%#v", result, captured.writes)
	}
	if !hasFailedEvent(events, failingFile) {
		t.Fatalf("expected failed event for %q, events=%#v", failingFile, events)
	}
}

func TestSyncOnceFTPToLocalAutoCreateRoot(t *testing.T) {
	defer withFTPClientFactory(openFTPClient)()
	destinationRoot := filepath.Join(t.TempDir(), "created-root")
	remoteRoot := "/outgoing"
	remoteFile := path.Join(remoteRoot, "nested", "child.txt")
	openFTPClient = func(ctx context.Context, opts FTPOptions) (ftpCore, error) {
		return &recordingFTPClient{walkEntries: []walkEntry{{path: remoteRoot, isDir: true}, {path: path.Join(remoteRoot, "nested"), isDir: true}, {path: remoteFile, isDir: false}}}, nil
	}
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
	if _, err := os.Stat(filepath.Join(destinationRoot, "nested", "child.txt")); err != nil {
		t.Fatalf("expected synced file under destination root: %v", err)
	}
	if result.PathsVisited != 3 || result.FilesAttempted != 1 || result.DirectoriesAttempted != 2 {
		t.Fatalf("expected compact FTP→local summary counts, got %#v", result)
	}
}

func TestSyncOnceFTPToLocalSuccess(t *testing.T) {
	defer withFTPClientFactory(openFTPClient)()
	destinationRoot := t.TempDir()
	remoteRoot := "/outgoing"
	remoteFile := path.Join(remoteRoot, "nested", "child.txt")
	var captured *recordingFTPClient
	openFTPClient = func(ctx context.Context, opts FTPOptions) (ftpCore, error) {
		captured = &recordingFTPClient{walkEntries: []walkEntry{{path: remoteRoot, isDir: true}, {path: remoteFile, isDir: false}}}
		return captured, nil
	}
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
	if !containsFTPRead(captured.readFiles, remoteFile) || result.Direction != DirectionFTPToLocal || result.DestinationRoot != destinationRoot {
		t.Fatalf("expected FTP→local result and read, result=%#v reads=%#v", result, captured.readFiles)
	}
}

func TestSyncOnceFTPToLocalNormalizesRemotePathAndIgnoresTrailingSlash(t *testing.T) {
	defer withFTPClientFactory(openFTPClient)()
	destinationRoot := t.TempDir()
	remoteFile := `/outgoing\nested\child.txt`
	var captured *recordingFTPClient
	openFTPClient = func(ctx context.Context, opts FTPOptions) (ftpCore, error) {
		captured = &recordingFTPClient{walkEntries: []walkEntry{{path: `/outgoing\`, isDir: true}, {path: remoteFile, isDir: false}}}
		return captured, nil
	}
	opts := completeFTPToLocalOptionsForOneShot()
	opts.Source.FTP.RemotePath = `/outgoing/`
	opts.Destination.LocalPath = destinationRoot + string(os.PathSeparator)
	svc, err := NewFTPSyncService(opts)
	if err != nil {
		t.Fatalf("construct service: %v", err)
	}

	_, err = svc.SyncOnce(context.Background())
	if err != nil {
		t.Fatalf("SyncOnce returned error: %v", err)
	}
	if !containsFTPRead(captured.readFiles, remoteFile) {
		t.Fatalf("expected FTP read for backslash remote path, reads=%#v", captured.readFiles)
	}
	if _, err := os.Stat(filepath.Join(destinationRoot, "nested", "child.txt")); err != nil {
		t.Fatalf("expected remote file mapped under local root independent of trailing slash: %v", err)
	}
}

func TestSyncOnceIgnoreRulesNormalizeWindowsStylePaths(t *testing.T) {
	ignore, err := newSyncOnceIgnore([]IgnoreRule{
		{Kind: IgnoreKindLiteral, Pattern: `tmp\cache\`},
		{Kind: IgnoreKindGlob, Pattern: `build\*.tmp`},
	})
	if err != nil {
		t.Fatalf("compile ignore rules: %v", err)
	}

	if !ignore.MatchPath(`tmp/cache`, "test", "") {
		t.Fatalf("expected slash-style path to match Windows literal ignore rule")
	}
	if !ignore.MatchPath(`build\output.tmp`, "test", "") {
		t.Fatalf("expected Windows-style path to match normalized glob ignore rule")
	}
	if ignore.MatchPath(`build\output.log`, "test", "") {
		t.Fatalf("expected non-matching extension not to match normalized glob ignore rule")
	}
}

func TestFTPPathHelpersNormalizeSeparatorsAndTrailingSlashes(t *testing.T) {
	if got := cleanFTPPath(`\incoming\nested\`); got != "/incoming/nested" {
		t.Fatalf("cleanFTPPath normalized incorrectly: %q", got)
	}
	if got := joinFTPPath(`/incoming/`, `nested\child.txt`); got != "/incoming/nested/child.txt" {
		t.Fatalf("joinFTPPath normalized incorrectly: %q", got)
	}
	if got := remoteRelativePath(`/incoming/`, `\incoming\nested\child.txt`); got != "nested/child.txt" {
		t.Fatalf("remoteRelativePath normalized incorrectly: %q", got)
	}
	root := filepath.Join(t.TempDir(), "dest") + string(os.PathSeparator)
	if got := localTargetPath(root, `/incoming/`, `/incoming\nested\child.txt`); got != filepath.Join(filepath.Clean(root), "nested", "child.txt") {
		t.Fatalf("localTargetPath normalized incorrectly: %q", got)
	}
}

func TestSyncOnceFTPToLocalNeverWritesToCWD(t *testing.T) {
	defer withFTPClientFactory(openFTPClient)()
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
	t.Cleanup(func() { _ = os.Chdir(originalWD) })
	openFTPClient = func(ctx context.Context, opts FTPOptions) (ftpCore, error) {
		return &recordingFTPClient{walkEntries: []walkEntry{{path: remoteRoot, isDir: true}, {path: path.Join(remoteRoot, "nested"), isDir: true}, {path: remoteFile, isDir: false}}}, nil
	}
	opts := completeFTPToLocalOptionsForOneShot()
	opts.Destination.LocalPath = destinationRoot
	svc, err := NewFTPSyncService(opts)
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
	defer withFTPClientFactory(openFTPClient)()
	destinationRoot := t.TempDir()
	remoteRoot := "/outgoing"
	failingFile := path.Join(remoteRoot, "nested", "fail.txt")
	afterFile := path.Join(remoteRoot, "after.txt")
	openFTPClient = func(ctx context.Context, opts FTPOptions) (ftpCore, error) {
		return &recordingFTPClient{walkEntries: []walkEntry{{path: remoteRoot, isDir: true}, {path: path.Join(remoteRoot, "nested"), isDir: true}, {path: failingFile, isDir: false}, {path: afterFile, isDir: false}}, failReadPath: failingFile}, nil
	}
	opts := completeFTPToLocalOptionsForOneShot()
	opts.Destination.LocalPath = destinationRoot
	svc, err := NewFTPSyncService(opts)
	if err != nil {
		t.Fatalf("construct service: %v", err)
	}
	result, err := svc.SyncOnce(context.Background())
	if err == nil || !IsKind(err, ErrTransfer) {
		t.Fatalf("expected transfer error, got %v", err)
	}
	if !result.Partial || result.FailureCount == 0 {
		t.Fatalf("expected partial FTP→local result, got %#v", result)
	}
	if _, err := os.Stat(filepath.Join(destinationRoot, "after.txt")); err != nil {
		t.Fatalf("expected later file to still be written after failure: %v", err)
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

func withFTPClientFactory(factory ftpClientFactory) func() {
	previous := openFTPClient
	openFTPClient = factory
	return func() {
		openFTPClient = previous
	}
}

type recordingFTPClient struct {
	mkdirs        []string
	writes        []recordedWrite
	removes       []string
	readFiles     []recordedRead
	walkEntries   []walkEntry
	failWritePath string
	failReadPath  string
	closed        bool
}

func newRecordingFTPClient() *recordingFTPClient {
	return &recordingFTPClient{}
}

type recordedWrite struct {
	remotePath string
	localPath  string
}

type recordedRead struct {
	remotePath string
	localPath  string
}

type walkEntry struct {
	path  string
	isDir bool
}

func (r *recordingFTPClient) mkdirAll(remotePath string) error {
	r.mkdirs = append(r.mkdirs, remotePath)
	return nil
}

func (r *recordingFTPClient) writeFile(remotePath string, localPath string) error {
	r.writes = append(r.writes, recordedWrite{remotePath: remotePath, localPath: localPath})
	if localPath == r.failWritePath || remotePath == r.failWritePath {
		return errors.New("simulated write failure")
	}
	return nil
}

func (r *recordingFTPClient) remove(remotePath string) error {
	r.removes = append(r.removes, remotePath)
	return nil
}

func (r *recordingFTPClient) walk(root string, fn fs.WalkDirFunc) error {
	for _, entry := range r.walkEntries {
		if err := fn(entry.path, stubDirEntry{entry: entry}, nil); err != nil {
			return err
		}
	}
	return nil
}

func (r *recordingFTPClient) readFile(remotePath string, localPath string) error {
	r.readFiles = append(r.readFiles, recordedRead{remotePath: remotePath, localPath: localPath})
	if remotePath == r.failReadPath {
		return errors.New("simulated read failure")
	}
	if err := os.MkdirAll(filepath.Dir(localPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(localPath, []byte("synced"), 0o644)
}

func (r *recordingFTPClient) close() error {
	r.closed = true
	return nil
}

func containsFTPWrite(writes []recordedWrite, localPath string) bool {
	for _, write := range writes {
		if write.localPath == localPath || write.remotePath == localPath {
			return true
		}
	}
	return false
}

func containsFTPWriteRemote(writes []recordedWrite, remotePath string) bool {
	for _, write := range writes {
		if write.remotePath == remotePath {
			return true
		}
	}
	return false
}

func containsFTPRead(reads []recordedRead, remotePath string) bool {
	for _, read := range reads {
		if read.remotePath == remotePath {
			return true
		}
	}
	return false
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
