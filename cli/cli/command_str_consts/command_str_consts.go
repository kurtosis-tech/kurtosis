package command_str_consts

import (
	"os"
	"path"
)

// We put all the command strings here so that when we need to give users remediation instructions, we can give them the
//  commands they need to run
var KurtosisCmdStr = path.Base(os.Args[0])
const (
	EnclaveCmdStr = "enclave"
		EnclaveInspectCmdStr = "inspect"
		EnclaveLsCmdStr = "ls"
		EnclaveNewCmdStr = "new"
	EngineCmdStr           = "engine"
		EngineStartCmdStr  = "ls"
		EngineStatusCmdStr = "status"
		EngineStopCmdStr   = "stop"
	ModuleCmdStr = "module"
		ModuleExecCmdStr = "exec"
	ReplCmdStr = "repl"
		ReplInstallCmdStr = "install"
		ReplNewCmdStr = "new"
	SandboxCmdStr = "sandbox"
	ServiceCmdStr = "service"
		ServiceLogsCmdStr = "logs"
	TestCmdStr = "test"
	VersionCmdStr = "version"
)

