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

	"github.com/go-redis/redis/v9"
)

type fakeRedis struct {
	client redis.Cmdable
}

func TestRedisClientSetup(t *testing.T) {
	tests := []struct {
		name string
		cert string
	}{{
		name: "test with fake cert from x509",
		cert: "-----BEGIN CERTIFICATE-----\n" +
			"MIIDBjCCAe6gAwIBAgIRANXM5I3gjuqDfTp/PYrs+u8wDQYJKoZIhvcNAQELBQAw\n" +
			"EjEQMA4GA1UEChMHQWNtZSBDbzAeFw0xODAzMjcxOTU2MjFaFw0xOTAzMjcxOTU2\n" +
			"MjFaMBIxEDAOBgNVBAoTB0FjbWUgQ28wggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAw\n" +
			"ggEKAoIBAQDK+9m3rjsO2Djes6bIYQZ3eV29JF09ZrjOrEHLtaKrD6/acsoSoTsf\n" +
			"cQr+rzzztdB5ijWXCS64zo/0OiqBeZUNZ67jVdToa9qW5UYe2H0Y+ZNdfA5GYMFD\n" +
			"yk/l3/uBu3suTZPfXiW2TjEi27Q8ruNUIZ54DpTcs6y2rBRFzadPWwn/VQMlvRXM\n" +
			"jrzl8Y08dgnYmaAHprxVzwMXcQ/Brol+v9GvjaH1DooHqkn8O178wsPQNhdtvN01\n" +
			"IXL46cYdcUwWrE/GX5u+9DaSi+0KWxAPQ+NVD5qUI0CKl4714yGGh7feXMjJdHgl\n" +
			"VG4QJZlJvC4FsURgCHJT6uHGIelnSwhbAgMBAAGjVzBVMA4GA1UdDwEB/wQEAwIF\n" +
			"oDATBgNVHSUEDDAKBggrBgEFBQcDATAMBgNVHRMBAf8EAjAAMCAGA1UdEQQZMBeC\n" +
			"FVRlc3RTeXN0ZW1DZXJ0UG9vbC5nbzANBgkqhkiG9w0BAQsFAAOCAQEAwuSRx/VR\n" +
			"BKh2ICxZjL6jBwk/7UlU1XKbhQD96RqkidDNGEc6eLZ90Z5XXTurEsXqdm5jQYPs\n" +
			"1cdcSW+fOSMl7MfW9e5tM66FaIPZl9rKZ1r7GkOfgn93xdLAWe8XHd19xRfDreub\n" +
			"YC8DVqgLASOEYFupVSl76ktPfxkU5KCvmUf3P2PrRybk1qLGFytGxfyice2gHSNI\n" +
			"gify3K/+H/7wCkyFW4xYvzl7WW4mXxoqPRPjQt1J423DhnnQ4G1P8V/vhUpXNXOq\n" +
			"N9IEPnWuihC09cyx/WMQIUlWnaQLHdfpPS04Iez3yy2PdfXJzwfPrja7rNE+skK6\n" +
			"pa/O1nF0AfWOpw==\n" +
			"-----END CERTIFICATE-----\n",
	}, {
		name: "test with empty cert",
		cert: "-----BEGIN CERTIFICATE-----" +
			"-----END CERTIFICATE-----",
	}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			env = envInfo{
				StreamName:   "mystream",
				RedisAddress: "rediss://redis.redis.svc.cluster.local:6379",
				TlsCert:      test.cert,
			}
			setUpRedis()
		})
	}
}

func TestHandleRequest(t *testing.T) {
	testserver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodPost {
			t.Errorf("Expected 'POST' OR 'GET' request, got '%s'", r.Method)
		}
	}))
	setupFakeRedis()

	tests := []struct {
		name             string
		async            bool
		method           string
		body             string
		contentLengthSet bool
		returncode       int
	}{{
		name:       "async get request",
		method:     http.MethodGet,
		body:       "",
		returncode: http.StatusAccepted,
	}, {
		name:       "async post request with too large payload",
		method:     http.MethodPost,
		body:       `{"body":"this is a larger body"}`,
		returncode: http.StatusInternalServerError,
	}, {
		name:       "async post request with smaller than limit payload",
		method:     http.MethodPost,
		body:       `{"body":"this is a body"}`,
		returncode: http.StatusAccepted,
	}, {
		name:       "test failure to write to Redis",
		method:     http.MethodPost,
		body:       "failure",
		returncode: http.StatusInternalServerError,
	}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			env = envInfo{
				StreamName:       "mystream",
				RedisAddress:     "address",
				RequestSizeLimit: 25,
			}
			request := httptest.NewRequest(http.MethodGet, testserver.URL, nil)
			if test.method == http.MethodPost {
				var body *strings.Reader
				if test.body != "" {
					body = strings.NewReader(test.body)
				}
				request = httptest.NewRequest(http.MethodPost, testserver.URL, body)
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

func setupFakeRedis() {
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
