package service

import (
	"encoding/base64"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/dgrijalva/jwt-go"
	repository "github.com/kazakhjuice/tech-assignment-medods/internal/repo"
)

type Service struct {
	repos  *repository.Repo
	jwtKey string
}

func NewService(repo *repository.Repo, jwtKey string) *Service {
	return &Service{
		repos:  repo,
		jwtKey: jwtKey,
	}
}

func (s *Service) NewJWT(uuid string, ttl time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.StandardClaims{
		ExpiresAt: time.Now().Add(ttl).Unix(),
		Id:        uuid,
	})

	return token.SignedString([]byte(s.jwtKey))
}

func NewRefreshToken() (string, error) {
	b := make([]byte, 32)

	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)

	if _, err := r.Read(b); err != nil {
		return "", err
	}

	tokenBase64 := base64.StdEncoding.EncodeToString([]byte(b))

	return tokenBase64, nil
}

func (s *Service) UploadToken(hashedToken string, uuid string) error {
	err := s.repos.UploadToken(hashedToken, uuid)

	if err != nil {
		return err
	}

	return nil
}

func (s *Service) UpdateToken(hashedToken string, uuid string) error {
	err := s.repos.UpdateToken(hashedToken, uuid)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) GetToken(UUID string) (*repository.Token, error) {
	tokenData, err := s.repos.GetToken(UUID)
	if err != nil {
		return nil, err
	}
	return tokenData, nil
}

func (s *Service) GetUUID(JWT string) (string, error) {

	token, err := jwt.Parse(JWT, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtKey), nil
	})

	if err != nil || !token.Valid {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("cannot read jwt")
	}

	expUnix, ok := claims["exp"].(float64)
	if !ok {
		return "", errors.New("expiration time not found")
	}

	expTime := time.Unix(int64(expUnix), 0)

	if time.Now().After(expTime) {
		return "", errors.New("token has expired")
	}

	fmt.Println(expTime)

	uuid, ok := claims["jti"].(string)
	if !ok {
		return "", errors.New("id not found")
	}

	return uuid, nil

}
