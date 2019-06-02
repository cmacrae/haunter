// haunter provides general purpose HTTP functionality via MPP Group

// Copyright 2019 Calum MacRae. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package haunter

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/prometheus/client_golang/prometheus"
)

const mppAPI = "https://api.myprivateproxy.net/v1/fetchProxies/json"

// Prometheus metrics
var (
	httpReqs = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "haunter_external_http_requests_total",
			Help: "How many external HTTP requests processed, partitioned by status code, method, and proxy IP.",
		},
		[]string{"code", "method", "proxy_ip"},
	)

	// TODO: Not sure a counter makes sense here...
	proxyCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "haunter_proxy_count",
			Help: "How many proxy servers are configured, partitioned by IP, status, city, region, and country.",
		},
		[]string{"ip", "status", "city", "region", "country"},
	)
)

// Proxies is a container of Proxy objects
type Proxies []Proxy

// Proxy is a set of data representing HTTP proxies retrieved from GhostProxies
type Proxy struct {
	ProxyIP       string `json:"proxy_ip"`
	ProxyPort     string `json:"proxy_port"`
	Username      string `json:"username"`
	Password      string `json:"password"`
	ProxyStatus   string `json:"proxy_status"`
	ProxyCountry  string `json:"proxy_country"`
	ProxyArea     string `json:"proxy_area"`
	ProxyLocation string `json:"proxy_location"`
}

// RetryOptions is a set of parameters expressing HTTP retry behavior
type RetryOptions struct {
	Max             int
	WaitMinSecs     int
	WaitMaxSecs     int
	BackoffStepSecs int
}

// Metrics is a set of controllers for the metrics haunter exposes
type Metrics struct {
	RequestCounter bool
	ProxyCounter   bool
}

// Expose passes up Prometheus metrics to the caller, based on control switches
// which dictate what metrics the caller wishes to use
func (m Metrics) Expose() []*prometheus.CounterVec {
	var exposed []*prometheus.CounterVec
	if m.RequestCounter {
		exposed = append(exposed, httpReqs)
	}

	if m.ProxyCounter {
		exposed = append(exposed, proxyCount)
	}

	return exposed
}

// RandProxy returns a random proxy from a Provider's list of proxies
func (p Proxies) RandProxy() Proxy {
	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)
	rand := r.Intn(len(p))

	return p[rand]
}

// NewClient returns a retryablehttp.Client configured to use a random proxy
func (p Proxies) NewClient(req *retryablehttp.Request, opts RetryOptions) (*retryablehttp.Client, string, error) {
	proxy := p.RandProxy()
	proxyURL, err := url.ParseRequestURI(fmt.Sprintf("http://%s:%s", proxy.ProxyIP, proxy.ProxyPort))
	if err != nil {
		return &retryablehttp.Client{}, "", fmt.Errorf("%v", err)
	}

	client := retryablehttp.NewClient()
	client.Logger = nil
	client.RetryMax = opts.Max
	client.RetryWaitMax = time.Second * time.Duration(opts.WaitMaxSecs)
	client.RetryWaitMin = time.Second * time.Duration(opts.WaitMinSecs)

	client.Backoff = func(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
		return (time.Second * time.Duration(opts.BackoffStepSecs)) * time.Duration((attemptNum))
	}

	client.HTTPClient = &http.Client{
		Timeout: (5 * time.Second),
		Transport: &http.Transport{
			Proxy:              http.ProxyURL(proxyURL),
			ProxyConnectHeader: req.Header,
		}}

	return client, proxy.ProxyIP, nil
}

// Get performs an HTTP GET request against the given url, with any headers and retry options provided.
// It will use a random proxy to do so
func (p Proxies) Get(url string, header http.Header, o RetryOptions) (http.Response, error) {
	req, err := retryablehttp.NewRequest("GET", url, nil)
	if err != nil {
		return http.Response{}, err
	}

	req.Header = header

	client, proxyIP, err := p.NewClient(req, o)
	if err != nil {
		return http.Response{}, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return http.Response{}, err
	}

	httpReqs.WithLabelValues(fmt.Sprintf("%d", resp.StatusCode), "GET", proxyIP).Inc()

	return *resp, nil
}

// NewProvider returns a configured Provider
func NewProvider(key string) (Proxies, error) {
	if key == "" {
		return Proxies{}, fmt.Errorf("empty API key")
	}

	p := Proxies{}
	client := &http.Client{Timeout: 10 * time.Second}
	r, err := client.Get(mppAPI + "/full/showLocation/" + key)
	if err != nil {
		return Proxies{}, err
	}
	defer r.Body.Close()
	json.NewDecoder(r.Body).Decode(&p)

	for _, v := range p {
		proxyCount.WithLabelValues(
			v.ProxyIP,
			v.ProxyStatus,
			v.ProxyLocation,
			v.ProxyArea,
			v.ProxyCountry,
		).Inc()
	}

	return p, nil
}
