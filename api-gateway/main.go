package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/redis/go-redis/v9"
)

var (
	ctx = context.Background()
	rdb *redis.Client
)

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

func health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func predict(w http.ResponseWriter, r *http.Request) {
	payload := map[string]any{
		"features": []float64{12.3, 45.6},
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
		"events_total":  int(total),
		"positive_total": int(pos),
		"score_avg":     avg,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

func main() {
	rdb = redisClient()

	http.HandleFunc("/health", health)
	http.HandleFunc("/predict", predict)
	http.HandleFunc("/stats", stats)

	http.ListenAndServe(":8080", nil)
}
