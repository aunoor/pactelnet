package pactelnet

type (
	TelnetCommands  byte
	TelnetOptions   byte
	TelnetEventType byte
	TelnetMSSP      byte
	TelnetFlags     byte

	TelnetOptionReq struct {
		// one of the TELOPT codes
		TelOpt TelnetOptions
		// TELNET_WILL or TELNET_WONT
		Us TelnetCommands
		// TELNET_DO or TELNET_DONT
		Him TelnetCommands
	}

	MSSPPair struct {
		variable []byte
		value    []byte
	}
)

const (
	// End of subnegotiation parameters.
	TELNET_SE TelnetCommands = 240
	// No operation.
	TELNET_NOP = 241
	// The data stream portion of a Synch.
	TELNET_DM = 242
	// NVT character BRK.
	TELNET_BRK = 243
	// The function IP(Interrupt Process).
	TELNET_IP = 244
	// The function AO(Abort output).
	TELNET_AO = 245
	// The function AYT(Are You There).
	TELNET_AYT = 246
	// The function EC(Erase character).
	TELNET_EC = 247
	// The function EL(Erase Line).
	TELNET_EL = 248
	// The GA signal(Go ahead).
	TELNET_GA = 249
	// Indicates that what follows is subnegotiation of the indicated option.
	TELNET_SB = 250
	// Indicates the desire to begin performing, or confirmation that
	// you are now performing, the indicated option.
	TELNET_WILL = 251
	// Indicates the refusal to perform, or continue performing, the indicated option.
	TELNET_WONT = 252
	/// Indicates the request that the other party perform, or confirmation that you are expecting
	/// the other party to perform, the indicated option.
	TELNET_DO = 253
	// Indicates the demand that the other party stop performing,
	// or confirmation that you are no longer expecting the other party to perform, the indicated option.
	TELNET_DONT = 254
	// Data Byte 255.
	TELNET_IAC = 255
)

const (
	// 8-bit data path
	TELOPT_BINARY TelnetOptions = 0
	// Echo
	TELOPT_ECHO = 1
	// Prepare to reconnect
	TELOPT_RCP = 2
	// Suppress go ahead
	TELOPT_SGA = 3
	// Approximate message size
	TELOPT_NAMS = 4
	// Give status
	TELOPT_STATUS = 5
	// Timing mark
	TELOPT_TM = 6
	// Remote controlled transmission and echo
	TELOPT_RCTE = 7
	// Negotiate about output line width
	TELOPT_NAOL = 8
	// Negotiate about output page size
	TELOPT_NAOP = 9
	// Negotiate about CR disposition
	TELOPT_NAOCRD = 10
	// Negotiate about horizontal tabstops
	TELOPT_NAOHTS       = 11
	TELOPT_NAOHTD       = 12 /* negotiate about horizontal tab disposition */
	TELOPT_NAOFFD       = 13 /* negotiate about formfeed disposition */
	TELOPT_NAOVTS       = 14 /* negotiate about vertical tab stops */
	TELOPT_NAOVTD       = 15 /* negotiate about vertical tab disposition */
	TELOPT_NAOLFD       = 16 /* negotiate about output LF disposition */
	TELOPT_XASCII       = 17 /* extended ascic character set */
	TELOPT_LOGOUT       = 18 /* force logout */
	TELOPT_BM           = 19 /* byte macro */
	TELOPT_DET          = 20 /* data entry terminal */
	TELOPT_SUPDUP       = 21 /* supdup protocol */
	TELOPT_SUPDUPOUTPUT = 22 /* supdup output */
	TELOPT_SNDLOC       = 23 /* send location */
	TELOPT_TTYPE        = 24 /* terminal type */
	TELOPT_EOR          = 25 /* end or record */
	TELOPT_TUID         = 26 /* TACACS user identification */
	TELOPT_OUTMRK       = 27 /* output marking */
	TELOPT_TTYLOC       = 28 /* terminal location number */
	TELOPT_3270REGIME   = 29 /* 3270 regime */
	TELOPT_X3PAD        = 30 /* X.3 PAD */
	TELOPT_NAWS         = 31 /* window size */
	TELOPT_TSPEED       = 32 /* terminal speed */
	TELOPT_LFLOW        = 33 /* remote flow control */
	// Linemode option
	TELOPT_LINEMODE        = 34
	TELOPT_XDISPLOC        = 35 /* X Display Location */
	TELOPT_OLD_ENVIRON     = 36 /* Old - Environment variables */
	TELOPT_AUTHENTICATION  = 37 /* Authenticate */
	TELOPT_ENCRYPT         = 38 /* Encryption option */
	TELOPT_NEW_ENVIRON     = 39 /* New - Environment variables */
	TELOPT_TN3270E         = 40 /* TN3270 enhancements */
	TELOPT_XAUTH           = 41
	TELOPT_CHARSET         = 42 /* Character set */
	TELOPT_RSP             = 43 /* Remote serial port */
	TELOPT_COM_PORT_OPTION = 44 /* Com port control */
	TELOPT_SLE             = 45 /* Suppress local echo */
	TELOPT_STARTTLS        = 46 /* Start TLS */
	TELOPT_KERMIT          = 47 /* Automatic Kermit file transfer */
	TELOPT_SEND_URL        = 48
	TELOPT_FORWARD_X       = 49
	// Mud serverstate protocol
	TELOPT_MSSP      = 70
	TELOPT_COMPRESS  = 85
	TELOPT_COMPRESS2 = 86
	TELOPT_MCCP2     = 86
	// ZMud protocol
	TELOPT_ZMP              = 93
	TELOPT_PRAGMA_LOGON     = 138
	TELOPT_SSPI_LOGON       = 139
	TELOPT_PRAGMA_HEARTBEAT = 140
	TELOPT_EXOPL            = 255
)

// Protocol codes for MSSP commands.
const (
	MSSP_VAR TelnetMSSP = 1
	MSSP_VAL            = 2
)

const (
	TELNET_EV_DATA           TelnetEventType = iota /*!< raw text data has been received */
	TELNET_EV_SEND                                  /*!< data needs to be sent to the peer */
	TELNET_EV_IAC                                   /*!< generic IAC code received */
	TELNET_EV_WILL                                  /*!< WILL option negotiation received */
	TELNET_EV_WONT                                  /*!< WONT option neogitation received */
	TELNET_EV_DO                                    /*!< DO option negotiation received */
	TELNET_EV_DONT                                  /*!< DONT option negotiation received */
	TELNET_EV_SUBNEGOTIATION                        /*!< sub-negotiation data received */
	TELNET_EV_COMPRESS                              /*!< compression has been enabled */
	TELNET_EV_ZMP                                   /*!< ZMP command has been received */
	TELNET_EV_TTYPE                                 /*!< TTYPE command has been received */
	TELNET_EV_ENVIRON                               /*!< ENVIRON command has been received */
	TELNET_EV_MSSP                                  /*!< MSSP command has been received */
	TELNET_EV_WARNING                               /*!< recoverable error has occured */
	TELNET_EV_ERROR                                 /*!< non-recoverable error has occured */
)

// Control behavior of telnet state tracker.
const (
	// Operate in proxy mode.  This disables the RFC1143 support and
	// enables automatic detection of COMPRESS2 streams.
	TELNET_FLAG_PROXY TelnetFlags = 1
	// Receive data with translation of the TELNET NVT CR NUL and CR LF
	// sequences specified in RFC854 to C carriage return (\r) and C
	// newline(\n), respectively.
	TELNET_FLAG_NVT_EOL = 2
)
