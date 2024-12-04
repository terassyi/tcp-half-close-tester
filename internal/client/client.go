package client

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"time"

	"github.com/terassyi/tcp-half-close-tester/internal/logutils"
)

type Client struct {
	cfg *Config
}

var log *slog.Logger

func New(cfg *Config) *Client {
	level, _ := logutils.LogLevelFromString(cfg.LogLevel)
	log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	log = log.WithGroup("client")
	log.Info("new client", slog.String("server", cfg.Server), slog.Duration("read-timeout", cfg.ReadTimeout), slog.Duration("write-timeout", cfg.WriteTimeout))
	return &Client{cfg: cfg}
}

func (c *Client) Run(ctx context.Context) error {

	writeTimeoutTicker := time.NewTicker(c.cfg.WriteTimeout)
	readTimeoutTicker := time.NewTicker(c.cfg.ReadTimeout)

	log = log.With("server", c.cfg.Server)

	addr, err := net.ResolveTCPAddr("tcp", c.cfg.Server)
	if err != nil {
		return fmt.Errorf("failed to resolve host: %w", err)
	}

	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return fmt.Errorf("failed to dial: %w", err)
	}

	defer func() {
		log.InfoContext(ctx, "finish handling")
	}()

	dataChan := make(chan []byte)
	done := make(chan error)
	log.InfoContext(ctx, "start to read")
	go c.readFromConn(ctx, conn, dataChan, done)

	readExpire, writeExpire := false, false
	for {
		select {
		case <-writeTimeoutTicker.C:
			if !writeExpire {
				log.InfoContext(ctx, "write timeout is expired", slog.Duration("timeout", c.cfg.WriteTimeout))
			}
			writeExpire = true

			if err := conn.CloseWrite(); err != nil {
				return fmt.Errorf("failed to close write: %w", err)
			}

			if writeExpire && readExpire {
				return nil
			}

		case <-readTimeoutTicker.C:
			if !readExpire {
				log.InfoContext(ctx, "read timeout is expired", slog.Duration("timeout", c.cfg.ReadTimeout))
			}
			readExpire = true

			if err := conn.CloseRead(); err != nil {
				return fmt.Errorf("failed to close read: %w", err)
			}

			if writeExpire && readExpire {
				return nil
			}
		case err := <-done:
			return err
		}

	}

}

func (c *Client) readFromConn(ctx context.Context, conn *net.TCPConn, dataChan chan<- []byte, done chan<- error) {
	defer close(dataChan)

	buf := make([]byte, c.cfg.BufSize)

	readTotal := 0
	writeTotal := 0
	writeClosed := false

	for {
		readSize, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				log.InfoContext(ctx, "got EOF in reader", slog.Int("total", readTotal))
				done <- nil
				return
			}
			log.ErrorContext(ctx, "failed to read data", slog.Any("error", err))
			done <- err
			return
		}

		readTotal += readSize

		log.DebugContext(ctx, "read data", slog.Int("size", readSize), slog.Int("total", readTotal))

		if c.cfg.Echo && !writeClosed {
			writeSize, err := conn.Write(buf)
			if err != nil {
				if err == io.EOF {
					log.InfoContext(ctx, "got EOF in writer", slog.Int("total", writeTotal))
				} else {
					log.ErrorContext(ctx, "failed to write data", slog.Any("error", err))
				}
				writeClosed = true
			}
			writeTotal += writeSize
			log.DebugContext(ctx, "write back data", slog.Int("size", writeSize), slog.Int("total", writeTotal))
		}

		// dataChan <- buf[:readSize]
	}

}
