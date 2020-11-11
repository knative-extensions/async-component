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

package service

import (
	"context"
	"testing"

	"knative.dev/net-contour/pkg/reconciler/contour/config"
	fakenetworkingclient "knative.dev/networking/pkg/client/injection/client/fake"

	netv1alpha1 "knative.dev/networking/pkg/apis/networking/v1alpha1"
	ingressreconciler "knative.dev/networking/pkg/client/injection/reconciler/networking/v1alpha1/ingress"
	kubeclient "knative.dev/pkg/client/injection/kube/client"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	. "knative.dev/async-component/pkg/reconciler/testing"
	network "knative.dev/networking/pkg"
	"knative.dev/networking/pkg/apis/networking"
	"knative.dev/networking/pkg/apis/networking/v1alpha1"
	. "knative.dev/pkg/reconciler/testing"
)

type testConfigStore struct {
	config *config.Config
}

var svc = service("default", "testing")
var ing = ingress("default", "testing", withAnnotations(map[string]string{networking.IngressClassAnnotationKey: asyncIngressClassName}))

// var createdIng = ingress("default", "test-ingress-new", withAnnotations(map[string]string{networking.IngressClassAnnotationKey: network.IstioIngressClassName}), withHeaderPaths())

// func TestMakeNewIngress(t *testing.T) {
// 	got := makeNewIngress(ing, network.IstioIngressClassName)
// 	want := createdIng
// 	if !cmp.Equal(want, got) {
// 		t.Errorf("Unexpected Ingress (-want, +got):\n%s", cmp.Diff(want, got))
// 	}
// }

// func TestMarkIngressReady(t *testing.T) {
// 	markIngressReady(ing)
// 	got := ing.Status.Conditions
// 	if got == nil {
// 		t.Fatal("Expected Conditions to return a non-nil value")
// 	}
// }

func TestReconcile(t *testing.T) {
	ing.Status.InitializeConditions()
	table := TableTest{{
		Name: "skip ingress not matching class key",
		Objects: []runtime.Object{
			ingress("testing", "testing", withAnnotations(
				map[string]string{networking.IngressClassAnnotationKey: "fake-class-annotation"})),
		},
	}, {
		Name: "create new service",
		Key:  "default/testing",
		Objects: []runtime.Object{
			ing,
		},
		WantCreates: []runtime.Object{
			svc,
		},
	}, {
		Name: "service already exists",
		Key:  "default/testing",
		Objects: []runtime.Object{
			ing,
			svc,
		},
	}}

	table.Test(t, MakeFactory(func(ctx context.Context, listers *Listers, cmw configmap.Watcher) controller.Reconciler {
		r := &Reconciler{
			serviceLister: listers.GetK8sServiceLister(),
			kubeclient:    kubeclient.Get(ctx),
		}
		return ingressreconciler.NewReconciler(ctx, logging.FromContext(ctx), fakenetworkingclient.Get(ctx),
			listers.GetIngressLister(), controller.GetEventRecorder(ctx), r, asyncIngressClassName, controller.Options{})
	}))
}

func service(namespace, name string) *corev1.Service {
	selector := make(map[string]string)
	selector["app"] = producerServiceName
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name + asyncSuffix,
			Namespace: namespace,
		},
		Spec: corev1.ServiceSpec{
			Type:         "ExternalName",
			ExternalName: producerServiceName + ".knative-serving.svc.cluster.local",
			Ports: []corev1.ServicePort{{
				Name:       networking.ServicePortName(networking.ProtocolHTTP1),
				Protocol:   corev1.ProtocolTCP,
				Port:       int32(networking.ServicePort(networking.ProtocolHTTP1)),
				TargetPort: intstr.FromInt(80),
			}},
			Selector: selector,
		},
	}
	return svc
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
