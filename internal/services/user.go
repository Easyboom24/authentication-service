package services

import (
	"fmt"
	"go-test/internal/config"
	"go-test/internal/domain"
	"math/rand"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func CreateJwtToken(user domain.User) (string, error) {
	payload := jwt.MapClaims{
		"sub": user.GUID,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
		"iat": time.Now().Unix(),
	}
	
	jwt := jwt.NewWithClaims(jwt.SigningMethodHS512, payload)
	jwtToken, err := jwt.SignedString([]byte(config.GetConfig().Jwt.Secret_key))
	if err != nil {
		return  "", err
	}
	return jwtToken, nil
}

func CreateRefreshToken() string {
	token := make([]byte, 32)
	_, err := rand.New(rand.NewSource(time.Now().Unix())).Read(token)
	if err != nil {
		fmt.Print(err)
	}
	
	return fmt.Sprintf("%x",token)
}
