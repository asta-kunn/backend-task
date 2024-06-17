package worker

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/asta-kunn/backend-task/config"
	"github.com/asta-kunn/backend-task/pkg/scraper"
	"github.com/segmentio/kafka-go"
	"github.com/spf13/cobra"
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
    config.InitConfig()

    redisClient := config.InitRedis()
    kafkaWriter := config.InitKafkaWriter()
    defer kafkaWriter.Close()

    tracer, closer := config.InitJaeger("backend-task")
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
