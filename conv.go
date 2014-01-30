package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"time"
)

var ErrLineTooLine = errors.New("Command line was too long!")

type Conv interface {
	Banner() string

	Domain(string)

	BeginTransaction()
	Envelope() Envelope
	EndTransaction()

	CmdReader() io.Reader
	DotReader() io.Reader

	handleError(error) stateFn

	ResponseWriter
}

type TimedReader struct {
	idleTimeout time.Duration

	net.Conn
}

func (t *TimedReader) Read(b []byte) (int, error) {
	t.Conn.SetReadDeadline(time.Now().Add(t.idleTimeout))

	return t.Conn.Read(b)
}

type Conversation struct {
	state stateFn

	conn        net.Conn
	idleTimeout time.Duration

	domain string

	envelope Envelope

	ResponseWriter
}

func (c *Conversation) handleError(err error) stateFn {
	if err == io.ErrUnexpectedEOF {
		c.WriteResponse(503, "Command Line Too Long!")
		return nil
	}

	if err == ErrLineTooLine {
		c.WriteResponse(550, ErrLineTooLine.Error())
		return nil
	}

	// no transition on command errors
	c.WriteResponse(501, err.Error())
	return c.state
}

func (c *Conversation) Banner() string {
	return fmt.Sprintf("Welcome to SMTP %s! %s", c.conn.RemoteAddr(), time.Now())
}

func (c *Conversation) Domain(domain string) {
	c.domain = domain
	logger.Printf("Got domain: %q", c.domain)
}

func (c *Conversation) BeginTransaction() {
	c.envelope = &SimpleEnvelope{}
}

func (c *Conversation) Envelope() Envelope {
	return c.envelope
}

func (c *Conversation) EndTransaction() {
	// Clear out the reference to the transaction
	c.envelope = nil
}

func (c *Conversation) DotReader() io.Reader {
	timedReader := &TimedReader{
		idleTimeout: c.idleTimeout,
		Conn:        c.conn,
	}
	return &dotReader{r: bufio.NewReader(timedReader)}
}

func (c *Conversation) CmdReader() io.Reader {
	timedReader := &TimedReader{
		idleTimeout: c.idleTimeout,
		Conn:        c.conn,
	}

	return &cmdReader{r: bufio.NewReader(timedReader)}
}

func (c *Conversation) Run() {
	defer c.conn.Close()

	c.state = bannerState
	c.ResponseWriter = &responseWriter{writer: bufio.NewWriter(c.conn)}

	for c.state != nil {
		c.state = c.state(c)
	}
}
