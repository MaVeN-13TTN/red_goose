package config

import (
    "fmt"
    "os"
    "path/filepath"

    "github.com/spf13/viper"
)

type Config struct {
    Download DownloadConfig `mapstructure:"download"`
    Output   OutputConfig   `mapstructure:"output"`
    Network  NetworkConfig  `mapstructure:"network"`
}

type DownloadConfig struct {
    DefaultQuality string `mapstructure:"default_quality"`
    MaxWorkers     int    `mapstructure:"max_workers"`
    SkipErrors     bool   `mapstructure:"skip_errors"`
    AudioOnly      bool   `mapstructure:"audio_only"`
}

type OutputConfig struct {
    Directory        string `mapstructure:"directory"`
    CreateSubfolders bool   `mapstructure:"create_subfolders"`
    NamingPattern    string `mapstructure:"naming_pattern"`
}

type NetworkConfig struct {
    Timeout       int    `mapstructure:"timeout_seconds"`
    Retries       int    `mapstructure:"retries"`
    UserAgent     string `mapstructure:"user_agent"`
    RateLimit     int    `mapstructure:"rate_limit_ms"`
}

// DefaultConfig returns a Config with default values
func DefaultConfig() *Config {
    return &Config{
        Download: DownloadConfig{
            DefaultQuality: "best",
            MaxWorkers:     3,
            SkipErrors:     false,
            AudioOnly:      false,
        },
        Output: OutputConfig{
            Directory:        "./downloads",
            CreateSubfolders: true,
            NamingPattern:    "{title}",
        },
        Network: NetworkConfig{
            Timeout:   1800, // 30 minutes
            Retries:   3,
            UserAgent: "red-goose/1.0",
            RateLimit: 100,
        },
    }
}

// LoadConfig loads the configuration from the specified file
// If no file is specified, it looks for .red-goose.yaml in the user's home directory
func LoadConfig(cfgFile string) (*Config, error) {
    config := DefaultConfig()

    if cfgFile != "" {
        // Use config file from the flag
        viper.SetConfigFile(cfgFile)
    } else {
        // Find home directory
        home, err := os.UserHomeDir()
        if err != nil {
            return nil, fmt.Errorf("failed to find home directory: %w", err)
        }

        // Search config in home directory with name ".red-goose" (without extension)
        viper.AddConfigPath(home)
        viper.SetConfigType("yaml")
        viper.SetConfigName(".red-goose")
    }

    // Set environment variable prefix
    viper.SetEnvPrefix("REDGOOSE")
    viper.AutomaticEnv() // read in environment variables that match

    // If a config file is found, read it in
    if err := viper.ReadInConfig(); err == nil {
        if err := viper.Unmarshal(config); err != nil {
            return nil, fmt.Errorf("failed to parse config: %w", err)
        }
    } else if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
        // Config file was found but another error was produced
        return nil, fmt.Errorf("failed to read config file: %w", err)
    }

    return config, nil
}

// SaveConfig saves the configuration to the specified file
// If no file is specified, it saves to .red-goose.yaml in the user's home directory
func SaveConfig(config *Config, cfgFile string) error {
    if cfgFile == "" {
        home, err := os.UserHomeDir()
        if err != nil {
            return fmt.Errorf("failed to find home directory: %w", err)
        }
        cfgFile = filepath.Join(home, ".red-goose.yaml")
    }

    // Set config values
    viper.Set("download", config.Download)
    viper.Set("output", config.Output)
    viper.Set("network", config.Network)

    // Write config file
    return viper.WriteConfigAs(cfgFile)
}

// CreateDefaultConfig creates a default configuration file at the specified path
func CreateDefaultConfig(configPath string) error {
    config := DefaultConfig()
    return SaveConfig(config, configPath)
}
