package models

type User struct {
    ID       string `json:"id"`
    Username string `json:"username"`
    Password string `json:"-"` // "-" means this field won't be included in JSON
    Role     string `json:"role"`
}