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

package v1alpha1_test

import (
	"context"
	. "github.com/klenkes74/k8s-installed-features-catalogue/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"time"
	// +kubebuilder:scaffold:imports
)

var _ = Describe("InstalledFeature API", func() {
	const (
		groupname      = "basic-libary"
		groupnamespace = "default"
		name           = "basic-feature"
		namespace      = "default"
		version        = "1.0.0-alpha1"
		provider       = "Kaiserpfalz EDV-Service"
		description    = "a basic demonstration feature"
		uri            = "https://www.kaiserpfalz-edv.de/k8s/"

		timeout  = time.Second * 10
		interval = time.Millisecond * 250
	)
	var (
		ift = &InstalledFeature{
			TypeMeta: metav1.TypeMeta{
				Kind:       "InstalledFeature",
				APIVersion: GroupVersion.Group + "/" + GroupVersion.Version,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: InstalledFeatureSpec{
				Group: &InstalledFeatureGroupRef{
					Namespace: groupnamespace,
					Name:      groupname,
				},
				Kind:        name,
				Version:     version,
				Provider:    provider,
				Description: description,
				Uri:         uri,
			},
			Status: InstalledFeatureStatus{
				Phase:   "provisioned",
				Message: "provisioned without problems",
			},
		}
		ctx          = context.Background()
		iftLookupKey = types.NamespacedName{Name: name, Namespace: namespace}
	)

	Context("When installing a InstalledFeature CR", func() {
		It("should be created when there are no conflicting features installed and all dependencies met", func() {
			By("By creating a new InstalledFeature")

			Expect(k8sClient.Create(ctx, ift)).Should(Succeed())

			createdIft := &InstalledFeature{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, iftLookupKey, createdIft)
				if err != nil {
					return false
				}

				return true
			}, timeout, interval).Should(BeTrue())

			Expect(createdIft.Spec.Kind).Should(Equal(name))
			Expect(createdIft.Spec.Version).Should(Equal(version))
		})
	})

	Context("When deleting an existing InstalledFeature", func() {
		It("should be deleted", func() {
			By("By deleting the InstalledFeature named: " + ift.Name)

			Expect(k8sClient.Delete(ctx, ift)).Should(Succeed())

			createdIft := &InstalledFeature{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, iftLookupKey, createdIft)
				if errors.IsNotFound(err) {
					return true
				}

				logf.Log.Info("found ift", "ift", createdIft)

				return false
			}, timeout, interval).Should(BeTrue())
		})
	})
})
