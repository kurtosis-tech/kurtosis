import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Message, proto3, Timestamp } from "@bufbuild/protobuf";
/**
 * @generated from enum kurtosis_package_indexer.ArgumentValueType
 */
export declare enum ArgumentValueType {
    /**
     * @generated from enum value: BOOL = 0;
     */
    BOOL = 0,
    /**
     * @generated from enum value: STRING = 1;
     */
    STRING = 1,
    /**
     * @generated from enum value: INTEGER = 2;
     */
    INTEGER = 2,
    /**
     * @generated from enum value: DICT = 4;
     */
    DICT = 4,
    /**
     * @generated from enum value: JSON = 5;
     */
    JSON = 5,
    /**
     * @generated from enum value: LIST = 6;
     */
    LIST = 6
}
/**
 * @generated from message kurtosis_package_indexer.ReadPackageRequest
 */
export declare class ReadPackageRequest extends Message<ReadPackageRequest> {
    /**
     * @generated from field: kurtosis_package_indexer.PackageRepository repository_metadata = 1;
     */
    repositoryMetadata?: PackageRepository;
    constructor(data?: PartialMessage<ReadPackageRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "kurtosis_package_indexer.ReadPackageRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ReadPackageRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ReadPackageRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ReadPackageRequest;
    static equals(a: ReadPackageRequest | PlainMessage<ReadPackageRequest> | undefined, b: ReadPackageRequest | PlainMessage<ReadPackageRequest> | undefined): boolean;
}
/**
 * @generated from message kurtosis_package_indexer.ReadPackageResponse
 */
export declare class ReadPackageResponse extends Message<ReadPackageResponse> {
    /**
     * @generated from field: optional kurtosis_package_indexer.KurtosisPackage package = 1;
     */
    package?: KurtosisPackage;
    constructor(data?: PartialMessage<ReadPackageResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "kurtosis_package_indexer.ReadPackageResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ReadPackageResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ReadPackageResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ReadPackageResponse;
    static equals(a: ReadPackageResponse | PlainMessage<ReadPackageResponse> | undefined, b: ReadPackageResponse | PlainMessage<ReadPackageResponse> | undefined): boolean;
}
/**
 * @generated from message kurtosis_package_indexer.GetPackagesResponse
 */
export declare class GetPackagesResponse extends Message<GetPackagesResponse> {
    /**
     * @generated from field: repeated kurtosis_package_indexer.KurtosisPackage packages = 1;
     */
    packages: KurtosisPackage[];
    constructor(data?: PartialMessage<GetPackagesResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "kurtosis_package_indexer.GetPackagesResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetPackagesResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetPackagesResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetPackagesResponse;
    static equals(a: GetPackagesResponse | PlainMessage<GetPackagesResponse> | undefined, b: GetPackagesResponse | PlainMessage<GetPackagesResponse> | undefined): boolean;
}
/**
 * @generated from message kurtosis_package_indexer.KurtosisPackage
 */
export declare class KurtosisPackage extends Message<KurtosisPackage> {
    /**
     * @generated from field: string name = 1;
     */
    name: string;
    /**
     * @generated from field: repeated kurtosis_package_indexer.PackageArg args = 2;
     */
    args: PackageArg[];
    /**
     * @generated from field: uint64 stars = 3;
     */
    stars: bigint;
    /**
     * @generated from field: string description = 4;
     */
    description: string;
    /**
     * deprecated: use a combination of repository_url and root_path instead
     *
     * @generated from field: optional string url = 5;
     */
    url?: string;
    /**
     * @generated from field: string entrypoint_description = 6;
     */
    entrypointDescription: string;
    /**
     * @generated from field: string returns_description = 7;
     */
    returnsDescription: string;
    /**
     * @generated from field: kurtosis_package_indexer.PackageRepository repository_metadata = 8;
     */
    repositoryMetadata?: PackageRepository;
    /**
     * @generated from field: string parsing_result = 9;
     */
    parsingResult: string;
    /**
     * @generated from field: google.protobuf.Timestamp parsing_time = 10;
     */
    parsingTime?: Timestamp;
    /**
     * @generated from field: string version = 11;
     */
    version: string;
    constructor(data?: PartialMessage<KurtosisPackage>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "kurtosis_package_indexer.KurtosisPackage";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): KurtosisPackage;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): KurtosisPackage;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): KurtosisPackage;
    static equals(a: KurtosisPackage | PlainMessage<KurtosisPackage> | undefined, b: KurtosisPackage | PlainMessage<KurtosisPackage> | undefined): boolean;
}
/**
 * @generated from message kurtosis_package_indexer.PackageArg
 */
export declare class PackageArg extends Message<PackageArg> {
    /**
     * @generated from field: string name = 1;
     */
    name: string;
    /**
     * @generated from field: bool is_required = 2;
     */
    isRequired: boolean;
    /**
     * @generated from field: string description = 4;
     */
    description: string;
    /**
     * @generated from field: kurtosis_package_indexer.PackageArgumentType typeV2 = 5;
     */
    typeV2?: PackageArgumentType;
    constructor(data?: PartialMessage<PackageArg>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "kurtosis_package_indexer.PackageArg";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): PackageArg;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): PackageArg;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): PackageArg;
    static equals(a: PackageArg | PlainMessage<PackageArg> | undefined, b: PackageArg | PlainMessage<PackageArg> | undefined): boolean;
}
/**
 * @generated from message kurtosis_package_indexer.PackageArgumentType
 */
export declare class PackageArgumentType extends Message<PackageArgumentType> {
    /**
     * @generated from field: kurtosis_package_indexer.ArgumentValueType top_level_type = 1;
     */
    topLevelType: ArgumentValueType;
    /**
     * @generated from field: optional kurtosis_package_indexer.ArgumentValueType inner_type_1 = 2;
     */
    innerType1?: ArgumentValueType;
    /**
     * @generated from field: optional kurtosis_package_indexer.ArgumentValueType inner_type_2 = 3;
     */
    innerType2?: ArgumentValueType;
    constructor(data?: PartialMessage<PackageArgumentType>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "kurtosis_package_indexer.PackageArgumentType";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): PackageArgumentType;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): PackageArgumentType;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): PackageArgumentType;
    static equals(a: PackageArgumentType | PlainMessage<PackageArgumentType> | undefined, b: PackageArgumentType | PlainMessage<PackageArgumentType> | undefined): boolean;
}
/**
 * @generated from message kurtosis_package_indexer.PackageRepository
 */
export declare class PackageRepository extends Message<PackageRepository> {
    /**
     * @generated from field: string base_url = 1;
     */
    baseUrl: string;
    /**
     * @generated from field: string owner = 2;
     */
    owner: string;
    /**
     * @generated from field: string name = 3;
     */
    name: string;
    /**
     * @generated from field: string root_path = 4;
     */
    rootPath: string;
    constructor(data?: PartialMessage<PackageRepository>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "kurtosis_package_indexer.PackageRepository";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): PackageRepository;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): PackageRepository;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): PackageRepository;
    static equals(a: PackageRepository | PlainMessage<PackageRepository> | undefined, b: PackageRepository | PlainMessage<PackageRepository> | undefined): boolean;
}
