import * as fs from "fs"
import {err, ok, Result} from "neverthrow";
import * as yaml from "js-yaml";

export class KurtosisYml {
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

export async  function parseKurtosisYml(kurtosisYmlFilepath: string): Promise<Result<KurtosisYml, Error>> {
    // check if the yml file actually exists
    if (!fs.existsSync(kurtosisYmlFilepath)) {
        return err(new Error(`The file '${kurtosisYmlFilepath}' does not exist.`))
    }
    const kurtosisYmlFile = fs.readFileSync(kurtosisYmlFilepath, UTF8_ENCODING)
    const parsedYAML = (yaml.load(kurtosisYmlFile) as KurtosisYml)

    if (parsedYAML.module === null || parsedYAML.module.name === null || parsedYAML.module.name === "") {
        return err(new Error(`Field module.name in kurtosis.yml needs to be set and cannot be empty`))
    }

    return ok(parsedYAML)
}
