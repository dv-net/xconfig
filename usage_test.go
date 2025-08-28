package xconfig_test

import (
	"fmt"
	"testing"

	"github.com/dv-net/xconfig"
	"github.com/dv-net/xconfig/flat"
	"github.com/dv-net/xconfig/internal/f"
	"github.com/dv-net/xconfig/plugins"
	"github.com/dv-net/xconfig/plugins/secret"
	"github.com/google/go-cmp/cmp"
)

const expectedUsageMessage = `
Supported Fields:
FIELD                   FLAG                     ENV                     DEFAULT    GOODPLUGIN              SECRET    USAGE
-----                   -----                    -----                   -------    ----------              ------    -----
Version                 -version                 VERSION                            Version                           
GoHard                  -gohard                  GO_HARD                 false      GoHard                            
Redis.Host              -redis-host              REDIS_HOST                         Redis.Host                        
Redis.Port              -redis-port              REDIS_PORT              0          Redis.Port                        
Rethink.Host.Address    -rethink-host-address    RETHINK_HOST_ADDRESS               Rethink.Host.Address              
Rethink.Host.Port       -rethink-host-port       RETHINK_HOST_PORT                  Rethink.Host.Port                 
Rethink.Db              -rethink-db              RETHINK_DB              primary    Rethink.Db                        main database used by our application
Rethink.Password        -rethink-password        RETHINK_PASSWORD                   Rethink.Password        ✅         
BaseURL.API             -baseurl-api             BASE_URL_API                       BaseURL.API                       
`

type UselessPluginVisitor struct {
	plugins.Plugin
}

func (*UselessPluginVisitor) Parse() error { return nil }

func (*UselessPluginVisitor) Visit(fields flat.Fields) error {
	for _, f := range fields {
		f.Meta()["goodplugin"] = f.Name()
	}
	return nil
}

func TestUsage(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Usage should not panic, but did: %v", r)
		}
	}()

	// good plugin is used just so that we have more than
	// one tag/field that isn't pre-weighted in "usage".
	noopPlugin := &UselessPluginVisitor{}

	value := f.Config{}

	secretProvider := func(name string) (string, error) { return "top secret token", nil }

	c, err := xconfig.Load(&value, xconfig.WithPlugins(secret.New(secretProvider), noopPlugin))
	if err != nil {
		fmt.Println("Error loading config:", err)
	}

	output, err := c.Usage()
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(output)

	if diff := cmp.Diff(expectedUsageMessage, output); diff != "" {
		t.Error(diff)
	}
}
