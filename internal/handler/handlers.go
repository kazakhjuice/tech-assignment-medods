package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/kazakhjuice/tech-assignment-medods/internal/service"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	UUID string `json:"uuid"`
}

type Tokens struct {
	JWT          string `json:"jwt"`
	RefreshToken string `json:"refreshToken"`
}

type Handler struct {
	service service.Service
}

func NewHandler(service *service.Service) *Handler {
	return &Handler{
		service: *service,
	}
}

func (h *Handler) GetTokens(w http.ResponseWriter, r *http.Request) {

	var user User

	err := json.NewDecoder(r.Body).Decode(&user)

	if err != nil {
		http.Error(w, "wrong json format", http.StatusBadRequest)
		log.Print(err)
		return
	}

	jwt, err := h.service.NewJWT(user.UUID, time.Minute*15)

	if err != nil {
		http.Error(w, "fail to get jwt", http.StatusInternalServerError)
		log.Print(err)
		return
	}

	refreshToken, err := service.NewRefreshToken()

	if err != nil {
		http.Error(w, "fail to get refresh token", http.StatusInternalServerError)
		log.Print(err)
		return
	}

	hashedToken, err := bcrypt.GenerateFromPassword([]byte(refreshToken), bcrypt.DefaultCost)

	if err != nil {
		http.Error(w, "cannot encrypt pass", http.StatusInternalServerError)
		log.Print(err)
		return
	}

	err = h.service.UploadToken(string(hashedToken), user.UUID)

	if err != nil {
		http.Error(w, "already in base", http.StatusBadRequest)
		log.Print(err)
		return
	}

	tokens := Tokens{
		JWT:          jwt,
		RefreshToken: refreshToken,
	}

	jsonData, err := json.Marshal(tokens)
	if err != nil {
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	w.Write(jsonData)

}

func (h *Handler) UpdateRefreshToken(w http.ResponseWriter, r *http.Request) {
	var tokens Tokens
	if err := json.NewDecoder(r.Body).Decode(&tokens); err != nil {
		http.Error(w, "failed to parse refresh token", http.StatusBadRequest)
		log.Print(err)
		return
	}

	//должен ли я проверять expire у jwt? я провреил на всякий

	UUID, err := h.service.GetUUID(tokens.JWT)

	if err != nil {
		http.Error(w, "failed to decypher token", http.StatusBadRequest)
		log.Print(err)
		return
	}

	tokenData, err := h.service.GetToken(UUID)
	if err != nil {
		http.Error(w, "expired or bad token", http.StatusInternalServerError)
		log.Print(err)
		return
	}

	if tokenData == nil || tokenData.ExpiresAt.Before(time.Now()) {
		http.Error(w, "refresh token not found or expired", http.StatusUnauthorized)
		log.Print(err)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(tokenData.Token), []byte(tokens.RefreshToken))

	if err != nil {
		http.Error(w, "bad refresh token", http.StatusBadRequest)
		log.Print(err)
		return
	}

	newJWT, err := h.service.NewJWT(tokenData.UUID, time.Minute*15)
	if err != nil {
		http.Error(w, "failed to generate new JWT", http.StatusInternalServerError)
		log.Print(err)
		return
	}

	newRefreshToken, err := service.NewRefreshToken()
	if err != nil {
		http.Error(w, "failed to generate new refresh token", http.StatusInternalServerError)
		log.Print(err)
		return
	}

	hashedToken, err := bcrypt.GenerateFromPassword([]byte(newRefreshToken), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "failed to encrypt new refresh token", http.StatusInternalServerError)
		log.Print(err)
		return
	}

	if err := h.service.UpdateToken(string(hashedToken), tokenData.UUID); err != nil {
		http.Error(w, "failed to update token", http.StatusInternalServerError)
		log.Print(err)
		return
	}

	tokensUpdated := Tokens{
		JWT:          newJWT,
		RefreshToken: newRefreshToken,
	}

	jsonData, err := json.Marshal(tokensUpdated)
	if err != nil {
		http.Error(w, "failed to encode JSON", http.StatusInternalServerError)
		return
	}

	log.Print("resfreshed tokens for", tokenData.UUID)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}
