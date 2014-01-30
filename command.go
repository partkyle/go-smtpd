package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

const (
	HELO = "HELO"
	EHLO = "EHLO"
	MAIL = "MAIL"
	RCPT = "RCPT"
	DATA = "DATA"
	RSET = "RSET"
	NOOP = "NOOP"
	QUIT = "QUIT"
)

type Command struct {
	Command  string
	Argument string
}

func (c *Command) String() string {
	return c.Command + c.Argument
}

func ReadCommand(r io.Reader) (*Command, error) {
	line := make([]byte, 1024)
	n, err := r.Read(line)
	if err != io.EOF {
		return nil, err
	}

	line = line[:n]

	if len(line) < 4 {
		return nil, fmt.Errorf("Invalid command %q", line)
	}

	command := &Command{Command: strings.ToUpper(string(line[:4])), Argument: string(line[4:])}
	logger.Printf("Got command: %+v", command)
	return command, nil
}

type cmdReader struct {
	r     *bufio.Reader
	state int
}

func (c *cmdReader) Read(b []byte) (n int, err error) {
	const (
		stateData = iota
		stateCR
		stateEOF
	)

	br := c.r
	for n < len(b) && c.state != stateEOF {
		var char byte
		char, err = br.ReadByte()
		if err != nil {
			if err == io.EOF {
				err = io.ErrUnexpectedEOF
			}
			break
		}

		switch c.state {
		case stateData:
			if char == '\r' {
				c.state = stateCR
				continue
			}
		case stateCR:
			if char == '\n' {
				c.state = stateEOF
				continue
			}
			// unwind the last change
			char = '\r'
			br.UnreadByte()
			c.state = stateData
		}
		b[n] = char
		n++
	}

	if err == nil && c.state == stateEOF {
		err = io.EOF
	}

	return
}
