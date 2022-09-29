/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

import { PartitionConnectionInfo } from "../../kurtosis_core_rpc_api_bindings/api_container_service_pb";
import { newPartitionConnectionInfo } from "../constructor_calls";

const UNBLOCKED_PARTITION_CONNECTION_PACKET_LOSS_VALUE: number = 0;
const BLOCKED_PARTITION_CONNECTION_PACKET_LOSS_VALUE: number = 100;
//Random packet loss is specified in the 'tc' command in percent. The smallest possible non-zero value is: 2^32 = 0.0000000232% More info: https://wiki.linuxfoundation.org/networking/netem
const SMALLEST_POSSIBLE_NON_ZERO_PACKET_LOSS_VALUE: number = 0.0000000232;
const MAX_POSSIBLE_PACKET_LOSS_VALUE = 100;

// PartitionConnection To get an instance of this type, use the UnblockedPartitionConnection, BlockedPartitionConnection or SoftPartitionConnection objects
export interface PartitionConnection {
    getPartitionConnectionInfo: () => PartitionConnectionInfo;
}

// ====================================================================================================
//                                    	 Implementations
// ====================================================================================================
// UnblockedPartitionConnection use this type of PartitionConnection when you want to establish that the connection is not partitioned
export class UnblockedPartitionConnection {
    private readonly packetLossPercentage: number = UNBLOCKED_PARTITION_CONNECTION_PACKET_LOSS_VALUE;

    constructor(){}

    public getPartitionConnectionInfo(): PartitionConnectionInfo {
        return newPartitionConnectionInfo(this.packetLossPercentage);
    }
}

// BlockedPartitionConnection use this type of PartitionConnection when you want to create a hard partition
export class BlockedPartitionConnection {
    private readonly packetLossPercentage: number = BLOCKED_PARTITION_CONNECTION_PACKET_LOSS_VALUE;

    constructor(){}

    public getPartitionConnectionInfo(): PartitionConnectionInfo {
        let partitionConnectionInfo: PartitionConnectionInfo = new PartitionConnectionInfo();
        partitionConnectionInfo.setPacketLossPercentage(this.packetLossPercentage);
        return partitionConnectionInfo;
    }
}

// SoftPartitionConnection use this type of PartitionConnection when you want to create a partition with x% percentage of packet loss
export class SoftPartitionConnection {
    private readonly packetLossPercentage: number

    constructor(packetLossPercentage: number) {
        if(!SoftPartitionConnection.isValidPacketLossValue(packetLossPercentage)){
            throw new Error(`The packet loss percentage value ${packetLossPercentage} is not allowed, 
            it should be >= ${SMALLEST_POSSIBLE_NON_ZERO_PACKET_LOSS_VALUE} 
            and <= ${MAX_POSSIBLE_PACKET_LOSS_VALUE}`);
        }
        this.packetLossPercentage = packetLossPercentage;
    }

    public getPartitionConnectionInfo(): PartitionConnectionInfo {
        let partitionConnectionInfo: PartitionConnectionInfo = new PartitionConnectionInfo();
        partitionConnectionInfo.setPacketLossPercentage(this.packetLossPercentage);
        return partitionConnectionInfo;
    }

    // ====================================================================================================
    // 									   Private helper methods
    // ====================================================================================================
    private static isValidPacketLossValue(packetLossPercentage: number): boolean {
        return packetLossPercentage >= SMALLEST_POSSIBLE_NON_ZERO_PACKET_LOSS_VALUE && packetLossPercentage <= MAX_POSSIBLE_PACKET_LOSS_VALUE;
    }
}
