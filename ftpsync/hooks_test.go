package ftpsync

import (
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestHookContractsUseLibraryLocalTypes(t *testing.T) {
	var logs []string
	var progress []Progress
	var events []SyncEvent

	hooks := HookSet{
		Logger: LoggerFunc(func(message string) {
			logs = append(logs, message)
		}),
		Progress: ProgressHook(func(snapshot Progress) {
			progress = append(progress, snapshot)
		}),
		Event: EventHook(func(event SyncEvent) {
			events = append(events, event)
		}),
	}

	hooks.Logger.Log("library message")
	hooks.Progress(Progress{Path: "file.txt", BytesTransferred: 10, BytesTotal: 20, FilesTransferred: 1, FilesTotal: 2})
	hooks.Event(SyncEvent{Operation: "upload", Path: "file.txt", Status: "complete", ErrorKind: ErrTransfer})

	if len(logs) != 1 || logs[0] != "library message" {
		t.Fatalf("expect custom logger to receive library message, got %#v", logs)
	}
	if len(progress) != 1 || progress[0].BytesTransferred != 10 || progress[0].FilesTotal != 2 {
		t.Fatalf("expect typed progress snapshot, got %#v", progress)
	}
	if len(events) != 1 || events[0].ErrorKind != ErrTransfer || events[0].Path != "file.txt" {
		t.Fatalf("expect typed sync event, got %#v", events)
	}
}

func TestHookContractsDoNotExposeLegacyRuntimeTypes(t *testing.T) {
	cmd := exec.Command("go", "doc", ".")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("go doc hook contract check failed: %v", err)
	}

	forbidden := []string{"logger.Logger", "report.Reporter", "eventlog.Event", "web report", "global logger"}
	for _, token := range forbidden {
		if strings.Contains(string(out), token) {
			t.Fatalf("public ftpsync docs expose forbidden runtime token %q", token)
		}
	}
}

func TestHookContractsZeroValueNormalizes(t *testing.T) {
	hooks := normalizeHooks(HookSet{})
	hooks.Logger.Log("no-op message")
	hooks.Progress(Progress{})
	hooks.Event(SyncEvent{})
}

func TestHookDefaultsAreNoOp(t *testing.T) {
	var zero FTPSyncService
	zero.log("zero service log")
	zero.reportProgress(Progress{Path: "zero.txt"})
	zero.reportEvent(SyncEvent{Operation: "noop", Path: "zero.txt", Status: "ignored"})

	svc, err := NewFTPSyncService(completeLocalToFTPHookOptions())
	if err != nil {
		t.Fatalf("construct service: %v", err)
	}

	svc.log("default log")
	svc.reportProgress(Progress{Path: "file.txt"})
	svc.reportEvent(SyncEvent{Operation: "upload", Path: "file.txt", Status: "complete"})
}

func TestCustomHooksReceiveServiceDispatches(t *testing.T) {
	const secret = "hook-secret-password"
	var logs []string
	var progress []Progress
	var events []SyncEvent

	opts := completeLocalToFTPHookOptions()
	opts.Destination.FTP.Password = secret
	opts.Hooks = HookSet{
		Logger: LoggerFunc(func(message string) {
			logs = append(logs, message)
		}),
		Progress: ProgressHook(func(snapshot Progress) {
			progress = append(progress, snapshot)
		}),
		Event: EventHook(func(event SyncEvent) {
			events = append(events, event)
		}),
	}

	svc, err := NewFTPSyncService(opts)
	if err != nil {
		t.Fatalf("construct service: %v", err)
	}
	svc.log("sync starting")
	svc.reportProgress(Progress{Path: "file.txt", BytesTransferred: 5, BytesTotal: 10, FilesTransferred: 1, FilesTotal: 1})
	svc.reportEvent(SyncEvent{Operation: "upload", Path: "file.txt", Status: "complete"})

	if len(logs) != 1 || logs[0] != "sync starting" {
		t.Fatalf("expect service log dispatch, got %#v", logs)
	}
	if strings.Contains(logs[0], secret) {
		t.Fatalf("log dispatch leaked FTP password")
	}
	if len(progress) != 1 || progress[0].BytesTransferred != 5 || progress[0].BytesTotal != 10 {
		t.Fatalf("expect service progress dispatch, got %#v", progress)
	}
	if len(events) != 1 || events[0].Operation != "upload" || events[0].ErrorKind != "" {
		t.Fatalf("expect service event dispatch, got %#v", events)
	}
}

func TestPackageDependencyBoundaryForHooks(t *testing.T) {
	cmd := exec.Command("go", "list", "-deps", ".")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("go list dependency check failed: %v", err)
	}

	forbidden := []string{
		"github.com/no-src/gofs/report",
		"github.com/no-src/gofs/eventlog",
		"github.com/no-src/gofs/server",
		"github.com/no-src/gofs/logger",
	}
	deps := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, dep := range deps {
		for _, forbiddenDep := range forbidden {
			if dep == forbiddenDep || strings.HasPrefix(dep, forbiddenDep+"/") {
				t.Fatalf("ftpsync imports forbidden hook dependency %s", dep)
			}
		}
	}
}

func completeLocalToFTPHookOptions() Options {
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
