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

func TestSyncOnceChecksValidationAndContext(t *testing.T) {
	var invalid ftpsync.FTPSyncService
	_, err := invalid.SyncOnce(context.Background())
	if err == nil || !ftpsync.IsKind(err, ftpsync.ErrValidation) {
		t.Fatalf("expect validation error before sync dispatch, got %v", err)
	}

	svc, err := ftpsync.NewFTPSyncService(completeLocalToFTPOptions())
	if err != nil {
		t.Fatalf("construct service: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = svc.SyncOnce(ctx)
	if err == nil {
		t.Fatalf("expect canceled context to fail")
	}
	if !errors.Is(err, context.Canceled) || !ftpsync.IsKind(err, ftpsync.ErrCanceled) {
		t.Fatalf("expect typed cancellation error, got %v", err)
	}

	_, err = svc.SyncOnce(context.Background())
	if err == nil || !ftpsync.IsKind(err, ftpsync.ErrUnsupportedCapability) {
		t.Fatalf("expect unsupported one-shot capability until implementation phase, got %v", err)
	}
}

func TestStartBackgroundChecksValidationAndContext(t *testing.T) {
	var invalid ftpsync.FTPSyncService
	_, err := invalid.StartBackground(context.Background())
	if err == nil || !ftpsync.IsKind(err, ftpsync.ErrValidation) {
		t.Fatalf("expect validation error before background dispatch, got %v", err)
	}

	svc, err := ftpsync.NewFTPSyncService(completeLocalToFTPOptions())
	if err != nil {
		t.Fatalf("construct service: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = svc.StartBackground(ctx)
	if err == nil {
		t.Fatalf("expect canceled context to fail")
	}
	if !errors.Is(err, context.Canceled) || !ftpsync.IsKind(err, ftpsync.ErrCanceled) {
		t.Fatalf("expect typed cancellation error, got %v", err)
	}

	_, err = svc.StartBackground(context.Background())
	if err == nil || !ftpsync.IsKind(err, ftpsync.ErrUnsupportedCapability) {
		t.Fatalf("expect unsupported background capability until implementation phase, got %v", err)
	}
}

func TestStartBackgroundRejectsFTPToLocal(t *testing.T) {
	svc, err := ftpsync.NewFTPSyncService(completeFTPToLocalOptions())
	if err != nil {
		t.Fatalf("construct service: %v", err)
	}
	_, err = svc.StartBackground(context.Background())
	if err == nil || !ftpsync.IsKind(err, ftpsync.ErrUnsupportedCapability) {
		t.Fatalf("expect FTP to local background to be unsupported, got %v", err)
	}
	if !strings.Contains(err.Error(), "StartBackground") || !strings.Contains(err.Error(), string(ftpsync.DirectionFTPToLocal)) {
		t.Fatalf("expect method and direction context in unsupported error, got %v", err)
	}
}

func TestContextAwarePublicContractsCompile(t *testing.T) {
	var result ftpsync.Result
	var handle ftpsync.Handle
	_ = result
	_ = handle

	svc, err := ftpsync.NewFTPSyncService(completeLocalToFTPOptions())
	if err != nil {
		t.Fatalf("construct service: %v", err)
	}
	_, _ = svc.SyncOnce(context.Background())
	_, _ = svc.StartBackground(context.Background())
}
