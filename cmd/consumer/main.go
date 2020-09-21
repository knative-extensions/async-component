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
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

type request struct {
	ID  string `json:"id"`
	Req string `json:"request"`
}

func consumeEvent(event cloudevents.Event) error {
	data := &request{}
	datastrings := make([]string, 0)
	event.DataAs(&datastrings)

	// unmarshal the string to request
	err := json.Unmarshal([]byte(datastrings[1]), data)
	if err != nil {
		fmt.Println("Error unmarshalling json")
		return err
	}

	// deserialize the request
	r := bufio.NewReader(strings.NewReader(data.Req))
	req, err := http.ReadRequest(r)
	if err != nil {
		fmt.Println("Problem reading request: ", err)
		return err
	}
	// client for sending request
	client := &http.Client{}

	// build new url - writing the request removes the URL and places in URI.
	req.URL, err = url.Parse("http://" + req.Host + req.RequestURI)
	if err != nil {
		fmt.Println("Problem parsing URL: ", req.URL)
		return err
	}
	// RequestURI must be unset for client.Do(req)
	req.RequestURI = ""
	req.Header.Del("Prefer") // We do not want to make this request as async
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Problem calling url: ", err)
		return err
	}
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
