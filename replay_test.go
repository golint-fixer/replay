package replay

import (
	"fmt"
	"github.com/nbio/st"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestReplay(t *testing.T) {
	replayed := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		replayed = true
		fmt.Fprintln(w, "Hello, replay")
	}))
	defer ts.Close()

	replay := New(ts.URL)
	req := &http.Request{Header: make(http.Header), URL: &url.URL{}}

	called := false
	replay.HandleHTTP(nil, req, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	time.Sleep(50 * time.Millisecond)
	st.Expect(t, called, true)
	st.Expect(t, replayed, true)
}
