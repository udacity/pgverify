package main

import (
	"fmt"

	"github.com/cjfinnell/pgverify"
	"github.com/jackc/pgx"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Flags
var (
	aliasesFlag, excludeSchemasFlag, excludeTablesFlag, includeSchemasFlag, includeTablesFlag *[]string
	logLevelFlag, strategyFlag                                                                *string
)

func init() {
	aliasesFlag = rootCmd.Flags().StringSlice("aliases", []string{}, "alias names for the supplied targets (comma separated)")
	excludeSchemasFlag = rootCmd.Flags().StringSlice("exclude-schemas", []string{}, "schemas to skip verification, ignored if '--include-schemas' used (comma separated)")
	excludeTablesFlag = rootCmd.Flags().StringSlice("exclude-tables", []string{}, "tables to skip verification, ignored if '--include-tables' used (comma separated)")
	includeSchemasFlag = rootCmd.Flags().StringSlice("include-schemas", []string{}, "schemas to verify (comma separated)")
	includeTablesFlag = rootCmd.Flags().StringSlice("include-tables", []string{}, "tables to verify (comma separated)")

	logLevelFlag = rootCmd.Flags().String("level", "info", "logging level")
	strategyFlag = rootCmd.Flags().StringP("strategy", "s", pgverify.StrategyFull, "strategy to use for verification")
}

var rootCmd = &cobra.Command{
	Use:  "pgverify [flags] target-uri...",
	Long: `Verify data consistency between PostgreSQL syntax compatible databases.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var targets []pgx.ConnConfig
		for _, target := range args {
			connConfig, err := pgx.ParseURI(target)
			if err != nil {
				return fmt.Errorf("invalid target URI %s: %w", target, err)
			}
			targets = append(targets, connConfig)
		}

		opts := []pgverify.Option{
			pgverify.IncludeTables(*includeTablesFlag...),
			pgverify.ExcludeTables(*excludeTablesFlag...),
			pgverify.IncludeSchemas(*includeSchemasFlag...),
			pgverify.ExcludeSchemas(*excludeSchemasFlag...),
			pgverify.WithStrategy(*strategyFlag),
		}

		logger := log.New()
		logger.SetFormatter(&log.TextFormatter{})
		levelInt, err := log.ParseLevel(*logLevelFlag)
		if err != nil {
			levelInt = log.InfoLevel
		}
		logger.SetLevel(log.Level(levelInt))
		opts = append(opts, pgverify.WithLogger(logger))

		if len(*aliasesFlag) > 0 {
			opts = append(opts, pgverify.WithAliases(*aliasesFlag))
		}

		report, err := pgverify.Verify(targets, opts...)
		report.WriteAsTable(cmd.OutOrStdout())
		return err
	},
}
