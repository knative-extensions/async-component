/*
Copyright 2020 The Knative Authors
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
package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-redis/redis/v8"
)

type fakeRedis struct {
	client redis.Cmdable
}

func TestAsyncRequestHeader(t *testing.T) {
	testserver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" && r.Method != "POST" {
			t.Errorf("Expected 'POST' OR 'GET' request, got '%s'", r.Method)
		}
	}))

	tests := []struct {
		name             string
		async            bool
		method           string
		largeBody        bool
		contentLengthSet bool
		returncode       int
	}{{
		name:       "async get request",
		async:      true,
		method:     "GET",
		largeBody:  false,
		returncode: 202,
	}, {
		name:       "non async get request",
		async:      false,
		method:     "GET",
		largeBody:  false,
		returncode: 200,
	}, {
		name:       "non async post request",
		async:      false,
		method:     "POST",
		largeBody:  false,
		returncode: 200,
	}, {
		name:       "async post request with too large payload",
		async:      true,
		method:     "POST",
		largeBody:  true,
		returncode: 500,
	}, {
		name:       "async post request with smaller than limit payload",
		async:      true,
		method:     "POST",
		largeBody:  false,
		returncode: 202,
	},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			env = envInfo{
				StreamName:       "mystream",
				RedisAddress:     "address",
				RequestSizeLimit: 25,
			}
			setupRedis()
			request, _ := http.NewRequest(http.MethodGet, testserver.URL, nil)
			if test.method == "POST" {
				body := strings.NewReader(`{"body":"this is a body"}`)
				if test.largeBody == true {
					body = strings.NewReader(`{"body":"this is a larger body"}`)
				}
				request, _ = http.NewRequest(http.MethodPost, testserver.URL, body)
			}
			if test.async {
				request.Header.Set("Prefer", "respond-async")
			}

			rr := httptest.NewRecorder()
			checkHeaderAndServe(rr, request)

			got := rr.Code
			want := test.returncode

			if got != want {
				t.Errorf("got %d, want %d", got, want)
			}
		})
	}
}

func setupRedis() {
	// set up redis client
	opts := &redis.UniversalOptions{
		Addrs: []string{env.RedisAddress},
	}
	theclient := redis.NewUniversalClient(opts)
	rc = &fakeRedis{
		client: theclient,
	}
}

func (fr *fakeRedis) write(ctx context.Context, s envInfo, reqJSON []byte, id string) (err error) {
	return // no need to actually write to redis stream for our test case.
}
