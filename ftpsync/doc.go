// Package ftpsync exposes a focused local Go module package for plain FTP file
// synchronization through FTPSyncService.
//
// While source remains in the ftpsync/ subdirectory under module ftpsync,
// consumers import the package as:
//
//	import "ftpsync/ftpsync"
//
// Callers configure typed Options, Endpoint, and FTPOptions values, then call
// SyncOnce for local-to-FTP or FTP-to-local one-shot sync, or StartBackground
// for local disk-to-FTP background sync.
package ftpsync
