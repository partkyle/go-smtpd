package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"time"
)

type Conv interface {
	Banner() string

	Domain(string)

	Start()
	Envelope() Envelope
	Finish()

	CmdReader() io.Reader
	DotReader() io.Reader

	ResponseWriter
}

type Conversation struct {
	state stateFn

	conn net.Conn

	domain string

	envelope Envelope

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

func (c *Conversation) Banner() string {
	return fmt.Sprintf("Welcome to SMTP %s! %s", c.conn.RemoteAddr(), time.Now())
}

func (c *Conversation) Domain(domain string) {
	c.domain = domain
	logger.Printf("Got domain: %q", c.domain)
}

func (c *Conversation) Start() {
	c.envelope = &SimpleEnvelope{}
}

func (c *Conversation) Envelope() Envelope {
	return c.envelope
}

func (c *Conversation) Finish() {
	c.envelope = &SimpleEnvelope{}
}

func (c *Conversation) DotReader() io.Reader {
	return &dotReader{r: bufio.NewReader(c.conn)}
}

func (c *Conversation) CmdReader() io.Reader {
	return &cmdReader{r: bufio.NewReader(c.conn)}
}

func (c *Conversation) Run() {
	defer c.conn.Close()

	c.state = beginState
	c.ResponseWriter = &responseWriter{writer: bufio.NewWriter(c.conn)}

	for c.state != nil {
		logger.Printf("Starting state: %q", c.state)
		c.state = c.state(c)
	}
}
