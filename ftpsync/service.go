package ftpsync

import "errors"

var errInvalidOptions = errors.New("invalid ftpsync options")

// FTPSyncService is the public FTP synchronization service boundary.
type FTPSyncService struct {
	opts Options
}

// NewFTPSyncService validates and stores a private copy of typed options.
func NewFTPSyncService(opts Options) (*FTPSyncService, error) {
	svc := &FTPSyncService{opts: copyOptions(opts)}
	if err := svc.Validate(); err != nil {
		return nil, err
	}

	return svc, nil
}

// Validate checks that the service options describe a supported FTP sync pair.
func (s *FTPSyncService) Validate() error {
	if s == nil {
		return newError(ErrValidation, "service is nil", errInvalidOptions)
	}
	return validateOptions(s.opts)
}

func validateOptions(opts Options) error {
	if err := validateEndpointShape("source", opts.Source); err != nil {
		return err
	}
	if err := validateEndpointShape("destination", opts.Destination); err != nil {
		return err
	}

	sourceLocal := isLocalEndpoint(opts.Source)
	sourceFTP := isFTPEndpoint(opts.Source)
	destinationLocal := isLocalEndpoint(opts.Destination)
	destinationFTP := isFTPEndpoint(opts.Destination)

	switch opts.Direction {
	case DirectionLocalToFTP:
		if sourceLocal && destinationFTP {
			return validateFTPFields("destination.ftp", opts.Destination.FTP)
		}
		if !hasEndpointFields(opts.Source) && destinationFTP {
			return newError(ErrValidation, "source local path is required", errInvalidOptions)
		}
		return unsupportedDirection(opts.Direction)
	case DirectionFTPToLocal:
		if sourceFTP && destinationLocal {
			return validateFTPFields("source.ftp", opts.Source.FTP)
		}
		if sourceFTP && !hasEndpointFields(opts.Destination) {
			return newError(ErrValidation, "destination local path is required", errInvalidOptions)
		}
		return unsupportedDirection(opts.Direction)
	default:
		return newError(ErrValidation, "invalid direction", errInvalidOptions)
	}
}

func validateEndpointShape(name string, endpoint Endpoint) error {
	if endpoint.LocalPath != "" && hasFTPFields(endpoint.FTP) {
		return newError(ErrValidation, name+" endpoint is ambiguous: local path and FTP fields are both set", errInvalidOptions)
	}
	return nil
}

func validateFTPFields(name string, ftp FTPOptions) error {
	if ftp.Host == "" {
		return newError(ErrValidation, name+" host is required", errInvalidOptions)
	}
	if ftp.RemotePath == "" {
		return newError(ErrValidation, name+" remote path is required", errInvalidOptions)
	}
	if ftp.Port <= 0 {
		return newError(ErrValidation, name+" port must be positive", errInvalidOptions)
	}
	return nil
}

func unsupportedDirection(direction Direction) error {
	return newError(ErrUnsupportedCapability, "unsupported endpoint combination for "+string(direction), errInvalidOptions)
}

func isLocalEndpoint(endpoint Endpoint) bool {
	return endpoint.LocalPath != "" && !hasFTPFields(endpoint.FTP)
}

func isFTPEndpoint(endpoint Endpoint) bool {
	return endpoint.LocalPath == "" && hasFTPFields(endpoint.FTP)
}

func hasFTPFields(ftp FTPOptions) bool {
	return ftp.Host != "" || ftp.Port != 0 || ftp.Username != "" || ftp.Password != "" || ftp.RemotePath != "" || ftp.PassiveMode || ftp.Timeout != 0 || ftp.PathEncoding != ""
}

func hasEndpointFields(endpoint Endpoint) bool {
	return endpoint.LocalPath != "" || hasFTPFields(endpoint.FTP)
}

func copyOptions(opts Options) Options {
	if len(opts.IgnoreRules) > 0 {
		ignoreRules := make([]IgnoreRule, len(opts.IgnoreRules))
		copy(ignoreRules, opts.IgnoreRules)
		opts.IgnoreRules = ignoreRules
	}

	return opts
}
