package kube

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	// "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

type KubeJob struct {
	Client     *kubernetes.Clientset
	Pod        *v1.Pod
	Container  *v1.Container
	KubeConfig *rest.Config
}

func NewKubeJob(client *kubernetes.Clientset, k8scfg *rest.Config, namespace string, labelSelector string, containerName string) (*KubeJob, error) {
	pod, err := findPodByLabels(client, namespace, labelSelector)
	if err != nil {
		return nil, err
	}

	container, err := findContainerByName(pod, containerName)
	if err != nil {
		return nil, err
	}

	return &KubeJob{
		Client:     client,
		KubeConfig: k8scfg,
		Pod:        pod,
		Container:  container,
	}, nil
}

func findPodByLabels(client *kubernetes.Clientset, namespace string, labelSelector string) (*v1.Pod, error) {
	pods, err := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, err
	}

	if len(pods.Items) < 1 {
		errMsg := fmt.Sprintf("Pod with labels %s in %s namespace not found!", labelSelector, namespace)

		return nil, errors.New(errMsg)
	}

	pod := pods.Items[0]

	return &pod, nil
}

func findContainerByName(pod *v1.Pod, containerName string) (*v1.Container, error) {
	var foundContainers []v1.Container

	for _, container := range pod.Spec.Containers {
		if container.Name == containerName {
			foundContainers = append(foundContainers, container)
		}
	}

	if len(foundContainers) < 1 {
		errMsg := fmt.Sprintf("Container %s in pod %s/%s not found!", containerName, pod.Namespace, pod.Name)

		return nil, errors.New(errMsg)
	}
	foundContainer := foundContainers[0]

	return &foundContainer, nil
}

func (kj *KubeJob) Exec(command string, stdin io.Reader) (string, error) {
	req := kj.Client.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(kj.Pod.Name).
		Namespace(kj.Pod.Namespace).
		SubResource("exec")
	scheme := runtime.NewScheme()
	err := v1.AddToScheme(scheme)
	if err != nil {
		return "", err
	}

	parameterCodec := runtime.NewParameterCodec(scheme)
	req.VersionedParams(&v1.PodExecOptions{
		Command: []string{
			"sh",
			"-c",
			command,
		},
		Container: kj.Container.Name,
		Stdin:     stdin != nil,
		Stdout:    true,
		Stderr:    true,
		TTY:       false,
	}, parameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(kj.KubeConfig, "POST", req.URL())
	if err != nil {
		return "", err
	}

	var stdout, stderr bytes.Buffer
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  stdin,
		Stdout: &stdout,
		Stderr: &stderr,
		Tty:    false,
	})
	if err != nil {
		return "", err
	}
	if stderr.String() != "" {
		return "", errors.New(stderr.String())
	}

	return stdout.String(), nil
}
