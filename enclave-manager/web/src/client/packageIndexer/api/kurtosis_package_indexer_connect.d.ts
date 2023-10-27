import { Empty, MethodKind } from "@bufbuild/protobuf";
import { GetPackagesResponse, ReadPackageRequest, ReadPackageResponse } from "./kurtosis_package_indexer_pb.js";
/**
 * @generated from service kurtosis_package_indexer.KurtosisPackageIndexer
 */
export declare const KurtosisPackageIndexer: {
    readonly typeName: "kurtosis_package_indexer.KurtosisPackageIndexer";
    readonly methods: {
        /**
         * @generated from rpc kurtosis_package_indexer.KurtosisPackageIndexer.IsAvailable
         */
        readonly isAvailable: {
            readonly name: "IsAvailable";
            readonly I: typeof Empty;
            readonly O: typeof Empty;
            readonly kind: MethodKind.Unary;
        };
        /**
         * @generated from rpc kurtosis_package_indexer.KurtosisPackageIndexer.GetPackages
         */
        readonly getPackages: {
            readonly name: "GetPackages";
            readonly I: typeof Empty;
            readonly O: typeof GetPackagesResponse;
            readonly kind: MethodKind.Unary;
        };
        /**
         * @generated from rpc kurtosis_package_indexer.KurtosisPackageIndexer.Reindex
         */
        readonly reindex: {
            readonly name: "Reindex";
            readonly I: typeof Empty;
            readonly O: typeof Empty;
            readonly kind: MethodKind.Unary;
        };
        /**
         * @generated from rpc kurtosis_package_indexer.KurtosisPackageIndexer.ReadPackage
         */
        readonly readPackage: {
            readonly name: "ReadPackage";
            readonly I: typeof ReadPackageRequest;
            readonly O: typeof ReadPackageResponse;
            readonly kind: MethodKind.Unary;
        };
    };
};
