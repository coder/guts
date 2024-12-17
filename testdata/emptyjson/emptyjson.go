package emptyjson

type JSONTag struct {
	HostName string
	CertName string `json:",omitempty"`
}
