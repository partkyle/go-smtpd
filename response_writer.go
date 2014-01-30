package main

import (
	"bufio"
	"bytes"
	"fmt"
)

type ResponseWriter interface {
	WriteHeader(int)
	Write([]byte) (int, error)
}

type responseWriter struct {
	code    int
	writer  *bufio.Writer
	written bool
}

func (r *responseWriter) WriteHeader(status int) {
	r.code = status
}

func (r *responseWriter) Write(p []byte) (n int, err error) {
	lines := bytes.Split(p, []byte("\r\n"))

	for i := 0; i < len(lines)-1; i++ {
		written, err := fmt.Fprintf(r.writer, "%d-%s\r\n", r.code, lines[i])
		n += written

		if err != nil {
			return n, err
		}
	}

	written, err := fmt.Fprintf(r.writer, "%d %s\r\n", r.code, lines[len(lines)-1])
	n += written
	if err != nil {
		return n, err
	}

	if n > 0 {
		r.written = true
		err = r.writer.Flush()
	}

	return n, err
}
