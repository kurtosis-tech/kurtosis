package enclave_data_directory

type FileNameAndUuid struct {
	uuid FilesArtifactUUID
	name string
}

func (nameAndUuid FileNameAndUuid) GetName() string {
	return nameAndUuid.name
}

func (nameAndUuid FileNameAndUuid) GetUuid() FilesArtifactUUID {
	return nameAndUuid.uuid
}
