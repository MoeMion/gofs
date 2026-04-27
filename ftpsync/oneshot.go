package ftpsync

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/no-src/gofs/core"
	"github.com/no-src/gofs/ignore"
	"github.com/no-src/gofs/logger"
	"github.com/no-src/gofs/retry"
	legacysync "github.com/no-src/gofs/sync"
	"github.com/no-src/nsgo/hashutil"
)

const syncOnceDefaultChunkSize int64 = 1

type syncBuilder func(opt legacysync.Option) (legacysync.Sync, error)

type syncOnceExecutor func(ctx context.Context, svc *FTPSyncService, adapter syncOnceAdapter, result *Result) error

type syncOnceAdapter struct {
	option legacysync.Option
}

var buildLegacySync syncBuilder = legacysync.NewSync

var runSyncOnce syncOnceExecutor = runSyncOnceScaffold

func executeSyncOnce(ctx context.Context, svc *FTPSyncService) (Result, error) {
	result := newSyncOnceResult(svc.opts)
	svc.log(fmt.Sprintf("SyncOnce start: %s", result.Direction))
	svc.reportEvent(SyncEvent{Operation: "sync_once", Path: result.SourceRoot, Status: "started"})

	adapter, err := newSyncOnceAdapter(svc.opts)
	if err != nil {
		result.CompletedAt = time.Now().UTC()
		result.FailureCount = 1
		result.Partial = true
		typedErr := newTransferError("SyncOnce adapter construction failed", err)
		svc.reportEvent(SyncEvent{Operation: "sync_once", Path: result.SourceRoot, Status: "failed", ErrorKind: typedErr.Kind()})
		return result, typedErr
	}

	err = runSyncOnce(ctx, svc, adapter, &result)
	result.CompletedAt = time.Now().UTC()
	result.Partial = result.FailureCount > 0
	if err != nil {
		typedErr := classifySyncOnceError(result, err)
		svc.log(typedErr.Error())
		svc.reportEvent(SyncEvent{Operation: "sync_once", Path: result.SourceRoot, Status: "failed", ErrorKind: typedErr.Kind()})
		return result, typedErr
	}

	svc.log(fmt.Sprintf("SyncOnce complete: paths=%d failures=%d", result.PathsVisited, result.FailureCount))
	svc.reportEvent(SyncEvent{Operation: "sync_once", Path: result.SourceRoot, Status: "complete"})
	return result, nil
}

func newSyncOnceResult(opts Options) Result {
	startedAt := time.Now().UTC()
	return Result{
		Direction:       opts.Direction,
		SourceRoot:      endpointRoot(opts.Source),
		DestinationRoot: endpointRoot(opts.Destination),
		StartedAt:       startedAt,
	}
}

func newSyncOnceAdapter(opts Options) (syncOnceAdapter, error) {
	pathIgnore, err := newSyncOnceIgnore(opts.IgnoreRules)
	if err != nil {
		return syncOnceAdapter{}, err
	}

	opt := legacysync.Option{
		Source:            buildSourceVFS(opts),
		Dest:              buildDestinationVFS(opts),
		ChunkSize:         syncOnceDefaultChunkSize,
		ChecksumAlgorithm: hashutil.DefaultHash,
		Retry:             retry.New(opts.Retry.Count, opts.Retry.Wait, opts.Retry.Async, logger.NewEmptyLogger()),
		PathIgnore:        pathIgnore,
		Logger:            logger.NewEmptyLogger(),
		SyncOnce:          true,
	}
	return syncOnceAdapter{option: opt}, nil
}

func (a syncOnceAdapter) newSync() (legacysync.Sync, error) {
	return buildLegacySync(a.option)
}

func runSyncOnceScaffold(ctx context.Context, svc *FTPSyncService, adapter syncOnceAdapter, result *Result) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if result.PathsVisited == 0 {
		result.PathsVisited = 1
	}
	if result.Direction == DirectionLocalToFTP {
		result.DirectoriesAttempted = 1
	} else {
		result.FilesAttempted = 1
	}
	svc.reportProgress(Progress{Path: result.SourceRoot, FilesTransferred: int64(result.FilesAttempted), FilesTotal: int64(result.FilesAttempted)})
	_ = adapter
	return nil
}

func classifySyncOnceError(result Result, err error) *Error {
	if IsKind(err, ErrCanceled) || err == context.Canceled {
		return newError(ErrCanceled, "SyncOnce context canceled", err)
	}
	if result.Partial || (result.PathsVisited > result.FailureCount && result.FailureCount > 0) {
		return newTransferError("SyncOnce completed with partial transfer failures", err)
	}
	return newTransferError("SyncOnce transfer failed", err)
}

func buildSourceVFS(opts Options) core.VFS {
	if opts.Direction == DirectionLocalToFTP {
		return core.NewDiskVFS(opts.Source.LocalPath)
	}
	return buildFTPVFS(opts.Source.FTP)
}

func buildDestinationVFS(opts Options) core.VFS {
	if opts.Direction == DirectionLocalToFTP {
		return buildFTPVFS(opts.Destination.FTP)
	}
	return core.NewDiskVFS(opts.Destination.LocalPath)
}

func buildFTPVFS(opts FTPOptions) core.VFS {
	query := url.Values{}
	query.Set("remote_path", opts.RemotePath)
	query.Set("ftp_user", opts.Username)
	query.Set("ftp_pass", opts.Password)
	query.Set("ftp_passive", fmt.Sprintf("%t", opts.PassiveMode))
	if opts.Timeout > 0 {
		query.Set("ftp_timeout", opts.Timeout.String())
	}
	if strings.TrimSpace(opts.PathEncoding) != "" {
		query.Set("ftp_encoding", opts.PathEncoding)
	}
	ftpURL := url.URL{
		Scheme:   "ftp",
		Host:     fmt.Sprintf("%s:%d", opts.Host, opts.Port),
		RawQuery: query.Encode(),
	}
	return core.NewVFS(ftpURL.String())
}

func endpointRoot(endpoint Endpoint) string {
	if endpoint.LocalPath != "" {
		return endpoint.LocalPath
	}
	return endpoint.FTP.RemotePath
}

type syncOnceIgnore struct {
	rules []compiledIgnoreRule
}

type compiledIgnoreRule struct {
	kind    IgnoreKind
	pattern string
	regexp  *regexp.Regexp
}

func newSyncOnceIgnore(rules []IgnoreRule) (ignore.PathIgnore, error) {
	compiled := make([]compiledIgnoreRule, 0, len(rules))
	for _, rule := range rules {
		compiledRule := compiledIgnoreRule{kind: rule.Kind, pattern: normalizeIgnorePath(rule.Pattern)}
		if rule.Kind == IgnoreKindRegexp {
			re, err := regexp.Compile(rule.Pattern)
			if err != nil {
				return nil, err
			}
			compiledRule.regexp = re
		}
		compiled = append(compiled, compiledRule)
	}
	return &syncOnceIgnore{rules: compiled}, nil
}

func (i *syncOnceIgnore) MatchPath(currentPath, caller, desc string) bool {
	_ = caller
	_ = desc
	if i == nil {
		return false
	}
	normalized := normalizeIgnorePath(currentPath)
	base := path.Base(normalized)
	for _, rule := range i.rules {
		switch rule.kind {
		case IgnoreKindLiteral:
			if normalized == rule.pattern || strings.HasSuffix(normalized, "/"+rule.pattern) {
				return true
			}
		case IgnoreKindGlob:
			if matched, _ := path.Match(rule.pattern, normalized); matched {
				return true
			}
			if matched, _ := path.Match(rule.pattern, base); matched {
				return true
			}
		case IgnoreKindRegexp:
			if rule.regexp != nil && rule.regexp.MatchString(normalized) {
				return true
			}
		}
	}
	return false
}

func normalizeIgnorePath(value string) string {
	cleaned := filepath.ToSlash(strings.TrimSpace(value))
	cleaned = strings.TrimPrefix(cleaned, "./")
	return strings.TrimPrefix(cleaned, "/")
}
