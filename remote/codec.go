package remote

// from github.com/prometheus/prometheus/storage/remote/codec.go

import (
	"sort"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/prompb"
)

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
