package sync

import (
	"github.com/no-src/gofs/core"
	"github.com/no-src/gofs/driver"
	"github.com/no-src/gofs/driver/ftp"
	"github.com/no-src/gofs/logger"
	"github.com/no-src/gofs/retry"
)

var newFTPPullDriver = func(remoteAddr string, ftpConfig core.FTPConfig, autoReconnect bool, r retry.Retry, maxTranRate int64, logger *logger.Logger) driver.Driver {
	return ftp.NewFTPDriver(remoteAddr, ftpConfig, autoReconnect, r, maxTranRate, logger)
}

type ftpPullClientSync struct {
	driverPullClientSync

	remoteAddr string
}

// NewFTPPullClientSync creates an FTP-specific sync entry point.
func NewFTPPullClientSync(opt Option) (Sync, error) {
	// the fields of option
	source := opt.Source
	chunkSize := opt.ChunkSize
	maxTranRate := opt.MaxTranRate
	r := opt.Retry
	logger := opt.Logger

	if chunkSize <= 0 {
		return nil, errInvalidChunkSize
	}

	ds, err := newDiskSync(opt)
	if err != nil {
		return nil, err
	}

	s := &ftpPullClientSync{
		driverPullClientSync: newDriverPullClientSync(*ds),
		remoteAddr:           source.Addr(),
	}
	s.driver = newFTPPullDriver(s.remoteAddr, source.FTPConfig(), true, r, maxTranRate, logger)

	err = s.start()
	if err != nil {
		return nil, err
	}

	// reset the sourceAbsPath because the source.RemotePath() is absolute representation of path and the source.RemotePath() may be cross-platform
	s.diskSync.sourceAbsPath = source.RemotePath().Base()

	// reset some functions for ftp
	s.diskSync.isDirFn = s.IsDir
	s.diskSync.statFn = s.driver.Stat
	s.diskSync.getFileTimeFn = s.driver.GetFileTime

	return s, nil
}
