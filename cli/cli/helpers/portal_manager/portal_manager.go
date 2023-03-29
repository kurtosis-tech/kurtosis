package portal_manager

import (
	"context"
	portal_constructors "github.com/kurtosis-tech/kurtosis-portal/api/golang/constructors"
	portal_generated_api "github.com/kurtosis-tech/kurtosis-portal/api/golang/generated"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/host_machine_directories"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"strconv"
	"syscall"
)

const (
	portalPidFileMode       = 0600
	portalProcessPingSignal = 0

	pidNumberBase    = 10
	pidNumberBitSize = 64
)

type PortalManager struct {
	// As it's fairly new, the portal daemon might not be running. If the context is local, it's not a problem and
	// therefore this being set to nil is fine. However, if the context is remote, this should be set.
	portalClientMaybe portal_generated_api.KurtosisPortalClientClient
}

func NewPortalManager() *PortalManager {
	return &PortalManager{
		portalClientMaybe: nil,
	}
}

func (portalManager *PortalManager) IsReachable() bool {
	if err := portalManager.instantiateClientIfUnset(); err != nil {
		return false
	}
	return true
}

func (portalManager *PortalManager) GetClient() portal_generated_api.KurtosisPortalClientClient {
	return portalManager.portalClientMaybe
}

// CurrentStatus returns the status of any current Kurtosis Portal process running locally
// It returns:
// - an int corresponding to the PID of the Portal process as stored in its PID file, 0 otherwise
// - a pointer to the os.Process object corresponding to the currently running Portal process, nil if none is running
// - a boolean flag corresponding to whether the Portal process is reachable on its ports
// - an error if something unexpected happened
func (portalManager *PortalManager) CurrentStatus(ctx context.Context) (int, *os.Process, bool, error) {
	pid, err := getPidFromPidFile()
	if err != nil {
		return 0, nil, false, stacktrace.Propagate(err, "Unexpected error reading PID from PID file")
	}
	if pid == 0 {
		return 0, nil, false, nil
	}

	process, err := getRunningProcessFromPid(pid)
	if err != nil {
		return 0, nil, false, stacktrace.Propagate(err, "Unexpected error getting process for PID '%d'", pid)
	}
	if process == nil {
		return pid, nil, false, nil
	}

	if !portalManager.IsReachable() {
		return pid, process, false, nil
	}
	return pid, process, true, nil
}

// StartNew starts a new Portal process. The caller needs to make sure a process is not already running, as ports
// might conflict
// It writes the PID of the new process to the Portal PID file
// It returns the PID of the new Portal process, or an error if something went wrong
func (portalManager *PortalManager) StartNew(ctx context.Context) (int, error) {
	portalPidFilePath, err := host_machine_directories.GetPortalPidFilePath()
	if err != nil {
		return 0, stacktrace.Propagate(err, "Unable to get file path to PID file")
	}
	portalBinaryFile, err := host_machine_directories.GetPortalBinaryFilePath()
	if err != nil {
		return 0, stacktrace.Propagate(err, "Unable to get file path to PID file")
	}

	// Start portal daemon
	portalLogFilePath, err := host_machine_directories.GetPortalLogFilePath()
	if err != nil {
		return 0, stacktrace.Propagate(err, "Unable to get file path to portal log file")
	}
	// Create will truncate the file if it already exists.
	// TODO: we can potentially do log rolling here
	portalLogFile, err := os.Create(portalLogFilePath)
	if err != nil {
		return 0, stacktrace.Propagate(err, "Error truncating Kurtosis Portal log file prior to starting it")
	}
	kurtosisPortalCmd := exec.Command(portalBinaryFile)
	kurtosisPortalCmd.Stdout = portalLogFile
	kurtosisPortalCmd.Stderr = portalLogFile
	if err = kurtosisPortalCmd.Start(); err != nil {
		// TODO: maybe print the portal logs here
		return 0, stacktrace.Propagate(err, "Error trying to start Kurtosis Portal")
	}

	// Persist new PID to PID file
	newPid := kurtosisPortalCmd.Process.Pid
	if err = os.WriteFile(portalPidFilePath, []byte(strconv.Itoa(newPid)), portalPidFileMode); err != nil {
		return 0, stacktrace.Propagate(err, "Portal was successfully started, but Kurtosis failed at persisting its "+
			"PID to disk. It will have trouble next time checking whether the Portal is running or not, and you "+
			"might need to kill is manually")
	}

	// TODO: check whether it's reachable with a retry
	return newPid, nil
}

// StopExisting stops the existing Portal process, if any, and removes the PID file
// It returns an error if something went wrong
func (portalManager *PortalManager) StopExisting(_ context.Context) error {
	portalPidFilePath, err := host_machine_directories.GetPortalPidFilePath()
	if err != nil {
		return stacktrace.Propagate(err, "Unable to get file path to PID file")
	}

	pid, err := getPidFromPidFile()
	if err != nil {
		return stacktrace.Propagate(err, "Error getting current Portal PID")
	}
	if pid == 0 {
		// Not running, nothing to do
		if portalManager.IsReachable() {
			logrus.Warnf("Kurtosis Portal process not registered in Kurtosis, but it seems a Portal can be " +
				"reached locally on its ports. This is unexpected and Kurtosis cannot stop it. Was the Portal started " +
				"with something else then Kurtosis CLI? If that's the case, please kill the current Portal process and " +
				"start it using Kurtosis CLI")
			return nil
		}
		return nil
	}

	process, err := getRunningProcessFromPid(pid)
	if err != nil {
		return stacktrace.Propagate(err, "Error getting Portal process from its PID %d", pid)
	}
	if process == nil {
		if portalManager.IsReachable() {
			logrus.Warnf("Kurtosis Portal process not registered in Kurtosis, but it seems a Portal can be " +
				"reached locally on its ports. This is unexpected and Kurtosis cannot stop it. Was the Portal started " +
				"with something else then Kurtosis CLI? If that's the case, please kill the current Portal process and " +
				"start it using Kurtosis CLI")
			return nil
		}
	}

	if err = process.Signal(syscall.SIGINT); err != nil {
		logrus.Warnf("Error stopping currently running portal on PID: '%d'. It might already be stopped. "+
			"PID file will be removed. Error was: %s", pid, err.Error())
	}

	if err = os.Remove(portalPidFilePath); err != nil {
		return stacktrace.Propagate(err, "Portal was successfully stopped but Kurtosis couldn't remove the "+
			"PID file on disk. This is not critical but might hide an underlying issue.")
	}
	return nil
}

// MapPorts maps a set of remote ports locally according to the mapping provided
// It returns the set of successfully mapped ports, and potential failed ports
// An error will be returned if the set of failed port is not empty
func (portalManager *PortalManager) MapPorts(ctx context.Context, localPortToRemotePortMapping map[uint16]*services.PortSpec) (map[uint16]*services.PortSpec, map[uint16]*services.PortSpec, error) {
	successfullyMappedPorts := map[uint16]*services.PortSpec{}
	failedPorts := map[uint16]*services.PortSpec{}
	if !portalManager.IsReachable() {
		failedPorts = localPortToRemotePortMapping
		return successfullyMappedPorts, failedPorts, stacktrace.NewError("Unable to instantiate a client to the Kurtosis Portal daemon")
	}
	if portalManager.portalClientMaybe == nil {
		successfullyMappedPorts = localPortToRemotePortMapping
		// context is local and portal not present. Port mapping doesn't make sense in a local context anyway, return
		// successfully
		logrus.Debug("Context is local, no ports to map via the Portal as they are naturally exposed")
		return successfullyMappedPorts, failedPorts, nil
	}

	for localPort, remotePort := range localPortToRemotePortMapping {
		var transportProtocol portal_generated_api.TransportProtocol
		if remotePort.GetTransportProtocol() == services.TransportProtocol_TCP {
			transportProtocol = portal_generated_api.TransportProtocol_TCP
		} else if remotePort.GetTransportProtocol() == services.TransportProtocol_UDP {
			transportProtocol = portal_generated_api.TransportProtocol_UDP
		} else {
			logrus.Warnf("Mapping other than TCP or UDP port is not supported right now. Will skip port '%d' because protocal is '%v'", remotePort.GetNumber(), remotePort.GetTransportProtocol())
		}
		forwardPortsArgs := portal_constructors.NewForwardPortArgs(uint32(localPort), uint32(remotePort.GetNumber()), &transportProtocol)
		if _, err := portalManager.portalClientMaybe.ForwardPort(ctx, forwardPortsArgs); err != nil {
			failedPorts[localPort] = remotePort
		} else {
			successfullyMappedPorts[localPort] = remotePort
		}
	}

	if len(failedPorts) > 0 {
		return successfullyMappedPorts, failedPorts, stacktrace.NewError("Some ports failed to be mapped")
	}
	return successfullyMappedPorts, failedPorts, nil
}

func (portalManager *PortalManager) instantiateClientIfUnset() error {
	portalDaemonClientMaybe, err := kurtosis_context.CreatePortalDaemonClient(true)
	if err != nil {
		return stacktrace.Propagate(err, "Unable to build client to Kurtosis Portal Daemon")
	}
	portalManager.portalClientMaybe = portalDaemonClientMaybe
	return nil
}

// getPidFromPidFile returns the PID of the current Portal process form the PID file. It doesn't check
// whether the PID is actually running or not
// It returns the PID value or 0 if no PID file was found (or if it was empty). It returns an error if something went
// wrong
func getPidFromPidFile() (int, error) {
	portalPidFilePath, err := host_machine_directories.GetPortalPidFilePath()
	if err != nil {
		return 0, stacktrace.Propagate(err, "Unable to get file path to PID file")
	}
	if _, err = CreateFileIfNecessary(portalPidFilePath); err != nil {
		return 0, stacktrace.Propagate(err, "Unable to get or create Kurtosis PID file")
	}
	pidFileContent, err := os.ReadFile(portalPidFilePath)
	if err != nil {
		return 0, stacktrace.Propagate(err, "Unable to read Kurtosis Portal PID file content")
	}
	if len(pidFileContent) == 0 {
		// PID file is empty, portal is not running
		return 0, nil
	}
	pidFileRawContent := string(pidFileContent)
	pid, err := strconv.ParseInt(pidFileRawContent, pidNumberBase, pidNumberBitSize)
	if err != nil {
		return 0, stacktrace.Propagate(err, "Unable to parse Kurtosis Portal PID file content. Was expecting a single PID number, got: '%s'", pidFileRawContent)
	}
	return int(pid), nil
}

// getRunningProcessFromPid returns the os.Process object corresponding to the Portal process, or nil if the process is not running
func getRunningProcessFromPid(pid int) (*os.Process, error) {
	process, err := os.FindProcess(pid)
	if err != nil {
		// this should never happen on Unix system, see FindProcess docs
		return nil, stacktrace.Propagate(err, "Unexpected error getting process attached to PID '%d'", pid)
	}
	if err = process.Signal(syscall.Signal(portalProcessPingSignal)); err != nil {
		// PID file exists but process seem to be dead
		return nil, nil
	}
	return process, nil
}
