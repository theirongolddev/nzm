package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/Dicklesworthstone/ntm/internal/resilience"
	"github.com/Dicklesworthstone/ntm/internal/tmux"
)

func newMonitorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "_monitor <session>",
		Short:  "Internal command to run resilience monitor (hidden)",
		Hidden: true,
		Args:   cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			session := args[0]
			return runMonitor(session)
		},
	}
	return cmd
}

func runMonitor(session string) error {
	// 1. Validate session
	if !tmux.SessionExists(session) {
		return fmt.Errorf("session %q not found", session)
	}

	dir := cfg.GetProjectDir(session)

	// 2. Start monitor
	monitor := resilience.NewMonitor(session, dir, cfg)

	// 3. Scan for existing agents
	if err := monitor.ScanAndRegisterAgents(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to scan agents: %v\n", err)
	}

	// 4. Start monitoring
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	monitor.Start(ctx)
	fmt.Printf("Monitoring session %s... (PID: %d)\n", session, os.Getpid())

	// 5. Wait for signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	monitor.Stop()
	return nil
}
