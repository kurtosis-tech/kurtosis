package exec_result

type ExecResult struct {
	exitCode int32
	output   string
}

func NewExecResult(exitCode int32, output string) *ExecResult {
	return &ExecResult{exitCode: exitCode, output: output}
}

func (execResult *ExecResult) GetExitCode() int32 {
	return execResult.exitCode
}

func (execResult *ExecResult) GetOutput() string {
	return execResult.output
}
