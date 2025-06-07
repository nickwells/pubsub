package main

import "log/slog"

// connID is a type representing a connection ID.
type connID int64

// Attr returns a slog Attr representing a connection ID
func (cID connID) Attr() slog.Attr {
	return slog.Int64(cltAttrPfx+"connID", int64(cID))
}
