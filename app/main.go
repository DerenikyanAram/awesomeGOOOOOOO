package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/segmentio/kafka-go"
)

type Config struct {
	HTTPAddr     string
	DBDSN        string
	KafkaBrokers []string
	KafkaTopic   string
	KafkaGroup   string
	CacheSize    int
	RepoKind     string
	KafkaOff     bool
}

func getenv(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}

func loadConfig() Config {
	_ = godotenv.Load(".env", "../.env")

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		host := getenv("PG_HOST", "localhost")
		port := getenv("PG_PORT", "5432")
		db := getenv("PG_DB", "orders")
		user := getenv("PG_USER", "postgres")
		pass := getenv("PG_PASSWORD", "postgres")
		ssl := getenv("PG_SSLMODE", "disable")
		dsn = "postgres://" + user + ":" + pass + "@" + host + ":" + port + "/" + db + "?sslmode=" + ssl
	}

	brokers := getenv("KAFKA_BROKERS", "localhost:9092")

	return Config{
		HTTPAddr:     getenv("HTTP_ADDR", ":8081"),
		DBDSN:        dsn,
		KafkaBrokers: strings.Split(brokers, ","),
		KafkaTopic:   getenv("KAFKA_TOPIC", "orders"),
		KafkaGroup:   getenv("KAFKA_GROUP", "order-consumer"),
		CacheSize:    1024,
		RepoKind:     strings.ToLower(getenv("REPO", "pg")),
		KafkaOff:     getenv("KAFKA_DISABLED", "") == "1",
	}
}

func main() {
	cfg := loadConfig()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := pgxpool.New(ctx, cfg.DBDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	repo := NewRepository(pool)
	cache := NewCache(cfg.CacheSize)

	go consume(ctx, cfg, repo, cache)

	r := mux.NewRouter()
	r.Use(recoverMW)

	r.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodGet)

	r.HandleFunc("/api/validate", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
		dec.DisallowUnknownFields()

		var o Order
		if err := dec.Decode(&o); err != nil {
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}
		if o.DateCreated.IsZero() {
			o.DateCreated = time.Now().UTC()
		}
		if err := ValidateOrder(o); err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(o); err != nil {
			http.Error(w, "encode error", http.StatusInternalServerError)
			return
		}
	}).Methods(http.MethodPost)

	r.HandleFunc("/api/orders/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		if id == "" {
			http.NotFound(w, r)
			return
		}
		if v, ok := cache.Get(id); ok {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(v)
			return
		}
		cctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		o, err := repo.Get(cctx, id)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		cache.Set(o)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(o)
	}).Methods(http.MethodGet)

	srv := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}

func recoverMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func consume(ctx context.Context, cfg Config, repo Repository, cache *Cache) {
	_ = ensureTopic(cfg)
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  cfg.KafkaBrokers,
		Topic:    cfg.KafkaTopic,
		GroupID:  cfg.KafkaGroup,
		MinBytes: 1,
		MaxBytes: 10e6,
	})
	defer r.Close()

	for {
		m, err := r.ReadMessage(ctx)
		if err != nil {
			break
		}
		var o Order
		if err := json.Unmarshal(m.Value, &o); err != nil {
			continue
		}
		if o.DateCreated.IsZero() {
			o.DateCreated = time.Now().UTC()
		}
		if err := ValidateOrder(o); err != nil {
			continue
		}
		cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		_ = repo.Upsert(cctx, o)
		cancel()
		cache.Set(o)
	}
}

func ensureTopic(cfg Config) error {
	conn, err := kafka.Dial("tcp", cfg.KafkaBrokers[0])
	if err != nil {
		return err
	}
	defer conn.Close()
	return conn.CreateTopics(kafka.TopicConfig{Topic: cfg.KafkaTopic, NumPartitions: 1, ReplicationFactor: 1})
}
