package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/pflag"
)

type PoolConfig struct {
	MaxConnLifetime       time.Duration
	MaxConnLifetimeJitter time.Duration
	MaxConnIdleTime       time.Duration
	MaxConns              int32
	MinConns              int32
	MinIdleConns          int32
	HealthCheckPeriod     time.Duration
	PingMaxRetries        int
	PingRetryDelay        time.Duration
}

type DBConfig struct {
	PgHost     string
	PgPort     string
	PgDB       string
	PgUser     string
	PgPassword string
	Pool       PoolConfig
}

type OrderConfig struct {
	ServiceHost   string
	CheckInterval time.Duration
	MaxRetries    int
	RetryDelay    time.Duration
}

type SaramaConfig struct {
	Version            string
	AutoCommitInterval time.Duration
}

type KafkaConfig struct {
	Brokers       []string
	Topic         string
	ConsumerGroup string
	Sarama        SaramaConfig
}

type RateLimitConfig struct {
	Capacity   int
	RefillRate int
}

type Config struct {
	Port                  string
	PprofPort             string
	ShutdownTimeout       time.Duration
	DB                    DBConfig
	DeliveryCheckInterval time.Duration
	Order                 OrderConfig
	Kafka                 KafkaConfig
	RateLimit             RateLimitConfig
}

func Load() *Config {
	_ = godotenv.Load()

	cfg := &Config{
		Port:                  getEnvAsString("PORT", "8080"),
		PprofPort:             getEnvAsString("PPROF_PORT", "6060"),
		ShutdownTimeout:       getEnvAsDuration("SHUTDOWN_TIMEOUT", 5*time.Second),
		DeliveryCheckInterval: getEnvAsDuration("DELIVERY_CHECK_INTERVAL", 10*time.Second),
		DB: DBConfig{
			PgHost:     getEnvAsString("POSTGRES_HOST", "localhost"),
			PgPort:     getEnvAsString("POSTGRES_PORT", "5432"),
			PgDB:       getEnvAsString("POSTGRES_DB", "testdb"),
			PgUser:     getEnvAsString("POSTGRES_USER", "myuser"),
			PgPassword: getEnvAsString("POSTGRES_PASSWORD", "mypassword"),
			Pool: PoolConfig{
				MaxConnLifetime:       getEnvAsDuration("POSTGRES_MAX_CONN_LIFETIME", time.Hour),
				MaxConnLifetimeJitter: getEnvAsDuration("POSTGRES_MAX_CONN_LIFETIME_JITTER", 5*time.Minute),
				MaxConnIdleTime:       getEnvAsDuration("POSTGRES_MAX_CONN_IDLE_TIME", 30*time.Minute),
				MaxConns:              getEnvAsInt32("POSTGRES_MAX_CONNS", 20),
				MinConns:              getEnvAsInt32("POSTGRES_MIN_CONNS", 5),
				MinIdleConns:          getEnvAsInt32("POSTGRES_MIN_IDLE_CONNS", 2),
				HealthCheckPeriod:     getEnvAsDuration("POSTGRES_HEALTH_CHECK_PERIOD", time.Minute),
				PingMaxRetries:        getEnvAsInt("POSTGRES_MAX_RETRIES", 5),
				PingRetryDelay:        getEnvAsDuration("POSTGRES_RETRY_DELAY", time.Second),
			},
		},
		Order: OrderConfig{
			ServiceHost:   getEnvAsString("ORDER_SERVICE_HOST", "service-order:50051"),
			CheckInterval: getEnvAsDuration("ORDER_CHECK_INTERVAL", 5*time.Second),
			MaxRetries:    getEnvAsInt("ORDER_MAX_RETRIES", 3),
			RetryDelay:    getEnvAsDuration("ORDER_RETRY_DELAY", 100*time.Millisecond),
		},
		Kafka: KafkaConfig{
			Brokers:       getEnvAsStringSlice("KAFKA_BROKERS", []string{"kafka:9092"}),
			Topic:         getEnvAsString("KAFKA_ORDERS_TOPIC", "orders"),
			ConsumerGroup: getEnvAsString("KAFKA_CONSUMER_GROUP", "courier-service"),
			Sarama: SaramaConfig{
				Version:            getEnvAsString("KAFKA_VERSION", "2.8.0"),
				AutoCommitInterval: getEnvAsDuration("KAFKA_AUTOCOMMIT_INTERVAL", 1*time.Second),
			},
		},
		RateLimit: RateLimitConfig{
			Capacity:   getEnvAsInt("RATE_LIMIT_CAPACITY", 100),
			RefillRate: getEnvAsInt("RATE_LIMIT_REFILL_RATE", 10),
		},
	}

	pflag.StringVarP(&cfg.Port, "port", "p", cfg.Port, "Port to listen on")
	pflag.Parse()

	return cfg
}

func getEnvAs[T any](key string, defaultVal T, parse func(string) (T, error)) T {
	if val := os.Getenv(key); val != "" {
		if v, err := parse(val); err == nil {
			return v
		}
	}
	return defaultVal
}

func getEnvAsString(key string, defaultVal string) string {
	return getEnvAs[string](key, defaultVal, func(s string) (string, error) {
		return s, nil
	})
}

func getEnvAsInt(key string, defaultVal int) int {
	return getEnvAs[int](key, defaultVal, strconv.Atoi)
}

func getEnvAsInt32(key string, defaultVal int32) int32 {
	return getEnvAs[int32](key, defaultVal, func(s string) (int32, error) {
		v, err := strconv.ParseInt(s, 10, 32)
		return int32(v), err
	})
}

func getEnvAsDuration(key string, defaultVal time.Duration) time.Duration {
	return getEnvAs(key, defaultVal, time.ParseDuration)
}

func getEnvAsStringSlice(key string, defaultVal []string) []string {
	return getEnvAs[[]string](key, defaultVal, func(s string) ([]string, error) {
		var result []string
		for _, v := range strings.Split(s, ",") {
			result = append(result, strings.TrimSpace(v))
		}
		return result, nil
	})
}
