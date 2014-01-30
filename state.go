package main

import (
	"fmt"
)

type stateFn func(Conv) stateFn

func beginState(c Conv) stateFn {
	c.WriteHeader(220)
	_, err := fmt.Fprintf(c, c.Banner())
	if err != nil {
		return nil
	}

	return greetingState
}

func greetingState(c Conv) stateFn {
	command, err := ReadCommand(c.CmdReader())
	if err != nil {
		logger.Printf("error during read command: %q", err)
		return greetingState
	}

	switch command.Command {
	case QUIT:
		c.WriteHeader(221)
		fmt.Fprintf(c, "Goodbye!")
		return nil
	case HELO, EHLO:
		c.Domain(command.Argument)
	default:
		c.WriteHeader(501)
		fmt.Fprintf(c, "Invalid Command %q", command)
		return greetingState
	}

	c.WriteHeader(250)
	fmt.Fprint(c, "Ok")

	return mailState
}

func mailState(c Conv) stateFn {
	command, err := ReadCommand(c.CmdReader())
	if err != nil {
		return mailState
	}

	switch command.Command {
	case QUIT:
		c.WriteHeader(221)
		fmt.Fprintf(c, "Goodbye!")
		return nil
	case MAIL:
		c.Start()
		c.Envelope().MailFrom(c, command.Argument)
	default:
		c.WriteHeader(501)
		fmt.Fprintf(c, "Invalid Command %q", command)
		return mailState
	}

	return rcptState
}

func rcptState(c Conv) stateFn {
	command, err := ReadCommand(c.CmdReader())
	if err != nil {
		return rcptState
	}

	switch command.Command {
	case QUIT:
		c.WriteHeader(221)
		fmt.Fprintf(c, "Goodbye!")
		return nil
	case RCPT:
		c.Envelope().RcptTo(c, command.Argument)
	default:
		c.WriteHeader(501)
		fmt.Fprintf(c, "Invalid Command %q", command)
		return rcptState
	}

	return rcptDataState
}

func rcptDataState(c Conv) stateFn {
	command, err := ReadCommand(c.CmdReader())
	if err != nil {
		return rcptState
	}

	switch command.Command {
	case QUIT:
		c.WriteHeader(221)
		fmt.Fprintf(c, "Goodbye!")
		return nil
	case RCPT:
		c.Envelope().RcptTo(c, command.Argument)
	case DATA:
		c.WriteHeader(354)
		fmt.Fprintf(c, "<CRLF>.<CRLF>")
		return dataState
	default:
		c.WriteHeader(501)
		fmt.Fprintf(c, "Invalid Command %q", command)
		return rcptState
	}

	return rcptDataState
}

func dataState(c Conv) stateFn {
	c.Envelope().Data(c, c.DotReader())
	c.Finish()
	return mailState
}
