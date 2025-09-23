package dependency_graph

// // computeParallelTimeSaved computes the time saved by running instructions in parallel.
// // It uses dependency graph to find instructions that can be run in parallel and assumes that instructions start running as soon as they can be scheduled, aka. as soon as all their dependencies have finished runnning.
// // Then it returns the total time spent running instructions in parallel.
// func ComputeParallelExecutionTime(dependencyGraph map[ScheduledInstructionUuid][]ScheduledInstructionUuid, instructionNumToDuration map[int]time.Duration) time.Duration {
// 	// Convert ScheduledInstructionUuid to int for duration lookup
// 	// We need to create a mapping from UUID to instruction number
// 	uuidToInstructionNum := make(map[ScheduledInstructionUuid]int)
// 	for uuid := range dependencyGraph {
// 		// Extract instruction number from UUID (assuming UUID format contains instruction number)
// 		// This is a simplified approach - in practice, you might need a more robust mapping
// 		if num, ok := extractInstructionNumber(uuid); ok {
// 			uuidToInstructionNum[uuid] = num
// 		}
// 	}

// 	// Calculate the earliest start time for each instruction based on its dependencies
// 	earliestStartTimes := make(map[ScheduledInstructionUuid]time.Duration)

// 	// Initialize start times for instructions with no dependencies
// 	for instruction := range dependencyGraph {
// 		if len(dependencyGraph[instruction]) == 0 {
// 			earliestStartTimes[instruction] = 0
// 		}
// 	}

// 	// Calculate start times for all instructions using topological sort approach
// 	visited := make(map[ScheduledInstructionUuid]bool)

// 	var calculateStartTime func(instruction ScheduledInstructionUuid) time.Duration
// 	calculateStartTime = func(instruction ScheduledInstructionUuid) time.Duration {
// 		if startTime, exists := earliestStartTimes[instruction]; exists {
// 			return startTime
// 		}

// 		if visited[instruction] {
// 			// Handle circular dependencies by returning 0
// 			return 0
// 		}
// 		visited[instruction] = true

// 		// Find the latest finish time among all dependencies
// 		maxDependencyFinishTime := time.Duration(0)
// 		for _, dependency := range dependencyGraph[instruction] {
// 			dependencyStartTime := calculateStartTime(dependency)
// 			dependencyNum := uuidToInstructionNum[dependency]
// 			dependencyDuration := instructionNumToDuration[dependencyNum]
// 			dependencyFinishTime := dependencyStartTime + dependencyDuration

// 			if dependencyFinishTime > maxDependencyFinishTime {
// 				maxDependencyFinishTime = dependencyFinishTime
// 			}
// 		}

// 		earliestStartTimes[instruction] = maxDependencyFinishTime
// 		return maxDependencyFinishTime
// 	}

// 	// Calculate start times for all instructions
// 	for instruction := range dependencyGraph {
// 		calculateStartTime(instruction)
// 	}

// 	// Calculate total parallel execution time
// 	totalParallelExecutionTime := time.Duration(0)
// 	for instruction, startTime := range earliestStartTimes {
// 		instructionNum := uuidToInstructionNum[instruction]
// 		duration := instructionNumToDuration[instructionNum]
// 		finishTime := startTime + duration

// 		if finishTime > totalParallelExecutionTime {
// 			totalParallelExecutionTime = finishTime
// 		}
// 	}

// 	// Calculate total sequential execution time
// 	totalSequentialExecutionTime := time.Duration(0)
// 	for _, duration := range instructionNumToDuration {
// 		totalSequentialExecutionTime += duration
// 	}

// 	// Return the time saved (sequential - parallel)
// 	return totalParallelExecutionTime
// }

// // extractInstructionNumber extracts the instruction number from a ScheduledInstructionUuid
// // This is a helper function - the actual implementation depends on how UUIDs are formatted
// func extractInstructionNumber(uuid ScheduledInstructionUuid) (int, bool) {
// 	// For simple string format like "1", "2", "123", etc.
// 	if len(uuid) == 0 {
// 		return 0, false
// 	}

// 	// Use strconv.Atoi for string to int conversion
// 	num, err := strconv.Atoi(string(uuid))
// 	if err != nil {
// 		return 0, false
// 	}

// 	return num, true
// }
