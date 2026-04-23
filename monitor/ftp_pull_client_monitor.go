package monitor

import "errors"

var errFTPMonitorDeferred = errors.New("ftp monitor backend is deferred to phase 2")

// NewFTPPullClientMonitor creates an FTP-specific pull monitor entry point.
func NewFTPPullClientMonitor(opt Option) (m Monitor, err error) {
	return nil, errFTPMonitorDeferred
}
