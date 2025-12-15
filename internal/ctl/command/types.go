package command

type CmdType string

const (
	CmdType_Ping CmdType = "ping"
	CmdType_Pong CmdType = "pong"
)

type PingPayload struct{}
type PongPayload struct{}

type Msg struct {
	CmdType CmdType
	Payload []byte
}
