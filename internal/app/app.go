package app

import (
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/suchimauz/walg-k8s-cron-backup/internal/config"
	"github.com/suchimauz/walg-k8s-cron-backup/pkg/kube"
	"github.com/suchimauz/walg-k8s-cron-backup/pkg/storage"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	cr "github.com/robfig/cron/v3"
	cjobs "github.com/suchimauz/walg-k8s-cron-backup/internal/job"
	klog "github.com/suchimauz/walg-k8s-cron-backup/pkg/logger"
)

func Run(dotenv func()) {
	// Initialize config
	cfg, err := config.Init(dotenv)
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

	// Create new object for telegram api when one of notification is enabled
	var tgbot *tgbotapi.BotAPI
	// Create new http client for telegram api
	tgclient := &http.Client{}

	// When http proxy for telegram api is declared,
	if cfg.Telegram.HttpProxy != "" {
		proxy, err := url.Parse(cfg.Telegram.HttpProxy)
		if err != nil {
			klog.Errorf("[TelegramBotApi] Proxy: %s", err.Error())
		}

		transport := &http.Transport{}
		transport.Proxy = http.ProxyURL(proxy)

		tgclient.Transport = transport
	}

	if cfg.Telegram.NotificationsEnabled() {
		tgbot, err = tgbotapi.NewBotAPIWithClient(cfg.Telegram.BotToken, cfg.Telegram.ApiEndpoint, tgclient)
		if err != nil {
			klog.Errorf("[TelegramBotApi] %s", err.Error())

			return
		}
	}

	// Init storage provider - minio if save logs is enabled
	var storageProvider storage.Provider
	if cfg.FileStorageRequired() {
		storageProvider, err = newStorageProvider(cfg)
		if err != nil {
			klog.Errorf("[FileStorage] Provider: %s", err.Error())

			return
		}
	}

	// Make new cron object, calls constructor
	cron := cr.New(cr.WithSeconds())
	// Insert jobs to cron
	jobIds, err := cjobs.InsertJobs(cron, cfg, kjob, tgbot, storageProvider)
	if err != nil {
		klog.Errorf("[Cron] Error inserting jobs: %s", err.Error())

		return
	}

	// Start the cron scheduler in its own goroutine
	cron.Start()

	klog.Infof("[Cron] Started! JobIds %v", jobIds)

	// Graceful Shutdown

	// Make new channel of size = 1
	quit := make(chan os.Signal, 1)

	// Listen system 15 and 2 signals, when one of they called, send info to quit channel
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	// Read channel, this block of code lock this thread, until someone writes to the channel
	<-quit

	// When someone call SIGTERM or SIGINT signals, we'll get to here
	// cron.Stop() -> Wait jobs and stop
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
