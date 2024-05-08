package cmd

import (
	"context"
	"en-order/cmd/migration"
	"log"
	"os"
	"os/signal"
	"syscall"

	"en-order/cmd/consumer"
	"en-order/cmd/http"

	"en-order/pkg/logger"

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
					consumer.RunConsumerOrderExpire(ctx)
				}()

				go func() {
					consumer.RunConsumerOrderComplete(ctx)
				}()
			},
		},
		{
			Use:   "consumer:expire-order",
			Short: "Run Consumer Expire Order",
			Run: func(cmd *cobra.Command, args []string) {
				consumer.RunConsumerOrderExpire(ctx)
			},
		},
		{
			Use:   "consumer:complete-order",
			Short: "Run Consumer Complete Order",
			Run: func(cmd *cobra.Command, args []string) {
				consumer.RunConsumerOrderComplete(ctx)
			},
		},
	}

	rootCmd.AddCommand(cmd...)
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
