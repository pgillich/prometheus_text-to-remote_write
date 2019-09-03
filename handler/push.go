package handler

import (
	"context"
	//"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang/glog"

	"github.com/spf13/viper"

	//config_util "github.com/prometheus/common/config"
	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/model"

	//"github.com/prometheus/prometheus/storage/remote/client"
	dto "github.com/prometheus/client_model/go"
	//"github.com/prometheus/prometheus/pkg/timestamp"
	"github.com/prometheus/prometheus/prompb"

	"github.com/pgillich/prometheus_text-to-remote_write/conf"
	"github.com/pgillich/prometheus_text-to-remote_write/remote"
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

		ProcessSeries(metricFamilies)
	}
}

// Timestamp series are listed to labels
func ProcessSeries(metricFamilies map[string]*dto.MetricFamily) {
	labelsToSeries := map[string]prompb.TimeSeries{}

	mergeMetrics(labelsToSeries, metricFamilies)

	serverURL, err := url.Parse(viper.GetString(conf.OPT_WRITE_TO))
	if err != nil {
		util.PrintFatalf("Runtime error: %+v\n", err)
	}
	glog.V(2).Infof("%s: %v\n", util.FUNCTION_NAME_SHORT(), serverURL)

	cc := remote.ClientConfig{
		URL:     serverURL,
		Timeout: time.Second,
	}

	c, err := remote.NewClient(0, &cc)
	if err != nil {
		glog.Warningf("%s: Client error: %+v\n", util.FUNCTION_NAME_SHORT(), err)
		return
	}

	writeRequest := SeriesToWriteRequest(labelsToSeries)
	util.LogObjAsJson(2, writeRequest, "writeRequest", true)
	err = c.Store(context.Background(), writeRequest)
	if err != nil {
		glog.Warningf("%s: Store error: %+v\n", util.FUNCTION_NAME_SHORT(), err)
	}
}

// Idea from github.com/prometheus/prometheus/storage/remote/codec.go:ToWriteRequest
func SeriesToWriteRequest(series map[string]prompb.TimeSeries) *prompb.WriteRequest {
	req := &prompb.WriteRequest{
		Timeseries: make([]prompb.TimeSeries, 0, len(series)),
	}

	for _, s := range series {
		glog.V(2).Infof("%s: serie %v\n", util.FUNCTION_NAME_SHORT(), s)
		req.Timeseries = append(req.Timeseries, s)
	}

	return req
}

// Idea from github.com/prometheus/prometheus/documentation/
//           examples/remote_storage/remote_storage_adapter/influxdb/client.go:mergeResult
func mergeMetrics(labelsToSeries map[string]prompb.TimeSeries, metricFamilies map[string]*dto.MetricFamily) error {
	for _, m := range metricFamilies {
		name := m.GetName()
		glog.V(2).Infof("%s: name = %v (%v)\n", util.FUNCTION_NAME_SHORT(), name, m.String())
		switch m.GetType() {
		case dto.MetricType_HISTOGRAM, dto.MetricType_SUMMARY:
			glog.Warningf("%s: Not supported metric type: %v, %v\n", util.FUNCTION_NAME_SHORT(),
				m.String(), m,
			)
			continue
		}
		for _, s := range m.GetMetric() {
			glog.V(2).Infof("%s: s.GetLabel() = %v\n", util.FUNCTION_NAME_SHORT(), s.GetLabel())
			k := concatLabels(name, s.GetLabel())
			glog.V(2).Infof("%s: k = %v\n", util.FUNCTION_NAME_SHORT(), k)
			ts, ok := labelsToSeries[k]
			if !ok {
				ts = prompb.TimeSeries{
					Labels: tagsToLabelPairs(name, s.GetLabel()),
				}
				labelsToSeries[k] = ts
			}

			value := float64(0)
			switch m.GetType() {
			case dto.MetricType_COUNTER:
				value = s.GetCounter().GetValue()
			case dto.MetricType_GAUGE:
				value = s.GetGauge().GetValue()
			case dto.MetricType_UNTYPED:
				value = s.GetUntyped().GetValue()
			}

			ts.Samples = append(ts.Samples, prompb.Sample{
				Timestamp: s.GetTimestampMs(),
				Value:     value,
			})
			glog.V(2).Infof("%s: ts = %v (%v)\n", util.FUNCTION_NAME_SHORT(), ts, m.String())
		}
		glog.V(2).Infof("%s:\n labelsToSeries = %v\n\n", util.FUNCTION_NAME_SHORT(), labelsToSeries)

	}

	util.LogObjAsJson(2, labelsToSeries, "labelsToSeries", true)

	return nil
}

// Idea from github.com/prometheus/prometheus/documentation/
//           examples/remote_storage/remote_storage_adapter/influxdb/client.go:tagsToLabelPairs
func tagsToLabelPairs(name string, labels []*dto.LabelPair) []prompb.Label {
	pairs := make([]prompb.Label, 0, len(labels)+1)
	for _, kv := range labels {
		pairs = append(pairs, prompb.Label{
			Name:  kv.GetName(),
			Value: kv.GetValue(),
		})
	}
	pairs = append(pairs, prompb.Label{
		Name:  model.MetricNameLabel,
		Value: name,
	})
	return pairs
}

// Same to github.com/prometheus/prometheus/documentation/
//         examples/remote_storage/remote_storage_adapter/influxdb/client.go:concatLabels
func concatLabels(name string, labels []*dto.LabelPair) string {
	// 0xff cannot cannot occur in valid UTF-8 sequences, so use it
	// as a separator here.
	separator := "\xff"
	pairs := make([]string, 0, len(labels)+1)
	pairs = append(pairs, name)
	for _, kv := range labels {
		pairs = append(pairs, kv.GetName()+separator+kv.GetValue())
	}
	return strings.Join(pairs, separator)
}
