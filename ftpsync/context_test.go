package ftpsync_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/no-src/gofs/ftpsync"
)

func TestErrorKindValidation(t *testing.T) {
	opts := completeLocalToFTPOptions()
	opts.Source.LocalPath = ""

	_, err := ftpsync.NewFTPSyncService(opts)
	if err == nil {
		t.Fatalf("expect missing source path to fail validation")
	}
	if !ftpsync.IsKind(err, ftpsync.ErrValidation) {
		t.Fatalf("expect validation error kind, got %v", err)
	}
}

func TestErrorWrapping(t *testing.T) {
	testCases := []struct {
		name string
		kind ftpsync.ErrorKind
		msg  string
	}{
		{name: "connection", kind: ftpsync.ErrConnection, msg: "dial ftp.example.test"},
		{name: "authentication", kind: ftpsync.ErrAuthentication, msg: "login ftp-user"},
		{name: "transfer", kind: ftpsync.ErrTransfer, msg: "upload file"},
		{name: "unsupported capability", kind: ftpsync.ErrUnsupportedCapability, msg: "StartBackground ftp_to_local"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cause := errors.New("underlying cause")
			err := &ftpsync.Error{KindValue: tc.kind, Message: tc.msg, Cause: cause}
			if !ftpsync.IsKind(err, tc.kind) {
				t.Fatalf("expect kind %s", tc.kind)
			}
			if !errors.Is(err, cause) {
				t.Fatalf("expect wrapped cause to be discoverable")
			}
			if !strings.Contains(err.Error(), tc.msg) {
				t.Fatalf("expect error message to include context %q, got %q", tc.msg, err.Error())
			}
		})
	}
}

func TestContextCancellationErrorKind(t *testing.T) {
	err := &ftpsync.Error{KindValue: ftpsync.ErrCanceled, Message: "SyncOnce context canceled", Cause: context.Canceled}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expect cancellation to wrap context.Canceled")
	}
	if !ftpsync.IsKind(err, ftpsync.ErrCanceled) {
		t.Fatalf("expect cancellation error kind")
	}
}
