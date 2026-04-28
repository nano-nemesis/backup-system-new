package web

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	RoleAdmin  = "admin"
	RoleViewer = "viewer"
)

type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"password_hash"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

type UserStore struct {
	path string
	mu   sync.RWMutex
	data []User
}

func NewUserStore(path string) (*UserStore, error) {
	s := &UserStore{path: path}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *UserStore) load() error {
	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		s.data = []User{}
		return nil
	}
	if err != nil {
		return fmt.Errorf("read users file: %w", err)
	}
	return json.Unmarshal(data, &s.data)
}

func (s *UserStore) save() error {
	data, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o600)
}

func (s *UserStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.data)
}

func (s *UserStore) All() []User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]User, len(s.data))
	copy(out, s.data)
	return out
}

func (s *UserStore) GetByUsername(username string) (User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, u := range s.data {
		if u.Username == username {
			return u, true
		}
	}
	return User{}, false
}

func (s *UserStore) GetByID(id string) (User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, u := range s.data {
		if u.ID == id {
			return u, true
		}
	}
	return User{}, false
}

func (s *UserStore) Create(username, password, role string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, u := range s.data {
		if u.Username == username {
			return fmt.Errorf("username already exists")
		}
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	s.data = append(s.data, User{
		ID:           fmt.Sprintf("%d", time.Now().UnixNano()),
		Username:     username,
		PasswordHash: string(hash),
		Role:         role,
		CreatedAt:    time.Now().UTC(),
	})
	return s.save()
}

func (s *UserStore) UpdateRole(id, role string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, u := range s.data {
		if u.ID == id {
			s.data[i].Role = role
			return s.save()
		}
	}
	return fmt.Errorf("user not found")
}

func (s *UserStore) UpdatePassword(id, password string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	for i, u := range s.data {
		if u.ID == id {
			s.data[i].PasswordHash = string(hash)
			return s.save()
		}
	}
	return fmt.Errorf("user not found")
}

func (s *UserStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, u := range s.data {
		if u.ID == id {
			s.data = append(s.data[:i], s.data[i+1:]...)
			return s.save()
		}
	}
	return fmt.Errorf("user not found")
}

func (s *UserStore) CheckPassword(u User, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) == nil
}
