package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
	"gopkg.in/yaml.v3"
)

// Config representa la configuración completa de la aplicación
type Config struct {
	Checker       CheckerConfig       `yaml:"checker"`
	Notifications NotificationsConfig `yaml:"notifications"`
	Logging       LoggingConfig       `yaml:"logging"`
	
	// Variables de entorno
	TelegramBotToken string
	TelegramChatID   string
	DockerHost       string
	LogLevel         string
}

// CheckerConfig configuración del verificador
type CheckerConfig struct {
	Schedule           string   `yaml:"schedule"`
	ExcludeImages      []string `yaml:"exclude_images"`
	IncludeBuildImages bool     `yaml:"include_build_images"`
}

// NotificationsConfig configuración de notificaciones
type NotificationsConfig struct {
	Telegram TelegramConfig `yaml:"telegram"`
}

// TelegramConfig configuración específica de Telegram
type TelegramConfig struct {
	Enabled      bool   `yaml:"enabled"`
	TemplateFile string `yaml:"template_file"`
}

// LoggingConfig configuración de logging
type LoggingConfig struct {
	File       string `yaml:"file"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
}

// Load carga la configuración desde archivos .env y YAML
func Load(configPath string) (*Config, error) {
	// Cargar variables de entorno
	if err := godotenv.Load(); err != nil {
		// No es un error crítico si no existe el archivo .env
		fmt.Printf("Warning: .env file not found: %v\n", err)
	}

	// Cargar configuración YAML
	yamlFile, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(yamlFile, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	// Cargar variables de entorno
	config.TelegramBotToken = getEnv("TELEGRAM_BOT_TOKEN", "")
	config.TelegramChatID = getEnv("TELEGRAM_CHAT_ID", "")
	config.DockerHost = getEnv("DOCKER_HOST", "unix:///var/run/docker.sock")
	config.LogLevel = getEnv("LOG_LEVEL", "info")

	// Validar configuración requerida
	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// ValidateCronSchedule valida que el schedule sea un formato cron válido
func (c *CheckerConfig) ValidateCronSchedule() error {
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	_, err := parser.Parse(c.Schedule)
	return err
}

// validate valida que la configuración sea correcta
func (c *Config) validate() error {
	if c.Notifications.Telegram.Enabled {
		if c.TelegramBotToken == "" {
			return fmt.Errorf("TELEGRAM_BOT_TOKEN is required when telegram notifications are enabled")
		}
		if c.TelegramChatID == "" {
			return fmt.Errorf("TELEGRAM_CHAT_ID is required when telegram notifications are enabled")
		}
	}

	if err := c.Checker.ValidateCronSchedule(); err != nil {
		return fmt.Errorf("invalid cron schedule format: %w", err)
	}

	return nil
}

// getEnv obtiene una variable de entorno con valor por defecto
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
