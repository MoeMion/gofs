package sync

import (
	"errors"

	"github.com/no-src/gofs/core"
	"github.com/no-src/gofs/driver"
	"github.com/no-src/gofs/driver/ftp"
	"github.com/no-src/gofs/logger"
	"github.com/no-src/gofs/retry"
)

var errFTPBackendDeferred = errors.New("ftp backend is deferred to phase 2")

var newFTPPushDriver = func(remoteAddr string, ftpConfig core.FTPConfig, autoReconnect bool, r retry.Retry, maxTranRate int64, logger *logger.Logger) driver.Driver {
	return ftp.NewFTPDriver(remoteAddr, ftpConfig, autoReconnect, r, maxTranRate, logger)
}

type ftpPushClientSync struct {
	driverPushClientSync

	remoteAddr string
}

// NewFTPPushClientSync creates an FTP-specific sync entry point.
func NewFTPPushClientSync(opt Option) (Sync, error) {
	// the fields of option
	dest := opt.Dest
	chunkSize := opt.ChunkSize
	maxTranRate := opt.MaxTranRate
	r := opt.Retry
	logger := opt.Logger
	syncOnce := opt.SyncOnce
	syncCron := opt.SyncCron

	if chunkSize <= 0 {
		return nil, errInvalidChunkSize
	}

	ds, err := newDiskSync(opt)
	if err != nil {
		return nil, err
	}

	s := &ftpPushClientSync{
		driverPushClientSync: newDriverPushClientSync(*ds, dest.RemotePath().Base()),
		remoteAddr:           dest.Addr(),
	}

	s.driver = newFTPPushDriver(s.remoteAddr, dest.FTPConfig(), true, r, maxTranRate, logger)

	isSync := syncOnce || len(syncCron) > 0
	err = s.start(isSync)
	if err != nil {
		return nil, err
	}
	return s, nil
}
