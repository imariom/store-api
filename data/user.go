package data

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"
)

// this mutex is used to control access to usersList
// slice.
var userRWMutex = &sync.RWMutex{}

var usersList = Users{
	&User{
		ID:       0,
		Username: "testuser",
		Password: "12345",
		Name:     "Test User",
		Phone:    "000-000-000",
		Address: &Address{
			City:    "Paris",
			Street:  "Liberee",
			Number:  0,
			ZipCode: "123-654",
		},
	},
}

type Address struct {
	City    string `json:"city"`
	Street  string `json:"street"`
	Number  uint64 `json:"number"`
	ZipCode string `json:"zip_code"`
}

type User struct {
	ID       uint64 `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	*Address
}

type Users []*User

func userExists(id uint64) (int, *User, error) {
	userRWMutex.RLock()
	defer userRWMutex.RUnlock()

	for i, u := range usersList {
		if id == uint64(i) {
			tmp := &User{}
			*tmp = *u // copy current user info
			return i, tmp, nil
		}
	}

	return -1, nil, fmt.Errorf("user doe not exist")
}

func getNextUserID() uint64 {
	userRWMutex.RLock()
	defer userRWMutex.RUnlock()

	if len(usersList) == 0 {
		return 0
	}

	// assumes the user with ID 0 will never be deleted
	lastUser := usersList[len(usersList)-1]

	return lastUser.ID + 1
}

func GetAllUsers() Users {
	userRWMutex.RLock()
	tmp := make(Users, 0, len(usersList))
	tmp = append(tmp, usersList...)
	userRWMutex.RUnlock()

	return tmp
}

func GetUser(id uint64) (*User, error) {
	_, user, err := userExists(id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func AddNewUser(u *User) {
	u.ID = getNextUserID()

	userRWMutex.Lock()
	usersList = append(usersList, u)
	userRWMutex.Unlock()
}

func (us *Users) ToJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(us)
}

func (u *User) ToJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(u)
}

func (u *User) FromJSON(r io.Reader) error {
	return json.NewDecoder(r).Decode(u)
}
