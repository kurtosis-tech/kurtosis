import { PortSpec } from "./port_spec";

// The UUID of an artifact containing files that should be mounted into a service container
export type FilesArtifactUUID = string;

// ====================================================================================================
//                                    Config Object
// ====================================================================================================
// TODO defensive copy when we're giving back complex objects?????
// Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
export class ContainerConfig {
    constructor(
        public readonly image: string,
        public readonly usedPorts: Map<string, PortSpec>,
        public readonly publicPorts: Map<string, PortSpec>, //TODO this is a huge hack to temporarily enable static ports for NEAR until we have a more productized solution
        public readonly filesArtifactMountpoints: Map<FilesArtifactUUID, string>,
        public readonly entrypointOverrideArgs: string[],
        public readonly cmdOverrideArgs: string[],
        public readonly environmentVariableOverrides: Map<string,string>,
    ) {}

    // No need for getters because all the fields are 'readonly'
}

// ====================================================================================================
//                                        Builder
// ====================================================================================================
// TODO Defensive copies on all these With... functions???
// Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
export class ContainerConfigBuilder {
    private readonly image: string;
    private usedPorts: Map<string, PortSpec>;
    private publicPorts: Map<string, PortSpec>; //TODO this is a huge hack to temporarily enable static ports for NEAR until we have a more productized solution
    private filesArtifactMountpoints: Map<FilesArtifactUUID, string>;
    private entrypointOverrideArgs: string[];
	private cmdOverrideArgs: string[];
	private environmentVariableOverrides: Map<string,string>;

    constructor (image: string) {
        this.image = image;
        this.usedPorts = new Map();
        this.filesArtifactMountpoints = new Map();
        this.entrypointOverrideArgs = [];
        this.cmdOverrideArgs = [];
        this.environmentVariableOverrides = new Map();
        this.publicPorts = new Map(); //TODO this is a huge hack to temporarily enable static ports for NEAR until we have a more productized solution
    }

    public withUsedPorts(usedPorts: Map<string, PortSpec>): ContainerConfigBuilder {
        this.usedPorts = usedPorts;
        return this;
    }

    public withFiles(filesArtifactMountpoints: Map<FilesArtifactUUID, string>): ContainerConfigBuilder {
        this.filesArtifactMountpoints = filesArtifactMountpoints;
        return this;
    }

    public withEntrypointOverride(args: string[]): ContainerConfigBuilder {
        this.entrypointOverrideArgs = args;
        return this;
	}
	
	public withCmdOverride(args: string[]): ContainerConfigBuilder {
        this.cmdOverrideArgs = args;
        return this;
	}
	
	public withEnvironmentVariableOverrides(envVars: Map<string, string>): ContainerConfigBuilder {
        this.environmentVariableOverrides = envVars;
        return this;
	}

    //TODO this is a huge hack to temporarily enable static ports for NEAR until we have a more productized solution
    public withPublicPorts(publicPorts: Map<string, PortSpec>): ContainerConfigBuilder {
        this.publicPorts = publicPorts;
        return this;
    }

    public build(): ContainerConfig {
        return new ContainerConfig(
            this.image,
            this.usedPorts,
            this.publicPorts,
            this.filesArtifactMountpoints,
            this.entrypointOverrideArgs,
            this.cmdOverrideArgs,
            this.environmentVariableOverrides,
        );
    }
}
