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
