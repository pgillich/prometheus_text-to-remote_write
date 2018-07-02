package handler

import (
	"bufio"
	"bytes"
	"context"
	//"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/golang/glog"

	//"github.com/spf13/viper"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	"golang.org/x/net/context/ctxhttp"

	//config_util "github.com/prometheus/common/config"
	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/model"
	//"github.com/prometheus/prometheus/storage/remote/client"
	dto "github.com/prometheus/client_model/go"
	//"github.com/prometheus/prometheus/pkg/timestamp"
	"github.com/prometheus/prometheus/prompb"

	//"github.com/pgillich/prometheus_text-to-remote_write/cmd"
	"github.com/pgillich/prometheus_text-to-remote_write/util"
)

func HandlePush(w http.ResponseWriter, req *http.Request) {
	glog.V(1).Infof("%s: %s %s\n", util.FUNCTION_NAME_SHORT(), req.Method, req.URL.String())
	switch req.Method {
	case "PUT", "POST":
		var parser expfmt.TextParser
		metricFamilies, _ := parser.TextToMetricFamilies(req.Body)

		glog.V(2).Infof("%s: %v\n", util.FUNCTION_NAME_SHORT(), metricFamilies)
		util.LogObjAsJson(2, metricFamilies, "metricFamilies", true)

		//ProcessSamples(metricFamilies)
		ProcessSeries(metricFamilies)
	}
}

// Method from // from github.com/prometheus/prometheus/storage/remote

func ProcessSamples(metricFamilies map[string]*dto.MetricFamily) {
	samples := make([]*model.Sample, 0)
	o := expfmt.DecodeOptions{Timestamp: model.Time(model.Now())}
	for _, metricFamily := range metricFamilies {
		sampleVector, _ := expfmt.ExtractSamples(&o, metricFamily)
		glog.V(2).Infof("sampleVector = %v\n", sampleVector)
		for _, sample := range sampleVector {
			glog.V(2).Infof("sample = %v\n", sample)
			samples = append(samples, sample)
		}
	}

	serverURL, err := url.Parse("http://172.17.0.1:1234/receive" /*viper.GetString(cmd.OPT_WRITE_TO)*/)
	if err != nil {
		util.PrintFatalf("Runtime error: %+v\n", err)
	}
	glog.V(2).Infof("%v\n", serverURL)

	c := Client{
		url:     serverURL,
		timeout: time.Second,
	}

	writeRequest := ToWriteRequest(samples)
	util.LogObjAsJson(2, writeRequest, "writeRequest", true)
	Store(context.Background(), &c, writeRequest)

}

// from github.com/prometheus/prometheus/storage/remote
type Client struct {
	//index   int // Used to differentiate clients in metrics.
	//url     *config_util.URL
	url     *url.URL
	client  *http.Client
	timeout time.Duration
}

// from github.com/prometheus/prometheus/storage/remote
func Store(ctx context.Context, c *Client, req *prompb.WriteRequest) error {
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

const maxErrMsgLen = 256

type recoverableError struct {
	error
}

// from github.com/prometheus/prometheus/storage/remote/codec.go
// ToWriteRequest converts an array of samples into a WriteRequest proto.
func ToWriteRequest(samples []*model.Sample) *prompb.WriteRequest {
	req := &prompb.WriteRequest{
		Timeseries: make([]*prompb.TimeSeries, 0, len(samples)),
	}

	for _, s := range samples {
		ts := prompb.TimeSeries{
			Labels: MetricToLabelProtos(s.Metric),
			Samples: []*prompb.Sample{
				{
					Value:     float64(s.Value),
					Timestamp: int64(s.Timestamp),
				},
			},
		}
		req.Timeseries = append(req.Timeseries, &ts)
	}

	return req
}

// from github.com/prometheus/prometheus/storage/remote/codec.go
// MetricToLabelProtos builds a []*prompb.Label from a model.Metric
func MetricToLabelProtos(metric model.Metric) []*prompb.Label {
	labels := make([]*prompb.Label, 0, len(metric))
	for k, v := range metric {
		labels = append(labels, &prompb.Label{
			Name:  string(k),
			Value: string(v),
		})
	}
	sort.Slice(labels, func(i int, j int) bool {
		return labels[i].Name < labels[j].Name
	})
	return labels
}

// Method from github.com/prometheus/prometheus/documentation/examples/remote_storage/remote_storage_adapter/influxdb

// Timestamp series are listed to labels
func ProcessSeries(metricFamilies map[string]*dto.MetricFamily) {
	labelsToSeries := map[string]*prompb.TimeSeries{}
	glog.V(2).Infof("%v\n", labelsToSeries)

	mergeMetrics(labelsToSeries, metricFamilies)

	serverURL, err := url.Parse("http://172.17.0.1:1234/receive" /*viper.GetString(cmd.OPT_WRITE_TO)*/)
	if err != nil {
		util.PrintFatalf("Runtime error: %+v\n", err)
	}
	glog.V(2).Infof("%v\n", serverURL)

	c := Client{
		url:     serverURL,
		timeout: time.Second,
	}

	writeRequest := SeriesToWriteRequest(labelsToSeries)
	util.LogObjAsJson(2, writeRequest, "writeRequest", true)
	Store(context.Background(), &c, writeRequest)
}

func SeriesToWriteRequest(series map[string]*prompb.TimeSeries) *prompb.WriteRequest {
	req := &prompb.WriteRequest{
		Timeseries: make([]*prompb.TimeSeries, 0, len(series)),
	}

	for _, s := range series {
		glog.V(2).Infof("serie %v\n", s)
		req.Timeseries = append(req.Timeseries, s)
	}

	return req
}

// Method from github.com/prometheus/prometheus/documentation/examples/remote_storage/remote_storage_adapter/influxdb
// Modified func of client.go:mergeResult
func mergeMetrics(labelsToSeries map[string]*prompb.TimeSeries, metricFamilies map[string]*dto.MetricFamily) error {
	for _, m := range metricFamilies {
		name := m.GetName()
		glog.V(2).Infof("%s: name = %v\n", util.FUNCTION_NAME_SHORT(), name)
		for _, s := range m.GetMetric() {
			glog.V(2).Infof("%s: s.GetLabel() = %v\n", util.FUNCTION_NAME_SHORT(), s.GetLabel())
			k := concatLabels(name, s.GetLabel())
			glog.V(2).Infof("%s: k = %v\n", util.FUNCTION_NAME_SHORT(), k)
			ts, ok := labelsToSeries[k]
			if !ok {
				ts = &prompb.TimeSeries{
					Labels: tagsToLabelPairs(name, s.GetLabel()),
				}
				labelsToSeries[k] = ts
			}

			ts.Samples = append(ts.Samples, &prompb.Sample{
				Timestamp: s.GetTimestampMs(),
				Value:     s.GetUntyped().GetValue(),
			})
			glog.V(2).Infof("%s: ts = %v\n", util.FUNCTION_NAME_SHORT(), ts)
		}
		glog.V(2).Infof("%s:\n labelsToSeries = %v\n\n", util.FUNCTION_NAME_SHORT(), labelsToSeries)

	}

	util.LogObjAsJson(2, labelsToSeries, "labelsToSeries", true)

	return nil
}

func tagsToLabelPairs(name string, labels []*dto.LabelPair) []*prompb.Label {
	pairs := make([]*prompb.Label, 0, len(labels)+1)
	for _, kv := range labels {
		pairs = append(pairs, &prompb.Label{
			Name:  kv.GetName(),
			Value: kv.GetValue(),
		})
	}
	pairs = append(pairs, &prompb.Label{
		Name:  model.MetricNameLabel,
		Value: name,
	})
	return pairs
}

/*
func tagsToLabelPairs(name string, tags []*dto.LabelPair) []*prompb.Label {
	pairs := make([]*prompb.Label, 0, len(tags))
	for k, v := range tags {
		if v == "" {
			// If we select metrics with different sets of labels names,
			// InfluxDB returns *all* possible tag names on all returned
			// series, with empty tag values on series where they don't
			// apply. In Prometheus, an empty label value is equivalent
			// to a non-existent label, so we just skip empty ones here
			// to make the result correct.
			continue
		}
		pairs = append(pairs, &prompb.Label{
			Name:  k,
			Value: v,
		})
	}
	pairs = append(pairs, &prompb.Label{
		Name:  model.MetricNameLabel,
		Value: name,
	})
	return pairs
}
*/

func concatLabels(name string, labels []*dto.LabelPair) string {
	// 0xff cannot cannot occur in valid UTF-8 sequences, so use it
	// as a separator here.
	//separator := "\xff"
	separator := "_"
	pairs := make([]string, 0, len(labels)+1)
	pairs = append(pairs, name)
	for _, kv := range labels {
		pairs = append(pairs, kv.GetName()+separator+kv.GetValue())
	}
	return strings.Join(pairs, separator)
}

/*
func valuesToSamples(values [][]interface{}) ([]*prompb.Sample, error) {
	samples := make([]*prompb.Sample, 0, len(values))
	for _, v := range values {
		if len(v) != 2 {
			return nil, fmt.Errorf("bad sample tuple length, expected [<timestamp>, <value>], got %v", v)
		}

		jsonTimestamp, ok := v[0].(json.Number)
		if !ok {
			return nil, fmt.Errorf("bad timestamp: %v", v[0])
		}

		jsonValue, ok := v[1].(json.Number)
		if !ok {
			return nil, fmt.Errorf("bad sample value: %v", v[1])
		}

		timestamp, err := jsonTimestamp.Int64()
		if err != nil {
			return nil, fmt.Errorf("unable to convert sample timestamp to int64: %v", err)
		}

		value, err := jsonValue.Float64()
		if err != nil {
			return nil, fmt.Errorf("unable to convert sample value to float64: %v", err)
		}

		samples = append(samples, &prompb.Sample{
			Timestamp: timestamp,
			Value:     value,
		})
	}
	return samples, nil
}

func mergeSamples(a, b []*prompb.Sample) []*prompb.Sample {
	result := make([]*prompb.Sample, 0, len(a)+len(b))
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		if a[i].Timestamp < b[j].Timestamp {
			result = append(result, a[i])
			i++
		} else if a[i].Timestamp > b[j].Timestamp {
			result = append(result, b[j])
			j++
		} else {
			result = append(result, a[i])
			i++
			j++
		}
	}
	result = append(result, a[i:]...)
	result = append(result, b[j:]...)
	return result
}
*/
