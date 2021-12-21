package job

import (
	"github.com/suchimauz/walg-k8s-cron-backup/internal/config"
	"github.com/suchimauz/walg-k8s-cron-backup/pkg/kube"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// BackupJob - struct for manage job, which send commands for make backup
type BackupJob struct {
	KubeJob        *kube.KubeJob
	Notification   *config.TelegramNotificationBackupConfig
	Exec           string
	TelegramBotApi *tgbotapi.BotAPI
}

// Constructor
func NewBackupJob(telegramCfg *config.TelegramConfig, kj *kube.KubeJob, botapi *tgbotapi.BotAPI, exec string) *BackupJob {
	return &BackupJob{
		KubeJob:        kj,
		Notification:   &telegramCfg.Notification.Backup,
		Exec:           exec,
		TelegramBotApi: botapi,
	}
}

// Main required method, which implements cron.Job interface
func (ij *BackupJob) Run() {
}
