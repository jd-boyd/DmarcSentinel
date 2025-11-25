package main

import (
	"fmt"
	"os"

	"dmarc-viewer/internal/config"
)

func main() {
	// Load configuration with CLI flags
	cfg, err := config.LoadWithFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Print loaded configuration
	fmt.Println("=== DMARC Report Viewer Configuration ===")
	fmt.Println()

	fmt.Println("IMAP Configuration:")
	fmt.Printf("  Host:     %s\n", cfg.IMAP.Host)
	fmt.Printf("  Port:     %d\n", cfg.IMAP.Port)
	fmt.Printf("  Username: %s\n", cfg.IMAP.Username)
	fmt.Printf("  Password: %s\n", maskPassword(cfg.IMAP.Password))
	fmt.Printf("  Folder:   %s\n", cfg.IMAP.Folder)
	fmt.Printf("  Use TLS:  %t\n", cfg.IMAP.UseTLS)
	fmt.Println()

	fmt.Println("Database Configuration:")
	fmt.Printf("  Path: %s\n", cfg.Database.Path)
	fmt.Println()

	fmt.Println("Web Server Configuration:")
	fmt.Printf("  Host: %s\n", cfg.Web.Host)
	fmt.Printf("  Port: %d\n", cfg.Web.Port)
	fmt.Println()

	fmt.Println("Sync Configuration:")
	fmt.Printf("  Interval:   %s\n", cfg.Sync.Interval)
	fmt.Printf("  On Startup: %t\n", cfg.Sync.OnStartup)
	fmt.Println()

	fmt.Println("Logging Configuration:")
	fmt.Printf("  Level:  %s\n", cfg.Logging.Level)
	fmt.Printf("  Format: %s\n", cfg.Logging.Format)
	fmt.Println()

	fmt.Println("Configuration loaded successfully!")
	fmt.Println()
	fmt.Println("Note: This is a basic configuration test.")
	fmt.Println("Full application functionality will be available in future tasks.")
}

// maskPassword masks the password for display, showing only first and last characters
func maskPassword(password string) string {
	if len(password) == 0 {
		return ""
	}
	if len(password) <= 2 {
		return "***"
	}
	return string(password[0]) + "***" + string(password[len(password)-1])
}
