package kurtosis_context

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
