// Copyright 2020 The Knative Authors

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build e2e
// +build !eventing

package e2e

import (
	"os"
	"path/filepath"
	"testing"

	"gotest.tools/v3/assert"
	"knative.dev/client/lib/test"
	"knative.dev/client/pkg/util"
)

func TestBasicWorkflow(t *testing.T) {
	t.Parallel()

	currentDir, err := os.Getwd()
	assert.NilError(t, err)

	it, err := NewE2ETest()
	assert.NilError(t, err)
	defer func() {
		assert.NilError(t, it.KnTest().Teardown())
	}()

	r := test.NewKnRunResultCollector(t, it.KnTest())
	defer r.DumpIfFailed()

	t.Log("install the consumer and async, redis, and producer components")
	it.install(t, r)

	t.Log("create demo app")
	it.createDemoApp(t, r)

	t.Log("test demo app")
	it.testDemoApp(t, r)

	t.Log("clean up")
	it.cleanup(t, r)
}

// Private
func (it *E2ETest) installConsumer(t *testing.T, r *test.KnRunResultCollector) {
	out := it.Ko().Run("apply -f config/async/100-async-consumer.yaml")
	r.AssertNoError(out)
	assert.Check(t, util.ContainsAllIgnoreCase(out.Stdout, "")) //TODO check output

	out = it.Ko().Run("apply -f config/ingress/controller.yaml")
	r.AssertNoError(out)
	assert.Check(t, util.ContainsAllIgnoreCase(out.Stdout, "")) //TODO check output

	//TODO automate redis install

	out = it.Ko().Run("apply -f config/async/100-async-producer.yaml")
	r.AssertNoError(out)
	assert.Check(t, util.ContainsAllIgnoreCase(out.Stdout, "")) //TODO check output
}

func (it *E2ETest) createDemoApp(t, r) {
	out := it.Kubectl().Run("apply -f test/app/service.yml")
	r.AssertNoError(out)
	assert.Check(t, util.ContainsAllIgnoreCase(out.Stdout, "")) //TODO check output
}

func (it *E2ETest) testDemoAppUrl(t, r) string {
	//TODO automate
	// 1. get URL for helloworld-sleep
	// 2. invoke (curl) service non-async and verify responses 200
	// 3. invoke (curl) service async and verify response 202
}

func (it *E2ETest) cleanup(t, r) {
	out := it.Kubectl().Run("delete -f test/app/service.yml")
	r.AssertNoError(out)
	assert.Check(t, util.ContainsAllIgnoreCase(out.Stdout, "")) //TODO check output

	out = it.Ko().Run("ko delete -f config/async/100-async-producer.yaml")
	r.AssertNoError(out)
	assert.Check(t, util.ContainsAllIgnoreCase(out.Stdout, "")) //TODO check output
	
	out = it.Kubectl().Run("delete -f config/async/100-async-redis-source.yaml")
	r.AssertNoError(out)
	assert.Check(t, util.ContainsAllIgnoreCase(out.Stdout, "")) //TODO check output
	
	out = it.Kubectl().Run("delete -f config/async/tls-secret.yaml")
	r.AssertNoError(out)
	assert.Check(t, util.ContainsAllIgnoreCase(out.Stdout, "")) //TODO check output

	out = it.Ko().Run("delete -f source/config")
	r.AssertNoError(out)
	assert.Check(t, util.ContainsAllIgnoreCase(out.Stdout, "")) //TODO check output
    
	out = it.Ko().Run("delete -f config/ingress/controller.yaml")
	r.AssertNoError(out)
	assert.Check(t, util.ContainsAllIgnoreCase(out.Stdout, "")) //TODO check output
    
	out = it.Ko().Run("delete -f config/async/100-async-consumer.yaml")
	r.AssertNoError(out)
	assert.Check(t, util.ContainsAllIgnoreCase(out.Stdout, "")) //TODO check output
}