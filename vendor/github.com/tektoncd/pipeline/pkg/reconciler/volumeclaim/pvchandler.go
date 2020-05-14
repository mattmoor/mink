/*
Copyright 2020 The Tekton Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package volumeclaim

import (
	"fmt"

	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	errorutils "k8s.io/apimachinery/pkg/util/errors"
	clientset "k8s.io/client-go/kubernetes"
)

const (
	// ReasonCouldntCreateWorkspacePVC indicates that a Pipeline expects a workspace from a
	// volumeClaimTemplate but couldn't create a claim.
	ReasonCouldntCreateWorkspacePVC = "CouldntCreateWorkspacePVC"
)

type PvcHandler interface {
	CreatePersistentVolumeClaimsForWorkspaces(wb []v1alpha1.WorkspaceBinding, ownerReference metav1.OwnerReference, namespace string) error
}

type defaultPVCHandler struct {
	clientset clientset.Interface
	logger    *zap.SugaredLogger
}

func NewPVCHandler(clientset clientset.Interface, logger *zap.SugaredLogger) PvcHandler {
	return &defaultPVCHandler{clientset, logger}
}

// CreatePersistentVolumeClaimsForWorkspaces checks if a PVC named <claim-name>-<workspace-name>-<owner-name> exists;
// where claim-name is provided by the user in the volumeClaimTemplate, and owner-name is the name of the
// resource with the volumeClaimTemplate declared, a PipelineRun or TaskRun. If the PVC did not exist, a new PVC
// with that name is created with the provided OwnerReference.
func (c *defaultPVCHandler) CreatePersistentVolumeClaimsForWorkspaces(wb []v1alpha1.WorkspaceBinding, ownerReference metav1.OwnerReference, namespace string) error {
	var errs []error
	for _, claim := range getPersistentVolumeClaims(wb, ownerReference, namespace) {
		_, err := c.clientset.CoreV1().PersistentVolumeClaims(claim.Namespace).Get(claim.Name, metav1.GetOptions{})
		switch {
		case apierrors.IsNotFound(err):
			_, err := c.clientset.CoreV1().PersistentVolumeClaims(claim.Namespace).Create(claim)
			if err != nil {
				errs = append(errs, fmt.Errorf("failed to create PVC %s: %s", claim.Name, err))
			}
			if err == nil || !apierrors.IsAlreadyExists(err) {
				c.logger.Infof("Created PersistentVolumeClaim %s in namespace %s", claim.Name, claim.Namespace)
			}
		case err != nil:
			errs = append(errs, fmt.Errorf("failed to retrieve PVC %s: %s", claim.Name, err))
		}
	}
	return errorutils.NewAggregate(errs)
}

func getPersistentVolumeClaims(workspaceBindings []v1alpha1.WorkspaceBinding, ownerReference metav1.OwnerReference, namespace string) map[string]*corev1.PersistentVolumeClaim {
	claims := make(map[string]*corev1.PersistentVolumeClaim)
	for _, workspaceBinding := range workspaceBindings {
		if workspaceBinding.VolumeClaimTemplate == nil {
			continue
		}

		claim := workspaceBinding.VolumeClaimTemplate.DeepCopy()
		claim.Name = GetPersistentVolumeClaimName(workspaceBinding.VolumeClaimTemplate, workspaceBinding, ownerReference)
		claim.Namespace = namespace
		claim.OwnerReferences = []metav1.OwnerReference{ownerReference}
		claims[workspaceBinding.Name] = claim
	}
	return claims
}

// GetPersistentVolumeClaimName gets the name of PersistentVolumeClaim for a Workspace and PipelineRun or TaskRun. claim
// must be a PersistentVolumeClaim from set's VolumeClaims template.
func GetPersistentVolumeClaimName(claim *corev1.PersistentVolumeClaim, wb v1alpha1.WorkspaceBinding, owner metav1.OwnerReference) string {
	if claim.Name == "" {
		return fmt.Sprintf("%s-%s", wb.Name, owner.Name)
	}
	return fmt.Sprintf("%s-%s-%s", claim.Name, wb.Name, owner.Name)
}
