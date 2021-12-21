package app

import (
	"os"
	"os/signal"
	"syscall"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	cr "github.com/robfig/cron/v3"
	"github.com/suchimauz/walg-k8s-cron-backup/internal/config"
	cjobs "github.com/suchimauz/walg-k8s-cron-backup/internal/job"
	"github.com/suchimauz/walg-k8s-cron-backup/pkg/kube"
	klog "github.com/suchimauz/walg-k8s-cron-backup/pkg/logger"
	"github.com/suchimauz/walg-k8s-cron-backup/pkg/storage"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func Run() {
	cfg, err := config.Init()
	if err != nil {
		klog.Errorf("[ENV] %s", err.Error())

		return
	}

	// Initialize Kubernetes tls config for set insecure: true of false from config
	tlsClientConfig := &rest.TLSClientConfig{Insecure: cfg.Kubernetes.Insecure}

	// Initialize Kubernetes BearerToken config
	kubeConfig := &rest.Config{
		Host:            cfg.Kubernetes.Host,
		APIPath:         cfg.Kubernetes.ApiVersion,
		BearerToken:     cfg.Kubernetes.BearerToken,
		TLSClientConfig: *tlsClientConfig,
	}

	// Create new kubernetes client from kubeConfig
	clientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		klog.Errorf("[KubeConfig] %s", err.Error())

		return
	}

	// Create new local KubeJob pkg object
	kjob, err := kube.NewKubeJob(clientset, kubeConfig, cfg.Kubernetes.Namespace,
		cfg.Kubernetes.LabelSelector, cfg.Kubernetes.ContainerName)
	if err != nil {
		klog.Errorf("[KubeJob] %s", err.Error())

		return
	}

	// Create new object for telegram api
	tgbot, err := tgbotapi.NewBotAPI(cfg.Telegram.BotToken)
	if err != nil {
		klog.Errorf("[TelegramBotApi] %s", err.Error())

		return
	}

	// Init storage provider - minio
	storageProvider, err := newStorageProvider(cfg)
	if err != nil {
		klog.Errorf("[FileStorage] Provider: %s", err.Error())

		return
	}

	cron := cr.New(cr.WithSeconds())
	// Insert jobs to cron
	jobIds, err := cjobs.InsertJobs(cron, cfg, kjob, tgbot, storageProvider)
	if err != nil {
		klog.Errorf("[Cron] Error inserting jobs: %s", err.Error())

		return
	}

	// Start the cron scheduler in its own goroutine
	cron.Start()

	klog.Infof("[Cron] Started! JobIds %a", jobIds)

	// Graceful Shutdown

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	<-quit

	// Wait jobs and stop
	cron.Stop()

	klog.Info("[Cron] Stopped! Exit")
}

func newStorageProvider(cfg *config.Config) (storage.Provider, error) {
	client, err := minio.New(cfg.FileStorage.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.FileStorage.AccessKey, cfg.FileStorage.SecretKey, ""),
		Secure: cfg.FileStorage.Secure,
	})
	if err != nil {
		return nil, err
	}

	provider := storage.NewFileStorage(client, cfg.FileStorage.Bucket, cfg.FileStorage.Endpoint)

	return provider, nil
}
