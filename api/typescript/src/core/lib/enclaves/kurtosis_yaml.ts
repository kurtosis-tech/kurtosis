import * as fs from "fs"
import {err, ok, Result} from "neverthrow";
import * as yaml from "js-yaml";
import {KURTOSIS_YAML_FILENAME} from "./enclave_context";

const DEPENDENCIES_URL = "https://docs.kurtosis.com/reference/starlark-reference/#dependencies";

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
        return err(new Error(`Couldn't find a '${KURTOSIS_YAML_FILENAME}' in the root of the package at '${kurtosisYamlFilepath}'. Packages are expected to have a '${KURTOSIS_YAML_FILENAME}' at root; have a look at '${DEPENDENCIES_URL}' for more`))
    }

    let kurtosisYamlFile: string
    try {
        kurtosisYamlFile = fs.readFileSync(kurtosisYamlFilepath, UTF8_ENCODING)
    } catch(error) {
        if (error instanceof Error) {
            return err(error);
        }
        return err(new Error(
            `An error occurred while reading the '${KURTOSIS_YAML_FILENAME}' file at '${kurtosisYamlFilepath}'`
        ));
    }

    let parsedYAML: KurtosisYaml
    try {
        parsedYAML = (yaml.load(kurtosisYamlFile) as KurtosisYaml)
    } catch(error) {
        if (error instanceof Error) {
            return err(error);
        }
        return err(new Error(
            `"An error occurred while parsing the '${KURTOSIS_YAML_FILENAME}' file at '${kurtosisYamlFilepath}'`
        ));
    }

    if (parsedYAML.module === null || parsedYAML.module.name === null || parsedYAML.module.name === "") {
        return err(new Error(`Field module.name in '${KURTOSIS_YAML_FILENAME}' needs to be set and cannot be empty`))
    }

    return ok(parsedYAML)
}
