package cmd

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/spf13/cobra"
	"github.com/terassyi/tcp-half-close-tester/internal/client"
)

var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Run the tcp-half-close-tester client",
	Run:   runClient,
}

var clientCfg = &client.Config{}

func init() {
	rootCmd.AddCommand(clientCmd)

	clientCmd.Flags().StringVarP(&clientCfg.Server, "server", "s", "127.0.0.1:4000", "Server address to connect")
	clientCmd.Flags().DurationVarP(&clientCfg.ReadTimeout, "read-timeout", "r", time.Minute*10, "Read timeout to close connection")
	clientCmd.Flags().DurationVarP(&clientCfg.WriteTimeout, "write-timeout", "w", time.Minute*10, "Write timeout to close connection")
	clientCmd.Flags().IntVarP(&clientCfg.BufSize, "buf-size", "b", 1024, "Buffer size to read")
	clientCmd.Flags().BoolVar(&clientCfg.Echo, "echo", false, "Run as echo client")
	clientCmd.Flags().StringVar(&clientCfg.LogLevel, "log-level", "info", "Log level(debug, info, warn, error)")
}

func runClient(cmd *cobra.Command, args []string) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	client := client.New(clientCfg)

	done := make(chan error)
	go func() {
		done <- client.Run(ctx)
	}()

	select {
	case <-ctx.Done():
		log.Info("Signal received. The client will be stopped after 5 seconds.")
		time.Sleep(5 * time.Second)
	case err := <-done:
		if err != nil {
			log.Error("client run failed", slog.Any("error", err))
		}
	}
}
