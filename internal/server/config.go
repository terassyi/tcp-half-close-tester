package server

import (
	"net"

	"github.com/terassyi/tcp-half-close-tester/internal/logutils"
)

type Config struct {
	Listen   string
	File     string
	Chunk    int
	LogLevel string
}

func (c *Config) Validate() error {
	if _, err := net.ResolveTCPAddr("tcp", c.Listen); err != nil {
		return err
	}
	if _, err := logutils.LogLevelFromString(c.LogLevel); err != nil {
		return err
	}
	return nil
}