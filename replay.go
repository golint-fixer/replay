package replay

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"runtime"
)

// Filter function is used to determine if a given http.Request should be replayed or not.
type Filter func(*http.Request) bool

// Modifier function is used to modify a given http.Request before replaying it.
type Modifier func(*http.Request)

// Handler function is used to
type Handler func(error, *http.Response, *http.Request)

// Replayer replays incoming HTTP traffic to one or multiple servers via isolated goroutine.
type Replayer struct {
	// Targets stores the replay URL targets.
	Targets []string

	// Filters stores the replay filter functions.
	Filters []Filter

	// Modifiers stores the replay request modifier functions.
	Modifiers []Modifier

	// Handler stores the optional replay response/error handler.
	Handler Handler

	// Client stores the http.Client to be used to replay the requests.
	// Defaults to http.DefaultClient.
	Client *http.Client
}

// New creates a new replayer ready to be attached as middleware.
func New(targets ...string) *Replayer {
	return &Replayer{Targets: targets, Client: http.DefaultClient}
}

// Modify attaches a new modifier function to the current replayer who
// is responsible to modify the http.Request to be replayed before the replay.
func (x *Replayer) Modify(fn ...Modifier) *Replayer {
	x.Modifiers = append(x.Modifiers, fn...)
	return x
}

// Filter attaches a new filter function to the current replayer who
// determines if a given request should be replayed or not.
func (x *Replayer) Filter(fn ...Filter) *Replayer {
	x.Filters = append(x.Filters, fn...)
	return x
}

// SetHandler is used to set a replay request handler, allowing the developer
// to deal with the replay response or error accordingly.
func (x *Replayer) SetHandler(fn Handler) *Replayer {
	x.Handler = fn
	return x
}

// HandleHTTP handles an incoming HTTP request received by the proxy.
func (x *Replayer) HandleHTTP(w http.ResponseWriter, r *http.Request, h http.Handler) {
	defer h.ServeHTTP(w, r)

	if len(x.Targets) == 0 {
		return
	}

	// Filter request to determine if should be replayed
	for _, filter := range x.Filters {
		if !filter(r) {
			return
		}
	}

	// Restore body to be consume by subsequent layers.
	var body io.Reader
	if r.Body != nil {
		buf, _ := ioutil.ReadAll(r.Body)
		body = bytes.NewReader(buf)
		r.Body = ioutil.NopCloser(body)
	}

	for _, target := range x.Targets {
		go x.Replay(r, target, body)
	}
}

// Replay replays the given http.Request to the given target hostname.
func (x *Replayer) Replay(r *http.Request, target string, body io.Reader) {
	// Create the replay request
	req, err := NewReplayRequest(r, target, body)
	if err != nil {
		if x.Handler != nil {
			x.Handler(err, nil, req)
		}
		return
	}

	// Trigger the request modifiers
	for _, modifier := range x.Modifiers {
		modifier(req)
	}

	res, err := x.Client.Do(req)

	if transport, ok := x.Client.Transport.(*http.Transport); ok {
		EnsureTransporterFinalized(transport)
	}

	if x.Handler != nil {
		x.Handler(err, res, req)
	}
}

// NewReplayRequest creates a new http.Request cloning on the given one replacing the target URL host.
func NewReplayRequest(req *http.Request, host string, body io.Reader) (*http.Request, error) {
	target, err := url.Parse(host)
	if err != nil {
		return nil, err
	}

	// Clone request
	r, err := http.NewRequest(req.Method, req.URL.String(), body)
	r.URL.Host = target.Host
	r.URL.Scheme = target.Scheme
	r.Header = req.Header
	r.TLS = req.TLS
	r.Host = req.Host
	r.RemoteAddr = req.RemoteAddr

	return r, nil
}

// EnsureTransporterFinalized will ensure that when the HTTP client is GCed
// the runtime will close the idle connections (so that they won't leak)
// this function was adopted from Hashicorp's go-cleanhttp package.
func EnsureTransporterFinalized(httpTransport *http.Transport) {
	runtime.SetFinalizer(&httpTransport, func(transportInt **http.Transport) {
		(*transportInt).CloseIdleConnections()
	})
}
