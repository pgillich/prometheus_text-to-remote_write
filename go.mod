module github.com/pgillich/prometheus_text-to-remote_write

go 1.15

require (
	emperror.dev/errors v0.4.3
	github.com/gogo/protobuf v1.3.1
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/snappy v0.0.2
	github.com/moogar0880/problems v0.1.1
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/common v0.15.0
	// git rev-list -n 1 v2.22.2
	// go get github.com/prometheus/prometheus@de1c1243f4dd66fbac3e8213e9a7bd8dbc9f38b2
	github.com/prometheus/prometheus v1.8.2-0.20201116123734-de1c1243f4dd
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.6.1
	golang.org/x/net v0.0.0-20201110031124-69a78807bb2b
)

// Make sure Prometheus version is pinned as Prometheus semver does not include Go APIs.
//replace github.com/prometheus/prometheus => github.com/prometheus/prometheus v1.8.2-0.20201116123734-de1c1243f4dd
