package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"
)

var (
	ctx     = context.Background()
	rdb     *redis.Client
	limiter = rate.NewLimiter(5, 10) // 5 req/sec, burst 10
)

// ---------- Redis ----------

func redisClient() *redis.Client {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "redis"
	}
	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = "6379"
	}
	return redis.NewClient(&redis.Options{
		Addr: host + ":" + port,
	})
}

// ---------- Middleware ----------

// CORS middleware (fixes browser blocked requests)
func withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

		// Preflight request
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next(w, r)
	}
}

// Security middleware (rate limit + API key)
func withSecurity(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Rate limit
		if !limiter.Allow() {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		// API key check (only if API_KEY is set)
		key := os.Getenv("API_KEY")
		if key != "" && r.Header.Get("X-API-Key") != key {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}

// ---------- Handlers ----------

func health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func predict(w http.ResponseWriter, r *http.Request) {
	// Option A: randomize features so score (and avg) changes
	payload := map[string]any{
		"features": []float64{
			rand.Float64() * 100,
			rand.Float64() * 100,
		},
	}

	body, _ := json.Marshal(payload)

	resp, err := http.Post("http://ml-service:5000/predict", "application/json", bytes.NewBuffer(body))
	if err != nil {
		http.Error(w, "ml-service unreachable", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(respBody)
}

func stats(w http.ResponseWriter, r *http.Request) {
	totalStr, _ := rdb.Get(ctx, "stats:events_total").Result()
	posStr, _ := rdb.Get(ctx, "stats:positive_total").Result()
	sumStr, _ := rdb.Get(ctx, "stats:score_sum").Result()

	// defaults
	if totalStr == "" {
		totalStr = "0"
	}
	if posStr == "" {
		posStr = "0"
	}
	if sumStr == "" {
		sumStr = "0"
	}

	total, _ := strconv.ParseFloat(totalStr, 64)
	pos, _ := strconv.ParseFloat(posStr, 64)
	sum, _ := strconv.ParseFloat(sumStr, 64)

	avg := 0.0
	if total > 0 {
		avg = sum / total
	}

	out := map[string]any{
		"events_total":   int(total),
		"positive_total": int(pos),
		"score_avg":      avg,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

// ---------- Main ----------

func main() {
	// Seed RNG once at startup (so features change each request)
	rand.Seed(time.Now().UnixNano())

	rdb = redisClient()

	// IMPORTANT: Order matters => CORS → Security → Handler
	http.HandleFunc("/health", withCORS(health))
	http.HandleFunc("/predict", withCORS(withSecurity(predict)))
	http.HandleFunc("/stats", withCORS(withSecurity(stats)))

	http.ListenAndServe(":8080", nil)
}
