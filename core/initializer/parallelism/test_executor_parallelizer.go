package parallelism

import (
	"github.com/docker/distribution/uuid"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/kurtosis/commons/testsuite"
	"github.com/sirupsen/logrus"
	"os"
	"sync"
)

// ================= Test params & result ============================
type ParallelTestParams struct {
	TestName            string
	Test 				testsuite.Test
	LogFp               *os.File
	SubnetMask          string
	ExecutionInstanceId uuid.UUID
}

func NewParallelTestParams(testName string, test testsuite.Test, logFp *os.File, subnetMask string, executionInstanceId uuid.UUID) *ParallelTestParams {
	return &ParallelTestParams{TestName: testName, Test: test, LogFp: logFp, SubnetMask: subnetMask, ExecutionInstanceId: executionInstanceId}
}

type ParallelTestOutput struct {
	TestName     string
	ExecutionErr error    // Indicates whether an error occurred during the execution of the test that prevented it from running
	TestPassed   bool     // Indicates whether the test passed or failed (undefined if the test had a setup error)
	LogFp        *os.File // Where the logs for this test got written to
}

// ================= Parallel executor ============================
type TestExecutorParallelizer struct {
	executionId                 uuid.UUID
	dockerClient                *client.Client
	testControllerImageName     string
	testControllerLogLevel      string
	customTestControllerEnvVars map[string]string
	parallelism                 uint
}

/*
Creates a new TestExecutorParallelizer which will run tests in parallel using the given parameters.

Args:
	executionId: The UUID uniquely identifying this execution of the tests
	dockerClient: The handle to manipulating the Docker environment
	testControllerImageName: The name of the Docker image that will be used to run the test controller
	testControllerLogLevel: A string, meaningful to the test controller, that represents the user's desired log level
	customTestControllerEnvVars: A custom user-defined map from <env variable name> -> <env variable value> that will be
		passed via Docker environment variables to the test controller
	parallelism: The number of tests to run concurrently
 */
func NewTestExecutorParallelizer(
			executionId uuid.UUID,
			dockerClient *client.Client,
			testControllerImageName string,
			testControllerLogLevel string,
			customTestControllerEnvVars map[string]string,
			parallelism uint) *TestExecutorParallelizer {
	return &TestExecutorParallelizer{
		executionId:                 executionId,
		dockerClient:                dockerClient,
		testControllerImageName:     testControllerImageName,
		testControllerLogLevel:      testControllerLogLevel,
		customTestControllerEnvVars: customTestControllerEnvVars,
		parallelism:                 parallelism,
	}
}

/*
Runs the given tests in parallel

Args:
	interceptor: A capturer for logs that are erroneously written to the system-level log during parallel test execution (since all
		logs should be written to the test-specific logs during parallel test execution)
	tests:
 */
func (executor TestExecutorParallelizer) RunInParallel(interceptor *ErroneousSystemLogCaptureWriter, allTestParams map[string]ParallelTestParams) map[string]ParallelTestOutput {
	// These need to be buffered else sending to the channel will be blocking
	testParamsChan := make(chan ParallelTestParams, len(allTestParams))
	testOutputChan := make(chan ParallelTestOutput, len(allTestParams))

	logrus.Info("Loading test params into work queue...")
	for _, testParams := range allTestParams {
		testParamsChan <- testParams
	}
	close(testParamsChan) // We close the channel so that when all params are consumed, the worker threads won't block on waiting for more params
	logrus.Info("All test params loaded into work queue")

	logrus.Infof("Launching %v tests with parallelism %v...", len(allTestParams), executor.parallelism)
	executor.disableSystemLogAndRunTestThreads(interceptor, testParamsChan, testOutputChan)
	logrus.Info("All tests exited")

	// Collect all results
	result := make(map[string]ParallelTestOutput)
	for i := 0; i < len(allTestParams); i++ {
		output := <-testOutputChan
		result[output.TestName] = output
	}
	return result
}

func (executor TestExecutorParallelizer) disableSystemLogAndRunTestThreads(interceptor *ErroneousSystemLogCaptureWriter, testParamsChan chan ParallelTestParams, testOutputChan chan ParallelTestOutput) {
	/*
    Because each test needs to have its logs written to an independent file to avoid getting logs all mixed up, we need to make
    sure that all code below this point uses the per-test logger rather than the systemwide logger. However, it's very difficult for
    a coder to remember to use 'log.Info' when they're used to doing 'logrus.Info'. To enforce this, we capture any systemwide logger usages
	during this function so we can show them later.
	*/
	currentSystemOut := logrus.StandardLogger().Out
	logrus.SetOutput(interceptor)
	defer logrus.SetOutput(currentSystemOut)

	var waitGroup sync.WaitGroup
	for i := uint(0); i < executor.parallelism; i++ {
		waitGroup.Add(1)
		go executor.runTestWorkerGoroutine(&waitGroup, testParamsChan, testOutputChan)
	}
	waitGroup.Wait()
}

/*
A function, designed to be run inside a worker thread, that will pull test params from the given test params channel, execute the test, and
push the result to the test results channel
 */
func (executor TestExecutorParallelizer) runTestWorkerGoroutine(
			waitGroup *sync.WaitGroup,
			testParamsChan chan ParallelTestParams,
			testOutputChan chan ParallelTestOutput) {
	// IMPORTANT: make sure that we mark a thread as done!
	defer waitGroup.Done()

	for testParams := range testParamsChan {
		// Create a separate logger just for this test that writes to the given file
		log := logrus.New()
		log.SetLevel(logrus.GetLevel())
		log.SetOutput(testParams.LogFp)
		log.SetFormatter(logrus.StandardLogger().Formatter)
		testExecutor := newTestExecutor(
			log,
			executor.executionId,
			executor.dockerClient,
			testParams.SubnetMask,
			executor.testControllerImageName,
			executor.testControllerLogLevel,
			executor.customTestControllerEnvVars,
			testParams.TestName,
			testParams.Test)

		passed, executionErr := testExecutor.runTest()

		testOutput := ParallelTestOutput{
			TestName:     testParams.TestName,
			ExecutionErr: executionErr,
			TestPassed:   passed,
			LogFp:        testParams.LogFp,
		}
		testOutputChan <- testOutput
	}
}
