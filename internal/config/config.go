package config

import (
	"fmt"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Config holds the complete application configuration
type Config struct {
	IMAP     IMAPConfig     `yaml:"imap"`
	Database DatabaseConfig `yaml:"database"`
	Web      WebConfig      `yaml:"web"`
	Sync     SyncConfig     `yaml:"sync"`
	Logging  LogConfig      `yaml:"logging"`
}

// IMAPConfig contains IMAP server connection settings
type IMAPConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Folder   string `yaml:"folder"`
	UseTLS   bool   `yaml:"use_tls"`
}

// DatabaseConfig contains database settings
type DatabaseConfig struct {
	Path string `yaml:"path"`
}

// WebConfig contains web server settings
type WebConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// SyncConfig contains sync schedule settings
type SyncConfig struct {
	Interval  string `yaml:"interval"` // e.g., "15m"
	OnStartup bool   `yaml:"on_startup"`
}

// LogConfig contains logging settings
type LogConfig struct {
	Level  string `yaml:"level"`  // debug, info, warn, error
	Format string `yaml:"format"` // json, text
}

// Load reads configuration from YAML file, environment variables, and CLI flags
// Priority order: CLI flags > Environment variables > YAML file
func Load(configFile string) (*Config, error) {
	v := viper.New()

	// Set default values
	setDefaults(v)

	// Read from config file if provided
	if configFile != "" {
		v.SetConfigFile(configFile)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Read from environment variables with DMARC_ prefix
	v.SetEnvPrefix("DMARC")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Unmarshal into Config struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate required fields
	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

// LoadWithFlags reads configuration with CLI flag overrides
func LoadWithFlags() (*Config, error) {
	// Define CLI flags
	configFile := pflag.String("config", "config.yaml", "Path to config file")
	imapHost := pflag.String("imap-host", "", "IMAP server host")
	imapPort := pflag.Int("imap-port", 0, "IMAP server port")
	imapUsername := pflag.String("imap-username", "", "IMAP username")
	imapPassword := pflag.String("imap-password", "", "IMAP password")
	imapFolder := pflag.String("imap-folder", "", "IMAP folder")
	imapUseTLS := pflag.Bool("imap-use-tls", true, "Use TLS for IMAP connection")
	databasePath := pflag.String("database", "", "Database file path")
	webHost := pflag.String("web-host", "", "Web server host")
	webPort := pflag.Int("web-port", 0, "Web server port")
	syncInterval := pflag.String("sync-interval", "", "Sync interval (e.g., 15m)")
	syncOnStartup := pflag.Bool("sync-on-startup", false, "Run sync on startup")
	logLevel := pflag.String("log-level", "", "Log level (debug, info, warn, error)")
	logFormat := pflag.String("log-format", "", "Log format (json, text)")

	pflag.Parse()

	v := viper.New()

	// Set default values
	setDefaults(v)

	// Read from config file
	if *configFile != "" {
		v.SetConfigFile(*configFile)
		// Ignore error if config file doesn't exist
		_ = v.ReadInConfig()
	}

	// Read from environment variables
	v.SetEnvPrefix("DMARC")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Override with CLI flags (highest priority)
	if pflag.Lookup("imap-host").Changed {
		v.Set("imap.host", *imapHost)
	}
	if pflag.Lookup("imap-port").Changed {
		v.Set("imap.port", *imapPort)
	}
	if pflag.Lookup("imap-username").Changed {
		v.Set("imap.username", *imapUsername)
	}
	if pflag.Lookup("imap-password").Changed {
		v.Set("imap.password", *imapPassword)
	}
	if pflag.Lookup("imap-folder").Changed {
		v.Set("imap.folder", *imapFolder)
	}
	if pflag.Lookup("imap-use-tls").Changed {
		v.Set("imap.use_tls", *imapUseTLS)
	}
	if pflag.Lookup("database").Changed {
		v.Set("database.path", *databasePath)
	}
	if pflag.Lookup("web-host").Changed {
		v.Set("web.host", *webHost)
	}
	if pflag.Lookup("web-port").Changed {
		v.Set("web.port", *webPort)
	}
	if pflag.Lookup("sync-interval").Changed {
		v.Set("sync.interval", *syncInterval)
	}
	if pflag.Lookup("sync-on-startup").Changed {
		v.Set("sync.on_startup", *syncOnStartup)
	}
	if pflag.Lookup("log-level").Changed {
		v.Set("logging.level", *logLevel)
	}
	if pflag.Lookup("log-format").Changed {
		v.Set("logging.format", *logFormat)
	}

	// Unmarshal into Config struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// IMAP defaults
	v.SetDefault("imap.port", 993)
	v.SetDefault("imap.folder", "INBOX")
	v.SetDefault("imap.use_tls", true)

	// Database defaults
	v.SetDefault("database.path", "./dmarc-reports.db")

	// Web defaults
	v.SetDefault("web.host", "localhost")
	v.SetDefault("web.port", 8080)

	// Sync defaults
	v.SetDefault("sync.interval", "15m")
	v.SetDefault("sync.on_startup", true)

	// Logging defaults
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "text")
}

// validate checks that required configuration fields are set
func validate(cfg *Config) error {
	if cfg.IMAP.Host == "" {
		return fmt.Errorf("imap.host is required")
	}
	if cfg.IMAP.Username == "" {
		return fmt.Errorf("imap.username is required")
	}
	if cfg.IMAP.Password == "" {
		return fmt.Errorf("imap.password is required")
	}
	if cfg.Database.Path == "" {
		return fmt.Errorf("database.path is required")
	}

	// Validate log level
	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[cfg.Logging.Level] {
		return fmt.Errorf("invalid log level: %s (must be debug, info, warn, or error)", cfg.Logging.Level)
	}

	// Validate log format
	validFormats := map[string]bool{"json": true, "text": true}
	if !validFormats[cfg.Logging.Format] {
		return fmt.Errorf("invalid log format: %s (must be json or text)", cfg.Logging.Format)
	}

	return nil
}
