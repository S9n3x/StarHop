package meta

type MetaInfo struct {
	Version  string
	Port     string
	DeviceID string
}

var Info = MetaInfo{
	Version: "v0.0.0",
}
