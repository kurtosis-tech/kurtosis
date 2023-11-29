import { createContext, PropsWithChildren, useContext } from "react";
import { useParams } from "react-router-dom";
import { AppPageLayout } from "../../../components/AppLayout";
import { KurtosisAlert } from "../../../components/KurtosisAlert";
import { useFullEnclave } from "../EnclavesContext";
import { EnclaveFullInfo } from "../types";

type EnclaveRouteContextState = {
  enclave: EnclaveFullInfo;
};

const EnclaveRouteContext = createContext<EnclaveRouteContextState>({ enclave: null as any });

export const EnclaveRouteContextProvider = ({ children }: PropsWithChildren) => {
  const { enclaveUUID } = useParams();
  const enclave = useFullEnclave(enclaveUUID || "Unknown");

  if (enclave.isErr) {
    return (
      <AppPageLayout>
        <KurtosisAlert message={enclave.error} />
      </AppPageLayout>
    );
  }

  return <EnclaveRouteContext.Provider value={{ enclave: enclave.value }}>{children}</EnclaveRouteContext.Provider>;
};

export const useEnclaveFromParams = () => {
  const { enclave } = useContext(EnclaveRouteContext);
  return enclave;
};
