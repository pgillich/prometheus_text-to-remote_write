# prometheus_text-to-remote_write
It's a microservice receiving Prometheus text exposition format and sending to Prometheus remote_write

The service expects Prometheus text expose format, described here: https://prometheus.io/docs/instrumenting/exposition_formats/

Example for the received input:
```
storage_used{DC="operator.com",Network="prom-lab",Region="R170",Host="host-1",Mount="/"} 3756675072 1484564635000
storage_used_p{DC="operator.com",Network="prom-lab",Region="R170",Host="host-1",Mount="/"} 7.1 1484564635000
```

Received text is converted to object representation by github.com/prometheus/common/expfmt/text_parse/goTextToMetricFamilies, for example (in JSON):
```
{
  "storage_used": {
    "name": "storage_used",
    "type": 3,
    "metric": [
      {
        "label": [
          {
            "name": "DC",
            "value": "operator.com"
          },
          {
            "name": "Network",
            "value": "prom-lab"
          },
          {
            "name": "Region",
            "value": "R170"
          },
          {
            "name": "Host",
            "value": "host-1"
          },
          {
            "name": "Mount",
            "value": "/"
          }
        ],
        "untyped": {
          "value": 3756675072
        },
        "timestamp_ms": 1484564635000
      }
    ]
  },
  "storage_used_p": {
    "name": "storage_used_p",
    "type": 3,
    "metric": [
      {
        "label": [
          {
            "name": "DC",
            "value": "operator.com"
          },
          {
            "name": "Network",
            "value": "prom-lab"
          },
          {
            "name": "Region",
            "value": "R170"
          },
          {
            "name": "Host",
            "value": "host-1"
          },
          {
            "name": "Mount",
            "value": "/"
          }
        ],
        "untyped": {
          "value": 7.1
        },
        "timestamp_ms": 1484564635000
      }
    ]
  }
}
```

Imported object is transformed to protobuf data by github.com/prometheus/common/expfmt/decode.go and github.com/prometheus/prometheus/storage/remote/codec.go, for example (in JSON):
```
{
  "timeseries": [
    {
      "labels": [
        {
          "name": "DC",
          "value": "operator.com"
        },
        {
          "name": "Host",
          "value": "host-1"
        },
        {
          "name": "Mount",
          "value": "/"
        },
        {
          "name": "Network",
          "value": "prom-lab"
        },
        {
          "name": "Region",
          "value": "R170"
        },
        {
          "name": "__name__",
          "value": "storage_used"
        }
      ],
      "samples": [
        {
          "value": 3756675072,
          "timestamp": 1484564635000
        }
      ]
    },
    {
      "labels": [
        {
          "name": "DC",
          "value": "operator.com"
        },
        {
          "name": "Host",
          "value": "host-1"
        },
        {
          "name": "Mount",
          "value": "/"
        },
        {
          "name": "Network",
          "value": "prom-lab"
        },
        {
          "name": "Region",
          "value": "R170"
        },
        {
          "name": "__name__",
          "value": "storage_used_p"
        }
      ],
      "samples": [
        {
          "value": 7.1,
          "timestamp": 1484564635000
        }
      ]
    }
  ]
}
```

It can be tested by github.com/prometheus/prometheus/documentation/examples/remote_storage/example_write_adapter, for example (in separated shells):
```
~/go/src/github.com/prometheus/prometheus/documentation/examples/remote_storage/example_write_adapter$ ./example_write_adapter

~/go/src/github.com/pgillich/prometheus_text-to-remote_write$ ./prometheus_text-to-remote_write service --write-to "http://179.17.0.1:1234/receive" --v 2 --logtostderr

~/go/src/github.com/pgillich/prometheus_text-to-remote_write$ curl -X PUT --data-binary @test/data/sample-2.txt localhost:9099
```

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
github.com/mitchellh/mapstructure                 branch master  branch master  bb74f1d   bb74f1d  1   
github.com/pelletier/go-toml                      v1.2.0         v1.2.0         c01d127   v1.2.0   1   
github.com/prometheus/client_model                branch master  branch master  99fa1f4   99fa1f4  1   
github.com/prometheus/common                      branch master  branch master  7600349   7600349  3   
github.com/prometheus/prometheus                  v2.3.1         v2.3.1         188ca45   v2.3.1   1   
github.com/spf13/afero                            v1.1.1         v1.1.1         787d034   v1.1.1   2   
github.com/spf13/cast                             v1.2.0         v1.2.0         8965335   v1.2.0   1   
github.com/spf13/cobra                            ^0.0.3         v0.0.3         ef82de7   v0.0.3   1   
github.com/spf13/jwalterweatherman                branch master  branch master  7c0cea3   7c0cea3  1   
github.com/spf13/pflag                            ^1.0.1         v1.0.1         583c0c0   v1.0.1   1   
github.com/spf13/viper                            ^1.0.2         v1.0.2         b5e8006   v1.0.2   1   
golang.org/x/net                                  branch master  branch master  4cb1c02   4cb1c02  8   
golang.org/x/sys                                  branch master  branch master  7138fd3   7138fd3  1   
golang.org/x/text                                 v0.3.0         v0.3.0         f21a4df   v0.3.0   14  
google.golang.org/genproto                        branch master  branch master  ff3583e   ff3583e  2   
google.golang.org/grpc                            v1.13.0        v1.13.0        168a619   v1.13.0  25  
gopkg.in/yaml.v2                                  v2.2.1         v2.2.1         5420a8b   v2.2.1   1   
```

TODO: code cleanup
