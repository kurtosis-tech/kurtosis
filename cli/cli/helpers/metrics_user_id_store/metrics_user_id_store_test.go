package metrics_user_id_store

import (
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	//!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	//DO NOT CHANGE THIS VALUE
	//!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	expectedApplicationID = "kurtosis"
)

//The application ID constant in this package has to be always "kurtosis"
//so we control that it does not change
func TestApplicationIdDoesNotChange(t *testing.T) {
	require.Equal(t, expectedApplicationID, applicationID)
}
