package data

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"
)

const (
	UserNotFoundError   = "requested user does not exist"
	UserPayloadError    = "invalid user payload"
	UserIDError         = "invalid user ID"
	UserConvertionError = "failed to convert user(s)"
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

func userExists(id uint64) (index int, user *User, err error) {
	userRWMutex.RLock()
	defer userRWMutex.RUnlock()

	for i, u := range usersList {
		if id == u.ID {
			tmp := &User{}
			*tmp = *u // copy current user info
			return i, tmp, nil
		}
	}

	return -1, nil, fmt.Errorf(UserNotFoundError)
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

func UpdateUser(user *User) error {
	userRWMutex.Lock()
	defer userRWMutex.Unlock()

	for i, u := range usersList {
		if u.ID == user.ID {
			usersList[i] = user
			return nil
		}
	}

	return fmt.Errorf(UserNotFoundError)
}

func SetUser(user *User) error {
	userRWMutex.Lock()
	defer userRWMutex.Unlock()

	// TODO: optimize this loop by using the userExists() function
	// and using a better data structure.
	for i, u := range usersList {
		if u.ID == user.ID {
			if user.Username != "" {
				usersList[i].Username = user.Username
			}

			if user.Password != "" {
				usersList[i].Password = user.Password
			}

			if user.Name != "" {
				usersList[i].Name = user.Name
			}

			if user.Phone != "" {
				usersList[i].Phone = user.Phone
			}

			if user.City != "" {
				usersList[i].City = user.City
			}

			if user.Street != "" {
				usersList[i].Street = user.Street
			}

			if user.Number != 0 {
				usersList[i].Number = user.Number
			}

			if user.ZipCode != "" {
				usersList[i].ZipCode = user.ZipCode
			}

			// set temporary product equal to original product
			*user = *usersList[i]
			return nil
		}
	}

	return fmt.Errorf(UserNotFoundError)
}

func AddNewUser(u *User) {
	u.ID = getNextUserID()

	userRWMutex.Lock()
	usersList = append(usersList, u)
	userRWMutex.Unlock()
}

func RemoveUser(id uint64) (*User, error) {
	index, user, err := userExists(id)
	if err != nil {
		return nil, err
	}

	deletedUser := &User{}

	userRWMutex.Lock()
	*deletedUser = *user
	tmpList := make(Users, 0, len(usersList)-1)

	for i, u := range usersList {
		if i == index {
			continue
		}

		tmpList = append(tmpList, u)
	}
	usersList = tmpList

	userRWMutex.Unlock()

	return deletedUser, nil
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
