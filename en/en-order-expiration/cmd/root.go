package cmd

import (
	"context"
	"en-order-expiration/pkg/logger"
	"log"
	"os"
	"os/signal"
	"syscall"

	"en-order-expiration/cmd/job"
	"github.com/spf13/cobra"
)

func Start() {
	logger.SetJSONFormatter()
	ctx, cancel := context.WithCancel(context.Background())

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-quit
		cancel()
	}()

	rootCmd := &cobra.Command{}
	cmd := []*cobra.Command{
		{
			Use:   "job:expire-order",
			Short: "Run Job Expire Order",
			Run: func(cmd *cobra.Command, args []string) {
				job.RunJobExpireOrder(ctx)
			},
		},
	}

	rootCmd.AddCommand(cmd...)
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
