import type { Path } from "path-browserify";
import { PathJoiner } from "./path_joiner";

export class WebPathJoiner implements PathJoiner{
    constructor(
        private readonly path: Path
    ){}
    public join(a: string, b: string){
        return this.path.join(a, b)
    }
}