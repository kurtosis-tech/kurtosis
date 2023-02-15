package kurtosis_context

//This is an object to represent a simple log line information
type ServiceLog struct {
	//lineTime time.Time //TODO add the time from loki logs result
	content string
}

func newServiceLog(content string) *ServiceLog {
	return &ServiceLog{content: content}
}

func (serviceLog ServiceLog) GetContent() string {
	return serviceLog.content
}
