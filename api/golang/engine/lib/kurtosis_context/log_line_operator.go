package kurtosis_context

//go:generate go run github.com/dmarkham/enumer -trimprefix "LogLineOperator" -type=logLineOperator -transform=lower
type logLineOperator uint8
const (
	logLineOperator_DoesContainText logLineOperator = iota
	logLineOperator_DoesNotContainText
	logLineOperator_DoesContainMatchRegex
	logLineOperator_DoesNotContainMatchRegex
)
