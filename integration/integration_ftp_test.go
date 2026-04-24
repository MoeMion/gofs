//go:build integration_test_ftp

package integration

import "testing"

func TestIntegration_FTP(t *testing.T) {
	testCases := []struct {
		name          string
		runServerConf string
		runClientConf string
		testConf      string
	}{
		{"gofs FTP push", "", "run-gofs-ftp-push-client.yaml", "test-gofs-ftp-push.yaml"},
		{"gofs FTP pull", "", "run-gofs-ftp-pull-client.yaml", "test-gofs-ftp-pull.yaml"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testIntegrationClientServer(t, tc.runServerConf, tc.runClientConf, tc.testConf)
		})
	}
}
