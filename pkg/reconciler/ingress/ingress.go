package ingress

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	corev1listers "k8s.io/client-go/listers/core/v1"
	networkpkg "knative.dev/networking/pkg"
	"knative.dev/networking/pkg/apis/networking"
	"knative.dev/networking/pkg/apis/networking/v1alpha1"
	netclientset "knative.dev/networking/pkg/client/clientset/versioned"
	networkinglisters "knative.dev/networking/pkg/client/listers/networking/v1alpha1"

	"knative.dev/pkg/kmeta"
	"knative.dev/pkg/logging"
	network "knative.dev/pkg/network"
	"knative.dev/pkg/reconciler"
	"knative.dev/pkg/system"
)

// Reconciler implements controller.Reconciler for Ingress resources.
type Reconciler struct {
	ingressLister networkinglisters.IngressLister
	serviceLister corev1listers.ServiceLister
	netclient     netclientset.Interface
	kubeclient    kubernetes.Interface
}

const (
	AsyncModeAnnotationKey  = "async.knative.dev/mode"
	asyncSuffix             = "-async"
	newSuffix               = "-new"
	preferHeaderField       = "Prefer"
	preferAsyncValue        = "respond-async"
	preferSyncValue         = "respond-sync"
	asyncAlwaysMode         = "always.async.knative.dev"
	asyncConditionalMode    = "conditional.async.knative.dev"
	publicLBDomain          = "istio-ingressgateway.istio-system.svc.cluster.local"
	privateLBDomain         = "cluster-local-gateway.istio-system.svc.cluster.local"
	producerServiceName     = "async-producer"
	asyncOriginalHostHeader = "Async-Original-Host"
)

// ReconcileKind implements Interface.ReconcileKind.
func (r *Reconciler) ReconcileKind(ctx context.Context, ing *v1alpha1.Ingress) reconciler.Event {
	logger := logging.FromContext(ctx)
	// TODO(beemarie): allow this ingress class to be configurable
	ingressClass := networkpkg.IstioIngressClassName
	err := validateAsyncModeAnnotation(ing.Annotations)
	if err != nil {
		logger.Errorf("error validating ingress annotations: %w", err)
		return err
	}

	markIngressReady(ing) //TODO(beemarie): this just sets the status of KIngress, but load balancer isn't needed.
	desired := makeNewIngress(ing, ingressClass)
	service := MakeK8sService(ing)
	_, err = r.reconcileIngress(ctx, desired)
	if err != nil {
		logger.Errorf("error reconciling ingress: %s", desired.Name)
		return err
	}
	err = r.reconcileService(ctx, service)
	if err != nil {
		logger.Errorf("error reconciling service: %s", service.Name)
		return err
	}
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

// makeNewIngress creates an Ingress object with respond-async headers pointing to async-producer
func makeNewIngress(ingress *v1alpha1.Ingress, ingressClass string) *v1alpha1.Ingress {
	original := ingress.DeepCopy()
	splits := make([]v1alpha1.IngressBackendSplit, 0, 1)
	splits = append(splits, v1alpha1.IngressBackendSplit{
		IngressBackend: v1alpha1.IngressBackend{
			ServiceName:      kmeta.ChildName(ingress.Name, asyncSuffix),
			ServiceNamespace: original.Namespace,
			ServicePort:      intstr.FromInt(80),
		},
		Percent: int(100),
	})
	theRules := []v1alpha1.IngressRule{}
	for _, rule := range original.Spec.Rules {
		newRule := rule
		newPaths := make([]v1alpha1.HTTPIngressPath, 0)
		if ingress.Annotations[AsyncModeAnnotationKey] == asyncAlwaysMode {
			for _, path := range rule.HTTP.Paths {
				defaultPath := path
				defaultPath.Splits = splits
				defaultPath.AppendHeaders = map[string]string{
					asyncOriginalHostHeader: getClusterLocalDomain(ingress.Name, ingress.Namespace),
				}
				defaultPath.RewriteHost = getClusterLocalDomain(producerServiceName, system.Namespace())
				if path.Headers == nil {
					path.Headers = map[string]v1alpha1.HeaderMatch{preferHeaderField: {Exact: preferSyncValue}}
				} else {
					path.Headers[preferHeaderField] = v1alpha1.HeaderMatch{Exact: preferSyncValue}
				}
				newPaths = append(newPaths, path, defaultPath)
				newRule.HTTP.Paths = newPaths
				theRules = append(theRules, newRule)
			}
		} else {
			newPaths = append(newPaths, v1alpha1.HTTPIngressPath{
				Headers: map[string]v1alpha1.HeaderMatch{preferHeaderField: {Exact: preferAsyncValue}},
				Splits:  splits,
				AppendHeaders: map[string]string{
					asyncOriginalHostHeader: getClusterLocalDomain(ingress.Name, ingress.Namespace),
				},
				RewriteHost: getClusterLocalDomain(producerServiceName, system.Namespace()),
			})
			newPaths = append(newPaths, newRule.HTTP.Paths...)
			newRule.HTTP.Paths = newPaths
			theRules = append(theRules, newRule)
		}
	}
	return &v1alpha1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      original.Name + newSuffix,
			Namespace: original.Namespace,
			Annotations: kmeta.FilterMap(kmeta.UnionMaps(map[string]string{
				networking.IngressClassAnnotationKey: ingressClass,
			}), func(key string) bool {
				return key == corev1.LastAppliedConfigAnnotation
			}),
			Labels:          original.Labels,
			OwnerReferences: original.OwnerReferences,
		},
		Spec: v1alpha1.IngressSpec{
			Rules: theRules,
		},
	}
}

// TODO(beemarie) track status of upstream ingress that is created "-new"
func markIngressReady(ingress *v1alpha1.Ingress) {
	privateDomain := domainForLocalGateway(ingress.Name, true)
	publicDomain := domainForLocalGateway(ingress.Name, false)

	ingress.Status.MarkLoadBalancerReady(
		[]v1alpha1.LoadBalancerIngressStatus{{
			DomainInternal: publicDomain,
		}},
		[]v1alpha1.LoadBalancerIngressStatus{{
			DomainInternal: privateDomain,
		}},
	)
	ingress.Status.MarkNetworkConfigured()
}

// TODO(beemarie) we need to pull this from the upstream ingress that is create "-new"
func domainForLocalGateway(ingressName string, isPrivate bool) string {
	if isPrivate {
		return privateLBDomain
	}
	return publicLBDomain
}

func (r *Reconciler) reconcileService(ctx context.Context, desiredSvc *corev1.Service) error {
	logger := logging.FromContext(ctx)

	sn := desiredSvc.Name
	service, err := r.serviceLister.Services(desiredSvc.Namespace).Get(sn)
	if apierrs.IsNotFound(err) {
		logger.Infof("K8s public service %s does not exist; creating.", sn)
		_, err := r.kubeclient.CoreV1().Services(desiredSvc.Namespace).Create(ctx, desiredSvc, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("Failed to create async K8s Service: %w", err)
		}
		logger.Info("Created K8s service: ", sn)
		return nil
	} else if err != nil {
		return fmt.Errorf("Failed to get async K8s Service: %w", err)
	} else {
		if !equality.Semantic.DeepEqual(service.Spec, desiredSvc.Spec) {
			// Don't modify the informers copy
			template := service.DeepCopy()
			template.Spec = desiredSvc.Spec
			if _, err = r.kubeclient.CoreV1().Services(service.Namespace).Update(ctx, template, metav1.UpdateOptions{}); err != nil {
				return fmt.Errorf("Failed to update public K8s Service: %w", err)
			}
		}
	}
	logger.Debug("Finished reconciling public K8s service: ", sn)
	return nil
}

// MakeK8sService constructs a K8s service, that is used to route service to the producer service
func MakeK8sService(ingress *v1alpha1.Ingress) *corev1.Service {
	selector := make(map[string]string)
	selector["app"] = producerServiceName
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            kmeta.ChildName(ingress.ObjectMeta.Name, asyncSuffix),
			Namespace:       ingress.Namespace,
			OwnerReferences: ingress.OwnerReferences,
		},
		Spec: corev1.ServiceSpec{
			Type:         "ExternalName",
			ExternalName: getClusterLocalDomain(producerServiceName, system.Namespace()),
			Ports: []corev1.ServicePort{{
				Name:       networking.ServicePortName(networking.ProtocolHTTP1),
				Protocol:   corev1.ProtocolTCP,
				Port:       int32(networking.ServicePort(networking.ProtocolHTTP1)),
				TargetPort: intstr.FromInt(80),
			}},
			Selector:        selector,
			SessionAffinity: "None",
		},
	}
}

func getClusterLocalDomain(serviceName string, namespace string) string {
	return serviceName + "." + namespace + ".svc." + network.GetClusterDomainName()
}

func validateAsyncModeAnnotation(annotations map[string]string) error {
	asyncMode := annotations[AsyncModeAnnotationKey]
	if asyncMode != "" && asyncMode != asyncAlwaysMode && asyncMode != asyncConditionalMode {
		return fmt.Errorf("Invalid value for key %s: ", AsyncModeAnnotationKey)
	}
	return nil
}
