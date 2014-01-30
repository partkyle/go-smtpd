package main

import (
	"bufio"
	"bytes"
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

	return ParseCommand(line)
}

func ParseCommand(line []byte) (*Command, error) {
	var verb, argument string
	verb = strings.ToUpper(string(line[:4]))

	switch verb {
	case MAIL:
		prefix := []byte("MAIL FROM:")
		if !bytes.HasPrefix(bytes.ToUpper(line), prefix) {
			return nil, fmt.Errorf("MAIL command is invalid")
		}
		argument = line[len(prefix)+1:]
	case RCPT:
		prefix := []byte("RCPT TO:")
		if !bytes.HasPrefix(bytes.ToUpper(line), prefix) {
			return nil, fmt.Errorf("MAIL command is invalid")
		}
		argument = line[len(prefix)+1:]
	case HELO, EHLO:
		var trash string
		fmt.Sscanf(string(line), "%s %s", &trash, &argument)
		if argument == "" {
			return nil, fmt.Errorf("Command %q requires a parameter", verb)
		}
	case QUIT, NOOP, RSET, DATA:
		if len(line) > 4 {
			return nil, fmt.Errorf("Command %q does not support arguments", verb)
		}
	default:
		return nil, fmt.Errorf("Unsupported command line: %q", line)
	}
	command := &Command{Command: verb, Argument: argument}
	logger.Printf("Got command: %q", command)
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
