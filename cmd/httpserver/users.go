package main

import (
	"encoding/json"
	"net/http"
)

type User struct {
	firstName string
	lastName  string
}
type UserStorage struct {
	users []User
}

func NewUserStorage() *UserStorage {
	return &UserStorage{users: make([]User, 10)}
}

func (s *UserStorage) listUsers() []User {
	return s.users
}

func handleGetUsers(store *UserStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		users := store.listUsers()
		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(users); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
