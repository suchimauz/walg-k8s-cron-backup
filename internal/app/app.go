package app

import (
	"github.com/suchimauz/walg-k8s-cron-backup/internal/config"
	"github.com/suchimauz/walg-k8s-cron-backup/pkg/kube"
	klog "github.com/suchimauz/walg-k8s-cron-backup/pkg/logger"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func Run() {
	cfg, err := config.Init()
	if err != nil {
		klog.Errorf("ENV: %s", err.Error())

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
		klog.Error(err.Error())

		return
	}

	kjob, err := kube.NewKubeJob(clientset, kubeConfig, cfg.Kubernetes.Namespace,
		cfg.Kubernetes.LabelSelector, cfg.Kubernetes.ContainerName)
	if err != nil {
		klog.Errorf("KubeJob: %s", err.Error())

		return
	}

	resp, err := kjob.Exec("wal-g backup-list --json --pretty --detail", nil)
	if err != nil {
		klog.Errorf("K8S Exec: %s", err.Error())

		return
	}
	klog.Info(resp)
}
