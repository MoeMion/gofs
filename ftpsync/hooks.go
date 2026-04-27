package ftpsync

// Logger receives library-local informational messages from FTPSyncService.
// Hooks run synchronously in the caller's process.
type Logger interface {
	Log(message string)
}

// LoggerFunc adapts a function to the Logger hook interface.
type LoggerFunc func(message string)

// Log sends a library-local message to f.
func (f LoggerFunc) Log(message string) {
	f(message)
}

// Progress describes file and byte progress for a sync operation.
type Progress struct {
	Path             string
	BytesTransferred int64
	BytesTotal       int64
	FilesTransferred int64
	FilesTotal       int64
}

// ProgressHook receives typed progress snapshots.
type ProgressHook func(snapshot Progress)

// SyncEvent describes a sync operation event without exposing credentials.
type SyncEvent struct {
	Operation string
	Path      string
	Status    string
	ErrorKind ErrorKind
}

// EventHook receives typed sync event notifications.
type EventHook func(event SyncEvent)

// HookSet contains optional FTPSyncService observability callbacks.
type HookSet struct {
	Logger   Logger
	Progress ProgressHook
	Event    EventHook
}

type noopLogger struct{}

func (noopLogger) Log(string) {}
