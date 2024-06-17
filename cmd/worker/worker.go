package worker

import (
    "context"
    "fmt"
    "log"
    "sync"

    "github.com/asta-kunn/backend-task/pkg/scraper"
    "github.com/go-redis/redis/v8"
    "github.com/segmentio/kafka-go"
    "github.com/spf13/cobra"
    "github.com/uber/jaeger-client-go"
    "github.com/uber/jaeger-client-go/config"
    "github.com/uber/jaeger-lib/metrics/prometheus"
)

var Cmd = &cobra.Command{
    Use:   "worker",
    Short: "Scrape data from Dummy API",
    Run: func(cmd *cobra.Command, args []string) {
        scrapeData()
    },
}

func scrapeData() {
    var wg sync.WaitGroup
    redisClient := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })
    kafkaWriter := kafka.Writer{
        Addr:     kafka.TCP("localhost:9092"),
        Topic:    "scraped-data",
        Balancer: &kafka.LeastBytes{},
    }
    defer kafkaWriter.Close()

    cfg := config.Configuration{
        ServiceName: "backend-task",
        Sampler: &config.SamplerConfig{
            Type:  "const",
            Param: 1,
        },
        Reporter: &config.ReporterConfig{
            LogSpans:           true,
            LocalAgentHostPort: "127.0.0.1:6831",
        },
    }
    tracer, closer, err := cfg.NewTracer(
        config.Logger(jaeger.StdLogger),
        config.Metrics(prometheus.New()),
    )
    if err != nil {
        log.Fatal("Could not initialize Jaeger tracer:", err)
    }
    defer closer.Close()

    log.Println("Connected to Kafka, Redis, and Jaeger successfully")

    for i := 1; i <= 10; i++ {
        wg.Add(3)
        go func(page int) {
            defer wg.Done()
            span := tracer.StartSpan(fmt.Sprintf("Scraping Users Page %d", page))
            users, err := scraper.ScrapeUsers(page)
            span.Finish()
            if err != nil {
                log.Printf("Error scraping users on page %d: %v\n", page, err)
                return
            }
            for _, user := range users {
                log.Printf("User: %s %s %s %s %s\n", user.Title, user.FirstName, user.LastName, user.Email, user.Gender)
                kafkaWriter.WriteMessages(context.Background(),
                    kafka.Message{
                        Key:   []byte(user.Email),
                        Value: []byte(fmt.Sprintf("User: %s %s %s %s %s", user.Title, user.FirstName, user.LastName, user.Email, user.Gender)),
                    },
                )
                redisClient.Set(context.Background(), user.Email, user, 0)
            }
        }(i)

        go func(page int) {
            defer wg.Done()
            span := tracer.StartSpan(fmt.Sprintf("Scraping Posts Page %d", page))
            posts, err := scraper.ScrapePosts(page)
            span.Finish()
            if err != nil {
                log.Printf("Error scraping posts on page %d: %v\n", page, err)
                return
            }
            for _, post := range posts {
                log.Printf("Post: Posted by %s %s:\n%s\nLikes: %d Tags: %v\nDate posted: %s\n", post.User.FirstName, post.User.LastName, post.Text, post.Likes, post.Tags, post.PublishDate)
                kafkaWriter.WriteMessages(context.Background(),
                    kafka.Message{
                        Key:   []byte(post.User.Email),
                        Value: []byte(fmt.Sprintf("Post: Posted by %s %s:\n%s\nLikes: %d Tags: %v\nDate posted: %s", post.User.FirstName, post.User.LastName, post.Text, post.Likes, post.Tags, post.PublishDate)),
                    },
                )
                redisClient.Set(context.Background(), post.User.Email, post, 0)
            }
        }(i)

        go func(page int) {
            defer wg.Done()
            span := tracer.StartSpan(fmt.Sprintf("Scraping Comments Page %d", page))
            comments, err := scraper.ScrapeComments(page)
            span.Finish()
            if err != nil {
                log.Printf("Error scraping comments on page %d: %v\n", page, err)
                return
            }
            for _, comment := range comments {
                log.Printf("Comment: %s by %s %s\n", comment.Message, comment.User.FirstName, comment.User.LastName)
                kafkaWriter.WriteMessages(context.Background(),
                    kafka.Message{
                        Key:   []byte(comment.ID),
                        Value: []byte(fmt.Sprintf("Comment: %s by %s %s", comment.Message, comment.User.FirstName, comment.User.LastName)),
                    },
                )
                redisClient.Set(context.Background(), comment.ID, comment, 0)
            }
        }(i)
    }

    wg.Wait()
    log.Println("Scraping completed")
}
