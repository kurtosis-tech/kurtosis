import type { PlatformPath } from "path";
import { PathJoiner } from "./path_joiner";

export class NodePathJoiner implements PathJoiner{
    constructor(
        private readonly path: PlatformPath
    ){}
    public join(a: string, b: string){
        return this.path.join(a, b)
    }
}