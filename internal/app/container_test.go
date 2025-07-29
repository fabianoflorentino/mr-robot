package app

import (
	"errors"
	"testing"

	"github.com/fabianoflorentino/mr-robot/config"
	"gorm.io/gorm"
)

// MockDatabaseConnection implementa a interface DatabaseConnection para testes
type MockDatabaseConnection struct {
	connectFunc func() (*gorm.DB, error)
	closeFunc   func() error
	db          *gorm.DB
}

func NewMockDatabaseConnection() *MockDatabaseConnection {
	return &MockDatabaseConnection{
		connectFunc: func() (*gorm.DB, error) {
			return &gorm.DB{}, nil
		},
		closeFunc: func() error {
			return nil
		},
	}
}

func (m *MockDatabaseConnection) Connect() (*gorm.DB, error) {
	if m.connectFunc != nil {
		db, err := m.connectFunc()
		m.db = db
		return db, err
	}
	return &gorm.DB{}, nil
}

// Close simulate the closing of the database connection
func (m *MockDatabaseConnection) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

// GetDB return the mock database instance
func (m *MockDatabaseConnection) GetDB() *gorm.DB {
	return m.db
}

// SetConnectFunc allows to configure the behavior of Connect
func (m *MockDatabaseConnection) SetConnectFunc(fn func() (*gorm.DB, error)) {
	m.connectFunc = fn
}

// SetCloseFunc allow to configure the behavior of Close
// This is useful to simulate errors during shutdown
func (m *MockDatabaseConnection) SetCloseFunc(fn func() error) {
	m.closeFunc = fn
}

// TestAppConfigCreation tests the creation of app configuration
func TestAppConfigCreation(t *testing.T) {
	t.Setenv("POSTGRES_HOST", "localhost")
	t.Setenv("POSTGRES_PORT", "5432")
	t.Setenv("POSTGRES_USER", "test")
	t.Setenv("POSTGRES_PASSWORD", "test")
	t.Setenv("POSTGRES_DB", "test_db")
	t.Setenv("POSTGRES_SSLMODE", "disable")
	t.Setenv("POSTGRES_TIMEZONE", "UTC")
	t.Setenv("DEFAULT_PROCESSOR_URL", "http://localhost:8080")
	t.Setenv("QUEUE_WORKERS", "4")
	t.Setenv("QUEUE_BUFFER_SIZE", "100")
	t.Setenv("SKIP_ENV_FILE", "true")

	cfg, err := config.LoadAppConfig()
	if err != nil {
		t.Fatalf("Expected no error loading config, got: %v", err)
	}
	if cfg == nil {
		t.Fatal("Expected config to be not nil")
	}

	// Verify config values
	if cfg.Database.Host != "localhost" {
		t.Errorf("Expected host to be 'localhost', got: %s", cfg.Database.Host)
	}
	if cfg.Database.Port != "5432" {
		t.Errorf("Expected port to be '5432', got: %s", cfg.Database.Port)
	}
	if cfg.Payment.DefaultProcessorURL != "http://localhost:8080" {
		t.Errorf("Expected processor URL to be 'http://localhost:8080', got: %s", cfg.Payment.DefaultProcessorURL)
	}
	if cfg.Queue.Workers != 4 {
		t.Errorf("Expected workers to be 4, got: %d", cfg.Queue.Workers)
	}
	if cfg.Queue.BufferSize != 100 {
		t.Errorf("Expected buffer size to be 100, got: %d", cfg.Queue.BufferSize)
	}
}

// TestContainerBuilder_Creation tests the basic creation of container builder
func TestContainerBuilder_Creation(t *testing.T) {
	builder := NewContainerBuilder()
	if builder == nil {
		t.Fatal("Expected builder to be not nil")
	}
}

// TestContainerBuilder_ChainedMethods tests the fluent interface of the builder
func TestContainerBuilder_ChainedMethods(t *testing.T) {
	// Test that builder methods can be chained
	builder := NewContainerBuilder()
	if builder == nil {
		t.Fatal("Expected builder to be not nil")
	}

	// Create a config
	cfg := &config.AppConfig{
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     "5432",
			User:     "test",
			Password: "test",
			Database: "test_db",
			SSLMode:  "disable",
			Timezone: "UTC",
		},
		Payment: config.PaymentConfig{
			DefaultProcessorURL: "http://localhost:8080",
		},
		Queue: config.QueueConfig{
			Workers:    3,
			BufferSize: 75,
		},
	}

	// Test method chaining
	builderWithConfig := builder.WithConfig(cfg)
	if builderWithConfig == nil {
		t.Error("Expected WithConfig to return builder instance")
	}

	// Verify it's the same instance (fluent interface)
	if builderWithConfig != builder {
		t.Error("Expected WithConfig to return the same builder instance")
	}
}

// TestConfigDefaults tests default values when environment variables are not set
func TestConfigDefaults(t *testing.T) {
	// Clear environment variables that might affect the test
	t.Setenv("POSTGRES_HOST", "")
	t.Setenv("POSTGRES_PORT", "")
	t.Setenv("POSTGRES_USER", "")
	t.Setenv("QUEUE_WORKERS", "")
	t.Setenv("QUEUE_BUFFER_SIZE", "")
	t.Setenv("SKIP_ENV_FILE", "true")

	cfg, err := config.LoadAppConfig()
	if err != nil {
		t.Fatalf("Expected no error loading config with defaults, got: %v", err)
	}

	// Test default values
	if cfg.Database.Host != "localhost" {
		t.Errorf("Expected default host to be 'localhost', got: %s", cfg.Database.Host)
	}
	if cfg.Database.Port != "5432" {
		t.Errorf("Expected default port to be '5432', got: %s", cfg.Database.Port)
	}
	if cfg.Database.User != "postgres" {
		t.Errorf("Expected default user to be 'postgres', got: %s", cfg.Database.User)
	}
	if cfg.Queue.Workers != 4 {
		t.Errorf("Expected default workers to be 4, got: %d", cfg.Queue.Workers)
	}
	if cfg.Queue.BufferSize != 100 {
		t.Errorf("Expected default buffer size to be 100, got: %d", cfg.Queue.BufferSize)
	}
}

// TestContainerInterface tests that the container implements the expected interface
func TestContainerInterface(t *testing.T) {
	// This test verifies that AppContainer implements Container interface
	var _ Container = &AppContainer{}

	// Test that the interface has the expected methods
	container := &AppContainer{}

	// These should not panic (nil pointer calls are expected to panic, but type assertions shouldn't)
	_ = container.GetDB
	_ = container.GetPaymentService
	_ = container.GetPaymentQueue
	_ = container.Shutdown
}

// TestConfigInvalidWorkers tests error handling for invalid worker count
func TestConfigInvalidWorkers(t *testing.T) {
	t.Setenv("QUEUE_WORKERS", "invalid")
	t.Setenv("QUEUE_BUFFER_SIZE", "50")
	t.Setenv("SKIP_ENV_FILE", "true")

	cfg, err := config.LoadAppConfig()
	if err != nil {
		t.Fatalf("Expected no error loading config, got: %v", err)
	}

	// Should fallback to default value (4) when parsing fails
	if cfg.Queue.Workers != 4 {
		t.Errorf("Expected workers to fallback to 4 when invalid, got: %d", cfg.Queue.Workers)
	}
}

// TestConfigInvalidBufferSize tests error handling for invalid buffer size
func TestConfigInvalidBufferSize(t *testing.T) {
	t.Setenv("QUEUE_WORKERS", "2")
	t.Setenv("QUEUE_BUFFER_SIZE", "not-a-number")
	t.Setenv("SKIP_ENV_FILE", "true")

	cfg, err := config.LoadAppConfig()
	if err != nil {
		t.Fatalf("Expected no error loading config, got: %v", err)
	}

	// Should fallback to default value (100) when parsing fails
	if cfg.Queue.BufferSize != 100 {
		t.Errorf("Expected buffer size to fallback to 100 when invalid, got: %d", cfg.Queue.BufferSize)
	}
}

// TestBuilderWithNilConfig tests builder behavior with nil config
func TestBuilderWithNilConfig(t *testing.T) {
	builder := NewContainerBuilder()

	// Test with nil config - should not panic
	builderWithNil := builder.WithConfig(nil)
	if builderWithNil != builder {
		t.Error("Expected WithConfig to return the same builder instance even with nil")
	}
}

// TestContainerBuilder_WithDatabaseConnection tests the WithDatabaseConnection method
func TestContainerBuilder_WithDatabaseConnection(t *testing.T) {
	builder := NewContainerBuilder()
	mockDB := NewMockDatabaseConnection()

	// Test method chaining with database connection
	builderWithDB := builder.WithDatabaseConnection(mockDB)
	if builderWithDB != builder {
		t.Error("Expected WithDatabaseConnection to return the same builder instance")
	}
}

// TestContainerGetters tests the getter methods with mock data
func TestContainerGetters(t *testing.T) {
	// Create a container with mock dependencies
	mockDB := NewMockDatabaseConnection()

	container := &AppContainer{
		db:             &gorm.DB{},
		paymentService: nil, // Will be nil for this test
		paymentQueue:   nil, // Will be nil for this test
		dbConnection:   mockDB,
	}

	// Test getters
	db := container.GetDB()
	if db == nil {
		t.Error("Expected GetDB to return the database instance")
	}

	paymentService := container.GetPaymentService()
	if paymentService != nil {
		t.Error("Expected GetPaymentService to return nil for this test")
	}

	paymentQueue := container.GetPaymentQueue()
	if paymentQueue != nil {
		t.Error("Expected GetPaymentQueue to return nil for this test")
	}
}

// TestContainerShutdown tests the Shutdown method
func TestContainerShutdown(t *testing.T) {
	mockDB := NewMockDatabaseConnection()

	container := &AppContainer{
		dbConnection: mockDB,
	}

	// Test successful shutdown
	err := container.Shutdown()
	if err != nil {
		t.Errorf("Expected no error during shutdown, got: %v", err)
	}
}

// TestContainerShutdown_WithError tests the Shutdown method with database error
func TestContainerShutdown_WithError(t *testing.T) {
	mockDB := NewMockDatabaseConnection()
	expectedError := errors.New("database close error")

	// Configure mock to return error on close
	mockDB.SetCloseFunc(func() error {
		return expectedError
	})

	container := &AppContainer{
		dbConnection: mockDB,
	}

	// Test shutdown with error
	err := container.Shutdown()
	if err == nil {
		t.Error("Expected error during shutdown, got nil")
	}
	if !errors.Is(err, expectedError) {
		t.Errorf("Expected error to contain database close error, got: %v", err)
	}
}

// TestContainerShutdown_NilConnection tests shutdown with nil connection
func TestContainerShutdown_NilConnection(t *testing.T) {
	container := &AppContainer{
		dbConnection: nil,
	}

	// Test shutdown with nil connection - should not panic
	err := container.Shutdown()
	if err != nil {
		t.Errorf("Expected no error with nil connection, got: %v", err)
	}
}

// TestContainerBuilder_Build_ConfigError tests Build method with config loading error
func TestContainerBuilder_Build_ConfigError(t *testing.T) {
	// This test is skipped because LoadEnv uses log.Fatalf which terminates the process
	// In a real scenario, we would need to refactor LoadEnv to return errors instead of calling log.Fatalf
	t.Skip("Skipping test that would cause log.Fatalf - needs refactoring of LoadEnv")
}

// TestContainerBuilder_Build_DatabaseConnectionError tests Build with database connection error
func TestContainerBuilder_Build_DatabaseConnectionError(t *testing.T) {
	// Setup valid config
	cfg := &config.AppConfig{
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     "5432",
			User:     "test",
			Password: "test",
			Database: "test_db",
			SSLMode:  "disable",
			Timezone: "UTC",
		},
		Payment: config.PaymentConfig{
			DefaultProcessorURL: "http://localhost:8080",
		},
		Queue: config.QueueConfig{
			Workers:    2,
			BufferSize: 50,
		},
	}

	// Create mock that fails on connect
	mockDB := NewMockDatabaseConnection()
	mockDB.SetConnectFunc(func() (*gorm.DB, error) {
		return nil, errors.New("connection failed")
	})

	builder := NewContainerBuilder().
		WithConfig(cfg).
		WithDatabaseConnection(mockDB)

	_, err := builder.Build()
	if err == nil {
		t.Error("Expected error when database connection fails, got nil")
	}
}

// TestContainerBuilder_Build_Success tests successful Build with mocks
func TestContainerBuilder_Build_Success(t *testing.T) {
	// This test is limited because we can't properly mock GORM without extensive setup
	// In a real scenario, we would use an in-memory database or testcontainers
	t.Skip("Skipping full build test - requires proper GORM setup. Test individual components instead.")
}

// BenchmarkContainerBuilder_Creation benchmarks the creation of container builder
func BenchmarkContainerBuilder_Creation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		builder := NewContainerBuilder()
		_ = builder
	}
}

// BenchmarkContainerBuilder_WithConfig benchmarks the WithConfig method
func BenchmarkContainerBuilder_WithConfig(b *testing.B) {
	cfg := &config.AppConfig{
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     "5432",
			User:     "test",
			Password: "test",
			Database: "test_db",
			SSLMode:  "disable",
			Timezone: "UTC",
		},
		Payment: config.PaymentConfig{
			DefaultProcessorURL: "http://localhost:8080",
		},
		Queue: config.QueueConfig{
			Workers:    2,
			BufferSize: 50,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		builder := NewContainerBuilder()
		_ = builder.WithConfig(cfg)
	}
}

// TestTableDriven_ConfigValidation uses table-driven tests for config validation
func TestTableDriven_ConfigValidation(t *testing.T) {
	tests := []struct {
		name           string
		envVars        map[string]string
		expectedResult func(*config.AppConfig) bool
		description    string
	}{
		{
			name: "valid_config",
			envVars: map[string]string{
				"POSTGRES_HOST":         "localhost",
				"POSTGRES_PORT":         "5432",
				"POSTGRES_USER":         "testuser",
				"QUEUE_WORKERS":         "8",
				"QUEUE_BUFFER_SIZE":     "200",
				"DEFAULT_PROCESSOR_URL": "http://test.com",
				"SKIP_ENV_FILE":         "true",
			},
			expectedResult: func(cfg *config.AppConfig) bool {
				return cfg.Database.Host == "localhost" &&
					cfg.Queue.Workers == 8 &&
					cfg.Queue.BufferSize == 200
			},
			description: "should load valid configuration from environment",
		},
		{
			name: "default_fallback",
			envVars: map[string]string{
				"SKIP_ENV_FILE": "true",
			},
			expectedResult: func(cfg *config.AppConfig) bool {
				return cfg.Database.Host == "localhost" &&
					cfg.Database.Port == "5432" &&
					cfg.Queue.Workers == 4
			},
			description: "should use default values when env vars are not set",
		},
		{
			name: "invalid_numbers_fallback",
			envVars: map[string]string{
				"QUEUE_WORKERS":     "not-a-number",
				"QUEUE_BUFFER_SIZE": "also-not-a-number",
				"SKIP_ENV_FILE":     "true",
			},
			expectedResult: func(cfg *config.AppConfig) bool {
				return cfg.Queue.Workers == 4 && cfg.Queue.BufferSize == 100
			},
			description: "should fallback to defaults when numbers are invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables for this test
			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			cfg, err := config.LoadAppConfig()
			if err != nil {
				t.Fatalf("Expected no error loading config, got: %v", err)
			}

			if !tt.expectedResult(cfg) {
				t.Errorf("Test %s failed: %s", tt.name, tt.description)
			}
		})
	}
}

// TestContainerBuilder_MultipleMethodCalls tests calling multiple methods
func TestContainerBuilder_MultipleMethodCalls(t *testing.T) {
	builder := NewContainerBuilder()

	cfg := &config.AppConfig{
		Database: config.DatabaseConfig{Host: "test"},
		Payment:  config.PaymentConfig{DefaultProcessorURL: "http://test"},
		Queue:    config.QueueConfig{Workers: 1, BufferSize: 10},
	}

	mockDB := NewMockDatabaseConnection()

	// Test multiple method calls don't interfere with each other
	result := builder.
		WithConfig(cfg).
		WithDatabaseConnection(mockDB).
		WithConfig(cfg) // Call again to ensure it's idempotent

	if result != builder {
		t.Error("Expected fluent interface to return same builder")
	}
}

// TestErrorWrapping tests that errors are properly wrapped
func TestErrorWrapping(t *testing.T) {
	mockDB := NewMockDatabaseConnection()
	originalError := errors.New("original database error")

	mockDB.SetConnectFunc(func() (*gorm.DB, error) {
		return nil, originalError
	})

	cfg := &config.AppConfig{
		Database: config.DatabaseConfig{Host: "test"},
		Payment:  config.PaymentConfig{},
		Queue:    config.QueueConfig{Workers: 1, BufferSize: 10},
	}

	builder := NewContainerBuilder().
		WithConfig(cfg).
		WithDatabaseConnection(mockDB)

	_, err := builder.Build()
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	// Check that the error is properly wrapped
	if !errors.Is(err, originalError) {
		t.Errorf("Expected error to wrap original error, got: %v", err)
	}
}
