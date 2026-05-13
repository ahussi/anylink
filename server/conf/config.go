package conf

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/spf13/viper"
)

// AppConfig holds the global application configuration
var (
	AppConfig *Config
	configMu  sync.RWMutex
)

// Config represents the main application configuration structure
type Config struct {
	// Server settings
	Server ServerConfig `mapstructure:"server" json:"server"`

	// Database settings
	DB DBConfig `mapstructure:"db" json:"db"`

	// VPN settings
	VPN VPNConfig `mapstructure:"vpn" json:"vpn"`

	// JWT settings
	JWT JWTConfig `mapstructure:"jwt" json:"jwt"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Addr     string `mapstructure:"addr" json:"addr"`
	HTTPPort int    `mapstructure:"http_port" json:"http_port"`
	HTTPSPort int   `mapstructure:"https_port" json:"https_port"`
	CertFile string `mapstructure:"cert_file" json:"cert_file"`
	KeyFile  string `mapstructure:"key_file" json:"key_file"`
	LogLevel string `mapstructure:"log_level" json:"log_level"`
}

// DBConfig holds database configuration
type DBConfig struct {
	Type string `mapstructure:"type" json:"type"`
	DSN  string `mapstructure:"dsn" json:"dsn"`
}

// VPNConfig holds VPN tunnel configuration
type VPNConfig struct {
	ServerAddress string `mapstructure:"server_address" json:"server_address"`
	IPPool        string `mapstructure:"ip_pool" json:"ip_pool"`
	IPMask        string `mapstructure:"ip_mask" json:"ip_mask"`
	DNS           []string `mapstructure:"dns" json:"dns"`
	Routes        []string `mapstructure:"routes" json:"routes"`
	SplitTunnel   bool   `mapstructure:"split_tunnel" json:"split_tunnel"`
	MTU           int    `mapstructure:"mtu" json:"mtu"`
}

// JWTConfig holds JWT authentication configuration
type JWTConfig struct {
	Secret     string `mapstructure:"secret" json:"secret"`
	ExpireTime int    `mapstructure:"expire_time" json:"expire_time"` // in hours
}

// InitConfig initializes the application configuration from the given file path
func InitConfig(cfgFile string) error {
	v := viper.New()

	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("toml")
		v.AddConfigPath(".")
		v.AddConfigPath("./conf")
		v.AddConfigPath("/etc/anylink/")
	}

	// Set default values
	setDefaults(v)

	// Read environment variables
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(*os.PathError); !ok {
			return fmt.Errorf("failed to read config file: %w", err)
		}
		fmt.Println("No config file found, using defaults and environment variables")
	}

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	configMu.Lock()
	AppConfig = cfg
	configMu.Unlock()

	return nil
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	v.SetDefault("server.addr", "0.0.0.0")
	v.SetDefault("server.http_port", 8080)
	v.SetDefault("server.https_port", 443)
	v.SetDefault("server.log_level", "info")

	v.SetDefault("db.type", "sqlite3")
	v.SetDefault("db.dsn", "./anylink.db")

	v.SetDefault("vpn.ip_pool", "192.168.10.0")
	v.SetDefault("vpn.ip_mask", "255.255.255.0")
	v.SetDefault("vpn.dns", []string{"8.8.8.8", "8.8.4.4"})
	v.SetDefault("vpn.split_tunnel", false)
	v.SetDefault("vpn.mtu", 1400)

	v.SetDefault("jwt.expire_time", 24)
}

// GetConfig returns a thread-safe copy of the current configuration
func GetConfig() *Config {
	configMu.RLock()
	defer configMu.RUnlock()
	return AppConfig
}

// String returns a JSON representation of the config (with sensitive fields masked)
func (c *Config) String() string {
	type masked struct {
		Server ServerConfig `json:"server"`
		DB     struct {
			Type string `json:"type"`
			DSN  string `json:"dsn"`
		} `json:"db"`
		VPN VPNConfig `json:"vpn"`
	}
	m := masked{}
	m.Server = c.Server
	m.DB.Type = c.DB.Type
	m.DB.DSN = "***"
	m.VPN = c.VPN

	b, _ := json.MarshalIndent(m, "", "  ")
	return string(b)
}
