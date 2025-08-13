package queue

import (
	"os"
	"testing"
)

func TestConfigManager_LoadConfig(t *testing.T) {
	// Save original env vars
	originalVars := map[string]string{
		"QUEUE_WORKERS":                 os.Getenv("QUEUE_WORKERS"),
		"QUEUE_BUFFER_SIZE":             os.Getenv("QUEUE_BUFFER_SIZE"),
		"QUEUE_MAX_ENQUEUE_RETRIES":     os.Getenv("QUEUE_MAX_ENQUEUE_RETRIES"),
		"QUEUE_MAX_SIMULTANEOUS_WRITES": os.Getenv("QUEUE_MAX_SIMULTANEOUS_WRITES"),
	}

	// Cleanup function
	defer func() {
		for key, value := range originalVars {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()

	t.Run("Default values", func(t *testing.T) {
		// Clear all env vars
		for key := range originalVars {
			os.Unsetenv(key)
		}

		cm := NewConfigManager()
		err := cm.LoadConfig()
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		config := cm.GetConfig()
		if config.Workers != 10 {
			t.Errorf("Expected workers to be 10, got: %d", config.Workers)
		}
		if config.BufferSize != 10000 {
			t.Errorf("Expected buffer size to be 10000, got: %d", config.BufferSize)
		}
		if config.MaxEnqueueRetries != 4 {
			t.Errorf("Expected max retries to be 4, got: %d", config.MaxEnqueueRetries)
		}
		if config.MaxSimultaneousWrites != 50 {
			t.Errorf("Expected max simultaneous writes to be 50, got: %d", config.MaxSimultaneousWrites)
		}
	})

	t.Run("Custom values", func(t *testing.T) {
		os.Setenv("QUEUE_WORKERS", "20")
		os.Setenv("QUEUE_BUFFER_SIZE", "20000")
		os.Setenv("QUEUE_MAX_ENQUEUE_RETRIES", "8")
		os.Setenv("QUEUE_MAX_SIMULTANEOUS_WRITES", "100")

		cm := NewConfigManager()
		err := cm.LoadConfig()
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		config := cm.GetConfig()
		if config.Workers != 20 {
			t.Errorf("Expected workers to be 20, got: %d", config.Workers)
		}
		if config.BufferSize != 20000 {
			t.Errorf("Expected buffer size to be 20000, got: %d", config.BufferSize)
		}
		if config.MaxEnqueueRetries != 8 {
			t.Errorf("Expected max retries to be 8, got: %d", config.MaxEnqueueRetries)
		}
		if config.MaxSimultaneousWrites != 100 {
			t.Errorf("Expected max simultaneous writes to be 100, got: %d", config.MaxSimultaneousWrites)
		}
	})

	t.Run("Invalid values", func(t *testing.T) {
		os.Setenv("QUEUE_WORKERS", "invalid")

		cm := NewConfigManager()
		err := cm.LoadConfig()
		if err == nil {
			t.Fatal("Expected error for invalid workers value")
		}
	})
}

func TestConfigManager_Validate(t *testing.T) {
	t.Run("Valid config", func(t *testing.T) {
		cm := NewConfigManager()
		cm.SetConfig(&Config{
			Workers:               10,
			BufferSize:            1000,
			MaxEnqueueRetries:     3,
			MaxSimultaneousWrites: 50,
		})

		err := cm.Validate()
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
	})

	t.Run("Invalid workers", func(t *testing.T) {
		cm := NewConfigManager()
		cm.SetConfig(&Config{
			Workers:               0,
			BufferSize:            1000,
			MaxEnqueueRetries:     3,
			MaxSimultaneousWrites: 50,
		})

		err := cm.Validate()
		if err == nil {
			t.Fatal("Expected error for invalid workers")
		}
	})

	t.Run("Invalid buffer size", func(t *testing.T) {
		cm := NewConfigManager()
		cm.SetConfig(&Config{
			Workers:               10,
			BufferSize:            0,
			MaxEnqueueRetries:     3,
			MaxSimultaneousWrites: 50,
		})

		err := cm.Validate()
		if err == nil {
			t.Fatal("Expected error for invalid buffer size")
		}
	})

	t.Run("Invalid simultaneous writes", func(t *testing.T) {
		cm := NewConfigManager()
		cm.SetConfig(&Config{
			Workers:               10,
			BufferSize:            1000,
			MaxEnqueueRetries:     3,
			MaxSimultaneousWrites: 0,
		})

		err := cm.Validate()
		if err == nil {
			t.Fatal("Expected error for invalid simultaneous writes")
		}
	})

	t.Run("Nil config", func(t *testing.T) {
		cm := NewConfigManager()

		err := cm.Validate()
		if err == nil {
			t.Fatal("Expected error for nil config")
		}
	})
}
