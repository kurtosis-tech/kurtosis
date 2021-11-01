import * as path from "path"

// Docs available at https://docs.kurtosistech.com/kurtosis-client/lib-documentation
export class SharedPath {

    //Absolute path in the container where this code is running
    private readonly absPathOnThisContainer: string
    //Absolute path in the service container
    private readonly absPathOnServiceContainer: string

    constructor (absPathOnThisContainer: string, absPathOnServiceContainer: string) {
        this.absPathOnThisContainer = absPathOnThisContainer;
        this.absPathOnServiceContainer = absPathOnServiceContainer;
    }
    // Docs available at https://docs.kurtosistech.com/kurtosis-client/lib-documentation
    public getAbsPathOnThisContainer(): string {
        return this.absPathOnThisContainer;
    }
    // Docs available at https://docs.kurtosistech.com/kurtosis-client/lib-documentation
    public getAbsPathOnServiceContainer(): string {
        return this.absPathOnServiceContainer;
    }
    // Docs available at https://docs.kurtosistech.com/kurtosis-client/lib-documentation
    public getChildPath(pathElement: string): SharedPath {
        const absPathOnThisContainer = path.join(this.absPathOnThisContainer, pathElement);

        const absPathOnServiceContainer = path.join(this.absPathOnServiceContainer, pathElement);

        const sharedPath = new SharedPath(absPathOnThisContainer, absPathOnServiceContainer)

        return sharedPath
    }
}
