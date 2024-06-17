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

func InitConfig() {
    viper.AddConfigPath(".")
    viper.SetConfigName(".env")
    viper.SetConfigType("env")

    if err := viper.ReadInConfig(); err != nil {
        log.Fatalf("Error reading config file, %s", err)
    }

    viper.AutomaticEnv()
}

func GetAppID() string {
    return viper.GetString("APP_ID")
}

func InitRedis() *redis.Client {
    return redis.NewClient(&redis.Options{
        Addr: viper.GetString("REDIS_ADDR"),
    })
}

func InitKafkaWriter() *kafka.Writer {
    return &kafka.Writer{
        Addr:     kafka.TCP(viper.GetString("KAFKA_ADDR")),
        Topic:    viper.GetString("KAFKA_TOPIC"),
        Balancer: &kafka.LeastBytes{},
    }
}

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
        log.Fatal("Could not initialize Jaeger tracer:", err)
    }
    return tracer, closer
}

func CheckKafkaConnection(broker string) error {
    conn, err := kafka.Dial("tcp", broker)
    if err != nil {
        return err
    }
    defer conn.Close()
    return nil
}

func CheckRedisConnection(addr string) error {
    client := redis.NewClient(&redis.Options{
        Addr: addr,
    })
    defer client.Close()
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    return client.Ping(ctx).Err()
}

func CheckJaegerConnection(agentHost string) error {
    cfg := config.Configuration{
        ServiceName: "check-jaeger",
        Sampler: &config.SamplerConfig{
            Type:  "const",
            Param: 1,
        },
        Reporter: &config.ReporterConfig{
            LogSpans:           true,
            LocalAgentHostPort: agentHost,
        },
    }
    _, closer, err := cfg.NewTracer(
        config.Logger(jaeger.StdLogger),
    )
    if err != nil {
        return err
    }
    defer closer.Close()
    return nil
}
