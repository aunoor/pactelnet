package pactelnet

import "testing"

func TestDataIAC(t *testing.T) {
	var rsvData []byte

	telnet := NewTelnet(nil, nil, nil)
	telnet.OnTelnetEvent = func(telnetEvent TelnetEventInterface) {
		switch telnetEvent.EventType() {
		case TELNET_EV_DATA:
			dataEvent := telnetEvent.(*TelnetDataEvent)
			rsvData = dataEvent.Buffer
		}
	}
	telnet.TelnetRecv([]byte{'a', 'b', 'c', TELNET_IAC})

	if !func() bool {
		if rsvData[0] != 'a' {
			return false
		}
		if rsvData[1] != 'b' {
			return false
		}
		if rsvData[2] != 'c' {
			return false
		}
		return true
	}() {
		t.Error("Data from telnet not equal with expected")
	}
}
