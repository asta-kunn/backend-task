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
    baseURL          = "https://dummyapi.io/data/v1"
    userAgent        = "Mozilla/5.0"
    httpClientTimeout = 30 * time.Second // Increased timeout
    maxRetries       = 3                 // Number of retries
)

var client = &http.Client{Timeout: httpClientTimeout}

func ScrapeUsers(page int) ([]models.User, error) {
    appID := config.GetAppID()

    req, err := http.NewRequest("GET", fmt.Sprintf("%s/user?page=%d&limit=10", baseURL, page), nil)
    if err != nil {
        return nil, err
    }
    req.Header.Set("app-id", appID)
    req.Header.Set("User-Agent", userAgent)

    resp, err := doRequestWithRetries(req, maxRetries)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var response struct {
        Data []struct {
            ID string `json:"id"`
        } `json:"data"`
    }
    err = json.NewDecoder(resp.Body).Decode(&response)
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
            user, err := ScrapeUserDetails(userID)
            if err != nil {
                fmt.Printf("Error scraping user details for userID %s: %v\n", userID, err)
                return
            }
            mu.Lock()
            users = append(users, user)
            mu.Unlock()
        }(userData.ID)
    }

    wg.Wait()
    return users, nil
}

func ScrapeUserDetails(userID string) (models.User, error) {
    appID := config.GetAppID()

    req, err := http.NewRequest("GET", fmt.Sprintf("%s/user/%s", baseURL, userID), nil)
    if err != nil {
        return models.User{}, err
    }
    req.Header.Set("app-id", appID)
    req.Header.Set("User-Agent", userAgent)

    resp, err := doRequestWithRetries(req, maxRetries)
    if err != nil {
        return models.User{}, err
    }
    defer resp.Body.Close()

    var user models.User
    err = json.NewDecoder(resp.Body).Decode(&user)
    if err != nil {
        return models.User{}, err
    }

    return user, nil
}

func ScrapePosts(page int) ([]models.Post, error) {
    appID := config.GetAppID()

    req, err := http.NewRequest("GET", fmt.Sprintf("%s/post?page=%d&limit=10", baseURL, page), nil)
    if err != nil {
        return nil, err
    }
    req.Header.Set("app-id", appID)
    req.Header.Set("User-Agent", userAgent)

    resp, err := doRequestWithRetries(req, maxRetries)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var response models.PostResponse
    err = json.NewDecoder(resp.Body).Decode(&response)
    if err != nil {
        return nil, err
    }

    return response.Data, nil
}

func ScrapeComments(page int) ([]models.Comment, error) {
    appID := config.GetAppID()

    req, err := http.NewRequest("GET", fmt.Sprintf("%s/comment?page=%d&limit=10", baseURL, page), nil)
    if err != nil {
        return nil, err
    }
    req.Header.Set("app-id", appID)
    req.Header.Set("User-Agent", userAgent)

    resp, err := doRequestWithRetries(req, maxRetries)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var response models.CommentResponse
    err = json.NewDecoder(resp.Body).Decode(&response)
    if err != nil {
        return nil, err
    }

    return response.Data, nil
}

func doRequestWithRetries(req *http.Request, retries int) (*http.Response, error) {
    var resp *http.Response
    var err error
    for i := 0; i < retries; i++ {
        resp, err = client.Do(req)
        if err == nil {
            return resp, nil
        }
        fmt.Printf("Request failed, retrying (%d/%d): %v\n", i+1, retries, err)
        time.Sleep(time.Second * 2) // Delay between retries
    }
    return nil, err
}
