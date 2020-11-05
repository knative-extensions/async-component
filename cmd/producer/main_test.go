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
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-redis/redis/v8"
)

type fakeRedis struct {
	client redis.Cmdable
}

func TestHandleRequest(t *testing.T) {
	testserver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" && r.Method != "POST" {
			t.Errorf("Expected 'POST' OR 'GET' request, got '%s'", r.Method)
		}
	}))
	setupRedis()

	tests := []struct {
		name             string
		async            bool
		method           string
		body             string
		contentLengthSet bool
		returncode       int
	}{{
		name:       "async get request",
		method:     "GET",
		body:       "",
		returncode: 202,
	}, {
		name:       "async post request with too large payload",
		method:     "POST",
		body:       `{"body":"this is a larger body"}`,
		returncode: 500,
	}, {
		name:       "async post request with smaller than limit payload",
		method:     "POST",
		body:       `{"body":"this is a body"}`,
		returncode: 202,
	}, {
		name:       "test failure to write to Redis",
		method:     "POST",
		body:       "failure",
		returncode: 500,
	},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			env = envInfo{
				StreamName:       "mystream",
				RedisAddress:     "address",
				RequestSizeLimit: 25,
			}
			request, _ := http.NewRequest(http.MethodGet, testserver.URL, nil)
			if test.method == "POST" {
				var body *strings.Reader
				if test.body != "" {
					body = strings.NewReader(test.body)
				}
				request, _ = http.NewRequest(http.MethodPost, testserver.URL, body)
			}

			rr := httptest.NewRecorder()
			handleRequest(rr, request)

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
	if strings.Contains(string(reqJSON), "failure") {
		return errors.New("Failure writing")
	}
	return // no need to actually write to redis stream for our test case.
}
