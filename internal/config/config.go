// internal/config/config.go
package config

import (
	"os"
)

// getEnv returns the value of an environment variable or a default value.
func getEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

// APIConfig holds configuration for the API gateway.
type APIConfig struct {
	HTTPAddr            string
	OrchestratorGRPCAddr string
}

// OrchestratorConfig holds configuration for the orchestrator service.
type OrchestratorConfig struct {
	GRPCAddr       string
	WorkerGRPCAddr string
	DBDSN          string
}

// NodeConfig holds configuration for the node worker service.
type NodeConfig struct {
	GRPCAddr string
}

// LoadAPIConfig loads API configuration from environment variables.
func LoadAPIConfig() APIConfig {
	return APIConfig{
		HTTPAddr:            getEnv("API_HTTP_ADDR", ":8080"),
		OrchestratorGRPCAddr: getEnv("ORCHESTRATOR_GRPC_ADDR", "localhost:50051"),
	}
}

// LoadOrchestratorConfig loads Orchestrator configuration from environment variables.
func LoadOrchestratorConfig() OrchestratorConfig {
	return OrchestratorConfig{
		GRPCAddr:      getEnv("ORCHESTRATOR_GRPC_ADDR", ":50051"),
		WorkerGRPCAddr: getEnv("WORKER_GRPC_ADDR", "localhost:50052"),
		DBDSN:          getEnv("ORCHESTRATOR_DB_DSN", "postgres://autofarm:autofarm@postgres:5432/autofarm?sslmode=disable"),
	}
}

// LoadNodeConfig loads Node worker configuration from environment variables.
func LoadNodeConfig() NodeConfig {
	return NodeConfig{
		GRPCAddr: getEnv("NODE_GRPC_ADDR", ":50052"),
	}
}

