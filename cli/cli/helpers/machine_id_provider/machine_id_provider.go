package machine_id_provider

import (
	"github.com/denisbrodbeck/machineid"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	applicationID = "kurtosis-cli"
)

func GetProtectedMachineID() (string, error) {

	//TODO create the logic to store the machine ID in the host machine

	id, err := machineid.ProtectedID(applicationID)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred generating protected machine ID")
	}
	return id, nil
}
