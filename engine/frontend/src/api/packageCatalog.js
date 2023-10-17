import axios from "axios"

const host = "http://localhost:9770"// "https://cloud.kurtosis.com:9770"

export const getKurtosisPackages = async () => {
    try {
        const response = await axios.post(
            `${host}/kurtosis_package_indexer.KurtosisPackageIndexer/GetPackages`,
            {"field": ""},
            {"headers": {'Content-Type': "application/json"}}
        )
        const {data} = response
        if ("packages" in data) {
            return data.packages
        }
        return []
    } catch {
        console.error("error occurred")
        return []
    }
}


/**
 * This call get's a package from the indexer directly, without having the indexer automatically discover or cache it
 * @param baseUrl
 * @param owner
 * @param name
 * @returns {Promise<{}|any>}
 */
export const getSinglePackageManually = async (
    baseUrl,
    owner,
    name,
) => {
    try {
        const response = await axios.post(
            `${host}/kurtosis_package_indexer.KurtosisPackageIndexer/ReadPackage`,
            {
                "repository_metadata": {
                    "base_url": baseUrl,
                    "owner": owner,
                    "name": name,
                }
            },
            {"headers": {'Content-Type': "application/json"}}
        )
        const {data} = response
        console.log("Read package", data)
        return data
    } catch {
        console.error("error occurred")
        return {}
    }
}

export const getSinglePackageManuallyWithFullUrl = (fullUrl) => {
    const parts = getOwnerNameFromUrl(fullUrl)
    return getSinglePackageManually(
        parts.base,
        parts.owner,
        parts.name,
    )
}

export const getOwnerNameFromUrl = (fullUrl) => {
    // // first check that it's even a valid url
    // new URL(fullUrl)

    // second check that it's a proper github url
    const components = fullUrl.split('/');
    console.log("components", components)
    if(components.length < 3) {
        throw `Illegal url, invalid number of components: ${fullUrl}`
    }
    if (components[1].length <1 || components[2].length < 1) {
        throw `Illegal url, empty components: ${fullUrl}`
    }
    return {
        "base": 'github.com',
        "owner": components[1],
        "name": components[2],
    }
}
