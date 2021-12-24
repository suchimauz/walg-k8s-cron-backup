package config

import (
	"os"
	"reflect"
	"testing"
)

func TestInit(t *testing.T) {
	requiredEnv := func() {
		os.Setenv("K8S_HOST", "localhost")
		os.Setenv("K8S_AUTH_TOKEN", "token")
		os.Setenv("K8S_NAMESPACE", "ns")
		os.Setenv("K8S_LABEL_SELECTOR", "labelSelector")
		os.Setenv("K8S_POD_CONTAINER_NAME", "podContainerName")

		os.Setenv("EXEC_BACKUP", "execBackup")
		os.Setenv("CRON_BACKUP", "cronBackup")
	}

	tests := []struct {
		name    string
		want    *Config
		envFunc func()
		wantErr bool
	}{
		// Test setting default and required values
		{
			name:    "test config default values",
			envFunc: requiredEnv,
			want: &Config{
				Timezone: "UTC",
				SaveLogs: false,
				Kubernetes: KubernetesConfig{
					ApiVersion:    "v1",
					Host:          "localhost",
					Insecure:      true,
					BearerToken:   "token",
					Namespace:     "ns",
					LabelSelector: "labelSelector",
					ContainerName: "podContainerName",
				},
				Exec: ExecConfig{
					Backup: "execBackup",
					Info:   "echo 1",
				},
				Cron: CronConfig{
					Backup: "cronBackup",
					Info:   "",
				},
				Telegram: TelegramConfig{
					BotToken: "",
					Notification: TelegramNotificationConfig{
						Backup: TelegramNotificationBackupConfig{
							Enabled: false,
							ChatIds: nil,
						},
						Info: TelegramNotificationInfoConfig{
							Enabled: false,
							ChatIds: nil,
						},
					},
				},
				FileStorage: FileStorageConfig{
					Endpoint:  "",
					Bucket:    "",
					AccessKey: "",
					SecretKey: "",
					Secure:    true,
				},
			},
		},

		// Test getting error when not passed required vars
		{
			name:    "test config without required variables",
			envFunc: func() {},
			wantErr: true,
		},

		// Tests validate if one of notifications are enabled
		{
			name: "tests validate if one of notifications are enabled, but token not passed",
			envFunc: func() {
				requiredEnv()
				os.Setenv("TG_BACKUP_NOTIFICATION_ENABLED", "true")
			},
			wantErr: true,
		},
		{
			name: "tests validate if one of notifications are enabled, but token passed",
			envFunc: func() {
				requiredEnv()
				os.Setenv("TG_BACKUP_NOTIFICATION_ENABLED", "true")
				os.Setenv("TG_BOT_TOKEN", "token")
			},
			want: &Config{
				Timezone: "UTC",
				SaveLogs: false,
				Kubernetes: KubernetesConfig{
					ApiVersion:    "v1",
					Host:          "localhost",
					Insecure:      true,
					BearerToken:   "token",
					Namespace:     "ns",
					LabelSelector: "labelSelector",
					ContainerName: "podContainerName",
				},
				Exec: ExecConfig{
					Backup: "execBackup",
					Info:   "echo 1",
				},
				Cron: CronConfig{
					Backup: "cronBackup",
					Info:   "",
				},
				Telegram: TelegramConfig{
					BotToken: "token",
					Notification: TelegramNotificationConfig{
						Backup: TelegramNotificationBackupConfig{
							Enabled: true,
							ChatIds: nil,
						},
						Info: TelegramNotificationInfoConfig{
							Enabled: false,
							ChatIds: nil,
						},
					},
				},
				FileStorage: FileStorageConfig{
					Endpoint:  "",
					Bucket:    "",
					AccessKey: "",
					SecretKey: "",
					Secure:    true,
				},
			},
		},

		// Tests validate if save logs are enable
		{
			name: "tests validate if APP_SAVE_LOGS=true, but FileStorage Config not passed",
			envFunc: func() {
				requiredEnv()
				os.Setenv("APP_SAVE_LOGS", "true")
			},
			wantErr: true,
		},
		{
			name: "tests validate if APP_SAVE_LOGS=true, FileStorage Config passed, but CRON_INFO not passed",
			envFunc: func() {
				requiredEnv()
				os.Setenv("APP_SAVE_LOGS", "true")
				os.Setenv("FS_HOST", "host")
				os.Setenv("FS_BUCKET", "bucket")
				os.Setenv("FS_ACCESS_KEY", "accessKey")
				os.Setenv("FS_SECRET_KEY", "secretKey")
				os.Setenv("FS_SECURE", "false")
			},
			wantErr: true,
		},
		{
			name: "tests validate if APP_SAVE_LOGS=true, FileStorage Config passed, CRON_INFO passed",
			envFunc: func() {
				requiredEnv()
				os.Setenv("APP_SAVE_LOGS", "true")
				os.Setenv("FS_HOST", "host")
				os.Setenv("FS_BUCKET", "bucket")
				os.Setenv("FS_ACCESS_KEY", "accessKey")
				os.Setenv("FS_SECRET_KEY", "secretKey")
				os.Setenv("FS_SECURE", "false")

				os.Setenv("CRON_INFO", "cronInfo")
			},
			want: &Config{
				Timezone: "UTC",
				SaveLogs: true,
				Kubernetes: KubernetesConfig{
					ApiVersion:    "v1",
					Host:          "localhost",
					Insecure:      true,
					BearerToken:   "token",
					Namespace:     "ns",
					LabelSelector: "labelSelector",
					ContainerName: "podContainerName",
				},
				Exec: ExecConfig{
					Backup: "execBackup",
					Info:   "echo 1",
				},
				Cron: CronConfig{
					Backup: "cronBackup",
					Info:   "cronInfo",
				},
				Telegram: TelegramConfig{
					BotToken: "",
					Notification: TelegramNotificationConfig{
						Backup: TelegramNotificationBackupConfig{
							Enabled: false,
							ChatIds: nil,
						},
						Info: TelegramNotificationInfoConfig{
							Enabled: false,
							ChatIds: nil,
						},
					},
				},
				FileStorage: FileStorageConfig{
					Endpoint:  "host",
					Bucket:    "bucket",
					AccessKey: "accessKey",
					SecretKey: "secretKey",
					Secure:    false,
				},
			},
		},

		//
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()

			got, err := Init(tt.envFunc)
			if (err != nil) != tt.wantErr {
				t.Errorf("Init() \nerror = %v\nwantErr %v", err, tt.wantErr)

				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Init() \ngot = %v\nwant %v", got, tt.want)
			}
		})
	}
}
