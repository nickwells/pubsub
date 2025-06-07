package main

import (
	"fmt"
	"log/slog"
	"net"
)

const (
	svrAttrPfx = "Svr-"
	cltAttrPfx = "Clt-"
)

// listeningPortAttr returns a slog.Attr for the listening port
func listeningPortAttr(port int) slog.Attr {
	return slog.String(svrAttrPfx+"Listening-Port", fmt.Sprintf("%d", port))
}

// progNameAttr returns a slog.Attr for the program name
func progNameAttr(name string) slog.Attr {
	return slog.String(svrAttrPfx+"Program-Name", name)
}

// netAddrAttr returns a slog.Attr for the network connection address
func netAddrAttr(conn net.Conn) slog.Attr {
	return slog.String(cltAttrPfx+"Net-Address", conn.RemoteAddr().String())
}
