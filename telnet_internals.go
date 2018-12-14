package pactelnet

type (
	telnetInternalFlags byte
	telnetState         byte
	telnetErrorCode     byte
	rfc1143StateNames   byte

	// RFC1143 option negotiation state
	TelnetRFC1143 struct {
		telopt byte
		state  byte
	}
)

// telnet NVT EOL sequences
var CRLF = []byte{'\r', '\n'}
var CRNUL = []byte{'\r', 0}

const (
	TELNET_FLAG_TRANSMIT_BINARY telnetInternalFlags = 32
	TELNET_FLAG_RECEIVE_BINARY                      = 64
	TELNET_PFLAG_DEFLATE                            = 128
)
const (
	TELNET_STATE_DATA telnetState = 0
	TELNET_STATE_EOL
	TELNET_STATE_IAC
	TELNET_STATE_WILL
	TELNET_STATE_WONT
	TELNET_STATE_DO
	TELNET_STATE_DONT
	TELNET_STATE_SB
	TELNET_STATE_SB_DATA
	TELNET_STATE_SB_DATA_IAC
)

/// Error codes
const (
	TELNET_EOK       telnetErrorCode = 0 /*!< no error */
	TELNET_EBADVAL                       /*!< invalid parameter, or API misuse */
	TELNET_ENOMEM                        /*!< memory allocation failure */
	TELNET_EOVERFLOW                     /*!< data exceeds buffer size */
	TELNET_EPROTOCOL                     /*!< invalid sequence of special bytes */
	TELNET_ECOMPRESS                     /*!< error handling compressed streams */
)

/// <summary>
/// RFC1143 state names
/// </summary>
const (
	Q_NO         rfc1143StateNames = 0
	Q_YES                          = 1
	Q_WANTNO                       = 2
	Q_WANTYES                      = 3
	Q_WANTNO_OP                    = 4
	Q_WANTYES_OP                   = 5
)

/* helper for Q-method option tracking */
func q_US(q TelnetRFC1143) byte {
	return (byte)(q.state & 0x0F)
}

func q_HIM(q TelnetRFC1143) byte {
	return (byte)((q.state & 0xF0) >> 4)
}

func q_MAKE(us byte, him byte) byte {
	return (byte)((us) | ((him) << 4))
}
