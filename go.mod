module knative.dev/async-component

go 1.14

require (
	github.com/bradleypeabody/gouuidv6 v0.0.0-20200224230637-90681a9a9294
	github.com/cloudevents/sdk-go/v2 v2.2.0
	github.com/go-redis/redis/v8 v8.0.0-beta.7
	github.com/kelseyhightower/envconfig v1.4.0
	k8s.io/api v0.20.7
	k8s.io/apimachinery v0.20.7
	k8s.io/client-go v0.20.7
	k8s.io/kube-openapi v0.0.0-20210527164424-3c818078ee3d // indirect
	knative.dev/hack v0.0.0-20210325223819-b6ab329907d3
	knative.dev/net-contour v0.22.0
	knative.dev/networking v0.0.0-20210331064822-999a7708876c
	knative.dev/pkg v0.0.0-20210331065221-952fdd90dbb0
)
