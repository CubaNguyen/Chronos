package sample

import (
    "database/sql"
    "fmt"
    mycache "github.com/myorg/cache"
)

type UserService struct {
    db    *sql.DB
    cache *mycache.Cache
    name  string
}

type Reader interface {
    Read(p []byte) (int, error)
}

func Add(a, b int) int {
    return a + b
}

func (s *UserService) GetUser(id int) (*User, error) {
    row := s.db.Query("SELECT * FROM users WHERE id = ?", id)
    fmt.Println("querying user", id)
    return nil, nil
}

func (s *UserService) helper() {
    s.GetUser(1)
}