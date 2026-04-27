package ftpsync

import (
	"context"
	"errors"
	"time"
)

var errInvalidOptions = errors.New("invalid ftpsync options")

// FTPSyncService is the public FTP synchronization service boundary.
type FTPSyncService struct {
	opts  Options
	hooks HookSet
}

// Result is the public summary contract returned by one-shot sync calls.
type Result struct {
	Direction            Direction
	SourceRoot           string
	DestinationRoot      string
	StartedAt            time.Time
	CompletedAt          time.Time
	PathsVisited         int
	FilesAttempted       int
	DirectoriesAttempted int
	FailureCount         int
	Partial              bool
}

// Handle is the public lifecycle contract returned by background sync calls.
type Handle interface {
	Done() <-chan struct{}
	Err() error
	Stop(context.Context) error
}

// NewFTPSyncService validates and stores a private copy of typed options.
func NewFTPSyncService(opts Options) (*FTPSyncService, error) {
	copied := copyOptions(opts)
	svc := &FTPSyncService{opts: copied, hooks: normalizeHooks(copied.Hooks)}
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

// SyncOnce validates options and context before dispatching a one-shot sync.
func (s *FTPSyncService) SyncOnce(ctx context.Context) (Result, error) {
	if err := s.Validate(); err != nil {
		return Result{}, err
	}
	if err := validateContext(ctx, "SyncOnce"); err != nil {
		return Result{}, err
	}
	return executeSyncOnce(ctx, s)
}

// StartBackground validates options and context before starting background sync.
func (s *FTPSyncService) StartBackground(ctx context.Context) (Handle, error) {
	if err := s.Validate(); err != nil {
		return nil, err
	}
	if err := validateContext(ctx, "StartBackground"); err != nil {
		return nil, err
	}
	return nil, unsupportedMethod("StartBackground", s.opts.Direction)
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

func unsupportedMethod(method string, direction Direction) error {
	return newError(ErrUnsupportedCapability, method+" is unsupported for "+string(direction), nil)
}

func validateContext(ctx context.Context, method string) error {
	if ctx == nil {
		return newError(ErrValidation, method+" context is required", errInvalidOptions)
	}
	if err := ctx.Err(); err != nil {
		return newError(ErrCanceled, method+" context canceled", err)
	}
	return nil
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

func normalizeHooks(hooks HookSet) HookSet {
	if hooks.Logger == nil {
		hooks.Logger = noopLogger{}
	}
	if hooks.Progress == nil {
		hooks.Progress = func(Progress) {}
	}
	if hooks.Event == nil {
		hooks.Event = func(SyncEvent) {}
	}
	return hooks
}

func (s *FTPSyncService) log(message string) {
	if s == nil {
		return
	}
	s.normalizedHooks().Logger.Log(message)
}

func (s *FTPSyncService) reportProgress(progress Progress) {
	if s == nil {
		return
	}
	s.normalizedHooks().Progress(progress)
}

func (s *FTPSyncService) reportEvent(event SyncEvent) {
	if s == nil {
		return
	}
	s.normalizedHooks().Event(event)
}

func (s *FTPSyncService) normalizedHooks() HookSet {
	return normalizeHooks(s.hooks)
}
