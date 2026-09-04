package control

import (
	"errors"
	"runtime"
)

var ErrProtectedFileStoreUnsupported = errors.New("conduit control: owner-only protected file persistence is unsupported on this platform")

// requireProtectedFileStoreSupport prevents durable Conduit control-state stores
// from claiming owner-only filesystem protection on platforms where Go's Unix
// permission bits do not provide the required security contract.
func requireProtectedFileStoreSupport() error {
	if runtime.GOOS == "windows" {
		return ErrProtectedFileStoreUnsupported
	}
	return nil
}
