package config

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

func setupEnv(t *testing.T) {
	t.Helper()
	originalEnv := map[string]string{
		"LOG_LEVEL":                os.Getenv("LOG_LEVEL"),
		"WEBSOCKET_PORT":           os.Getenv("WEBSOCKET_PORT"),
		"MULTICAST_ADDRESS":        os.Getenv("MULTICAST_ADDRESS"),
		"MULTICAST_PORT":           os.Getenv("MULTICAST_PORT"),
		"MULTICAST_INTERFACE_NAME": os.Getenv("MULTICAST_INTERFACE_NAME"),
		"CACHE_PATH":               os.Getenv("CACHE_PATH"),
	}

	t.Cleanup(func() {
		for k, v := range originalEnv {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}
	})
}

func TestConfig_DefaultValues(t *testing.T) {
	setupEnv(t)

	os.Unsetenv("LOG_LEVEL")
	os.Unsetenv("WEBSOCKET_PORT")
	os.Unsetenv("MULTICAST_ADDRESS")
	os.Unsetenv("MULTICAST_PORT")
	os.Unsetenv("MULTICAST_INTERFACE_NAME")
	os.Unsetenv("CACHE_PATH")

	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envFile, []byte(""), 0o644); err != nil {
		t.Fatalf("failed to create .env file: %v", err)
	}

	os.Setenv("LOG_LEVEL", "INFO")
	os.Setenv("WEBSOCKET_PORT", "9000")
	os.Setenv("MULTICAST_ADDRESS", "235.5.5.11")
	os.Setenv("MULTICAST_PORT", "9000")
	os.Setenv("MULTICAST_INTERFACE_NAME", "lo")
	os.Setenv("CACHE_PATH", tmpDir+"/cache/")

	origArgs := os.Args
	os.Args = []string{"test", "-env-file", envFile}
	t.Cleanup(func() { os.Args = origArgs })

	cfg := mustLoadForTest(envFile)

	if cfg.LogLevel != slog.LevelInfo {
		t.Errorf("expected LogLevel INFO, got %v", cfg.LogLevel)
	}
	if cfg.WebSocketPort != 9000 {
		t.Errorf("expected WebSocketPort 9000, got %d", cfg.WebSocketPort)
	}
	if cfg.MulticastAddress != "235.5.5.11" {
		t.Errorf("expected MulticastAddress 235.5.5.11, got %s", cfg.MulticastAddress)
	}
	if cfg.MulticastPort != 9000 {
		t.Errorf("expected MulticastPort 9000, got %d", cfg.MulticastPort)
	}
	if cfg.MulticastInterfaceName != "lo" {
		t.Errorf("expected MulticastInterfaceName 'lo', got %s", cfg.MulticastInterfaceName)
	}
}

func TestConfig_CustomCachePath(t *testing.T) {
	setupEnv(t)

	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envFile, []byte(""), 0o644); err != nil {
		t.Fatalf("failed to create .env file: %v", err)
	}

	customPath := tmpDir + "/custom"
	os.Setenv("CACHE_PATH", customPath)
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("WEBSOCKET_PORT", "8080")
	os.Setenv("MULTICAST_ADDRESS", "235.5.5.11")
	os.Setenv("MULTICAST_PORT", "8080")
	os.Setenv("MULTICAST_INTERFACE_NAME", "lo")

	origArgs := os.Args
	os.Args = []string{"test", "-env-file", envFile}
	t.Cleanup(func() { os.Args = origArgs })

	cfg := mustLoadForTest(envFile)

	expectedPath := customPath + "/"
	if cfg.CachePath != expectedPath {
		t.Errorf("expected CachePath %s, got %s", expectedPath, cfg.CachePath)
	}
}

func TestConfig_EnvFileFlag(t *testing.T) {
	setupEnv(t)

	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	envContent := `LOG_LEVEL=DEBUG
WEBSOCKET_PORT=7070
MULTICAST_ADDRESS=235.5.5.10
MULTICAST_PORT=7070
MULTICAST_INTERFACE_NAME=eth0
CACHE_PATH=` + tmpDir + `/test-cache/`

	if err := os.WriteFile(envFile, []byte(envContent), 0o644); err != nil {
		t.Fatalf("failed to create .env file: %v", err)
	}

	os.Unsetenv("LOG_LEVEL")
	os.Unsetenv("WEBSOCKET_PORT")
	os.Unsetenv("MULTICAST_ADDRESS")
	os.Unsetenv("MULTICAST_PORT")
	os.Unsetenv("MULTICAST_INTERFACE_NAME")
	os.Unsetenv("CACHE_PATH")

	origArgs := os.Args
	os.Args = []string{"test", "-env-file", envFile}
	t.Cleanup(func() { os.Args = origArgs })

	cfg := mustLoadForTest(envFile)

	if cfg.LogLevel != slog.LevelDebug {
		t.Errorf("expected LogLevel DEBUG, got %v", cfg.LogLevel)
	}
	if cfg.WebSocketPort != 7070 {
		t.Errorf("expected WebSocketPort 7070, got %d", cfg.WebSocketPort)
	}
	if cfg.MulticastAddress != "235.5.5.10" {
		t.Errorf("expected MulticastAddress 235.5.5.10, got %s", cfg.MulticastAddress)
	}
}

func mustLoadForTest(envFile string) *Config {
	cfg := &Config{}

	err := godotenv.Load(envFile)
	if err != nil {
		panic(err.Error())
	}

	err = cleanenv.ReadEnv(cfg)
	if err != nil {
		panic(err.Error())
	}

	if cfg.CachePath == "" {
		var err error
		cfg.CachePath, err = os.UserCacheDir()
		if err != nil {
			panic(err)
		}
		cfg.CachePath += "/p2p/"
	} else {
		if cfg.CachePath[len(cfg.CachePath)-1] != '/' {
			cfg.CachePath += "/"
		}
	}

	_, stat := os.Stat(cfg.CachePath)
	if os.IsNotExist(stat) {
		if err := os.MkdirAll(cfg.CachePath, 0o755); err != nil {
			panic(err)
		}
	}

	cfg.LogFile = cfg.CachePath + "log.log"
	cfg.IDFile = cfg.CachePath + "id"

	return cfg
}
