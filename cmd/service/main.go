package main

import (
	"github.com/Sheridanlk/Music-Service/internal/config"
	"github.com/Sheridanlk/Music-Service/internal/logger"
)

func main() {
	cfg := config.Load()
	log := logger.SetupLogger(cfg.Env)
	

}
