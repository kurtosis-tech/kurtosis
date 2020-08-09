package parallelism

import (
	"fmt"
	"github.com/docker/distribution/uuid"
	"github.com/docker/docker/client"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"sync"
)

/*
Executor that will coordinate the execution of multiple tests in parallel
 */
type TestExecutorParallelizer struct {
	// The ID of the test suite execution to which all the individual test executions belong
	executionId                 uuid.UUID

	// Docker client which will be used for manipulating the Docker environment when running a test
	dockerClient                *client.Client

	// Name of the Docker image of the test controller that will be used to orchestrate each test
	testControllerImageName     string

	// A string, meaningful only to the controller image, which tells it what log level it should run at
	testControllerLogLevel      string

	// A ke-value map of custom Docker environment variables that will be passed as-is to the controller container during startup
	customTestControllerEnvVars map[string]string

	// The number of tests to run in parallel
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
		logs should be written to the test-specific logs during parallel test execution to avoid test logs getting jumbled)
	allTestParams: A mapping of test_name -> parameters for running the test

Returns:
	True if all tests passed, false otherwise
 */
func (executor TestExecutorParallelizer) RunInParallel(allTestParams map[string]ParallelTestParams) bool {
	// These need to be buffered else sending to the channel will be blocking
	testParamsChan := make(chan ParallelTestParams, len(allTestParams))

	logrus.Info("Loading test params into work queue...")
	for _, testParams := range allTestParams {
		testParamsChan <- testParams
	}
	close(testParamsChan) // We close the channel so that when all params are consumed, the worker threads won't block on waiting for more params
	logrus.Info("All test params loaded into work queue")

	outputManager := newParallelTestOutputManager()

	logrus.Infof("Launching %v tests with parallelism %v...", len(allTestParams), executor.parallelism)
	executor.disableSystemLogAndRunTestThreads(outputManager, testParamsChan)
	logrus.Info("All tests exited")

	outputManager.printSummary()
	return outputManager.getAllTestsPassed()
}

func (executor TestExecutorParallelizer) disableSystemLogAndRunTestThreads(
		outputManager *ParallelTestOutputManager,
		testParamsChan chan ParallelTestParams) {
	/*
    Because each test needs to have its logs written to an independent file to avoid getting logs all mixed up, we need to make
    sure that all code below this point uses the per-test logger rather than the systemwide logger. However, it's very difficult for
    a coder to remember to use 'log.Info' when they're used to doing 'logrus.Info'. To enforce this, we capture any systemwide logger usages
	during this function so we can show them later.
	*/
	outputManager.startInterceptingStdLogger()
	defer outputManager.stopInterceptingStdLogger()

	var waitGroup sync.WaitGroup
	for i := uint(0); i < executor.parallelism; i++ {
		waitGroup.Add(1)
		go executor.runTestWorkerGoroutine(outputManager, &waitGroup, testParamsChan)
	}
	waitGroup.Wait()
}

/*
A function, designed to be run inside a worker thread, that will pull test params from the given test params channel, execute the test, and
push the result to the test results channel
 */
func (executor TestExecutorParallelizer) runTestWorkerGoroutine(
			outputManager *ParallelTestOutputManager,
			waitGroup *sync.WaitGroup,
			testParamsChan chan ParallelTestParams) {
	// IMPORTANT: make sure that we mark a thread as done!
	defer waitGroup.Done()

	for testParams := range testParamsChan {
		testName := testParams.TestName

		tempFilename := fmt.Sprintf("%v-%v", executor.executionId, testName)
		writingTempFp, err := ioutil.TempFile("", tempFilename)
		if err != nil {
			emptyOutputReader := &strings.Reader{}
			executionErr := stacktrace.Propagate(err, "An error occurred creating temporary file to contain logs of test %v", testName)
			outputManager.logTestOutput(testName, executionErr, false, emptyOutputReader)
			continue
		}
		defer os.Remove(writingTempFp.Name())

		// Create a separate logger just for this test that writes to a tempfile
		log := logrus.New()
		log.SetLevel(logrus.GetLevel())
		log.SetOutput(writingTempFp)
		log.SetFormatter(logrus.StandardLogger().Formatter)

		testExecutor := newTestExecutor(
			log,
			executor.executionId,
			executor.dockerClient,
			testParams.SubnetMask,
			executor.testControllerImageName,
			executor.testControllerLogLevel,
			executor.customTestControllerEnvVars,
			testName,
			testParams.Test)

		passed, executionErr := testExecutor.runTest()
		writingTempFp.Close() // Close to flush out anything remaining in the buffer

		// Create a new FP to read the logfile from the start
		var testOutputReader io.Reader
		readingTempFp, err := os.Open(writingTempFp.Name())
		if err != nil {
			errorMsg := fmt.Sprintf("An error occurred opening the test's logfile for reading; logs for this test are unavailable:\n%s", err)
			testOutputReader = strings.NewReader(errorMsg)
		} else {
			defer readingTempFp.Close()
			testOutputReader = readingTempFp
		}
		outputManager.logTestOutput(testName, executionErr, passed, testOutputReader)
	}
}
