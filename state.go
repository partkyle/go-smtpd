package main

import (
	"fmt"
)

func isSuccess(code int) bool {
	return (code / 100) == 2
}

type stateFn func(Conv) stateFn

func bannerState(c Conv) stateFn {
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
		c.WriteHeader(503)
		fmt.Fprintf(c, "You did something bad: %q", err)
		logger.Printf("error during read command: %q", err)
		return nil
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

	return beginTransactionState
}

func beginTransactionState(c Conv) stateFn {
	c.BeginTransaction()
	return mailState
}

func mailState(c Conv) stateFn {
	command, err := ReadCommand(c.CmdReader())
	if err != nil {
		c.WriteHeader(503)
		fmt.Fprintf(c, "You did something bad: %q", err)
		return nil
	}

	switch command.Command {
	case QUIT:
		c.WriteHeader(221)
		fmt.Fprintf(c, "Goodbye!")
		return nil
	case MAIL:
		c.Envelope().MailFrom(c, command.Argument)
		if !isSuccess(c.lastResult()) {
			return mailState
		}
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
		c.WriteHeader(503)
		fmt.Fprintf(c, "You did something bad: %q", err)
		return nil
	}

	switch command.Command {
	case QUIT:
		c.WriteHeader(221)
		fmt.Fprintf(c, "Goodbye!")
		return nil
	case RCPT:
		c.Envelope().RcptTo(c, command.Argument)
		if !isSuccess(c.lastResult()) {
			return rcptState
		}
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
		// whether this is success or not, it will remain in this state
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

	return endTransactionState
}

func endTransactionState(c Conv) stateFn {
	c.EndTransaction()
	return beginTransactionState
}
