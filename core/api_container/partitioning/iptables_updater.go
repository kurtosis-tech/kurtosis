/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package partitioning

import "github.com/kurtosis-tech/kurtosis/api_container/partitioning/service_id_set"

/*
Tracks the state of IP tables on given hosts, and realizes any changes on the hosts
 */
type ipTablesUpdater struct {
	ipTablesDropState map[ServiceID]service_id_set.ServiceIDSet
}

func setBlocksForService(serviceId ServiceID, blocks service_id_set.ServiceIDSet) {

}

