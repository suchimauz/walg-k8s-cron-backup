package job

import (
	"github.com/suchimauz/walg-k8s-cron-backup/internal/config"
	"github.com/suchimauz/walg-k8s-cron-backup/pkg/kube"
	"github.com/suchimauz/walg-k8s-cron-backup/pkg/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	cr "github.com/robfig/cron/v3"
)

// Help func for insert need jobs to cron scheduler
func InsertJobs(cron *cr.Cron, cfg *config.Config, kj *kube.KubeJob, botapi *tgbotapi.BotAPI, storageProvider storage.Provider) ([]cr.EntryID, error) {
	// Init variables
	var entryIds []cr.EntryID
	var eId cr.EntryID
	var err error

	// InfoJob - object for manage job, which send notifications of backups and etc
	// Required when save logs is enabled or telegram notification is enabled
	if cfg.CronInfoRequired() {
		ij := NewInfoJob(&cfg.Telegram, kj, botapi, storageProvider, cfg.Exec.Info)

		// Add to exists cron object new InfoJob object
		eId, err = cron.AddJob(cfg.Cron.Info, ij)
		if err != nil {
			return nil, err
		}
		entryIds = append(entryIds, eId)
	}

	// BackupJob - object for manage job, which send command for backuping postgres db and etc.
	bj := NewBackupJob(&cfg.Telegram, kj, botapi, cfg.Exec.Backup)
	// Add to exists cron object new BackupJob object
	eId, err = cron.AddJob(cfg.Cron.Backup, bj)
	if err != nil {
		return nil, err
	}
	entryIds = append(entryIds, eId)

	// Return array of new cron job ids
	return entryIds, nil
}
