export class ModuleContainerArgs {
    constructor(
        // The ID of the enclave that the module will run inside
        public readonly enclaveId: string,

        // The port number that the module should listen on
        public readonly listenPortNum: number,

        // IP:port of the Kurtosis API container
        public readonly apiContainerSocket: string,

        // Arbitrary serialized data that the module can consume at startup to modify its behaviour
        // Analogous to the "constructor"
        public readonly serializedCustomParams: string,

        // The location on the module container where the enclave data directory has been mounted during launch
        public readonly enclaveDataDirMountpoint: string,
    ) {}
}