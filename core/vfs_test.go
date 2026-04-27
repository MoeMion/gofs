package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/no-src/nsgo/jsonutil"
)

const (
	testVFSServerPath                               = "rs://127.0.0.1:8105?mode=server&local_sync_disabled=true&path=./source&fs_server=https://127.0.0.1"
	testVFSServerPathWithNoPort                     = "rs://127.0.0.1?mode=server&local_sync_disabled=true&path=./source&fs_server=https://127.0.0.1"
	testVFSServerPathWithNoSchemeFsServer           = "rs://127.0.0.1:8105?mode=server&local_sync_disabled=true&path=./source&fs_server=127.0.0.1"
	testVFSSFTPDestPath                             = "sftp://127.0.0.1:22?mode=server&local_sync_disabled=true&path=./source&remote_path=/home/remote/dest&ssh_user=sftp_user&ssh_pass=sftp_pwd&ssh_key=./id_rsa&ssh_key_pass=123456&ssh_host_key=/root/.ssh/known_hosts"
	testVFSSFTPDestPathWithNoPort                   = "sftp://127.0.0.1?mode=server&local_sync_disabled=true&path=./source&remote_path=/home/remote/dest&ssh_user=sftp_user&ssh_pass=sftp_pwd&ssh_key=./id_rsa&ssh_key_pass=123456&ssh_host_key=/root/.ssh/known_hosts"
	testVFSSFTPSSHConfigDestPath                    = "sftp://example.com:22?mode=server&local_sync_disabled=true&path=./source&remote_path=/home/remote/dest&ssh_pass=sftp_pwd&ssh_key=./id_rsa&ssh_key_pass=123456&ssh_host_key=/root/.ssh/known_hosts&ssh_config=true"
	testVFSSFTPSSHConfigDestPathWithNoPort          = "sftp://example.com?mode=server&local_sync_disabled=true&path=./source&remote_path=/home/remote/dest&ssh_user=sftp_user&ssh_pass=sftp_pwd&ssh_key=./id_rsa&ssh_key_pass=123456&ssh_host_key=/root/.ssh/known_hosts&ssh_config=true"
	testVFSSFTPSSHConfigDestPathWithCover           = "sftp://example.com:22?mode=server&local_sync_disabled=true&path=./source&remote_path=/home/remote/dest&ssh_user=sftp_user&ssh_pass=sftp_pwd&ssh_key=./id_rsa&ssh_key_pass=123456&ssh_host_key=/root/.ssh/known_hosts&ssh_config=true"
	testVFSSFTPSSHConfigDestPathWithDefaultIdentity = "sftp://default-identity?mode=server&local_sync_disabled=true&path=./source&remote_path=/home/remote/dest&ssh_pass=sftp_pwd&ssh_config=true"
	testVFSFTPSrcPath                               = "ftp://127.0.0.1:21?path=./dest&remote_path=/srv/source&ftp_user=ftp_user&ftp_pass=ftp_pwd&ftp_timeout=30s"
	testVFSFTPDestPath                              = "ftp://127.0.0.1:21?path=./source&remote_path=/home/remote/dest&ftp_user=ftp_user&ftp_pass=ftp_pwd&ftp_timeout=30s"
	testVFSFTPDestPathWithNoPort                    = "ftp://127.0.0.1?path=./source&remote_path=/home/remote/dest&ftp_user=ftp_user&ftp_pass=ftp_pwd&ftp_passive=false"
	testVFSFTPDestPathWithGBKEncoding               = "ftp://127.0.0.1:21?path=./source&remote_path=/home/remote/dest&ftp_user=ftp_user&ftp_pass=ftp_pwd&ftp_timeout=30s&ftp_encoding=gbk"
	testVFSMinIODestPath                            = "minio://127.0.0.1:9000?mode=server&local_sync_disabled=true&path=./source&remote_path=/home/remote/dest&secure=true"
	testVFSMinIODestPathWithNoPort                  = "minio://127.0.0.1?mode=server&local_sync_disabled=true&path=./source&remote_path=/home/remote/dest&secure=false"
)

func TestVFS_MarshalText(t *testing.T) {
	testCases := []struct {
		path string
	}{
		{""},
		{testVFSServerPath},
		{testVFSServerPathWithNoPort},
		{testVFSServerPathWithNoSchemeFsServer},
		{testVFSSFTPDestPath},
		{testVFSSFTPDestPathWithNoPort},
		{testVFSSFTPSSHConfigDestPath},
		{testVFSSFTPSSHConfigDestPathWithNoPort},
		{testVFSSFTPSSHConfigDestPathWithCover},
		{testVFSSFTPSSHConfigDestPathWithDefaultIdentity},
		{testVFSFTPSrcPath},
		{testVFSFTPDestPath},
		{testVFSFTPDestPathWithNoPort},
		{testVFSMinIODestPath},
		{testVFSMinIODestPathWithNoPort},
	}

	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			vfs := NewVFS(tc.path)
			data, err := jsonutil.Marshal(vfs)
			if err != nil {
				t.Errorf("test duration marshal error =>%s", err)
				return
			}
			var buf bytes.Buffer
			json.HTMLEscape(&buf, []byte(tc.path))
			expect := fmt.Sprintf("\"%s\"", buf.String())
			actual := string(data)
			if actual != expect {
				t.Errorf("test vfs marshal error, expect:%s, actual:%s", expect, actual)
			}
		})
	}
}

func TestVFS_UnmarshalText(t *testing.T) {
	testCases := []struct {
		path string
	}{
		{""},
		{testVFSServerPath},
		{testVFSServerPathWithNoPort},
		{testVFSServerPathWithNoSchemeFsServer},
		{testVFSSFTPDestPath},
		{testVFSSFTPDestPathWithNoPort},
		{testVFSSFTPSSHConfigDestPath},
		{testVFSSFTPSSHConfigDestPathWithNoPort},
		{testVFSSFTPSSHConfigDestPathWithCover},
		{testVFSSFTPSSHConfigDestPathWithDefaultIdentity},
		{testVFSFTPSrcPath},
		{testVFSFTPDestPath},
		{testVFSFTPDestPathWithNoPort},
		{testVFSMinIODestPath},
		{testVFSMinIODestPathWithNoPort},
	}

	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			var actual VFS
			data := []byte(fmt.Sprintf("\"%s\"", tc.path))
			err := jsonutil.Unmarshal(data, &actual)
			if err != nil {
				t.Errorf("test vfs unmarshal error =>%s", err)
				return
			}
			compareVFS(t, NewVFS(tc.path), actual)
		})
	}
}

func TestNewVFS_WithDefaultPort(t *testing.T) {
	testCases := []struct {
		path       string
		expectPort int
	}{
		{testVFSServerPathWithNoPort, remoteServerDefaultPort},
		{testVFSFTPDestPathWithNoPort, ftpServerDefaultPort},
		{testVFSSFTPDestPathWithNoPort, sftpServerDefaultPort},
	}

	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			actual := NewVFS(tc.path)
			if tc.expectPort != actual.Port() {
				t.Errorf("test new vfs with default port error, expect:%d, actual:%d", tc.expectPort, actual.Port())
			}
		})
	}
}

func TestNewVFS_FTPPassiveModeDefaultsToTrue(t *testing.T) {
	vfs := NewVFS(testVFSFTPDestPath)
	if !vfs.FTPPassiveMode() {
		t.Fatal("expect ftp passive mode default to true when ftp_passive is omitted")
	}

	vfs = NewVFS(testVFSFTPDestPathWithNoPort)
	if vfs.FTPPassiveMode() {
		t.Fatal("expect explicit ftp_passive=false to disable passive mode")
	}
}

func TestNewVFS_WithNoSchemeFsServer(t *testing.T) {
	testCases := []struct {
		path   string
		expect string
	}{
		{testVFSServerPathWithNoSchemeFsServer, "https://127.0.0.1"},
	}

	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			actual := NewVFS(tc.path)
			if tc.expect != actual.FsServer() {
				t.Errorf("test new vfs with no scheme fs server error, expect:%s, actual:%s", tc.expect, actual.FsServer())
			}
		})
	}
}

func TestNewVFS_ReturnError(t *testing.T) {
	testCases := []struct {
		path   string
		expect VFS
	}{
		{testVFSServerPath + string([]byte{127}), NewEmptyVFS()}, // 0x7F DEL
		{testVFSSFTPDestPath + string([]byte{127}), NewEmptyVFS()},
		{testVFSFTPDestPath + string([]byte{127}), NewEmptyVFS()},
		{testVFSMinIODestPath + string([]byte{127}), NewEmptyVFS()},
	}

	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			compareVFS(t, tc.expect, NewVFS(tc.path))
		})
	}
}

func TestVFSVar_DefaultValue(t *testing.T) {
	testCases := []struct {
		name         string
		defaultValue VFS
	}{
		{"default_empty_vfs", NewEmptyVFS()},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var actual VFS
			testCommandLine.VFSVar(&actual, "core_test_vfs_var_default"+tc.name, tc.defaultValue, "test vfs var")
			parseFlag()
			compareVFS(t, tc.defaultValue, actual)
		})
	}
}

func TestVFSVar(t *testing.T) {
	testCases := []struct {
		name         string
		path         string
		defaultValue VFS
	}{
		{"testVFSServerPath", testVFSServerPath, NewEmptyVFS()},
		{"testVFSServerPathWithNoPort", testVFSServerPathWithNoPort, NewEmptyVFS()},
		{"testVFSServerPathWithNoSchemeFsServer", testVFSServerPathWithNoSchemeFsServer, NewEmptyVFS()},

		{"testVFSSFTPDestPath", testVFSSFTPDestPath, NewEmptyVFS()},
		{"testVFSSFTPDestPathWithNoPort", testVFSSFTPDestPathWithNoPort, NewEmptyVFS()},

		{"testVFSSFTPSSHConfigDestPath", testVFSSFTPSSHConfigDestPath, NewEmptyVFS()},
		{"testVFSSFTPSSHConfigDestPathWithNoPort", testVFSSFTPSSHConfigDestPathWithNoPort, NewEmptyVFS()},
		{"testVFSSFTPSSHConfigDestPathWithCover", testVFSSFTPSSHConfigDestPathWithCover, NewEmptyVFS()},
		{"testVFSSFTPSSHConfigDestPathWithDefaultIdentity", testVFSSFTPSSHConfigDestPathWithDefaultIdentity, NewEmptyVFS()},
		{"testVFSFTPSrcPath", testVFSFTPSrcPath, NewEmptyVFS()},
		{"testVFSFTPDestPath", testVFSFTPDestPath, NewEmptyVFS()},
		{"testVFSFTPDestPathWithNoPort", testVFSFTPDestPathWithNoPort, NewEmptyVFS()},

		{"testVFSMinIODestPath", testVFSMinIODestPath, NewEmptyVFS()},
		{"testVFSMinIODestPathWithNoPort", testVFSMinIODestPathWithNoPort, NewEmptyVFS()},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var actual VFS
			expect := NewVFS(tc.path)
			flagName := "core_test_vfs_var" + tc.name
			testCommandLine.VFSVar(&actual, flagName, tc.defaultValue, "test vfs var")
			parseFlag(fmt.Sprintf("-%s=%s", flagName, tc.path))
			compareVFS(t, expect, actual)
		})
	}
}

func TestVFSFlag_DefaultValue(t *testing.T) {
	testCases := []struct {
		name         string
		defaultValue VFS
	}{
		{"default_empty_vfs", NewEmptyVFS()},
		{"with_normal_vfs", NewVFS(testVFSServerPath)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var actual *VFS
			flagName := "core_test_vfs_flag_default" + tc.name
			actual = testCommandLine.VFSFlag(flagName, tc.defaultValue, "test vfs flag")
			parseFlag()
			compareVFS(t, tc.defaultValue, *actual)
		})
	}
}

func TestVFSFlag(t *testing.T) {
	testCases := []struct {
		name         string
		path         string
		defaultValue VFS
	}{
		{"testVFSServerPath", testVFSServerPath, NewEmptyVFS()},
		{"testVFSServerPathWithNoPort", testVFSServerPathWithNoPort, NewEmptyVFS()},
		{"testVFSServerPathWithNoSchemeFsServer", testVFSServerPathWithNoSchemeFsServer, NewEmptyVFS()},

		{"testVFSSFTPDestPath", testVFSSFTPDestPath, NewEmptyVFS()},
		{"testVFSSFTPDestPathWithNoPort", testVFSSFTPDestPathWithNoPort, NewEmptyVFS()},
		{"testVFSFTPSrcPath", testVFSFTPSrcPath, NewEmptyVFS()},
		{"testVFSFTPDestPath", testVFSFTPDestPath, NewEmptyVFS()},
		{"testVFSFTPDestPathWithNoPort", testVFSFTPDestPathWithNoPort, NewEmptyVFS()},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			expect := NewVFS(tc.path)
			flagName := "core_test_vfs_flag" + tc.name
			actual := testCommandLine.VFSFlag(flagName, tc.defaultValue, "test vfs flag")
			parseFlag(fmt.Sprintf("-%s=%s", flagName, tc.path))
			compareVFS(t, expect, *actual)
		})
	}
}

func compareVFS(t *testing.T, expect, actual VFS) {
	assert(t, expect.original == actual.original, "compare vfs original error, expect:%s, actual:%s", expect.original, actual.original)
	assert(t, expect.Path() == actual.Path(), "compare vfs Path error, expect:%s, actual:%s", expect.Path(), actual.Path())
	assert(t, expect.RemotePath() == actual.RemotePath(), "compare vfs RemotePath error, expect:%s, actual:%s", expect.RemotePath(), actual.RemotePath())

	expectAbs, err := expect.Abs()
	if err != nil {
		t.Errorf("compare vfs Abs error, parse expect abs error =>%s", err)
		return
	}

	actualAbs, err := actual.Abs()
	if err != nil {
		t.Errorf("compare vfs Abs error, parse actual abs error =>%s", err)
		return
	}

	assert(t, expectAbs == actualAbs, "compare vfs Abs error, expect:%s, actual:%s", expectAbs, actualAbs)
	assert(t, expect.IsEmpty() == actual.IsEmpty(), "compare vfs IsEmpty error, expect:%v, actual:%v", expect.IsEmpty(), actual.IsEmpty())
	assert(t, expect.Type() == actual.Type(), "compare vfs Type error, expect:%v, actual:%v", expect.Type(), actual.Type())
	assert(t, expect.Host() == actual.Host(), "compare vfs Host error, expect:%s, actual:%s", expect.Host(), actual.Host())
	assert(t, expect.Port() == actual.Port(), "compare vfs Port error, expect:%d, actual:%d", expect.Port(), actual.Port())
	assert(t, expect.Addr() == actual.Addr(), "compare vfs Addr error, expect:%s, actual:%s", expect.Addr(), actual.Addr())
	assert(t, expect.IsDisk() == actual.IsDisk(), "compare vfs IsDisk error, expect:%v, actual:%v", expect.IsDisk(), actual.IsDisk())
	assert(t, expect.Server() == actual.Server(), "compare vfs Server error, expect:%v, actual:%v", expect.Server(), actual.Server())
	assert(t, expect.FsServer() == actual.FsServer(), "compare vfs FsServer error, expect:%s, actual:%s", expect.FsServer(), actual.FsServer())
	assert(t, expect.LocalSyncDisabled() == actual.LocalSyncDisabled(), "compare vfs LocalSyncDisabled error, expect:%v, actual:%v", expect.LocalSyncDisabled(), actual.LocalSyncDisabled())
	assert(t, expect.Secure() == actual.Secure(), "compare vfs Secure error, expect:%v, actual:%v", expect.Secure(), actual.Secure())
	expectFTPConfig := expect.FTPConfig()
	actualFTPConfig := actual.FTPConfig()
	assert(t, expectFTPConfig.Username == actualFTPConfig.Username, "compare vfs FTPConfig.Username error, expect:%v, actual:%v", expectFTPConfig.Username, actualFTPConfig.Username)
	assert(t, expectFTPConfig.Password == actualFTPConfig.Password, "compare vfs FTPConfig.Password error, expect:%v, actual:%v", expectFTPConfig.Password, actualFTPConfig.Password)
	assert(t, expectFTPConfig.Timeout == actualFTPConfig.Timeout, "compare vfs FTPConfig.Timeout error, expect:%v, actual:%v", expectFTPConfig.Timeout, actualFTPConfig.Timeout)
	assert(t, expectFTPConfig.Encoding == actualFTPConfig.Encoding, "compare vfs FTPConfig.Encoding error, expect:%v, actual:%v", expectFTPConfig.Encoding, actualFTPConfig.Encoding)
	assert(t, expectFTPConfig.PassiveMode == actualFTPConfig.PassiveMode, "compare vfs FTPConfig.PassiveMode error, expect:%v, actual:%v", expectFTPConfig.PassiveMode, actualFTPConfig.PassiveMode)
	expectSSHConfig := expect.SSHConfig()
	actualSSHConfig := actual.SSHConfig()
	assert(t, expectSSHConfig.Username == actualSSHConfig.Username, "compare vfs SSHConfig.Username error, expect:%v, actual:%v", expectSSHConfig.Username, actualSSHConfig.Username)
	assert(t, expectSSHConfig.Password == actualSSHConfig.Password, "compare vfs SSHConfig.Password error, expect:%v, actual:%v", expectSSHConfig.Password, actualSSHConfig.Password)
	assert(t, expectSSHConfig.Key == actualSSHConfig.Key, "compare vfs SSHConfig.Key error, expect:%v, actual:%v", expectSSHConfig.Key, actualSSHConfig.Key)
	assert(t, expectSSHConfig.KeyPass == actualSSHConfig.KeyPass, "compare vfs SSHConfig.KeyPass error, expect:%v, actual:%v", expectSSHConfig.KeyPass, actualSSHConfig.KeyPass)
	assert(t, expectSSHConfig.HostKey == actualSSHConfig.HostKey, "compare vfs SSHConfig.HostKey error, expect:%v, actual:%v", expectSSHConfig.HostKey, actualSSHConfig.HostKey)
}

func TestNewVFS_FTPConfig(t *testing.T) {
	vfs := NewVFS(testVFSFTPDestPath)
	if vfs.Type() != FTP {
		t.Errorf("test ftp vfs type error, expect:%v, actual:%v", FTP, vfs.Type())
	}
	if vfs.FTPUsername() != "ftp_user" {
		t.Errorf("test ftp username error, expect:%s, actual:%s", "ftp_user", vfs.FTPUsername())
	}
	if vfs.FTPPassword() != "ftp_pwd" {
		t.Errorf("test ftp password error, expect:%s, actual:%s", "ftp_pwd", vfs.FTPPassword())
	}
	if vfs.FTPTimeout() != "30s" {
		t.Errorf("test ftp timeout error, expect:%s, actual:%s", "30s", vfs.FTPTimeout())
	}
	if vfs.FTPEncoding() != "auto" {
		t.Errorf("test ftp encoding default error, expect:%s, actual:%s", "auto", vfs.FTPEncoding())
	}
	if !vfs.FTPPassiveMode() {
		t.Errorf("test ftp passive mode error, expect:true, actual:%v", vfs.FTPPassiveMode())
	}
	if vfs.SSHConfig() != (SSHConfig{}) {
		t.Errorf("test ftp ssh config isolation error, expect empty ssh config, actual:%+v", vfs.SSHConfig())
	}
}

func TestNewVFS_FTPConfigWithoutOptionalTimeout(t *testing.T) {
	vfs := NewVFS(testVFSFTPDestPathWithNoPort)
	if vfs.Type() != FTP {
		t.Errorf("test ftp vfs type without port error, expect:%v, actual:%v", FTP, vfs.Type())
	}
	if vfs.Port() != ftpServerDefaultPort {
		t.Errorf("test ftp default port error, expect:%d, actual:%d", ftpServerDefaultPort, vfs.Port())
	}
	if vfs.FTPTimeout() != "" {
		t.Errorf("test ftp timeout optional error, expect empty, actual:%s", vfs.FTPTimeout())
	}
	if vfs.FTPEncoding() != "auto" {
		t.Errorf("test ftp encoding default without explicit value error, expect:%s, actual:%s", "auto", vfs.FTPEncoding())
	}
	if vfs.FTPPassiveMode() {
		t.Errorf("test ftp passive mode false error, expect:false, actual:%v", vfs.FTPPassiveMode())
	}
}

func TestNewVFS_FTPConfigWithExplicitEncoding(t *testing.T) {
	vfs := NewVFS(testVFSFTPDestPathWithGBKEncoding)
	if vfs.FTPEncoding() != "gbk" {
		t.Errorf("test ftp encoding explicit error, expect:%s, actual:%s", "gbk", vfs.FTPEncoding())
	}
}

func assert(t *testing.T, ok bool, format string, args ...any) {
	if !ok {
		t.Errorf(format, args...)
	}
}
