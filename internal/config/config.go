package config

import (
	"os"

	"github.com/joho/godotenv"
)

// Email содержит параметры для отправки почты
type Email struct {
	Host     string
	Port     string
	Username string
	Password string
}

// App содержит всю конфигурацию приложения
type App struct {
	Email Email
}

// ConfigLoader определяет метод для загрузки конфигурации
type ConfigLoader interface {
	Load() (App, error)
}

// DotenvConfigLoader реализует интерфейс ConfigLoader и загружает конфигурацию из .env файла
type DotenvConfigLoader struct{}

// Load загружает конфигурацию из .env файла
func (d *DotenvConfigLoader) Load() (App, error) {
	err := godotenv.Load()
	if err != nil {
		return App{}, err
	}

	email := Email{
		Host:     os.Getenv("SMTP_HOST"),
		Port:     os.Getenv("SMTP_PORT"),
		Username: os.Getenv("SMTP_USERNAME"),
		Password: os.Getenv("SMTP_PASSWORD"),
	}

	return App{
		Email: email,
	}, nil
}
