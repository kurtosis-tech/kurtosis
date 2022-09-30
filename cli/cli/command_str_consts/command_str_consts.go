package command_str_consts

import (
	"os"
	"path"
)

// We put all the command strings here so that when we need to give users remediation instructions, we can give them the
// commands they need to run
var KurtosisCmdStr = path.Base(os.Args[0])

const (
	// clean
	CleanCmdStr = "clean"

	// cluster
	ClusterCmdStr    = "cluster"
	ClusterSetCmdStr = "set"
	ClusterGetCmdStr = "get"
	ClusterLsCmdStr  = "ls"

	// discord
	DiscordCmdStr = "discord"

	// enclave
	EnclaveCmdStr        = "enclave"
	EnclaveInspectCmdStr = "inspect"
	EnclaveLsCmdStr      = "ls"
	EnclaveAddCmdStr     = "add"
	EnclaveStopCmdStr    = "stop"
	EnclaveRmCmdStr      = "rm"
	EnclaveDumpCmdStr    = "dump"

	// engine
	EngineCmdStr        = "engine"
	EngineStartCmdStr   = "start"
	EngineStatusCmdStr  = "status"
	EngineStopCmdStr    = "stop"
	EngineRestartCmdStr = "restart"

	// files
	FilesCmdStr             = "files"
	FilesUploadCmdStr       = "upload"
	FilesStoreWebCmdStr     = "storeweb"
	FilesStoreServiceCmdStr = "storeservice"
	FilesRenderTemplate     = "rendertemplate"

	// module
	ModuleCmdStr     = "module"
	ModuleExecCmdStr = "exec"

	// service
	ServiceCmdStr        = "service"
	ServiceAddCmdStr     = "add"
	ServiceLogsCmdStr    = "logs"
	ServicePauseCmdStr   = "pause"
	ServiceRmCmdStr      = "rm"
	ServiceUnpauseCmdStr = "unpause"
	ServiceShellCmdStr   = "shell"

	// startosis
	StartosisCmdStr     = "startosis"
	StartosisExecCmdStr = "exec"

	// config
	ConfigCmdStr = "config"
	InitCmdStr   = "init"
	PathCmdStr   = "path"

	// version
	VersionCmdStr = "version"

	// gateway
	GatewayCmdStr = "gateway"
)
