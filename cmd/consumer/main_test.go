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
	"encoding/json"
	"flag"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

var (
	eventSource string
	eventType   string
	data        requestData
)

func TestConsumeEvent(t *testing.T) {
	// t.Run("consume cloud event", func(t *testing.T) {
	myEvent := cloudevents.NewEvent("1.0")
	flag.StringVar(&eventSource, "eventSource", "redis-source", "the event-source (CloudEvents)")
	flag.StringVar(&eventType, "eventType", "dev.knative.async.request", "the event-type (CloudEvents)")
	myEvent.SetType(eventType)
	myEvent.SetSource(eventSource)
	myEvent.SetID("123")

	testserver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" && r.Method != "POST" {
			t.Errorf("Expected 'POST' OR 'GET' request, got '%s'", r.Method)
		}
		if r.Method == "POST" {
			b, _ := ioutil.ReadAll(r.Body)
			bodyString := string(b)
			expectedBodyString := "{\"body\":\"test body\"}"
			if bodyString != expectedBodyString {
				t.Errorf("Expected body with POST request to match %s", expectedBodyString)
			}
		}
	}))

	tests := []struct {
		name        string
		method      string
		body        string
		reqURL      string
		expectedErr string
	}{{
		name:        "proper request data, get request",
		method:      http.MethodGet,
		reqURL:      testserver.URL,
		expectedErr: "",
	}, {
		name:        "proper request data, post request",
		method:      http.MethodPost,
		reqURL:      testserver.URL,
		expectedErr: "",
	}, {
		name:        "bad url format",
		method:      http.MethodGet,
		reqURL:      "http://badurl",
		expectedErr: "no such host",
	}, {
		name:        "no request URL, get request",
		method:      http.MethodGet,
		reqURL:      "",
		expectedErr: "unsupported protocol scheme",
	}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// create data for Request.
			data.ID = "123"
			data.ReqURL = test.reqURL
			data.ReqMethod = test.method
			if data.ReqMethod == "POST" {
				data.ReqBody = "{\"body\":\"test body\"}"
			}
			// marshal data to json and then translate to string to encode as base64
			out, err := json.Marshal(data)
			if err != nil {
				t.Errorf("Error marshaling json for test")
			}
			testData := []string{"data", string(out)}

			// setdata in the event
			myEvent.SetData(cloudevents.ApplicationJSON, testData)

			got := consumeEvent(myEvent)
			if test.expectedErr != "" {
				msg := got.Error()
				if !strings.Contains(msg, test.expectedErr) {
					t.Errorf("got %s, want %s", msg, test.expectedErr)
				}
			} else if got != nil {
				t.Errorf("got error when one was unexpected")
			}
		})
	}
}
