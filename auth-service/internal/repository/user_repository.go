package repository

import (
	"auth-service/internal/logger"
	"auth-service/internal/model"
	"database/sql"
	"errors"
	"time"
)

//type UserRepositoryInterface interface {
//	GetUserByEmail(email string) (*model.User, error)
//	InsertUser(user model.User) (int, error)
//	InsertRefreshToken(user *model.User, token string) error
//	GetRefreshToken(token string) (*model.User, error)
//	DeleteRefreshToken(token string) error
//}

type UserRepository struct {
	db   *sql.DB
	logs *logger.Logger
}

func NewUserRepository(db *sql.DB, logs *logger.Logger) *UserRepository {
	return &UserRepository{db: db, logs: logs}
}

func (r *UserRepository) GetUserByEmail(email string) (*model.User, error) {
	var user model.User

	query := `SELECT id, name, email, password, created_at, updated_at  FROM users WHERE email = $1`
	err := r.db.QueryRow(query, email).Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		r.logs.Error.Printf("Database error in GetUserEmail: %v", err)
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetUserByID(id int) (*model.User, error) {
	var user model.User

	query := `SELECT id, name, email, password, created_at, updated_at  FROM users WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		r.logs.Error.Printf("Database error in GetUserEmail: %v", err)
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) InsertUser(user model.User) (int, error) {
	query := `INSERT INTO users (name, email, password, created_at, updated_at) VALUES ($1,$2,$3,$4,$4) RETURNING id`

	err := r.db.QueryRow(query, user.Name, user.Email, user.Password, user.CreatedAt).Scan(&user.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logs.Error.Printf("Database error in InsertUser: %v", err)
			return -1, errors.New("database error: failed to insert user")
		}
		return -1, errors.New("database error: failed to insert user")
	}
	r.logs.Info.Printf("User inserted successfully: ID=%d, Email=%s", user.ID, user.Email)
	return user.ID, nil
}

func (r *UserRepository) InsertRefreshToken(user *model.User, token string) error {
	query := `INSERT INTO refresh_tokens (user_id, token, expires_at, created_at) VALUES ($1, $2, $3, $4);`

	_, err := r.db.Exec(query, user.ID, token, time.Now().Add(7*24*time.Hour), time.Now())
	if err != nil {
		r.logs.Error.Printf("Database error in InsertRefreshToken: %v", err)
		return errors.New("database error: failed to insert refresh token")
	}
	r.logs.Info.Printf("Refresh token inserted for user ID=%d", user.ID)
	return nil
}

func (r *UserRepository) GetRefreshToken(token string) (*model.User, error) {
	var user model.User
	query := `SELECT u.id, u.name, u.email FROM users u INNER JOIN refresh_tokens r ON u.id = r.user_id WHERE r.token = $1 AND r.expires_at > NOW()`

	err := r.db.QueryRow(query, token).Scan(&user.ID, &user.Name, &user.Email)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		r.logs.Error.Printf("Database error in GetRefreshToken: %v", err)
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) DeleteRefreshToken(token string) error {
	query := `DELETE FROM refresh_tokens WHERE token = $1`
	_, err := r.db.Exec(query, token)

	if err != nil {
		r.logs.Error.Printf("Database error in DeleteRefreshToken: %v", err)
		return errors.New("database error: failed to delete refresh token")
	}
	return nil
}

func (r *UserRepository) UpdateUser(userID int, newUserName string, newUserEmail string) (*model.UserInfo, error) {
	query := `UPDATE user SET name $1, email $2, updated_at $3 WHERE id $4 RETURNING id, name, email, created_at, updated_at`
	var userInfo model.UserInfo
	err := r.db.QueryRow(query, newUserName, newUserEmail, time.Now(), userID).Scan(&userInfo)
	if err != nil {
		//r.logs.Error.Printf("Database error in UpdateUser: %v", err)
		return nil, err
	}
	return &userInfo, nil
}

func (r *UserRepository) DeleteCurrentUser(userID int) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.Exec(query, userID)
	if err != nil {
		r.logs.Error.Printf("Database error in DeleteUser: %v", err)
		return errors.New("database error: failed to delete user")
	}
	return nil
}
