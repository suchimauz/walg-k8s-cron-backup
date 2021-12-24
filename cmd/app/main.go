package main

import (
	"github.com/joho/godotenv"
	"github.com/suchimauz/walg-k8s-cron-backup/internal/app"
)

func main() {
	app.Run(dotenv)
}

func dotenv() {
	godotenv.Load()
}
