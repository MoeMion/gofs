package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ftpsync/ftpsync"
)

const (
	localSourcePath = "./source"
	ftpHost         = "127.0.0.1"
	ftpPort         = 21
	ftpUser         = "user"
	ftpPassword     = "pass"
	ftpRemotePath   = "/mirror"
	ftpPassiveMode  = true
	ftpTimeout      = 10 * time.Second
	shutdownTimeout = 10 * time.Second
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "ftp background sync failed: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if err := os.MkdirAll(localSourcePath, 0o755); err != nil {
		return fmt.Errorf("create local source directory: %w", err)
	}

	ctx, stopSignals := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopSignals()

	svc, err := ftpsync.NewFTPSyncService(ftpsync.Options{
		Direction: ftpsync.DirectionLocalToFTP,
		Source:    ftpsync.Endpoint{LocalPath: localSourcePath},
		Destination: ftpsync.Endpoint{FTP: ftpsync.FTPOptions{
			Host:         ftpHost,
			Port:         ftpPort,
			Username:     ftpUser,
			Password:     ftpPassword,
			RemotePath:   ftpRemotePath,
			PassiveMode:  ftpPassiveMode,
			Timeout:      ftpTimeout,
			PathEncoding: "utf8",
		}},
		Retry: ftpsync.RetryOptions{Count: 2, Wait: time.Second},
		Hooks: ftpsync.HookSet{
			Logger: ftpsync.LoggerFunc(func(message string) {
				fmt.Fprintf(os.Stderr, "ftpsync: %s\n", message)
			}),
			Event: func(event ftpsync.SyncEvent) {
				if event.ErrorKind != "" {
					fmt.Fprintf(os.Stderr, "ftpsync event: operation=%s path=%s status=%s error_kind=%s\n", event.Operation, event.Path, event.Status, event.ErrorKind)
				}
			},
		},
	})
	if err != nil {
		return fmt.Errorf("create FTP sync service: %w", err)
	}

	handle, err := svc.StartBackground(ctx)
	if err != nil {
		return fmt.Errorf("start background sync: %w", err)
	}

	fmt.Printf("FTP background sync started: %s -> ftp://%s:%d%s\n", localSourcePath, ftpHost, ftpPort, ftpRemotePath)
	fmt.Println("Press Ctrl+C to stop.")

	select {
	case <-ctx.Done():
		fmt.Fprintln(os.Stderr, "shutdown signal received")
	case <-handle.Done():
		if err := handle.Wait(); err != nil {
			return fmt.Errorf("background sync stopped: %w", err)
		}
		if err := handle.Err(); err != nil {
			return fmt.Errorf("background sync stopped after error: %w", err)
		}
		return nil
	}

	stopCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := handle.Stop(stopCtx); err != nil {
		return fmt.Errorf("stop background sync: %w", err)
	}
	if err := handle.Wait(); err != nil {
		return fmt.Errorf("wait for background sync: %w", err)
	}
	if err := handle.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "last background sync error before shutdown: %v\n", err)
	}

	fmt.Println("FTP background sync stopped")
	return nil
}
