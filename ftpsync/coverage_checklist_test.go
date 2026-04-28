package ftpsync

import (
	"os"
	"strings"
	"testing"
)

func TestVerificationCoverageChecklist(t *testing.T) {
	testCases := []struct {
		file string
		want []string
	}{
		{file: "validation_test.go", want: []string{"NewFTPSyncService", "TestValidateAcceptsSupportedDirections"}},
		{file: "oneshot_test.go", want: []string{"TestSyncOnceLocalToFTP", "TestSyncOnceFTPToLocalSuccess", "TestSyncOnceFTPToLocalNeverWritesToCWD"}},
		{file: "internal_ftp.go", want: []string{"newFTPPathCodec", "ftpEncodingAuto", "ftpEncodingUTF8", "ftpEncodingGBK"}},
		{file: "options.go", want: []string{"PassiveMode"}},
		{file: "context_test.go", want: []string{"TestSyncOnceChecksValidationAndContext", "TestStartBackgroundChecksValidationAndContext"}},
		{file: "background_test.go", want: []string{"TestBackgroundHandleWait", "TestBackgroundStopShutsDownDeterministically", "TestBackgroundContextCancelStopsRunner"}},
	}

	for _, tc := range testCases {
		t.Run(tc.file, func(t *testing.T) {
			content, err := os.ReadFile(tc.file)
			if err != nil {
				t.Fatalf("read coverage target %s: %v", tc.file, err)
			}
			text := string(content)
			for _, want := range tc.want {
				if !strings.Contains(text, want) {
					t.Fatalf("%s missing verification coverage target %q", tc.file, want)
				}
			}
		})
	}
}
