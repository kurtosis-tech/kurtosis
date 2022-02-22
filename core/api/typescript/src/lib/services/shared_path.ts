import * as path_browserify from "path-browserify"
import { isNode as  isExecutionEnvNode} from "browser-or-node";

// Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
export class SharedPath {

    //Absolute path in the container where this code is running
    private readonly absPathOnThisContainer: string
    //Absolute path in the service container
    private readonly absPathOnServiceContainer: string

    constructor (absPathOnThisContainer: string, absPathOnServiceContainer: string) {
        this.absPathOnThisContainer = absPathOnThisContainer;
        this.absPathOnServiceContainer = absPathOnServiceContainer;
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
    public async getChildPath(pathElement: string): Promise<SharedPath> {
        let absPathOnThisContainer;
        let absPathOnServiceContainer;
        
        if(isExecutionEnvNode){
            const path = await import( /* webpackIgnore: true */ "path")
            absPathOnThisContainer = path.join(this.absPathOnThisContainer, pathElement);
            absPathOnServiceContainer = path.join(this.absPathOnServiceContainer, pathElement);
        }else{
            absPathOnThisContainer = path_browserify.join(this.absPathOnThisContainer, pathElement);
            absPathOnServiceContainer = path_browserify.join(this.absPathOnServiceContainer, pathElement);
        }

        const sharedPath = new SharedPath(absPathOnThisContainer, absPathOnServiceContainer)
        return sharedPath
    }
}
