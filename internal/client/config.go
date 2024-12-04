package client

import (
	"net"
	"time"

	"github.com/terassyi/tcp-half-close-tester/internal/logutils"
)

type Config struct {
	Server       string
	WriteTimeout time.Duration
	ReadTimeout  time.Duration
	BufSize      int
	Echo         bool
	LogLevel     string
}

func (c *Config) Validate() error {
	if _, err := net.ResolveTCPAddr("tcp", c.Server); err != nil {
		return err
	}
	if _, err := logutils.LogLevelFromString(c.LogLevel); err != nil {
		return err
	}
	return nil
}
