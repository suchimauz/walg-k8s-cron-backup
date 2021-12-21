package job

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/suchimauz/walg-k8s-cron-backup/internal/config"
	"github.com/suchimauz/walg-k8s-cron-backup/pkg/kube"
	klog "github.com/suchimauz/walg-k8s-cron-backup/pkg/logger"
	"github.com/suchimauz/walg-k8s-cron-backup/pkg/storage"
)

// InfoJob - struct for manage job, which send notifications of backups and etc
type InfoJob struct {
	Storage        storage.Provider
	KubeJob        *kube.KubeJob
	Notification   *config.TelegramNotificationInfoConfig
	Exec           string
	TelegramBotApi *tgbotapi.BotAPI
}

// Constructor
func NewInfoJob(telegramCfg *config.TelegramConfig, kj *kube.KubeJob, botapi *tgbotapi.BotAPI, storageProvider storage.Provider, exec string) *InfoJob {
	return &InfoJob{
		Storage:        storageProvider,
		KubeJob:        kj,
		Notification:   &telegramCfg.Notification.Info,
		Exec:           exec,
		TelegramBotApi: botapi,
	}
}

// Main func for Run this job, implements for cron.Job interface
func (ij *InfoJob) Run() {
	klog.Info("[NotifierJob] Start processing Job!")

	// Execute on container EXEC_BACKUP cmd and return backups info
	backupsJson, err := ij.KubeJob.Exec(ij.Exec, nil)
	if err != nil {
		klog.Errorf("[NotifierJob] %s", err.Error())
		klog.Error("[NotifierJob] Exit Job!")

		return
	}

	// Parse backups info json to array of objects
	backupsInfo, err := parseBackupsInfoJson(backupsJson)
	if err != nil {
		klog.Errorf("[NotifierJob] parse json: %s", err.Error())
		klog.Error("[NotifierJob] Exit Job!")

		return
	}

	// Send tg notifications
	ij.SendNotifications(backupsInfo)

	klog.Info("[NotifierJob] End processing job!")
}

func (ij *InfoJob) SendNotifications(bi []*BackupInfo) {
	var msg string

	// If not backups send message for users that backups is not exists
	// Else send backups info
	if len(bi) < 1 {
		klog.Warn("[NotifierJob] Backups not found!")
		klog.Info("[NotifierJob] Send notifications of backups not found!")

		msg = "<b>Список бэкапов пуст!</b>"
	} else {
		msg = MakeBackupsInfoMessage(bi)
	}

	// Iterate with config users chat-ids, who get notifications
	for _, chatId := range ij.Notification.ChatIds {
		tgmsg := tgbotapi.NewMessage(chatId, msg)
		tgmsg.ParseMode = "HTMl"
		tgmsg.DisableNotification = true

		go func(gij *InfoJob, gtgmsg tgbotapi.MessageConfig) {
			_, err := gij.TelegramBotApi.Send(gtgmsg)
			if err != nil {
				klog.Errorf("[NotifierJob] Can't send tg notification: %s", err.Error())
			}
		}(ij, tgmsg)
	}
}

// Help func for make message
func MakeBackupsInfoMessage(bi []*BackupInfo) string {
	msg := "<b>Список бэкапов:</b>"

	for _, backupInfo := range bi {
		// Bytes to Gigabytes
		backupSize := backupInfo.CompressedSize / (1024 * 1024 * 1024)

		msg += "\n<code>-------------------</code>"
		msg += fmt.Sprintf("\nНазвание: <b>%s</b>", backupInfo.BackupName)
		msg += fmt.Sprintf("\nДата: %s", backupInfo.Time.In(config.TimeZone).Format("02.01.2006 15:04"))
		msg += fmt.Sprintf("\nРазмер бэкапа: <b>%dGB</b>", backupSize)
	}

	return msg
}
