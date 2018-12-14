package pactelnet

import "bytes"
import "github.com/yourbasic/bit"

type telnet struct {
	telOpts  []TelnetOptionReq
	flags    *bit.Set
	userData interface{}
	state    telnetState
	//internalFlags []telnetInternalFlags
	internalFlags *bit.Set
	sb_telopt     byte
	buffer        *bytes.Buffer
	rfc1143List   []TelnetRFC1143
	OnTelnetEvent func(telnetEvent telnetEventInterface)
}

func NewTelnet(options []TelnetOptionReq, flags []TelnetFlags, userData interface{}) *telnet {
	tl := new(telnet)
	tl.telOpts = options
	tl.flags = new(bit.Set)
	for _, v := range flags {
		tl.flags.Add(int(v))
	}
	tl.userData = userData

	tl.state = TELNET_STATE_DATA
	//tl.internalFlags = make([]telnetInternalFlags, 0)
	tl.internalFlags = new(bit.Set)
	tl.buffer = bytes.NewBuffer(make([]byte, 0, 512))
	tl.rfc1143List = make([]TelnetRFC1143, 0)

	return tl
}

//------------------------------------------------------------------------------------------------//

func (tl *telnet) TelnetRecv(buffer []byte) {
	tl.process(buffer)
}

//------------------------------------------------------------------------------------------------//

// Send negotiation
func (tl *telnet) TelnetNegotiate(cmd TelnetCommands, telopt byte) {
	// if we're in proxy mode, just send it now
	if tl.flags.Contains(int(TELNET_FLAG_PROXY)) {
		data := []byte{TELNET_IAC, byte(cmd), telopt}
		tl.send(data)
		return
	}

	// get current option states
	q := tl.getRFC1143(telopt)

	switch cmd {
	// advertise willingess to support an option
	case TELNET_WILL:
		switch q_US(q) {
		case byte(Q_NO):
			tl.setRFC1143(telopt, byte(Q_WANTYES), q_HIM(q))
			tl.sendNegotiate(TELNET_WILL, telopt)
			break
		case byte(Q_WANTNO):
			tl.setRFC1143(telopt, byte(Q_WANTNO_OP), q_HIM(q))
			break
		case byte(Q_WANTYES_OP):
			tl.setRFC1143(telopt, byte(Q_WANTYES), q_HIM(q))
			break
		}
		break
	// force turn-off of locally enabled option
	case TELNET_WONT:
		switch q_US(q) {
		case byte(Q_YES):
			tl.setRFC1143(telopt, byte(Q_WANTNO), q_HIM(q))
			tl.sendNegotiate(TELNET_WONT, telopt)
			break
		case byte(Q_WANTYES):
			tl.setRFC1143(telopt, byte(Q_WANTYES_OP), q_HIM(q))
			break
		case byte(Q_WANTNO_OP):
			tl.setRFC1143(telopt, byte(Q_WANTNO), q_HIM(q))
			break
		}
		break

	// ask remote end to enable an option
	case TELNET_DO:
		switch q_HIM(q) {
		case byte(Q_NO):
			tl.setRFC1143(telopt, q_US(q), byte(Q_WANTYES))
			tl.sendNegotiate(TELNET_DO, telopt)
			break
		case byte(Q_WANTNO):
			tl.setRFC1143(telopt, q_US(q), byte(Q_WANTNO_OP))
			break
		case byte(Q_WANTYES_OP):
			tl.setRFC1143(telopt, q_US(q), byte(Q_WANTYES))
			break
		}
		break

	// demand remote end disable an option
	case TELNET_DONT:
		switch q_HIM(q) {
		case byte(Q_YES):
			tl.setRFC1143(telopt, q_US(q), byte(Q_WANTNO))
			tl.sendNegotiate(TELNET_DONT, telopt)
			break
		case byte(Q_WANTYES):
			tl.setRFC1143(telopt, q_US(q), byte(Q_WANTYES_OP))
			break
		case byte(Q_WANTNO_OP):
			tl.setRFC1143(telopt, q_US(q), byte(Q_WANTNO))
			break
		}
		break
	}
}

//------------------------------------------------------------------------------------------------//

// Send an iac command
func (tl *telnet) TelnetIAC(cmd byte) {
	data := []byte{TELNET_IAC, byte(cmd)}
	tl.send(data)
}

//------------------------------------------------------------------------------------------------//

// Send non-command data (escapes IAC bytes)
func (tl *telnet) TelnetSend(buffer []byte) {
	var ln, i int
	for i, _ = range buffer {
		// dump prior portion of text, send escaped bytes
		if buffer[i] == byte(TELNET_IAC) {
			// dump prior text if any
			if i != ln {
				tl.send(buffer[:ln])
			}
			ln = i + 1

			// send escape
			tl.TelnetIAC(byte(TELNET_IAC))
		}
	}

	// send whatever portion of buffer is left
	if i != ln {
		tl.send(buffer[:ln])
	}
}

//------------------------------------------------------------------------------------------------//

func (tl *telnet) TelnetSendText(buffer []byte) {
	var ln, i int

	for i, _ = range buffer {
		// dump prior portion of text, send escaped bytes
		if buffer[i] == byte(TELNET_IAC) {
			// dump prior text if any
			if i != ln {
				tl.send(buffer[:ln])
			}
			ln = i + 1

			// send escape
			tl.TelnetIAC(byte(TELNET_IAC))
		} else if tl.internalFlags.Contains(int(TELNET_FLAG_TRANSMIT_BINARY)) && (buffer[i] == '\r' || buffer[i] == '\n') {
			// dump prior portion of text
			if i != ln {
				tl.send(buffer[:ln])
			}
			ln = i + 1

			// automatic translation of \r -> CRNUL
			if buffer[i] == '\r' {
				tl.send(CRNUL)
			} else {
				// automatic translation of \n -> CRLF
				tl.send(CRLF)
			}
		}
	} //for

	// send whatever portion of buffer is left
	if i != ln {
		tl.send(buffer[:ln])
	}
}

//-------------------------------Private functions------------------------------------------------//

func (tl *telnet) callEventHandler(telnetEvent telnetEventInterface) {
	if tl.OnTelnetEvent != nil {
		telnetEvent.SetUserData(tl.userData)
		tl.OnTelnetEvent(telnetEvent)
	}
}

//------------------------------------------------------------------------------------------------//

func (tl *telnet) send(buffer []byte) {
	ev := NewTelnetSendEvent()
	ev.Buffer = buffer
	tl.callEventHandler(ev)
}

//------------------------------------------------------------------------------------------------//

func (tl *telnet) process(buffer []byte) {

}

//------------------------------------------------------------------------------------------------//

// Retrieve RFC1143 option state
func (tl *telnet) getRFC1143(telopt byte) TelnetRFC1143 {

	// search for entry
	for _, v := range tl.rfc1143List {
		if v.telopt == telopt {
			return v
		}
	}

	// not found, return empty value
	return TelnetRFC1143{telopt: telopt, state: 0}
}

//------------------------------------------------------------------------------------------------//

// Save RFC1143 option state
func (tl *telnet) setRFC1143(telopt byte, us byte, him byte) {
	var qtmp TelnetRFC1143

	// search for entry
	for i, _ := range tl.rfc1143List {
		if tl.rfc1143List[i].telopt == telopt {
			qtmp = TelnetRFC1143{state: q_MAKE(us, him), telopt: telopt}
			tl.rfc1143List[i] = qtmp
			if telopt != byte(TELOPT_BINARY) {
				return
			}
			tl.internalFlags.Delete(int(TELNET_FLAG_TRANSMIT_BINARY))
			tl.internalFlags.Delete(int(TELNET_FLAG_RECEIVE_BINARY))
			if us == byte(Q_YES) {
				tl.internalFlags.Add(int(TELNET_FLAG_TRANSMIT_BINARY))
			}
			if him == byte(Q_YES) {
				tl.internalFlags.Add(int(TELNET_FLAG_RECEIVE_BINARY))
			}
			return
		}
	}

	qtmp.telopt = telopt
	qtmp.state = q_MAKE(us, him)
	tl.rfc1143List = append(tl.rfc1143List, qtmp)
}

//------------------------------------------------------------------------------------------------//

// Send negotiation bytes
func (tl *telnet) sendNegotiate(cmd TelnetCommands, telopt byte) {
	data := []byte{TELNET_IAC, byte(cmd), telopt}
	tl.send(data)
}
