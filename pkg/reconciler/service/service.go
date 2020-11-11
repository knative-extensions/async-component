package service

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
	"knative.dev/networking/pkg/apis/networking"
	"knative.dev/networking/pkg/apis/networking/v1alpha1"
	netclientset "knative.dev/networking/pkg/client/clientset/versioned"
	networkinglisters "knative.dev/networking/pkg/client/listers/networking/v1alpha1"

	"knative.dev/pkg/kmeta"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/reconciler"
)

// Reconciler implements controller.Reconciler for Ingress resources.
type Reconciler struct {
	ingressLister networkinglisters.IngressLister
	serviceLister corev1listers.ServiceLister
	netclient     netclientset.Interface
	kubeclient    kubernetes.Interface
}

const (
	asyncSuffix         = "-async"
	producerServiceName = "producer-service"
)

// ReconcileKind implements Interface.ReconcileKind.
func (r *Reconciler) ReconcileKind(ctx context.Context, ing *v1alpha1.Ingress) reconciler.Event {
	logger := logging.FromContext(ctx)
	service := MakeK8sService(ing)
	err := r.reconcileService(ctx, service)
	if err != nil {
		logger.Errorf("error reconciling service: %s", service.Name)
		return err
	}
	return nil
}

func (r *Reconciler) reconcileService(ctx context.Context, desiredSvc *corev1.Service) error {
	logger := logging.FromContext(ctx)

	sn := desiredSvc.Name
	service, err := r.serviceLister.Services(desiredSvc.Namespace).Get(sn)
	if apierrs.IsNotFound(err) {
		logger.Infof("K8s public service %s does not exist; creating.", sn)
		_, err := r.kubeclient.CoreV1().Services(desiredSvc.Namespace).Create(ctx, desiredSvc, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create async K8s Service: %w", err)
		}
		logger.Info("Created K8s service: ", sn)
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to get async K8s Service: %w", err)
	} else if !equality.Semantic.DeepEqual(service.Spec, desiredSvc.Spec) {
		// Don't modify the informers copy
		origin := service.DeepCopy()
		origin.Spec = desiredSvc.Spec
		if _, err = r.kubeclient.CoreV1().Services(service.Namespace).Update(ctx, desiredSvc, metav1.UpdateOptions{}); err != nil {
			return fmt.Errorf("failed to update public K8s Service: %w", err)
		}
		return nil
	}
	return err
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
}
