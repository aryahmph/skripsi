package cmd

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"ecst-order/cmd/consumer"
	"ecst-order/cmd/http"
	"ecst-order/cmd/migration"

	"ecst-order/pkg/logger"

	"github.com/spf13/cobra"
)

func Start() {
	rootCmd := &cobra.Command{}
	logger.SetJSONFormatter()
	ctx, cancel := context.WithCancel(context.Background())

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-quit
		cancel()
	}()

	migrateCmd := &cobra.Command{
		Use:   "db:migrate",
		Short: "database migration",
		Run: func(c *cobra.Command, args []string) {
			migration.MigrateDatabase()
		},
	}

	migrateCmd.Flags().BoolP("version", "", false, "print version")
	migrateCmd.Flags().StringP("dir", "", "database/migration/", "directory with migration files")
	migrateCmd.Flags().StringP("table", "", "db", "migrations table name")
	migrateCmd.Flags().BoolP("verbose", "", false, "enable verbose mode")
	migrateCmd.Flags().BoolP("guide", "", false, "print help")

	cmd := []*cobra.Command{
		migrateCmd,
		{
			Use:   "serve-http",
			Short: "Run HTTP Server",
			Run: func(cmd *cobra.Command, args []string) {
				http.Start(ctx)
			},
			PreRun: func(cmd *cobra.Command, args []string) {
				go func() {
					consumer.RunConsumerOrderExpirationComplete(ctx)
				}()

				go func() {
					consumer.RunConsumerCreateTicket(ctx)
				}()

				go func() {
					consumer.RunConsumerUpdateTicket(ctx)
				}()

				go func() {
					consumer.RunConsumerCreatePayment(ctx)
				}()
			},
		},
		{
			Use:   "consumer:expiration-complete",
			Short: "Run Consumer Expire Order",
			Run: func(cmd *cobra.Command, args []string) {
				consumer.RunConsumerOrderExpirationComplete(ctx)
			},
		},
		{
			Use:   "consumer:create-ticket",
			Short: "Run Consumer Create Ticket",
			Run: func(cmd *cobra.Command, args []string) {
				consumer.RunConsumerCreateTicket(ctx)
			},
		},
		{
			Use:   "consumer:update-ticket",
			Short: "Run Consumer Update Ticket",
			Run: func(cmd *cobra.Command, args []string) {
				consumer.RunConsumerUpdateTicket(ctx)
			},
		},
		{
			Use:   "consumer:create-payment",
			Short: "Run Consumer Create Payment",
			Run: func(cmd *cobra.Command, args []string) {
				consumer.RunConsumerCreatePayment(ctx)
			},
		},
	}

	rootCmd.AddCommand(cmd...)
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
