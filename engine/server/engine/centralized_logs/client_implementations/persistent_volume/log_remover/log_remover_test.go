package log_remover

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/logs_clock"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/volume_consts"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/volume_filesystem"
	"github.com/stretchr/testify/require"
	"os"
	"strconv"
	"testing"
)

const (
	testEnclaveUuid      = "test-enclave"
	testUserService1Uuid = "test-user-service-1"

	defaultDay = 0
)

func TestLogRemover_Run(t *testing.T) {
	mockFs := volume_filesystem.NewMockedVolumeFilesystem()

	week49filepath := getWeekFilepathStr(2022, 49)
	week50filepath := getWeekFilepathStr(2022, 50)
	week51filepath := getWeekFilepathStr(2022, 51)
	week52filepath := getWeekFilepathStr(2022, 52)
	week1filepath := getWeekFilepathStr(2023, 1)
	week2filepath := getWeekFilepathStr(2023, 2)

	mockFs.Create(week49filepath)
	mockFs.Create(week50filepath)
	mockFs.Create(week51filepath)
	mockFs.Create(week52filepath)
	mockFs.Create(week1filepath)
	mockFs.Create(week2filepath)

	currentWeek := 2

	mockTime := logs_clock.NewMockLogsClock(2023, currentWeek, defaultDay)
	logRemover := NewLogRemover(mockFs, mockTime)

	// log remover should remove week 49 logs
	logRemover.Run()

	_, err := mockFs.Stat(week49filepath)
	require.Error(t, err)
	require.True(t, os.IsNotExist(err))
}

func getWeekFilepathStr(year, week int) string {
	return fmt.Sprintf(volume_consts.PerWeekFilePathFmtStr, volume_consts.LogsStorageDirpath, strconv.Itoa(year), strconv.Itoa(week), testEnclaveUuid, testUserService1Uuid, volume_consts.Filetype)
}
