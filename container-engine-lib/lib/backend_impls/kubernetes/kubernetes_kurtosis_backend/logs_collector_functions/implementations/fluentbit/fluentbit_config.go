package fluentbit

const (
	filterRulesSeparator  = "\n	"
	outputLabelsSeparator = ", "

	renameModifyFilterRuleAction = "rename"

	//This is the "record accesor" character used by Fluentbit to dinamically get content from
	//a log stream in JSON format
	labelsVarPrefix = "$"
)

// TODO: Maybe refactor with the struct we're already using the docker backend codebase.
type FluentbitConfig struct {
	Service *Service
	Input   *Input
	Filter  *Filter
	Output  *Output
}

type Service struct {
	LogLevel          string
	HttpServerEnabled string
	HttpServerHost    string
	HttpServerPort    uint16
	StoragePath       string
}

type Input struct {
	Name        string
	Listen      string
	Port        uint16
	StorageType string
}

type Filter struct {
	Name  string
	Match string
	Rules []string
}

type Output struct {
	Name  string
	Match string
	Host  string
	Port  uint16
}

func newDefaultFluentbitConfigForKurtosisCentralizedLogs(
	fluentdHost string,
	fluentdPort uint16,
	tcpPortNumber uint16,
	httpPortNumber uint16,
) *FluentbitConfig {
	return &FluentbitConfig{
		Service: &Service{
			LogLevel:          logLevel,
			HttpServerEnabled: httpServerEnabledValue,
			HttpServerHost:    httpServerLocalhost,
			HttpServerPort:    httpPortNumber,
			StoragePath:       filesystemBufferStorageDirpath,
		},
		Input: &Input{
			Name:        inputName,
			Listen:      inputListenIP,
			Port:        tcpPortNumber,
			StorageType: inputFilesystemStorageType,
		},
		Output: &Output{
			Name:  "forward",
			Match: matchAllRegex,
			Host:  fluentdHost,
			Port:  fluentdPort,
		},
	}
}
