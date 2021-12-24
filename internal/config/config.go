package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"

	_ "github.com/joho/godotenv/autoload"
)

const (
	appName = "cron_backup"
)

// Not modify this variable!!!
// This variable will be filled when initializing the config
var TimeZone *time.Location

type (
	Config struct {
		Timezone    string `envconfig:"app_timezone" default:"UTC"` // String timezone format
		Kubernetes  KubernetesConfig
		Exec        ExecConfig
		Cron        CronConfig
		Telegram    TelegramConfig
		FileStorage FileStorageConfig
	}

	KubernetesConfig struct {
		// Kind default Pod
		ApiVersion    string `ignored:"true" default:"v1"`
		Host          string `envconfig:"k8s_host" required:"true"`
		Insecure      bool   `envconfig:"k8s_insecure" default:"false"`
		BearerToken   string `envconfig:"k8s_auth_token" required:"true"`
		Namespace     string `envconfig:"k8s_namespace" required:"true"`
		LabelSelector string `envconfig:"k8s_label_selector" required:"true"`
		ContainerName string `envconfig:"k8s_pod_container_name" required:"true"`
	}

	ExecConfig struct {
		Backup string `envconfig:"exec_backup" required:"true"`
		Info   string `envconfig:"exec_info" default:"echo 1"`
	}

	CronConfig struct {
		Backup string `envconfig:"cron_backup" required:"true"`
		Info   string `envconfig:"cron_info" default:"0 0 0 31 2 1"` // Never execute
	}

	TelegramConfig struct {
		BotToken     string `envconfig:"tg_bot_token" required:"true"` // This is test bot
		Notification TelegramNotificationConfig
	}

	TelegramNotificationConfig struct {
		Backup TelegramNotificationBackupConfig
		Info   TelegramNotificationInfoConfig
	}

	TelegramNotificationBackupConfig struct {
		Enabled bool    `envconfig:"tg_backup_notification_enabled" default:"false"`
		ChatIds []int64 `envconfig:"tg_backup_notification_chats" split_words:"true" default:"000000"`
	}

	TelegramNotificationInfoConfig struct {
		Enabled bool    `envconfig:"tg_info_notification_enabled" default:"false"`
		ChatIds []int64 `envconfig:"tg_info_notification_chats" split_words:"true" default:"000000"`
	}

	FileStorageConfig struct {
		Endpoint  string `envconfig:"fs_host"`
		Bucket    string `envconfig:"fs_bucket"`
		AccessKey string `envconfig:"fs_access_key"`
		SecretKey string `envconfig:"fs_secret_key"`
		Secure    bool   `envconfig:"fs_secure" default:"true"`
	}
)

func Init() (*Config, error) {
	var cfg Config

	// Parse variables from environment or return err
	err := envconfig.Process(appName, &cfg)
	if err != nil {
		return nil, err
	}

	// Parse timezone from cfg.tz or return err
	TimeZone, err = time.LoadLocation(cfg.Timezone)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
