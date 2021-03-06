# Prometheus text to remote_write

It's a microservice receiving Prometheus text exposition format and sending to Prometheus remote_write.

The interest of this project is importing old (1-day, 1-year) values into Prometheus for making offline analysis. Limitations of Pushgateway incoming format (timestamp is not enabled) and scraper (older than 5-minute values aren't processed by Prometheus) don't give 'legal' way to import old measurements. There is a 'semi-legal' way to import it: sendig data directly to the Prometheus DB by remote_write protocol.

Received text is converted to object representation by Prometheus expfmt library.
Imported objects are transformed to Prometheus protobuf format by Prometheus libraries,
following Prometheus remote_storage_adapter example for InfluxDB.

See background info at [Details](doc/details.md).

Docker images are pushed to Docker hub: [pgillich/prometheus_text-to-remote_write](https://hub.docker.com/r/pgillich/prometheus_text-to-remote_write/)

*It's in alpha phase. No automatic tests and source code should be refactored. If you would like to improve it, you are welcomed! ;-)*

# Protocols

The service expects Prometheus text expose format, described here: https://prometheus.io/docs/instrumenting/exposition_formats/

Example for the receiving input:
```
storage_used{DC="operator.com",Network="prom-lab",Region="R170",Host="host-1",Mount="/"} 3756675072 1484564635000
storage_used_p{DC="operator.com",Network="prom-lab",Region="R170",Host="host-1",Mount="/"} 7.1 1484564635000
```

The service sends data to target on Prometheus remote_write protocol.

# Supported metric types

Below metric types are supported:
* Counter
* Gauge
* Untyped (Counter or Gauge)

Below metric types are NOT supported:
* Histogram
* Summary

# Usage

Service can be started as a container, for example (with default env variables):
```
docker run --rm -P --env RECEIVE_ON=':9099' --env RECEIVE_PATH='/' \
    --env WRITE_TO='http://influxdb:8086/api/v1/prom/write?u=prom&p=prom&db=prometheus' --env GLOG_V=0 \
	pgillich/prometheus_text-to-remote_write
```
The default exposed port is 9099.

Example for starting container for testing, see [Testing](#Testing) below:
```
docker run --rm -P --env WRITE_TO="http://172.17.0.1:1234/receive" --env GLOG_V=2 pgillich/prometheus_text-to-remote_write
```

Mapping CLI options to environment variables (including Glog):

| CLI option | Environment variable |
| --- | --- |
| receive-on | RECEIVE_ON |
| receive-path | RECEIVE_PATH |
| write-to | WRITE_TO |
| v | GLOG_V |
| alsologtostderr | GLOG_ALSOLOGTOSTDERR |
| log_backtrace_at | GLOG_LOG_BACKTRACE_AT |
| log_dir | GLOG_LOG_DIR |
| logtostderr | GLOG_LOGTOSTDERR |
| stderrthreshold | GLOG_STDERRTHRESHOLD |
| vmodule | GLOG_VMODULE |

# Repo config

A subdirectory from Prometheus repo (prometheus/documentation/examples/remote_storage/example_write_adapter) is linked for making test target.
FYI, below commands were executed (you don't have to do it):
```
git clone https://github.com/pgillich/prometheus_text-to-remote_write.git
cd prometheus_text-to-remote_write/
git remote add -f -t release-2.2 --no-tags prometheus https://github.com/prometheus/prometheus.git
git read-tree --prefix=example_write_adapter/ -u prometheus/release-2.2:documentation/examples/remote_storage/example_write_adapter
git add .
git commit -m 'Linking Prometheus example_write_adapter'
git push
```
See more details: https://stackoverflow.com/questions/23937436/add-subdirectory-of-remote-repo-with-git-subtree

# Build

Docker image can be built by following command:
```
docker build -t pgillich/prometheus_text-to-remote_write:vA.B .
```
Where A.B is the version number

Binary executable can be built by following command:
```
go build
```

# Testing

Binary executable can be tested without any real DB backend by github.com/prometheus/prometheus/documentation/examples/remote_storage/example_write_adapter, for example (in separated shells):
```
~/go/src/github.com/prometheus/prometheus/documentation/examples/remote_storage/example_write_adapter$ ./example_write_adapter

~/go/src/github.com/pgillich/prometheus_text-to-remote_write$ ./prometheus_text-to-remote_write service --write-to "http://172.17.0.1:1234/receive" --v 2

~/go/src/github.com/pgillich/prometheus_text-to-remote_write$ curl -X PUT --data-binary @test/data/sample-2.txt localhost:9099
```
If it runs in container, the target address should be the container IP address, instead of `localhost`.

# Version info

Output of dep status:
```
PROJECT                                           CONSTRAINT     VERSION        REVISION  LATEST   PKGS USED
github.com/fsnotify/fsnotify                      v1.4.7         v1.4.7         c282820   v1.4.7   1   
github.com/gogo/protobuf                          v1.0.0         v1.0.0         1adfc12   v1.0.0   5   
github.com/golang/glog                            branch master  branch master  23def4e   23def4e  1   
github.com/golang/protobuf                        v1.1.0         v1.1.0         b4deda0   v1.1.0   8   
github.com/golang/snappy                          branch master  branch master  2e65f85   2e65f85  1   
github.com/grpc-ecosystem/grpc-gateway            v1.4.1         v1.4.1         9258377   v1.4.1   3   
github.com/hashicorp/hcl                          branch master  branch master  ef8a98b   ef8a98b  10  
github.com/inconshreveable/mousetrap              v1.0           v1.0           76626ae   v1.0     1   
github.com/magiconair/properties                  v1.8.0         v1.8.0         c235336   v1.8.0   1   
github.com/matttproud/golang_protobuf_extensions  v1.0.1         v1.0.1         c12348c   v1.0.1   1   
github.com/mitchellh/mapstructure                 branch master  branch master  bb74f1d   f15292f  1   
github.com/pelletier/go-toml                      v1.2.0         v1.2.0         c01d127   v1.2.0   1   
github.com/prometheus/client_model                branch master  branch master  99fa1f4   5c3871d  1   
github.com/prometheus/common                      branch master  branch master  7600349   c7de230  3   
github.com/prometheus/prometheus                  ~2.2.1         v2.2.1         bc6058c   v2.2.1   1   
github.com/spf13/afero                            v1.1.1         v1.1.1         787d034   v1.1.1   2   
github.com/spf13/cast                             v1.2.0         v1.2.0         8965335   v1.2.0   1   
github.com/spf13/cobra                            ^0.0.3         v0.0.3         ef82de7   v0.0.3   1   
github.com/spf13/jwalterweatherman                branch master  branch master  7c0cea3   14d3d4c  1   
github.com/spf13/pflag                            ^1.0.1         v1.0.1         583c0c0   v1.0.2   1   
github.com/spf13/viper                            ^1.0.2         v1.0.2         b5e8006   v1.1.0   1   
golang.org/x/net                                  branch master  branch master  4cb1c02   aaf6012  8   
golang.org/x/sys                                  branch master  branch master  7138fd3   1c95834  1   
golang.org/x/text                                 v0.3.0         v0.3.0         f21a4df   v0.3.0   14  
google.golang.org/genproto                        branch master  branch master  ff3583e   c66870c  2   
google.golang.org/grpc                            v1.13.0        v1.13.0        168a619   v1.13.0  25  
gopkg.in/yaml.v2                                  v2.2.1         v2.2.1         5420a8b   v2.2.1   1   
```
