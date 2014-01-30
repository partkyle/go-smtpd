package main

import (
	"bufio"
	"fmt"
)

type ResponseWriter interface {
	WriteResponse(int, ...string) error
	lastResult() int
}

type responseWriter struct {
	code    int
	writer  *bufio.Writer
	written bool
}

func (r *responseWriter) WriteResponse(status int, lines ...string) error {
	r.code = status

	for i := 0; i < len(lines)-1; i++ {
		_, err := fmt.Fprintf(r.writer, "%d-%s\r\n", status, lines[i])
		if err != nil {
			return err
		}
	}

	_, err := fmt.Fprintf(r.writer, "%d %s\r\n", status, lines[len(lines)-1])
	if err != nil {
		return err
	}

	return r.writer.Flush()
}

func (r *responseWriter) lastResult() int {
	return r.code
}
