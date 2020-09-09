package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSimpleServer(t *testing.T) {
	testserver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	}))

	tests := []struct {
		name       string
		returncode int
	}{{
		name:       "simple request",
		returncode: 200,
	}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testserver.URL = testserver.URL + "/myname"
			request, _ := http.NewRequest(http.MethodGet, testserver.URL, nil)
			rr := httptest.NewRecorder()

			hello(rr, request)

			got := rr.Code
			want := test.returncode

			if got != want {
				t.Errorf("got %d, want %d", got, want)
			}
		})
	}
}
