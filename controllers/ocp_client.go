//go:generate mockgen -copyright_file=../hack/boilerplate.go.txt -destination=../generated/mock_ocp_client.go -package=generated github.com/klenkes74/k8s-installed-features-catalogue/controllers OcpClient

/*
 * Copyright 2020 Kaiserpfalz EDV-Service, Roland T. Lichti.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package controllers

import (
	"context"
	"github.com/klenkes74/k8s-installed-features-catalogue/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type OcpClient interface {
	// LoadInstalledFeature loads the feature defined by the parameter lookup.
	// ctx The context to be used for this resource load
	// lookup The namespaced name for the resource to loead.
	LoadInstalledFeature(ctx context.Context, lookup types.NamespacedName) (*v1alpha1.InstalledFeature, error)
	SaveInstalledFeature(ctx context.Context, instance *v1alpha1.InstalledFeature) error
	GetInstalledFeaturePatchBase(instance *v1alpha1.InstalledFeature) client.Patch
	PatchInstalledFeatureStatus(ctx context.Context, instance *v1alpha1.InstalledFeature, patch client.Patch) error

	LoadInstalledFeatureGroup(ctx context.Context, lookup types.NamespacedName) (*v1alpha1.InstalledFeatureGroup, error)
	SaveInstalledFeatureGroup(ctx context.Context, instance *v1alpha1.InstalledFeatureGroup) error
	GetInstalledFeatureGroupPatchBase(instance *v1alpha1.InstalledFeatureGroup) client.Patch
	PatchInstalledFeatureGroupStatus(ctx context.Context, instance *v1alpha1.InstalledFeatureGroup, patch client.Patch) error

	// InfoEvent sends an informative event for the given instance with a note what happened.
	// instance is the instance the event is generated for.
	// reason is the UpperCamelCase unique reason of this event <Create|Update|Delete>
	// note is the human readable note which will be used as fmt string for the optional list of spaces.
	// args are the variables that are inserted into the note
	InfoEvent(instance runtime.Object, reason, note string, args ...interface{})

	// WarnEvent sends a warning for the given instance with a note what happened.
	// instance is the instance the event is generated for.
	// reason is the UpperCamelCase unique reason of this event <Create|Update|Delete>
	// note is the human readable note which will be used as fmt string for the optional list of spaces.
	// args are the variables that are inserted into the note
	WarnEvent(instance runtime.Object, reason, note string, args ...interface{})
}

var _ OcpClient = &OcpClientProd{}

type OcpClientProd struct {
	Client   client.Client
	Recorder record.EventRecorder
}

func (o OcpClientProd) InfoEvent(instance runtime.Object, reason, note string, args ...interface{}) {
	o.Recorder.Eventf(instance, "Normal", reason, note, args...)
}

func (o OcpClientProd) WarnEvent(instance runtime.Object, reason, note string, args ...interface{}) {
	o.Recorder.Eventf(instance, "Warning", reason, note, args...)
}

func (o OcpClientProd) LoadInstalledFeature(ctx context.Context, lookup types.NamespacedName) (*v1alpha1.InstalledFeature, error) {
	instance := &v1alpha1.InstalledFeature{}

	err := o.Client.Get(ctx, lookup, instance)
	if err != nil {
		return nil, err
	}

	return instance, nil
}

func (o OcpClientProd) SaveInstalledFeature(ctx context.Context, instance *v1alpha1.InstalledFeature) error {
	return o.Client.Update(ctx, instance)
}

func (o OcpClientProd) GetInstalledFeaturePatchBase(instance *v1alpha1.InstalledFeature) client.Patch {
	return client.MergeFrom(instance.DeepCopy())
}

func (o OcpClientProd) PatchInstalledFeatureStatus(ctx context.Context, instance *v1alpha1.InstalledFeature, patch client.Patch) error {
	return o.Client.Status().Patch(ctx, instance, patch)
}

func (o OcpClientProd) LoadInstalledFeatureGroup(ctx context.Context, lookup types.NamespacedName) (*v1alpha1.InstalledFeatureGroup, error) {
	instance := &v1alpha1.InstalledFeatureGroup{}

	err := o.Client.Get(ctx, lookup, instance)
	if err != nil {
		return nil, err
	}

	return instance, nil
}

func (o OcpClientProd) SaveInstalledFeatureGroup(ctx context.Context, instance *v1alpha1.InstalledFeatureGroup) error {
	return o.Client.Update(ctx, instance)
}

func (o OcpClientProd) GetInstalledFeatureGroupPatchBase(instance *v1alpha1.InstalledFeatureGroup) client.Patch {
	return client.MergeFrom(instance.DeepCopy())
}

func (o OcpClientProd) PatchInstalledFeatureGroupStatus(ctx context.Context, instance *v1alpha1.InstalledFeatureGroup, patch client.Patch) error {
	return o.Client.Status().Patch(ctx, instance, patch)
}
