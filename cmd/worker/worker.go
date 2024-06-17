package worker

import (
    "fmt"
    "github.com/asta-kunn/backend-task/pkg/scraper"
    "github.com/spf13/cobra"
    "sync"
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

    for i := 1; i <= 10; i++ {
        wg.Add(2)
        go func(page int) {
            defer wg.Done()
            users, err := scraper.ScrapeUsers(page)
            if err != nil {
                fmt.Printf("Error scraping users: %v\n", err)
                return
            }
            for _, user := range users {
                fmt.Printf("User: %s %s %s %s %s\n", user.Title, user.FirstName, user.LastName, user.Email, user.Gender)
            }
        }(i)

        go func(page int) {
            defer wg.Done()
            posts, err := scraper.ScrapePosts(page)
            if err != nil {
                fmt.Printf("Error scraping posts: %v\n", err)
                return
            }
            for _, post := range posts {
                fmt.Printf("Post: Posted by %s %s:\n%s\nLikes: %d Tags: %v\nDate posted: %s\n", post.User.FirstName, post.User.LastName, post.Text, post.Likes, post.Tags, post.PublishDate)
            }
        }(i)
    }

    wg.Wait()
}
