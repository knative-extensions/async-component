module knative.dev/async-component

go 1.14

require (
	github.com/bradleypeabody/gouuidv6 v0.0.0-20200224230637-90681a9a9294
	github.com/cloudevents/sdk-go/v2 v2.2.0
	github.com/go-redis/redis/v8 v8.0.0-beta.7
	github.com/google/go-cmp v0.5.2
	github.com/kelseyhightower/envconfig v1.4.0
	k8s.io/api v0.18.8
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	knative.dev/net-contour v0.18.1
	knative.dev/networking v0.0.0-20201016015257-30b677481f47
	knative.dev/pkg v0.0.0-20201016021557-c1a8664276b4
	knative.dev/test-infra v0.0.0-20201015231956-d236fb0ea9ff
)

replace (
	github.com/prometheus/client_golang => github.com/prometheus/client_golang v0.9.2

	k8s.io/api => k8s.io/api v0.18.8
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.18.8
	k8s.io/apimachinery => k8s.io/apimachinery v0.18.8
	k8s.io/apiserver => k8s.io/apiserver v0.18.8
	k8s.io/client-go => k8s.io/client-go v0.18.8
	k8s.io/code-generator => k8s.io/code-generator v0.18.8
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20200410145947-61e04a5be9a6
)
