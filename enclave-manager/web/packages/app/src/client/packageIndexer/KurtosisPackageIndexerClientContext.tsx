import { assertDefined } from "kurtosis-ui-components";
import { createContext, PropsWithChildren, useContext, useMemo } from "react";
import { KurtosisPackageIndexerClient } from "./KurtosisPackageIndexerClient";

type KurtosisPackageIndexerClientContextState = {
  client: KurtosisPackageIndexerClient | null;
};

const KurtosisPackageIndexerClientContext = createContext<KurtosisPackageIndexerClientContextState>({ client: null });

export const KurtosisPackageIndexerProvider = ({ children }: PropsWithChildren) => {
  const errorHandlingClient = useMemo(() => {
    return new Proxy(new KurtosisPackageIndexerClient(), {
      get(target, prop: string | symbol) {
        if (prop === "getPackages" || prop === "readPackage") {
          return new Proxy(target[prop], {
            apply: (target, thisArg, argumentsList) => {
              const methodResult = Reflect.apply(target, thisArg, argumentsList) as ReturnType<typeof target>;
              return methodResult.then((r) => {
                if (r.isErr) {
                  console.error(r.error);
                }
                return r;
              });
            },
          });
        } else {
          return Reflect.get(target, prop);
        }
      },
    });
  }, []);

  return (
    <KurtosisPackageIndexerClientContext.Provider value={{ client: errorHandlingClient }}>
      {children}
    </KurtosisPackageIndexerClientContext.Provider>
  );
};

export const useKurtosisPackageIndexerClient = (): KurtosisPackageIndexerClient => {
  const { client } = useContext(KurtosisPackageIndexerClientContext);

  assertDefined(
    client,
    `useKurtosisPackageIndexerClient used incorrectly - KurtosisPackageIndexerClient is not currently available.`,
  );

  return client;
};
