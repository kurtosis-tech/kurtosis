import * as fs from "fs"
import {err, ok, Result} from "neverthrow";
import * as yaml from "js-yaml";
import {KURTOSIS_YAML_FILENAME} from "./enclave_context";

const PACKAGES_URL = "https://docs.kurtosis.com/concepts-reference/packages";

export class KurtosisYaml {
    constructor(
        public readonly  name: string,
    ){}
}

const UTF8_ENCODING = "utf-8";

export async  function parseKurtosisYaml(kurtosisYamlFilepath: string): Promise<Result<KurtosisYaml, Error>> {
    // check if the yml file actually exists
    if (!fs.existsSync(kurtosisYamlFilepath)) {
        return err(new Error(`Couldn't find a '${KURTOSIS_YAML_FILENAME}' in the root of the package at '${kurtosisYamlFilepath}'. Packages are expected to have a '${KURTOSIS_YAML_FILENAME}' at root; have a look at '${PACKAGES_URL}' for more`))
    }

    let kurtosisYamlFile: string
    try {
        kurtosisYamlFile = fs.readFileSync(kurtosisYamlFilepath, UTF8_ENCODING)
    } catch(error) {
        return err(new Error(
            `An error occurred while reading the '${KURTOSIS_YAML_FILENAME}' file at '${kurtosisYamlFilepath}'`
        ));
    }

    let parsedYAML: KurtosisYaml
    try {
        parsedYAML = (yaml.load(kurtosisYamlFile) as KurtosisYaml)
    } catch(error) {
        return err(new Error(
            `"An error occurred while parsing the '${KURTOSIS_YAML_FILENAME}' file at '${kurtosisYamlFilepath}'`
        ));
    }

    if ( parsedYAML.name === null || parsedYAML.name === "") {
        return err(new Error(`Field 'name', which is the Starlark package's name, in '${KURTOSIS_YAML_FILENAME}' needs to be set and cannot be empty`))
    }

    return ok(parsedYAML)
}
