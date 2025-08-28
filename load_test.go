package xconfig_test

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/dv-net/xconfig"
	"github.com/dv-net/xconfig/internal/f"
	"github.com/dv-net/xconfig/plugins/loader"
	"github.com/dv-net/xconfig/plugins/secret"
	"github.com/google/go-cmp/cmp"
)

func TestMain(m *testing.M) {
	// for go test framework.
	flag.Parse()

	os.Exit(m.Run())
}

func TestClassicBasic(t *testing.T) {
	expect := f.Config{
		Anon: f.Anon{
			Version: "from-flags",
		},

		GoHard: true,

		Redis: f.Redis{
			Host: "from-envs",
			Port: 6379,
		},

		Rethink: f.RethinkConfig{
			Host: f.Host{
				Address: "rethink-cluster",
				Port:    "28015",
			},
			Db: "base",
		},
	}

	l, err := loader.NewLoader(map[string]loader.Unmarshal{
		".json": json.Unmarshal,
	})
	if err != nil {
		t.Fatalf("failed to create loader: %v", err)
	}

	l.AddFile("testdata/classic.json", true)

	value := f.Config{}

	// set some env vars to test env var and plugin orders.
	os.Setenv("VERSION", "bad-value-overrided-with-flags")
	os.Setenv("REDIS_HOST", "from-envs")

	defer os.Unsetenv("VERSION")
	defer os.Unsetenv("REDIS_HOST")

	// patch the os.Args. for our tests.
	os.Args = append(os.Args[:1], "-version=from-flags")

	_, err = xconfig.Load(&value, xconfig.WithLoader(l))
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if diff := cmp.Diff(expect, value); diff != "" {
		t.Error(diff)
	}
}

func TestClassicWithSecret(t *testing.T) {
	// Config is part of text fixtures.
	type Creds struct {
		APIKey   string `secret:""`
		APIToken string `secret:"API_TOKEN"`
	}

	type Config struct {
		Redis   f.Redis
		Rethink f.RethinkConfig
		Creds   Creds
	}
	expect := Config{
		Redis: f.Redis{
			Host: "redis-host",
			Port: 6379,
		},

		Rethink: f.RethinkConfig{
			Host: f.Host{
				Address: "rethink-cluster",
				Port:    "28015",
			},
			Db:       "base",
			Password: "top secret token",
		},

		Creds: Creds{
			APIKey:   "top secret token",
			APIToken: "top secret token",
		},
	}

	l, err := loader.NewLoader(map[string]loader.Unmarshal{
		".json": json.Unmarshal,
	})
	if err != nil {
		t.Fatalf("failed to create loader: %v", err)
	}

	err = l.AddFile("testdata/classic.json", true)
	if err != nil {
		t.Fatalf("failed to add file: %v", err)
	}

	value := Config{}

	SecretProvider := func(name string) (string, error) {
		// known secrets.
		if name == "API_TOKEN" || name == "RETHINK_PASSWORD" || name == "CREDS_APIKEY" {
			return "top secret token", nil
		}

		return "", fmt.Errorf("Secret not found %s", name)
	}

	// patch the os.Args. for our tests.
	os.Args = os.Args[:1]
	os.Unsetenv("REDIS_HOST")

	_, err = xconfig.Load(&value, xconfig.WithLoader(l), xconfig.WithPlugins(secret.New(SecretProvider)))
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if diff := cmp.Diff(expect, value); diff != "" {
		t.Error(diff)
	}
}

func TestClassicBadPlugin(t *testing.T) {
	var badPlugin BadPlugin

	config := f.Config{}

	_, err := xconfig.Load(&config, xconfig.WithPlugins(badPlugin))

	if err == nil {
		t.Error("expected error for bad plugin, got nil")
	}

	if err.Error() != "unsupported plugins. expecting a Walker or Visitor" {
		t.Errorf("Expected unsupported plugin error, got: %v", err)
	}
}
