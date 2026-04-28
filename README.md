# ftpsync

`ftpsync` is a focused local Go module for plain FTP file synchronization through a typed Go API. It is no longer the old `gofs` CLI/server runtime; v2.0 keeps the FTP sync library and removes the application surfaces that are not required by embedders.

The module path is `ftpsync`. Because package source currently remains in the `ftpsync/` subdirectory, consumer code imports the package as:

```go
import "ftpsync/ftpsync"
```

## Install as a local module

Embed this repository as local source and use a `replace` directive from your application module:

```go
module example.com/my-app

go 1.24

require ftpsync v0.0.0

replace ftpsync => ../path/to/ftpsync
```

Then import the package from its current subdirectory package path:

```go
import "ftpsync/ftpsync"
```

## One-shot local to FTP

```go
package main

import (
	"context"
	"log"
	"time"

	"ftpsync/ftpsync"
)

func main() {
	svc, err := ftpsync.NewFTPSyncService(ftpsync.Options{
		Direction: ftpsync.DirectionLocalToFTP,
		Source:    ftpsync.Endpoint{LocalPath: "./site"},
		Destination: ftpsync.Endpoint{FTP: ftpsync.FTPOptions{
			Host:         "ftp.example.test",
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
		log.Fatal(err)
	}

	result, err := svc.SyncOnce(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("synced %d files", result.FilesAttempted)
}
```

## One-shot FTP to local

```go
package main

import (
	"context"
	"log"
	"time"

	"ftpsync/ftpsync"
)

func main() {
	svc, err := ftpsync.NewFTPSyncService(ftpsync.Options{
		Direction: ftpsync.DirectionFTPToLocal,
		Source: ftpsync.Endpoint{FTP: ftpsync.FTPOptions{
			Host:         "ftp.example.test",
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
		log.Fatal(err)
	}

	result, err := svc.SyncOnce(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("downloaded %d files", result.FilesAttempted)
}
```

## Background local to FTP

```go
package main

import (
	"context"
	"log"
	"time"

	"ftpsync/ftpsync"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	svc, err := ftpsync.NewFTPSyncService(ftpsync.Options{
		Direction: ftpsync.DirectionLocalToFTP,
		Source:    ftpsync.Endpoint{LocalPath: "./watched"},
		Destination: ftpsync.Endpoint{FTP: ftpsync.FTPOptions{
			Host:         "ftp.example.test",
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
		log.Fatal(err)
	}

	handle, err := svc.StartBackground(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = handle.Stop(context.Background()) }()

	// Run your application. Stop the handle during shutdown.
	<-handle.Done()
}
```

## Supported package contract

- `ftpsync.NewFTPSyncService(opts)` constructs a service from typed options.
- `ftpsync.Options`, `ftpsync.Endpoint`, and `ftpsync.FTPOptions` configure local paths, FTP host, port, username, password, remote path, passive mode, timeout, and path encoding.
- `ftpsync.DirectionLocalToFTP` supports one-shot local→FTP sync and background local→FTP sync.
- `ftpsync.DirectionFTPToLocal` supports one-shot FTP→local sync.
- `ftpsync.RetryOptions`, `ftpsync.IgnoreRule`, and `ftpsync.HookSet` provide retry, filtering, logging, progress, and event hooks without CLI or YAML configuration.

## Limitations

- Plain FTP only; no FTPS in v2.0.
- Client-side sync only; no FTP server mode.
- No FTP<->FTP sync.
- No FTP->disk background polling.
- No bidirectional conflict resolution.
- No CLI runtime, YAML/CLI config parser, URL parser, HTTP/gRPC/file server, task/auth/session runtime, SFTP, MinIO, Docker release artifact, or old application daemon mode.

## Migration from old gofs runtime

v2.0 is a breaking migration from the old `gofs` CLI/server application to a local Go library module. See [MIGRATION.md](MIGRATION.md) for the removed surfaces, package import path, and capability mapping.

## Development verification

Run the full suite with:

```bash
go test ./...
```
