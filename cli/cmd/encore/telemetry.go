package main

import (
	"context"
	"fmt"

	"github.com/logrusorgru/aurora/v3"
	"github.com/spf13/cobra"

	"encr.dev/cli/cmd/encore/cmdutil"
	"encr.dev/cli/cmd/encore/root"
	"encr.dev/pkg/option"
	daemonpb "encr.dev/proto/encore/daemon"
)

func printTelemetryStatus(resp *daemonpb.TelemetryResponse) {
	status := aurora.Green("Enabled")
	if !resp.Enabled {
		status = aurora.Red("Disabled")
	}
	fmt.Println(aurora.Sprintf("%s\n", aurora.Bold("Encore Telemetry")))
	if root.Verbosity == 0 {
		fmt.Println(aurora.Sprintf("Status: %s", status))
	} else {
		fmt.Println(aurora.Sprintf("Status:     %s", status))
		fmt.Println(aurora.Sprintf("Install ID: %s", resp.AnonId))
	}
	fmt.Println(aurora.Sprintf("\nLearn more: %s", aurora.Underline("https://encore.dev/docs/telemetry")))
}

func updateTelemetry(ctx context.Context, enabled option.Option[bool]) {
	daemon := cmdutil.ConnectDaemon(ctx)
	resp, err := daemon.Telemetry(ctx, &daemonpb.TelemetryRequest{Enabled: enabled.PtrOrNil()})
	if err != nil {
		fatalf("could not execute telemetry request: %s", err)
	}
	printTelemetryStatus(resp)
}

var telemetryCommand = &cobra.Command{
	Use:   "telemetry",
	Short: "Reports the current telemetry status",

	Run: func(cmd *cobra.Command, args []string) {
		updateTelemetry(cmd.Context(), option.None[bool]())
	},
}

var telemetryEnableCommand = &cobra.Command{
	Use:   "enable",
	Short: "Enables telemetry reporting",
	Run: func(cmd *cobra.Command, args []string) {
		updateTelemetry(cmd.Context(), option.Some(true))
	},
}

var telemetryDisableCommand = &cobra.Command{
	Use:   "disable",
	Short: "Disables telemetry reporting",
	Run: func(cmd *cobra.Command, args []string) {
		updateTelemetry(cmd.Context(), option.Some(false))
	},
}

func init() {
	telemetryCommand.AddCommand(telemetryEnableCommand, telemetryDisableCommand)
	rootCmd.AddCommand(telemetryCommand)
}
