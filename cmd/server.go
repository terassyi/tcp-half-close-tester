package cmd

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/spf13/cobra"
	"github.com/terassyi/tcp-half-close-tester/internal/server"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run the tcp-half-close-tester server",
	Run:   runServer,
}

var serverCfg = &server.Config{}

func init() {
	rootCmd.AddCommand(serverCmd)

	serverCmd.Flags().StringVarP(&serverCfg.File, "file", "f", "", "File path to send")
	serverCmd.Flags().StringVarP(&serverCfg.Listen, "listen", "l", "0.0.0.0:4000", "Listen address:port")
	serverCmd.Flags().IntVarP(&serverCfg.Chunk, "chunk", "c", 1024, "Chunk size to write")
	serverCmd.Flags().DurationVarP(&serverCfg.Interval, "interval", "i", time.Second*0, "Interval to write data")
	serverCmd.Flags().StringVar(&serverCfg.LogLevel, "log-level", "info", "Log level(debug, info, warn, error)")
}

func runServer(cmd *cobra.Command, args []string) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	server := server.New(serverCfg)

	done := make(chan error)
	go func() {
		done <- server.Run(ctx)
	}()

	select {
	case <-ctx.Done():
		log.Info("Signal received. The server will be stopped after 5 seconds.")
		time.Sleep(5 * time.Second)
	case err := <-done:
		if err != nil {
			log.Error("server run failed", slog.Any("error", err))
		}
	}

}
