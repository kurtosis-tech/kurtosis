package logs_database

type LogsDatabase interface {
	//Returns the config content
	GetConfigContent() (string, error)
}
