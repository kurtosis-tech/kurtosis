package to_logline

import (
	"github.com/dzobbe/PoTE-kurtosis/engine/server/engine/centralized_logs/logline"
	"github.com/kurtosis-tech/stacktrace"

	api_type "github.com/dzobbe/PoTE-kurtosis/api/golang/http_rest/api_types"
)

func ToLoglineLogLineFilters(logLineFilters []api_type.LogLineFilter) (logline.ConjunctiveLogLineFilters, error) {
	var conjunctiveLogLineFilters logline.ConjunctiveLogLineFilters

	for _, logLineFilter := range logLineFilters {
		var filter *logline.LogLineFilter
		operator := logLineFilter.Operator
		filterTextPattern := logLineFilter.TextPattern
		switch operator {
		case api_type.DOESCONTAINTEXT:
			filter = logline.NewDoesContainTextLogLineFilter(filterTextPattern)
		case api_type.DOESNOTCONTAINTEXT:
			filter = logline.NewDoesNotContainTextLogLineFilter(filterTextPattern)
		case api_type.DOESCONTAINMATCHREGEX:
			filter = logline.NewDoesContainMatchRegexLogLineFilter(filterTextPattern)
		case api_type.DOESNOTCONTAINMATCHREGEX:
			filter = logline.NewDoesNotContainMatchRegexLogLineFilter(filterTextPattern)
		default:
			return nil, stacktrace.NewError("Unrecognized log line filter operator '%v' in GRPC filter '%v'; this is a bug in Kurtosis", operator, logLineFilter)
		}
		conjunctiveLogLineFilters = append(conjunctiveLogLineFilters, *filter)
	}

	return conjunctiveLogLineFilters, nil
}
