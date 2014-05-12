package crankshaft

type Config struct {
	Host        string
	Port        int
	TLSEnabled  bool `toml:"tls_enabled"`
	Clusters    []string
	BackendType string
	Statsd      StatsDConfig
	InfluxDB    InfluxDbConfig
}

type StatsDConfig struct {
	Host   string
	Port   int
	Prefix string
}

type InfluxDbConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
}

type EventChannel chan *TurbineEvent

type StatWriter interface {
	WriteEvent(event *TurbineEvent)
}

type TurbineEvent struct {
	clusterName string
	data        map[string]interface{}
}
