module knative.dev/async-component

go 1.14

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bradleypeabody/gouuidv6 v0.0.0-20200224230637-90681a9a9294
	github.com/cloudevents/sdk-go/v2 v2.2.0
	github.com/go-redis/redis/v8 v8.0.0-beta.7
	github.com/json-iterator/go v1.1.10 // indirect
	github.com/kelseyhightower/envconfig v1.4.0
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	k8s.io/kube-openapi v0.0.0-20200831175022-64514a1d5d59 // indirect
	knative.dev/hack v0.0.0-20210325223819-b6ab329907d3
	knative.dev/net-contour v0.22.0
	knative.dev/networking v0.0.0-20210331064822-999a7708876c
	knative.dev/pkg v0.0.0-20210331065221-952fdd90dbb0
	sigs.k8s.io/structured-merge-diff/v3 v3.0.1-0.20200706213357-43c19bbb7fba // indirect
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
