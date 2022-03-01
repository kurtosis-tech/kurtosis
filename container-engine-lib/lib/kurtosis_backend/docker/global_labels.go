package docker

// TODO MOVE TO FOREVER CONSTS!!!!!
var GlobalLabels = map[*DockerLabelKey]*DockerLabelValue{
	AppIDLabelKey: AppIDLabelValue,
	// TODO container engine lib version??
}