// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tencentgit

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/jenkins-x/go-scm/scm"
)

const mimeJSON = "application/json"

func TestClient(t *testing.T) {
	client, err := New("https://git.code.tencent.com")
	if err != nil {
		t.Error(err)
	}
	if got, want := client.BaseURL.String(), "https://git.code.tencent.com/"; got != want {
		t.Errorf("Want Client URL %q, got %q", want, got)
	}
}

func TestClient_Base(t *testing.T) {
	client, err := New("https://server.example.com/tencentgit")
	if err != nil {
		t.Error(err)
	}
	if got, want := client.BaseURL.String(), "https://server.example.com/tencentgit/"; got != want {
		t.Errorf("Want Client URL %q, got %q", want, got)
	}
}

func TestClient_Default(t *testing.T) {
	client := NewDefault()
	if got, want := client.BaseURL.String(), "https://git.code.tencent.com/"; got != want {
		t.Errorf("Want Client URL %q, got %q", want, got)
	}
}

func TestClient_Error(t *testing.T) {
	_, err := New("http://a b.com/")
	if err == nil {
		t.Errorf("Expect error when invalid URL")
	}
}

func TestClientDo(t *testing.T) {
	var body []byte
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if v := r.Header.Get("Content-Type"); v != mimeJSON {
			t.Errorf("client.do() got Content-Type: %s, want %s", v, mimeJSON)
		}
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed: %s", err), http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()
		w.Header().Set("Content-Type", mimeJSON)
		body = bytes.TrimSpace(b)
		w.Write([]byte(`{"new":"value"}`))
	}))
	defer ts.Close()

	client, err := New(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	w := &wrapper{Client: client}
	in := map[string]string{"test": "test"}
	out := map[string]string{}

	_, err = w.do(context.Background(), "POST", "/", in, &out)
	if err != nil {
		t.Fatalf("do failed: %s", err)
	}
	wantRequestBody := []byte(`{"test":"test"}`)
	if !reflect.DeepEqual(wantRequestBody, body) {
		t.Errorf("incorrect request body: got %s, want %s", body, wantRequestBody)
	}
	wantResponseBody := map[string]string{"new": "value"}
	if !reflect.DeepEqual(out, wantResponseBody) {
		t.Errorf("incorrect response body: got %s, want %s", out, wantResponseBody)
	}
}

func TestClientDoResponseValues(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Content-Type", mimeJSON)
		h.Set("X-Request-Id", "test-id")
		h.Set("RateLimit-Limit", "200")
		h.Set("RateLimit-Remaining", "180")
		h.Set("RateLimit-Reset", "3600")
		w.Write([]byte(`{"new":"value"}`))
	}))
	defer ts.Close()

	client, err := New(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	w := &wrapper{Client: client}
	out := map[string]string{}

	resp, err := w.do(context.Background(), "POST", "/", nil, &out)
	if err != nil {
		t.Fatalf("do failed: %s", err)
	}
	if resp.ID != "test-id" {
		t.Errorf("response.ID, got %s, want %s", resp.ID, "test-id")
	}

	want := scm.Rate{
		Limit:     200,
		Remaining: 180,
		Reset:     3600,
	}

	if resp.Rate != want {
		t.Errorf("response.Rate got %#v, want %#v", resp.Rate, want)
	}
}

func testRate(res *scm.Response) func(t *testing.T) {
	return func(t *testing.T) {
		if got, want := res.Rate.Limit, 600; got != want {
			t.Errorf("Want RateLimit-Limit %d, got %d", want, got)
		}
		if got, want := res.Rate.Remaining, 599; got != want {
			t.Errorf("Want RateLimit-Remaining %d, got %d", want, got)
		}
		if got, want := res.Rate.Reset, int64(1512454441); got != want {
			t.Errorf("Want RateLimit-Reset %d, got %d", want, got)
		}
	}
}

func testPage(res *scm.Response) func(t *testing.T) {
	return func(t *testing.T) {
		if got, want := res.Page.Next, 2; got != want {
			t.Errorf("Want next page %d, got %d", want, got)
		}
		if got, want := res.Page.Prev, 1; got != want {
			t.Errorf("Want prev page %d, got %d", want, got)
		}
		if got, want := res.Page.First, 1; got != want {
			t.Errorf("Want first page %d, got %d", want, got)
		}
		if got, want := res.Page.Last, 5; got != want {
			t.Errorf("Want last page %d, got %d", want, got)
		}
	}
}

func testRequest(res *scm.Response) func(t *testing.T) {
	return func(t *testing.T) {
		if got, want := res.ID, "0d511a76-2ade-4c34-af0d-d17e84adb255"; got != want {
			t.Errorf("Want X-Request-Id: %q, got %q", want, got)
		}
	}
}

func TestTime_UnmarshalJSON(t1 *testing.T) {
	type args struct {
		TimeString string
	}
	tests := []struct {
		name     string
		args     args
		wantErr  bool
		wantTime time.Time
	}{
		{
			name: "",
			args: args{
				TimeString: "2012-09-19T09:29:19+0000",
			},
			wantErr:  false,
			wantTime: time.Date(2012, 9, 19, 9, 29, 19, 0, time.UTC),
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := Time{}
			if err := t.UnmarshalJSON([]byte(tt.args.TimeString)); (err != nil) != tt.wantErr {
				t1.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !t.Time.Equal(tt.wantTime) {
				t1.Errorf("UnmarshalJSON() time = %v, wantTime %v", t.Time, tt.wantTime)
			}
		})
	}
}
