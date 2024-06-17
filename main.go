package main

import (
    "github.com/asta-kunn/backend-task/cmd/worker"
    "github.com/spf13/cobra"
)

func main() {
    var rootCmd = &cobra.Command{Use: "app"}
    rootCmd.AddCommand(worker.Cmd)
    rootCmd.Execute()
}
