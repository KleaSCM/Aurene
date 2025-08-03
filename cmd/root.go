/*
Author: KleaSCM
Email: KleaSCM@gmail.com
File: root.go
Description: Root CLI command for the Aurene scheduler with cobra integration
and command structure for running, adding tasks, and viewing stats.
*/

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	cfgFile string
	verbose bool
)

var rootCmd = &cobra.Command{
	Use:   "aurene",
	Short: "🌌 Aurene - The Crystalline CPU Scheduler",
	Long: `Aurene is a real, production-ready CPU scheduler that can replace
the Linux scheduler. It implements MLFQ, FCFS, Round-Robin, and SJF
algorithms with preemption, task dispatching, and system integration.

Features:
• Real CPU scheduling with 250Hz tick rate
• MLFQ (Multi-Level Feedback Queue) as default
• Task injection and dynamic workload management
• Performance metrics and logging
• System integration capabilities

Example usage:
  aurene run                    # Start the scheduler
  aurene add --name "Firefox"   # Add a new task
  aurene stats                  # View performance metrics
  aurene reset                  # Clear all tasks and logs`,
	Version: "1.0.0",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.aurene.toml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}

func initConfig() {
	if verbose {
		fmt.Println("Verbose mode enabled")
	}
}
