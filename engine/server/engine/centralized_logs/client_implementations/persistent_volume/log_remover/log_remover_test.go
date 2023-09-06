package log_remover

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/logs_clock"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/volume_consts"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/volume_filesystem"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
	"testing/fstest"
)

const (
	logsStorageDirpathForTests = "var/log/kurtosis/"

	testEnclaveUuid      = "test-enclave"
	testUserService1Uuid = "test-user-service-1"

	defaultDay = 0
)

func TestLogRemover_Run(t *testing.T) {
	week49filepath := getWeekFilepathStr(2022, 49)
	week50filepath := getWeekFilepathStr(2022, 50)
	week51filepath := getWeekFilepathStr(2022, 51)
	week52filepath := getWeekFilepathStr(2022, 52)
	week1filepath := getWeekFilepathStr(2023, 1)
	week2filepath := getWeekFilepathStr(2023, 2)

	mapFs := &fstest.MapFS{
		week49filepath: {
			Data: []byte{},
		},
		week50filepath: {
			Data: []byte{},
		},
		week51filepath: {
			Data: []byte{},
		},
		week52filepath: {
			Data: []byte{},
		},
		week1filepath: {
			Data: []byte{},
		},
		week2filepath: {
			Data: []byte{},
		},
	}

	mockFs := volume_filesystem.NewMockedVolumeFilesystem(mapFs)
	currentWeek := 2

	mockTime := logs_clock.NewMockLogsClock(2023, currentWeek, defaultDay)
	logRemover := NewLogRemover(mockFs, mockTime)

	// log remover should remove week 49 logs
	logRemover.Run()

	_, err := mockFs.Stat(week49filepath)
	require.Error(t, err) // check the file doesn't exist
}

func getWeekFilepathStr(year, week int) string {
	return fmt.Sprintf(volume_consts.PerWeekFilePathFmtStr, logsStorageDirpathForTests, strconv.Itoa(year), strconv.Itoa(week), testEnclaveUuid, testUserService1Uuid, volume_consts.Filetype)
}
