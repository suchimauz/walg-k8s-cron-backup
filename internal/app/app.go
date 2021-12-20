package app

import (
	"github.com/suchimauz/walg-k8s-cron-backup/internal/config"

	"github.com/davecgh/go-spew/spew"
	"github.com/suchimauz/walg-k8s-cron-backup/pkg/logger"
)

func Run() {
	cfg, err := config.Init()
	if err != nil {
		logger.Errorf("ENV: %s", err.Error())

		return
	}

	spew.Dump(cfg)

	return
}
