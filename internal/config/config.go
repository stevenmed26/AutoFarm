package config

// Config holds basic runtime configuration for the AutoFarm services.
type Config struct {
	HTTPPort    string
	GRPCPort    string
	Environment string
}

// Default returns a sensible default configuration for local development.
func Default() Config {
	return Config{
		HTTPPort:    "8080",
		GRPCPort:    "50051",
		Environment: "local",
	}
}
