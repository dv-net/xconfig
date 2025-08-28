package loader_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/dv-net/xconfig"
	"github.com/dv-net/xconfig/internal/f"
	"github.com/dv-net/xconfig/plugins/loader"
	"github.com/google/go-cmp/cmp"
)

func TestFiles(t *testing.T) {
	expect := f.Config{
		Anon: f.Anon{
			Version: "0.2",
		},

		GoHard: true,

		Redis: f.Redis{
			Host: "redis-host",
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

	err = l.AddFile("testdata/config_rethink.json", true)
	if err != nil {
		t.Fatalf("failed to add file: %v", err)
	}

	err = l.AddFile("testdata/config_partial.json", true)
	if err != nil {
		t.Fatalf("failed to add file: %v", err)
	}

	value := f.Config{}

	os.Args = os.Args[:1]
	_, err = xconfig.Load(&value, xconfig.WithLoader(l))
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if diff := cmp.Diff(expect, value); diff != "" {
		t.Error(diff)
	}
}
