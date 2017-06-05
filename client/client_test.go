package client

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eremetic-framework/eremetic"
)

func TestClient_AddTask(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	}))
	defer ts.Close()

	var httpClient http.Client

	c, err := New(ts.URL, &httpClient)
	if err != nil {
		t.Fatal(err)
	}

	var req eremetic.Request

	if err := c.AddTask(req); err != nil {
		t.Fatal(err)
	}
}

func TestClient_Tasks(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[{
			"id": "eremetic-id-12345"
		}]`))
	}))
	defer ts.Close()

	var httpClient http.Client

	c, err := New(ts.URL, &httpClient)
	if err != nil {
		t.Fatal(err)
	}

	tasks, err := c.Tasks()
	if err != nil {
		t.Fatal(err)
	}

	if len(tasks) != 1 {
		t.Fail()
	}
}

func TestClient_KillTask(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()

	var httpClient http.Client

	c, err := New(ts.URL, &httpClient)
	if err != nil {
		t.Fatal(err)
	}

	if err := c.Kill("1234"); err != nil {
		t.Fatal(err)
	}
}
