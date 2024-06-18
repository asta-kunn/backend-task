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

// Cmd defines the command line command for the worker process.
var Cmd = &cobra.Command{
    Use:   "worker",
    Short: "Scrape data from Dummy API", // Short description of the command
    Run: func(cmd *cobra.Command, args []string) {
        scrapeData() // Function called when the command is executed
    },
}

// scrapeData orchestrates the data scraping process.
func scrapeData() {
    var wg sync.WaitGroup // Wait group to manage concurrent goroutines
    config.InitConfig() // Initialize configuration settings

    redisClient := config.InitRedis() // Initialize Redis client
    kafkaWriter := config.InitKafkaWriter() // Initialize Kafka writer
    defer kafkaWriter.Close() // Ensure Kafka writer is closed after function exits

    tracer, closer := config.InitJaeger("backend-task") // Initialize Jaeger for tracing
    defer closer.Close() // Ensure Jaeger is closed after function exits

    for i := 1; i <= 10; i++ { // Loop through pages 1 to 10
        wg.Add(3) // Add three tasks to the wait group for users, posts, and comments

        // Goroutine to scrape user data
        go func(page int) {
            defer wg.Done() // Notify wait group on goroutine completion
            span := tracer.StartSpan(fmt.Sprintf("Scraping Users Page %d", page)) // Start a tracing span
            users, err := scraper.ScrapeUsers(page) // Scrape users from API
            span.Finish() // Finish the tracing span
            if err != nil {
                log.Printf("Error scraping users on page %d: %v\n", page, err)
                return
            }
            // Send user data to Kafka
            for _, user := range users {
                message := kafka.Message{
                    Key:   []byte(user.Email),
                    Value: []byte(fmt.Sprintf("User: %s %s %s %s %s", user.Title, user.FirstName, user.LastName, user.Email, user.Gender)),
                }
                if err := kafkaWriter.WriteMessages(context.Background(), message); err != nil {
                    log.Printf("Failed to write user message to Kafka: %v", err)
                } else {
                    log.Printf("Successfully wrote user message to Kafka: %s", message.Value)
                }
            }
        }(i)

        // Goroutine to scrape post data
        go func(page int) {
            defer wg.Done() // Notify wait group on goroutine completion
            span := tracer.StartSpan(fmt.Sprintf("Scraping Posts Page %d", page)) // Start a tracing span
            posts, err := scraper.ScrapePosts(page) // Scrape posts from API
            span.Finish() // Finish the tracing span
            if err != nil {
                log.Printf("Error scraping posts on page %d: %v\n", page, err)
                return
            }
            // Process each post
            for _, post := range posts {
                log.Printf("Post: Posted by %s %s:\n%s\nLikes: %d Tags: %v\nDate posted: %s\n", post.User.FirstName, post.User.LastName, post.Text, post.Likes, post.Tags, post.PublishDate)
                kafkaWriter.WriteMessages(context.Background(),
                    kafka.Message{
                        Key:   []byte(post.User.Email),
                        Value: []byte(fmt.Sprintf("Post: Posted by %s %s:\n%s\nLikes: %d Tags: %v\nDate posted: %s", post.User.FirstName, post.User.LastName, post.Text, post.Likes, post.Tags, post.PublishDate)),
                    },
                )
                redisClient.Set(context.Background(), post.User.Email, post, 0) // Save post data in Redis
            }
        }(i)

        // Goroutine to scrape comment data
        go func(page int) {
            defer wg.Done() // Notify wait group on goroutine completion
            span := tracer.StartSpan(fmt.Sprintf("Scraping Comments Page %d", page)) // Start a tracing span
            comments, err := scraper.ScrapeComments(page) // Scrape comments from API
            span.Finish() // Finish the tracing span
            if err != nil {
                log.Printf("Error scraping comments on page %d: %v\n", page, err)
                return
            }
            // Process each comment
            for _, comment := range comments {
                log.Printf("Comment: %s by %s %s\n", comment.Message, comment.User.FirstName, comment.User.LastName)
                kafkaWriter.WriteMessages(context.Background(),
                    kafka.Message{
                        Key:   []byte(comment.ID),
                        Value: []byte(fmt.Sprintf("Comment: %s by %s %s", comment.Message, comment.User.FirstName, comment.User.LastName)),
                    },
                )
                redisClient.Set(context.Background(), comment.ID, comment, 0) // Save comment data in Redis
            }
        }(i)
    }

    wg.Wait() // Wait for all goroutines to finish
    log.Println("Scraping completed") // Log completion of scraping process
}
