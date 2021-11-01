// The ID of an artifact containing files that should be mounted into a service container
export type FilesArtifactID = string;

// ====================================================================================================
//                                    Config Object
// ====================================================================================================
// TODO defensive copy when we're giving back complex objects?????
// Docs available at https://docs.kurtosistech.com/kurtosis-client/lib-documentation
export class ContainerConfig {
	
    private readonly image: string;
    private readonly usedPortsSet: Set<string>;
    private readonly filesArtifactMountpoints: Map<FilesArtifactID, string>;
    private readonly entrypointOverrideArgs: string[];
	private readonly cmdOverrideArgs: string[];
	private readonly environmentVariableOverrides: Map<string,string>;

    constructor(
            image: string,
            usedPortsSet: Set<string>,
            filesArtifactMountpoints: Map<FilesArtifactID, string>,
            entrypointOverrideArgs: string[],
            cmdOverrideArgs: string[],
            environmentVariableOverrides: Map<string,string>) {
        this.image = image;
        this.usedPortsSet = usedPortsSet;
        this.filesArtifactMountpoints = filesArtifactMountpoints;
        this.entrypointOverrideArgs = entrypointOverrideArgs;
        this.cmdOverrideArgs = cmdOverrideArgs;
        this.environmentVariableOverrides = environmentVariableOverrides;    
    }

    public getImage(): string {
        return this.image;
    }

    public getUsedPortsSet(): Set<string> {
        return this.usedPortsSet;
    }

    public getFilesArtifactMountpoints(): Map<FilesArtifactID, string> {
        return this.filesArtifactMountpoints;
    }

    public getEntrypointOverrideArgs(): string[] {
        return this.entrypointOverrideArgs;
	}
	
	public getCmdOverrideArgs(): string[] {
        return this.cmdOverrideArgs;
	}
	
	public getEnvironmentVariableOverrides(): Map<string, string> {
        return this.environmentVariableOverrides;
	}
}

// ====================================================================================================
//                                        Builder
// ====================================================================================================
// TODO Defensive copies on all these With... functions???
// Docs available at https://docs.kurtosistech.com/kurtosis-client/lib-documentation
export class ContainerConfigBuilder {
    private readonly image: string;
    private usedPortsSet: Set<string>;
    private filesArtifactMountpoints: Map<FilesArtifactID, string>;
    private entrypointOverrideArgs: string[];
	private cmdOverrideArgs: string[];
	private environmentVariableOverrides: Map<string,string>;

    constructor (image: string) {
        this.image = image;
        this.usedPortsSet = new Set();
        this.filesArtifactMountpoints = new Map();
        this.entrypointOverrideArgs = [];
        this.cmdOverrideArgs = [];
        this.environmentVariableOverrides = new Map();
    }

    public withUsedPorts(usedPortsSet: Set<string>): ContainerConfigBuilder {
        this.usedPortsSet = usedPortsSet;
        return this;
    }

    public withFilesArtifacts(filesArtifactMountpoints: Map<FilesArtifactID, string>): ContainerConfigBuilder {
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

    public build(): ContainerConfig {
        return new ContainerConfig(
            this.image,
            this.usedPortsSet,
            this.filesArtifactMountpoints,
            this.entrypointOverrideArgs,
            this.cmdOverrideArgs,
            this.environmentVariableOverrides
        );
    }
}