package shell

import (
	"bufio"
	"net"
)

// CloseWriter is an interface that implements structs
// that close input streams to prevent from writing.
type CloseWriter interface {
	CloseWrite() error
}

type Shell struct {
	conn   net.Conn
	reader *bufio.Reader
}

func NewShell(conn net.Conn, reader *bufio.Reader) *Shell {
	return &Shell{conn: conn, reader: reader}
}

func (shell *Shell) GetConn() net.Conn {
	return shell.conn
}

func (shell *Shell) GetReader() *bufio.Reader {
	return shell.reader
}

// Close closes the  connection and reader.
func (shell *Shell) Close() {
	shell.conn.Close()
}

// CloseWrite closes a readWriter for writing.
func (shell *Shell) CloseWrite() error {
	if conn, ok := shell.conn.(CloseWriter); ok {
		return conn.CloseWrite()
	}
	return nil
}
