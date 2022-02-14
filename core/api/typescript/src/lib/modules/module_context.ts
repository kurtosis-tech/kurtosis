import { Result } from "neverthrow";
import { ApiContainerServiceClientNode, ApiContainerServiceClientWeb } from "../..";
import { GrpcNodeModuleContextBackend } from "./module_context_backend_node";
import { GrpcWebModuleContextBackend } from "./module_context_backend_web";

export type ModuleID = string;

export interface ModuleContextBackend {
    execute(serializedParams: string, moduleId: ModuleID): Promise<Result<string, Error>>
}

// Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
export class ModuleContext {
    private readonly backend: ModuleContextBackend
    private readonly moduleId: ModuleID;

    
    constructor (client: ApiContainerServiceClientWeb | ApiContainerServiceClientNode, moduleId: ModuleID) {
        if(client instanceof ApiContainerServiceClientWeb){
            this.backend = new GrpcWebModuleContextBackend(client)
        }else{
            this.backend = new GrpcNodeModuleContextBackend(client)
        }

        this.moduleId = moduleId
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
    public async execute(serializedParams: string): Promise<Result<string, Error>> {
        return this.backend.execute(serializedParams, this.moduleId)
    }
}
