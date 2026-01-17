package urlShortener

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"strings"
	"time"

	"url_shortner/internal/middleware"

	"github.com/redis/go-redis/v9"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	Repo  Repository
	Redis *redis.Client
}

func (h *Handler) GetUrl(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	cachedUrl, err := h.Redis.Get(r.Context(), code).Result()

	if err == nil {
		fmt.Println("⚡ Cache Hit!")
		go h.Redis.Incr(r.Context(), "clicks:"+code)
		http.Redirect(w, r, cachedUrl, http.StatusFound)
		return
	}

	if err != redis.Nil {
		fmt.Println("⚠️ Redis is unreachable:", err)
	}

	fmt.Println("🐢 Cache Miss (Hitting DB)...")
	u, err := h.Repo.GetVal(r.Context(), code)
	if err != nil {
		http.Error(w, "Url not found", http.StatusNotFound)
		return
	}

	finalUrl := u.Url
	if !strings.HasPrefix(finalUrl, "http://") && !strings.HasPrefix(finalUrl, "https://") {
		finalUrl = "http://" + finalUrl
	}

	err = h.Redis.Set(r.Context(), code, finalUrl, time.Hour*6).Err()
	if err != nil {
		fmt.Println("⚠️ Failed to update cache:", err)
	}
	go h.Redis.Incr(r.Context(), "clicks:"+code)
	http.Redirect(w, r, finalUrl, http.StatusFound)

}

func (h *Handler) AddUrl(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	type req struct {
		Url string `json:"url"`
	}
	var reqData req
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		http.Error(w, "Invalid Format data", http.StatusBadRequest)
		return
	}

	if !strings.HasPrefix(reqData.Url, "http://") && !strings.HasPrefix(reqData.Url, "https://") {
		reqData.Url = "http://" + reqData.Url
	}
	userID := r.Context().Value(middleware.UserIDKey).(int)
	code := generateShortCode()
	model := UrlDbModel{
		Url:       reqData.Url,
		ShortCode: code,
		CreatedAt: time.Now(),
		UserID:    userID,
	}

	if err := h.Repo.Add(r.Context(), model); err != nil {
		fmt.Println(err.Error())
		http.Error(w, "Failed to save data", http.StatusInternalServerError)
		return
	}
	response := map[string]string{
		"shortCode": code,
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) GetMyUrls(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := r.Context().Value(middleware.UserIDKey).(int)

	urls, err := h.Repo.GetByUser(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to fetch URLs", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(urls)
}

func (h *Handler) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	clicks, err := h.Redis.Get(r.Context(), "clicks:"+code).Int64()

	if err == redis.Nil {
		clicks = 0
	} else if err != nil {
		http.Error(w, "Error fetching analytics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int64{"clicks": clicks})
}

func generateShortCode() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 6
	code := ""
	for i := 0; i < length; i++ {
		randomIndex := rand.IntN(len(charset))
		code += string(charset[randomIndex])

	}
	return code
}
