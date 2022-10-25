import * as fs from "fs"
import {err, ok, Result} from "neverthrow";
import * as yaml from "js-yaml";

export class KurtosisMod {
    constructor(
        public readonly module: Module,
    ){}
}

class Module {
    constructor(
        public readonly  name: string,
    ){}
}

const UTF8_ENCODING = "utf-8";

export async  function parseKurtosisMod(kurtosisModFilepath: string): Promise<Result<KurtosisMod, Error>> {
    // check if the mod file actually exists
    if (!fs.existsSync(kurtosisModFilepath)) {
        return err(new Error(`The file '${kurtosisModFilepath}' does not exist.`))
    }
    const kurtosisModFile = fs.readFileSync(kurtosisModFilepath, UTF8_ENCODING)
    const parsedYAML = (yaml.load(kurtosisModFile) as KurtosisMod)

    if (parsedYAML.module === null || parsedYAML.module.name === null || parsedYAML.module.name === "") {
        return err(new Error(`Field module.name in kurtosis.mod needs to be set and cannot be empty`))
    }

    return ok(parsedYAML)
}
