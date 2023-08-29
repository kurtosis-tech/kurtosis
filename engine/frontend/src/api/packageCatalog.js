import axios from "axios"

export const getKurtosisPackages = async () => {
    const response = await axios.post(`https://cloud.kurtosis.com:9770/kurtosis_package_indexer.KurtosisPackageIndexer/GetPackages`, {"field":""}, {"headers":{'Content-Type': "application/json"}})
    const {data} = response
    return data.packages
}