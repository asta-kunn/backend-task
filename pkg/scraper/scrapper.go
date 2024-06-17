package scraper

import (
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "time"
)

const (
    baseURL          = "https://dummyapi.io/data/v1"
    userAgent        = "Mozilla/5.0"
    httpClientTimeout = 10 * time.Second
)

type User struct {
    Title     string `json:"title"`
    FirstName string `json:"firstName"`
    LastName  string `json:"lastName"`
    Email     string `json:"email"`
    Gender    string `json:"gender"`
}

type Post struct {
    User        User     `json:"owner"`
    Text        string   `json:"text"`
    Likes       int      `json:"likes"`
    Tags        []string `json:"tags"`
    PublishDate string   `json:"publishDate"`
}

type Comment struct {
    ID      string `json:"id"`
    Message string `json:"message"`
    User    User   `json:"owner"`
}

type Response struct {
    Data []User `json:"data"`
}

type PostResponse struct {
    Data []Post `json:"data"`
}

type CommentResponse struct {
    Data []Comment `json:"data"`
}

var client = &http.Client{Timeout: httpClientTimeout}

func ScrapeUsers(page int) ([]User, error) {
    appID := os.Getenv("APP_ID")

    req, err := http.NewRequest("GET", fmt.Sprintf("%s/user?page=%d&limit=10", baseURL, page), nil)
    if err != nil {
        return nil, err
    }
    req.Header.Set("app-id", appID)
    req.Header.Set("User-Agent", userAgent)

    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var response Response
    err = json.NewDecoder(resp.Body).Decode(&response)
    if err != nil {
        return nil, err
    }

    return response.Data, nil
}

func ScrapePosts(page int) ([]Post, error) {
    appID := os.Getenv("APP_ID")

    req, err := http.NewRequest("GET", fmt.Sprintf("%s/post?page=%d&limit=10", baseURL, page), nil)
    if err != nil {
        return nil, err
    }
    req.Header.Set("app-id", appID)
    req.Header.Set("User-Agent", userAgent)

    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var response PostResponse
    err = json.NewDecoder(resp.Body).Decode(&response)
    if err != nil {
        return nil, err
    }

    return response.Data, nil
}

func ScrapeComments(page int) ([]Comment, error) {
    appID := os.Getenv("APP_ID")

    req, err := http.NewRequest("GET", fmt.Sprintf("%s/comment?page=%d&limit=10", baseURL, page), nil)
    if err != nil {
        return nil, err
    }
    req.Header.Set("app-id", appID)
    req.Header.Set("User-Agent", userAgent)

    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var response CommentResponse
    err = json.NewDecoder(resp.Body).Decode(&response)
    if err != nil {
        return nil, err
    }

    return response.Data, nil
}
