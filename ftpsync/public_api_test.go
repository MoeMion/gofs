package ftpsync

import (
	"context"
	"errors"
	"os/exec"
	"reflect"
	"strings"
	"testing"
)

func TestPublicAPIStaysTypedOptionsOnly(t *testing.T) {
	errorType := reflect.TypeOf((*error)(nil)).Elem()
	contextType := reflect.TypeOf((*context.Context)(nil)).Elem()
	optionsType := reflect.TypeOf(Options{})
	resultType := reflect.TypeOf(Result{})
	handleType := reflect.TypeOf((*Handle)(nil)).Elem()
	servicePointerType := reflect.TypeOf((*FTPSyncService)(nil))

	constructor := reflect.TypeOf(NewFTPSyncService)
	if constructor.NumIn() != 1 || constructor.In(0) != optionsType {
		t.Fatalf("NewFTPSyncService must accept exactly ftpsync.Options, got %s", constructor)
	}
	if constructor.NumOut() != 2 || constructor.Out(0) != servicePointerType || constructor.Out(1) != errorType {
		t.Fatalf("NewFTPSyncService must return (*ftpsync.FTPSyncService, error), got %s", constructor)
	}

	syncOnce := reflect.TypeOf((*FTPSyncService).SyncOnce)
	if syncOnce.NumIn() != 2 || syncOnce.In(0) != servicePointerType || syncOnce.In(1) != contextType {
		t.Fatalf("SyncOnce must accept receiver and context.Context only, got %s", syncOnce)
	}
	if syncOnce.NumOut() != 2 || syncOnce.Out(0) != resultType || syncOnce.Out(1) != errorType {
		t.Fatalf("SyncOnce must return (ftpsync.Result, error), got %s", syncOnce)
	}

	startBackground := reflect.TypeOf((*FTPSyncService).StartBackground)
	if startBackground.NumIn() != 2 || startBackground.In(0) != servicePointerType || startBackground.In(1) != contextType {
		t.Fatalf("StartBackground must accept receiver and context.Context only, got %s", startBackground)
	}
	if startBackground.NumOut() != 2 || startBackground.Out(0) != handleType || startBackground.Out(1) != errorType {
		t.Fatalf("StartBackground must return (ftpsync.Handle, error), got %s", startBackground)
	}

	assertApprovedTypedOptionsCompile(t)
	assertNoLegacyPublicTypeNames(t)
}

func assertApprovedTypedOptionsCompile(t *testing.T) {
	t.Helper()
	_ = Options{
		Source: Endpoint{LocalPath: "/local/source"},
		Destination: Endpoint{FTP: FTPOptions{
			Host:         "ftp.example.test",
			Port:         21,
			Username:     "user",
			Password:     "password",
			RemotePath:   "/remote",
			PassiveMode:  true,
			PathEncoding: "utf-8",
		}},
		Direction: DirectionLocalToFTP,
		Retry: RetryOptions{
			Count: 1,
		},
		IgnoreRules: []IgnoreRule{{Kind: IgnoreKindLiteral, Pattern: "tmp"}},
		Hooks: HookSet{
			Logger:   LoggerFunc(func(string) {}),
			Progress: ProgressHook(func(Progress) {}),
			Event:    EventHook(func(SyncEvent) {}),
		},
	}
	_ = DirectionFTPToLocal
	_ = errors.Is
}

func assertNoLegacyPublicTypeNames(t *testing.T) {
	t.Helper()

	cmd := exec.Command("go", "doc", "-all", ".")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("go doc public API scan failed: %v", err)
	}

	rejectedMarkers := []string{"Config", "VFS", "Server", "Task"}
	approvedLoggerTypes := map[string]bool{"Logger": true, "LoggerFunc": true}

	for _, line := range strings.Split(string(out), "\n") {
		name, ok := publicTypeNameFromDocLine(line)
		if !ok {
			continue
		}
		for _, marker := range rejectedMarkers {
			if strings.Contains(name, marker) {
				t.Fatalf("public ftpsync type %q leaks rejected marker %q", name, marker)
			}
		}
		if strings.Contains(name, "Logger") && !approvedLoggerTypes[name] {
			t.Fatalf("public ftpsync type %q leaks rejected Logger marker", name)
		}
	}
}

func publicTypeNameFromDocLine(line string) (string, bool) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "type ") {
		return "", false
	}

	name := strings.TrimPrefix(line, "type ")
	name = strings.TrimSpace(name)
	if name == "" {
		return "", false
	}
	fields := strings.FieldsFunc(name, func(r rune) bool {
		return r == ' ' || r == '[' || r == '{' || r == '='
	})
	if len(fields) == 0 || fields[0] == "(" {
		return "", false
	}
	return fields[0], true
}
