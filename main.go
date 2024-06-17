package main

import (
    "log"
    "github.com/asta-kunn/backend-task/cmd/worker"
    "github.com/asta-kunn/backend-task/config"
    "github.com/spf13/cobra"
    "github.com/joho/godotenv"
    "github.com/spf13/viper"
)

func main() {
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }

    config.InitConfig()

    kafkaBroker := viper.GetString("KAFKA_ADDR")
    redisAddr := viper.GetString("REDIS_ADDR")
    jaegerAgentHost := viper.GetString("JAEGER_AGENT_HOST")

    if err := config.CheckKafkaConnection(kafkaBroker); err != nil {
        log.Fatalf("Failed to connect to Kafka: %v", err)
    } else {
        log.Println("Connected to Kafka successfully")
    }

    if err := config.CheckRedisConnection(redisAddr); err != nil {
        log.Fatalf("Failed to connect to Redis: %v", err)
    } else {
        log.Println("Connected to Redis successfully")
    }

    if err := config.CheckJaegerConnection(jaegerAgentHost); err != nil {
        log.Fatalf("Failed to connect to Jaeger: %v", err)
    } else {
        log.Println("Connected to Jaeger successfully")
    }

    var rootCmd = &cobra.Command{Use: "app"}
    rootCmd.AddCommand(worker.Cmd)
    rootCmd.Execute()
}
