package app

import (
	"os"
	"os/signal"
	"syscall"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	cr "github.com/robfig/cron/v3"
	"github.com/suchimauz/walg-k8s-cron-backup/internal/config"
	cjobs "github.com/suchimauz/walg-k8s-cron-backup/internal/job"
	"github.com/suchimauz/walg-k8s-cron-backup/pkg/kube"
	klog "github.com/suchimauz/walg-k8s-cron-backup/pkg/logger"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func Run() {
	cfg, err := config.Init()
	if err != nil {
		klog.Errorf("[ENV] %s", err.Error())

		return
	}

	tlsClientConfig := &rest.TLSClientConfig{Insecure: cfg.Kubernetes.Insecure}

	kubeConfig := &rest.Config{
		Host:            cfg.Kubernetes.Host,
		APIPath:         cfg.Kubernetes.ApiVersion,
		BearerToken:     cfg.Kubernetes.BearerToken,
		TLSClientConfig: *tlsClientConfig,
	}

	clientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		klog.Errorf("[KubeConfig] %s", err.Error())

		return
	}

	kjob, err := kube.NewKubeJob(clientset, kubeConfig, cfg.Kubernetes.Namespace,
		cfg.Kubernetes.LabelSelector, cfg.Kubernetes.ContainerName)
	if err != nil {
		klog.Errorf("[KubeJob] %s", err.Error())

		return
	}

	tgbot, err := tgbotapi.NewBotAPI(cfg.Telegram.BotToken)
	if err != nil {
		klog.Errorf("[TelegramBotApi] %s", err.Error())

		return
	}

	cron := cr.New(cr.WithSeconds())
	// Insert jobs to cron
	jobIds, err := cjobs.InsertJobs(cron, cfg, kjob, tgbot)
	if err != nil {
		klog.Errorf("[Cron] Error inserting jobs: %s", err.Error())

		return
	}

	go cron.Run()

	klog.Infof("[Cron] Started! JobIds %a", jobIds)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	<-quit

	cron.Stop()

	klog.Info("[Cron] Stopped! Exit")
}
