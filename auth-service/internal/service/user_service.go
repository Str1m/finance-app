package service

import (
	"auth-service/internal/logger"
	"auth-service/internal/model"
	"auth-service/internal/repository"
	"errors"
	"time"
)

//	type UserServiceInterface interface {
//		RegisterUser(user model.User) (*model.User, error)
//		LoginUser(loginInfo model.Login) (*model.Tokens, error)
//	}
type UserService struct {
	repo       *repository.UserRepository
	logs       *logger.Logger
	jwtService *JWTService
}

func NewUserService(repo *repository.UserRepository, logs *logger.Logger, jwtService *JWTService) *UserService {
	return &UserService{
		repo:       repo,
		logs:       logs,
		jwtService: jwtService,
	}
}

func (s *UserService) RegisterUser(user model.User) (*model.User, error) {
	existingUser, err := s.repo.GetUserByEmail(user.Email)
	if err != nil {
		s.logs.Error.Printf("Database error: %v", err)
		return nil, errors.New("database error")
	}

	if existingUser != nil {
		s.logs.Info.Printf("User with email %s already exists", user.Email)
		return nil, errors.New("user already exists")
	}

	hashedPassword, err := HashPassword(user.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}
	user.Password = hashedPassword
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	id, err := s.repo.InsertUser(user)
	if err != nil {
		s.logs.Error.Printf("Database error: could not create user: %v", err)
		return nil, errors.New("database error: could not create user")
	}
	user.ID = id
	s.logs.Info.Printf("User registered successfully: ID=%d, Email=%s", user.ID, user.Email)
	return &user, nil
}

func (s *UserService) LoginUser(loginInfo model.Login) (*model.Tokens, error) {
	existingUser, err := s.repo.GetUserByEmail(loginInfo.Email)
	if err != nil {
		return nil, errors.New("database error")
	}

	if existingUser == nil {
		s.logs.Info.Printf("Failed login attempt: email not found (%s)", loginInfo.Email)
		return nil, errors.New("user not found")
	}

	if !CheckPasswordHash(loginInfo.Password, existingUser.Password) {
		return nil, errors.New("wrong user password")
	}

	accessToken, err := s.jwtService.GenerateAccessToken(existingUser.ID)
	if err != nil {
		return nil, errors.New("error in access token generation")
	}

	refreshToken := GenerateRefreshToken()

	if err := s.repo.InsertRefreshToken(existingUser, refreshToken); err != nil {
		return nil, errors.New("database error: could not insert token")
	}

	s.logs.Info.Printf("User logged in: ID=%d, Email=%s", existingUser.ID, existingUser.Email)

	return &model.Tokens{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (s *UserService) RefreshAccessToken(refreshToken string) (*model.Tokens, error) {
	user, err := s.repo.GetRefreshToken(refreshToken)
	if err != nil {
		return nil, errors.New("database error")
	}

	if user == nil {
		s.logs.Error.Printf("Invalid refresh token: %s", refreshToken)
		return nil, errors.New("refresh token not found")
	}

	accessToken, err := s.jwtService.GenerateAccessToken(user.ID)
	if err != nil {
		return nil, errors.New("error in access token generation")
	}

	newRefreshToken := GenerateRefreshToken()
	err = s.repo.InsertRefreshToken(user, newRefreshToken)
	if err != nil {
		return nil, errors.New("database error: could not insert refresh token")
	}

	err = s.repo.DeleteRefreshToken(refreshToken)
	if err != nil {
		s.logs.Error.Printf("Failed to delete old refresh token: %v", err)
	}

	return &model.Tokens{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (s *UserService) GetUserByID(userID int) (*model.User, error) {
	return s.repo.GetUserByID(userID)
}

func (s *UserService) UpdateCurrentUser(userID int, userName string, userEmail string) (*model.UserInfo, error) {
	existingUser, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, errors.New("database error")
	}

	if existingUser != nil && existingUser.ID != userID {
		return nil, errors.New("email already in use")
	}

	updateUserInfo, err := s.repo.UpdateUser(userID, userName, userEmail)
	if err != nil {
		return nil, errors.New("database error: could not update user")
	}
	return updateUserInfo, nil
}

func (s *UserService) DeleteCurrentUser(userID int) error {
	err := s.repo.DeleteCurrentUser(userID)
	if err != nil {
		return errors.New("database error: could not delete user")
	}
	return nil
}
