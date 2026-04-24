package monitor

import (
	"errors"

	"github.com/no-src/gofs/wait"
)

var errFTPMonitorRequiresPolling = errors.New("ftp source monitor requires -sync_once or -sync_cron because FTP has no event stream")

type ftpPullClientMonitor struct {
	driverPullClientMonitor
}

// NewFTPPullClientMonitor creates an FTP-specific pull monitor entry point.
func NewFTPPullClientMonitor(opt Option) (m Monitor, err error) {
	m = &ftpPullClientMonitor{
		driverPullClientMonitor: driverPullClientMonitor{
			baseMonitor: newBaseMonitor(opt),
		},
	}
	return m, nil
}

func (m *ftpPullClientMonitor) Start() (wait.Wait, error) {
	if !m.syncOnce && len(m.syncSpec) == 0 {
		return nil, errFTPMonitorRequiresPolling
	}
	return m.driverPullClientMonitor.Start()
}
