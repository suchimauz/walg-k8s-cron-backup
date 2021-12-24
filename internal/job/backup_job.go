package job

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/suchimauz/walg-k8s-cron-backup/internal/config"
	"github.com/suchimauz/walg-k8s-cron-backup/pkg/kube"
	"github.com/suchimauz/walg-k8s-cron-backup/pkg/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	klog "github.com/suchimauz/walg-k8s-cron-backup/pkg/logger"
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
func (bj *BackupJob) Run() {
	klog.Info("[BackupJob] Start processing Job!")

	// Make start message for send notification
	guid, startMsg := bj.startBackupMessage()

	if bj.Notification.Enabled {
		// Send notification about start backup db
		bj.sendNotifications(startMsg)
	}

	// Execute on container EXEC_BACKUP cmd and return backups info
	// !!! For some reason wal-g writes logs to stderr !!!
	stdout, stderr := bj.KubeJob.Exec(bj.Exec, nil)
	if stderr != nil {
		klog.Infof("[BackupJob] %s", stderr.Error())

		if bj.Notification.Enabled {
			// Make end message
			endMsg := bj.endBackupMessage(guid)
			// Send notification about end backup db
			bj.sendNotifications(endMsg)
		}

		return
	} else {
		klog.Infof("[BackupJob] %s", stdout)

		if bj.Notification.Enabled {
			// Make end message
			endMsg := bj.endBackupMessage(guid)
			// Send notification about end backup db
			bj.sendNotifications(endMsg)
		}
	}

	klog.Infof("[BackupJob] End processing Job!")
}

// Private method for send telegram notifications
func (bj *BackupJob) sendNotifications(msg string) {
	// Iterate with config users chat-ids, who get backup notifications
	for _, chatId := range bj.Notification.ChatIds {
		tgmsg := tgbotapi.NewMessage(chatId, msg)
		tgmsg.ParseMode = "HTMl"
		tgmsg.DisableNotification = true

		go func(gij *BackupJob, gtgmsg tgbotapi.MessageConfig) {
			_, err := gij.TelegramBotApi.Send(gtgmsg)
			if err != nil {
				klog.Errorf("[BackupJob] Can't send tg notification: %s", err.Error())
			}
		}(bj, tgmsg)
	}
}

// Private method for generate start backup message
func (bj *BackupJob) startBackupMessage() (uuid.UUID, string) {
	// Generate new uuid for set id for this backup context
	id := uuid.New()

	// Get now date with Russian format
	date := utils.NowDateTz().Format("02.01.2006 15:04")

	msg := fmt.Sprintf("<b>%s</b>: start backup", strings.ToUpper(bj.KubeJob.Pod.Namespace))
	msg += fmt.Sprintf("\n\nUuid: <b>%s</b>", id.String())
	msg += fmt.Sprintf("\nCommand: <code>%s</code>", bj.Exec)
	msg += fmt.Sprintf("\nDate: <b>%s</b>\n", date)

	return id, msg
}

// Private method for generate end backup message
func (bj *BackupJob) endBackupMessage(id uuid.UUID) string {
	// Get now date with Russian format
	date := utils.NowDateTz().Format("02.01.2006 15:04")

	msg := fmt.Sprintf("<b>%s</b>: end backup", strings.ToUpper(bj.KubeJob.Pod.Namespace))
	msg += fmt.Sprintf("\n\nUuid: <b>%s</b>", id.String())
	msg += fmt.Sprintf("\nDate: <b>%s</b>\n", date)

	return msg
}
