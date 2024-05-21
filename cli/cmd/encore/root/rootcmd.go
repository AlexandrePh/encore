package root

import (
	"fmt"

	"github.com/logrusorgru/aurora/v3"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"encr.dev/cli/internal/telemetry"
	"encr.dev/pkg/errlist"
)

var (
	Verbosity int
	traceFile string

	// TraceFile is the file to write trace logs to.
	// If nil (the default), trace logs are not written.
	TraceFile *string
)

var Cmd = &cobra.Command{
	Use:           "encore",
	Short:         "encore is the fastest way of developing backend applications",
	SilenceErrors: true, // We'll handle displaying an error in our main func
	CompletionOptions: cobra.CompletionOptions{
		HiddenDefaultCmd: true, // Hide the "completion" command from help (used for generating auto-completions for the shell)
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if traceFile != "" {
			TraceFile = &traceFile
		}

		level := zerolog.InfoLevel
		if Verbosity == 1 {
			level = zerolog.DebugLevel
		} else if Verbosity >= 2 {
			level = zerolog.TraceLevel
		}

		if Verbosity >= 1 {
			errlist.Verbose = true
		}
		log.Logger = log.Logger.Level(level)

		if !telemetry.ShouldShownWarning() && cmd.Short != "version" {
			fmt.Println()
			fmt.Println(aurora.Sprintf("%s: This CLI tool collects usage data to help us improve the product.", aurora.Bold("Notice")))
			fmt.Println("        You can disable this by running 'encore telemetry disable'.\n")
			telemetry.SetShownWarning()
		}
	},
}

func init() {
	Cmd.PersistentFlags().CountVarP(&Verbosity, "verbose", "v", "verbose output")
	Cmd.PersistentFlags().StringVar(&traceFile, "trace", "", "file to write execution trace data to")
}
