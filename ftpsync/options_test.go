package ftpsync_test

import (
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/no-src/gofs/ftpsync"
)

func TestOptionsExposeTypedConfiguration(t *testing.T) {
	opts := ftpsync.Options{
		Direction: ftpsync.DirectionLocalToFTP,
		Source: ftpsync.Endpoint{
			LocalPath: "/data/source",
		},
		Destination: ftpsync.Endpoint{
			FTP: ftpsync.FTPOptions{
				Host:         "ftp.example.test",
				Port:         21,
				Username:     "ftp-user",
				Password:     "top-secret-test-password",
				RemotePath:   "/incoming",
				PassiveMode:  true,
				Timeout:      15 * time.Second,
				PathEncoding: "utf-8",
			},
		},
		Retry: ftpsync.RetryOptions{
			Count: 3,
			Wait:  time.Second,
			Async: true,
		},
		IgnoreRules: []ftpsync.IgnoreRule{
			{Kind: ftpsync.IgnoreKindLiteral, Pattern: "tmp/cache"},
			{Kind: ftpsync.IgnoreKindGlob, Pattern: "*.part"},
			{Kind: ftpsync.IgnoreKindRegexp, Pattern: `^debug/.*`},
		},
	}

	if opts.Direction != ftpsync.DirectionLocalToFTP {
		t.Errorf("expect local to FTP direction")
	}
	if opts.Source.LocalPath != "/data/source" {
		t.Errorf("expect local source path to be stored as typed field")
	}
	ftp := opts.Destination.FTP
	if ftp.Host != "ftp.example.test" || ftp.Port != 21 || ftp.Username != "ftp-user" {
		t.Errorf("expect FTP endpoint host, port, and username to be stored as typed fields")
	}
	if ftp.Password != "top-secret-test-password" || ftp.RemotePath != "/incoming" {
		t.Errorf("expect FTP password and remote path to be stored as typed fields")
	}
	if !ftp.PassiveMode || ftp.Timeout != 15*time.Second || ftp.PathEncoding != "utf-8" {
		t.Errorf("expect FTP passive mode, timeout, and path encoding to be stored as typed fields")
	}
	if opts.Retry.Count != 3 || opts.Retry.Wait != time.Second || !opts.Retry.Async {
		t.Errorf("expect retry count, wait, and async fields to be stored as typed values")
	}
	if len(opts.IgnoreRules) != 3 {
		t.Fatalf("expect ignore rules to be stored, got %d", len(opts.IgnoreRules))
	}
}

func TestOptionsSupportFTPToLocalDirection(t *testing.T) {
	opts := ftpsync.Options{
		Direction: ftpsync.DirectionFTPToLocal,
		Source: ftpsync.Endpoint{
			FTP: ftpsync.FTPOptions{
				Host:       "ftp.example.test",
				RemotePath: "/outgoing",
			},
		},
		Destination: ftpsync.Endpoint{LocalPath: "/data/destination"},
	}

	if opts.Direction != ftpsync.DirectionFTPToLocal {
		t.Errorf("expect FTP to local direction")
	}
	if opts.Source.FTP.RemotePath != "/outgoing" || opts.Destination.LocalPath != "/data/destination" {
		t.Errorf("expect typed FTP source and local destination fields")
	}
}

func TestOptionsZeroValueCompiles(t *testing.T) {
	var opts ftpsync.Options
	if opts.Direction != "" {
		t.Errorf("expect zero-value direction to remain unset")
	}
}

func TestNewFTPSyncServiceConstructsFromTypedOptions(t *testing.T) {
	opts := completeLocalToFTPOptions()
	svc, err := ftpsync.NewFTPSyncService(opts)
	if err != nil {
		t.Fatalf("expect service construction to succeed, got %v", err)
	}
	if svc == nil {
		t.Fatalf("expect service instance")
	}
}

func TestNewFTPSyncServiceDoesNotLeakPasswordInErrors(t *testing.T) {
	const secret = "do-not-leak-this-secret"
	opts := completeLocalToFTPOptions()
	opts.Direction = "unsupported"
	opts.Destination.FTP.Password = secret

	_, err := ftpsync.NewFTPSyncService(opts)
	if err == nil {
		t.Fatalf("expect invalid direction to fail")
	}
	if strings.Contains(err.Error(), secret) {
		t.Fatalf("constructor error leaked FTP password")
	}
}

func TestPackageDependencyBoundary(t *testing.T) {
	cmd := exec.Command("go", "list", "-deps", ".")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("go list dependency check failed: %v", err)
	}

	forbidden := []string{}
	deps := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, dep := range deps {
		for _, forbiddenDep := range forbidden {
			if dep == forbiddenDep || strings.HasPrefix(dep, forbiddenDep+"/") {
				t.Fatalf("ftpsync imports forbidden dependency %s", dep)
			}
		}
	}
}

func completeLocalToFTPOptions() ftpsync.Options {
	return ftpsync.Options{
		Direction: ftpsync.DirectionLocalToFTP,
		Source:    ftpsync.Endpoint{LocalPath: "/data/source"},
		Destination: ftpsync.Endpoint{
			FTP: ftpsync.FTPOptions{
				Host:         "ftp.example.test",
				Port:         21,
				Username:     "ftp-user",
				Password:     "secret-password",
				RemotePath:   "/incoming",
				PassiveMode:  true,
				Timeout:      15 * time.Second,
				PathEncoding: "utf-8",
			},
		},
	}
}
