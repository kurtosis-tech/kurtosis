package metrics_client

type MetricsClient interface {
	TrackShouldSendMetricsUserElection(didUserAcceptSendingMetrics bool) error
	TrackUserSharedEmailAddress(userSharedEmailAddress string) error
	TrackCreateEnclave(enclaveId string, isSubnetworkingEnabled bool) error
	TrackStopEnclave(enclaveId string) error
	TrackDestroyEnclave(enclaveId string) error
	TrackKurtosisRun(packageId string, isRemote bool, isDryRun bool, isScript bool) error
	TrackServiceUpdate(packageId string, isRemote bool, isDryRun bool, isScript bool) error
	TrackStartService(packageId string, isRemote bool, isDryRun bool, isScript bool) error
	TrackStopService(packageId string, isRemote bool, isDryRun bool, isScript bool) error
	TrackKurtosisRunFinishedEvent(packageId string, numberOfServices int, isSuccess bool) error
	TrackKurtosisAnalyticsToggle(analyticsStatus bool) error
	close() (err error)
}
