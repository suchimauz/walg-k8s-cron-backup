package job

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	cr "github.com/robfig/cron/v3"
	"github.com/suchimauz/walg-k8s-cron-backup/internal/config"
	"github.com/suchimauz/walg-k8s-cron-backup/pkg/kube"
)

func InsertJobs(cron *cr.Cron, cfg *config.Config, kj *kube.KubeJob, botapi *tgbotapi.BotAPI) ([]cr.EntryID, error) {
	var entryIds []cr.EntryID
	var eId cr.EntryID
	var err error

	ij := NewInfoJob(&cfg.Telegram, kj, botapi, cfg.Exec.Info)
	bj := NewBackupJob(&cfg.Telegram, kj, botapi, cfg.Exec.Backup)

	eId, err = cron.AddJob(cfg.Cron.Info, ij)
	if err != nil {
		return nil, err
	}
	entryIds = append(entryIds, eId)

	eId, err = cron.AddJob(cfg.Cron.Backup, bj)
	if err != nil {
		return nil, err
	}
	entryIds = append(entryIds, eId)

	return entryIds, nil
}
