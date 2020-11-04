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
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/bradleypeabody/gouuidv6"

	"github.com/go-redis/redis/v8"
	"github.com/kelseyhightower/envconfig"
)

type envInfo struct {
	StreamName       string `envconfig:"REDIS_STREAM_NAME"`
	RedisAddress     string `envconfig:"REDIS_ADDRESS"`
	RequestSizeLimit string `envconfig:"REQUEST_SIZE_LIMIT"`
}

type requestData struct {
	ID      string //`json:"id"`
	Request string //`json:"request"`
}

type redisInterface interface {
	write(ctx context.Context, s envInfo, reqJSON []byte, id string) error
}

type myRedis struct {
	client redis.Cmdable
}

// request size limit in bytes
const bitsInMB = 1000000

var env envInfo
var rc redisInterface

func main() {
	// get env info for queue
	err := envconfig.Process("", &env)
	if err != nil {
		log.Fatal(err.Error())
	}

	// set up redis client
	opts := &redis.UniversalOptions{
		Addrs: []string{env.RedisAddress},
	}
	rc = &myRedis{
		client: redis.NewUniversalClient(opts),
	}

	// Start an HTTP Server
	http.HandleFunc("/", checkHeaderAndServe)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

/*
check for a Prefer: respond-async header.
if async is preferred, then write request to redis.
if symnchronous is preferred, then proxy the request.
*/
func checkHeaderAndServe(w http.ResponseWriter, r *http.Request) {
	// If request body exists, check that length doesn't exceed limit.
	requestSizeInt, err := strconv.Atoi(env.RequestSizeLimit)
	if err != nil {
		log.Fatal("Error parsing request size string to integer")
	}
	if r.Body != nil {
		r.Body = http.MaxBytesReader(w, r.Body, int64(requestSizeInt))
	}
	// write the request into buff
	var buff = &bytes.Buffer{}
	if err := r.Write(buff); err != nil {
		if err.Error() == "http: request body too large" {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			log.Print("Error writing to buffer: ", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	// Translate to string, then json including an id.
	reqString := buff.String()
	id := gouuidv6.NewFromTime(time.Now()).String()
	reqData := requestData{
		ID:      id,
		Request: reqString,
	}
	reqJSON, err := json.Marshal(reqData)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(w, "Failed to marshal request: ", err)
		return
	}
	// write the request information to the storage
	if writeErr := rc.write(r.Context(), env, reqJSON, reqData.ID); writeErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Error asynchronous writing request to storage ", writeErr)
		return
	}
	w.WriteHeader(http.StatusAccepted)
	return
}

// function to write to redis stream
func (mr *myRedis) write(ctx context.Context, s envInfo, reqJSON []byte, id string) (err error) {
	strCMD := mr.client.XAdd(ctx, &redis.XAddArgs{
		Stream: s.StreamName,
		Values: map[string]interface{}{
			"data": reqJSON,
		},
	})
	if strCMD.Err() != nil {
		return fmt.Errorf("failed to publish %q: %v", id, strCMD.Err())
	}
	return
}
