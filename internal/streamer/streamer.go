package streamer

import (
	"context"
	"io"
	"log/slog"
	"net"
	"os"
)

type Streamer struct {
	conn   *net.TCPConn
	file   *os.File
	chunk  int
	logger *slog.Logger
}

func New(conn *net.TCPConn, file *os.File, chunkSize int, logger *slog.Logger) *Streamer {
	return &Streamer{conn: conn, file: file, chunk: chunkSize, logger: logger}
}

func (s *Streamer) Stream(ctx context.Context) error {

	if s.chunk == 0 {
		return s.stream(ctx)
	} else {
		return s.streamWithChunk(ctx)
	}
}

func (s *Streamer) stream(ctx context.Context) error {

	s.logger.InfoContext(ctx, "start to stream")

	size, err := io.Copy(s.conn, s.file)
	if err != nil {
		return err
	}

	s.logger.InfoContext(ctx, "finish streaming", slog.Int64("size", size))

	return nil
}

func (s *Streamer) streamWithChunk(ctx context.Context) error {
	buf := make([]byte, s.chunk)

	s.logger.InfoContext(ctx, "start to stream with chunk", slog.Int("chunk", s.chunk))

	total := 0

	for {
		readSize, err := s.file.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}
		s.logger.DebugContext(ctx, "read data", slog.Int("size", readSize))

		writeSize, err := s.conn.Write(buf)
		if err != nil {
			return err
		}

		total += writeSize
		s.logger.DebugContext(ctx, "streaming data", slog.Int("size", writeSize), slog.Int("total", total))

	}

	s.logger.InfoContext(ctx, "finish streaming with chunk", slog.Int("size", total))

	return nil
}
