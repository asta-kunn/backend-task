package config

import (
	"context"
	"io"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/opentracing/opentracing-go"
	"github.com/segmentio/kafka-go"
	"github.com/spf13/viper"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-lib/metrics/prometheus"
)

// Initialize the configuration using viper.
func InitConfig() {
    viper.AddConfigPath(".")
    viper.SetConfigName(".env")
    viper.SetConfigType("env")
    if err := viper.ReadInConfig(); err != nil {
        log.Fatalf("Error reading config file: %s", err)
    }
    viper.AutomaticEnv()
}

// Initialize and return a new Redis client.
func InitRedis() *redis.Client {
    client := redis.NewClient(&redis.Options{
        Addr: viper.GetString("REDIS_ADDR"),
    })
    _, err := client.Ping(context.Background()).Result()
    if err != nil {
        log.Fatalf("Failed to connect to Redis: %v", err)
    }
    return client
}

// Initialize and return a Kafka writer.
func InitKafkaWriter() *kafka.Writer {
    writer := &kafka.Writer{
        Addr:     kafka.TCP(viper.GetString("KAFKA_ADDR")),
        Topic:    viper.GetString("KAFKA_TOPIC"),
        Balancer: &kafka.LeastBytes{},
    }
    conn, err := kafka.Dial("tcp", viper.GetString("KAFKA_ADDR"))
    if err != nil {
        log.Fatalf("Failed to connect to Kafka: %v", err)
    }
    conn.Close()
    return writer
}
// GetAppID retrieves the application ID from the environment/config.
func GetAppID() string {
    return viper.GetString("APP_ID")     // Get the value associated with the key "APPID"
}
// Initialize Jaeger for distributed tracing.
func InitJaeger(serviceName string) (opentracing.Tracer, io.Closer) {
    cfg := config.Configuration{
        ServiceName: serviceName,
        Sampler: &config.SamplerConfig{
            Type:  "const",
            Param: 1,
        },
        Reporter: &config.ReporterConfig{
            LogSpans:           true,
            LocalAgentHostPort: viper.GetString("JAEGER_AGENT_HOST"),
        },
    }
    tracer, closer, err := cfg.NewTracer(
        config.Logger(jaeger.StdLogger),
        config.Metrics(prometheus.New()),
    )
    if err != nil {
        log.Fatalf("Could not initialize Jaeger tracer: %v", err)
    }
    return tracer, closer
}


// CheckKafkaConnection tests the connection to a Kafka broker.
func CheckKafkaConnection(broker string) error {
    conn, err := kafka.Dial("tcp", broker)
    if err != nil {
        return err  // Handle errors connecting to Kafka
    }
    defer conn.Close()
    return nil
}

// CheckRedisConnection tests the connection to a Redis server.
func CheckRedisDataConnection(addr string) error {
    client := redis.NewClient(&redis.Options{
        Addr: addr,  // Address to the Redis server
    })
    defer client.Close()
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second) // Context with timeout for connection test
    defer cancel()
    return client.Ping(ctx).Err()  // Ping Redis server
}

// CheckJaegerConnection tests the connection to a Jaeger agent.
func CheckJaegerConnection(agentHost string) error {
    cfg := config.Configuration{
        ServiceName: "check-jaeger",
        Sampler: &config.SamplerConfig{
            Type:  "const",     // Sampler type
            Param: 1,           // Sampler parameter
        },
        Reporter: &config.ReporterConfig{
            LogSpans:           true,
            LocalAgentHostPort: agentHost, // Jaeger agent host
        },
    }
    _, closer, err := cfg.NewTracer(
        config.Logger(jaeger.StdLogger), // Logging configuration
    )
    if err != nil {
        return err  // Handle tracer initialization errors
    }
    defer closer.Close()
    return nil
}
