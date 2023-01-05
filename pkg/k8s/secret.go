// Package k8s provides abstracted access to kubernetes APIs.
package k8s

import (
	"context"
	"fmt"

	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/aws/amazon-eks-connector/pkg/config"
)

type Secret interface {
	Put(map[string][]byte) error
	Get() (map[string][]byte, error)
}

// NewSecretInCluster creates a Secret that is suitable for eks-connector pods
// based on stateConfig.
// The secret will be accessed using in-cluster Kubernetes credentials
// and suffixed with pod ordinal index in StatefulSet to avoid conflicts.
func NewSecretInCluster(stateConfig *config.StateConfig) (Secret, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	podIndexProvider := NewPodIndexProvider()
	podIndex, err := podIndexProvider.Get()
	if err != nil {
		return nil, err
	}
	secretName := fmt.Sprintf("%s-%s", stateConfig.SecretNamePrefix, podIndex)

	return NewSecret(secretName, stateConfig.SecretNamespace, k8sClient), nil
}

func NewSecret(name, namespace string, clientset kubernetes.Interface) Secret {
	return &k8sSecret{
		k8s:       clientset,
		namespace: namespace,
		name:      name,
	}
}

type k8sSecret struct {
	k8s kubernetes.Interface

	namespace string
	name      string
}

func (secret *k8sSecret) Get() (map[string][]byte, error) {
	secretV1, err := secret.k8s.CoreV1().Secrets(secret.namespace).Get(context.TODO(), secret.name, metaV1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		} else {
			return nil, err
		}
	}
	return secretV1.Data, nil
}

func (secret *k8sSecret) Put(data map[string][]byte) error {
	secretV1, err := secret.k8s.CoreV1().Secrets(secret.namespace).Get(context.TODO(), secret.name, metaV1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			// secret not found, create

			secretV1 = &coreV1.Secret{
				TypeMeta: metaV1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Secret",
				},
				ObjectMeta: metaV1.ObjectMeta{
					Name:      secret.name,
					Namespace: secret.namespace,
					Labels: map[string]string{
						"name": secret.name,
					},
				},
				Type: coreV1.SecretTypeOpaque,
				Data: data,
			}

			_, err = secret.k8s.CoreV1().Secrets(secret.namespace).Create(context.TODO(), secretV1, metaV1.CreateOptions{})

			return err
		} else {
			return err
		}
	} else {
		secretV1.Data = data
		_, err = secret.k8s.CoreV1().Secrets(secret.namespace).Update(context.TODO(), secretV1, metaV1.UpdateOptions{})

		return err
	}
}
