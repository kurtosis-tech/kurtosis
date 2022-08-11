package logs_database_server_config

type LogsDatabaseConfigProvider interface {
	//Returns the config content
	GetConfigContent() (string, error)
}
