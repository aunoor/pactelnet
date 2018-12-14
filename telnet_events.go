package pactelnet

type (
	telnetEventInterface interface {
		EventType() TelnetEventType
		SetUserData(value interface{})
		UserData() interface{}
	}

	telnetEvent struct {
		telnetEventInterface
		eventType TelnetEventType
		userData  interface{}
	}

	// Data event: for SEND
	TelnetSendEvent struct {
		telnetEvent
		// Data buffer
		Buffer []byte
	}

	// Data event: for DATA
	TelnetDataEvent struct {
		telnetEvent
		// Data buffer
		Buffer []byte
	}

	// Command event: for IAC
	TelnetIacEvent struct {
		telnetEvent
		// Telnet command received
		Cmd TelnetCommands
	}

	// Negotiation event: WILL, WONT, DO, DONT
	TelnetNegotiateEvent struct {
		telnetEvent

		// Option being negotiated
		TelOpt TelnetOptions

		//public byte telOpt;
	}

	TelnetSubnegotiateEvent struct {
		telnetEvent
		// Data of sub-negotiation
		Buffer []byte
		// Option code for negotiation
		TelOpt TelnetOptions
	}
)

func (te *telnetEvent) EventType() TelnetEventType {
	return te.eventType
}

func (te *telnetEvent) SetUserData(value interface{}) {
	te.userData = value
}

func (te *telnetEvent) UserData() interface{} {
	return te.userData
}

func NewTelnetSendEvent() *TelnetSendEvent {
	se := &TelnetSendEvent{}
	se.eventType = TELNET_EV_SEND
	return se
}

func NewTelnetDataEvent() *TelnetDataEvent {
	de := &TelnetDataEvent{}
	de.eventType = TELNET_EV_DATA
	return de
}

func NewTelnetIacEvent() *TelnetIacEvent {
	ie := &TelnetIacEvent{}
	ie.eventType = TELNET_EV_IAC
	return ie
}

func NewTelnetNegotiateEvent(eventType TelnetEventType) *TelnetNegotiateEvent {
	ne := &TelnetNegotiateEvent{}
	ne.eventType = eventType
	return ne
}

func NewTelnetSubnegotiateEvent() *TelnetSubnegotiateEvent {
	se := &TelnetSubnegotiateEvent{}
	se.eventType = TELNET_EV_SUBNEGOTIATION
	return se
}
