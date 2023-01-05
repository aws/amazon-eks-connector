package k8s

import (
	"fmt"
	"os"
	"strings"

	"k8s.io/klog/v2"
)

const (
	EnvPodName = "POD_NAME"
)

// PodIndexProvider provides pod ordinal index inside StatefulSet.
type PodIndexProvider interface {
	Get() (string, error)
}

func NewPodIndexProvider() PodIndexProvider {
	return &statefulSetPodIndexProvider{}
}

type statefulSetPodIndexProvider struct {
}

func (p *statefulSetPodIndexProvider) Get() (string, error) {
	klog.V(2).Infof("extracting pod index in StatefulSet with env %s", EnvPodName)

	pod := os.Getenv(EnvPodName)
	if pod == "" {
		return "", fmt.Errorf("cannot get pod name from env %s", EnvPodName)
	}
	klog.V(2).Infof("env %s = %s", EnvPodName, pod)

	strIndex := strings.LastIndex(pod, "-")
	if strIndex < 0 || strIndex >= len(pod)-1 {
		return "", fmt.Errorf("unexpected pod name %s", pod)
	}

	podIndex := pod[strIndex+1:]
	klog.V(2).Infof("extracted pod index %s", podIndex)

	return podIndex, nil
}
