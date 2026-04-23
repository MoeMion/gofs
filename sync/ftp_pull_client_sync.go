package sync

// NewFTPPullClientSync creates an FTP-specific sync entry point.
func NewFTPPullClientSync(opt Option) (Sync, error) {
	return nil, errFTPBackendDeferred
}
