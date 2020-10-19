/*
Copyright 2020 The Knative Authors

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

package ingress

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	network "knative.dev/networking/pkg"
	"knative.dev/networking/pkg/apis/networking/v1alpha1"
	netv1alpha1 "knative.dev/networking/pkg/apis/networking/v1alpha1"
)

func TestMakeNewIngress(t *testing.T) {

	ing := &netv1alpha1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-ingress",
			Namespace: "namespace",
			Annotations: map[string]string{
				"networking.knative.dev/ingress.class": "async.ingress.networking.knative.dev",
			},
		},
		Spec: netv1alpha1.IngressSpec{
			Rules: []netv1alpha1.IngressRule{{
				Hosts:      []string{"example.com"},
				Visibility: netv1alpha1.IngressVisibilityExternalIP,
				HTTP: &netv1alpha1.HTTPIngressRuleValue{
					Paths: []netv1alpha1.HTTPIngressPath{{
						Splits: []netv1alpha1.IngressBackendSplit{{
							Percent: 100,
							AppendHeaders: map[string]string{
								network.OriginalHostHeader: "test.com",
							},
							IngressBackend: netv1alpha1.IngressBackend{
								ServiceName:      "servicename",
								ServiceNamespace: "namespace",
								ServicePort:      intstr.FromInt(80),
							},
						}},
					}},
				},
			}},
		},
	}

	got := makeNewIngress(ing, "async.ingress.networking.knative.dev")

	booltrue := true
	want := &netv1alpha1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-ingress-new",
			Namespace: "namespace",
			Annotations: map[string]string{
				"networking.knative.dev/ingress.class": "async.ingress.networking.knative.dev",
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         "networking.internal.knative.dev/v1alpha1",
					Kind:               "Ingress",
					Name:               "test-ingress",
					Controller:         &booltrue,
					BlockOwnerDeletion: &booltrue,
				},
			},
		},
		Spec: netv1alpha1.IngressSpec{
			Rules: []netv1alpha1.IngressRule{{
				Hosts:      []string{"example.com"},
				Visibility: netv1alpha1.IngressVisibilityExternalIP,
				HTTP: &netv1alpha1.HTTPIngressRuleValue{
					Paths: []netv1alpha1.HTTPIngressPath{
						{
							Headers: map[string]v1alpha1.HeaderMatch{"Prefer": {Exact: "respond-async"}},
							Splits: []v1alpha1.IngressBackendSplit{
								{
									IngressBackend: netv1alpha1.IngressBackend{
										ServiceName:      "producer-service",
										ServiceNamespace: "default",
										ServicePort:      intstr.FromInt(80),
									},
									Percent: 100}},
						},
						{
							Splits: []netv1alpha1.IngressBackendSplit{{
								Percent: 100,
								AppendHeaders: map[string]string{
									network.OriginalHostHeader: "test.com",
								},
								IngressBackend: netv1alpha1.IngressBackend{
									ServiceName:      "servicename",
									ServiceNamespace: "namespace",
									ServicePort:      intstr.FromInt(80),
								},
							}},
						}},
				},
			}},
		},
	}

	if !cmp.Equal(want, got) {
		t.Errorf("Unexpected Ingress (-want, +got):\n%s", cmp.Diff(want, got))
	}
}
