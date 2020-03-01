package protocol

const (
	MsgTypePTYReq = "pty-req"
	MsgTypeX11Req = "x11-req"
	MsgTypeEnv    = "env"
	MsgTypeExec   = "exec"
	MsgTypeShell  = "shell"
	MsgTypeSignal = "signal"

	MsgTypeSubsystem    = "subsystem"
	MsgTypeExitStatus   = "exit-status"
	MsgTypeExitSignal   = "exit-signal"
	MsgTypeTcpIpForward = "tcpip-forward"
)

// RFC 4254 Section 6.2 Requesting a Pseudo-Terminal
// type: "pty-req"
type MsgRequestPTY struct {
	Term     string
	Columns  uint32
	Rows     uint32
	Width    uint32
	Height   uint32
	Modelist string
}

// RFC 4254 Section 6.3.1 Requesting X11 Forwarding
// type: "x11-req"
type MsgRequestX11Forward struct {
	Single       bool
	AuthProtocol string
	AuthCookie   string
	ScreenNumber uint32
}

// RFC 4254 Section 6.4 Environment Variable Passing
// type: "env"
// Environment variables
type MsgRequestSetEnv struct {
	Name  string
	Value string
}

// RFC 4254 Section 6.5 Starting a Shell or a Command
// type: "exec"
//   This message will request that the server start the execution of the
// given command.
type MsgRequestExec struct {
	Command string
}

// RFC 4254 Section 6.5 Starting a Shell or a Command
// type: "shell"
//   This message will request that the user's default shell (typically
// defined in /etc/passwd in UNIX systems) be started at the other end.
type MsgRequestShell struct{}

// RFC 4254 Section 6.5 Starting a Shell or a Command
// type: "subsystem"
// Predefined subsystem
type MsgRequestSubsystem struct {
	Name string
}

// RFC 4254 Section 6.7 Window Dimension Change Message
// type: "window-change"
type MsgRequestPTYWindowChange struct {
	Columns uint32
	Rows    uint32
	Width   uint32
	Height  uint32
}

// RFC 4254 Section 6.9 Signals
// type: "signal"
type MsgSignal struct {
	// signal name (without the "SIG" prefix)
	Signal string
}

// RFC 4254 Section 6.10 Returning Exit Status
// type: "exit-status"
type MsgExitStatus struct {
	ExitStatus uint32
}

// RFC 4254 Section 6.10 Returning Exit Status
// type: "exit-signal"
type MsgExitSignal struct {
	// signal name (without the "SIG" prefix)
	Signal     string
	CoreDumped bool
	Error      string
	Lang       string
}

// RFC 4254 Section 7.1 Requesting Port Forwarding
// type: "tcpip-forward"
type MsgRequestPortForward struct {
	Address string
	Port    uint32
}

// RFC 4254 Section 7.1 Requesting Port Forwarding
// type: "cancel-tcpip-forward"
type MsgRequestCancelPortForward struct {
	Address string
	Port    uint32
}

// RFC 4254 Section 7.2 TCP/IP Forwarding Channels
// type: "forwarded-tcpip"
type MsgChannelOpenForwarded struct {
	RAddr string
	RPort uint32
	LAddr string
	LPort uint32
}

// RFC 4254 Section 7.2 TCP/IP Forwarding Channels
// type: "direct-tcpip"
type MsgChannelOpenDirect struct {
	RAddr string
	LAddr string
	LPort uint32
}

type MsgUnparsed struct {
	Type    string
	Payload []byte
}
