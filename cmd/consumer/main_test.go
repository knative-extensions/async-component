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
	"bytes"
	"encoding/json"
	"flag"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

var (
	eventSource string
	eventType   string
	data        request
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
	}))

	getreq, _ := http.NewRequest(http.MethodGet, testserver.URL, nil)
	postreq, _ := http.NewRequest(http.MethodPost, testserver.URL, nil)
	badreq, _ := http.NewRequest(http.MethodGet, "http://badurl", nil)

	tests := []struct {
		name        string
		reqString   string
		expectedErr string
	}{{
		name:        "proper request data, get request",
		reqString:   getRequestString(getreq, t),
		expectedErr: "",
	}, {
		name:        "proper request data, post request",
		reqString:   getRequestString(postreq, t),
		expectedErr: "",
	}, {
		name:        "bad url format",
		reqString:   getRequestString(badreq, t),
		expectedErr: "no such host",
	}, {
		name:        "no request data, get request",
		reqString:   "",
		expectedErr: "EOF",
	}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// create data for Request. This is how xadd formats data when added to a
			// stream: ["data","id: a123, request: a123"] (an array of strings)
			data.ID = "123"
			data.Req = test.reqString

			// marshal data to json and then translate to string to encode as base64
			out, err := json.Marshal(data)
			if err != nil {
				t.Errorf("Error marshaling json for test")
			}
			testData := []string{"data", string(out)}

			// setdata in the event
			myEvent.SetData(cloudevents.ApplicationJSON, testData)

			theResponse := consumeEvent(myEvent)
			got := theResponse
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

func getRequestString(theReq *http.Request, t *testing.T) string {
	// write the request into b
	var b = &bytes.Buffer{}
	if err := theReq.Write(b); err != nil {
		t.Errorf("Error writing request to buffer")
	}
	return b.String()
}
