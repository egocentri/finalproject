package config

import "os"

type EnvConfig struct {
    HTTPPort       string
    GRPCPort       string
    JWTSecret      []byte
    TimeEvaluation int
}

func InitEnv() *EnvConfig {
    return &EnvConfig{
        HTTPPort:       getEnv("HTTP_PORT", "8080"),
        GRPCPort:       getEnv("GRPC_PORT", "50051"),
        JWTSecret:      []byte(getEnv("JWT_SECRET", "supersecret")),
        TimeEvaluation: getInt("TIME_EVALUATION_MS", 500),
    }
}

func getEnv(key, def string) string {
    if v := os.Getenv(key); v != "" { return v }
    return def
}

func getInt(key string, def int) int {
    if v := os.Getenv(key); v != "" {
        if i, err := strconv.Atoi(v); err == nil {
            return i
        }
    }
    return def
}
