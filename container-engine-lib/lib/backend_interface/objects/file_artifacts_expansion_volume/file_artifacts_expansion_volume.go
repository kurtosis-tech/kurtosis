package file_artifacts_expansion_volume

type FileArtifactExpansionVolumeGUID string

type FileArtifactsExpansionVolume struct {
	guid FileArtifactExpansionVolumeGUID
}

func NewFileArtifactsExpansionVolume(guid FileArtifactExpansionVolumeGUID) *FileArtifactsExpansionVolume {
	return &FileArtifactsExpansionVolume{guid: guid}
}

func (expansionVolume *FileArtifactsExpansionVolume) GetGUID() FileArtifactExpansionVolumeGUID {
	return expansionVolume.guid
}
