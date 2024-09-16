package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	GeneralSettings GeneralSettings  `yaml:"generalSettings"`
	Providers       []ProviderConfig `yaml:"providers"`
}

type GeneralSettings struct {
	UpdateInterval            int    `yaml:"updateInterval"`
	ConnectivityCheckInterval int    `yaml:"connectivityCheckInterval"`
	ConnectivityCheckIP       string `yaml:"connectivityCheckIP"`
	ConnectivityCheckPort     string `yaml:"connectivityCheckPort"`
}

type ProviderConfig struct {
	Type     string                 `yaml:"type"`
	Settings map[string]interface{} `yaml:"settings"`
	Records  []DNSRecord            `yaml:"records"`
}

type DNSRecord struct {
	Name        string `yaml:"name"`
	Type        string `yaml:"type"`
	Proxied     bool   `yaml:"proxied,omitempty"`
	TTL         int    `yaml:"ttl"`
	UpdateToken string `yaml:"updateToken,omitempty"`
}

func LoadConfig() (*Config, error) {
	configPath := getConfigFilePath()

	if configPath == "" {
		return nil, fmt.Errorf("config file not found")
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}

	var config Config
	if err := yaml.Unmarshal(content, &config); err != nil {
		return nil, fmt.Errorf("error unmarshalling yaml: %v", err)
	}

	// Validate general settings with default values
	if config.GeneralSettings.UpdateInterval <= 0 {
		config.GeneralSettings.UpdateInterval = 300
	}
	if config.GeneralSettings.ConnectivityCheckInterval <= 0 {
		config.GeneralSettings.ConnectivityCheckInterval = 10
	}
	if config.GeneralSettings.ConnectivityCheckIP == "" {
		config.GeneralSettings.ConnectivityCheckIP = "8.8.8.8"
	}
	if config.GeneralSettings.ConnectivityCheckPort == "" {
		config.GeneralSettings.ConnectivityCheckPort = "53"
	}

	// Validate providers
	for _, provider := range config.Providers {
		switch provider.Type {
		case "cloudflare":
			settings := provider.Settings
			_, hasAPIToken := settings["apiToken"]
			_, hasEmail := settings["email"]
			_, hasGlobalAPIKey := settings["globalApiKey"]
			if !hasAPIToken && (!hasEmail || !hasGlobalAPIKey) {
				return nil, fmt.Errorf("cloudflare provider requires either apiToken or both email and globalApiKey")
			}
		case "route53":
			settings := provider.Settings
			_, hasAccessKeyID := settings["accessKeyId"]
			_, hasSecretAccessKey := settings["secretAccessKey"]
			if !hasAccessKeyID || !hasSecretAccessKey {
				return nil, fmt.Errorf("route53 provider requires both accessKeyId and secretAccessKey")
			}
		case "digitalocean":
			settings := provider.Settings
			_, hasAPIToken := settings["apiToken"]
			_, hasDomain := settings["domain"]
			if !hasAPIToken || !hasDomain {
				return nil, fmt.Errorf("digitalocean provider requires both apiToken and domain")
			}
		case "clouddns":
			settings := provider.Settings
			_, hasProjectID := settings["projectId"]
			_, hasCredentialsJSONPath := settings["credentialsJsonPath"]
			_, hasZone := settings["zone"]
			if !hasProjectID || !hasCredentialsJSONPath || !hasZone {
				return nil, fmt.Errorf("clouddns provider requires projectId, credentialsJsonPath, and zone")
			}
		case "duckdns":
			settings := provider.Settings
			_, hasToken := settings["token"]
			if !hasToken {
				return nil, fmt.Errorf("duckdns provider requires token")
			}
		case "noip":
			settings := provider.Settings
			_, hasUsername := settings["username"]
			_, hasPassword := settings["password"]
			if !hasUsername || !hasPassword {
				return nil, fmt.Errorf("noip provider requires username and password")
			}
		case "freedns":
			for _, record := range provider.Records {
				if record.UpdateToken == "" {
					return nil, fmt.Errorf("freedns provider requires updateToken per record")
				}
			}
		case "dynu":
			settings := provider.Settings
			_, hasUsername := settings["username"]
			_, hasPassword := settings["password"]
			if !hasUsername || !hasPassword {
				return nil, fmt.Errorf("dynu provider requires username and password")
			}
		default:
			return nil, fmt.Errorf("unsupported provider type: %s", provider.Type)
		}
	}

	return &config, nil
}

func getConfigFilePath() string {
	if configPath := os.Getenv("CFDDNS_CONFIG_PATH"); configPath != "" {
		return configPath
	}

	homeDir, _ := os.UserHomeDir()
	userConfigPath := filepath.Join(homeDir, ".config", "cfddns", "cfddns.yml")
	if fileExists(userConfigPath) {
		return userConfigPath
	}

	systemConfigPath := "/etc/cfddns/cfddns.yml"
	if fileExists(systemConfigPath) {
		return systemConfigPath
	}

	if fileExists("cfddns.yml") {
		return "cfddns.yml"
	}

	return ""
}

func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}
