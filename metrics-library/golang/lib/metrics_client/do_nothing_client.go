package metrics_client

import "github.com/sirupsen/logrus"

// doNothingClient: This metrics client implementation has been created for instantiate when user rejects
// sending metrics, so it doesn't really track metrics the only logic that it contains is loging
// the traking methods calls. It also can be used for test purpose
type doNothingClient struct {
	callback Callback
}

func newDoNothingClient(callback Callback) *doNothingClient {
	return &doNothingClient{callback: callback}
}

func (client *doNothingClient) TrackShouldSendMetricsUserElection(didUserAcceptSendingMetrics bool) error {
	logrus.Debugf("Do-nothing metrics client TrackShouldSendMetricsUserElection called with argument didUserAcceptSendingMetrics '%v'; skipping sending event", didUserAcceptSendingMetrics)
	client.callback.Success()
	return nil
}

func (client *doNothingClient) TrackUserSharedEmailAddress(userSharedEmailAddress string) error {
	logrus.Debugf("Do-nothing metrics client TrackUserSharedEmailAddress called with argument; skipping sending event")
	client.callback.Success()
	return nil
}

func (client *doNothingClient) TrackCreateEnclave(enclaveId string, isSubnetworkingEnabled bool) error {
	logrus.Debugf("Do-nothing metrics client TrackCreateEnclave called with argument enclaveId '%v', isSubnetworkingEnabled '%v'; skipping sending event", enclaveId, isSubnetworkingEnabled)
	client.callback.Success()
	return nil
}

func (client *doNothingClient) TrackStopEnclave(enclaveId string) error {
	logrus.Debugf("Do-nothing metrics client TrackStopEnclave called with argument enclaveId '%v'; skipping sending event", enclaveId)
	client.callback.Success()
	return nil
}

func (client *doNothingClient) TrackDestroyEnclave(enclaveId string) error {
	logrus.Debugf("Do-nothing metrics client TrackDestroyEnclave called with argument enclaveId '%v'; skipping sending event", enclaveId)
	client.callback.Success()
	return nil
}

func (client *doNothingClient) TrackKurtosisRun(packageId string, isRemote bool, isDryRun bool, isScript bool, serializedParams string) error {
	logrus.Debugf("Do-nothing metrics client TrackKurtosisRun called with arguments packageId '%v', isRemote '%v', isDryRun '%v', isScript '%v'; skipping sending event", packageId, isRemote, isDryRun, isScript)
	client.callback.Success()
	return nil
}

func (client *doNothingClient) TrackServiceUpdate(enclaveId string, serviceId string) error {
	logrus.Debugf("Do-nothing metrics client TrackServiceUpdate called with arguments enclaveId '%v', serviceId '%v'", enclaveId, serviceId)
	client.callback.Success()
	return nil
}

func (client *doNothingClient) TrackStartService(enclaveId string, serviceId string) error {
	logrus.Debugf("Do-nothing metrics client TrackStartService called with arguments enclaveId '%v', serviceId '%v'", enclaveId, serviceId)
	client.callback.Success()
	return nil
}

func (client *doNothingClient) TrackStopService(enclaveId string, serviceId string) error {
	logrus.Debugf("Do-nothing metrics client TrackStopService called with arguments enclaveId '%v', serviceId '%v'", enclaveId, serviceId)
	client.callback.Success()
	return nil
}

func (client *doNothingClient) TrackKurtosisRunFinishedEvent(packageId string, numberOfServices int, isSuccess bool, serializedParams string) error {
	logrus.Debugf("Do-nothing metrics client TrackKurtosisRunFinishedEvent called with arguments packageId '%v', numberOfServices '%v', isSuccess '%v'; skipping sending event", packageId, numberOfServices, isSuccess)
	client.callback.Success()
	return nil
}

func (client *doNothingClient) TrackKurtosisAnalyticsToggle(analyticsStatus bool) error {
	logrus.Debugf("Do-nothing metrics client TrackKurtosisAnalyticsToggle called with arguments analyticsStatus '%v'", analyticsStatus)
	client.callback.Success()
	return nil
}

func (client *doNothingClient) close() (err error) {
	logrus.Debugf("Do-nothing metrics client close method called")
	return nil
}
