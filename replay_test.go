package replay

import (
	"github.com/nbio/st"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestReplay(t *testing.T) {
	replayed := make(chan bool)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		replayed <- true
	}))
	defer ts.Close()

	called := make(chan bool)
	go func() {
		replay := New(ts.URL)
		req := &http.Request{Header: make(http.Header), URL: &url.URL{}}

		replay.HandleHTTP(nil, req, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called <- true
		}))
	}()

	st.Expect(t, <-called, true)
	st.Expect(t, <-replayed, true)
}
