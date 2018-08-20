package conf

const (
	OPT_RECEIVE_ON        = "receive-on"
	OPT_RECEIVE_PATH_TEXT = "receive-path"
	OPT_WRITE_TO          = "write-to"

	OPT_COPYSTANDARDLOGTO      = "copystandardlogto"
	OPT_GLOG_COPYSTANDARDLOGTO = "glog." + OPT_COPYSTANDARDLOGTO

	DEFAULT_RECEIVE_ON        = ":9099"
	DEFAULT_RECEIVE_PATH_TEXT = "/"
	DEFAULT_WRITE_TO          = "http://influxdb:8086/api/v1/prom/write?u=prom&p=prom&db=prometheus"
)
