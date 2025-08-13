package config

import (
	"os"
	"testing"
)

func TestConfigManager_Integration(t *testing.T) {
	// Save original env vars
	envVars := []string{
		"POSTGRES_HOST", "POSTGRES_PORT", "POSTGRES_USER", "POSTGRES_PASSWORD",
		"POSTGRES_DB", "POSTGRES_SSLMODE", "POSTGRES_TIMEZONE",
		"DEFAULT_PROCESSOR_URL", "FALLBACK_PROCESSOR_URL",
		"QUEUE_WORKERS", "QUEUE_BUFFER_SIZE", "QUEUE_MAX_ENQUEUE_RETRIES", "QUEUE_MAX_SIMULTANEOUS_WRITES",
		"CIRCUIT_BREAKER_TIMEOUT", "CIRCUIT_BREAKER_MAX_FAILURES", "CIRCUIT_BREAKER_RESET_TIMEOUT", "CIRCUIT_BREAKER_RATE_LIMIT",
		"HOSTNAME",
	}

	originalValues := make(map[string]string)
	for _, env := range envVars {
		originalValues[env] = os.Getenv(env)
	}

	// Cleanup function
	defer func() {
		for _, env := range envVars {
			if original := originalValues[env]; original == "" {
				os.Unsetenv(env)
			} else {
				os.Setenv(env, original)
			}
		}
	}()

	t.Run("Load all configurations successfully", func(t *testing.T) {
		// Set required environment variables
		os.Setenv("DEFAULT_PROCESSOR_URL", "http://default.example.com")
		os.Setenv("FALLBACK_PROCESSOR_URL", "http://fallback.example.com")

		manager := NewManager()

		// Test loading
		err := manager.LoadConfiguration()
		if err != nil {
			t.Fatalf("Failed to load configuration: %v", err)
		}

		// Test validation
		err = manager.ValidateConfiguration()
		if err != nil {
			t.Fatalf("Failed to validate configuration: %v", err)
		}

		// Test individual configs
		dbConfig := manager.GetDatabaseConfig()
		if dbConfig == nil {
			t.Fatal("Database config is nil")
		}
		if dbConfig.Host != "localhost" {
			t.Errorf("Expected database host to be 'localhost', got: %s", dbConfig.Host)
		}

		paymentConfig := manager.GetPaymentConfig()
		if paymentConfig == nil {
			t.Fatal("Payment config is nil")
		}
		if paymentConfig.DefaultProcessorURL != "http://default.example.com" {
			t.Errorf("Expected default processor URL to be 'http://default.example.com', got: %s", paymentConfig.DefaultProcessorURL)
		}

		queueConfig := manager.GetQueueConfig()
		if queueConfig == nil {
			t.Fatal("Queue config is nil")
		}
		if queueConfig.Workers != 10 {
			t.Errorf("Expected workers to be 10, got: %d", queueConfig.Workers)
		}

		cbConfig := manager.GetCircuitBreakerConfig()
		if cbConfig == nil {
			t.Fatal("Circuit breaker config is nil")
		}
		if cbConfig.MaxFailures != 5 {
			t.Errorf("Expected max failures to be 5, got: %d", cbConfig.MaxFailures)
		}

		controllerConfig := manager.GetControllerConfig()
		if controllerConfig == nil {
			t.Fatal("Controller config is nil")
		}
		if controllerConfig.HostName != "localhost" {
			t.Errorf("Expected hostname to be 'localhost', got: %s", controllerConfig.HostName)
		}
	})

	t.Run("Fail on missing required configuration", func(t *testing.T) {
		// Clear required variables
		os.Unsetenv("DEFAULT_PROCESSOR_URL")
		os.Unsetenv("FALLBACK_PROCESSOR_URL")

		manager := NewManager()

		// Should fail to load
		err := manager.LoadConfiguration()
		if err == nil {
			t.Fatal("Expected error when missing required configuration")
		}
	})

	t.Run("Individual manager access", func(t *testing.T) {
		// Set required environment variables
		os.Setenv("DEFAULT_PROCESSOR_URL", "http://default.example.com")
		os.Setenv("FALLBACK_PROCESSOR_URL", "http://fallback.example.com")

		manager := NewManager()
		err := manager.LoadConfiguration()
		if err != nil {
			t.Fatalf("Failed to load configuration: %v", err)
		}

		// Test manager access
		dbManager := manager.GetDatabaseManager()
		if dbManager == nil {
			t.Fatal("Database manager is nil")
		}

		paymentManager := manager.GetPaymentManager()
		if paymentManager == nil {
			t.Fatal("Payment manager is nil")
		}

		queueManager := manager.GetQueueManager()
		if queueManager == nil {
			t.Fatal("Queue manager is nil")
		}

		cbManager := manager.GetCircuitBreakerManager()
		if cbManager == nil {
			t.Fatal("Circuit breaker manager is nil")
		}

		controllerManager := manager.GetControllerManager()
		if controllerManager == nil {
			t.Fatal("Controller manager is nil")
		}
	})
}
