package ftpsync

import (
	"context"
	"sync"
)

type backgroundHandle struct {
	cancel context.CancelFunc
	done   chan struct{}
	ready  chan struct{}

	stopOnce sync.Once
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
	return h.Err()
}

func (h *backgroundHandle) Stop(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	h.stopOnce.Do(h.cancel)
	select {
	case <-h.done:
		return h.Err()
	case <-ctx.Done():
		return newError(ErrCanceled, "StartBackground stop context canceled", ctx.Err())
	}
}

func (h *backgroundHandle) run(ctx context.Context, svc *FTPSyncService) {
	defer close(h.done)
	if _, err := executeSyncOnce(ctx, svc); err != nil {
		h.setCurrent(err)
	}
	close(h.ready)
	<-ctx.Done()
	if err := ctx.Err(); err != nil && err != context.Canceled {
		h.setFinal(newError(ErrCanceled, "StartBackground context canceled", err))
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
