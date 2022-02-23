import type { GenericPathJoiner } from "../enclaves/generic_path_joiner";

// Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
export class SharedPath {

    //Absolute path in the container where this code is running
    private readonly absPathOnThisContainer: string
    //Absolute path in the service container
    private readonly absPathOnServiceContainer: string
    private readonly pathJoiner: GenericPathJoiner;

    constructor (absPathOnThisContainer: string, absPathOnServiceContainer: string, pathJoiner: GenericPathJoiner) {
        this.absPathOnThisContainer = absPathOnThisContainer;
        this.absPathOnServiceContainer = absPathOnServiceContainer;
        this.pathJoiner = pathJoiner
    }
    // Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
    public getAbsPathOnThisContainer(): string {
        return this.absPathOnThisContainer;
    }
    // Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
    public getAbsPathOnServiceContainer(): string {
        return this.absPathOnServiceContainer;
    }
    // Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
    public getChildPath(pathElement: string): SharedPath {
        const absPathOnThisContainer = this.pathJoiner.join(this.absPathOnThisContainer, pathElement);
        const absPathOnServiceContainer = this.pathJoiner.join(this.absPathOnServiceContainer, pathElement);

        const sharedPath = new SharedPath(absPathOnThisContainer, absPathOnServiceContainer, this.pathJoiner)
        return sharedPath
    }
}
