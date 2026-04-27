package ftpsync

import "time"

// Direction identifies the supported FTP sync direction.
type Direction string

const (
	// DirectionLocalToFTP syncs from a local filesystem path to an FTP endpoint.
	DirectionLocalToFTP Direction = "local_to_ftp"
	// DirectionFTPToLocal syncs from an FTP endpoint to a local filesystem path.
	DirectionFTPToLocal Direction = "ftp_to_local"
)

// Endpoint describes one side of a sync pair using typed Go values.
type Endpoint struct {
	// LocalPath is the local filesystem root for a disk endpoint.
	LocalPath string
	// FTP is the FTP connection and remote path description for an FTP endpoint.
	FTP FTPOptions
}

// FTPOptions describes an FTP endpoint without URL, YAML, or CLI parsing.
type FTPOptions struct {
	Host         string
	Port         int
	Username     string
	Password     string
	RemotePath   string
	PassiveMode  bool
	Timeout      time.Duration
	PathEncoding string
}

// RetryOptions describes retry behavior for future sync execution.
type RetryOptions struct {
	Count int
	Wait  time.Duration
	Async bool
}

// IgnoreKind identifies how an ignore rule pattern should be interpreted.
type IgnoreKind string

const (
	// IgnoreKindLiteral matches a literal relative path.
	IgnoreKindLiteral IgnoreKind = "literal"
	// IgnoreKindGlob matches a glob-style path pattern.
	IgnoreKindGlob IgnoreKind = "glob"
	// IgnoreKindRegexp matches a regular-expression path pattern.
	IgnoreKindRegexp IgnoreKind = "regexp"
)

// IgnoreRule describes one typed ignore rule.
type IgnoreRule struct {
	Kind    IgnoreKind
	Pattern string
}

// Options contains the public FTPSyncService construction contract.
type Options struct {
	Source      Endpoint
	Destination Endpoint
	Direction   Direction
	Retry       RetryOptions
	IgnoreRules []IgnoreRule

	// Hook fields are reserved for Plan 03 and intentionally remain library-local
	// callback slots instead of depending on global logger or report packages.
	LogHook      func(message string)
	ProgressHook func(transferred, total int64)
	EventHook    func(name, path string)
}
