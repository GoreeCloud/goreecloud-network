package dialer

import "errors"

// ErrTransportUnavailable reports that the selected relay transport is not
// offered by the endpoint. Callers may use this to distinguish an explicit
// capability absence from a transport or network failure without exposing
// endpoint details.
var ErrTransportUnavailable = errors.New("relay transport unavailable")
