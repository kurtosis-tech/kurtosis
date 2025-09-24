package dependency_graph

import (
	"testing"
	"time"
)

func TestComputeParallelExecutionTime(t *testing.T) {
	tests := []struct {
		name                     string
		dependencyGraph          map[ScheduledInstructionUuid][]ScheduledInstructionUuid
		instructionNumToDuration map[int]time.Duration
		expectedParallelTime     time.Duration
	}{
		{
			name: "No dependencies - all instructions run in parallel",
			dependencyGraph: map[ScheduledInstructionUuid][]ScheduledInstructionUuid{
				"1": {},
				"2": {},
				"3": {},
			},
			instructionNumToDuration: map[int]time.Duration{
				1: 10 * time.Second,
				2: 15 * time.Second,
				3: 20 * time.Second,
			},
			expectedParallelTime: 20 * time.Second, // Longest instruction determines parallel time
		},
		{
			name: "Linear dependency chain - no parallel execution possible",
			dependencyGraph: map[ScheduledInstructionUuid][]ScheduledInstructionUuid{
				"1": {},
				"2": {"1"},
				"3": {"2"},
			},
			instructionNumToDuration: map[int]time.Duration{
				1: 10 * time.Second,
				2: 15 * time.Second,
				3: 20 * time.Second,
			},
			expectedParallelTime: 45 * time.Second, // Sequential execution: 10 + 15 + 20
		},
		{
			name: "Partial parallel execution",
			dependencyGraph: map[ScheduledInstructionUuid][]ScheduledInstructionUuid{
				"1": {},
				"2": {"1"},
				"3": {"1"},
				"4": {"2", "3"},
			},
			instructionNumToDuration: map[int]time.Duration{
				1: 10 * time.Second,
				2: 15 * time.Second,
				3: 20 * time.Second,
				4: 5 * time.Second,
			},
			expectedParallelTime: 35 * time.Second, // 10 + max(15, 20) + 5
		},
		{
			name: "Complex dependency graph with multiple paths",
			dependencyGraph: map[ScheduledInstructionUuid][]ScheduledInstructionUuid{
				"1": {},
				"2": {},
				"3": {"1"},
				"4": {"2"},
				"5": {"3", "4"},
				"6": {"5"},
			},
			instructionNumToDuration: map[int]time.Duration{
				1: 5 * time.Second,
				2: 8 * time.Second,
				3: 12 * time.Second,
				4: 10 * time.Second,
				5: 7 * time.Second,
				6: 15 * time.Second,
			},
			expectedParallelTime: 40 * time.Second,
		},
		{
			name: "Single instruction",
			dependencyGraph: map[ScheduledInstructionUuid][]ScheduledInstructionUuid{
				"1": {},
			},
			instructionNumToDuration: map[int]time.Duration{
				1: 10 * time.Second,
			},
			expectedParallelTime: 10 * time.Second,
		},
		{
			name:                     "Empty dependency graph",
			dependencyGraph:          map[ScheduledInstructionUuid][]ScheduledInstructionUuid{},
			instructionNumToDuration: map[int]time.Duration{},
			expectedParallelTime:     0 * time.Second,
		},
		{
			name: "Instructions with same duration",
			dependencyGraph: map[ScheduledInstructionUuid][]ScheduledInstructionUuid{
				"1": {},
				"2": {"1"},
				"3": {"1"},
			},
			instructionNumToDuration: map[int]time.Duration{
				1: 10 * time.Second,
				2: 10 * time.Second,
				3: 10 * time.Second,
			},
			expectedParallelTime: 20 * time.Second, // 10 + 10 (parallel)
		},
		// {
		// 	name: "Eth network",
		// 	dependencyGraph: map[ScheduledInstructionUuid][]ScheduledInstructionUuid{
		// 		"1": {},         // generate genesis
		// 		"2": {"1"},      // start el bootnode
		// 		"3": {"2"},      // start second el node
		// 		"4": {"2"},      // start third el node
		// 		"5": {"4", "3"}, // start cl bootnode
		// 		"6": {"5"},      // start second cl node
		// 		"7": {"5"},      // start third cl node
		// 		"8": {"1"},      // a random run sh
		// 	},
		// 	instructionNumToDuration: map[int]time.Duration{
		// 		1: 5 * time.Second,
		// 		2: 3 * time.Second,
		// 		3: 3 * time.Second,
		// 		4: 3 * time.Second,
		// 		5: 20 * time.Second,
		// 		6: 20 * time.Second,
		// 		7: 20 * time.Second,
		// 		8: 2 * time.Second,
		// 	},
		// 	expectedParallelTime: 20 * time.Second, // 5 + 3 + 3 + 20 + 20
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ComputeParallelExecutionTime(tt.dependencyGraph, tt.instructionNumToDuration)

			if result != tt.expectedParallelTime {
				OutputDependencyGraphVisual(tt.dependencyGraph)
				t.Errorf("%v computeParallelExecutionTime() = %v, want %v", tt.name, result, tt.expectedParallelTime)
			}
		})
	}
}

func TestExtractInstructionNumber(t *testing.T) {
	tests := []struct {
		name     string
		uuid     ScheduledInstructionUuid
		expected int
		ok       bool
	}{
		{
			name:     "Valid instruction UUID",
			uuid:     "1",
			expected: 1,
			ok:       true,
		},
		{
			name:     "UUID with multiple digits",
			uuid:     "123",
			expected: 123,
			ok:       true,
		},
		{
			name:     "Empty UUID",
			uuid:     "",
			expected: 0,
			ok:       false,
		},
		{
			name:     "UUID without numbers",
			uuid:     "abc",
			expected: 0,
			ok:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := extractInstructionNumber(tt.uuid)

			if result != tt.expected || ok != tt.ok {
				t.Errorf("extractInstructionNumber(%s) = (%d, %t), want (%d, %t)",
					tt.uuid, result, ok, tt.expected, tt.ok)
			}
		})
	}
}

// Benchmark test for performance
func BenchmarkComputeParallelExecutionTime(b *testing.B) {
	dependencyGraph := map[ScheduledInstructionUuid][]ScheduledInstructionUuid{
		"1":  {},
		"2":  {},
		"3":  {"1"},
		"4":  {"2"},
		"5":  {"3", "4"},
		"6":  {"5"},
		"7":  {"1"},
		"8":  {"2"},
		"9":  {"7", "8"},
		"10": {"9"},
	}

	instructionNumToDuration := map[int]time.Duration{
		1:  5 * time.Second,
		2:  8 * time.Second,
		3:  12 * time.Second,
		4:  10 * time.Second,
		5:  7 * time.Second,
		6:  15 * time.Second,
		7:  6 * time.Second,
		8:  9 * time.Second,
		9:  11 * time.Second,
		10: 13 * time.Second,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ComputeParallelExecutionTime(dependencyGraph, instructionNumToDuration)
	}
}
