package logs_collector_functions

//These are the list of container's labels that the Docker's logging driver (for instance the Fluetd logging driver)
//will add into the logs stream when it sends them to the destination (for instance Loki, the logs database)
type LogsCollectorLabels []string
