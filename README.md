# Knative Asynchronous Component

>Warning: Experimental and still under development. Not meant for production deployment.
>Note: This component is currently only functional with Istio as the networking layer.

This is an add-on component that, when installed, will enable your Knative services to be called asynchronously. You can set a service to be always or conditionally asynchronous. Conditionally asynchronous services will respond when the `Prefer: respond-async` header is provided as a part of the request, while always asynchronous services do not need a special header for asynchronous functionality.

## Architecture

![diagram](./README-images/async-all-components.png)

When Knative Serving creates a service, one of the artifacts created is a KIngress for the service. The yaml for the KIngress will contain an annotation which is read by the networking controller (net-istio, net-countour, etc.). The ingress controller in our asynchronous component looks for KIngresses with the `async.ingress.networking.knative.dev` annotation. The ingress controller will then create a KIngress for Istio (annotated with `istio.ingress.networking.knative.dev`), which will create the required istio components, such as a virtual service, to route asynchronous service calls appropriately. If the service was an always asynchronous service, then all requests are routed to the producer component. If it was a conditional asynchronous service, then only requests with the `Prefer: respond-async` header will be routed. The producer component writes the request information to the redis stream and returns a `202 Accepted` response to the user. The consumer component reads from that stream and synchronously makes the service call.

## Install Knative Serving & Eventing to your Cluster

1. https://knative.dev/docs/install/any-kubernetes-cluster/

## Install the consumer and async controller components

1. Apply the following config files:
    ```
    ko apply -f config/async/100-async-consumer.yaml
    ko apply -f config/ingress/controller.yaml
    ```

## Install the Redis source 

### Using a cloud based Redis instance
1. Follow the `Getting Started - Install` Instructions for the [Redis Source](https://github.com/knative-sandbox/eventing-redis/tree/main/source#install).

1. Update the [producer .yaml file](config/async/100-async-producer.yaml) with the value for the `REDIS_ADDRESS`.

1. Update the [config-tls.yaml file](config/async/config-tls.yaml) with the cert.pem data key from your cloud instance. This will be the same key used in `Getting Started - Install` instructions.

1. There is a [.yaml file](config/async/100-async-redis-source.yaml) in the `async-component` describing the `RedisStreamSource`. It points to the `async-consumer` as the sink. First, update the `address` value in this .yaml file. You can then apply it to your cluster.
    ```
    kubectl apply -f config/async/100-async-redis-source.yaml
    ```

### For a local installation of Redis
1. Follow the `Getting Started` Instructions for the
   [Redis Source](https://github.com/knative-sandbox/eventing-redis/tree/master/source). For the `Example` section, do not install the entire `samples` folder, as you
   don't need the event-display sink. Only install redis with:
   `kubectl apply -f samples/redis`.

1. Update the [producer .yaml file](config/async/100-async-producer.yaml) with the value for the `REDIS_ADDRESS`. This should be `redis.redis.svc.cluster.local:6379`.

1. There is a [.yaml file](config/async/100-async-redis-source.yaml) in the `async-component` describing the `RedisStreamSource`. It points to the `async-consumer` as the sink. First update the address to `rediss://redis.redis.svc.cluster.local:6379`. You can then apply it to your cluster.
    ```
    kubectl apply -f config/async/100-async-redis-source.yaml
    ```

## Install the producer component.

1. Apply the producer config file to install the component:
    ```
    ko apply -f config/async/100-async-producer.yaml
    ```

## Create your demo application

1. This can be any simple hello world application. There is a sample application that sleeps for 10 seconds in the [`test/app`](test/app) folder. To deploy, use the `kubectl apply` command:
    ```
    kubectl apply -f test/app/service.yml
    ```

1. Note that your application has an annotation setting the `ingress.class` as `async.ingress.networking.knative.dev`. This enables just this application to respond to the `Prefer: respond-async` header.
    ```
    networking.knative.dev/ingress.class: async.ingress.networking.knative.dev
    ```

1. Make note of your application URL, which you can get with the following command:
    ```
    kubectl get kservice helloworld-sleep
    ```
    
1. (Optional) If you wanted every service created by knative to respond to the `Prefer: respond-async` header, you can configure Knative Serving to use the async ingress class for every service.

    ```
    kubectl patch configmap/config-network \
    -n knative-serving \
    --type merge \
    -p '{"data":{"ingress.class":"async.ingress.networking.knative.dev"}}'
    ```

    You can remove this setting by updating the ingress.class to null or by updating the ingress.class to the ingress.class you would like to use, for example `kourier`.
    ```
    kubectl patch configmap/config-network \
    -n knative-serving --type merge \
    -p '{"data":{"ingress.class":null}}'
    ```

    ```
    kubectl patch configmap/config-network \
    -n knative-serving \
    --type merge \
    -p '{"data":{"ingress.class":"kourier.ingress.networking.knative.dev"}}'
    ```
    
## Test your application
1. Curl your application. Try both asynchronous and non asynchronous requests.
    ```
    curl helloworld-sleep.default.11.112.113.14.xip.io
    curl helloworld-sleep.default.11.112.113.14.xip.io -H "Prefer: respond-async" -v
    ```

1. For the synchronous case, you should see that the connection remains open to the client, and does not close until about 10 seconds have passed, which is the amount of time this application sleeps. For the asynchronous case, you should see a `202` response returned immediately. 

## Update your Knative service to be always asynchronous.
1. To set a service to always respond asynchronously, rather than conditionally requiring the header, you can add the following annotation in the `.yml` for the service.
    ```
    async.knative.dev/mode: always.async.knative.dev
    ```

1. You can find an example of this (commented) in the [`test/app/service.yml`](test/app/service.yml) file. Uncomment the annotation `async.knative.dev/mode: always.async.knative.dev`.

1. Update the application by applying the `.yaml` file:
    ```
    kubectl apply -f test/app/service.yml
    ```

## Test your application
1. Curl the application, this time without the `Prefer: respond-async` header. You should see a `202` response returned while some pods are spun up to handle your request.
    ```
    curl helloworld-sleep.default.11.112.113.14.xip.io -v
    ```

1. You can see the pods with `kubectl get pods.`

