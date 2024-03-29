package pactelnet

import "bytes"
import "github.com/yourbasic/bit"

type Telnet struct {
	telOpts  []TelnetOptionReq
	flags    *bit.Set
	userData interface{}
	state    telnetState
	//internalFlags []telnetInternalFlags
	internalFlags *bit.Set
	sb_telopt     byte
	buffer        *bytes.Buffer
	rfc1143List   []TelnetRFC1143
	OnTelnetEvent func(telnetEvent TelnetEventInterface)
}

func NewTelnet(options []TelnetOptionReq, flags []TelnetFlags, userData interface{}) *Telnet {
	tl := new(Telnet)
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

func (tl *Telnet) TelnetRecv(buffer []byte) {
	tl.process(buffer)
}

//------------------------------------------------------------------------------------------------//

// Send negotiation
func (tl *Telnet) TelnetNegotiate(cmd TelnetCommands, telopt byte) {
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
		case byte(Q_WANTNO):
			tl.setRFC1143(telopt, byte(Q_WANTNO_OP), q_HIM(q))
		case byte(Q_WANTYES_OP):
			tl.setRFC1143(telopt, byte(Q_WANTYES), q_HIM(q))
		}

	// force turn-off of locally enabled option
	case TELNET_WONT:
		switch q_US(q) {
		case byte(Q_YES):
			tl.setRFC1143(telopt, byte(Q_WANTNO), q_HIM(q))
			tl.sendNegotiate(TELNET_WONT, telopt)
		case byte(Q_WANTYES):
			tl.setRFC1143(telopt, byte(Q_WANTYES_OP), q_HIM(q))
		case byte(Q_WANTNO_OP):
			tl.setRFC1143(telopt, byte(Q_WANTNO), q_HIM(q))
		}

	// ask remote end to enable an option
	case TELNET_DO:
		switch q_HIM(q) {
		case byte(Q_NO):
			tl.setRFC1143(telopt, q_US(q), byte(Q_WANTYES))
			tl.sendNegotiate(TELNET_DO, telopt)
		case byte(Q_WANTNO):
			tl.setRFC1143(telopt, q_US(q), byte(Q_WANTNO_OP))
		case byte(Q_WANTYES_OP):
			tl.setRFC1143(telopt, q_US(q), byte(Q_WANTYES))
		}

	// demand remote end disable an option
	case TELNET_DONT:
		switch q_HIM(q) {
		case byte(Q_YES):
			tl.setRFC1143(telopt, q_US(q), byte(Q_WANTNO))
			tl.sendNegotiate(TELNET_DONT, telopt)
		case byte(Q_WANTYES):
			tl.setRFC1143(telopt, q_US(q), byte(Q_WANTYES_OP))
		case byte(Q_WANTNO_OP):
			tl.setRFC1143(telopt, q_US(q), byte(Q_WANTNO))
		}
	}
}

//------------------------------------------------------------------------------------------------//

// Send an iac command
func (tl *Telnet) TelnetIAC(cmd byte) {
	data := []byte{TELNET_IAC, byte(cmd)}
	tl.send(data)
}

//------------------------------------------------------------------------------------------------//

// Send non-command data (escapes IAC bytes)
func (tl *Telnet) TelnetSend(buffer []byte) {
	var ln, i int
	for i, _ = range buffer {
		// dump prior portion of text, send escaped bytes
		if buffer[i] == byte(TELNET_IAC) {
			// dump prior text if any
			if i != ln {
				tl.send(buffer[ln:])
			}
			ln = i + 1

			// send escape
			tl.TelnetIAC(byte(TELNET_IAC))
		}
	}

	// send whatever portion of buffer is left
	if i != ln {
		tl.send(buffer[ln:])
	}
}

//------------------------------------------------------------------------------------------------//

func (tl *Telnet) TelnetSendText(buffer []byte) {
	var ln, i int

	for i, _ = range buffer {
		// dump prior portion of text, send escaped bytes
		if buffer[i] == byte(TELNET_IAC) {
			// dump prior text if any
			if i != ln {
				tl.send(buffer[ln:])
			}
			ln = i + 1

			// send escape
			tl.TelnetIAC(byte(TELNET_IAC))
		} else if tl.internalFlags.Contains(int(TELNET_FLAG_TRANSMIT_BINARY)) && (buffer[i] == '\r' || buffer[i] == '\n') {
			// dump prior portion of text
			if i != ln {
				tl.send(buffer[ln:])
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
		tl.send(buffer[ln:])
	}
}

//-------------------------------Private functions------------------------------------------------//

func (tl *Telnet) callEventHandler(telnetEvent TelnetEventInterface) {
	if tl.OnTelnetEvent != nil {
		telnetEvent.SetUserData(tl.userData)
		tl.OnTelnetEvent(telnetEvent)
	}
}

//------------------------------------------------------------------------------------------------//

func (tl *Telnet) send(buffer []byte) {
	ev := NewTelnetSendEvent()
	ev.Buffer = buffer
	tl.callEventHandler(ev)
}

//------------------------------------------------------------------------------------------------//

func (tl *Telnet) process(buffer []byte) {
	var i, start int
	var dataByte byte

	for i, dataByte = range buffer {
		switch tl.state {
		// regular data
		case TELNET_STATE_DATA:
			// on an IAC byte, pass through all pending bytes and switch states
			if dataByte == byte(TELNET_IAC) {
				if i != start {
					dataEvent := NewTelnetDataEvent()
					dataEvent.Buffer = buffer[start:i]
					tl.callEventHandler(dataEvent)
				}
				tl.state = TELNET_STATE_IAC
			} else if dataByte == '\r' && (tl.flags.Contains(TELNET_FLAG_NVT_EOL)) && !(tl.internalFlags.Contains(TELNET_FLAG_RECEIVE_BINARY)) {
				if i != start {
					dataEvent := NewTelnetDataEvent()
					dataEvent.Buffer = buffer[start:i]
					tl.callEventHandler(dataEvent)
				}
				tl.state = TELNET_STATE_EOL
			}

			// NVT EOL to be translated
		case TELNET_STATE_EOL:
			if dataByte != '\n' {
				dataByte = '\r'
				dataEvent := NewTelnetDataEvent()
				dataEvent.Buffer = []byte{dataByte}
				tl.callEventHandler(dataEvent)
				dataByte = buffer[i]
			}
			// any byte following '\r' other than '\n' or '\0' is invalid,
			// so pass both \r and the byte
			start = i
			if dataByte == 0 {
				start++
			}
			// state update
			tl.state = TELNET_STATE_DATA

			// IAC command
		case TELNET_STATE_IAC:
			switch dataByte {
			// subnegotiation
			case byte(TELNET_SB):
				tl.state = TELNET_STATE_SB
				// negotiation commands
			case byte(TELNET_WILL):
				tl.state = TELNET_STATE_WILL
			case byte(TELNET_WONT):
				tl.state = TELNET_STATE_WONT
			case byte(TELNET_DO):
				tl.state = TELNET_STATE_DO
			case byte(TELNET_DONT):
				tl.state = TELNET_STATE_DONT

				// IAC escaping
			case byte(TELNET_IAC):
				// event
				dataEvent := NewTelnetDataEvent()
				dataEvent.Buffer = []byte{dataByte}
				tl.callEventHandler(dataEvent)
				// state update
				start = i + 1
				tl.state = TELNET_STATE_DATA

				// some other command
			default:
				// event
				iacEvent := NewTelnetIacEvent()
				iacEvent.Cmd = TELNET_IAC
				tl.callEventHandler(iacEvent)
				// state update
				start = i + 1
				tl.state = TELNET_STATE_DATA
			}

			// negotiation commands
		case TELNET_STATE_WILL:
			fallthrough
		case TELNET_STATE_WONT:
			fallthrough
		case TELNET_STATE_DO:
			fallthrough
		case TELNET_STATE_DONT:
			tl.negotiate(dataByte)
			start = i + 1
			tl.state = TELNET_STATE_DATA

			// subnegotiation -- determine subnegotiation telopt
		case TELNET_STATE_SB:
			tl.sb_telopt = dataByte
			//_buffer.Position = 0;
			tl.buffer.Reset()
			tl.state = TELNET_STATE_SB_DATA

		// subnegotiation -- buffer bytes until end request
		case TELNET_STATE_SB_DATA:
			// IAC command in subnegotiation -- either IAC SE or IAC IAC
			if dataByte == byte(TELNET_IAC) {
				tl.state = TELNET_STATE_SB_DATA_IAC
			} else if tl.sb_telopt == byte(TELOPT_COMPRESS) && dataByte == byte(TELNET_WILL) {
				/* In 1998 MCCP used TELOPT 85 and the protocol defined an invalid
				 * subnegotiation sequence (IAC SB 85 WILL SE) to start compression.
				 * Subsequently MCCP version 2 was created in 2000 using TELOPT 86
				 * and a valid subnegotiation (IAC SB 86 IAC SE). libtelnet for now
				 * just captures and discards MCCPv1 sequences. */
				start = i + 2
				tl.state = TELNET_STATE_DATA
			} else {
				tl.buffer.Write([]byte{dataByte})
				start = i + 1
				tl.state = TELNET_STATE_DATA
			}

			// IAC escaping inside a subnegotiation
		case TELNET_STATE_SB_DATA_IAC:
			switch dataByte {
			// end subnegotiation
			case byte(TELNET_SE):
				// return to default state
				start = i + 1
				tl.state = TELNET_STATE_DATA

				// process subnegotiation
				if tl.subnegotiate() == true {
					/* any remaining bytes in the buffer are compressed.
					 * we have to re-invoke telnet_recv to get those
					 * bytes inflated and abort trying to process the
					 * remaining compressed bytes in the current _process
					 * buffer argument
					 */

					//byte[] tmp = new byte[buffer.Length - start];
					//buffer.CopyTo(tmp, start);
					//tmp := make([]byte, 0, len(buffer)- start)
					tmp := buffer[start:]
					tl.TelnetRecv(tmp)
					return
				}

			// escaped IAC byte
			case byte(TELNET_IAC):
				// push IAC into buffer
				/*
				   if (BufferByte((byte)TelnetCommands.TELNET_IAC) != TelnetErrorCode.TELNET_EOK)
				   {
				   start = i + 1;
				   _state = TelnetState.TELNET_STATE_DATA;
				   }
				   else
				   {
				   _state = TelnetState.TELNET_STATE_SB_DATA;
				   }
				*/
				tl.buffer.Write([]byte{byte(TELNET_IAC)})
				tl.state = TELNET_STATE_SB_DATA

			/* something else -- protocol error.  attempt to process
			 * content in subnegotiation buffer, then evaluate the
			 * given command as an IAC code.
			 */
			default:
				/* TODO:
				   _error(Telnet, __LINE__, __func__, TELNET_EPROTOCOL, 0,
					   "unexpected byte after IAC inside SB: %d",
					   dataByte);
				*/

				// enter IAC state
				start = i + 1
				tl.state = TELNET_STATE_IAC

				/* process subnegotiation; see comment in
				 * TELNET_STATE_SB_DATA_IAC about invoking telnet_recv()
				 */
				if tl.subnegotiate() == true {
					//				   byte[] tmp = new byte[buffer.Length - start];
					//				   buffer.CopyTo(tmp, start);
					tmp := buffer[start:]
					tl.TelnetRecv(tmp)
					return
				} else {
					/* recursive call to get the current input byte processed
					 * as a regular IAC command.  we could use a goto, but
					 * that would be gross.
					 */
					tl.process([]byte{dataByte})
				}

			}
		} //switch
	} //for

	// pass through any remaining bytes
	if tl.state == TELNET_STATE_DATA && start < len(buffer) {
		dataEvent := NewTelnetDataEvent()
		dataEvent.Buffer = buffer[start:]
		tl.callEventHandler(dataEvent)
	}

}

//------------------------------------------------------------------------------------------------//

// Negotiation handling magic for RFC1143
func (tl *Telnet) negotiate(telopt byte) {

	/* in PROXY mode, just pass it thru and do nothing */
	if tl.flags.Contains(int(TELNET_FLAG_PROXY)) {
		switch tl.state {
		case TELNET_STATE_WILL:
			tl.negotiateEvent(TELNET_EV_WILL, telopt)
		case TELNET_STATE_WONT:
			tl.negotiateEvent(TELNET_EV_WONT, telopt)
		case TELNET_STATE_DO:
			tl.negotiateEvent(TELNET_EV_DO, telopt)
		case TELNET_STATE_DONT:
			tl.negotiateEvent(TELNET_EV_DONT, telopt)
		}
		return
	}

	// lookup the current state of the option
	var q TelnetRFC1143 = tl.getRFC1143(telopt)

	// start processing...
	switch tl.state {
	// request to enable option on remote end or confirm DO
	case TELNET_STATE_WILL:
		switch q_HIM(q) {
		case byte(Q_NO):
			if tl.checkTelOpt(telopt, false) {
				tl.setRFC1143(telopt, q_US(q), byte(Q_YES))
				tl.sendNegotiate(TELNET_DO, telopt)
				tl.negotiateEvent(TELNET_EV_WILL, telopt)
			} else {
				tl.sendNegotiate(TELNET_DONT, telopt)
			}

		case byte(Q_WANTNO):
			tl.setRFC1143(telopt, q_US(q), byte(Q_NO))
			tl.negotiateEvent(TELNET_EV_WONT, telopt)
			//_error(Telnet, __LINE__, __func__, TELNET_EPROTOCOL, 0, "DONT answered by WILL");

		case byte(Q_WANTNO_OP):
			tl.setRFC1143(telopt, q_US(q), byte(Q_YES))
			tl.negotiateEvent(TELNET_EV_WILL, telopt)
			//_error(Telnet, __LINE__, __func__, TELNET_EPROTOCOL, 0, "DONT answered by WILL");

		case byte(Q_WANTYES):
			tl.setRFC1143(telopt, q_US(q), byte(Q_YES))
			tl.negotiateEvent(TELNET_EV_WILL, telopt)

		case byte(Q_WANTYES_OP):
			tl.setRFC1143(telopt, q_US(q), byte(Q_WANTNO))
			tl.sendNegotiate(TELNET_DONT, telopt)
			tl.negotiateEvent(TELNET_EV_WILL, telopt)
		}

	// request to disable option on remote end, confirm DONT, reject DO
	case TELNET_STATE_WONT:
		switch q_HIM(q) {
		case byte(Q_YES):
			tl.setRFC1143(telopt, q_US(q), byte(Q_NO))
			tl.sendNegotiate(TELNET_DONT, telopt)
			tl.negotiateEvent(TELNET_EV_WONT, telopt)
		case byte(Q_WANTNO):
			tl.setRFC1143(telopt, q_US(q), byte(Q_NO))
			tl.negotiateEvent(TELNET_EV_WONT, telopt)
		case byte(Q_WANTNO_OP):
			tl.setRFC1143(telopt, q_US(q), byte(Q_WANTYES))
			tl.negotiateEvent(TELNET_EV_DO, telopt)
		case byte(Q_WANTYES):
			fallthrough
		case byte(Q_WANTYES_OP):
			tl.setRFC1143(telopt, q_US(q), byte(Q_NO))
		}

	// request to enable option on local end or confirm WILL
	case TELNET_STATE_DO:
		switch q_US(q) {
		case byte(Q_NO):
			if tl.checkTelOpt(telopt, true) {
				tl.setRFC1143(telopt, byte(Q_YES), q_HIM(q))
				tl.sendNegotiate(TELNET_WILL, telopt)
				tl.negotiateEvent(TELNET_EV_DO, telopt)
			} else {
				tl.sendNegotiate(TELNET_WONT, telopt)
			}

		case byte(Q_WANTNO):
			tl.setRFC1143(telopt, byte(Q_NO), q_HIM(q))
			tl.negotiateEvent(TELNET_EV_DONT, telopt)
			//_error(Telnet, __LINE__, __func__, TELNET_EPROTOCOL, 0, "WONT answered by DO");

		case byte(Q_WANTNO_OP):
			tl.setRFC1143(telopt, byte(Q_YES), q_HIM(q))
			tl.negotiateEvent(TELNET_EV_DO, telopt)
			//_error(Telnet, __LINE__, __func__, TELNET_EPROTOCOL, 0, "WONT answered by DO");

		case byte(Q_WANTYES):
			tl.setRFC1143(telopt, byte(Q_YES), q_HIM(q))
			tl.negotiateEvent(TELNET_EV_DO, telopt)

		case byte(Q_WANTYES_OP):
			tl.setRFC1143(telopt, byte(Q_WANTNO), q_HIM(q))
			tl.sendNegotiate(TELNET_WONT, telopt)
			tl.negotiateEvent(TELNET_EV_DO, telopt)
		}

	// request to disable option on local end, confirm WONT, reject WILL
	case TELNET_STATE_DONT:
		switch q_US(q) {
		case byte(Q_YES):
			tl.setRFC1143(telopt, byte(Q_NO), q_HIM(q))
			tl.sendNegotiate(TELNET_WONT, telopt)
			tl.negotiateEvent(TELNET_EV_DONT, telopt)

		case byte(Q_WANTNO):
			tl.setRFC1143(telopt, byte(Q_NO), q_HIM(q))
			tl.negotiateEvent(TELNET_EV_WONT, telopt)

		case byte(Q_WANTNO_OP):
			tl.setRFC1143(telopt, byte(Q_WANTYES), q_HIM(q))
			tl.sendNegotiate(TELNET_WILL, telopt)
			tl.negotiateEvent(TELNET_EV_WILL, telopt)

		case byte(Q_WANTYES_OP):
			fallthrough
		case byte(Q_WANTYES):
			tl.setRFC1143(telopt, byte(Q_NO), q_HIM(q))
		}
	} //switch
}

// Process a subnegotiation buffer; return non-zero if the current buffer
// must be aborted and reprocessed due to COMPRESS2 being activated
func (tl *Telnet) subnegotiate() bool {
	subnEvent := NewTelnetSubnegotiateEvent()
	subnEvent.TelOpt = TelnetOptions(tl.sb_telopt)
	subnEvent.Buffer = tl.buffer.Bytes()
	tl.callEventHandler(subnEvent)

	switch tl.sb_telopt {
	// specially handled subnegotiation telopt types
	case byte(TELOPT_ZMP):
		//TODO: return ZMPTelnet();
	case byte(TELOPT_TTYPE):
		//TODO: return TTypeTelnet();
		/*
			case (byte)TelnetOptions.TELOPT_OLD_ENVIRON:
			case (byte)TelnetOptions.TELOPT_NEW_ENVIRON:
				return _environ_telnet(_sb_telopt, _buffer, _buffer_pos);
		*/
	case byte(TELOPT_MSSP):
		//TODO: return MSSPTelnet();
	}
	return false
}

//------------------------------------------------------------------------------------------------//

// Retrieve RFC1143 option state
func (tl *Telnet) getRFC1143(telopt byte) TelnetRFC1143 {

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
func (tl *Telnet) setRFC1143(telopt byte, us byte, him byte) {
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

// Check if we support a particular telopt; if us is non-zero, we
// check if we(local) supports it, otherwise we check if he(remote)
// supports it.  return non-zero if supported, zero if not supported.
func (tl *Telnet) checkTelOpt(telopt byte, us bool) bool {
	// if we have no telopts table, we obviously don't support it
	if len(tl.telOpts) == 0 {
		return false
	}

	// loop until found or end marker (us and him both 0)
	for _, v := range tl.telOpts {
		if byte(v.TelOpt) == telopt {
			if us && v.Us == TELNET_WILL {
				return true
			} else if us && v.Him == TELNET_DO {
				return true
			} else {
				return false
			}
		}
	}

	// not found, so not supported
	return false
}

//------------------------------------------------------------------------------------------------//

// Send negotiation bytes
func (tl *Telnet) sendNegotiate(cmd TelnetCommands, telopt byte) {
	data := []byte{TELNET_IAC, byte(cmd), telopt}
	tl.send(data)
}

//------------------------------------------------------------------------------------------------//

// helper for the negotiation routines
func (tl *Telnet) negotiateEvent(EventType TelnetEventType, opt byte) {
	ne := NewTelnetNegotiateEvent(EventType)
	ne.TelOpt = TelnetOptions(opt)
	tl.callEventHandler(ne)
}
