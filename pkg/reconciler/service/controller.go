/*
Copyright 2019 The Knative Authors
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

	"knative.dev/networking/pkg/apis/networking"

	"k8s.io/client-go/tools/cache"
	kubeclient "knative.dev/pkg/client/injection/kube/client"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
	knativeReconciler "knative.dev/pkg/reconciler"

	ingressinformer "knative.dev/networking/pkg/client/injection/informers/networking/v1alpha1/ingress"
	v1alpha1ingress "knative.dev/networking/pkg/client/injection/reconciler/networking/v1alpha1/ingress"
	serviceinformer "knative.dev/pkg/client/injection/kube/informers/core/v1/service"
)

const (
	asyncIngressClassName = "async.ingress.networking.knative.dev"
)

// NewController creates a Reconciler and returns the result of NewImpl.
func NewController(
	ctx context.Context,
	cmw configmap.Watcher,
) *controller.Impl {
	logger := logging.FromContext(ctx)

	ingressInformer := ingressinformer.Get(ctx)
	serviceInformer := serviceinformer.Get(ctx)

	r := &Reconciler{
		ingressLister: ingressInformer.Lister(),
		serviceLister: serviceInformer.Lister(),
		kubeclient:    kubeclient.Get(ctx),
	}
	impl := v1alpha1ingress.NewImpl(ctx, r, asyncIngressClassName)

	logger.Info("Setting up event handlers.")

	// Ingresses need to be filtered by ingress class, so async-component does not
	// react to nor modify ingresses created by other gateways.
	classFilter := knativeReconciler.AnnotationFilterFunc(
		networking.IngressClassAnnotationKey, asyncIngressClassName, false,
	)

	ingressInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: classFilter,
		Handler:    controller.HandleAll(impl.Enqueue),
	})

	return impl
}
