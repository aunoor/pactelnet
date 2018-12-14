package pactelnet

import "bytes"

type telnet struct {
	telOpts       []TelnetOptionReq
	flags         []TelnetFlags
	userData      interface{}
	state         telnetState
	internalFlags []telnetInternalFlags
	sb_telopt     byte
	buffer        bytes.Buffer
	_rfc1143List
}

func NewTelnet() *telnet {
	tl := new(telnet)
	return tl
}
