package handlers

import (
	"fmt"
	"log"
	"net/http"
	"regexp"

	"github.com/imariom/products-api/data"
)

var (
	listUsersRe = regexp.MustCompile(`^/users[/]?$`)
	getUserRe   = regexp.MustCompile(`^/users/(\d+)$`)
	crateUserRe = regexp.MustCompile(`^/users[/]?$`)
)

type User struct {
	logger *log.Logger
}

func NewUser(l *log.Logger) *User {
	return &User{l}
}

func (h *User) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	// set API to be json based (send and receive JSON data)
	rw.Header().Set("Content-Type", "application/json")

	switch {
	case r.Method == http.MethodGet && listUsersRe.MatchString(r.URL.Path):
		h.list(rw, r)
		return

	case r.Method == http.MethodGet && getUserRe.MatchString(r.URL.Path):
		h.get(rw, r)
		return

	case r.Method == http.MethodPost && crateUserRe.MatchString(r.URL.Path):
		h.create(rw, r)
		return

	default:
		http.Error(rw, "HTTP verb not implemented", http.StatusNotImplemented)
		return
	}
}

func (h *User) list(rw http.ResponseWriter, r *http.Request) {
	h.logger.Println("[INFO] received a GET list user request")

	users := data.GetAllUsers()
	if err := users.ToJSON(rw); err != nil {
		msg := "internal server error, while converting users to JSON"
		h.logger.Println("[ERROR] ", msg)
		http.Error(rw, msg, http.StatusInternalServerError)
	}
}

func (h *User) get(rw http.ResponseWriter, r *http.Request) {
	h.logger.Println("[INFO] received a GET user request")

	// get user id
	userID, err := getID(*getUserRe, r.URL.Path)
	if err != nil {
		http.Error(rw, "user not found", http.StatusNotFound)
		return
	}

	// get user
	user, err := data.GetUser(uint64(userID))
	if err != nil {
		http.Error(rw, "user not found", http.StatusNotFound)
		return
	}

	// try to return the user
	if err := user.ToJSON(rw); err != nil {
		http.Error(rw, InternalServerError.Error(), http.StatusInternalServerError)
	}
}

func (h *User) create(rw http.ResponseWriter, r *http.Request) {
	h.logger.Println("received a POST user request")

	newUser := &data.User{}
	if err := newUser.FromJSON(r.Body); err != nil {
		http.Error(rw, "invalid payload", http.StatusBadRequest)
		return
	}
	data.AddNewUser(newUser)

	// try to return created user
	if err := newUser.ToJSON(rw); err != nil {
		http.Error(rw,
			fmt.Sprintf("user created with ID: '%d', but failed to retrieve it",
				newUser.ID),
			http.StatusInternalServerError)
	}
}
