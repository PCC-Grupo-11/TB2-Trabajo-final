package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/logger"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/storage"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateUser = errors.New("username already exists")
	ErrInvalidCreds  = errors.New("invalid credentials")
)

type Auth struct {
	store      *storage.Repository
	jwtSecret  []byte
	expiration time.Duration
}

func New(store *storage.Repository, jwtSecret string, expiration time.Duration) *Auth {
	return &Auth{
		store:      store,
		jwtSecret:  []byte(jwtSecret),
		expiration: expiration,
	}
}

func (a *Auth) Register(ctx context.Context, username, password string) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	err = a.store.CreateUser(ctx, username, string(hashed))
	if err != nil {
		logger.Warn("register failed", "username", username, "error", err)
		if mongo.IsDuplicateKeyError(err) {
			return ErrDuplicateUser
		}
		return fmt.Errorf("register: %w", err)
	}
	return nil
}

func (a *Auth) Login(ctx context.Context, username, password string) (string, error) {
	user, err := a.store.FindUser(ctx, username)
	if err != nil {
		logger.Warn("login find user failed", "username", username, "error", err)
		if errors.Is(err, mongo.ErrNoDocuments) {
			return "", ErrInvalidCreds
		}
		return "", fmt.Errorf("login: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password))
	if err != nil {
		logger.Warn("login wrong password", "username", username)
		return "", ErrInvalidCreds
	}

	return a.generateToken(user.ID.Hex())
}

func (a *Auth) generateToken(userID string) (string, error) {
	now := time.Now()

	claims := jwt.MapClaims{
		"sub": userID,
		"iat": now.Unix(),
		"exp": now.Add(a.expiration).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(a.jwtSecret)
}
