package ircd

import (
	"bufio"
	"io"
	"net"

	"nxircd/interfaces"
)

// IRCSocket extends Socket interface
// and holds information regarding a IRC Connection
type IRCSocket struct {
	interfaces.Socket

	Closed bool

	conn net.Conn

	scanner *bufio.Scanner
	writer  *bufio.Writer
}

// NewIRCSocket takes a connection and returns a new ptr
// to an IRCSocket struct
func NewIRCSocket(conn net.Conn) *IRCSocket {
	return &IRCSocket{
		conn:    conn,
		Closed:  false,
		scanner: bufio.NewScanner(conn),
		writer:  bufio.NewWriter(conn),
	}
}

func (sock *IRCSocket) Write(line string) (int, error) {
	return sock.writer.WriteString(line)
}

func (sock *IRCSocket) Read() (string, error) {
	for sock.scanner.Scan() {
		if sock.Closed {
			return "", io.EOF
		}

		line := sock.scanner.Text()
		if len(line) == 0 {
			continue
		}
		return line, nil
	}
	return "", io.EOF
}

func (sock *IRCSocket) Close() {
	sock.Closed = true
	// Ignore error but indicate in the code we are ignoring it.
	_ = sock.conn.Close()
}
