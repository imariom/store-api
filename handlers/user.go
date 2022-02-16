package handlers

import (
	"fmt"
	"log"
	"net/http"
	"regexp"

	"github.com/imariom/products-api/data"
)

type User struct {
	logger *log.Logger
}

func NewUser(l *log.Logger) *User {
	return &User{l}
}

func parseUser(regex *regexp.Regexp, r *http.Request) (*data.User, error) {
	// try to parse user id
	id, err := getID(*regex, r.URL.Path)
	if err != nil {
		return nil, fmt.Errorf(data.UserIDError)
	}

	// decode the user from the request body
	user := &data.User{}
	if err := user.FromJSON(r.Body); err != nil {
		return nil, fmt.Errorf(data.UserPayloadError)
	}

	// This block avoid nil pointer reference error when none of
	// Address struct fields is provided in the payload.
	if user.Address == nil {
		user.Address = &data.Address{}
	}

	user.ID = uint64(id)
	return user, nil
}

func (h *User) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	// set API to be json based (send and receive JSON data)
	rw.Header().Set("Content-Type", "application/json")

	// route each incoming request
	switch r.Method {
	case http.MethodGet:
		h.get(rw, r)
		return

	case http.MethodPost:
		h.create(rw, r)
		return

	case http.MethodPut:
		fallthrough
	case http.MethodPatch:
		h.update(rw, r)
		return

	case http.MethodDelete:
		h.delete(rw, r)
		return

	default:
		http.Error(rw, "HTTP verb not implemented", http.StatusNotImplemented)
	}
}

func (h *User) get(rw http.ResponseWriter, r *http.Request) {
	h.logger.Println("received a GET user request")

	// serve list all users request
	listUsersRe := regexp.MustCompile(`^/users[/]?$`)

	if listUsersRe.MatchString(r.URL.Path) {
		users := data.GetAllUsers()

		if err := users.ToJSON(rw); err != nil {
			h.logger.Println(data.UserConvertionError)
			http.Error(rw, data.UserConvertionError, http.StatusInternalServerError)
		}

		return
	}

	// serve get user request
	getUserRe := regexp.MustCompile(`^/users/(\d+)$`)
	if getUserRe.MatchString(r.URL.Path) {
		userID, err := getID(*getUserRe, r.URL.Path)
		if err != nil {
			http.Error(rw, data.UserNotFoundError, http.StatusNotFound)
			return
		}

		user, err := data.GetUser(uint64(userID))
		if err != nil {
			http.Error(rw, data.UserNotFoundError, http.StatusNotFound)
			return
		}

		if err := user.ToJSON(rw); err != nil {
			http.Error(rw, data.UserConvertionError, http.StatusInternalServerError)
		}

		return
	}

	http.Error(rw, "bad GET request", http.StatusBadRequest)
}

func (h *User) create(rw http.ResponseWriter, r *http.Request) {
	h.logger.Println("received a POST user request")

	// parse user from request object
	user := &data.User{}
	if err := user.FromJSON(r.Body); err != nil {
		http.Error(rw, data.UserPayloadError, http.StatusBadRequest)
		return
	}

	// add user to data store
	data.AddNewUser(user)

	// try to return created user
	if err := user.ToJSON(rw); err != nil {
		http.Error(rw,
			fmt.Sprintf("user created with ID: '%d', but failed to retrieve it",
				user.ID),
			http.StatusInternalServerError)
	}
}

func (h *User) update(rw http.ResponseWriter, r *http.Request) {
	h.logger.Println("received a PUT user request")

	// try to parse user from request object
	updateUserRe := regexp.MustCompile(`^/users/(\d+)$`)

	user, err := parseUser(updateUserRe, r)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusNotFound)
		return
	}

	// match request method (PUT or PATCH)
	if r.Method == http.MethodPut {
		// update whole user information
		if err := data.UpdateUser(user); err != nil {
			http.Error(rw, err.Error(), http.StatusNotFound)
			return
		}
	} else if r.Method == http.MethodPatch {
		// update user attributes
		if err := data.SetUser(user); err != nil {
			http.Error(rw, err.Error(), http.StatusNotFound)
			return
		}
	}

	// return updated user
	if err := user.ToJSON(rw); err != nil {
		http.Error(rw,
			fmt.Sprintf("user with ID: '%d' was updated sucessfully, but failed to retrieve it",
				user.ID),
			http.StatusInternalServerError)
	}
}

func (h *User) delete(rw http.ResponseWriter, r *http.Request) {
	h.logger.Println("received a DELETE user request")

	// get user id
	deleteUserRe := regexp.MustCompile(`^/users/(\d+)$`)
	userID, err := getID(*deleteUserRe, r.URL.Path)
	if err != nil {
		http.Error(rw, data.UserIDError, http.StatusNotFound)
		return
	}

	// delete user from datastore
	user, err := data.RemoveUser(uint64(userID))
	if err != nil {
		http.Error(rw, data.UserNotFoundError, http.StatusNotFound)
		return
	}

	// return deleted user
	if err := user.ToJSON(rw); err != nil {
		http.Error(rw,
			fmt.Sprintf("user with ID: '%d' was deleted, but failed to retrieve it",
				user.ID),
			http.StatusInternalServerError)
	}
}
