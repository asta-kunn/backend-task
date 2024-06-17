package main

import (
    "github.com/asta-kunn/backend-task/cmd/worker"
    "github.com/spf13/cobra"
    "github.com/joho/godotenv"
    "log"
)

func main() {
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }

    var rootCmd = &cobra.Command{Use: "app"}
    rootCmd.AddCommand(worker.Cmd)
    rootCmd.Execute()
}
