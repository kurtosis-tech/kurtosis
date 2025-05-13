package metrics_client

type MetricsClient interface {
	TrackShouldSendMetricsUserElection(didUserAcceptSendingMetrics bool) error
	TrackUserSharedEmailAddress(userSharedEmailAddress string) error
	TrackCreateEnclave(enclaveId string, isSubnetworkingEnabled bool) error
	TrackStopEnclave(enclaveId string) error
	TrackDestroyEnclave(enclaveId string) error
	TrackKurtosisRun(packageId string, isRemote bool, isDryRun bool, isScript bool) error
	TrackServiceUpdate(enclaveId string, serviceId string) error
	TrackStartService(enclaveId string, serviceId string) error
	TrackStopService(enclaveId string, serviceId string) error
	TrackKurtosisRunFinishedEvent(packageId string, numberOfServices int, isSuccess bool) error
	TrackKurtosisAnalyticsToggle(analyticsStatus bool) error
	close() (err error)
}
