import * as fs from "fs"
import {err, ok, Result} from "neverthrow";
import * as yaml from "js-yaml";

export class KurtosisYaml {
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

export async  function parseKurtosisYaml(kurtosisYamlFilepath: string): Promise<Result<KurtosisYaml, Error>> {
    // check if the yml file actually exists
    if (!fs.existsSync(kurtosisYamlFilepath)) {
        return err(new Error(`The file '${kurtosisYamlFilepath}' does not exist.`))
    }
    const kurtosisYamlFile = fs.readFileSync(kurtosisYamlFilepath, UTF8_ENCODING)
    const parsedYAML = (yaml.load(kurtosisYamlFile) as KurtosisYaml)

    if (parsedYAML.module === null || parsedYAML.module.name === null || parsedYAML.module.name === "") {
        return err(new Error(`Field module.name in kurtosis.yml needs to be set and cannot be empty`))
    }

    return ok(parsedYAML)
}
