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

package e2e

import (
	"knative.dev/client/lib/test"
)

type Ko struct {
}

type Kperf struct {
	kn: test.Kn
}

type E2ETest struct {
	knTest   *test.KnTest
	kperf *Kperf
	ko *Ko
}

// NewE2ETest with params
func NewE2ETest() (*E2ETest, error) {
	knTest, err := test.NewKnTest()
	if err != nil {
		return nil, err
	}

	kperf := &Kperf{
		kn:         knTest.Kn(),
	}

	ko := &Ko{		
	}

	e2eTest := &E2ETest{
		knTest:   knTest,
		kperf: kperf,
		ko: ko,
	}

	return e2eTest, nil
}

// KnTest object
func (e2eTest *E2ETest) KnTest() *test.KnTest {
	return e2eTest.knTest
}

// Kperf object
func (e2eTest *E2ETest) Kperf() *Kperf {
	return e2eTest.kperf
}

// Ko object
func (e2eTest *E2ETest) Ko() *Ko {
	return e2eTest.ko
}
