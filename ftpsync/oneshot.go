package ftpsync

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

type syncOnceExecutor func(ctx context.Context, svc *FTPSyncService, result *Result) error

type ftpClientFactory func(ctx context.Context, opts FTPOptions) (ftpCore, error)

type ftpCore interface {
	mkdirAll(remotePath string) error
	writeFile(remotePath string, localPath string) error
	remove(remotePath string) error
	walk(root string, fn fs.WalkDirFunc) error
	readFile(remotePath string, localPath string) error
	close() error
}

var openFTPClient ftpClientFactory = func(ctx context.Context, opts FTPOptions) (ftpCore, error) {
	return newFTPClient(ctx, opts)
}

var runSyncOnce syncOnceExecutor = runSyncOnceScaffold

func executeSyncOnce(ctx context.Context, svc *FTPSyncService) (Result, error) {
	result := newSyncOnceResult(svc.opts)
	svc.log(fmt.Sprintf("SyncOnce start: %s", result.Direction))
	svc.reportEvent(SyncEvent{Operation: "sync_once", Path: result.SourceRoot, Status: "started"})

	err := runSyncOnce(ctx, svc, &result)
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

func runSyncOnceScaffold(ctx context.Context, svc *FTPSyncService, result *Result) error {
	if result.Direction == DirectionLocalToFTP {
		return runSyncOnceLocalToFTP(ctx, svc, result)
	}
	if result.Direction == DirectionFTPToLocal {
		return runSyncOnceFTPToLocal(ctx, svc, result)
	}

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
	return nil
}

func runSyncOnceFTPToLocal(ctx context.Context, svc *FTPSyncService, result *Result) error {
	destinationRoot, err := filepath.Abs(svc.opts.Destination.LocalPath)
	if err != nil {
		return err
	}
	if strings.TrimSpace(svc.opts.Destination.LocalPath) == "" {
		return newTransferError("SyncOnce destination local path is required", errInvalidOptions)
	}
	if err := os.MkdirAll(destinationRoot, fs.ModePerm); err != nil {
		return err
	}

	client, err := openFTPClient(ctx, svc.opts.Source.FTP)
	if err != nil {
		return err
	}
	defer client.close()

	remoteRoot := cleanFTPPath(svc.opts.Source.FTP.RemotePath)
	if err := ensureTargetUnderRoot(destinationRoot, remoteRoot, remoteRoot); err != nil {
		return err
	}

	pathIgnore, err := newSyncOnceIgnore(svc.opts.IgnoreRules)
	if err != nil {
		return err
	}

	var failureErrs []error
	err = client.walk(remoteRoot, func(currentPath string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			failureErrs = append(failureErrs, walkErr)
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if pathIgnore != nil && pathIgnore.MatchPath(currentPath, "ftp pull client sync", "sync once") {
			return nil
		}

		if err := ensureTargetUnderRoot(destinationRoot, remoteRoot, currentPath); err != nil {
			result.FailureCount++
			result.Partial = true
			failureErrs = append(failureErrs, err)
			svc.reportEvent(SyncEvent{Operation: "write", Path: currentPath, Status: "failed", ErrorKind: ErrTransfer})
			return nil
		}

		result.PathsVisited++
		opName := "create"
		var opErr error
		isSymlink := d.Type()&os.ModeSymlink != 0
		target := localTargetPath(destinationRoot, remoteRoot, currentPath)

		if d.IsDir() {
			result.DirectoriesAttempted++
			opErr = os.MkdirAll(target, 0o755)
		} else if isSymlink {
			result.FilesAttempted++
			opName = "symlink"
			opErr = newTransferError("SyncOnce FTP symlink pull is unsupported", errInvalidOptions)
		} else {
			result.FilesAttempted++
			opName = "write"
			opErr = client.readFile(currentPath, target)
		}

		if opErr != nil {
			result.FailureCount++
			result.Partial = true
			failureErrs = append(failureErrs, fmt.Errorf("%s %s: %w", opName, currentPath, opErr))
			svc.reportEvent(SyncEvent{Operation: opName, Path: currentPath, Status: "failed", ErrorKind: ErrTransfer})
			return nil
		}

		svc.reportEvent(SyncEvent{Operation: opName, Path: currentPath, Status: "complete"})
		return nil
	})
	if err != nil {
		if err == context.Canceled {
			return err
		}
		failureErrs = append(failureErrs, err)
	}

	if len(failureErrs) > 0 && result.PathsVisited == 0 {
		result.PathsVisited = 1
		result.FilesAttempted = 1
	}

	if len(failureErrs) > 0 {
		result.Partial = true
		return errors.Join(failureErrs...)
	}
	return nil
}

func runSyncOnceLocalToFTP(ctx context.Context, svc *FTPSyncService, result *Result) error {
	client, err := openFTPClient(ctx, svc.opts.Destination.FTP)
	if err != nil {
		return err
	}
	defer client.close()

	sourceRoot, err := filepath.Abs(svc.opts.Source.LocalPath)
	if err != nil {
		return err
	}
	remoteRoot := cleanFTPPath(svc.opts.Destination.FTP.RemotePath)
	pathIgnore, err := newSyncOnceIgnore(svc.opts.IgnoreRules)
	if err != nil {
		return err
	}
	seenRemotePaths := map[string]struct{}{}

	var failureErrs []error
	err = filepath.WalkDir(sourceRoot, func(currentPath string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			failureErrs = append(failureErrs, walkErr)
			result.FailureCount++
			result.Partial = true
			svc.reportEvent(SyncEvent{Operation: "walk", Path: currentPath, Status: "failed", ErrorKind: ErrTransfer})
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		relPath, err := filepath.Rel(sourceRoot, currentPath)
		if err != nil {
			failureErrs = append(failureErrs, err)
			return nil
		}
		if relPath == "." {
			relPath = ""
		}
		normalizedRel := normalizeIgnorePath(relPath)
		if pathIgnore != nil && pathIgnore.MatchPath(normalizedRel, "ftp push client sync", "sync once") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		remotePath := joinFTPPath(remoteRoot, filepath.ToSlash(relPath))
		if relPath == "" {
			remotePath = remoteRoot
		}
		seenRemotePaths[cleanFTPPath(remotePath)] = struct{}{}
		result.PathsVisited++
		opName := "create"
		var opErr error
		var bytesTotal int64

		isSymlink := d.Type()&os.ModeSymlink != 0
		switch {
		case d.IsDir():
			result.DirectoriesAttempted++
			opErr = retryWithContext(ctx, svc.opts.Retry, "ftp mkdir", func() error { return client.mkdirAll(remotePath) })
		case isSymlink:
			result.FilesAttempted++
			opName = "symlink"
			opErr = newTransferError("SyncOnce FTP symlink push is unsupported", errInvalidOptions)
		default:
			result.FilesAttempted++
			if info, infoErr := d.Info(); infoErr == nil {
				bytesTotal = info.Size()
			}
			if opErr = retryWithContext(ctx, svc.opts.Retry, "ftp mkdir", func() error { return client.mkdirAll(path.Dir(remotePath)) }); opErr == nil {
				opName = "write"
				opErr = retryWithContext(ctx, svc.opts.Retry, "ftp write", func() error { return client.writeFile(remotePath, currentPath) })
			}
		}

		if opErr != nil {
			result.FailureCount++
			result.Partial = true
			failureErrs = append(failureErrs, fmt.Errorf("%s %s: %w", opName, currentPath, opErr))
			svc.log(fmt.Sprintf("SyncOnce %s failed: %s", opName, currentPath))
			svc.reportEvent(SyncEvent{Operation: opName, Path: currentPath, Status: "failed", ErrorKind: ErrTransfer})
			return nil
		}

		svc.reportEvent(SyncEvent{Operation: opName, Path: currentPath, Status: "complete"})
		svc.reportProgress(Progress{
			Path:             currentPath,
			BytesTransferred: bytesTotal,
			BytesTotal:       bytesTotal,
			FilesTransferred: int64(result.FilesAttempted - result.FailureCount),
			FilesTotal:       int64(result.FilesAttempted),
		})
		return nil
	})
	if err != nil {
		return err
	}
	if err := removeRemotePathsMissingLocally(ctx, svc, client, sourceRoot, remoteRoot, pathIgnore, seenRemotePaths, result, &failureErrs); err != nil {
		return err
	}
	if len(failureErrs) > 0 {
		return errors.Join(failureErrs...)
	}
	return nil
}

func removeRemotePathsMissingLocally(ctx context.Context, svc *FTPSyncService, client ftpCore, sourceRoot string, remoteRoot string, pathIgnore *syncOnceIgnore, seenRemotePaths map[string]struct{}, result *Result, failureErrs *[]error) error {
	var removeTargets []string
	err := client.walk(remoteRoot, func(currentPath string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			*failureErrs = append(*failureErrs, walkErr)
			result.FailureCount++
			result.Partial = true
			svc.reportEvent(SyncEvent{Operation: "walk", Path: currentPath, Status: "failed", ErrorKind: ErrTransfer})
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		cleanRemotePath := cleanFTPPath(currentPath)
		if cleanRemotePath == cleanFTPPath(remoteRoot) {
			return nil
		}
		if _, ok := seenRemotePaths[cleanRemotePath]; ok {
			return nil
		}

		relPath := remoteRelativePath(remoteRoot, currentPath)
		if pathIgnore != nil && pathIgnore.MatchPath(relPath, "ftp push client sync", "sync once") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		localPath := filepath.Join(sourceRoot, filepath.FromSlash(relPath))
		if _, err := os.Lstat(localPath); err == nil {
			return nil
		} else if !errors.Is(err, os.ErrNotExist) {
			*failureErrs = append(*failureErrs, err)
			result.FailureCount++
			result.Partial = true
			svc.reportEvent(SyncEvent{Operation: "stat", Path: localPath, Status: "failed", ErrorKind: ErrTransfer})
			return nil
		}

		removeTargets = append(removeTargets, currentPath)
		return nil
	})
	if err != nil {
		return err
	}

	for _, remotePath := range removeTargets {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		result.PathsVisited++
		if err := retryWithContext(ctx, svc.opts.Retry, "ftp remove", func() error { return client.remove(remotePath) }); err != nil {
			result.FailureCount++
			result.Partial = true
			*failureErrs = append(*failureErrs, fmt.Errorf("remove %s: %w", remotePath, err))
			svc.log(fmt.Sprintf("SyncOnce remove failed: %s", remotePath))
			svc.reportEvent(SyncEvent{Operation: "remove", Path: remotePath, Status: "failed", ErrorKind: ErrTransfer})
			continue
		}
		svc.reportEvent(SyncEvent{Operation: "remove", Path: remotePath, Status: "complete"})
	}
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

func newSyncOnceIgnore(rules []IgnoreRule) (*syncOnceIgnore, error) {
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
	cleaned := strings.ReplaceAll(strings.TrimSpace(value), `\`, "/")
	cleaned = strings.TrimPrefix(cleaned, "./")
	cleaned = strings.TrimPrefix(cleaned, "/")
	if cleaned == "" {
		return ""
	}
	return strings.Trim(path.Clean(cleaned), "/")
}

func ensureTargetUnderRoot(root string, remoteRoot string, remotePath string) error {
	cleanRoot := filepath.Clean(root)
	target := localTargetPath(cleanRoot, remoteRoot, remotePath)
	rel, err := filepath.Rel(cleanRoot, target)
	if err != nil {
		return err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return newTransferError("SyncOnce destination path escaped configured root", errInvalidOptions)
	}
	if runtime.GOOS == "windows" && filepath.VolumeName(target) != filepath.VolumeName(cleanRoot) {
		return newTransferError("SyncOnce destination path escaped configured root", errInvalidOptions)
	}
	return nil
}

func localTargetPath(root string, remoteRoot string, remotePath string) string {
	cleanRemoteRoot := cleanFTPPath(remoteRoot)
	cleanRemotePath := cleanFTPPath(remotePath)
	relative := strings.TrimPrefix(cleanRemotePath, cleanRemoteRoot)
	relative = strings.TrimPrefix(relative, "/")
	return filepath.Join(filepath.Clean(root), filepath.FromSlash(relative))
}

func remoteRelativePath(remoteRoot string, remotePath string) string {
	cleanRemoteRoot := cleanFTPPath(remoteRoot)
	cleanRemotePath := cleanFTPPath(remotePath)
	relative := strings.TrimPrefix(cleanRemotePath, cleanRemoteRoot)
	return strings.TrimPrefix(relative, "/")
}
