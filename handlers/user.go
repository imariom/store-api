package handlers

import (
	"fmt"
	"log"
	"net/http"
)

type User struct {
	logger *log.Logger
}

func NewUser(l *log.Logger) *User {
	return &User{l}
}

func (h *User) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(rw, "Hello from User Handler\n")
}
