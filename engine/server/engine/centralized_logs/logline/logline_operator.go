package logline

//go:generate go run github.com/dmarkham/enumer -trimprefix "LogLineOperator" -type=logLineOperator -transform=lower
type logLineOperator uint8

const (
	LogLineOperator_DoesContainText logLineOperator = iota
	LogLineOperator_DoesNotContainText
	LogLineOperator_DoesContainMatchRegex
	LogLineOperator_DoesNotContainMatchRegex
)
