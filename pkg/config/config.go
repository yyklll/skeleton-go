package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/yyklll/skeleton/pkg/version"
)

type AllConfig struct {
	SrvConfig ServerConfig
	Db        Database
}

type ServerConfig struct {
	HttpClientTimeout         time.Duration `mapstructure:"http-client-timeout"`
	HttpServerTimeout         time.Duration `mapstructure:"http-server-timeout"`
	HttpServerShutdownTimeout time.Duration `mapstructure:"http-server-shutdown-timeout"`
	ConfigPath                string        `mapstructure:"config-path"`
	Port                      int           `mapstructure:"port"`
	PortMetrics               int           `mapstructure:"port-metrics"`
	Hostname                  string        `mapstructure:"hostname"`
	H2C                       bool          `mapstructure:"h2c"`
	JWTSecret                 string        `mapstructure:"jwt-secret"`
	LogLevel                  string        `mapstructure:"log-level"`
}

// Database backend configuration
type Database struct {
	Enable  bool   `mapstructure:"enable"`
	Backend string `mapstructure:"backend"`
	Source  string `mapstructure:"source"`
}

// ParseSQLStorage tries to parse out Database backend from Viper.  If backend and
// URL are not provided, returns a nil pointer.  Database is required (if
// a backend is not provided, an error will be returned.)
func ParseSQLStorage(configuration *viper.Viper) (*Database, error) {
	store := Database{
		Backend: configuration.GetString("database.backend"),
		Source:  configuration.GetString("database.db_url"),
	}

	switch {
	case store.Source == "":
		return nil, fmt.Errorf(
			"must provide a non-empty database source for %s",
			store.Backend,
		)
	case store.Backend == "MySQL":
		urlConfig, err := mysql.ParseDSN(store.Source)
		if err != nil {
			return nil, fmt.Errorf("failed to parse the database source for %s",
				store.Backend,
			)
		}

		urlConfig.ParseTime = true
		store.Source = urlConfig.FormatDSN()
	default:
		return nil, fmt.Errorf(
			"%s is not a supported SQL backend driver",
			store.Backend,
		)
	}
	return &store, nil
}

// SetupViper sets up an instance of viper to also look at environment
// variables
func SetupConfig(envPrefix string, fs *pflag.FlagSet) {
	viper.BindPFlags(fs)
	hostname, _ := os.Hostname()
	viper.SetDefault("jwt-secret", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9")
	viper.Set("hostname", hostname)
	viper.Set("version", version.VERSION)
	viper.Set("revision", version.REVISION)
	viper.SetEnvPrefix(envPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
}

// ParseViper tries to parse out a Viper from a configuration file.
func LoadConfig() error {
	if _, err := os.Stat(filepath.Join(viper.GetString("config-path"), viper.GetString("config"))); err == nil {
		viper.SetConfigName(strings.Split(viper.GetString("config"), ".")[0])
		viper.AddConfigPath(viper.GetString("config-path"))
		if err := viper.ReadInConfig(); err != nil {
			fmt.Printf("Error reading config file, %v\n", err)
		}
	}
	return nil
}

func GetConfigField(key string) interface{} {
	return viper.Get(key)
}
