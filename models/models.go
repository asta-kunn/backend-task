package models

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
