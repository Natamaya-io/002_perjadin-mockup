package main

import (
	"fmt"
	"github.com/kemnaker/perjadin-backend/internal/config"
)

func main() {
	cfg := config.LoadConfig()
	fmt.Printf("Config loaded: %+v\n", cfg)
}
