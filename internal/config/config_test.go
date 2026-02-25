package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	t.Setenv("DB_USER", "testuser")
	t.Setenv("DB_PASSWORD", "testpass")
	t.Setenv("HTTP_SERVER_ADDRESS", ":9090")

	cfg := LoadConfig("")

	require.Equal(t, "testuser", cfg.Db.User)
	require.Equal(t, ":9090", cfg.HTTPServer.Address)
}

func TestLoadConfig_Defaults(t *testing.T) {
	cfg := LoadConfig("")

	require.Equal(t, ":8080", cfg.HTTPServer.Address)
	require.Equal(t, 5*time.Second, cfg.HTTPServer.ReadTimeout)
	require.Equal(t, int32(100), cfg.Db.MaxConns)
}

func TestLoadConfig_FromFile(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), ".env")
	require.NoError(t, err)

	_, err = f.WriteString("DB_USER=fileuser\nDB_PASSWORD=filepass\n")
	require.NoError(t, err)

	require.NoError(t, f.Close())

	cfg := LoadConfig(f.Name())
	require.Equal(t, "fileuser", cfg.Db.User)
}

func TestParseEnvValue(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		got, err := parseEnvValue[int]("PORT", "8080")
		require.NoError(t, err)
		require.Equal(t, 8080, got)
	})

	t.Run("int invalid", func(t *testing.T) {
		_, err := parseEnvValue[int]("PORT", "abc")
		require.Error(t, err)
	})

	t.Run("duration", func(t *testing.T) {
		got, err := parseEnvValue[time.Duration]("TIMEOUT", "5s")
		require.NoError(t, err)
		require.Equal(t, 5*time.Second, got)
	})

	t.Run("bool", func(t *testing.T) {
		got, err := parseEnvValue[bool]("FLAG", "true")
		require.NoError(t, err)
		assert.True(t, got)
	})

	t.Run("int32", func(t *testing.T) {
		got, err := parseEnvValue[int32]("CONNS", "100")
		require.NoError(t, err)
		require.Equal(t, int32(100), got)
	})
}
