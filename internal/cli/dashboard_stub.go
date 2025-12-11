package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Stub dashboard command to keep builds green when the full TUI dashboard
// implementation is unavailable in this checkout.
func newDashboardCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "dashboard",
		Short:  "TUI dashboard (stubbed)",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("dashboard not available in this build")
		},
	}
}
