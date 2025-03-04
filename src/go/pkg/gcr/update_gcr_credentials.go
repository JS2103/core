// Copyright 2019 The Cloud Robotics Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

/*
Library for updating the token used to pull images from GCR in the surrounding cluster.
*/
package gcr

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/googlecloudrobotics/core/src/go/pkg/kubeutils"
	"github.com/googlecloudrobotics/core/src/go/pkg/robotauth"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

// Name of the secret that stores the GCR pull token.
const SecretName = "gcr-json-key"

// DockerCfgJSON takes a service account key, and converts it into the JSON
// format required for k8s's docker-registry secrets.
func DockerCfgJSON(token string) []byte {
	type dockercfg struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
		Auth     []byte `json:"auth"`
	}

	m := map[string]interface{}{}
	for _, r := range []string{"gcr.io", "asia.gcr.io", "eu.gcr.io", "us.gcr.io"} {
		m["https://"+r] = dockercfg{
			Username: "oauth2accesstoken",
			Password: string(token),
			Email:    "not@val.id",
			Auth:     []byte("oauth2accesstoken:" + token),
		}
	}
	b, err := json.Marshal(m)
	if err != nil {
		log.Fatal("unexpected error marshalling dockercfg: ", err)
	}
	return b
}

func patchServiceAccount(ctx context.Context, k8s *kubernetes.Clientset, name string, namespace string, patchData []byte) error {
	sa := k8s.CoreV1().ServiceAccounts(namespace)
	return backoff.Retry(
		func() error {
			_, err := sa.Patch(ctx, name, types.StrategicMergePatchType, patchData, metav1.PatchOptions{})
			if err != nil && !k8serrors.IsNotFound(err) {
				return backoff.Permanent(fmt.Errorf("failed to apply %q: %v", patchData, err))
			}
			return err
		},
		backoff.NewConstantBackOff(time.Second),
	)
}

// UpdateGcrCredentials authenticates to the cloud cluster using the auth config given and updates
// the credentials used to pull images from GCR.
func UpdateGcrCredentials(ctx context.Context, k8s *kubernetes.Clientset, auth *robotauth.RobotAuth) error {
	tokenSource := auth.CreateRobotTokenSource(ctx)
	token, err := tokenSource.Token()
	if err != nil {
		return fmt.Errorf("failed to get token: %v", err)
	}

	nsList, err := k8s.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list namespaces: %v", err)
	}
	cfgData := map[string][]byte{".dockercfg": DockerCfgJSON(token.AccessToken)}
	patchData := []byte(`{"imagePullSecrets": [{"name": "` + SecretName + `"}]}`)
	haveError := false
	for _, ns := range nsList.Items {
		if ns.DeletionTimestamp != nil {
			log.Printf("namespace %q is marked for deletion, skipping", ns.ObjectMeta.Name)
			continue
		}
		namespace := ns.ObjectMeta.Name

		// Only ever create secrets in the 'default' namespace. For app namespaces the ChartAssignment
		// controller will create the initial secret and patch the service account.
		// This avoids us putting pull secrets into eg foreign namespaces.
		s := k8s.CoreV1().Secrets(namespace)
		if _, err := s.Get(ctx, SecretName, metav1.GetOptions{}); k8serrors.IsNotFound(err) {
			if namespace != "default" {
				continue
			}
		}
		// If we get here, the namespace has a secret that we need to update or
		// it is the default namespace where it is okay to create the secret.

		// Create or update a secret containing a docker config with the access-token.
		err = kubeutils.UpdateSecret(ctx, k8s, SecretName, namespace, corev1.SecretTypeDockercfg,
			cfgData)
		if err != nil {
			log.Printf("failed to update kubernetes secret for namespace %s: %v", namespace, err)
			haveError = true
			continue
		}

		// Tell k8s to use this key by pointing the default SA at it.
		err = patchServiceAccount(ctx, k8s, "default", namespace, patchData)
		if err != nil {
			log.Printf("failed to update kubernetes service account for namespace %s: %v", namespace, err)
			haveError = true
		}
	}
	if haveError {
		return fmt.Errorf("failed to update one or more namespaces")
	}
	return nil
}
