package shell

import "github.com/moby/term"

// TerminalChecker holds logic to check if environment is a terminal
type TerminalChecker interface {
	IsTerminal(...interface{}) bool
}

// DefaultTerminalChecker holds data to check if environment is a terminal
type DefaultTerminalChecker struct{}

// IsTerminal checks if input is a terminal
func (t *DefaultTerminalChecker) IsTerminal(fds ...interface{}) bool {
	for fd := range fds {
		_, isTerm := term.GetFdInfo(fd)

		if !isTerm {
			return false
		}
	}

	return true
}

// NewTerminalChecker creates a new terminal checker
func NewTerminalChecker() TerminalChecker {
	return &DefaultTerminalChecker{}
}
