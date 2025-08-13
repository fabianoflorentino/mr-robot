package payment

import (
	"os"
	"testing"
)

func TestConfigManager_LoadConfig(t *testing.T) {
	// Save original env vars
	originalDefault := os.Getenv("DEFAULT_PROCESSOR_URL")
	originalFallback := os.Getenv("FALLBACK_PROCESSOR_URL")

	// Cleanup function
	defer func() {
		if originalDefault == "" {
			os.Unsetenv("DEFAULT_PROCESSOR_URL")
		} else {
			os.Setenv("DEFAULT_PROCESSOR_URL", originalDefault)
		}
		if originalFallback == "" {
			os.Unsetenv("FALLBACK_PROCESSOR_URL")
		} else {
			os.Setenv("FALLBACK_PROCESSOR_URL", originalFallback)
		}
	}()

	t.Run("Valid URLs", func(t *testing.T) {
		os.Setenv("DEFAULT_PROCESSOR_URL", "http://default.example.com")
		os.Setenv("FALLBACK_PROCESSOR_URL", "http://fallback.example.com")

		cm := NewConfigManager()
		err := cm.LoadConfig()
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		config := cm.GetConfig()
		if config.DefaultProcessorURL != "http://default.example.com" {
			t.Errorf("Expected default URL to be 'http://default.example.com', got: %s", config.DefaultProcessorURL)
		}
		if config.FallbackProcessorURL != "http://fallback.example.com" {
			t.Errorf("Expected fallback URL to be 'http://fallback.example.com', got: %s", config.FallbackProcessorURL)
		}
	})

	t.Run("Missing default URL", func(t *testing.T) {
		os.Unsetenv("DEFAULT_PROCESSOR_URL")
		os.Setenv("FALLBACK_PROCESSOR_URL", "http://fallback.example.com")

		cm := NewConfigManager()
		err := cm.LoadConfig()
		if err == nil {
			t.Fatal("Expected error for missing default processor URL")
		}
	})

	t.Run("Missing fallback URL", func(t *testing.T) {
		os.Setenv("DEFAULT_PROCESSOR_URL", "http://default.example.com")
		os.Unsetenv("FALLBACK_PROCESSOR_URL")

		cm := NewConfigManager()
		err := cm.LoadConfig()
		if err == nil {
			t.Fatal("Expected error for missing fallback processor URL")
		}
	})
}

func TestConfigManager_Validate(t *testing.T) {
	t.Run("Valid config", func(t *testing.T) {
		cm := NewConfigManager()
		cm.SetConfig(&Config{
			DefaultProcessorURL:  "http://default.example.com",
			FallbackProcessorURL: "http://fallback.example.com",
		})

		err := cm.Validate()
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
	})

	t.Run("Empty default URL", func(t *testing.T) {
		cm := NewConfigManager()
		cm.SetConfig(&Config{
			DefaultProcessorURL:  "",
			FallbackProcessorURL: "http://fallback.example.com",
		})

		err := cm.Validate()
		if err == nil {
			t.Fatal("Expected error for empty default URL")
		}
	})

	t.Run("Empty fallback URL", func(t *testing.T) {
		cm := NewConfigManager()
		cm.SetConfig(&Config{
			DefaultProcessorURL:  "http://default.example.com",
			FallbackProcessorURL: "",
		})

		err := cm.Validate()
		if err == nil {
			t.Fatal("Expected error for empty fallback URL")
		}
	})

	t.Run("Invalid default URL", func(t *testing.T) {
		cm := NewConfigManager()
		cm.SetConfig(&Config{
			DefaultProcessorURL:  "://invalid-url",
			FallbackProcessorURL: "http://fallback.example.com",
		})

		err := cm.Validate()
		if err == nil {
			t.Fatal("Expected error for invalid default URL")
		}
	})

	t.Run("Invalid fallback URL", func(t *testing.T) {
		cm := NewConfigManager()
		cm.SetConfig(&Config{
			DefaultProcessorURL:  "http://default.example.com",
			FallbackProcessorURL: "://invalid-url",
		})

		err := cm.Validate()
		if err == nil {
			t.Fatal("Expected error for invalid fallback URL")
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
