package secrets

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFromEnvFile(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")

	envContent := `
# This is a comment
DATABASE_URL=postgres://user:pass@localhost/db
API_KEY=secret-key-123
EMPTY_VALUE=

# Another comment
JWT_SECRET="quoted-value"
SINGLE_QUOTED='single-quoted'
`
	if err := os.WriteFile(envPath, []byte(envContent), 0644); err != nil {
		t.Fatal(err)
	}

	secrets, err := LoadFromEnvFile(envPath, []string{"DATABASE_URL", "API_KEY", "JWT_SECRET", "SINGLE_QUOTED"})
	if err != nil {
		t.Fatalf("LoadFromEnvFile failed: %v", err)
	}

	if len(secrets) != 4 {
		t.Errorf("expected 4 secrets, got %d", len(secrets))
	}

	// Check values
	expected := map[string]string{
		"DATABASE_URL":  "postgres://user:pass@localhost/db",
		"API_KEY":       "secret-key-123",
		"JWT_SECRET":    "quoted-value",
		"SINGLE_QUOTED": "single-quoted",
	}

	for _, s := range secrets {
		exp, ok := expected[s.Name]
		if !ok {
			t.Errorf("unexpected secret: %s", s.Name)
			continue
		}
		if s.Value != exp {
			t.Errorf("secret %s: expected '%s', got '%s'", s.Name, exp, s.Value)
		}
	}
}

func TestLoadFromEnvFileWithMissingSecrets(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")

	envContent := `
DATABASE_URL=postgres://localhost/db
`
	if err := os.WriteFile(envPath, []byte(envContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Request secrets that don't exist in file
	secrets, err := LoadFromEnvFile(envPath, []string{"DATABASE_URL", "NONEXISTENT"})
	if err != nil {
		t.Fatalf("LoadFromEnvFile failed: %v", err)
	}

	// Only DATABASE_URL should be found
	if len(secrets) != 1 {
		t.Errorf("expected 1 secret, got %d", len(secrets))
	}

	if secrets[0].Name != "DATABASE_URL" {
		t.Errorf("expected DATABASE_URL, got %s", secrets[0].Name)
	}
}

func TestLoadFromEnvFileFallbackToEnv(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")

	envContent := `
FILE_SECRET=from-file
`
	if err := os.WriteFile(envPath, []byte(envContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Set environment variable for secret not in file
	os.Setenv("ENV_SECRET", "from-env")
	defer os.Unsetenv("ENV_SECRET")

	secrets, err := LoadFromEnvFile(envPath, []string{"FILE_SECRET", "ENV_SECRET"})
	if err != nil {
		t.Fatalf("LoadFromEnvFile failed: %v", err)
	}

	if len(secrets) != 2 {
		t.Errorf("expected 2 secrets, got %d", len(secrets))
	}

	found := map[string]string{}
	for _, s := range secrets {
		found[s.Name] = s.Value
	}

	if found["FILE_SECRET"] != "from-file" {
		t.Errorf("expected FILE_SECRET='from-file', got '%s'", found["FILE_SECRET"])
	}

	if found["ENV_SECRET"] != "from-env" {
		t.Errorf("expected ENV_SECRET='from-env', got '%s'", found["ENV_SECRET"])
	}
}

func TestLoadFromEnvFileNotFound(t *testing.T) {
	_, err := LoadFromEnvFile("/nonexistent/.env", []string{"SECRET"})
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestLoadFromEnvFileEmptySecretNames(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")

	envContent := `
DATABASE_URL=postgres://localhost/db
`
	if err := os.WriteFile(envPath, []byte(envContent), 0644); err != nil {
		t.Fatal(err)
	}

	secrets, err := LoadFromEnvFile(envPath, []string{})
	if err != nil {
		t.Fatalf("LoadFromEnvFile failed: %v", err)
	}

	if len(secrets) != 0 {
		t.Errorf("expected 0 secrets, got %d", len(secrets))
	}
}

func TestLoadFromEnv(t *testing.T) {
	os.Setenv("TEST_SECRET_1", "value1")
	os.Setenv("TEST_SECRET_2", "value2")
	defer os.Unsetenv("TEST_SECRET_1")
	defer os.Unsetenv("TEST_SECRET_2")

	secrets := LoadFromEnv([]string{"TEST_SECRET_1", "TEST_SECRET_2", "NONEXISTENT"})

	if len(secrets) != 2 {
		t.Errorf("expected 2 secrets, got %d", len(secrets))
	}

	found := map[string]string{}
	for _, s := range secrets {
		found[s.Name] = s.Value
	}

	if found["TEST_SECRET_1"] != "value1" {
		t.Errorf("expected TEST_SECRET_1='value1', got '%s'", found["TEST_SECRET_1"])
	}

	if found["TEST_SECRET_2"] != "value2" {
		t.Errorf("expected TEST_SECRET_2='value2', got '%s'", found["TEST_SECRET_2"])
	}
}

func TestLoadFromEnvEmpty(t *testing.T) {
	secrets := LoadFromEnv([]string{"DEFINITELY_NOT_SET_12345"})

	if len(secrets) != 0 {
		t.Errorf("expected 0 secrets, got %d", len(secrets))
	}
}

func TestLoadFromEnvEmptyList(t *testing.T) {
	secrets := LoadFromEnv([]string{})

	if len(secrets) != 0 {
		t.Errorf("expected 0 secrets, got %d", len(secrets))
	}
}

func TestNewManager(t *testing.T) {
	// We can't fully test without SSH, but we can verify the constructor
	m := NewManager(nil, "myapp")

	if m.stackName != "myapp" {
		t.Errorf("expected stackName 'myapp', got '%s'", m.stackName)
	}
}
