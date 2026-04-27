package ftpsync

import "errors"

var errInvalidOptions = errors.New("invalid ftpsync options")

// FTPSyncService is the public FTP synchronization service boundary.
type FTPSyncService struct {
	opts Options
}

// NewFTPSyncService validates and stores a private copy of typed options.
func NewFTPSyncService(opts Options) (*FTPSyncService, error) {
	if err := validateOptions(opts); err != nil {
		return nil, err
	}

	return &FTPSyncService{opts: copyOptions(opts)}, nil
}

func validateOptions(opts Options) error {
	switch opts.Direction {
	case DirectionLocalToFTP:
		if opts.Source.LocalPath == "" || opts.Destination.FTP.Host == "" || opts.Destination.FTP.RemotePath == "" {
			return errInvalidOptions
		}
	case DirectionFTPToLocal:
		if opts.Source.FTP.Host == "" || opts.Source.FTP.RemotePath == "" || opts.Destination.LocalPath == "" {
			return errInvalidOptions
		}
	default:
		return errInvalidOptions
	}

	return nil
}

func copyOptions(opts Options) Options {
	if len(opts.IgnoreRules) > 0 {
		ignoreRules := make([]IgnoreRule, len(opts.IgnoreRules))
		copy(ignoreRules, opts.IgnoreRules)
		opts.IgnoreRules = ignoreRules
	}

	return opts
}
