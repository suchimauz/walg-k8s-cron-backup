package config

import (
	"errors"
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
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
		SaveLogs    bool   `envconfig:"app_save_logs" default:"false"`
		Kubernetes  KubernetesConfig
		Exec        ExecConfig
		Cron        CronConfig
		Telegram    TelegramConfig
		FileStorage FileStorageConfig
	}

	KubernetesConfig struct {
		// Kind default Pod
		ApiVersion    string `default:"v1"`
		Host          string `envconfig:"k8s_host" required:"true"`
		Insecure      bool   `envconfig:"k8s_insecure" default:"true"`
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
		Info   string `envconfig:"cron_info"`
	}

	TelegramConfig struct {
		ApiEndpoint  string `envconfig:"tg_bot_api_endpoint" default:"https://api.telegram.org/bot%s/%s"`
		HttpProxy    string `envconfig:"tg_bot_http_proxy"`
		BotToken     string `envconfig:"tg_bot_token"`
		Notification TelegramNotificationConfig
	}

	TelegramNotificationConfig struct {
		Backup TelegramNotificationBackupConfig
		Info   TelegramNotificationInfoConfig
	}

	TelegramNotificationBackupConfig struct {
		Enabled bool    `envconfig:"tg_backup_notification_enabled" default:"false"`
		ChatIds []int64 `envconfig:"tg_backup_notification_chats" split_words:"true"`
	}

	TelegramNotificationInfoConfig struct {
		Enabled bool    `envconfig:"tg_info_notification_enabled" default:"false"`
		ChatIds []int64 `envconfig:"tg_info_notification_chats" split_words:"true"`
	}

	FileStorageConfig struct {
		Endpoint  string `envconfig:"fs_host"`
		Bucket    string `envconfig:"fs_bucket"`
		AccessKey string `envconfig:"fs_access_key"`
		SecretKey string `envconfig:"fs_secret_key"`
		Secure    bool   `envconfig:"fs_secure" default:"true"`
	}
)

func Init(dotenv func()) (*Config, error) {
	var cfg Config

	dotenv()

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

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (cfg *Config) validate() error {
	// When one of telegram notifications are enabled - required bot token
	if cfg.Telegram.NotificationsEnabled() {
		if cfg.Telegram.BotToken == "" {
			return errors.New("Telegram bot token is required, when one of notifications enable is true")
		}
	}

	// When save logs is true, file storage environment are required
	if cfg.FileStorageRequired() {
		if err := cfg.FileStorage.allRequired(); err != nil {
			msg := fmt.Sprintf("If save logs config value is true: %s", err.Error())

			return errors.New(msg)
		}
	}

	// When save logs is enabled or telegram info notifications are enabled - cron.Info is required
	if cfg.CronInfoRequired() {
		if cfg.Cron.Info == "" {
			return errors.New("If save logs is enabled or telegram info notifications are enabled: cron info is required")
		}
		if cfg.Exec.Info == "" {
			return errors.New("If save logs is enabled or telegram info notifications are enabled: exec info is required")
		}
	}

	return nil
}

func (cfg *Config) CronInfoRequired() bool {
	return cfg.SaveLogs || cfg.Telegram.Notification.Info.Enabled
}

func (cfg *Config) FileStorageRequired() bool {
	return cfg.SaveLogs
}

// Private func for set all FileStorageConfig fields required
func (fscfg *FileStorageConfig) allRequired() error {
	if fscfg.Endpoint == "" {
		return errors.New("FileStorage Endpoint is required")
	}
	if fscfg.Bucket == "" {
		return errors.New("FileStorage Bucket is required")
	}
	if fscfg.AccessKey == "" {
		return errors.New("FileStorage AccessKey is required")
	}
	if fscfg.SecretKey == "" {
		return errors.New("FileStorage SecretKey is required")
	}

	return nil
}

// Func for check telegram notifications enabled
func (tgcfg *TelegramConfig) NotificationsEnabled() bool {
	return tgcfg.Notification.Info.Enabled || tgcfg.Notification.Backup.Enabled
}
