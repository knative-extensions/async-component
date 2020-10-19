package ingress

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	networkpkg "knative.dev/networking/pkg"
	"knative.dev/networking/pkg/apis/networking"
	"knative.dev/networking/pkg/apis/networking/v1alpha1"
	netclientset "knative.dev/networking/pkg/client/clientset/versioned"
	networkinglisters "knative.dev/networking/pkg/client/listers/networking/v1alpha1"
	"knative.dev/networking/pkg/ingress"
	"knative.dev/pkg/kmeta"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/network"
	"knative.dev/pkg/reconciler"
	"knative.dev/pkg/system"
)

// Reconciler implements controller.Reconciler for Ingress resources.
type Reconciler struct {
	ingressLister networkinglisters.IngressLister
	netclient     netclientset.Interface
}

// ReconcileKind implements Interface.ReconcileKind.
func (r *Reconciler) ReconcileKind(ctx context.Context, ing *v1alpha1.Ingress) reconciler.Event {
	logger := logging.FromContext(ctx)
	// TODO(beemarie): allow this ingress class to be configurable
	ingressClass := networkpkg.IstioIngressClassName

	markIngressReady(ing) //TODO (beemarie): this just sets the status of KIngress, but load balancer isn't needed.

	desired := makeNewIngress(ing, ingressClass)
	_, err := ingress.InsertProbe(desired)
	if err != nil {
		logger.Errorf("failed to add knative probe header in ingress %s", ing.GetName())
		return err
	}
	_, err = r.netclient.NetworkingV1alpha1().Ingresses(desired.Namespace).Create(ctx, desired, metav1.CreateOptions{})
	if err != nil {
		logger.Errorf("error creating ingress %s", desired.Name)
		return err
	}
	_, err = r.reconcileIngress(ctx, ing)
	if err != nil {
		logger.Errorf("error reconciling ingress: %s", ing.Name)
		return err
	}
	// return err
	return nil
}

func (r *Reconciler) reconcileIngress(ctx context.Context, desired *v1alpha1.Ingress) (*v1alpha1.Ingress, error) {
	desired.Status.InitializeConditions()
	ingress, err := r.ingressLister.Ingresses(desired.Namespace).Get(desired.Name)
	if apierrs.IsNotFound(err) {
		ingress, err = r.netclient.NetworkingV1alpha1().Ingresses(desired.Namespace).Create(ctx, desired, metav1.CreateOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create Ingress: %w", err)
		}
		return ingress, nil
	} else if err != nil {
		return nil, err
	} else if !equality.Semantic.DeepEqual(ingress.Spec, desired.Spec) ||
		!equality.Semantic.DeepEqual(ingress.Annotations, desired.Annotations) {
		// Don't modify the informers copy
		origin := ingress.DeepCopy()
		origin.Spec = desired.Spec
		origin.Annotations = desired.Annotations
		updated, err := r.netclient.NetworkingV1alpha1().Ingresses(origin.Namespace).Update(ctx, origin, metav1.UpdateOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to update Ingress: %w", err)
		}
		return updated, nil
	}

	return ingress, err
}

// makeNewIngress creates an Ingress object with respond-async headers pointing to producer-service
func makeNewIngress(ingress *v1alpha1.Ingress, ingressClass string) *v1alpha1.Ingress {
	splits := make([]v1alpha1.IngressBackendSplit, 0, 1)
	splits = append(splits, v1alpha1.IngressBackendSplit{
		IngressBackend: v1alpha1.IngressBackend{
			ServiceName:      "producer-service", // TODO(beemarie): make this configurable
			ServiceNamespace: "default",
			ServicePort:      intstr.FromInt(80),
		},
		Percent: int(100),
	})
	theRules := []v1alpha1.IngressRule{}
	for _, rule := range ingress.Spec.Rules {
		newRule := rule
		newPaths := make([]v1alpha1.HTTPIngressPath, 0)
		newPaths = append(newPaths, v1alpha1.HTTPIngressPath{
			Headers: map[string]v1alpha1.HeaderMatch{"Prefer": {Exact: "respond-async"}},
			Splits:  splits,
		})
		newPaths = append(newPaths, newRule.HTTP.Paths...)
		newRule.HTTP.Paths = newPaths
		theRules = append(theRules, newRule)
	}
	return &v1alpha1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ingress.Name + "-new",
			Namespace: ingress.Namespace,
			Annotations: kmeta.FilterMap(kmeta.UnionMaps(map[string]string{
				networking.IngressClassAnnotationKey: ingressClass,
			}), func(key string) bool {
				return key == corev1.LastAppliedConfigAnnotation
			}),
			Labels:          ingress.Labels,
			OwnerReferences: []metav1.OwnerReference{*kmeta.NewControllerRef(ingress)},
		},
		Spec: v1alpha1.IngressSpec{
			Rules: theRules,
		},
	}
}

func markIngressReady(ingress *v1alpha1.Ingress) {
	internalDomain := domainForServiceName(ingress.Name)
	externalDomain := domainForServiceName(ingress.Name)

	ingress.Status.MarkLoadBalancerReady(
		[]v1alpha1.LoadBalancerIngressStatus{{
			//DomainInternal: "doesitmatter", // TODO:(beemarie) this isn't a "real ingress", what can we put here?
			DomainInternal: externalDomain,
		}},
		[]v1alpha1.LoadBalancerIngressStatus{{
			//DomainInternal: "shouldntmatter", // TODO:(beemarie) this isn't a "real ingress", what can we put here?
			DomainInternal: internalDomain,
		}},
	)
	ingress.Status.MarkNetworkConfigured()
}

func domainForServiceName(serviceName string) string {
	return serviceName + "." + getGatewayNamespace() + ".svc." + network.GetClusterDomainName()
}

func getGatewayNamespace() string {
	namespace := "" //os.Getenv(config.GatewayNamespaceEnv)
	if namespace == "" {
		return system.Namespace()
	}
	return namespace
}
