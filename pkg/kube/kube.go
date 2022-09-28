package kube

import (
	"context"
	"errors"
	"fmt"
	"io"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PodSelector struct {
	LabelSelector string
	ContainerName string
	Namespace     string
}

// Assembly of necessary methods and fields for kubernetes for project
type KubeJob struct {
	Client      *kubernetes.Clientset
	KubeConfig  *rest.Config
	PodSelector *PodSelector
}

// Create new KubeJob struct object [Constructor]
func NewKubeJob(client *kubernetes.Clientset, k8scfg *rest.Config, namespace string, labelSelector string, containerName string) (*KubeJob, error) {
	return &KubeJob{
		Client:     client,
		KubeConfig: k8scfg,
		PodSelector: &PodSelector{
			LabelSelector: labelSelector,
			ContainerName: containerName,
			Namespace:     namespace,
		},
	}, nil
}

func findPodByLabels(client *kubernetes.Clientset, namespace string, labelSelector string) (*v1.Pod, error) {
	pods, err := client.CoreV1().Pods(namespace).List(
		context.TODO(), // empty context
		metav1.ListOptions{
			LabelSelector: labelSelector, // labelSelector for pod from config
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

	// Search need container by name on pod containers list
	for _, container := range pod.Spec.Containers {
		if container.Name == containerName {
			foundContainers = append(foundContainers, container)
		}
	}

	// When not found container
	if len(foundContainers) < 1 {
		errMsg := fmt.Sprintf("Container %s in pod %s/%s not found!", containerName, pod.Namespace, pod.Name)

		return nil, errors.New(errMsg)
	}
	foundContainer := foundContainers[0]

	return &foundContainer, nil
}

func (kj *KubeJob) GetPod() (*v1.Pod, error) {
	pod, err := findPodByLabels(kj.Client, kj.PodSelector.Namespace, kj.PodSelector.LabelSelector)
	if err != nil {
		return nil, err
	}

	return pod, nil
}

func (kj *KubeJob) GetContainerFromPod(pod *v1.Pod) (*v1.Container, error) {
	// Get need container in pod
	container, err := findContainerByName(pod, kj.PodSelector.ContainerName)
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (kj *KubeJob) Exec(command string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	pod, err := kj.GetPod()
	if err != nil {
		return err
	}

	container, err := kj.GetContainerFromPod(pod)
	if err != nil {
		return err
	}

	req := kj.Client.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(pod.Name).
		Namespace(pod.Namespace).
		SubResource("exec")
	scheme := runtime.NewScheme()
	err = v1.AddToScheme(scheme)
	if err != nil {
		return err
	}

	parameterCodec := runtime.NewParameterCodec(scheme)
	req.VersionedParams(&v1.PodExecOptions{
		Command: []string{
			"sh",
			"-c",
			command,
		},
		Container: container.Name,
		Stdin:     stdin != nil,
		Stdout:    true,
		Stderr:    true,
		TTY:       false,
	}, parameterCodec)

	// Execute over remotecommand
	exec, err := remotecommand.NewSPDYExecutor(kj.KubeConfig, "POST", req.URL())
	if err != nil {
		return err
	}

	// Write container stream to buffers
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
		Tty:    false,
	})
	if err != nil {
		return err
	}

	return nil
}
