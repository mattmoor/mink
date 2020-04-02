/*
Copyright 2019 The Knative Authors.

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

package v1alpha1

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/apis/duck"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/tracker"

	"github.com/mattmoor/bindings/pkg/cloudsql"
	"github.com/mattmoor/bindings/pkg/github"
	"github.com/mattmoor/bindings/pkg/slack"
	"github.com/mattmoor/bindings/pkg/sql"
	"github.com/mattmoor/bindings/pkg/twitter"
)

const (
	// GithubBindingConditionReady is set when the binding has been applied to the subjects.
	GithubBindingConditionReady = apis.ConditionReady
)

var ghCondSet = apis.NewLivingConditionSet()

// GetGroupVersionKind implements kmeta.OwnerRefable
func (fb *GithubBinding) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("GithubBinding")
}

// GetSubject implements Bindable
func (fb *GithubBinding) GetSubject() tracker.Reference {
	return fb.Spec.Subject
}

// GetBindingStatus implements Bindable
func (fb *GithubBinding) GetBindingStatus() duck.BindableStatus {
	return &fb.Status
}

// SetObservedGeneration implements BindableStatus
func (fbs *GithubBindingStatus) SetObservedGeneration(gen int64) {
	fbs.ObservedGeneration = gen
}

func (fbs *GithubBindingStatus) InitializeConditions() {
	ghCondSet.Manage(fbs).InitializeConditions()
}

func (fbs *GithubBindingStatus) MarkBindingUnavailable(reason, message string) {
	ghCondSet.Manage(fbs).MarkFalse(
		GithubBindingConditionReady, reason, message)
}

func (fbs *GithubBindingStatus) MarkBindingAvailable() {
	ghCondSet.Manage(fbs).MarkTrue(GithubBindingConditionReady)
}

func (fb *GithubBinding) Do(ctx context.Context, ps *duckv1.WithPod) {

	// First undo so that we can just unconditionally append below.
	fb.Undo(ctx, ps)

	// Make sure the PodSpec has a Volume like this:
	volume := corev1.Volume{
		Name: github.VolumeName,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: fb.Spec.Secret.Name,
			},
		},
	}
	ps.Spec.Template.Spec.Volumes = append(ps.Spec.Template.Spec.Volumes, volume)

	// Make sure that each [init]container in the PodSpec has a VolumeMount like this:
	volumeMount := corev1.VolumeMount{
		Name:      github.VolumeName,
		ReadOnly:  true,
		MountPath: github.MountPath,
	}
	spec := ps.Spec.Template.Spec
	for i := range spec.InitContainers {
		spec.InitContainers[i].VolumeMounts = append(spec.InitContainers[i].VolumeMounts, volumeMount)
	}
	for i := range spec.Containers {
		spec.Containers[i].VolumeMounts = append(spec.Containers[i].VolumeMounts, volumeMount)
	}
}

func (fb *GithubBinding) Undo(ctx context.Context, ps *duckv1.WithPod) {
	spec := ps.Spec.Template.Spec

	// Make sure the PodSpec does NOT have the github volume.
	for i, v := range spec.Volumes {
		if v.Name == github.VolumeName {
			ps.Spec.Template.Spec.Volumes = append(spec.Volumes[:i], spec.Volumes[i+1:]...)
			break
		}
	}

	// Make sure that none of the [init]containers have the github volume mount
	for i, c := range spec.InitContainers {
		for j, vm := range c.VolumeMounts {
			if vm.Name == github.VolumeName {
				spec.InitContainers[i].VolumeMounts = append(c.VolumeMounts[:j], c.VolumeMounts[j+1:]...)
				break
			}
		}
	}
	for i, c := range spec.Containers {
		for j, vm := range c.VolumeMounts {
			if vm.Name == github.VolumeName {
				spec.Containers[i].VolumeMounts = append(c.VolumeMounts[:j], c.VolumeMounts[j+1:]...)
				break
			}
		}
	}
}

const (
	// SlackBindingConditionReady is set when the binding has been applied to the subjects.
	SlackBindingConditionReady = apis.ConditionReady
)

var slackCondSet = apis.NewLivingConditionSet()

// GetGroupVersionKind implements kmeta.OwnerRefable
func (fb *SlackBinding) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("SlackBinding")
}

// GetSubject implements Bindable
func (fb *SlackBinding) GetSubject() tracker.Reference {
	return fb.Spec.Subject
}

// GetBindingStatus implements Bindable
func (fb *SlackBinding) GetBindingStatus() duck.BindableStatus {
	return &fb.Status
}

// SetObservedGeneration implements BindableStatus
func (fbs *SlackBindingStatus) SetObservedGeneration(gen int64) {
	fbs.ObservedGeneration = gen
}

func (fbs *SlackBindingStatus) InitializeConditions() {
	slackCondSet.Manage(fbs).InitializeConditions()
}

func (fbs *SlackBindingStatus) MarkBindingUnavailable(reason, message string) {
	slackCondSet.Manage(fbs).MarkFalse(
		SlackBindingConditionReady, reason, message)
}

func (fbs *SlackBindingStatus) MarkBindingAvailable() {
	slackCondSet.Manage(fbs).MarkTrue(SlackBindingConditionReady)
}

func (fb *SlackBinding) Do(ctx context.Context, ps *duckv1.WithPod) {

	// First undo so that we can just unconditionally append below.
	fb.Undo(ctx, ps)

	// Make sure the PodSpec has a Volume like this:
	volume := corev1.Volume{
		Name: slack.VolumeName,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: fb.Spec.Secret.Name,
			},
		},
	}
	ps.Spec.Template.Spec.Volumes = append(ps.Spec.Template.Spec.Volumes, volume)

	// Make sure that each [init]container in the PodSpec has a VolumeMount like this:
	volumeMount := corev1.VolumeMount{
		Name:      slack.VolumeName,
		ReadOnly:  true,
		MountPath: slack.MountPath,
	}
	spec := ps.Spec.Template.Spec
	for i := range spec.InitContainers {
		spec.InitContainers[i].VolumeMounts = append(spec.InitContainers[i].VolumeMounts, volumeMount)
	}
	for i := range spec.Containers {
		spec.Containers[i].VolumeMounts = append(spec.Containers[i].VolumeMounts, volumeMount)
	}
}

func (fb *SlackBinding) Undo(ctx context.Context, ps *duckv1.WithPod) {
	spec := ps.Spec.Template.Spec

	// Make sure the PodSpec does NOT have the slack volume.
	for i, v := range spec.Volumes {
		if v.Name == slack.VolumeName {
			ps.Spec.Template.Spec.Volumes = append(spec.Volumes[:i], spec.Volumes[i+1:]...)
			break
		}
	}

	// Make sure that none of the [init]containers have the slack volume mount
	for i, c := range spec.InitContainers {
		for j, ev := range c.VolumeMounts {
			if ev.Name == slack.VolumeName {
				spec.InitContainers[i].VolumeMounts = append(spec.InitContainers[i].VolumeMounts[:j], spec.InitContainers[i].VolumeMounts[j+1:]...)
				break
			}
		}
	}
	for i, c := range spec.Containers {
		for j, ev := range c.VolumeMounts {
			if ev.Name == slack.VolumeName {
				spec.Containers[i].VolumeMounts = append(spec.Containers[i].VolumeMounts[:j], spec.Containers[i].VolumeMounts[j+1:]...)
				break
			}
		}
	}
}

const (
	// TwitterBindingConditionReady is set when the binding has been applied to the subjects.
	TwitterBindingConditionReady = apis.ConditionReady
)

var twitterCondSet = apis.NewLivingConditionSet()

// GetGroupVersionKind implements kmeta.OwnerRefable
func (fb *TwitterBinding) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("TwitterBinding")
}

// GetSubject implements Bindable
func (fb *TwitterBinding) GetSubject() tracker.Reference {
	return fb.Spec.Subject
}

// GetBindingStatus implements Bindable
func (fb *TwitterBinding) GetBindingStatus() duck.BindableStatus {
	return &fb.Status
}

// SetObservedGeneration implements BindableStatus
func (fbs *TwitterBindingStatus) SetObservedGeneration(gen int64) {
	fbs.ObservedGeneration = gen
}

func (fbs *TwitterBindingStatus) InitializeConditions() {
	twitterCondSet.Manage(fbs).InitializeConditions()
}

func (fbs *TwitterBindingStatus) MarkBindingUnavailable(reason, message string) {
	twitterCondSet.Manage(fbs).MarkFalse(
		TwitterBindingConditionReady, reason, message)
}

func (fbs *TwitterBindingStatus) MarkBindingAvailable() {
	twitterCondSet.Manage(fbs).MarkTrue(TwitterBindingConditionReady)
}

func (fb *TwitterBinding) Do(ctx context.Context, ps *duckv1.WithPod) {

	// First undo so that we can just unconditionally append below.
	fb.Undo(ctx, ps)

	// Make sure the PodSpec has a Volume like this:
	volume := corev1.Volume{
		Name: twitter.VolumeName,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: fb.Spec.Secret.Name,
			},
		},
	}
	ps.Spec.Template.Spec.Volumes = append(ps.Spec.Template.Spec.Volumes, volume)

	// Make sure that each [init]container in the PodSpec has a VolumeMount like this:
	volumeMount := corev1.VolumeMount{
		Name:      twitter.VolumeName,
		ReadOnly:  true,
		MountPath: twitter.MountPath,
	}
	spec := ps.Spec.Template.Spec
	for i := range spec.InitContainers {
		spec.InitContainers[i].VolumeMounts = append(spec.InitContainers[i].VolumeMounts, volumeMount)
	}
	for i := range spec.Containers {
		spec.Containers[i].VolumeMounts = append(spec.Containers[i].VolumeMounts, volumeMount)
	}
}

func (fb *TwitterBinding) Undo(ctx context.Context, ps *duckv1.WithPod) {
	spec := ps.Spec.Template.Spec

	// Make sure the PodSpec does NOT have the twitter volume.
	for i, v := range spec.Volumes {
		if v.Name == twitter.VolumeName {
			ps.Spec.Template.Spec.Volumes = append(spec.Volumes[:i], spec.Volumes[i+1:]...)
			break
		}
	}

	// Make sure that none of the [init]containers have the twitter volume mount
	for i, c := range spec.InitContainers {
		for j, ev := range c.VolumeMounts {
			if ev.Name == twitter.VolumeName {
				spec.InitContainers[i].VolumeMounts = append(spec.InitContainers[i].VolumeMounts[:j], spec.InitContainers[i].VolumeMounts[j+1:]...)
				break
			}
		}
	}
	for i, c := range spec.Containers {
		for j, ev := range c.VolumeMounts {
			if ev.Name == twitter.VolumeName {
				spec.Containers[i].VolumeMounts = append(spec.Containers[i].VolumeMounts[:j], spec.Containers[i].VolumeMounts[j+1:]...)
				break
			}
		}
	}
}

const (
	// GoogleCloudSQLBindingConditionReady is set when the binding has been applied to the subjects.
	GoogleCloudSQLBindingConditionReady = apis.ConditionReady
)

var gcSqlCondSet = apis.NewLivingConditionSet()

// GetGroupVersionKind implements kmeta.OwnerRefable
func (fb *GoogleCloudSQLBinding) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("GoogleCloudSQLBinding")
}

// GetSubject implements Bindable
func (fb *GoogleCloudSQLBinding) GetSubject() tracker.Reference {
	return fb.Spec.Subject
}

// GetBindingStatus implements Bindable
func (fb *GoogleCloudSQLBinding) GetBindingStatus() duck.BindableStatus {
	return &fb.Status
}

// SetObservedGeneration implements BindableStatus
func (fbs *GoogleCloudSQLBindingStatus) SetObservedGeneration(gen int64) {
	fbs.ObservedGeneration = gen
}

func (fbs *GoogleCloudSQLBindingStatus) InitializeConditions() {
	gcSqlCondSet.Manage(fbs).InitializeConditions()
}

func (fbs *GoogleCloudSQLBindingStatus) MarkBindingUnavailable(reason, message string) {
	gcSqlCondSet.Manage(fbs).MarkFalse(
		GoogleCloudSQLBindingConditionReady, reason, message)
}

func (fbs *GoogleCloudSQLBindingStatus) MarkBindingAvailable() {
	gcSqlCondSet.Manage(fbs).MarkTrue(GoogleCloudSQLBindingConditionReady)
}

func (fb *GoogleCloudSQLBinding) Do(ctx context.Context, ps *duckv1.WithPod) {
	// First undo so that we can just unconditionally append below.
	fb.Undo(ctx, ps)

	c := corev1.Container{
		Name:  cloudsql.ContainerName,
		Image: "gcr.io/cloudsql-docker/gce-proxy:1.14",
		Command: []string{
			"/cloud_sql_proxy",
			"-dir=" + cloudsql.SocketMountPath,
			fmt.Sprintf("-instances=%s,%s=tcp:3306", fb.Spec.Instance, fb.Spec.Instance),
			// If running on a VPC, the Cloud SQL proxy can connect via Private IP. See:
			// https://cloud.google.com/sql/docs/mysql/private-ip for more info.
			// "-ip_address_types=PRIVATE",
			fmt.Sprintf("-credential_file=%s/credentials.json", cloudsql.SecretMountPath),
		},
		VolumeMounts: []corev1.VolumeMount{{
			Name:      cloudsql.SecretVolumeName,
			ReadOnly:  true,
			MountPath: cloudsql.SecretMountPath,
		}, {
			Name:      cloudsql.SocketVolumeName,
			MountPath: cloudsql.SocketMountPath,
		}},
	}

	v := []corev1.Volume{{
		Name: cloudsql.SecretVolumeName,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: fb.Spec.Secret.Name,
			},
		},
	}, {
		Name: cloudsql.SocketVolumeName,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}}

	spec := ps.Spec.Template.Spec
	for i := range spec.Containers {
		spec.Containers[i].VolumeMounts = append(spec.Containers[i].VolumeMounts, c.VolumeMounts...)
	}

	ps.Spec.Template.Spec.Containers = append(ps.Spec.Template.Spec.Containers, c)
	ps.Spec.Template.Spec.Volumes = append(ps.Spec.Template.Spec.Volumes, v...)
}

func (fb *GoogleCloudSQLBinding) Undo(ctx context.Context, ps *duckv1.WithPod) {
	spec := ps.Spec.Template.Spec

	// Make sure the PodSpec does NOT have the cloudsql container.
	for i, v := range spec.Containers {
		if v.Name == cloudsql.ContainerName {
			ps.Spec.Template.Spec.Containers = append(spec.Containers[:i], spec.Containers[i+1:]...)
			break
		}
	}

	// Make sure the PodSpec does NOT have the cloudsql volumes.
	vs := make([]corev1.Volume, 0, len(spec.Volumes))
	for _, v := range spec.Volumes {
		if v.Name == cloudsql.SecretVolumeName || v.Name == cloudsql.SocketVolumeName {
			continue
		}
		vs = append(vs, v)
	}
	ps.Spec.Template.Spec.Volumes = vs

	// Make sure that none of the containers have the cloudsql socket volume mount
	for i, c := range spec.Containers {
		vms := make([]corev1.VolumeMount, 0, len(c.VolumeMounts))
		for _, vm := range c.VolumeMounts {
			if vm.Name == cloudsql.SecretVolumeName || vm.Name == cloudsql.SocketVolumeName {
				continue
			}
			vms = append(vms, vm)
		}
		spec.Containers[i].VolumeMounts = vms
	}
}

const (
	// SQLBindingConditionReady is set when the binding has been applied to the subjects.
	SQLBindingConditionReady = apis.ConditionReady
)

var sqlCondSet = apis.NewLivingConditionSet()

// GetGroupVersionKind implements kmeta.OwnerRefable
func (fb *SQLBinding) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("SQLBinding")
}

// GetSubject implements Bindable
func (fb *SQLBinding) GetSubject() tracker.Reference {
	return fb.Spec.Subject
}

// GetBindingStatus implements Bindable
func (fb *SQLBinding) GetBindingStatus() duck.BindableStatus {
	return &fb.Status
}

// SetObservedGeneration implements BindableStatus
func (fbs *SQLBindingStatus) SetObservedGeneration(gen int64) {
	fbs.ObservedGeneration = gen
}

func (fbs *SQLBindingStatus) InitializeConditions() {
	sqlCondSet.Manage(fbs).InitializeConditions()
}

func (fbs *SQLBindingStatus) MarkBindingUnavailable(reason, message string) {
	sqlCondSet.Manage(fbs).MarkFalse(SQLBindingConditionReady, reason, message)
}

func (fbs *SQLBindingStatus) MarkBindingAvailable() {
	sqlCondSet.Manage(fbs).MarkTrue(SQLBindingConditionReady)
}

func (fb *SQLBinding) Do(ctx context.Context, ps *duckv1.WithPod) {
	// First undo so that we can just unconditionally append below.
	fb.Undo(ctx, ps)

	// Make sure the PodSpec has a Volume like this:
	volume := corev1.Volume{
		Name: sql.VolumeName,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: fb.Spec.Secret.Name,
			},
		},
	}
	ps.Spec.Template.Spec.Volumes = append(ps.Spec.Template.Spec.Volumes, volume)

	// Make sure that each [init]container in the PodSpec has a VolumeMount like this:
	volumeMount := corev1.VolumeMount{
		Name:      sql.VolumeName,
		ReadOnly:  true,
		MountPath: sql.MountPath,
	}
	spec := ps.Spec.Template.Spec
	for i := range spec.InitContainers {
		spec.InitContainers[i].VolumeMounts = append(spec.InitContainers[i].VolumeMounts, volumeMount)
	}
	for i := range spec.Containers {
		spec.Containers[i].VolumeMounts = append(spec.Containers[i].VolumeMounts, volumeMount)
	}
}

func (fb *SQLBinding) Undo(ctx context.Context, ps *duckv1.WithPod) {
	spec := ps.Spec.Template.Spec

	// Make sure the PodSpec does NOT have the sql volume.
	for i, v := range spec.Volumes {
		if v.Name == sql.VolumeName {
			ps.Spec.Template.Spec.Volumes = append(spec.Volumes[:i], spec.Volumes[i+1:]...)
			break
		}
	}

	// Make sure that none of the [init]containers have the sql volume mount
	for i, c := range spec.InitContainers {
		for j, ev := range c.VolumeMounts {
			if ev.Name == sql.VolumeName {
				spec.InitContainers[i].VolumeMounts = append(spec.InitContainers[i].VolumeMounts[:j], spec.InitContainers[i].VolumeMounts[j+1:]...)
				break
			}
		}
	}
	for i, c := range spec.Containers {
		for j, ev := range c.VolumeMounts {
			if ev.Name == sql.VolumeName {
				spec.Containers[i].VolumeMounts = append(spec.Containers[i].VolumeMounts[:j], spec.Containers[i].VolumeMounts[j+1:]...)
				break
			}
		}
	}

}
