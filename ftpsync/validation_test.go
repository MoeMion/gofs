package ftpsync_test

import (
	"strings"
	"testing"
	"time"

	"github.com/no-src/gofs/ftpsync"
)

func TestValidateAcceptsSupportedDirections(t *testing.T) {
	testCases := []struct {
		name string
		opts ftpsync.Options
	}{
		{name: "local to FTP", opts: completeLocalToFTPOptions()},
		{name: "FTP to local", opts: completeFTPToLocalOptions()},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc, err := ftpsync.NewFTPSyncService(tc.opts)
			if err != nil {
				t.Fatalf("expect construction to succeed, got %v", err)
			}
			if err := svc.Validate(); err != nil {
				t.Fatalf("expect validation to succeed, got %v", err)
			}
		})
	}
}

func TestValidateRejectsUnsupportedEndpointCombinations(t *testing.T) {
	testCases := []struct {
		name string
		opts ftpsync.Options
		kind ftpsync.ErrorKind
	}{
		{name: "local to local", opts: ftpsync.Options{Direction: ftpsync.DirectionLocalToFTP, Source: ftpsync.Endpoint{LocalPath: "/src"}, Destination: ftpsync.Endpoint{LocalPath: "/dst"}}, kind: ftpsync.ErrUnsupportedCapability},
		{name: "FTP to FTP", opts: ftpsync.Options{Direction: ftpsync.DirectionFTPToLocal, Source: completeFTPToLocalOptions().Source, Destination: completeLocalToFTPOptions().Destination}, kind: ftpsync.ErrUnsupportedCapability},
		{name: "non FTP endpoints", opts: ftpsync.Options{Direction: ftpsync.DirectionLocalToFTP}, kind: ftpsync.ErrUnsupportedCapability},
		{name: "missing local source path", opts: withLocalSourcePath(completeLocalToFTPOptions(), ""), kind: ftpsync.ErrValidation},
		{name: "missing FTP destination host", opts: withDestinationFTPHost(completeLocalToFTPOptions(), ""), kind: ftpsync.ErrValidation},
		{name: "missing FTP destination remote path", opts: withDestinationFTPRemotePath(completeLocalToFTPOptions(), ""), kind: ftpsync.ErrValidation},
		{name: "invalid FTP destination port", opts: withDestinationFTPPort(completeLocalToFTPOptions(), -1), kind: ftpsync.ErrValidation},
		{name: "invalid FTP destination timeout", opts: withDestinationFTPTimeout(completeLocalToFTPOptions(), -1), kind: ftpsync.ErrValidation},
		{name: "invalid FTP source timeout", opts: withSourceFTPTimeout(completeFTPToLocalOptions(), -1), kind: ftpsync.ErrValidation},
		{name: "ambiguous local to FTP source", opts: withSourceFTP(completeLocalToFTPOptions(), completeFTPToLocalOptions().Source.FTP), kind: ftpsync.ErrValidation},
		{name: "ambiguous local to FTP destination", opts: withDestinationLocalPath(completeLocalToFTPOptions(), "/also-local"), kind: ftpsync.ErrValidation},
		{name: "mismatched direction source and destination", opts: ftpsync.Options{Direction: ftpsync.DirectionFTPToLocal, Source: ftpsync.Endpoint{LocalPath: "/src"}, Destination: completeLocalToFTPOptions().Destination}, kind: ftpsync.ErrUnsupportedCapability},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ftpsync.NewFTPSyncService(tc.opts)
			if err == nil {
				t.Fatalf("expect validation to reject %s", tc.name)
			}
			if !ftpsync.IsKind(err, tc.kind) {
				t.Fatalf("expect error kind %s, got %v", tc.kind, err)
			}
		})
	}
}

func TestValidateDoesNotLeakFTPPassword(t *testing.T) {
	const secret = "validation-secret-password"
	opts := completeLocalToFTPOptions()
	opts.Destination.FTP.Password = secret
	opts.Destination.FTP.Port = -1

	_, err := ftpsync.NewFTPSyncService(opts)
	if err == nil {
		t.Fatalf("expect invalid port to fail")
	}
	if strings.Contains(err.Error(), secret) {
		t.Fatalf("validation error leaked FTP password")
	}
}

func completeFTPToLocalOptions() ftpsync.Options {
	return ftpsync.Options{
		Direction: ftpsync.DirectionFTPToLocal,
		Source: ftpsync.Endpoint{FTP: ftpsync.FTPOptions{
			Host:       "ftp.example.test",
			Port:       21,
			Username:   "ftp-user",
			Password:   "secret-password",
			RemotePath: "/outgoing",
		}},
		Destination: ftpsync.Endpoint{LocalPath: "/data/destination"},
	}
}

func withLocalSourcePath(opts ftpsync.Options, path string) ftpsync.Options {
	opts.Source.LocalPath = path
	return opts
}

func withDestinationFTPHost(opts ftpsync.Options, host string) ftpsync.Options {
	opts.Destination.FTP.Host = host
	return opts
}

func withDestinationFTPRemotePath(opts ftpsync.Options, path string) ftpsync.Options {
	opts.Destination.FTP.RemotePath = path
	return opts
}

func withDestinationFTPPort(opts ftpsync.Options, port int) ftpsync.Options {
	opts.Destination.FTP.Port = port
	return opts
}

func withDestinationFTPTimeout(opts ftpsync.Options, timeout int) ftpsync.Options {
	opts.Destination.FTP.Timeout = time.Duration(timeout)
	return opts
}

func withSourceFTPTimeout(opts ftpsync.Options, timeout int) ftpsync.Options {
	opts.Source.FTP.Timeout = time.Duration(timeout)
	return opts
}

func withSourceFTP(opts ftpsync.Options, ftp ftpsync.FTPOptions) ftpsync.Options {
	opts.Source.FTP = ftp
	return opts
}

func withDestinationLocalPath(opts ftpsync.Options, path string) ftpsync.Options {
	opts.Destination.LocalPath = path
	return opts
}
