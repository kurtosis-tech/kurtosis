package initializer

// TODO TODO TODO Combine this entire file into test_suite_runner????

import (
	"context"
	"github.com/docker/distribution/uuid"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
	"os"
	"sync"
)

// ================= Test params & result ============================
type ParallelTestParams struct {
	testName string
	logFp *os.File
	subnetMask string
	executionInstanceId uuid.UUID
}
type TestOutput struct {
	Name		 string
	ExecutionErr error    // Indicates whether an error occurred during the execution of the test that prevented it from running
	Passed       bool     // Indicates whether the test passed or failed (undefined if the test had a setup error)
	LogFp        *os.File // Where the logs for this test got written to
}

// ================= Parallel executor ============================
type ParallelTestExecutor struct {
	executionId uuid.UUID
	dockerClient *client.Client
	testControllerImageName string
	testControllerLogLevel string
	testServiceImageName string
	parallelism int
}

func NewParallelTestExecutor(
			executionId uuid.UUID,
			dockerClient *client.Client,
			testControllerImageName string,
			testControllerLogLevel string,
			testServiceImageName string,
			parallelism int) *ParallelTestExecutor {
	return &ParallelTestExecutor{
		executionId: executionId,
		dockerClient: dockerClient,
		testControllerImageName: testControllerImageName,
		testControllerLogLevel: testControllerLogLevel,
		testServiceImageName: testServiceImageName,
		parallelism: parallelism,
	}
}



func (executor ParallelTestExecutor) RunTestsInParallel(tests map[string]ParallelTestParams) map[string]TestOutput {

	testParamsChan := make(chan ParallelTestParams)
	for _, testParams := range tests {
		testParamsChan <- testParams
	}
	close(testParamsChan)

	logrus.Info("Launching %v tests with %v parallelism...", len(tests), executor.parallelism)
	var waitGroup sync.WaitGroup
	testOutputChan := make(chan TestOutput)
	for i := 0; i < executor.parallelism; i++ {
		waitGroup.Add(1)
		go executor.runTestWorker(&waitGroup, testParamsChan, testOutputChan)
	}
	logrus.Info("All worker threads started successfully")

	// TODO add a timeout which the total execution must not exceed
	logrus.Info("Waiting for tests to finish...")
	waitGroup.Wait()
	logrus.Info("All worker threads exited")

	// Collect all results
	result := make(map[string]TestOutput)
	for i := 0; i < len(tests); i++ {
		output := <-testOutputChan
		result[output.Name] = output
	}
	return result
}

/*
A function, designed to be run inside a goroutine, that will pull from the given test params channel, execute a test, and
push the result to the test results channel
 */
func (executor ParallelTestExecutor) runTestWorker(
			waitGroup *sync.WaitGroup,
			testParamsChan chan ParallelTestParams,
			testOutputChan chan TestOutput) {
	// IMPORTANT: make sure that we mark a thread as done!
	defer waitGroup.Done()

	for testParams := range testParamsChan {
		// Create a separate logger just for this test that writes to the given file
		log := logrus.New()
		log.Level = logrus.GetLevel()
		log.Out = testParams.logFp
		loggedExecutor := NewLoggedTestExecutor(log)

		// TODO create a new context for the test itself probably, so we can cancel it if it's running too long!
		testContext := context.Background()

		passed, executionErr := loggedExecutor.RunTest(
			executor.executionId,
			testContext,
			executor.dockerClient,
			testParams.subnetMask,
			executor.testServiceImageName,
			executor.testControllerLogLevel,
			executor.testServiceImageName,
			testParams.testName)

		result := TestOutput{
			Name:         testParams.testName,
			ExecutionErr: executionErr,
			Passed:       passed,
			LogFp:        testParams.logFp,
		}
		testOutputChan <- result
	}
}
