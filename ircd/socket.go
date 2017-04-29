package ircd

import (
  "bufio"
  "io"
  "net"
)

/**************************************************************/

// Socket holds the internals how we actually interact
// with the users descriptor
type Socket struct {
  closed  bool
  conn    net.Conn
  scanner *bufio.Scanner
  writer  *bufio.Writer
}

/**************************************************************/

// NewSocket creates an instance of our socket struct
func NewSocket(c net.Conn) *Socket {
  return &Socket{
    conn:    c,
    scanner: bufio.NewScanner(c),
    writer:  bufio.NewWriter(c),
  }
}

/**************************************************************/

func (socket *Socket) Read() (line string, err error) {
  if socket.closed {
    err = io.EOF
    return
  }
  for socket.scanner.Scan() {
    line = socket.scanner.Text()
    if len(line) == 0 {
      continue
    }
    return
  }
  err = socket.scanner.Err()
  if err == nil {
    err = io.EOF
  }
  return
}

/**************************************************************/

func (socket *Socket) Write(line string) (err error) {
  if socket.closed {
    err = io.EOF
    return
  }

  if _, err = socket.writer.WriteString(line); err != nil {
    return
  }

  // if _, err = socket.writer.WriteString(CRLF); err != nil {
  //   return
  // }

  if err = socket.writer.Flush(); err != nil {
    return
  }

  return
}
