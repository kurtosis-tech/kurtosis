package graph

import (
	"crypto/md5"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
)

type InstructionHash string

func newInstructionHash(instruction kurtosis_instruction.KurtosisInstruction) InstructionHash {
	hash := md5.New()
	hash.Write([]byte(instruction.String()))
	return InstructionHash(hash.Sum(nil))
}
