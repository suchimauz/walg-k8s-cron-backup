package utils

import (
	"time"

	"github.com/suchimauz/walg-k8s-cron-backup/internal/config"
)

// Func for get now datetime with timezone in cfg
func NowDateTz() time.Time {
	return time.Now().In(config.TimeZone)
}
