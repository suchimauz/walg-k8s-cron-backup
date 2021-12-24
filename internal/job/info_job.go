package job

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/suchimauz/walg-k8s-cron-backup/internal/config"
	"github.com/suchimauz/walg-k8s-cron-backup/pkg/kube"
	"github.com/suchimauz/walg-k8s-cron-backup/pkg/storage"
	"github.com/suchimauz/walg-k8s-cron-backup/pkg/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	klog "github.com/suchimauz/walg-k8s-cron-backup/pkg/logger"
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

	var stdout, stderr bytes.Buffer

	// Execute on container EXEC_BACKUP cmd and return backups info
	if err := ij.KubeJob.Exec(ij.Exec, nil, &stdout, &stderr); err != nil {
		klog.Errorf("[NotifierJob] %s", err.Error())
		klog.Error("[NotifierJob] Exit Job!")

		return
	}
	if stderr.Len() > 0 {
		klog.Errorf("[NotifierJob] %s", stderr.String())
		klog.Error("[NotifierJob] Exit Job!")

		return
	}

	// Parse backups info json to array of objects
	backupsInfo, err := parseBackupsInfoJson(stdout.String())
	if err != nil {
		klog.Errorf("[NotifierJob] parse json: %s", err.Error())
		klog.Error("[NotifierJob] Exit Job!")

		return
	}

	// If TG_INFO_NOTIFICATION_ENABLED is true
	if ij.Notification.Enabled {
		// Send tg notifications
		ij.sendNotifications(backupsInfo)
	}

	// Save backupsInfo log file to storage
	err = ij.saveBackupsInfoFile(backupsInfo)
	if err != nil {
		klog.Errorf("[NotifierJob] Error on upload file: %s", err.Error())
	}

	klog.Info("[NotifierJob] End processing job!")
}

// Private method for send telegram notifications
func (ij *InfoJob) sendNotifications(bi []*BackupInfo) {
	var msg string

	// Get only full backups
	fullBackupsInfo := getOnlyFullBackups(bi)

	// If not full backups send message for users that backups is not exists
	// Else send backups info
	msg = MakeBackupsInfoMessage(fullBackupsInfo)
	if !(len(fullBackupsInfo) < 1) {
		klog.Warn("[NotifierJob] Backups not found!")
		klog.Info("[NotifierJob] Send notifications of backups not found!")

		msg = "<b>Список бэкапов пуст!</b>"
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

// Private function for save backups info to storage
func (ij *InfoJob) saveBackupsInfoFile(bi []*BackupInfo) error {
	fullBackupsInfo := getOnlyFullBackups(bi)

	// If not full backups, return non error
	if len(fullBackupsInfo) < 1 {
		return nil
	}

	// Parse fullBackupsInfo to json
	backupsInfoJson, err := json.Marshal(fullBackupsInfo)
	if err != nil {
		return err
	}

	// Get storage from InfoJob object
	s3 := ij.Storage

	// Init new empty UploadInput object
	file := storage.UploadInput{
		Name: fmt.Sprintf("walg_k8s_cron_backup/logs/backups_%s.json",
			utils.NowDateTz().Format("2006_01_02T15_04_05")),
		ContentType: "application/octet-stream",
		Size:        int64(len(backupsInfoJson)),
		File:        bytes.NewReader(backupsInfoJson),
	}

	// Upload file to storage
	path, err := s3.Upload(context.TODO(), file)
	if err != nil {
		return err
	}

	klog.Infof("[FileStorage] Save backups info: %s", path)

	return nil
}

// Get only full wal-g backups
// base_00000005000034600000006B -> true
// base_00000005000034600000006B_D_00000005000033A50000006C -> false
// backup name, which have _D_SOME is incremental backups
func getOnlyFullBackups(bi []*BackupInfo) []*BackupInfo {
	var preparedBackupsInfo []*BackupInfo

	// Get only full backups info, check backup name
	for _, backupInfo := range bi {
		matched, _ := regexp.MatchString("^(.*)_(.*)_(.*)$", backupInfo.BackupName)
		if !matched {
			preparedBackupsInfo = append(preparedBackupsInfo, backupInfo)
		}
	}

	return preparedBackupsInfo
}
