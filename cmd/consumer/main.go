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
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

type requestData struct {
	ID        string              `json:"id"`
	ReqURL    string              `json:"url"`
	ReqBody   string              `json:"body"`
	ReqHeader map[string][]string `json:"header"`
	ReqMethod string              `json:"method"`
}

const (
	preferHeaderField = "Prefer"
	preferSyncValue   = "respond-sync"
)

func consumeEvent(event cloudevents.Event) error {
	data := &requestData{}
	datastrings := make([]string, 0)
	event.DataAs(&datastrings)
	// unmarshal the string to request
	if err := json.Unmarshal([]byte(datastrings[1]), data); err != nil {
		return fmt.Errorf("error unmarshalling json: %w", err)
	}

	// client for sending request
	client := &http.Client{}
	req, err := http.NewRequest(data.ReqMethod, data.ReqURL, strings.NewReader(data.ReqBody))
	if err != nil {
		return fmt.Errorf("unable to create new request %w", err)
	}
	req.Header = data.ReqHeader
	if req.Header == nil {
		req.Header = make(map[string][]string)
	}
	req.Header.Set(preferHeaderField, preferSyncValue) // We do not want to make this request as async
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("problem calling url: %w", err)
	}
	log.Printf("Received response: %s\n", resp.Status)
	defer resp.Body.Close()
	return nil
}

func main() {
	c, err := cloudevents.NewDefaultClient()
	if err != nil {
		log.Fatal("Failed to create client, ", err)
	}
	log.Fatal(c.StartReceiver(context.Background(), consumeEvent))
}
