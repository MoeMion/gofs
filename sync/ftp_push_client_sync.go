package sync

import "errors"

var errFTPBackendDeferred = errors.New("ftp backend is deferred to phase 2")

// NewFTPPushClientSync creates an FTP-specific sync entry point.
func NewFTPPushClientSync(opt Option) (Sync, error) {
	return nil, errFTPBackendDeferred
}
