package main

import (
	"fmt"
	"io"
	"io/ioutil"
)

type Envelope interface {
	MailFrom(ResponseWriter, string)
	RcptTo(ResponseWriter, string)

	Data(ResponseWriter, io.Reader)
}

type SimpleEnvelope struct {
	Sender     string
	Recipients []string
}

func NewSimpleEnvelope() *SimpleEnvelope {
	return &SimpleEnvelope{Recipients: make([]string, 0, 1)}
}

func (s *SimpleEnvelope) MailFrom(w ResponseWriter, sender string) {
	s.Sender = sender

	logger.Printf("Got sender: %q", s.Sender)

	w.WriteHeader(250)
	fmt.Fprintf(w, "Ok")
}

func (s *SimpleEnvelope) RcptTo(w ResponseWriter, recipient string) {
	s.Recipients = append(s.Recipients, recipient)

	logger.Printf("Got recipient. Recipients: %s", s.Recipients)

	w.WriteHeader(250)
	fmt.Fprintf(w, "Ok")
}

func (s *SimpleEnvelope) Data(w ResponseWriter, r io.Reader) {
	body, err := ioutil.ReadAll(r)
	if err != nil {
		logger.Printf("Error reading: %q", err)
		w.WriteHeader(450)
		fmt.Fprintf(w, "Error occured during data: %q", err)
		return
	}

	logger.Printf("Got data: %q", body)

	w.WriteHeader(250)
	fmt.Fprintf(w, "Ok")
}
