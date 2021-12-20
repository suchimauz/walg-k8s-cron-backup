package config

import (
	_ "github.com/joho/godotenv/autoload"
	"github.com/kelseyhightower/envconfig"
)

const (
	appName = "cron-backup"
)

type (
	Config struct {
		Kubernetes Kubernetes
		Exec       Exec
		Cron       Cron
		Telegram   Telegram
	}

	Kubernetes struct {
		// Kind default Pod
		ApiVersion    string `ignored:"true" default:"v1"`
		Host          string `envconfig:"k8s_host" required:"true"`
		Insecure      bool   `envconfig:"k8s_insecure" default:"false"`
		BearerToken   string `envconfig:"k8s_auth_token" required:"true"`
		Namespace     string `envconfig:"k8s_namespace" required:"true"`
		LabelSelector string `envconfig:"k8s_label_selector" required:"true"`
		ContainerName string `envconfig:"k8s_pod_container_name" required:"true"`
	}

	Exec struct {
		Backup string `envconfig:"exec_backup" required:"true"`
		Info   string `envconfig:"exec_info"`
	}

	Cron struct {
		Backup string `envconfig:"cron_backup" required:"true"`
		Info   string `envconfig:"cron_info"`
	}

	Telegram struct {
		BotToken     string `envconfig:"tg_bot_token"`
		Notification TelegramNotification
	}

	TelegramNotification struct {
		Backup TelegramNotificationBackup
		Info   TelegramNotificationInfo
	}

	TelegramNotificationBackup struct {
		Enabled bool    `envconfig:"tg_backup_notification_enabled" default:"true"`
		ChatIds []int64 `envconfig:"tg_backup_notification_chats" split_words:"true"`
	}

	TelegramNotificationInfo struct {
		Enabled bool    `envconfig:"tg_info_notification_enabled" default:"true"`
		ChatIds []int64 `envconfig:"tg_info_notification_chats" split_words:"true"`
	}
)

func Init() (*Config, error) {
	var cfg Config

	err := envconfig.Process(appName, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
