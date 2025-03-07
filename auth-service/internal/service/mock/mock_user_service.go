package mock

import (
	"auth-service/internal/model"
	"errors"
)

type MockUserService struct {
	users map[string]model.User
}

func NewMockUserService() *MockUserService {
	return &MockUserService{users: make(map[string]model.User)}
}

func (m *MockUserService) RegisterUser(user model.User) (*model.User, error) {
	if _, exists := m.users[user.Email]; exists {
		return nil, errors.New("user already exists")
	}
	user.ID = len(m.users) + 1
	m.users[user.Email] = user
	return &user, nil
}

func (m *MockUserService) LoginUser(loginInfo model.Login) (*model.Tokens, error) {
	user, exists := m.users[loginInfo.Email]
	if !exists {
		return nil, errors.New("user not found")
	}
	if user.Password != loginInfo.Password {
		return nil, errors.New("wrong user password")
	}
	return &model.Tokens{
		AccessToken:  "mock_access_token",
		RefreshToken: "mock_refresh_token",
	}, nil
}
