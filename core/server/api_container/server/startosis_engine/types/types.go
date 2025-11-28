package types

import (
	"fmt"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"
)

type ScheduledInstructionUuid string

type ScheduledInstructionUuidGenerator interface {
	GenerateUUIDString() (string, error)
}

type ScheduledInstructionUuidGeneratorImpl struct {
}

func NewScheduledInstructionUuidGenerator() ScheduledInstructionUuidGenerator {
	return &ScheduledInstructionUuidGeneratorImpl{}
}

func (g *ScheduledInstructionUuidGeneratorImpl) GenerateUUIDString() (string, error) {
	return uuid_generator.GenerateUUIDString()
}

type ScheduledInstructionUuidGeneratorForDependencyGraphTests struct {
	uuidCount int
}

func NewScheduledInstructionUuidGeneratorForTests() ScheduledInstructionUuidGenerator {
	return &ScheduledInstructionUuidGeneratorForDependencyGraphTests{uuidCount: 0}
}

func (g *ScheduledInstructionUuidGeneratorForDependencyGraphTests) GenerateUUIDString() (string, error) {
	g.uuidCount++
	return fmt.Sprintf("%d", g.uuidCount), nil
}
