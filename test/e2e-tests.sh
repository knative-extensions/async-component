#!/usr/bin/env bash

# Copyright 2021 The Knative Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# This script runs the end-to-end tests for the async component.

source $(dirname $0)/../vendor/knative.dev/hack/e2e-tests.sh
# Add local dir to have access to built in

TEST_KOURIER=${TEST_KOURIER:-0}
TEST_ISTIO=${TEST_ISTIO:-0}
TEST_CONTOUR=${TEST_CONTOUR:-0}

run(){
  # Create cluster
  initialize $@

  set -x

  install_prerequisites
  manage_dependencies

  # Smoke test
  eval smoke_test || fail_test

  smoke_test_clean_up
  delete_prerequisites

  success
}

function parse_flags() {
  # This function will be called repeatedly by initialize() with one fewer
  # argument each time and expects a return value of "the number of arguments to skip"
  # so we can just check the first argument and return 1 (to have it redirected to the
  # test container) or 0 (to have initialize() parse it normally).
  case $1 in
    --kourier)
      TEST_KOURIER=1
      return 1
      ;;
    --istio)
      TEST_ISTIO=1
      return 1
      ;;
    --contour)
      TEST_CONTOUR=1
      return 1
      ;;
  esac
  return 0
}

manage_dependencies(){
  git clone https://github.com/knative-sandbox/eventing-redis.git --branch main
}

install_prerequisites(){
  # Set up knative serving
  kubectl apply -f https://github.com/knative/serving/releases/latest/download/serving-crds.yaml || fail_test
  kubectl apply -f https://github.com/knative/serving/releases/latest/download/serving-core.yaml || fail_test

  # Set up networking layer
  set_up_networking || fail_test

  # Configure DNS
  kubectl apply -f https://github.com/knative/serving/releases/latest/download/serving-default-domain.yaml || fail_test

  # Set up knative eventing
  kubectl apply -f https://github.com/knative/eventing/releases/latest/download/eventing-crds.yaml || fail_test
  kubectl apply -f https://github.com/knative/eventing/releases/latest/download/eventing-core.yaml || fail_test
}

delete_prerequisites(){
  kubectl delete -f https://storage.googleapis.com/knative-nightly/eventing/latest/eventing-crds.yaml
  kubectl delete -f https://storage.googleapis.com/knative-nightly/eventing/latest/eventing-core.yaml
  kubectl delete -f https://github.com/knative/serving/releases/download/knative-v1.0.0/serving-default-domain.yaml
  delete_networking
  kubectl delete -f https://storage.googleapis.com/knative-nightly/serving/latest/serving-crds.yaml
  kubectl delete -f https://storage.googleapis.com/knative-nightly/serving/latest/serving-core.yaml
}

set_up_networking(){
  if [[ $TEST_KOURIER == 1 ]]; then
    echo "Setting up Kourier as networking layer"
    set_up_networking_kourier
  elif [[ $TEST_ISTIO == 1 ]]; then
    echo "Setting up Istio as networking layer"
    set_up_networking_istio
  elif [[ $TEST_CONTOUR == 1 ]]; then
    echo "Setting up Contour as networking layer"
    set_up_networking_contour
  else
    echo "No networking flag found - setting up default networking layer Kourier"
    set_up_networking_kourier
  fi
}

delete_networking(){
  if [[ $TEST_KOURIER == 1 ]]; then
    echo "Deleting networking layer Kourier"
    delete_networking_kourier
  elif [[ $TEST_ISTIO == 1 ]]; then
    echo "Deleting networking layer Istio"
    delete_networking_istio
  elif [[ $TEST_CONTOUR == 1 ]]; then
    echo "Deleting networking layer Contour"
    delete_networking_contour
  else
    echo "No networking flag found - deleting networking layer Kourier"
    delete_networking_kourier
  fi
}

set_up_networking_kourier(){
  kubectl apply -f https://github.com/knative/net-kourier/releases/latest/download/kourier.yaml || fail_test
  sleep 10
  kubectl patch configmap/config-network --namespace knative-serving --type merge --patch '{"data":{"ingress.class":"kourier.ingress.networking.knative.dev"}}' || fail_test
}

delete_networking_kourier(){
  kubectl delete -f https://github.com/knative/net-kourier/releases/latest/download/kourier.yaml
}

set_up_networking_contour(){
  kubectl apply -f https://github.com/knative/net-contour/releases/latest/download/contour.yaml || fail_test
  kubectl apply -f https://github.com/knative/net-contour/releases/latest/download/net-contour.yaml || fail_test
  sleep 10
  kubectl patch configmap/config-network \
    --namespace knative-serving \
    --type merge \
    --patch '{"data":{"ingress-class":"contour.ingress.networking.knative.dev"}}' || fail_test
}

delete_networking_contour(){
  kubectl delete -f https://github.com/knative/net-contour/releases/latest/download/net-contour.yaml
  kubectl delete -f https://github.com/knative/net-contour/releases/latest/download/contour.yaml
}

set_up_networking_istio(){
  kubectl apply -l knative.dev/crd-install=true -f https://github.com/knative/net-istio/releases/latest/download/istio.yaml
  kubectl apply -f https://github.com/knative/net-istio/releases/latest/download/istio.yaml
  kubectl apply -f https://github.com/knative/net-istio/releases/latest/download/net-istio.yaml
}

delete_networking_istio(){
  kubectl delete -f https://github.com/knative/net-istio/releases/latest/download/net-istio.yaml
  kubectl delete -f https://github.com/knative/net-istio/releases/latest/download/istio.yaml
}

smoke_test_clean_up(){
  # Remove the demo application
  kubectl delete -f test/app/service.yml

  # Remove the producer component
  ko delete -f config/async/100-async-producer.yaml

  # Remove the RedisStreamSource and tls secret
  kubectl delete -f config/async/100-async-redis-source.yaml
  kubectl delete -f config/async/tls-secret.yaml

  # Switch to the redis dir, delete component, switch back to async dir
  kubectl delete -f ./eventing-redis/samples/redis
  ko delete -f ./eventing-redis/source/config

  # Remove the consumer and async controller components
  ko delete -f config/ingress/controller.yaml
  ko delete -f config/async/100-async-consumer.yaml
}

smoke_test() {
  header "Running smoke tests"
  # Assume at this point we have ko/kubectl/curl

  # Async uses default namespace
  set -x

  # Install the consumer and async controller components
  ko apply -f config/async/100-async-consumer.yaml || fail_test
  install_ingress_controller || fail_test

  # Install the Redis Source
  cd ./eventing-redis
  ko apply -f ./source/config || fail_test
  kubectl apply -f ./samples/redis || fail_test
  cd ..

  kubectl apply -f config/async/tls-secret.yaml || fail_test
  kubectl apply -f config/async/100-async-redis-source.yaml || fail_test

  # Install the producer component
  ko apply -f config/async/100-async-producer.yaml || fail_test

  # Create the demo application
  kubectl apply -f test/app/service.yml || fail_test

  # Wait for helloworld-sleep route to be set up
  sleep 20

  set_gateway_ip
  export DOMAIN_NAME=`kubectl get route helloworld-sleep --output jsonpath="{.status.url}" | cut -d'/' -f 3`

  # Add the record of Gateway IP and domain name into file "/etc/hosts"
  echo -e "$GATEWAY_IP\t$DOMAIN_NAME" | sudo tee -a /etc/hosts

  # Get the url for application
  helloworld_url=$(kubectl get kservice helloworld-sleep --output jsonpath="{.status.url}" | cut -d'/' -f 3)

  # Verify synchronous response
  regular_response=$(curl -s -o /dev/null -w "%{http_code}" $helloworld_url)

  if [[ $regular_response != 200 ]]
  then
    fail_test
  fi

  # Verify asynchronous response
  async_response=$(curl -s -o /dev/null -w "%{http_code}" $helloworld_url -H "Prefer: respond-async")

  if [[ $async_response != 202 ]]
  then
    fail_test
  fi

  # Additional testing
  # Tests should restore environment to default test service / ingress classes before ending
  always_async
  no_ingress_annotation
  serving_async_ingress
}

always_async(){
  # Recreate test service in always async mode, query without async header, expect 202
  kubectl delete -f test/app/service.yml
  kubectl apply -f test/app/alwaysAsyncService.yml
  sleep 20
  # Get the url for application
  helloworld_url=$(kubectl get kservice helloworld-sleep --output jsonpath="{.status.url}" | cut -d'/' -f 3)
  always_async_response=$(curl -s -o /dev/null -w "%{http_code}" $helloworld_url)
  if [[ $always_async_response != 202 ]]
  then
    fail_test
  fi

  kubectl apply -f test/app/service.yml
}

no_ingress_annotation(){
  # Deleting original demo service, recreating one with no defined ingress annotation"
  kubectl delete -f test/app/service.yml
  kubectl apply -f ./test/app/noAnnoService.yml
  sleep 20
  # Calling demo service with async header, but expect it to be handled synchronously by using serving default ingress
  helloworld_url=$(kubectl get kservice helloworld-sleep --output jsonpath="{.status.url}" | cut -d'/' -f 3)
  no_ingress_annotation_response=$(curl -s -o /dev/null -w "%{http_code}" $helloworld_url -H "Prefer: respond-async")

  if [[ $no_ingress_annotation_response != 200 ]]
  then
    fail_test
  fi

  kubectl delete -f ./test/app/noAnnoService.yml
  kubectl apply -f test/app/service.yml
}

serving_async_ingress(){
  # Deleting demo service, patching service network configmap to use async as the default ingress
  kubectl delete -f test/app/service.yml
  kubectl patch configmap/config-network \
  -n knative-serving \
  --type merge \
  -p '{"data":{"ingress.class":"async.ingress.networking.knative.dev"}}'

  # Recreating the demo service with no defined ingress annotation, wait for ready, call with async header, expect 202
  kubectl apply -f ./test/app/noAnnoService.yml
  sleep 20
  helloworld_url=$(kubectl get kservice helloworld-sleep --output jsonpath="{.status.url}" | cut -d'/' -f 3)
  serving_aasync_ingress_response=$(curl -s -o /dev/null -w "%{http_code}" $helloworld_url -H "Prefer: respond-async")

  if [[ $serving_aasync_ingress_response != 202 ]]
  then
    fail_test
  fi

  # Cleaning up, restoring to defaults
  kubectl delete -f ./test/app/noAnnoService.yml
  if [[ $TEST_KOURIER == 1 ]]; then
    restore_ingress="kourier.ingress.networking.knative.dev"
  elif [[ $TEST_ISTIO == 1 ]]; then
    restore_ingress="istio.ingress.networking.knative.dev"
  elif [[ $TEST_CONTOUR == 1 ]]; then
    restore_ingress="contour.ingress.networking.knative.dev"
  elif [[ $TEST_AMBASSADOR == 1 ]]; then
    restore_ingress="ambassador.ingress.networking.knative.dev"
  else
    echo "No networking flag found - restore ingress class Kourier"
    restore_ingress="kourier.ingress.networking.knative.dev"
  fi

  kubectl patch configmap/config-network \
  -n knative-serving \
  --type merge \
  -p '{"data":{"ingress.class":"'$restore_ingress'"}}'
  kubectl apply -f test/app/service.yml
}

install_ingress_controller(){
  if [[ $TEST_KOURIER == 1 ]]; then
      echo "Setting up ingress controller for Kourier"
      ko apply -f config/ingress/kourier.yaml
    elif [[ $TEST_ISTIO == 1 ]]; then
      echo "Setting up ingress controller for Istio"
      ko apply -f config/ingress/istio.yaml
    elif [[ $TEST_CONTOUR == 1 ]]; then
      echo "Setting up ingress controller for Contour"
      ko apply -f config/ingress/contour.yaml
    else
      echo "Setting up ingress controller for custom local (default - Kourier)"
      ko apply -f config/ingress/controller.yaml
    fi
}

set_gateway_ip(){
  if [[ $TEST_KOURIER == 1 ]]; then
    echo "Setting gateway ip for Kourier"
    set_gateway_ip_kourier
  elif [[ $TEST_ISTIO == 1 ]]; then
    echo "Setting gateway ip for Istio"
    set_gateway_ip_istio
  elif [[ $TEST_CONTOUR == 1 ]]; then
    echo "Setting gateway ip for Contour"
    set_gateway_ip_contour
  else
    echo "Setting gateway ip for default Kourier"
    set_gateway_ip_kourier
  fi
}

set_gateway_ip_kourier(){
  INGRESSGATEWAY=kourier
  export GATEWAY_IP=`kubectl get svc $INGRESSGATEWAY --namespace kourier-system --output jsonpath="{.status.loadBalancer.ingress[*]['ip']}"`
}

set_gateway_ip_istio(){
  INGRESSGATEWAY=istio-ingressgateway
  export GATEWAY_IP=`kubectl get svc $INGRESSGATEWAY --namespace istio-system --output jsonpath="{.status.loadBalancer.ingress[*]['ip']}"`
}

set_gateway_ip_contour(){
  INGRESSGATEWAY=envoy
  export GATEWAY_IP=`kubectl get svc $INGRESSGATEWAY --namespace contour-external --output jsonpath="{.status.loadBalancer.ingress[*]['ip']}"`
}

run $@