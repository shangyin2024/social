package main

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the configuration structure
type Config struct {
	Server   ServerConfig                 `yaml:"server"`
	Redis    RedisConfig                  `yaml:"redis"`
	OAuth    OAuthConfig                  `yaml:"oauth"`
	Platform PlatformConfig               `yaml:"platform"`
	Servers  map[string]ServerOAuthConfig `yaml:"servers"`
}

type ServerConfig struct {
	Port    string `yaml:"port"`
	BaseURL string `yaml:"base_url"`
}

type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type OAuthConfig struct {
	YouTube   ProviderConfig `yaml:"youtube"`
	X         ProviderConfig `yaml:"x"`
	Facebook  ProviderConfig `yaml:"facebook"`
	TikTok    ProviderConfig `yaml:"tiktok"`
	Instagram ProviderConfig `yaml:"instagram"`
}

type ProviderConfig struct {
	ClientID     string   `yaml:"client_id"`
	ClientSecret string   `yaml:"client_secret"`
	Scopes       []string `yaml:"scopes"`
}

type PlatformConfig struct {
	SupportedProviders []string `yaml:"supported_providers"`
}

type ServerOAuthConfig struct {
	YouTube   ProviderConfig `yaml:"youtube"`
	X         ProviderConfig `yaml:"x"`
	Facebook  ProviderConfig `yaml:"facebook"`
	TikTok    ProviderConfig `yaml:"tiktok"`
	Instagram ProviderConfig `yaml:"instagram"`
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]
	configFile := "config.yaml"

	switch command {
	case "list":
		listServers(configFile)
	case "add":
		if len(os.Args) < 3 {
			fmt.Println("Usage: go run config-helper.go add <server_name>")
			return
		}
		addServer(configFile, os.Args[2])
	case "remove":
		if len(os.Args) < 3 {
			fmt.Println("Usage: go run config-helper.go remove <server_name>")
			return
		}
		removeServer(configFile, os.Args[2])
	case "show":
		if len(os.Args) < 3 {
			fmt.Println("Usage: go run config-helper.go show <server_name>")
			return
		}
		showServer(configFile, os.Args[2])
	case "validate":
		validateConfig(configFile)
	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Println("配置管理工具")
	fmt.Println("用法:")
	fmt.Println("  go run config-helper.go list                    - 列出所有服务器")
	fmt.Println("  go run config-helper.go add <server_name>       - 添加新服务器配置")
	fmt.Println("  go run config-helper.go remove <server_name>    - 删除服务器配置")
	fmt.Println("  go run config-helper.go show <server_name>      - 显示服务器配置")
	fmt.Println("  go run config-helper.go validate                - 验证配置文件")
}

func loadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func saveConfig(filename string, config *Config) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

func listServers(configFile string) {
	config, err := loadConfig(configFile)
	if err != nil {
		fmt.Printf("错误: 无法加载配置文件 %s: %v\n", configFile, err)
		return
	}

	if len(config.Servers) == 0 {
		fmt.Println("没有配置任何服务器")
		return
	}

	fmt.Println("已配置的服务器:")
	for name := range config.Servers {
		fmt.Printf("  - %s\n", name)
	}
}

func addServer(configFile string, serverName string) {
	config, err := loadConfig(configFile)
	if err != nil {
		fmt.Printf("错误: 无法加载配置文件 %s: %v\n", configFile, err)
		return
	}

	if config.Servers == nil {
		config.Servers = make(map[string]ServerOAuthConfig)
	}

	if _, exists := config.Servers[serverName]; exists {
		fmt.Printf("服务器 '%s' 已存在\n", serverName)
		return
	}

	// 创建默认配置
	serverConfig := ServerOAuthConfig{
		YouTube: ProviderConfig{
			ClientID:     fmt.Sprintf("%s_youtube_client_id", serverName),
			ClientSecret: fmt.Sprintf("%s_youtube_client_secret", serverName),
			Scopes: []string{
				"https://www.googleapis.com/auth/youtube.upload",
				"openid",
				"email",
			},
		},
		X: ProviderConfig{
			ClientID:     fmt.Sprintf("%s_x_client_id", serverName),
			ClientSecret: fmt.Sprintf("%s_x_client_secret", serverName),
			Scopes: []string{
				"tweet.read",
				"tweet.write",
				"users.read",
				"offline.access",
			},
		},
		Facebook: ProviderConfig{
			ClientID:     fmt.Sprintf("%s_facebook_app_id", serverName),
			ClientSecret: fmt.Sprintf("%s_facebook_app_secret", serverName),
			Scopes: []string{
				"pages_manage_posts",
				"pages_read_engagement",
				"pages_show_list",
				"pages_read_user_content",
			},
		},
		TikTok: ProviderConfig{
			ClientID:     fmt.Sprintf("%s_tiktok_client_id", serverName),
			ClientSecret: fmt.Sprintf("%s_tiktok_client_secret", serverName),
			Scopes: []string{
				"video.upload",
				"user.info.basic",
			},
		},
		Instagram: ProviderConfig{
			ClientID:     fmt.Sprintf("%s_instagram_client_id", serverName),
			ClientSecret: fmt.Sprintf("%s_instagram_client_secret", serverName),
			Scopes: []string{
				"instagram_content_publish",
				"pages_read_engagement",
			},
		},
	}

	config.Servers[serverName] = serverConfig

	err = saveConfig(configFile, config)
	if err != nil {
		fmt.Printf("错误: 无法保存配置文件: %v\n", err)
		return
	}

	fmt.Printf("成功添加服务器 '%s'\n", serverName)
	fmt.Println("请编辑配置文件，填入真实的OAuth凭据")
}

func removeServer(configFile string, serverName string) {
	config, err := loadConfig(configFile)
	if err != nil {
		fmt.Printf("错误: 无法加载配置文件 %s: %v\n", configFile, err)
		return
	}

	if _, exists := config.Servers[serverName]; !exists {
		fmt.Printf("服务器 '%s' 不存在\n", serverName)
		return
	}

	delete(config.Servers, serverName)

	err = saveConfig(configFile, config)
	if err != nil {
		fmt.Printf("错误: 无法保存配置文件: %v\n", err)
		return
	}

	fmt.Printf("成功删除服务器 '%s'\n", serverName)
}

func showServer(configFile string, serverName string) {
	config, err := loadConfig(configFile)
	if err != nil {
		fmt.Printf("错误: 无法加载配置文件 %s: %v\n", configFile, err)
		return
	}

	serverConfig, exists := config.Servers[serverName]
	if !exists {
		fmt.Printf("服务器 '%s' 不存在\n", serverName)
		return
	}

	fmt.Printf("服务器 '%s' 的配置:\n", serverName)

	providers := map[string]ProviderConfig{
		"YouTube":   serverConfig.YouTube,
		"X":         serverConfig.X,
		"Facebook":  serverConfig.Facebook,
		"TikTok":    serverConfig.TikTok,
		"Instagram": serverConfig.Instagram,
	}

	for name, provider := range providers {
		if provider.ClientID != "" {
			fmt.Printf("\n%s:\n", name)
			fmt.Printf("  Client ID: %s\n", provider.ClientID)
			fmt.Printf("  Client Secret: %s\n", maskSecret(provider.ClientSecret))
			fmt.Printf("  Scopes: %s\n", strings.Join(provider.Scopes, ", "))
		}
	}
}

func validateConfig(configFile string) {
	config, err := loadConfig(configFile)
	if err != nil {
		fmt.Printf("错误: 无法加载配置文件 %s: %v\n", configFile, err)
		return
	}

	fmt.Println("配置文件验证结果:")

	// 验证基本配置
	if config.Server.Port == "" {
		fmt.Println("❌ 服务器端口未配置")
	} else {
		fmt.Printf("✅ 服务器端口: %s\n", config.Server.Port)
	}

	if config.Server.BaseURL == "" {
		fmt.Println("❌ 服务器基础URL未配置")
	} else {
		fmt.Printf("✅ 服务器基础URL: %s\n", config.Server.BaseURL)
	}

	// 验证Redis配置
	if config.Redis.Addr == "" {
		fmt.Println("❌ Redis地址未配置")
	} else {
		fmt.Printf("✅ Redis地址: %s\n", config.Redis.Addr)
	}

	// 验证服务器配置
	if len(config.Servers) == 0 {
		fmt.Println("⚠️  没有配置任何服务器，将使用默认配置")
	} else {
		fmt.Printf("✅ 配置了 %d 个服务器:\n", len(config.Servers))
		for name := range config.Servers {
			fmt.Printf("  - %s\n", name)
		}
	}

	fmt.Println("\n配置文件验证完成")
}

func maskSecret(secret string) string {
	if len(secret) <= 8 {
		return "***"
	}
	return secret[:4] + "***" + secret[len(secret)-4:]
}
