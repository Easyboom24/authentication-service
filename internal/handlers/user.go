package handlers

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"go-test/internal/apperror"
	"go-test/internal/domain"
	"go-test/internal/repository"
	"go-test/internal/services"
	"go-test/pkg/logging"
	"log"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"golang.org/x/crypto/bcrypt"
)

const (
	loginURL = "/login"
	refreshTokensURL = "/refresh-tokens"
)

type handler struct {
	userStorage repository.UserStorage
	logger *logging.Logger
}

func NewUserHandler(logger *logging.Logger, userStorage repository.UserStorage) Handler {
	return &handler{
		logger: logger,
		userStorage: userStorage,
	}
}

func (h *handler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodPost, loginURL, apperror.Middleware(h.SignIn))
	router.HandlerFunc(http.MethodPost, refreshTokensURL, apperror.Middleware(h.RefreshTokens))
}

type requestForSignIn struct {
	Guid string `json:"guid"`
	FingerPrint string `json:"fingerprint"`
}

type requestForRefreshTokens struct {
	FingerPrint string  `json:"fingerprint"`
}

type tokensResponse struct {
	AccessToken string
	RefreshToken string
}

func (h *handler) SignIn(w http.ResponseWriter, r *http.Request) error {
	var request requestForSignIn
	h.logger.Debug("Decoding request body")
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		return err
	}	

	u, err := h.userStorage.GetByGUID(r.Context(), request.Guid)
	if err != nil {
		return err
	}
	
	//генерируем случайный refresh token
	h.logger.Debug("Generation refresh token")
	generatedRefreshToken := services.CreateRefreshToken()
	
	//создаем jwt access токен
	h.logger.Debug("Creation access jwt token")
	jwtToken, err := services.CreateJwtToken(u)
	if err != nil {
		return err
	}
	
	//хэшируем refresh token
	hashedRefreshToken, err := bcrypt.GenerateFromPassword([]byte(generatedRefreshToken), bcrypt.MinCost)
	if err != nil {
		return err
	}

	//задаем сессию для записи её в бд
	session := domain.Session{
		FingerPrint: request.FingerPrint, 
		RefreshToken: string(hashedRefreshToken),
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Hour * 24 * 60),
	}
	u.Sessions = nil
	u.Sessions = append(u.Sessions, session)
	
	err = h.userStorage.CreateSession(r.Context(), u)
	if err != nil {
		return err
	}
	
	base64RefreshToken := base64.StdEncoding.EncodeToString([]byte(generatedRefreshToken))

	//создаем куку с refresh token-ом
	h.logger.Debug("Create refresh-token cookie")
	cookie := http.Cookie{
		Name: "refresh-token",
		Value: base64RefreshToken,
		Expires: u.Sessions[0].ExpiresAt,
		HttpOnly: true,
	}

	res, err := json.Marshal(tokensResponse{
		AccessToken: jwtToken,
		RefreshToken: base64RefreshToken,
	})
	if err != nil {
		return err
	}
	//задаем заголовки
	h.logger.Debug("Sending response")

	w.Header().Set("Content-Type", "aplication/json")
	w.Header().Set("Authorization", "Bearer " + jwtToken)
	//записываем куку с refresh-token-ом
	http.SetCookie(w, &cookie)
	w.Write(res)

	return nil
}


func (h *handler) RefreshTokens(w http.ResponseWriter, r *http.Request) error {
	var request requestForRefreshTokens
	h.logger.Debug("Decoding request body")
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		return err
	}	

	h.logger.Debug("Search refresh-token cookie")
	cookie, err := r.Cookie("refresh-token")
	if err != nil {
		switch {
        case errors.Is(err, http.ErrNoCookie):
            return fmt.Errorf("Cookie not found")
        default:
            log.Println(err)
            return fmt.Errorf("Server error")
        }
	}

	refreshTokenString, err := base64.StdEncoding.DecodeString(cookie.Value)
	if err != nil {
		return err
	}
	
	//ищем пользователя с конкретным отпечатком сессии
	u, err := h.userStorage.GetByFingerPrint(r.Context(), request.FingerPrint)
	if err != nil {
		fmt.Printf(err.Error())
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.Sessions[0].RefreshToken),[]byte(refreshTokenString))
	if err != nil {
		return err
	}
	
	h.logger.Debug("Create new refresh token")
	newRefreshToken := services.CreateRefreshToken()
	hashedRefreshToken, err := bcrypt.GenerateFromPassword([]byte(newRefreshToken), bcrypt.MinCost)
	if err != nil {
		return err
	}
	base64RefreshToken := base64.StdEncoding.EncodeToString([]byte(newRefreshToken))
	
	h.logger.Debug("Update refresh-token cookie")
	cookie = &http.Cookie{
		Name: "refresh-token",
		Value: base64RefreshToken,
		Expires: u.Sessions[len(u.Sessions)-1].ExpiresAt,
		HttpOnly: true,
	}

	h.logger.Debug("Create new access jwt  token")
	newJWTtoken, err := services.CreateJwtToken(u)
	if err != nil {
		return err
	}

	newSession := domain.Session{
		FingerPrint: u.Sessions[0].FingerPrint,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Hour * 24 * 60),
		RefreshToken: string(hashedRefreshToken),
	} 
	err = h.userStorage.DeleteSession(r.Context(), u)
	if err != nil {
		fmt.Println(err)
	}

	u.Sessions = nil
	u.Sessions = append(u.Sessions, newSession)
	err = h.userStorage.CreateSession(r.Context(), u)
	if err != nil {
		return err
	}
	res, err := json.Marshal(tokensResponse{
		AccessToken: newJWTtoken,
		RefreshToken: base64RefreshToken,
	})
	if err != nil {
		return err
	}
	//задаем заголовки 
	h.logger.Debug("Sending response")
	w.Header().Set("Content-Type", "aplication/json")
	w.Header().Set("Authorization", "Bearer " + newJWTtoken)
	//записываем куку с refresh-token-ом
	http.SetCookie(w, cookie)
	w.Write(res)

	return nil
}