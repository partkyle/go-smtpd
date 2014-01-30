package main

import (
	"fmt"
)

func isSuccess(code int) bool {
	return (code / 100) == 2
}

type stateFn func(Conv) stateFn

func bannerState(c Conv) stateFn {
	err := c.WriteResponse(220, c.Banner())
	if err != nil {
		return nil
	}

	return greetingState
}

func greetingState(c Conv) stateFn {
	command, err := ReadCommand(c.CmdReader())
	if err != nil {
		return c.handleError(err)
	}

	switch command.Command {
	case QUIT:
		c.WriteResponse(221, "Goodbye!")
		return nil
	case HELO, EHLO:
		c.Domain(command.Argument)
	default:
		c.WriteResponse(501, fmt.Sprintf("Invalid Command %q", command))
		return greetingState
	}

	c.WriteResponse(250, "Ok")

	return beginTransactionState
}

func beginTransactionState(c Conv) stateFn {
	c.BeginTransaction()
	return mailState
}

func mailState(c Conv) stateFn {
	command, err := ReadCommand(c.CmdReader())
	if err != nil {
		return c.handleError(err)
	}

	switch command.Command {
	case QUIT:
		c.WriteResponse(221, "Goodbye!")
		return nil
	case MAIL:
		c.Envelope().MailFrom(c, command.Argument)
		if !isSuccess(c.lastResult()) {
			return mailState
		}
	default:
		c.WriteResponse(501, fmt.Sprintf("Invalid Command %q", command))
		return mailState
	}

	return rcptState
}

func rcptState(c Conv) stateFn {
	command, err := ReadCommand(c.CmdReader())
	if err != nil {
		return c.handleError(err)
	}

	switch command.Command {
	case QUIT:
		c.WriteResponse(221, "Goodbye!")
		return nil
	case RCPT:
		c.Envelope().RcptTo(c, command.Argument)
		if !isSuccess(c.lastResult()) {
			return rcptState
		}
	default:
		c.WriteResponse(501, fmt.Sprintf("Invalid Command %q", command))
		return rcptState
	}

	return rcptDataState
}

func rcptDataState(c Conv) stateFn {
	command, err := ReadCommand(c.CmdReader())
	if err != nil {
		return c.handleError(err)
	}

	switch command.Command {
	case QUIT:
		c.WriteResponse(221, "Goodbye!")
		return nil
	case RCPT:
		// whether this is success or not, it will remain in this state
		c.Envelope().RcptTo(c, command.Argument)
	case DATA:
		c.WriteResponse(354, "<CRLF>.<CRLF>")
		return dataState
	default:
		c.WriteResponse(501, fmt.Sprintf("Invalid Command %q", command))
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
