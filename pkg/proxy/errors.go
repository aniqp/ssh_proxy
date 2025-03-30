package proxy

import "fmt"

type ConnectionErr struct {
	Message string
	Err     error
}

func (c *ConnectionErr) Error() string {
	return fmt.Sprintf("Connection error: %v", c.Message)
}

func (c *ConnectionErr) Unwrap() error {
	return c.Err
}
