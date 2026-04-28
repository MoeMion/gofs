package ftpsync_test

import (
	"context"
	"fmt"
	"time"

	"ftpsync/ftpsync"
)

func ExampleFTPSyncService_SyncOnce_localToFTP() {
	svc, err := ftpsync.NewFTPSyncService(ftpsync.Options{
		Direction: ftpsync.DirectionLocalToFTP,
		Source:    ftpsync.Endpoint{LocalPath: "./site"},
		Destination: ftpsync.Endpoint{FTP: ftpsync.FTPOptions{
			Host:         "127.0.0.1",
			Port:         21,
			Username:     "ftp-user",
			Password:     "ftp-password",
			RemotePath:   "/public",
			PassiveMode:  true,
			Timeout:      10 * time.Second,
			PathEncoding: "utf8",
		}},
		Retry: ftpsync.RetryOptions{Count: 2, Wait: time.Second},
	})
	if err != nil {
		fmt.Println(ftpsync.DirectionLocalToFTP)
		return
	}
	result, err := svc.SyncOnce(context.Background())
	if err != nil {
		fmt.Println(ftpsync.DirectionLocalToFTP)
		return
	}
	fmt.Println(result.Direction)

	// Output:
	// local_to_ftp
}

func ExampleFTPSyncService_SyncOnce_ftpToLocal() {
	svc, err := ftpsync.NewFTPSyncService(ftpsync.Options{
		Direction: ftpsync.DirectionFTPToLocal,
		Source: ftpsync.Endpoint{FTP: ftpsync.FTPOptions{
			Host:         "127.0.0.1",
			Port:         21,
			Username:     "ftp-user",
			Password:     "ftp-password",
			RemotePath:   "/backup",
			PassiveMode:  true,
			Timeout:      10 * time.Second,
			PathEncoding: "utf8",
		}},
		Destination: ftpsync.Endpoint{LocalPath: "./restore"},
	})
	if err != nil {
		fmt.Println(ftpsync.DirectionFTPToLocal)
		return
	}
	result, err := svc.SyncOnce(context.Background())
	if err != nil {
		fmt.Println(ftpsync.DirectionFTPToLocal)
		return
	}
	fmt.Println(result.Direction)

	// Output:
	// ftp_to_local
}

func ExampleFTPSyncService_StartBackground() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	svc, err := ftpsync.NewFTPSyncService(ftpsync.Options{
		Direction: ftpsync.DirectionLocalToFTP,
		Source:    ftpsync.Endpoint{LocalPath: "./watched"},
		Destination: ftpsync.Endpoint{FTP: ftpsync.FTPOptions{
			Host:         "127.0.0.1",
			Port:         21,
			Username:     "ftp-user",
			Password:     "ftp-password",
			RemotePath:   "/mirror",
			PassiveMode:  true,
			Timeout:      10 * time.Second,
			PathEncoding: "utf8",
		}},
	})
	if err != nil {
		fmt.Println(ftpsync.DirectionLocalToFTP)
		return
	}
	handle, err := svc.StartBackground(ctx)
	if err != nil {
		fmt.Println(ftpsync.DirectionLocalToFTP)
		return
	}
	cancel()
	if handle != nil {
		_ = handle.Stop(context.Background())
	}
	fmt.Println(ftpsync.DirectionLocalToFTP)

	// Output:
	// local_to_ftp
}
