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
	k8s.io/kube-openapi v0.0.0-20210113233702-8566a335510f // indirect
	knative.dev/hack v0.0.0-20210622141627-e28525d8d260
	knative.dev/net-contour v0.22.0
	knative.dev/networking v0.0.0-20210628063847-2315e141d4f1
	knative.dev/pkg v0.0.0-20210628225612-51cfaabbcdf6
)
