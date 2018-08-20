package remote

// from github.com/prometheus/prometheus/storage/remote/client.go

import (
	"bufio"
	"bytes"
	"context"
	"net/http"
	"net/url"
	"time"
	//"encoding/json"
	"fmt"
	"io"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	"golang.org/x/net/context/ctxhttp"

	//"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/prompb"
)

const maxErrMsgLen = 256

// MODIFIED
// Client allows reading and writing from/to a remote HTTP endpoint.
type Client struct {
	//index   int // Used to differentiate clients in metrics.
	//url     *config_util.URL
	url     *url.URL
	client  *http.Client
	timeout time.Duration
}

// MODIFIED
// ClientConfig configures a Client.
type ClientConfig struct {
	//URL     *config_util.URL
	URL *url.URL
	//Timeout model.Duration
	Timeout time.Duration
	//HTTPClientConfig config_util.HTTPClientConfig
}

// MODIFIED
// NewClient creates a new Client.
func NewClient(index int, conf *ClientConfig) (*Client, error) {
	/*
		httpClient, err := config_util.NewClientFromConfig(conf.HTTPClientConfig, "remote_storage")
		if err != nil {
			return nil, err
		}
	*/
	return &Client{
		//index:   index,
		url: conf.URL,
		//client:  httpClient,
		timeout: time.Duration(conf.Timeout),
	}, nil
}

type recoverableError struct {
	error
}

// Store sends a batch of samples to the HTTP endpoint.
func (c *Client) Store(ctx context.Context, req *prompb.WriteRequest) error {
	data, err := proto.Marshal(req)
	if err != nil {
		return err
	}

	compressed := snappy.Encode(nil, data)
	httpReq, err := http.NewRequest("POST", c.url.String(), bytes.NewReader(compressed))
	if err != nil {
		// Errors from NewRequest are from unparseable URLs, so are not
		// recoverable.
		return err
	}
	httpReq.Header.Add("Content-Encoding", "snappy")
	httpReq.Header.Set("Content-Type", "application/x-protobuf")
	httpReq.Header.Set("X-Prometheus-Remote-Write-Version", "0.1.0")
	httpReq = httpReq.WithContext(ctx)

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	httpResp, err := ctxhttp.Do(ctx, c.client, httpReq)
	if err != nil {
		// Errors from client.Do are from (for example) network errors, so are
		// recoverable.
		return recoverableError{err}
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode/100 != 2 {
		scanner := bufio.NewScanner(io.LimitReader(httpResp.Body, maxErrMsgLen))
		line := ""
		if scanner.Scan() {
			line = scanner.Text()
		}
		err = fmt.Errorf("server returned HTTP status %s: %s", httpResp.Status, line)
	}
	if httpResp.StatusCode/100 == 5 {
		return recoverableError{err}
	}
	return err
}
