package cmd

import (
	"github.com/spf13/cobra"
)

// NewRestartCommand initializes new kool start command
func NewRestartCommand(stop KoolService, start KoolService) *cobra.Command {
	stopTask := NewKoolTask("Stopping", stop)
	startTask := NewKoolTask("Starting", start)

	return &cobra.Command{
		Use:                   "restart",
		Short:                 "Restart containers - the same as stop followed by start.",
		Run:                   LongTaskCommandRunFunction(stopTask, startTask),
		DisableFlagsInUseLine: true,
	}
}

func init() {
	rootCmd.AddCommand(NewRestartCommand(NewKoolStop(), NewKoolStart()))
}
