package domain

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/mashfeii/chirpy/internal/infrastructure/api"
	"github.com/mashfeii/chirpy/internal/infrastructure/database"
	"github.com/mashfeii/chirpy/pkg/auth"
	stringshelpers "github.com/mashfeii/chirpy/pkg/strings_helpers"
	"github.com/samber/lo"
)

const (
	UpgradeEvent = "user.upgraded"
)

type APIConfig struct {
	FileserverHits atomic.Int32
	Database       *database.Queries
	Platform       string
	Secret         string
	Polka          string
}

func errorRespond(w http.ResponseWriter, code int, message string) {
	if respondErr := api.RespondWithError(w, code, message); respondErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("unable to respond: %s", respondErr.Error())
	}
}

func successRespond(w http.ResponseWriter, code int, data interface{}) {
	if respondErr := api.RespondWithJSON(w, code, data); respondErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("unable to respond: %s", respondErr.Error())
	}
}

func (conf *APIConfig) MiddlewareInc(next http.Handler) http.Handler {
	// BUG: incrementing will run ones on first function call
	// conf.FileserverHits.Add(1)
	// return next
	// NOTE: instead, we need to return a new function
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conf.FileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (conf *APIConfig) HitsHandler(r http.ResponseWriter, _ *http.Request) {
	r.Header().Add("Content-Type", "text/html")

	answer := fmt.Sprintf(`
    <html>
      <body>
        <h1>Welcome, Chirpy Admin</h1>
        <p>Chirpy has been visited %d times!</p>
      </body>
    </html>`,
		conf.FileserverHits.Load())

	_, err := r.Write([]byte(answer))
	if err != nil {
		panic(err)
	}
}

func (conf *APIConfig) ResetHandler(w http.ResponseWriter, r *http.Request) {
	if conf.Platform != "dev" {
		w.WriteHeader(403)
		return
	}

	prevValue := conf.FileserverHits.Load()
	conf.FileserverHits.Store(0)

	err := conf.Database.DeleteUsers(r.Context())
	if err != nil {
		errorRespond(w, http.StatusInternalServerError, err.Error())

		conf.FileserverHits.Store(prevValue)

		return
	}
}

func (conf *APIConfig) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)

	var params parameters

	err := decoder.Decode(&params)
	if err != nil {
		errorRespond(w, http.StatusInternalServerError, err.Error())

		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		errorRespond(w, http.StatusInternalServerError, err.Error())

		return
	}

	user, err := conf.Database.CreateUser(r.Context(), database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		errorRespond(w, http.StatusInternalServerError, err.Error())

		return
	}

	successRespond(w, http.StatusCreated, User{
		ID:          user.ID,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed,
	})
}

func (conf *APIConfig) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetAuthorizationToken(r.Header, "Bearer")
	if err != nil {
		errorRespond(w, http.StatusUnauthorized, err.Error())
		return
	}

	userID, err := auth.ValidateJWT(token, conf.Secret)
	if err != nil {
		errorRespond(w, http.StatusUnauthorized, err.Error())
		return
	}

	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)

	var params parameters

	err = decoder.Decode(&params)
	if err != nil {
		errorRespond(w, http.StatusInternalServerError, err.Error())
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		errorRespond(w, http.StatusInternalServerError, err.Error())
		return
	}

	newUser, err := conf.Database.UpdateUser(r.Context(), database.UpdateUserParams{
		ID:             userID,
		Email:          params.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		errorRespond(w, http.StatusInternalServerError, err.Error())
		return
	}

	successRespond(w, http.StatusOK, User{
		ID:          newUser.ID,
		CreatedAt:   newUser.CreatedAt,
		UpdatedAt:   newUser.UpdatedAt,
		Email:       newUser.Email,
		IsChirpyRed: newUser.IsChirpyRed,
	})
}

func (conf *APIConfig) LoginUserHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)

	var params parameters

	err := decoder.Decode(&params)
	if err != nil {
		errorRespond(w, http.StatusInternalServerError, err.Error())

		return
	}

	user, err := conf.Database.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		errorRespond(w, http.StatusInternalServerError, err.Error())

		return
	}

	if err := auth.CheckPasswordHash(params.Password, user.HashedPassword); err != nil {
		errorRespond(w, http.StatusUnauthorized, "incorrect email or password")

		return
	}

	token, err := auth.MakeJWT(user.ID, conf.Secret, time.Hour)
	if err != nil {
		errorRespond(w, http.StatusUnauthorized, err.Error())

		return
	}

	refreshToken := auth.MakeRefreshToken()

	_, err = conf.Database.InsertRefreshToken(r.Context(), database.InsertRefreshTokenParams{
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 60),
		RevokedAt: sql.NullTime{
			Time:  time.Time{},
			Valid: false,
		},
		UserID: user.ID,
	})
	if err != nil {
		errorRespond(w, http.StatusInternalServerError, err.Error())
	}

	successRespond(w, http.StatusOK, User{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        token,
		RefreshToken: refreshToken,
		IsChirpyRed:  user.IsChirpyRed,
	})
}

func (conf *APIConfig) CreateChirpsHandler(w http.ResponseWriter, r *http.Request) {
	type parameter struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)

	var params parameter

	err := decoder.Decode(&params)
	if err != nil {
		errorRespond(w, http.StatusInternalServerError, err.Error())
		return
	}

	token, err := auth.GetAuthorizationToken(r.Header, "Bearer")
	if err != nil {
		errorRespond(w, http.StatusUnauthorized, err.Error())
		return
	}

	userID, err := auth.ValidateJWT(token, conf.Secret)
	if err != nil {
		errorRespond(w, http.StatusUnauthorized, err.Error())
		return
	}

	if len(params.Body) > 140 {
		errorRespond(w, 400, "Chirp is too long")
		return
	}

	cleanedBody := stringshelpers.CleanString(params.Body, []string{
		"kerfuffle", "sharbert", "fornax",
	})

	chirp, err := conf.Database.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleanedBody,
		UserID: userID,
	})
	if err != nil {
		errorRespond(w, http.StatusInternalServerError, err.Error())
		return
	}

	successRespond(w, http.StatusCreated, Chirp(chirp))
}

func (conf *APIConfig) ShowChirpsHandler(w http.ResponseWriter, r *http.Request) {
	sortOrder := r.URL.Query().Get("sort")

	authorID, err := uuid.Parse(r.URL.Query().Get("author_id"))
	if err != nil && !uuid.IsInvalidLengthError(err) {
		errorRespond(w, http.StatusBadRequest, err.Error())
		return
	}

	var chirps []database.Chirp

	switch authorID {
	case uuid.UUID{}:
		chirps, err = conf.Database.GetChirps(r.Context())
	default:
		chirps, err = conf.Database.GetChirpsByUser(r.Context(), authorID)
	}

	if err != nil {
		errorRespond(w, http.StatusInternalServerError, err.Error())
		return
	}

	convertedChirps := lo.Map(chirps, func(chirp database.Chirp, _ int) Chirp {
		return Chirp(chirp)
	})

	if sortOrder == "desc" {
		sort.Slice(convertedChirps, func(i, j int) bool {
			return convertedChirps[i].CreatedAt.After(convertedChirps[j].CreatedAt)
		})
	}

	successRespond(w, http.StatusOK, convertedChirps)
}

func (conf *APIConfig) ShowChirpHandler(w http.ResponseWriter, r *http.Request) {
	pattern, err := uuid.Parse(r.PathValue("chirp_id"))
	if err != nil {
		errorRespond(w, http.StatusInternalServerError, err.Error())
		return
	}

	chirp, err := conf.Database.GetChirp(r.Context(), pattern)
	if err != nil {
		errorRespond(w, http.StatusNotFound, err.Error())
		return
	}

	successRespond(w, http.StatusOK, Chirp(chirp))
}

func (conf *APIConfig) DeleteChirpHandler(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetAuthorizationToken(r.Header, "Bearer")
	if err != nil {
		errorRespond(w, http.StatusUnauthorized, err.Error())
		return
	}

	userID, err := auth.ValidateJWT(token, conf.Secret)
	if err != nil {
		errorRespond(w, http.StatusUnauthorized, err.Error())
		return
	}

	chirpID, err := uuid.Parse(r.PathValue("chirp_id"))
	if err != nil {
		errorRespond(w, http.StatusInternalServerError, err.Error())
		return
	}

	chirp, err := conf.Database.GetChirp(r.Context(), chirpID)
	if err != nil {
		errorRespond(w, http.StatusNotFound, err.Error())
		return
	}

	if chirp.UserID != userID {
		errorRespond(w, http.StatusForbidden, "user does not own chirp")
		return
	}

	if err = conf.Database.DeleteChirp(r.Context(), chirpID); err != nil {
		errorRespond(w, http.StatusInternalServerError, err.Error())
		return
	}

	successRespond(w, http.StatusNoContent, nil)
}

func (conf *APIConfig) RefreshHandler(w http.ResponseWriter, r *http.Request) {
	type returnValue struct {
		Token string `json:"token"`
	}

	token, err := auth.GetAuthorizationToken(r.Header, "Bearer")
	if err != nil {
		errorRespond(w, http.StatusUnauthorized, err.Error())
		return
	}

	DBToken, err := conf.Database.GetRefreshToken(r.Context(), token)
	if err != nil {
		errorRespond(w, http.StatusUnauthorized, err.Error())
		return
	} else if DBToken.RevokedAt.Valid {
		errorRespond(w, http.StatusUnauthorized, "refresh token has been revoked")
		return
	}

	refreshedToken, err := auth.MakeJWT(DBToken.UserID, conf.Secret, time.Hour)
	if err != nil {
		errorRespond(w, http.StatusInternalServerError, err.Error())
		return
	}

	successRespond(w, http.StatusOK, returnValue{
		Token: refreshedToken,
	})
}

func (conf *APIConfig) RevokeHandler(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetAuthorizationToken(r.Header, "Bearer")
	if err != nil {
		errorRespond(w, http.StatusUnauthorized, err.Error())
		return
	}

	if err = conf.Database.RevokeRefreshToken(r.Context(), token); err != nil {
		errorRespond(w, http.StatusInternalServerError, err.Error())
		return
	}

	successRespond(w, http.StatusNoContent, nil)
}

func (conf *APIConfig) PolkaWebhookHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Event string `json:"event"`
		Data  struct {
			UserID string `json:"user_id"`
		} `json:"data"`
	}

	token, err := auth.GetAuthorizationToken(r.Header, "ApiKey")
	if err != nil || token != conf.Polka {
		errorRespond(w, http.StatusUnauthorized, err.Error())
		return
	}

	decoder := json.NewDecoder(r.Body)

	var params parameters

	err = decoder.Decode(&params)
	if err != nil {
		errorRespond(w, http.StatusInternalServerError, err.Error())
		return
	}

	if params.Event != UpgradeEvent {
		errorRespond(w, http.StatusNoContent, "")
		return
	}

	userID, err := uuid.Parse(params.Data.UserID)
	if err != nil {
		errorRespond(w, http.StatusNoContent, "")
		return
	}

	if _, err = conf.Database.GetUserByID(r.Context(), userID); err != nil {
		errorRespond(w, http.StatusNotFound, err.Error())
		return
	}

	if _, err = conf.Database.UpgradeUserRedChirp(r.Context(), userID); err != nil {
		errorRespond(w, http.StatusInternalServerError, err.Error())
		return
	}

	successRespond(w, http.StatusNoContent, nil)
}
