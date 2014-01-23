package gorainbow

type (
	socketHandler interface {
		OnMessage(m Message)
	}
)

func NewSocket(sockType string, conn *Connection) socketHandler {
	switch sockType {
	case "ssh":
		return &SSHSocket{Conn: conn}
	default:
		return nil
	}
}
