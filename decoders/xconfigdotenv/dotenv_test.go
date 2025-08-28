package xconfigdotenv_test

import (
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/dv-net/xconfig/decoders/xconfigdotenv"
	"github.com/stretchr/testify/assert"
)

func TestDecoderFormat(t *testing.T) {
	decoder := xconfigdotenv.New()
	assert.Equal(t, "env", decoder.Format())
}

func TestDecoderUnmarshal(t *testing.T) {
	decoder := xconfigdotenv.New()

	tests := []struct {
		name     string
		data     []byte
		expected map[string]string
		wantErr  bool
	}{
		{
			name:     "simple key-value",
			data:     []byte("KEY=value\nANOTHER_KEY=another_value"),
			expected: map[string]string{"KEY": "value", "ANOTHER_KEY": "another_value"},
			wantErr:  false,
		},
		{
			name:     "empty value",
			data:     []byte("KEY="),
			expected: map[string]string{"KEY": ""},
			wantErr:  false,
		},
		{
			name:     "quoted value",
			data:     []byte(`KEY="quoted value"`),
			expected: map[string]string{"KEY": "quoted value"},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result map[string]string
			err := decoder.Unmarshal(tt.data, &result)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestNewDecoder(t *testing.T) {
	decoder := xconfigdotenv.New()
	assert.NotNil(t, decoder)
	assert.IsType(t, &xconfigdotenv.Decoder{}, decoder)
}

// Complex nested structure with different naming conventions and types
type baseURLs struct {
	API             string            // regular camelCase field
	Version_number  int               // field with underscore
	Endpoints       map[string]string // map field
	EnabledFeatures []string          // slice field
}

type S3Config struct {
	API         string  // starts with uppercase
	rateLimit   float64 // starts with lowercase
	MaxUploadMB int     // mixed case field
	IsActive    bool    // boolean field
	REGION_CODE string  // all caps with underscore
	timeout_sec int     // lowercase with underscore
}

type OAuth2Settings struct {
	Client_id     string
	CLIENT_SECRET string
	Scopes        []string
	ExpiresIn     int
	token_type    string
}

type DB_CONFIG struct {
	HOST        string
	port        int
	User_name   string
	Pass_WORD   string
	MaxPoolSize int
	readOnly    bool
}

type nestedConfig struct {
	boolVar    bool
	IntVar     int
	StringVar  string
	FloatVar   float64
	Complex128 complex128
	PtrValue   *string
	Time       time.Duration
}

type testConfig struct {
	BaseURL    baseURLs
	S3         S3Config
	Nested     nestedConfig
	DB         DB_CONFIG
	oauth2     OAuth2Settings
	Meta_DATA  map[string]interface{}
	ENV_MODE   string
	debugLevel int
	Test1      string
	TEST2      string
	test_3     string
	Test_4     string
}

var testDotEnvData = []byte(`
BASE_URL_API=http://example.com/api
BASE_URL_VERSION_NUMBER=2
BASE_URL_ENDPOINTS_AUTH=/auth
BASE_URL_ENDPOINTS_USER=/user
BASE_URL_ENABLED_FEATURES_0=login
BASE_URL_ENABLED_FEATURES_1=register
BASE_URL_ENABLED_FEATURES_2=oauth

S3_API=http://example.com/s3
S3_RATE_LIMIT=5.5
S3_MAX_UPLOAD_MB=100
S3_IS_ACTIVE=true
S3_REGION_CODE=us-east-1
S3_TIMEOUT_SEC=30

OAUTH2_CLIENT_ID=client123
OAUTH2_CLIENT_SECRET=secret456
OAUTH2_SCOPES_0=read
OAUTH2_SCOPES_1=write
OAUTH2_EXPIRES_IN=3600
OAUTH2_TOKEN_TYPE=bearer

DB_CONFIG_HOST=localhost
DB_CONFIG_PORT=5432
DB_CONFIG_USER_NAME=admin
DB_CONFIG_PASS_WORD=secure123
DB_CONFIG_MAX_POOL_SIZE=20
DB_CONFIG_READ_ONLY=false

NESTED_BOOL_VAR=true
NESTED_INT_VAR=42
NESTED_STRING_VAR=Hello, World!
NESTED_FLOAT_VAR=3.14
NESTED_COMPLEX128=(1+2i)
NESTED_PTR_VALUE=pointer-value
NESTED_TIME=5s

META_DATA_VERSION=1.0
META_DATA_BUILD=20250604
META_DATA_AUTHOR=GoTeam

ENV_MODE=production
DEBUGLEVEL=1
TEST1=CamelCase
TEST2=ALLCAPS
TEST_3=snake_case
TEST_4=Mixed_Snake_Case
`)

func TestDecoderUnmarshalToStruct(t *testing.T) {
	decoder := xconfigdotenv.New()

	var config testConfig
	err := decoder.Unmarshal(testDotEnvData, &config)
	assert.NoError(t, err)

	spew.Dump(config) // Debugging output to see the structure after unmarshalling

	// BaseURL structure assertions
	assert.Equal(t, "http://example.com/api", config.BaseURL.API)
	assert.Equal(t, 2, config.BaseURL.Version_number)
	assert.Equal(t, "/auth", config.BaseURL.Endpoints["AUTH"])
	assert.Equal(t, "/user", config.BaseURL.Endpoints["USER"])
	assert.Equal(t, 3, len(config.BaseURL.EnabledFeatures))
	assert.Equal(t, "login", config.BaseURL.EnabledFeatures[0])
	assert.Equal(t, "register", config.BaseURL.EnabledFeatures[1])
	assert.Equal(t, "oauth", config.BaseURL.EnabledFeatures[2])

	// S3 structure assertions
	assert.Equal(t, "http://example.com/s3", config.S3.API)
	assert.Equal(t, 5.5, config.S3.rateLimit)
	assert.Equal(t, 100, config.S3.MaxUploadMB)
	assert.True(t, config.S3.IsActive)
	assert.Equal(t, "us-east-1", config.S3.REGION_CODE)
	assert.Equal(t, 30, config.S3.timeout_sec)

	// OAuth2 structure assertions
	assert.Equal(t, "client123", config.oauth2.Client_id)
	assert.Equal(t, "secret456", config.oauth2.CLIENT_SECRET)
	assert.Equal(t, 2, len(config.oauth2.Scopes))
	assert.Equal(t, "read", config.oauth2.Scopes[0])
	assert.Equal(t, "write", config.oauth2.Scopes[1])
	assert.Equal(t, 3600, config.oauth2.ExpiresIn)
	assert.Equal(t, "bearer", config.oauth2.token_type)

	// DB_CONFIG structure assertions
	assert.Equal(t, "localhost", config.DB.HOST)
	assert.Equal(t, 5432, config.DB.port)
	assert.Equal(t, "admin", config.DB.User_name)
	assert.Equal(t, "secure123", config.DB.Pass_WORD)
	assert.Equal(t, 20, config.DB.MaxPoolSize)
	assert.False(t, config.DB.readOnly)

	// Nested structure assertions
	assert.True(t, config.Nested.boolVar)
	assert.Equal(t, 42, config.Nested.IntVar)
	assert.Equal(t, "Hello, World!", config.Nested.StringVar)
	assert.Equal(t, 3.14, config.Nested.FloatVar)
	assert.Equal(t, complex(1, 2), config.Nested.Complex128)
	assert.Equal(t, "pointer-value", *config.Nested.PtrValue)
	assert.Equal(t, 5*time.Second, config.Nested.Time)

	// Other fields
	assert.Equal(t, "1.0", config.Meta_DATA["VERSION"])
	assert.Equal(t, "20250604", config.Meta_DATA["BUILD"])
	assert.Equal(t, "GoTeam", config.Meta_DATA["AUTHOR"])
	assert.Equal(t, "production", config.ENV_MODE)
	assert.Equal(t, 1, config.debugLevel)
	assert.Equal(t, "CamelCase", config.Test1)
	assert.Equal(t, "ALLCAPS", config.TEST2)
	assert.Equal(t, "snake_case", config.test_3)
	assert.Equal(t, "Mixed_Snake_Case", config.Test_4)
}
