package main

import (
    "context"
    "fmt"
    "log"

    "github.com/asta-kunn/backend-task/cmd/worker"
    "github.com/asta-kunn/backend-task/config"
    "github.com/joho/godotenv"
    "github.com/segmentio/kafka-go"
    "github.com/spf13/cobra"

)

func consumeKafkaMessages() {
    r := kafka.NewReader(kafka.ReaderConfig{
        Brokers:   []string{"localhost:9092"}, // Ensure this matches your actual Kafka setup
        Topic:     "scraped-data",
        Partition: 0,
        MinBytes:  10e3, // 10KB
        MaxBytes:  10e6, // 10MB
    })
    defer r.Close()

    for {
        m, err := r.ReadMessage(context.Background())
        if err != nil {
            log.Printf("Failed to read message from Kafka: %v", err)
            continue
        }
        fmt.Printf("Message at offset %d: %s = %s\n", m.Offset, string(m.Key), string(m.Value))
    }
}

func main() {
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }

    config.InitConfig()

    var rootCmd = &cobra.Command{Use: "app"}
    rootCmd.AddCommand(worker.Cmd)  // Add worker as a subcommand

    // Optionally, you could add a command to start consuming Kafka messages directly
    kafkaCmd := &cobra.Command{
        Use:   "kafkaConsume",
        Short: "Consume messages from Kafka",
        Run: func(cmd *cobra.Command, args []string) {
            consumeKafkaMessages()
        },
    }
    rootCmd.AddCommand(kafkaCmd)

    if err := rootCmd.Execute(); err != nil {
        log.Fatalf("Command execution failed: %v", err)
    }
}
