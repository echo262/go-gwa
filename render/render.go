package render

// graphite web render implementation
//
// https://graphite.readthedocs.io/en/latest/render_api.html
//

import (
	"io"
	"io/ioutil"
	"strconv"
	"time"

	"encoding/json"

	"fmt"

	"net/http"

	"context"
	"net/url"
)

const (
	defaultTimeout = 5 * time.Second
)

// Client Renderer client
type Client struct {
	httpClient *http.Client
	renderURL  *url.URL
}

// MetricRequest configuration for requesting metrics
type MetricRequest struct {
	From   string
	Until  string
	Target []string
}

// Metrics collection of graphite data for json representation
type Metrics []Metric

// Metric graphite metric representation
type Metric struct {
	Target     string      `json:"target"`
	DataPoints []DataPoint `json:"datapoints"`
}

// DataPoint graphite metric data
type DataPoint struct {
	Value     float64
	TimeStamp JSONTime
}

// JSONTime struct to help marshal/unmarshal timestamps
type JSONTime struct {
	time.Time
}

// Encode helper that encode struct to data compatible with Uri.RawQuery
func (m *MetricRequest) Encode() string {
	v := url.Values{}
	v.Set("format", "json")
	if len(m.From) > 0 {
		v.Set("from", m.From)
	}

	if len(m.Until) > 0 {
		v.Set("until", m.Until)
	}

	for _, t := range m.Target {
		if len(t) > 0 {
			v.Add("target", t)
		}
	}

	return v.Encode()
}

// UnmarshalJSON Implement unmarshaler interface useful for timestamp unmarshal
func (t *JSONTime) UnmarshalJSON(buf []byte) error {
	var err error
	stamp, err := strconv.ParseInt(string(buf), 10, 64)
	if err != nil {
		return err
	}
	t.Time = time.Unix(stamp, 0)
	return nil
}

// UnmarshalJSON Implement json unmarshaller for graphite datapoint
func (d *DataPoint) UnmarshalJSON(buf []byte) error {
	tmp := make([]json.RawMessage, 0, 2)
	err := json.Unmarshal(buf, &tmp)
	if err != nil {
		return err
	}

	if len(tmp) != 2 {
		return fmt.Errorf("Wrong argument count for datapoint")
	}

	err = json.Unmarshal(tmp[0], &d.Value)
	if err != nil {
		return err
	}

	err = json.Unmarshal(tmp[1], &d.TimeStamp)
	if err != nil {
		return err
	}

	return nil
}

// NewClient helper for create new client
func NewClient(httpClient *http.Client, renderURI string) (*Client, error) {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: defaultTimeout,
		}
	}

	u, err := url.Parse(renderURI)
	if err != nil {
		return nil, err
	}

	c := &Client{
		httpClient: httpClient,
		renderURL:  u,
	}

	return c, nil
}

// Fetch retrieve data from graphite
func (c *Client) Fetch(ctx context.Context, mr *MetricRequest) (*Metrics, error) {
	u := *c.renderURL // copy
	u.RawQuery = mr.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, nil
	}

	req = req.WithContext(ctx)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-agent", "FooBAr/200.1")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		// Drain up to 512 bytes and close the body to let the Transport reuse the connection
		io.CopyN(ioutil.Discard, resp.Body, 512)
		resp.Body.Close()
	}()

	// Skip bad answers
	if s := resp.StatusCode; s < 200 || s >= 300 {
		return nil, fmt.Errorf("Error: %v", s)
	}
	m := new(Metrics)

	if resp.StatusCode != http.StatusNoContent {
		err := json.NewDecoder(resp.Body).Decode(&m)
		if err != nil {
			return nil, err
		}
	}

	return m, nil
}
