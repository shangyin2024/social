package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"social/internal/config"
)

func main() {
	var (
		validate = flag.Bool("validate", false, "Validate configuration")
		show     = flag.Bool("show", false, "Show current configuration")
		env      = flag.String("env", "", "Set environment (development, staging, production)")
		format   = flag.String("format", "yaml", "Output format (yaml, json)")
	)
	flag.Parse()

	// Set environment if specified
	if *env != "" {
		os.Setenv("ENVIRONMENT", *env)
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Validate configuration
	if *validate {
		if err := cfg.Validate(); err != nil {
			log.Fatalf("Configuration validation failed: %v", err)
		}

		// Show warnings
		warnings := cfg.GetWarnings()
		if len(warnings) > 0 {
			fmt.Println("Configuration warnings:")
			for _, warning := range warnings {
				fmt.Printf("  ⚠️  %s\n", warning)
			}
		} else {
			fmt.Println("✅ Configuration is valid")
		}
	}

	// Show configuration
	if *show {
		switch *format {
		case "json":
			data, err := json.MarshalIndent(cfg, "", "  ")
			if err != nil {
				log.Fatalf("Failed to marshal configuration: %v", err)
			}
			fmt.Println(string(data))
		case "yaml":
			fmt.Printf("Server:\n")
			fmt.Printf("  Port: %s\n", cfg.Server.Port)
			fmt.Printf("  Base URL: %s\n", cfg.Server.BaseURL)
			fmt.Printf("Redis:\n")
			fmt.Printf("  Address: %s\n", cfg.Redis.Addr)
			fmt.Printf("  Database: %d\n", cfg.Redis.DB)
			fmt.Printf("Environment: %s\n", config.GetEnvironment())
			fmt.Printf("Multi-server configurations: %d\n", len(cfg.Servers))
		default:
			log.Fatalf("Unsupported format: %s", *format)
		}
	}

	// If no flags specified, show help
	if !*validate && !*show {
		fmt.Println("Configuration Management Tool")
		fmt.Println("Usage:")
		fmt.Println("  -validate    Validate configuration")
		fmt.Println("  -show        Show current configuration")
		fmt.Println("  -env         Set environment (development, staging, production)")
		fmt.Println("  -format      Output format (yaml, json)")
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  go run cmd/config/main.go -validate")
		fmt.Println("  go run cmd/config/main.go -show -format json")
		fmt.Println("  go run cmd/config/main.go -validate -env production")
	}
}
