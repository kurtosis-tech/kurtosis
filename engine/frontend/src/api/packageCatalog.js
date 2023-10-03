import axios from "axios"

export const getKurtosisPackages = async () => {
    try {
        const response = await axios.post(`https://cloud.kurtosis.com:9770/kurtosis_package_indexer.KurtosisPackageIndexer/GetPackages`, {"field":""}, {"headers":{'Content-Type': "application/json"}})
        const {data} = response
        if ("packages" in data) {
            return data.packages
        } 
        return []
    } catch {
        console.log("error occurred")
        return []
    } 
}

