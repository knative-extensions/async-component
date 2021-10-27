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

# TODO explain how to use this locally

source $(dirname $0)/../vendor/knative.dev/hack/e2e-tests.sh
# Add local dir to have access to built in

run(){
  # Create cluster
  initialize $@

  set -x

  install_prerequisites

  manage_dependencies

  # Smoke test
  eval smoke_test || fail_test

  smoke_test_clean_up

  success
}

manage_dependencies(){
  git clone https://github.com/knative-sandbox/eventing-redis.git --branch release-0.26
}

install_prerequisites(){
  # Set up knative serving
  kubectl apply -f https://github.com/knative/serving/releases/download/v0.26.0/serving-crds.yaml || fail_test
  kubectl apply -f https://github.com/knative/serving/releases/download/v0.26.0/serving-core.yaml || fail_test

  # Set up Networking layer (kourier is knative default now) TODO make this swappable in the future
  kubectl apply -f https://github.com/knative/net-kourier/releases/download/v0.26.0/kourier.yaml || fail_test
  kubectl patch configmap/config-network --namespace knative-serving --type merge --patch '{"data":{"ingress.class":"kourier.ingress.networking.knative.dev"}}' || fail_test

  # Configure DNS
  kubectl apply -f https://github.com/knative/serving/releases/download/v0.26.0/serving-default-domain.yaml || fail_test

  # Set up knative eventing
  kubectl apply -f https://github.com/knative/eventing/releases/download/v0.26.0/eventing-crds.yaml || fail_test
  kubectl apply -f https://github.com/knative/eventing/releases/download/v0.26.0/eventing-core.yaml || fail_test
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
  ko delete -f ./eventing-redis/source/config

  # Remove the consumer and async controller components
  ko delete -f config/ingress/controller.yaml
  ko delete -f config/async/100-async-consumer.yaml
}

# Currently expects config to be set up ahead of time
# TODO verify we are setting up environment correctly with eventing-redis in the same location as async-component
smoke_test() {
  header "Running smoke tests"
  # Assume at this point we have ko/kubectl/curl

  # Async uses default namespace
  set -x

  # Install the consumer and async controller components
  ko apply -f config/async/100-async-consumer.yaml || fail_test
  ko apply -f config/ingress/controller.yaml || fail_test

  # Install the Redis Source
  cd ./eventing-redis
  ko apply -f ./source/config || fail_test
  cd ..

  kubectl apply -f config/async/tls-secret.yaml || fail_test
  kubectl apply -f config/async/100-async-redis-source.yaml || fail_test

  # Install the producer component
  ko apply -f config/async/100-async-producer.yaml || fail_test

  # Create the demo application
  kubectl apply -f test/app/service.yml || fail_test

  # Wait for helloworld-sleep route to be set up
  sleep 20

  ingress_fix

  # Get the url for application
  helloworld_url=$(kubectl get kservice helloworld-sleep --output jsonpath="{.status.url}" | cut -d'/' -f 3)

  regular_response=$(curl -s -o /dev/null -w "%{http_code}" $helloworld_url)

  if [[ $regular_response != 200 ]]
  then
    fail_test
  fi

  async_response=$(curl -s -o /dev/null -w "%{http_code}" $helloworld_url -H "Prefer: respond-async")

  if [[ $async_response != 202 ]]
  then
    fail_test
  fi
}

#TODO let this use alternate ingresses, also is there a way to get around this? do we even have sudo?
ingress_fix(){

  INGRESSGATEWAY=istio-ingressgateway

  export GATEWAY_IP=`kubectl get svc $INGRESSGATEWAY --namespace istio-system --output jsonpath="{.status.loadBalancer.ingress[*]['ip']}"`

  export DOMAIN_NAME=`kubectl get route helloworld-sleep --output jsonpath="{.status.url}" | cut -d'/' -f 3`

  # Add the record of Gateway IP and domain name into file "/etc/hosts"
  echo -e "$GATEWAY_IP\t$DOMAIN_NAME" | sudo tee -a /etc/hosts

}

run $@