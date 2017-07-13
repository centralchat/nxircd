package interfaces

import (
	"strings"
)

// Socket Holds the basic methods for an ircd.Socket
// this allows us to override for non external clients
// like services etc
type Socket interface {
	Write(string) (int, error) // Write to the socket
	Read() (string, error)     // Read from the socket
	Close()
	IP() string // Return the IP Address for the connection
	// (For services should return 127.0.0.1 or w/e)
}

type TestSocket struct {
	Socket

	Out []string
	In  []string
}

func NewTestSocket() *TestSocket {
	return &TestSocket{
		Out: []string{},
		In:  []string{},
	}
}

func (sock *TestSocket) Read() (string, error) {
	return sock.GrabReadLine(), nil
}

func (sock *TestSocket) Write(line string) (int, error) {
	sock.Out = append(sock.Out, line)
	return len(line), nil
}

func (sock *TestSocket) IP() string {
	return "127.0.0.1"
}

func (sock *TestSocket) GrabWriteLine() (line string) {
	if len(sock.Out) > 0 {
		line = sock.Out[0]
		if len(sock.Out) > 1 {
			sock.Out = sock.Out[1:]
		} else {
			sock.Out = []string{}
		}
	}
	line = strings.Replace(line, "\r\n", "", -1)
	return
}

func (sock *TestSocket) GrabReadLine() (line string) {
	if len(sock.In) > 0 {
		line = sock.In[0]
		if len(sock.In) > 1 {
			sock.In = sock.In[1:]
		} else {
			sock.In = []string{}
		}
	}

	line = strings.Replace(line, "\r\n", "", -1)
	return
}
