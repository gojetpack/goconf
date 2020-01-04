package goconf

import (
	"fmt"
	"github.com/joho/godotenv"
	"reflect"
	"testing"
)

const testEnv string = "testdata/test.env"

type Env string

const (
	Test        = Env("test")
	Development = Env("development")
	Staging     = Env("staging")
	Production  = Env("production")
)

type serverConfig struct {
	GRPCPort     int  `env:"GRPC_PORT"`
	HTTPPort     int  `env:"HTTP_PORT"`
	RunHTTPProxy bool `env:"RUN_HTTP_PROXY"`
	Environment  Env
}

var cfg = serverConfig{
	GRPCPort:     3000,
	HTTPPort:     3001,
	RunHTTPProxy: true,
	Environment:  "DEFAULT_VALUE",
}

var cfgWant = serverConfig{
	GRPCPort:     5000,
	HTTPPort:     3001,
	RunHTTPProxy: true,
	Environment:  "DEVELOPMENT",
}

var opts = ExtractorOptions{
	OmitNotTagged: false,
}

func TestExtract(t *testing.T) {
	tests := []struct {
		name    string
		envFile string
		args    ExtractorArgs
		want    []interface{}
		wantErr bool
	}{
		{
			name:    "Success",
			envFile: testEnv,
			args: ExtractorArgs{
				Options: opts,
				Configs: []interface{}{&cfg, ""},
			},
			want:    []interface{}{&cfgWant, ""},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := godotenv.Load(tt.envFile)
			if err != nil {
				t.Errorf("Extract() invalid test env file error = %v", err)
			}
			err = Extract(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Extract() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i := 0; i < len(tt.args.Configs); i++ {
				if !reflect.DeepEqual(tt.args.Configs[i], tt.want[i]) {
					t.Errorf("Extract() = %v, want %v", tt.args.Configs, tt.want)
				}
			}

		})
	}
}

/**************************************/
/************ EXAMPLES ****************/
/**************************************/

func ExampleExtract() {
	type CustomAlias string

	type ConfigExample struct {
		RegularConfigField   int         `env:"CRAZY_CONFIG_FIELD"`
		FieldWithoutTag      string      // linked to  "FIELD_WITHOUT_TAG" env var
		FieldWithChangedTag  bool        `env:"OTHER_CRAZY_NAME"`
		CustomAliasTypeField CustomAlias `env:"CUSTOM_ALIAS_TYPE_FIELD"`
	}
	prefix := ""
	config := ConfigExample{}
	args := ExtractorArgs{
		Options: ExtractorOptions{
			EnvFile:       "testdata/env_test",
			OmitNotTagged: false,
		},
		Configs: []interface{}{&config, prefix},
	}
	err := Extract(args)
	if err != nil {
		fmt.Print(err)
		return
	}
	fmt.Print(config.CustomAliasTypeField)
	// Output: Garfield: the mad cat
}

func ExampleExtractWithPrefix() {
	type RedisConfig struct {
		Host             string
		Port             int
		AuthToken        string
		SecureConnection bool `env:"SECURE_CONNECTION"`
	}
	redisConfig := RedisConfig{}

	type MongoConfig struct {
		Host             string
		Port             int
		DatabaseName     string
		Test             bool
		SecureConnection bool `env:"SECURE_CONNECTION"`
	}
	mongoConfig := MongoConfig{}

	args := ExtractorArgs{
		Options: ExtractorOptions{
			EnvFile: "testdata/env_with_prefix",
		},
		Configs: []interface{}{
			&redisConfig, "REDIS",
			&mongoConfig, "MONGO",
		},
	}
	err := Extract(args)
	if err != nil {
		fmt.Print(err)
		return
	}
	fmt.Print(redisConfig.Host)
	// Output: 127.0.0.1
}
