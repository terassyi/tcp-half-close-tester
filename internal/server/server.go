package server

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"

	"github.com/terassyi/tcp-half-close-tester/internal/logutils"
	"github.com/terassyi/tcp-half-close-tester/internal/streamer"
)

var log *slog.Logger

type Server struct {
	cfg *Config
}

func New(cfg *Config) *Server {
	level, _ := logutils.LogLevelFromString(cfg.LogLevel)
	log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	log = log.WithGroup("server")
	log.Info("new server", slog.String("listen", cfg.Listen), slog.String("file", cfg.File), slog.Int("chunk", cfg.Chunk))
	return &Server{cfg}
}

func (s *Server) Run(ctx context.Context) error {
	log = log.With("listen", s.cfg.Listen)

	addr, err := net.ResolveTCPAddr("tcp", s.cfg.Listen)
	if err != nil {
		return fmt.Errorf("failed to resolve host: %w", err)
	}

	size, err := fileSize(s.cfg.File)
	if err != nil {
		return fmt.Errorf("failed to get file size: %w", err)
	}

	log.InfoContext(ctx, "ready to send data", slog.Int64("size", size))

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	defer func() {
		if err := listener.Close(); err != nil {
			log.ErrorContext(ctx, "failed to close a listener", slog.Any("error", err))
		}
	}()

	log.InfoContext(ctx, "start the server")

	for {
		if ctx.Err() != nil {
			return fmt.Errorf("propagated from ctx: %w", err)
		}

		conn, err := listener.AcceptTCP()
		if err != nil {
			return fmt.Errorf("failed to accept: %w", err)
		}

		closeReadChan := make(chan struct{})
		closeWriteChan := make(chan struct{})

		go func() {
			defer conn.Close()

			file, err := os.Open(s.cfg.File)
			if err != nil {
				log.ErrorContext(ctx, "failed to open file", slog.Any("error", err), slog.String("file", s.cfg.File))
				return
			}
			defer func() {
				file.Close()
				closeWriteChan <- struct{}{}
			}()

			streamer := streamer.New(conn, file, s.cfg.Chunk, s.cfg.Interval, log)

			if err := streamer.Stream(ctx); err != nil {
				log.ErrorContext(ctx, "failed to stream data", slog.Any("error", err))
				return
			}
		}()

		go func() {
			total := 0
			buf := make([]byte, 1024)

			defer func() {
				closeReadChan <- struct{}{}
			}()

			for {
				l, err := conn.Read(buf)
				if err != nil {
					if err == io.EOF {
						log.InfoContext(ctx, "got EOF in reader", slog.Int("total", total))
						return
					}
					log.ErrorContext(ctx, "failed to read from client", slog.Any("error", err))
					return
				}
				total += l
				log.DebugContext(ctx, "read from client", slog.Int("size", l), slog.Int("total", total))
			}
		}()

		go func() {
			closeRead, closeWrite := false, false
			for {
				select {
				case <-closeReadChan:
					closeRead = true
				case <-closeWriteChan:
					closeWrite = true
				default:
					if closeRead && closeWrite {
						log.InfoContext(ctx, "close server side connection")
						conn.Close()
						return
					}
				}
			}
		}()
	}

}

func fileSize(file string) (int64, error) {
	stat, err := os.Stat(file)
	if err != nil {
		return 0, err
	}

	return stat.Size(), nil
}
