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
	"context"
	"testing"

	"knative.dev/net-contour/pkg/reconciler/contour/config"
	netclient "knative.dev/networking/pkg/client/injection/client"
	fakenetworkingclient "knative.dev/networking/pkg/client/injection/client/fake"

	ingressreconciler "knative.dev/networking/pkg/client/injection/reconciler/networking/v1alpha1/ingress"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	. "knative.dev/async-component/pkg/reconciler/testing"
	network "knative.dev/networking/pkg"
	"knative.dev/networking/pkg/apis/networking"
	"knative.dev/networking/pkg/apis/networking/v1alpha1"
	netv1alpha1 "knative.dev/networking/pkg/apis/networking/v1alpha1"
	. "knative.dev/pkg/reconciler/testing"
)

type testConfigStore struct {
	config *config.Config
}

var ing = ingress("default", "test-ingress", withAnnotations(map[string]string{networking.IngressClassAnnotationKey: asyncIngressClassName}))
var booltrue = true
var createdIng = &netv1alpha1.Ingress{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "test-ingress-new",
		Namespace: "default",
		Annotations: map[string]string{
			networking.IngressClassAnnotationKey: network.IstioIngressClassName,
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
								ServiceNamespace: "default",
								ServicePort:      intstr.FromInt(80),
							},
						}},
					}},
			},
		}},
	},
}

func TestMakeNewIngress(t *testing.T) {
	got := makeNewIngress(ing, network.IstioIngressClassName)
	want := createdIng
	if !cmp.Equal(want, got) {
		t.Errorf("Unexpected Ingress (-want, +got):\n%s", cmp.Diff(want, got))
	}
}

func TestMarkIngressReady(t *testing.T) {
	markIngressReady(ing)
	c := ing.Status.Conditions
	if c == nil {
		t.Fatal("Expected NewController to return a non-nil value")
	}
}

func TestReconcile(t *testing.T) {
	table := TableTest{{
		Name: "skip ingress not matching class key",
		Objects: []runtime.Object{
			ingress("testing", "testing", withAnnotations(
				map[string]string{networking.IngressClassAnnotationKey: "fake-class-annotation"})),
		},
	},
		{
			Name: "create new ingress",
			Key:  "default/test-ingress",
			Objects: []runtime.Object{
				ing,
			},
			WantCreates: []runtime.Object{
				createdIng,
			},
		}}

	table.Test(t, MakeFactory(func(ctx context.Context, listers *Listers, cmw configmap.Watcher) controller.Reconciler {
		r := &Reconciler{
			netclient:     netclient.Get(ctx),
			ingressLister: listers.GetIngressLister(),
		}
		return ingressreconciler.NewReconciler(ctx, logging.FromContext(ctx), fakenetworkingclient.Get(ctx),
			listers.GetIngressLister(), controller.GetEventRecorder(ctx), r, asyncIngressClassName, controller.Options{})
	}))
}

type ingressCreationOption func(ing *v1alpha1.Ingress)

func ingress(namespace, name string, opt ...ingressCreationOption) *v1alpha1.Ingress {
	ing := &netv1alpha1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
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
								ServiceNamespace: namespace,
								ServicePort:      intstr.FromInt(80),
							},
						}},
					}},
				},
			}},
		},
	}
	for _, o := range opt {
		o(ing)
	}
	return ing
}

func withAnnotations(ans map[string]string) ingressCreationOption {
	return func(ing *v1alpha1.Ingress) {
		ing.Annotations = ans
	}
}
