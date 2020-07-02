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

type ParallelTestOutput struct {
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

func (executor ParallelTestExecutor) RunTestsInParallel(tests map[string]ParallelTestParams) map[string]ParallelTestOutput {
	testParamsChan := make(chan ParallelTestParams)
	for _, testParams := range tests {
		testParamsChan <- testParams
	}
	close(testParamsChan)

	testOutputChan := make(chan ParallelTestOutput)

	logrus.Infof("Launching %v tests with %v parallelism...", len(tests), executor.parallelism)
	executor.disableSystemLogAndRunTestThreads(testParamsChan, testOutputChan)
	logrus.Info("All tests exited")

	// Collect all results
	result := make(map[string]ParallelTestOutput)
	for i := 0; i < len(tests); i++ {
		output := <-testOutputChan
		result[output.Name] = output
	}
	return result
}

func (executor ParallelTestExecutor) disableSystemLogAndRunTestThreads(testParamsChan chan ParallelTestParams, testOutputChan chan ParallelTestOutput) {
	/*
    Because each test needs to have its logs written to an independent file to avoid getting logs all mixed up, we need to make
    sure that all code below this point uses the per-test logger rather than the systemwide logger. However, it's very difficult for
    a coder to remember to use 'log.Info' when they're used to doing 'logrus.Info'. To enforce this, we make the systemwide logger throw
	a panic during just this function call.
	*/
	logrus.SetOutput(PanickingLogWriter{})
	defer logrus.SetOutput(os.Stdout)

	var waitGroup sync.WaitGroup
	for i := 0; i < executor.parallelism; i++ {
		waitGroup.Add(1)
		go executor.runTestWorker(&waitGroup, testParamsChan, testOutputChan)
	}

	// TODO add a timeout which the total execution must not exceed
	waitGroup.Wait()
}

/*
A function, designed to be run inside a goroutine, that will pull from the given test params channel, execute a test, and
push the result to the test results channel
 */
func (executor ParallelTestExecutor) runTestWorker(
			waitGroup *sync.WaitGroup,
			testParamsChan chan ParallelTestParams,
			testOutputChan chan ParallelTestOutput) {
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

		result := ParallelTestOutput{
			Name:         testParams.testName,
			ExecutionErr: executionErr,
			Passed:       passed,
			LogFp:        testParams.logFp,
		}
		testOutputChan <- result
	}
}
