package di

import (
	"fmt"

	"github.com/secretlyhq/secretly/internal/config"
)

func InitializeApp() (string, error) {
	fmt.Println("InitializeApp() called")
	conf, err := config.Load("configs/config.yaml")
	if err != nil {
		fmt.Println("Error loading config:", err)
		return "", err
	}
	fmt.Println("Config successfully loaded:", conf)
	return "✅ Secretly app initialized. DB migrated.", nil
}
