package snapshots

// func TestGetMainScriptToExecuteFromSnapshotPackage(t *testing.T) {
// 	packageRootPathOnDisk := "/Users/tewodrosmitiku/craft/sandbox/tests/snapshot-test"

// 	mainScriptToExecute, err := GetMainScriptToExecuteFromSnapshotPackage(packageRootPathOnDisk)
// 	require.NoError(t, err)

// 	expectedMainScript := `
// 	 def run(plan, args):
// 		plan.add_service(name="test",config=ServiceConfig(image="test-1746656296-snapshot-img",cmd=["sleep", "1000"],env_vars={"TEST": "test"}))
// 	`

// 	require.Equal(t, expectedMainScript, mainScriptToExecute)
// }
