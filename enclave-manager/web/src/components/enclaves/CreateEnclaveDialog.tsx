import {useCallback, useEffect, useState} from "react";
import {KurtosisPackage} from "../../client/packageIndexer/api/kurtosis_package_indexer_pb";
import {PreloadPackage} from "./PreloadPackage";
import {ManualCreateEnclaveModal} from "./modals/ManualCreateEnclaveModal";
import {isDefined} from "../../utils";
import {ConfigureEnclaveModal} from "./modals/ConfigureEnclaveModal";

type CreateEnclaveInput = {
    defaultIsOpen: boolean | undefined;
    visibilityChangedCallback: (isOpen: boolean) => void;
};

export const CreateEnclaveDialog = ({
                                        defaultIsOpen = false,
                                        visibilityChangedCallback = () => {},
                                    }: CreateEnclaveInput) => {
    const [configureEnclaveOpen, setConfigureEnclaveOpen] = useState(false);
    const [kurtosisPackage, setKurtosisPackage] = useState<KurtosisPackage>();
    const [manualCreateEnclaveOpen, setManualCreateEnclaveOpen] = useState(false);

    useEffect(() => {
        setManualCreateEnclaveOpen(defaultIsOpen)
    }, [defaultIsOpen])

    useEffect(() => {
        visibilityChangedCallback(manualCreateEnclaveOpen)
    }, [manualCreateEnclaveOpen, configureEnclaveOpen])

    const handleManualCreateEnclaveConfirmed = (kurtosisPackage: KurtosisPackage) => {
        setKurtosisPackage(kurtosisPackage);
        setManualCreateEnclaveOpen(false);
        setConfigureEnclaveOpen(true);
    };

    const handlePreloadPackage = useCallback((kurtosisPackage: KurtosisPackage) => {
        setKurtosisPackage(kurtosisPackage);
        setConfigureEnclaveOpen(true);
    }, []);


    return (
        <>
            <PreloadPackage onPackageLoaded={handlePreloadPackage}/>
            <ManualCreateEnclaveModal
                isOpen={manualCreateEnclaveOpen}
                onClose={() => setManualCreateEnclaveOpen(false)}
                onConfirm={handleManualCreateEnclaveConfirmed}
            />
            {isDefined(kurtosisPackage) && (
                <ConfigureEnclaveModal
                    isOpen={configureEnclaveOpen}
                    onClose={() => setConfigureEnclaveOpen(false)}
                    kurtosisPackage={kurtosisPackage}
                />
            )}
        </>
    )
}
