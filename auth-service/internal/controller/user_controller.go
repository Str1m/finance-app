package controller

import (
	"auth-service/internal/logger"
	"auth-service/internal/model"
	"auth-service/internal/service"
	"encoding/json"
	"net/http"
)

type UserController struct {
	userService *service.UserService
	logs        *logger.Logger
}

func NewUserHandler(userService *service.UserService, logs *logger.Logger) *UserController {
	return &UserController{
		userService: userService,
		logs:        logs,
	}
}

func (c *UserController) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var user model.User

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		c.logs.Error.Printf("Failed to decode JSON: %v", err)
		SendErrorResponse(w, http.StatusBadRequest, "Invalid Request Format")
		return
	}

	createdUser, err := c.userService.RegisterUser(user)
	if err != nil {
		if err.Error() == "user already exists" {
			c.logs.Info.Printf("Attempt to register existing user: %s", user.Email)
			SendErrorResponse(w, http.StatusConflict, "User with this email already exists")
			return
		}
		c.logs.Error.Printf("Registration error: %v", err)
		SendErrorResponse(w, http.StatusInternalServerError, "Registration failed, please try again later")
		return
	}
	c.logs.Info.Printf("User registered successfully: ID=%d, Email=%s", createdUser.ID, createdUser.Email)
	SendSuccessResponse(w, http.StatusCreated, model.RegisterResponse{
		ID:    createdUser.ID,
		Name:  createdUser.Name,
		Email: createdUser.Email,
	})
}

func (c *UserController) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var loginInfo model.Login

	err := json.NewDecoder(r.Body).Decode(&loginInfo)
	if err != nil {
		c.logs.Error.Printf("Failed to decode JSON: %v", err)
		SendErrorResponse(w, http.StatusBadRequest, "Invalid Request Format")
		return
	}

	tokens, err := c.userService.LoginUser(loginInfo)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorMessage := "Login failed, please try again later"

		if err.Error() == "user not found" || err.Error() == "wrong user password" {
			statusCode = http.StatusUnauthorized
			errorMessage = "Invalid email or password"
		}

		SendErrorResponse(w, statusCode, errorMessage)
		c.logs.Error.Printf("Error in user login: %v", err)
		return
	}

	SendSuccessResponse(w, http.StatusOK, tokens)
}

func (c *UserController) RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		RefreshToken string `json:"refresh_token"`
	}

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		c.logs.Error.Printf("Failed to decode JSON: %v", err)
		SendErrorResponse(w, http.StatusBadRequest, "Invalid Request Format")
		return
	}

	tokens, err := c.userService.RefreshAccessToken(request.RefreshToken)
	if err != nil {
		statusCode := http.StatusUnauthorized
		errorMessage := "Invalid refresh token"

		if err.Error() == "refresh token not found" {
			statusCode = http.StatusUnauthorized
		} else if err.Error() == "database error" {
			statusCode = http.StatusInternalServerError
			errorMessage = "Something went wrong, please try again later"
		}

		SendErrorResponse(w, statusCode, errorMessage)
		c.logs.Error.Printf("Error refreshing token: %v", err)
		return
	}

	SendSuccessResponse(w, http.StatusOK, tokens)

}

func (c *UserController) GetCurrentUserHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		SendErrorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	user, err := c.userService.GetUserByID(userID)

	if err != nil {
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve user")
		return
	}

	userInfo := model.UserInfo{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
	SendSuccessResponse(w, http.StatusOK, userInfo)
}

func (c *UserController) UpdateCurrentUserHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		SendErrorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var updateData struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	err := json.NewDecoder(r.Body).Decode(&updateData)
	if err != nil {
		c.logs.Error.Printf("Failed to decode JSON: %v", err)
		SendErrorResponse(w, http.StatusBadRequest, "Invalid Request Format")
		return
	}

	updateUserInfo, err := c.userService.UpdateCurrentUser(userID, updateData.Name, updateData.Email)
	if err != nil {
		if err.Error() == "email already in use" {
			SendErrorResponse(w, http.StatusConflict, "Email already in use")
		} else {
			SendErrorResponse(w, http.StatusInternalServerError, "Failed to update profile")
		}
		c.logs.Error.Printf("Error updating user: %v", err)
		return
	}

	SendSuccessResponse(w, http.StatusOK, updateUserInfo)
}

func (c *UserController) DeleteCurrentUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		SendErrorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}
	err := c.userService.DeleteCurrentUser(userID)
	if err != nil {
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to delete account")
		c.logs.Error.Printf("Error deleting user: %v", err)
		return
	}
	SendSuccessResponse(w, http.StatusNoContent, nil)
}
