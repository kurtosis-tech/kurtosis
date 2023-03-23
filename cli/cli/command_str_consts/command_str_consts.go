package command_str_consts

import (
	"os"
	"path"
)

// We put all the command strings here so that when we need to give users remediation instructions, we can give them the
//
//	commands they need to run
var KurtosisCmdStr = path.Base(os.Args[0])

const (
	Analytics               = "analytics"
	CleanCmdStr             = "clean"
	ClusterCmdStr           = "cluster"
	ClusterSetCmdStr        = "set"
	ClusterGetCmdStr        = "get"
	ClusterLsCmdStr         = "ls"
	ContextCmdStr           = "context"
	ContextAddCmdStr        = "add"
	ContextLsCmdStr         = "ls"
	DiscordCmdStr           = "discord"
	DocsCmdStr              = "docs"
	EnclaveCmdStr           = "enclave"
	EnclaveInspectCmdStr    = "inspect"
	EnclaveLsCmdStr         = "ls"
	EnclaveAddCmdStr        = "add"
	EnclaveStopCmdStr       = "stop"
	EnclaveRmCmdStr         = "rm"
	EnclaveDumpCmdStr       = "dump"
	EngineCmdStr            = "engine"
	EngineLogsCmdStr        = "logs"
	EngineStartCmdStr       = "start"
	EngineStatusCmdStr      = "status"
	EngineStopCmdStr        = "stop"
	EngineRestartCmdStr     = "restart"
	FeedbackCmdStr          = "feedback"
	FilesCmdStr             = "files"
	FilesUploadCmdStr       = "upload"
	FilesDownloadCmdStr     = "download"
	FilesStoreWebCmdStr     = "storeweb"
	FilesStoreServiceCmdStr = "storeservice"
	FilesRenderTemplate     = "rendertemplate"
	ServiceCmdStr           = "service"
	ServiceAddCmdStr        = "add"
	ServiceLogsCmdStr       = "logs"
	ServiceRmCmdStr         = "rm"
	ServiceShellCmdStr      = "shell"
	StarlarkRunCmdStr       = "run"
	TwitterCmdStr           = "twitter"
	ConfigCmdStr            = "config"
	InitCmdStr              = "init"
	PathCmdStr              = "path"
	VersionCmdStr           = "version"
	GatewayCmdStr           = "gateway"
)
