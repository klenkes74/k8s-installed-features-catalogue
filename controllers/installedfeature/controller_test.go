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

package installedfeature_test

import (
	"errors"
	"github.com/golang/mock/gomock"
	. "github.com/klenkes74/k8s-installed-features-catalogue/api/v1alpha1"
	. "github.com/klenkes74/k8s-installed-features-catalogue/controllers/installedfeature"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pborman/uuid"
	errors2 "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
	// +kubebuilder:scaffold:imports
)

var _ = Describe("InstalledFeature controller", func() {
	const (
		group       = "basic-library"
		name        = "basic-feature"
		otherName   = "other-feature"
		namespace   = "default"
		version     = "1.0.0-alpha1"
		provider    = "Kaiserpfalz EDV-Service"
		description = "a basic demonstration feature"
		uri         = "https://www.kaiserpfalz-edv.de/k8s/"
	)
	var (
		iftLookupKey        = types.NamespacedName{Name: name, Namespace: namespace}
		iftReconcileRequest = reconcile.Request{
			NamespacedName: iftLookupKey,
		}

		iftgLookupKey = types.NamespacedName{Name: group, Namespace: namespace}
	)

	Context("When installing a InstalledFeature CR", func() {
		It("should be created when there are no conflicting features installed and all dependencies met", func() {
			By("By creating a new InstalledFeature")

			ift := createIFT(name, namespace, version, provider, description, uri, true, false)
			client.EXPECT().LoadInstalledFeature(gomock.Any(), iftLookupKey).Return(ift, nil)

			client.EXPECT().GetInstalledFeaturePatchBase(gomock.Any()).Return(k8sclient.MergeFrom(ift))
			client.EXPECT().PatchInstalledFeatureStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			result, err := sut.Reconcile(iftReconcileRequest)

			Expect(result).Should(Equal(reconcile.Result{Requeue: false}))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should add the finalizer when the finalizer is not set", func() {
			By("By creating a new InstalledFeature without finalizer")

			ift := createIFT(name, namespace, version, provider, description, uri, false, false)
			client.EXPECT().LoadInstalledFeature(gomock.Any(), iftLookupKey).Return(ift, nil)

			expected := copyIFT(ift)
			expected.Finalizers = make([]string, 1)
			expected.Finalizers[0] = FinalizerName

			client.EXPECT().SaveInstalledFeature(gomock.Any(), expected).Return(nil)

			client.EXPECT().GetInstalledFeaturePatchBase(gomock.Any()).Return(k8sclient.MergeFrom(ift))
			client.EXPECT().PatchInstalledFeatureStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			result, err := sut.Reconcile(iftReconcileRequest)
			Expect(result).Should(Equal(reconcile.Result{Requeue: false}))
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("Delete an existing InstalledFeature", func() {
		It("should remove the finalizer when the finalizer is set", func() {
			By("By creating a new InstalledFeature without finalizer")

			ift := createIFT(name, namespace, version, provider, description, uri, true, true)
			client.EXPECT().LoadInstalledFeature(gomock.Any(), iftLookupKey).Return(ift, nil)

			expected := copyIFT(ift)
			expected.Finalizers = make([]string, 0)

			client.EXPECT().SaveInstalledFeature(gomock.Any(), expected).Return(nil)

			client.EXPECT().GetInstalledFeaturePatchBase(gomock.Any()).Return(k8sclient.MergeFrom(ift))
			client.EXPECT().PatchInstalledFeatureStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			result, err := sut.Reconcile(iftReconcileRequest)
			Expect(result).Should(Equal(reconcile.Result{Requeue: false}))
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("Handle Library Groups", func() {
		It("should add the status entry on the IFTG when the IFTG has no features yet", func() {
			ift := createIFT(name, namespace, version, provider, description, uri, true, false)
			setGroupToIFT(ift, group, namespace)
			iftg := createIFTG(group, namespace, provider, description, uri, true, false)

			client.EXPECT().LoadInstalledFeature(gomock.Any(), iftLookupKey).Return(ift, nil)

			client.EXPECT().LoadInstalledFeatureGroup(gomock.Any(), iftgLookupKey).Return(iftg, nil)
			client.EXPECT().GetInstalledFeatureGroupPatchBase(gomock.Any()).Return(k8sclient.MergeFrom(iftg))
			client.EXPECT().PatchInstalledFeatureGroupStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			client.EXPECT().GetInstalledFeaturePatchBase(gomock.Any()).Return(k8sclient.MergeFrom(ift))
			client.EXPECT().PatchInstalledFeatureStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			result, err := sut.Reconcile(iftReconcileRequest)

			Expect(result).Should(Equal(reconcile.Result{Requeue: false}))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should add the status entry on the IFTG when the IFTG has already features", func() {
			ift := createIFT(name, namespace, version, provider, description, uri, true, false)
			setGroupToIFT(ift, group, namespace)
			iftg := createIFTG(group, namespace, provider, description, uri, true, false)
			iftg.Status.Features = make([]InstalledFeatureGroupListedFeature, 1)
			iftg.Status.Features[0] = InstalledFeatureGroupListedFeature{
				Namespace: namespace,
				Name:      "other-feature",
			}

			client.EXPECT().LoadInstalledFeature(gomock.Any(), iftLookupKey).Return(ift, nil)

			client.EXPECT().LoadInstalledFeatureGroup(gomock.Any(), iftgLookupKey).Return(iftg, nil)
			client.EXPECT().GetInstalledFeatureGroupPatchBase(gomock.Any()).Return(k8sclient.MergeFrom(iftg))
			client.EXPECT().PatchInstalledFeatureGroupStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			client.EXPECT().GetInstalledFeaturePatchBase(gomock.Any()).Return(k8sclient.MergeFrom(ift))
			client.EXPECT().PatchInstalledFeatureStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			result, err := sut.Reconcile(iftReconcileRequest)

			Expect(result).Should(Equal(reconcile.Result{Requeue: false}))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should not add the feature to the status entry on the IFTG when the IFTG already lists this feature", func() {
			ift := createIFT(name, namespace, version, provider, description, uri, true, false)
			setGroupToIFT(ift, group, namespace)
			iftg := createIFTG(group, namespace, provider, description, uri, true, false)
			iftg.Status.Features = make([]InstalledFeatureGroupListedFeature, 1)
			iftg.Status.Features[0] = InstalledFeatureGroupListedFeature{
				Namespace: namespace,
				Name:      name,
			}

			client.EXPECT().LoadInstalledFeature(gomock.Any(), iftLookupKey).Return(ift, nil)

			client.EXPECT().LoadInstalledFeatureGroup(gomock.Any(), iftgLookupKey).Return(iftg, nil)
			client.EXPECT().GetInstalledFeatureGroupPatchBase(gomock.Any()).Return(k8sclient.MergeFrom(iftg))

			client.EXPECT().GetInstalledFeaturePatchBase(gomock.Any()).Return(k8sclient.MergeFrom(ift))
			client.EXPECT().PatchInstalledFeatureStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			result, err := sut.Reconcile(iftReconcileRequest)

			Expect(result).Should(Equal(reconcile.Result{Requeue: false}))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should remove the status entry on the IFTG when IFT is deleted and is listed in IFTG", func() {
			ift := createIFT(name, namespace, version, provider, description, uri, false, true)
			setGroupToIFT(ift, group, namespace)
			iftg := createIFTG(group, namespace, provider, description, uri, true, false)
			iftg.Status.Features = make([]InstalledFeatureGroupListedFeature, 1)
			iftg.Status.Features[0] = InstalledFeatureGroupListedFeature{
				Namespace: namespace,
				Name:      name,
			}

			client.EXPECT().LoadInstalledFeature(gomock.Any(), iftLookupKey).Return(ift, nil)

			client.EXPECT().LoadInstalledFeatureGroup(gomock.Any(), iftgLookupKey).Return(iftg, nil)
			client.EXPECT().GetInstalledFeatureGroupPatchBase(gomock.Any()).Return(k8sclient.MergeFrom(iftg))
			client.EXPECT().PatchInstalledFeatureGroupStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			client.EXPECT().GetInstalledFeaturePatchBase(gomock.Any()).Return(k8sclient.MergeFrom(ift))
			client.EXPECT().PatchInstalledFeatureStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			result, err := sut.Reconcile(iftReconcileRequest)

			Expect(result).Should(Equal(reconcile.Result{Requeue: false}))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should requeue the request when IFTG can't be loaded", func() {
			ift := createIFT(name, namespace, version, provider, description, uri, false, true)
			setGroupToIFT(ift, group, namespace)

			client.EXPECT().LoadInstalledFeature(gomock.Any(), iftLookupKey).Return(ift, nil)

			client.EXPECT().LoadInstalledFeatureGroup(gomock.Any(), iftgLookupKey).Return(nil, errors.New("can not load IFTG"))

			result, err := sut.Reconcile(iftReconcileRequest)

			Expect(result).Should(Equal(reconcile.Result{RequeueAfter: 60}))
			Expect(err).Should(HaveOccurred())
		})
	})

	Context("Handling dependencies", func() {
		It("Should add dependency status when there is a dependency defined that has already other depending features", func() {
			ift := createIFT(name, namespace, version, provider, description, uri, true, false)
			ift.Spec.DependsOn = []InstalledFeatureDependency{
				{
					Feature: InstalledFeatureRef{
						Namespace: namespace,
						Name:      otherName,
					},
				},
			}

			By("Loading and saving the feature", func() {
				client.EXPECT().LoadInstalledFeature(gomock.Any(), iftLookupKey).Return(ift, nil)

				ift.Status.MissingDependencies = ift.Spec.DependsOn

				iftPatch := k8sclient.MergeFrom(ift)
				client.EXPECT().GetInstalledFeaturePatchBase(ift).Return(iftPatch)
				client.EXPECT().GetInstalledFeaturePatchBase(ift).Return(iftPatch)

				client.EXPECT().PatchInstalledFeatureStatus(gomock.Any(), gomock.Any(), iftPatch).Return(nil).Times(2)
			})

			other := createIFT(otherName, namespace, version, provider, description, uri, true, false)
			other.Status.DependingFeatures = []InstalledFeatureRef{
				{
					Namespace: "other",
					Name:      "other",
				},
			}
			By("Updating the dependent list in the status of the dependency", func() {
				client.EXPECT().LoadInstalledFeature(gomock.Any(), types.NamespacedName{Name: otherName, Namespace: namespace}).Return(other, nil)

				iftPatch := k8sclient.MergeFrom(other)
				client.EXPECT().GetInstalledFeaturePatchBase(other).Return(iftPatch)
				client.EXPECT().PatchInstalledFeatureStatus(gomock.Any(), gomock.Any(), iftPatch).Return(nil)
			})

			result, err := sut.Reconcile(iftReconcileRequest)

			Expect(result).Should(Equal(reconcile.Result{}))
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("Should add dependency status when the dependency is already listed in dependency status", func() {
			ift := createIFT(name, namespace, version, provider, description, uri, true, false)
			ift.Spec.DependsOn = []InstalledFeatureDependency{
				{
					Feature: InstalledFeatureRef{
						Namespace: namespace,
						Name:      otherName,
					},
				},
			}

			By("Loading and saving the feature", func() {
				client.EXPECT().LoadInstalledFeature(gomock.Any(), iftLookupKey).Return(ift, nil)

				iftPatch := k8sclient.MergeFrom(ift)
				client.EXPECT().GetInstalledFeaturePatchBase(ift).Return(iftPatch).Times(2)
				client.EXPECT().PatchInstalledFeatureStatus(gomock.Any(), gomock.Any(), iftPatch).Return(nil).Times(2)
			})

			other := createIFT(otherName, namespace, version, provider, description, uri, true, false)
			other.Status.DependingFeatures = []InstalledFeatureRef{
				{
					Namespace: namespace,
					Name:      name,
				},
			}
			By("Updating the dependent list in the status of the dependency", func() {
				client.EXPECT().LoadInstalledFeature(gomock.Any(), types.NamespacedName{Name: otherName, Namespace: namespace}).Return(other, nil)

				iftPatch := k8sclient.MergeFrom(other)
				client.EXPECT().GetInstalledFeaturePatchBase(other).Return(iftPatch)
			})

			result, err := sut.Reconcile(iftReconcileRequest)

			Expect(result).Should(Equal(reconcile.Result{}))
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("Should remove dependency status when the instance is deleted and already listed in the status of dependency", func() {
			ift := createIFT(name, namespace, version, provider, description, uri, true, true)
			ift.Spec.DependsOn = []InstalledFeatureDependency{
				{
					Feature: InstalledFeatureRef{
						Namespace: namespace,
						Name:      otherName,
					},
				},
			}

			By("Loading and saving the feature", func() {
				client.EXPECT().LoadInstalledFeature(gomock.Any(), iftLookupKey).Return(ift, nil)

				client.EXPECT().SaveInstalledFeature(gomock.Any(), ift).Return(nil)

				iftPatch := k8sclient.MergeFrom(ift)
				client.EXPECT().GetInstalledFeaturePatchBase(ift).Return(iftPatch).Times(2)
				client.EXPECT().PatchInstalledFeatureStatus(gomock.Any(), gomock.Any(), iftPatch).Return(nil).Times(2)
			})

			other := createIFT(otherName, namespace, version, provider, description, uri, true, false)
			other.Status.DependingFeatures = []InstalledFeatureRef{
				{
					Namespace: namespace,
					Name:      name,
				},
			}
			By("Updating the dependent list in the status of the dependency", func() {
				client.EXPECT().LoadInstalledFeature(gomock.Any(), types.NamespacedName{Name: otherName, Namespace: namespace}).Return(other, nil)

				iftPatch := k8sclient.MergeFrom(other)
				client.EXPECT().GetInstalledFeaturePatchBase(other).Return(iftPatch)
				client.EXPECT().PatchInstalledFeatureStatus(gomock.Any(), gomock.Any(), iftPatch).Return(nil)
			})

			result, err := sut.Reconcile(iftReconcileRequest)

			Expect(result).Should(Equal(reconcile.Result{}))
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("Should add dependency status when there is a dependency defined that has no other dependencies", func() {
			ift := createIFT(name, namespace, version, provider, description, uri, true, false)

			ift.Spec.DependsOn = []InstalledFeatureDependency{
				{
					Feature: InstalledFeatureRef{
						Namespace: namespace,
						Name:      otherName,
					},
				},
			}
			By("Loading and saving the feature", func() {
				client.EXPECT().LoadInstalledFeature(gomock.Any(), iftLookupKey).Return(ift, nil)

				iftPatch := k8sclient.MergeFrom(ift)
				client.EXPECT().GetInstalledFeaturePatchBase(ift).Return(iftPatch).Times(2)
				client.EXPECT().PatchInstalledFeatureStatus(gomock.Any(), gomock.Any(), iftPatch).Return(nil).Times(2)
			})

			other := createIFT(otherName, namespace, version, provider, description, uri, true, false)
			By("Updating the dependent list in the status of the dependency", func() {
				client.EXPECT().LoadInstalledFeature(gomock.Any(), types.NamespacedName{Name: otherName, Namespace: namespace}).Return(other, nil)

				iftPatch := k8sclient.MergeFrom(ift)
				client.EXPECT().GetInstalledFeaturePatchBase(other).Return(iftPatch)
				client.EXPECT().PatchInstalledFeatureStatus(gomock.Any(), gomock.Any(), iftPatch).Return(nil)
			})

			result, err := sut.Reconcile(iftReconcileRequest)

			Expect(result).Should(Equal(reconcile.Result{}))
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("Should requeue the reconcile when the dependency status can not be changed", func() {
			ift := createIFT(name, namespace, version, provider, description, uri, true, false)
			ift.Spec.DependsOn = []InstalledFeatureDependency{
				{
					Feature: InstalledFeatureRef{
						Namespace: namespace,
						Name:      otherName,
					},
				},
			}

			By("Loading and saving the feature", func() {
				client.EXPECT().LoadInstalledFeature(gomock.Any(), iftLookupKey).Return(ift, nil)

				ift.Status.MissingDependencies = ift.Spec.DependsOn

				iftPatch := k8sclient.MergeFrom(ift)
				client.EXPECT().GetInstalledFeaturePatchBase(ift).Return(iftPatch)
				client.EXPECT().GetInstalledFeaturePatchBase(ift).Return(iftPatch)

				client.EXPECT().PatchInstalledFeatureStatus(gomock.Any(), gomock.Any(), iftPatch).Return(nil)
				client.EXPECT().PatchInstalledFeatureStatus(gomock.Any(), gomock.Any(), iftPatch).Return(errors.New("could not update status"))
			})

			other := createIFT(otherName, namespace, version, provider, description, uri, true, false)
			other.Status.DependingFeatures = []InstalledFeatureRef{
				{
					Namespace: "other",
					Name:      "other",
				},
			}
			By("Updating the dependent list in the status of the dependency", func() {
				client.EXPECT().LoadInstalledFeature(gomock.Any(), types.NamespacedName{Name: otherName, Namespace: namespace}).Return(other, nil)

				iftPatch := k8sclient.MergeFrom(other)
				client.EXPECT().GetInstalledFeaturePatchBase(other).Return(iftPatch)
				client.EXPECT().PatchInstalledFeatureStatus(gomock.Any(), gomock.Any(), iftPatch).Return(nil)
			})

			result, err := sut.Reconcile(iftReconcileRequest)

			Expect(result).Should(Equal(reconcile.Result{RequeueAfter: 60}))
			Expect(err).Should(HaveOccurred())
		})

		It("Should requeue the reconcile when the instance status can not be changed", func() {
			ift := createIFT(name, namespace, version, provider, description, uri, true, false)

			ift.Spec.DependsOn = []InstalledFeatureDependency{
				{
					Feature: InstalledFeatureRef{
						Namespace: namespace,
						Name:      otherName,
					},
				},
			}
			By("Loading and saving the feature", func() {
				client.EXPECT().LoadInstalledFeature(gomock.Any(), iftLookupKey).Return(ift, nil)

				ift.Status.MissingDependencies = ift.Spec.DependsOn

				iftPatch := k8sclient.MergeFrom(ift)
				client.EXPECT().GetInstalledFeaturePatchBase(ift).Return(iftPatch)
				client.EXPECT().PatchInstalledFeatureStatus(gomock.Any(), gomock.Any(), iftPatch).Return(errors.New("dependency status can not be patched"))
			})

			other := createIFT(otherName, namespace, version, provider, description, uri, true, false)
			By("Updating the dependent list in the status of the dependency", func() {
				client.EXPECT().LoadInstalledFeature(gomock.Any(), types.NamespacedName{Name: otherName, Namespace: namespace}).Return(other, nil)

				iftPatch := k8sclient.MergeFrom(ift)
				client.EXPECT().GetInstalledFeaturePatchBase(other).Return(iftPatch)
			})

			result, err := sut.Reconcile(iftReconcileRequest)

			Expect(result).Should(Equal(reconcile.Result{RequeueAfter: 60}))
			Expect(err).Should(HaveOccurred())
		})

		It("Should mark missing dependency when dependency is marked as deleted", func() {
			ift := createIFT(name, namespace, version, provider, description, uri, true, false)

			ift.Spec.DependsOn = []InstalledFeatureDependency{
				{
					Feature: InstalledFeatureRef{
						Namespace: namespace,
						Name:      otherName,
					},
				},
			}
			By("Loading and saving the feature", func() {
				client.EXPECT().LoadInstalledFeature(gomock.Any(), iftLookupKey).Return(ift, nil)

				iftPatch := k8sclient.MergeFrom(ift)

				ift.Status.MissingDependencies = []InstalledFeatureDependency{
					ift.Spec.DependsOn[0],
				}

				client.EXPECT().GetInstalledFeaturePatchBase(ift).Return(iftPatch)
				client.EXPECT().PatchInstalledFeatureStatus(gomock.Any(), gomock.Any(), iftPatch).Return(nil)
			})

			other := createIFT(otherName, namespace, version, provider, description, uri, true, true)
			By("Updating the dependent list in the status of the dependency", func() {
				client.EXPECT().LoadInstalledFeature(gomock.Any(), types.NamespacedName{Name: otherName, Namespace: namespace}).Return(other, nil)

			})

			result, err := sut.Reconcile(iftReconcileRequest)

			Expect(result).Should(Equal(reconcile.Result{RequeueAfter: 60}))
			Expect(err).Should(HaveOccurred())
		})

		It("Should mark missing dependency when dependency can not be loaded", func() {
			ift := createIFT(name, namespace, version, provider, description, uri, true, false)

			ift.Spec.DependsOn = []InstalledFeatureDependency{
				{
					Feature: InstalledFeatureRef{
						Namespace: namespace,
						Name:      otherName,
					},
				},
			}
			By("Loading and saving the feature", func() {
				client.EXPECT().LoadInstalledFeature(gomock.Any(), iftLookupKey).Return(ift, nil)

				iftPatch := k8sclient.MergeFrom(ift)
				client.EXPECT().GetInstalledFeaturePatchBase(ift).Return(iftPatch)
				client.EXPECT().PatchInstalledFeatureStatus(gomock.Any(), gomock.Any(), iftPatch).Return(nil)
			})

			By("Updating the dependent list in the status of the dependency", func() {
				client.EXPECT().LoadInstalledFeature(gomock.Any(), types.NamespacedName{Name: otherName, Namespace: namespace}).Return(nil, errors.New("other feature not found"))
			})

			result, err := sut.Reconcile(iftReconcileRequest)

			Expect(result).Should(Equal(reconcile.Result{RequeueAfter: 60}))
			Expect(err).Should(HaveOccurred())
		})
	})

	Context("Technical handling", func() {
		It("should drop the request when the ift can't be loaded due to NotFoundError", func() {
			By("By having a problem loading the ift")

			client.EXPECT().LoadInstalledFeature(gomock.Any(), iftLookupKey).Return(nil, errors.New("some error"))

			result, err := sut.Reconcile(iftReconcileRequest)
			Expect(result).Should(Equal(reconcile.Result{RequeueAfter: 60}))
			Expect(err).To(HaveOccurred())
		})

		It("should requeue request when the ift can't be loaded due to another error but NotFoundError", func() {
			client.EXPECT().LoadInstalledFeature(gomock.Any(), iftLookupKey).Return(nil, errors2.NewNotFound(schema.GroupResource{
				Group:    "features.kaiserpfalz-edv.de",
				Resource: "installedfeatures",
			}, iftLookupKey.Name))

			result, err := sut.Reconcile(iftReconcileRequest)
			Expect(result).Should(Equal(reconcile.Result{Requeue: false}))
			Expect(err).To(HaveOccurred())
		})

		It("should requeue request when writing the reconciled object fails", func() {
			By("By getting a failure while saving the data back into the k8s cluster")

			ift := createIFT(name, namespace, version, provider, description, uri, false, false)
			client.EXPECT().LoadInstalledFeature(gomock.Any(), iftLookupKey).Return(ift, nil)

			expected := copyIFT(ift)
			expected.Finalizers = make([]string, 1)
			expected.Finalizers[0] = FinalizerName

			client.EXPECT().SaveInstalledFeature(gomock.Any(), expected).Return(errors.New("some error"))

			result, err := sut.Reconcile(iftReconcileRequest)
			Expect(result).Should(Equal(reconcile.Result{RequeueAfter: 60}))
			Expect(err).To(HaveOccurred())

		})

		It("should requeue the request when updating the status fails", func() {
			By("By getting an error when updating the status")

			ift := createIFT(name, namespace, version, provider, description, uri, false, false)
			client.EXPECT().LoadInstalledFeature(gomock.Any(), iftLookupKey).Return(ift, nil)

			expected := copyIFT(ift)
			expected.Finalizers = make([]string, 1)
			expected.Finalizers[0] = FinalizerName

			client.EXPECT().SaveInstalledFeature(gomock.Any(), expected).Return(nil)

			client.EXPECT().GetInstalledFeaturePatchBase(gomock.Any()).Return(k8sclient.MergeFrom(expected))
			client.EXPECT().PatchInstalledFeatureStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("patching status failed"))

			result, err := sut.Reconcile(iftReconcileRequest)

			Expect(result).Should(Equal(reconcile.Result{RequeueAfter: 60}))
			Expect(err).To(HaveOccurred())
		})
	})
})

func createIFT(name string, namespace string, version string, provider string, description string, uri string, finalizer bool, deleted bool) *InstalledFeature {
	result := &InstalledFeature{
		TypeMeta: metav1.TypeMeta{
			Kind:       "InstalledFeature",
			APIVersion: GroupVersion.Group + "/" + GroupVersion.Version,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Namespace:         namespace,
			CreationTimestamp: metav1.Time{Time: time.Now().Add(24 * time.Hour)},
			ResourceVersion:   "1",
			Generation:        0,
			UID:               types.UID(uuid.New()),
		},
		Spec: InstalledFeatureSpec{
			Kind:        name,
			Version:     version,
			Provider:    provider,
			Description: description,
			Uri:         uri,
		},
	}

	if finalizer {
		result.Finalizers = make([]string, 1)
		result.Finalizers[0] = FinalizerName
	}

	if deleted {
		deletionGracePeriod := int64(60)
		result.DeletionGracePeriodSeconds = &deletionGracePeriod
		result.DeletionTimestamp = &metav1.Time{Time: time.Now().Add(2 * time.Minute)}
	}

	return result
}

func copyIFT(orig *InstalledFeature) *InstalledFeature {
	//goland:noinspection GoDeprecation
	result := &InstalledFeature{
		TypeMeta: metav1.TypeMeta{
			Kind:       orig.TypeMeta.Kind,
			APIVersion: orig.TypeMeta.APIVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:                       orig.ObjectMeta.Name,
			GenerateName:               orig.ObjectMeta.GenerateName,
			Namespace:                  orig.ObjectMeta.Namespace,
			SelfLink:                   orig.ObjectMeta.SelfLink,
			UID:                        orig.ObjectMeta.UID,
			ResourceVersion:            orig.ObjectMeta.ResourceVersion,
			Generation:                 orig.ObjectMeta.Generation,
			CreationTimestamp:          orig.ObjectMeta.CreationTimestamp,
			DeletionTimestamp:          orig.ObjectMeta.DeletionTimestamp,
			DeletionGracePeriodSeconds: orig.ObjectMeta.DeletionGracePeriodSeconds,
			ClusterName:                orig.ObjectMeta.ClusterName,
		},
		Spec: InstalledFeatureSpec{
			Kind:        orig.Spec.Kind,
			Version:     orig.Spec.Version,
			Provider:    orig.Spec.Provider,
			Description: orig.Spec.Description,
			Uri:         orig.Spec.Uri,
		},
	}

	if orig.Spec.Group != nil {
		result = setGroupToIFT(result, orig.Spec.Group.Name, orig.Spec.Group.Namespace)
	}

	if len(orig.ObjectMeta.Labels) > 0 {
		result.ObjectMeta.Labels = make(map[string]string)
		for key, value := range orig.ObjectMeta.Labels {
			result.ObjectMeta.Labels[key] = value
		}
	}

	if len(orig.ObjectMeta.Annotations) > 0 {
		result.ObjectMeta.Annotations = make(map[string]string)
		for key, value := range orig.ObjectMeta.Annotations {
			result.ObjectMeta.Annotations[key] = value
		}
	}

	if len(orig.ObjectMeta.Finalizers) > 0 {
		result.ObjectMeta.Finalizers = make([]string, len(orig.ObjectMeta.Finalizers))
		for i, value := range orig.ObjectMeta.Finalizers {
			result.ObjectMeta.Finalizers[i] = value
		}
	}

	if len(orig.ObjectMeta.OwnerReferences) > 0 {
		result.ObjectMeta.OwnerReferences = make([]metav1.OwnerReference, len(orig.ObjectMeta.OwnerReferences))
		for i, r := range orig.ObjectMeta.OwnerReferences {
			result.ObjectMeta.OwnerReferences[i] = metav1.OwnerReference{
				APIVersion:         r.APIVersion,
				Kind:               r.Kind,
				Name:               r.Name,
				UID:                r.UID,
				Controller:         r.Controller,
				BlockOwnerDeletion: r.BlockOwnerDeletion,
			}
		}
	}

	if len(orig.ObjectMeta.ManagedFields) > 0 {
		result.ObjectMeta.ManagedFields = make([]metav1.ManagedFieldsEntry, len(orig.ObjectMeta.ManagedFields))
		for i, r := range orig.ObjectMeta.ManagedFields {
			result.ObjectMeta.ManagedFields[i] = metav1.ManagedFieldsEntry{
				Manager:    r.Manager,
				Operation:  r.Operation,
				APIVersion: r.APIVersion,
				Time:       r.Time,
				FieldsType: r.FieldsType,
				FieldsV1: &metav1.FieldsV1{
					Raw: r.FieldsV1.Raw,
				},
			}
		}
	}

	return result
}

func setGroupToIFT(instance *InstalledFeature, name string, namespace string) *InstalledFeature {
	instance.Spec.Group = &InstalledFeatureRef{
		Namespace: namespace,
		Name:      name,
	}

	return instance
}

func createIFTG(name string, namespace string, provider string, description string, uri string, finalizer bool, deleted bool) *InstalledFeatureGroup {
	result := &InstalledFeatureGroup{
		TypeMeta: metav1.TypeMeta{
			Kind:       "InstalledFeature",
			APIVersion: GroupVersion.Group + "/" + GroupVersion.Version,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Namespace:         namespace,
			CreationTimestamp: metav1.Time{Time: time.Now().Add(24 * time.Hour)},
			ResourceVersion:   "1",
			Generation:        0,
			UID:               types.UID(uuid.New()),
		},
		Spec: InstalledFeatureGroupSpec{
			Provider:    provider,
			Description: description,
			Uri:         uri,
		},
	}

	if finalizer {
		result.Finalizers = make([]string, 1)
		result.Finalizers[0] = FinalizerName
	}

	if deleted {
		deletionGracePeriod := int64(60)
		result.DeletionGracePeriodSeconds = &deletionGracePeriod
		result.DeletionTimestamp = &metav1.Time{Time: time.Now().Add(2 * time.Minute)}
	}

	return result
}
