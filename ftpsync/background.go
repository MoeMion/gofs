package ftpsync

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

const backgroundDebounceDelay = 100 * time.Millisecond

type backgroundHandle struct {
	cancel context.CancelFunc
	done   chan struct{}
	ready  chan struct{}

	stopOnce sync.Once
	workers  sync.WaitGroup
	mu       sync.Mutex
	current  error
	final    error
}

func executeStartBackground(ctx context.Context, svc *FTPSyncService) (Handle, error) {
	runnerCtx, cancel := context.WithCancel(ctx)
	h := &backgroundHandle{
		cancel: cancel,
		done:   make(chan struct{}),
		ready:  make(chan struct{}),
	}
	go h.run(runnerCtx, svc)
	return h, nil
}

func (h *backgroundHandle) Done() <-chan struct{} {
	return h.done
}

func (h *backgroundHandle) Ready() <-chan struct{} {
	return h.ready
}

func (h *backgroundHandle) Err() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.final != nil {
		return h.final
	}
	return h.current
}

func (h *backgroundHandle) Wait() error {
	<-h.done
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.final
}

func (h *backgroundHandle) Stop(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	h.stopOnce.Do(h.cancel)
	select {
	case <-h.done:
		return h.Wait()
	case <-ctx.Done():
		return newError(ErrCanceled, "StartBackground stop context canceled", ctx.Err())
	}
}

func (h *backgroundHandle) run(ctx context.Context, svc *FTPSyncService) {
	defer func() {
		h.cancel()
		h.workers.Wait()
		close(h.done)
	}()
	h.runInitialSync(ctx, svc)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		h.setFinal(newTransferError("StartBackground watcher construction failed", err))
		close(h.ready)
		return
	}
	defer watcher.Close()

	sourceRoot, err := filepath.Abs(svc.opts.Source.LocalPath)
	if err != nil {
		h.setFinal(newTransferError("StartBackground source path resolution failed", err))
		close(h.ready)
		return
	}
	if err := h.watchTree(watcher, sourceRoot); err != nil {
		h.setFinal(newTransferError("StartBackground source watcher registration failed", err))
		close(h.ready)
		return
	}

	trigger := make(chan struct{}, 1)
	h.workers.Add(1)
	go func() {
		defer h.workers.Done()
		h.runSyncTriggers(ctx, svc, trigger)
	}()
	close(h.ready)
	h.runWatchLoop(ctx, watcher, trigger)
	if err := ctx.Err(); err != nil && err != context.Canceled {
		h.setFinal(newError(ErrCanceled, "StartBackground context canceled", err))
	}
}

func (h *backgroundHandle) runInitialSync(ctx context.Context, svc *FTPSyncService) {
	if _, err := executeSyncOnce(ctx, svc); err != nil {
		svc.log("StartBackground initial sync failed: " + err.Error())
		svc.reportEvent(SyncEvent{Operation: "background_sync", Path: svc.opts.Source.LocalPath, Status: "failed", ErrorKind: backgroundErrorKind(err)})
		h.setCurrent(err)
	}
}

func (h *backgroundHandle) runWatchLoop(ctx context.Context, watcher *fsnotify.Watcher, trigger chan<- struct{}) {
	timer := time.NewTimer(backgroundDebounceDelay)
	if !timer.Stop() {
		<-timer.C
	}
	dirty := false
	for {
		select {
		case <-ctx.Done():
			if dirty && !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			return
		case event, ok := <-watcher.Events:
			if !ok {
				h.setFinal(newTransferError("StartBackground watcher event stream closed", nil))
				return
			}
			if !h.isDirtyEvent(event) {
				continue
			}
			if event.Op&fsnotify.Create == fsnotify.Create {
				h.watchCreatedDirectory(watcher, event.Name)
			}
			dirty = true
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(backgroundDebounceDelay)
		case err, ok := <-watcher.Errors:
			if !ok {
				h.setFinal(newTransferError("StartBackground watcher error stream closed", nil))
				return
			}
			h.setCurrent(newTransferError("StartBackground watcher error", err))
		case <-timer.C:
			if dirty {
				h.queueSync(trigger)
				dirty = false
			}
		}
	}
}

func (h *backgroundHandle) runSyncTriggers(ctx context.Context, svc *FTPSyncService, trigger <-chan struct{}) {
	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-trigger:
			if !ok {
				return
			}
			if _, err := executeSyncOnce(ctx, svc); err != nil {
				svc.log("StartBackground sync pass failed: " + err.Error())
				svc.reportEvent(SyncEvent{Operation: "background_sync", Path: svc.opts.Source.LocalPath, Status: "failed", ErrorKind: backgroundErrorKind(err)})
				h.setCurrent(err)
			}
		}
	}
}

func backgroundErrorKind(err error) ErrorKind {
	for _, kind := range []ErrorKind{ErrAuthentication, ErrConnection, ErrTransfer, ErrCanceled, ErrValidation, ErrUnsupportedCapability} {
		if IsKind(err, kind) {
			return kind
		}
	}
	return ErrTransfer
}

func (h *backgroundHandle) watchTree(watcher *fsnotify.Watcher, root string) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			_ = watcher.Remove(path)
			if err := watcher.Add(path); err != nil {
				return err
			}
		}
		return nil
	})
}

func (h *backgroundHandle) watchCreatedDirectory(watcher *fsnotify.Watcher, path string) {
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return
	}
	_ = h.watchTree(watcher, path)
}

func (h *backgroundHandle) isDirtyEvent(event fsnotify.Event) bool {
	return event.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Remove|fsnotify.Rename) != 0
}

func (h *backgroundHandle) queueSync(trigger chan<- struct{}) {
	select {
	case trigger <- struct{}{}:
	default:
	}
}

func (h *backgroundHandle) setCurrent(err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.current = err
}

func (h *backgroundHandle) setFinal(err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.final = err
}
