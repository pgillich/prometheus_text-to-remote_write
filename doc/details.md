# Parsing received data

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

# Preparing data to send

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
