package scraper

import (
    "encoding/json"
    "fmt"
    "net/http"
    "sync"
    "time"

    "github.com/asta-kunn/backend-task/config"
    "github.com/asta-kunn/backend-task/models"
)

const (
    baseURL          = "https://dummyapi.io/data/v1"   // Base URL of the API being scraped
    userAgent        = "Mozilla/5.0"                   // User-Agent to mimic a real browser
    httpClientTimeout = 30 * time.Second               // Timeout for HTTP client operations
    maxRetries       = 3                               // Maximum number of retries for HTTP requests
)

var client = &http.Client{Timeout: httpClientTimeout} // HTTP client with timeout

// ScrapeUsers fetches user data from the API for a given page.
func ScrapeUsers(page int) ([]models.User, error) {
    appID := config.GetAppID() // Retrieve application ID from configuration

    req, err := http.NewRequest("GET", fmt.Sprintf("%s/user?page=%d&limit=10", baseURL, page), nil) // Create new HTTP request
    if err != nil {
        return nil, err
    }
    req.Header.Set("app-id", appID) // Set application ID header for API authentication
    req.Header.Set("User-Agent", userAgent) // Set User-Agent header

    resp, err := doRequestWithRetries(req, maxRetries) // Make the request with retries
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close() // Ensure response body is closed after handling

    var response struct {
        Data []struct {
            ID string `json:"id"`
        } `json:"data"`
    }
    err = json.NewDecoder(resp.Body).Decode(&response) // Decode JSON response into struct
    if err != nil {
        return nil, err
    }

    var users []models.User
    var wg sync.WaitGroup
    var mu sync.Mutex

    for _, userData := range response.Data {
        wg.Add(1)
        go func(userID string) {
            defer wg.Done()
            user, err := ScrapeUserDetails(userID) // Fetch detailed user information concurrently
            if err != nil {
                fmt.Printf("Error scraping user details for userID %s: %v\n", userID, err)
                return
            }
            mu.Lock()
            users = append(users, user) // Append user data in a thread-safe manner
            mu.Unlock()
        }(userData.ID)
    }

    wg.Wait()
    return users, nil
}

// ScrapeUserDetails fetches detailed information for a specific user.
func ScrapeUserDetails(userID string) (models.User, error) {
    appID := config.GetAppID() // Retrieve application ID from configuration

    req, err := http.NewRequest("GET", fmt.Sprintf("%s/user/%s", baseURL, userID), nil) // Create new HTTP request
    if err != nil {
        return models.User{}, err
    }
    req.Header.Set("app-id", appID) // Set application ID header for API authentication
    req.Header.Set("User-Agent", userAgent) // Set User-Agent header

    resp, err := doRequestWithRetries(req, maxRetries) // Make the request with retries
    if err != nil {
        return models.User{}, err
    }
    defer resp.Body.Close() // Ensure response body is closed after handling

    var user models.User
    err = json.NewDecoder(resp.Body).Decode(&user) // Decode JSON response into user struct
    if err != nil {
        return models.User{}, err
    }

    return user, nil
}

// ScrapePosts fetches post data from the API for a given page.
func ScrapePosts(page int) ([]models.Post, error) {
    appID := config.GetAppID() // Retrieve application ID from configuration

    req, err := http.NewRequest("GET", fmt.Sprintf("%s/post?page=%d&limit=10", baseURL, page), nil) // Create new HTTP request
    if err != nil {
        return nil, err
    }
    req.Header.Set("app-id", appID) // Set application ID header for API authentication
    req.Header.Set("User-Agent", userAgent) // Set User-Agent header

    resp, err := doRequestWithRetries(req, maxRetries) // Make the request with retries
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close() // Ensure response body is closed after handling

    var response models.PostResponse
    err = json.NewDecoder(resp.Body).Decode(&response) // Decode JSON response into PostResponse struct
    if err != nil {
    return nil, err
    }

    return response.Data, nil
}

// ScrapeComments fetches comment data from the API for a given page.
func ScrapeComments(page int) ([]models.Comment, error) {
    appID := config.GetAppID() // Retrieve application ID from configuration

    req, err := http.NewRequest("GET", fmt.Sprintf("%s/comment?page=%d&limit=10", baseURL, page), nil) // Create new HTTP request
    if err != nil {
        return nil, err
    }
    req.Header.Set("app-id", appID) // Set application ID header for API authentication
    req.Header.Set("User-Agent", userAgent) // Set User-Agent header

    resp, err := doRequestWithRetries(req, maxRetries) // Make the request with retries
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close() // Ensure response body is closed after handling

    var response models.CommentResponse
    err = json.NewDecoder(resp.Body).Decode(&response) // Decode JSON response into CommentResponse struct
    if err != nil {
        return nil, err
    }

    return response.Data, nil
}

// doRequestWithRetries attempts an HTTP request with retries on failure.
func doRequestWithRetries(req *http.Request, retries int) (*http.Response, error) {
    var resp *http.Response
    var err error
    for i := 0; i < retries; i++ { // Loop through the retries
        resp, err = client.Do(req) // Perform the HTTP request
        if err == nil {
            return resp, nil // Return the response if successful
        }
        fmt.Printf("Request failed, retrying (%d/%d): %v\n", i+1, retries, err) // Log retry attempts
        time.Sleep(time.Second * 2) // Delay before retrying
    }
    return nil, err // Return the last error if all retries fail
}
