import {useEffect, useState} from "react";
import {useNavigate} from "react-router";
import {getSinglePackageManuallyWithFullUrl} from "../api/packageCatalog";
import {useLocation} from "react-router-dom";
import {AbsoluteCenter, Center, Code, Spinner, Text} from "@chakra-ui/react";
import LoadingOverlay from "./LoadingOverflow";

const LoadSinglePackageAutomatically = () => {
    const navigate = useNavigate();
    const {state} = useLocation();
    const [thisPackage, setThisPackage] = useState({})
    const {preloadedPackage} = state;
    const [content, setContent] = useState(<><LoadingOverlay/></>)

    useEffect(() => {
        if (preloadedPackage) {
            getSinglePackageManuallyWithFullUrl(preloadedPackage)
                .then(r => {
                    if (!r.package) {
                        setContent(
                            <AbsoluteCenter axis={"both"}>
                                <Text>
                                    An error occurred while loading the package.
                                    <br/>
                                    Please verify the following url is correct:
                                    <br/>
                                    <Code>{preloadedPackage}</Code>
                                </Text>
                            </AbsoluteCenter>
                        )
                    } else {
                        setThisPackage(r.package)
                    }
                })
        }
    }, [preloadedPackage])

    useEffect(() => {
        if (thisPackage.args) {
            navigate("/catalog/create", {state: {kurtosisPackage: thisPackage}})
        }
    }, [thisPackage])

    return content;
}

export default LoadSinglePackageAutomatically;
