package job

import (
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/suchimauz/walg-k8s-cron-backup/internal/config"
	"github.com/suchimauz/walg-k8s-cron-backup/pkg/kube"
	klog "github.com/suchimauz/walg-k8s-cron-backup/pkg/logger"
)

type InfoJob struct {
	KubeJob        *kube.KubeJob
	Notification   *config.TelegramNotificationInfo
	Exec           string
	TelegramBotApi *tgbotapi.BotAPI
}

func NewInfoJob(telegramCfg *config.Telegram, kj *kube.KubeJob, botapi *tgbotapi.BotAPI, exec string) *InfoJob {
	return &InfoJob{
		KubeJob:        kj,
		Notification:   &telegramCfg.Notification.Info,
		Exec:           exec,
		TelegramBotApi: botapi,
	}
}

func (ij *InfoJob) Run() {
	backupsJson, err := ij.KubeJob.Exec(ij.Exec, nil)
	if err != nil {
		klog.Errorf("[NotifierJob] %s", err.Error())
		klog.Error("[NotifierJob] Exit Job!")

		return
	}

	backupsInfo, err := parseBackupsInfoJson(backupsJson)
	if err != nil {
		klog.Errorf("[NotifierJob] parse json: %s", err.Error())
		klog.Error("[NotifierJob] Exit Job!")

		return
	}

	ij.SendNotifications(backupsInfo)
}

func (ij *InfoJob) SendNotifications(bi []*BackupInfo) {
	var msg string

	if len(bi) < 1 {
		klog.Warn("[NotifierJob] Backups not found!")
		klog.Info("[NotifierJob] Send notifications of backups not found!")

		msg = "<b>Список бэкапов пуст!</b>"
	} else {
		msg = MakeBackupsInfoMessage(bi)
	}

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

func MakeBackupsInfoMessage(bi []*BackupInfo) string {
	msg := "<b>Список бэкапов:</b>"

	for _, backupInfo := range bi {
		tz, _ := time.LoadLocation("Europe/Moscow")
		// Bytes to Gigabytes
		backupSize := backupInfo.CompressedSize / (1024 * 1024 * 1024)

		msg += "\n-------------------"
		msg += fmt.Sprintf("\nНазвание: <b>%s</b>", backupInfo.BackupName)
		msg += fmt.Sprintf("\nДата: %s", backupInfo.Time.In(tz).Format("02.01.2006 15:04"))
		msg += fmt.Sprintf("\nРазмер бэкапа: <b>%dGB</b>", backupSize)
	}

	return msg
}
