package main

import (
	"context"
	"time"

	"github.com/Dicklesworthstone/ntm/internal/nzm"
	"github.com/Dicklesworthstone/ntm/internal/output"
	"github.com/Dicklesworthstone/ntm/internal/zellij"
	"github.com/spf13/cobra"
)

var sendCmd = &cobra.Command{
	Use:   "send SESSION TARGET TEXT",
	Short: "Send text to a pane",
	Long: `Send text or commands to a specific pane in a session.

Target can be:
  - Agent type: "cc", "cod", "gmi" (sends to first matching pane)
  - Pane name: "cc_1", "gmi_2" (short form)
  - Full pane name: "proj__cc_1" (includes session prefix)

Examples:
  # Send text to first Claude pane
  nzm send myproj cc "hello world"

  # Send text to specific Claude pane with Enter
  nzm send myproj cc_2 "npm test" --enter

  # Send Ctrl+C to interrupt a pane
  nzm send myproj cc_1 --interrupt`,
	Args: cobra.RangeArgs(2, 3),
	RunE: runSend,
}

var (
	sendEnter     bool
	sendInterrupt bool
)

func init() {
	rootCmd.AddCommand(sendCmd)

	sendCmd.Flags().BoolVarP(&sendEnter, "enter", "e", false, "Press Enter after sending text")
	sendCmd.Flags().BoolVarP(&sendInterrupt, "interrupt", "i", false, "Send Ctrl+C interrupt")
}

func runSend(cmd *cobra.Command, args []string) error {
	session := args[0]
	target := args[1]
	text := ""
	if len(args) > 2 {
		text = args[2]
	}

	client := zellij.NewClient()
	sender := nzm.NewSender(client)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := nzm.SendOptions{
		Session:   session,
		Target:    target,
		Text:      text,
		Enter:     sendEnter,
		Interrupt: sendInterrupt,
	}

	if err := sender.Send(ctx, opts); err != nil {
		return err
	}

	formatter := output.NZMDefaultFormatter(jsonFlag)
	if formatter.IsJSON() {
		return formatter.JSON(map[string]interface{}{
			"action":    "send",
			"session":   session,
			"target":    target,
			"text":      text,
			"enter":     sendEnter,
			"interrupt": sendInterrupt,
			"success":   true,
		})
	}

	// Silent success for text output (like echo)
	return nil
}
