package command_str_consts

import (
	"os"
	"path"
)

// We put all the command strings here so that when we need to give users remediation instructions, we can give them the
//  commands they need to run
var KurtosisCmdStr = path.Base(os.Args[0])
const (
	CleanCmdStr = "clean"
	ClusterCmdStr = "cluster"
		ClusterSetCmdStr = "set"
		ClusterGetCmdStr = "get"
		ClusterLsCmdStr = "ls"
	EnclaveCmdStr = "enclave"
		EnclaveInspectCmdStr = "inspect"
		EnclaveLsCmdStr   = "ls"
		EnclaveAddCmdStr  = "add"
		EnclaveStopCmdStr = "stop"
		EnclaveRmCmdStr = "rm"
		EnclaveDumpCmdStr = "dump"
	EngineCmdStr              = "engine"
		EngineStartCmdStr   = "start"
		EngineStatusCmdStr  = "status"
		EngineStopCmdStr    = "stop"
		EngineRestartCmdStr = "restart"
	FilesCmdStr           = "files"
		FilesUploadCmdStr = "upload"
		FilesStoreWebCmdStr = "storeweb"
		FilesStoreServiceCmdStr = "storeservice"
	ModuleCmdStr = "module"
		ModuleExecCmdStr = "exec"
	ServiceCmdStr = "service"
		ServiceAddCmdStr = "add"
		ServiceLogsCmdStr = "logs"
		ServicePauseCmdStr = "pause"
		ServiceRmCmdStr = "rm"
		ServiceUnpauseCmdStr = "unpause"
		ServiceShellCmdStr   = "shell"
	ConfigCmdStr             = "config"
		InitCmdStr = "init"
		PathCmdStr = "path"
	VersionCmdStr = "version"
	GatewayCmdStr = "gateway"
)

