import {Box, Button, Flex, Heading, Icon, Input, Text, useToast, UseToastOptions} from "@chakra-ui/react";
import { ColumnDef, createColumnHelper } from "@tanstack/react-table";
import {
    GetServicesResponse,
    ServiceInfo,
    ServiceStatus, StarlarkRunResponseLine
} from "enclave-manager-sdk/build/api_container_service_pb";
import {DataTable, isDefined, RemoveFunctions, stringifyError} from "kurtosis-ui-components";
import {useMemo, useState} from "react";
import {Link, NavigateFunction, useNavigate} from "react-router-dom";
import { ImageButton } from "../widgets/ImageButton";
import { PortsSummary } from "../widgets/PortsSummary";
import { ServiceStatusTag } from "../widgets/ServiceStatus";
import { getPortTableRows, PortsTableRow } from "./PortsTable";
import {useEnclavesContext} from "../../EnclavesContext";
import {EnclaveInfo} from "enclave-manager-sdk/build/engine_service_pb";
import {EnclaveFullInfo} from "../../types";

type ServicesTableRow = {
  serviceUUID: string;
  name: string;
  status: ServiceStatus;
  // started: DateTime | null; TODO: The api needs to support this field
  image?: string;
  ports: PortsTableRow[];
  setimage?: string
};

const serviceToRow = (enclaveUUID: string, service: ServiceInfo): ServicesTableRow => {
  return {
    serviceUUID: service.shortenedUuid,
    name: service.name,
    status: service.serviceStatus,
    image: service.container?.imageName,
    setimage: service.container?.imageName, // set to same as container image, initially
    ports: getPortTableRows(
      enclaveUUID,
      service.serviceUuid,
      service.privatePorts,
      service.maybePublicPorts,
      service.maybePublicIpAddr,
    ),
  };
};

const columnHelper = createColumnHelper<ServicesTableRow>();

type ServicesTableProps = {
  enclaveUUID: string;
  enclaveShortUUID: string;
  servicesResponse: RemoveFunctions<GetServicesResponse>;
  enclave?: RemoveFunctions<EnclaveFullInfo>;
};

export const ServicesTable = ({ enclaveUUID, enclaveShortUUID, servicesResponse, enclave }: ServicesTableProps) => {
  const { runStarlarkScript } = useEnclavesContext(); // this call will be used to run the starlark script against the enclave
  const toast = useToast(); // i have no idea what this toast is going to do
  const services = Object.values(servicesResponse.serviceInfo).map((service) => serviceToRow(enclaveUUID, service));
  const [error, setError] = useState<string>();
  const navigator = useNavigate();


  const getSetImageColumn = (
    toast: (options?: UseToastOptions) => void,
    runStarlarkScript: (
        enclave: RemoveFunctions<EnclaveInfo>,
        script: string,
        args: Record<string, string>,
        dryRun?: boolean
    ) => Promise<AsyncIterable<StarlarkRunResponseLine>>,
  ) => {
        if (!isDefined(enclave)) {
            return []
        }

        return [
            columnHelper.accessor("setimage", {
                id: "service_setimage",
                header: "Set Image",
                cell: ({ row, getValue }) => {
                    const handleSetImageSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
                        setError(undefined)

                        e.preventDefault();
                        console.log(e);
                        const inputImage = e.currentTarget.elements.namedItem("setimage") as HTMLInputElement
                        const newImage = inputImage.value.trim()
                        // fix this
                        console.log(`in handle set image ${newImage} and ${inputImage}`)

                        // TODO: prepare script:
                        // TODO: get services name
                        const script = "";
                        const args : Record<string, string> = {};
                        try {
                            const logsIterator = await runStarlarkScript(enclave, script, args, false);
                            navigator(`/enclave/${enclaveUUID}/logs`, { state: { logs: logsIterator } });
                        } catch (error: any) {
                            setError(stringifyError(error));
                        }
                    };

                    return (
                        <Flex flexDirection={"column"} gap={"10px"}>
                            <form onSubmit={handleSetImageSubmit}>
                                <Input name="Set Image" placeholder="Set new docker image" />
                            </form>
                        </Flex>
                    );
                },
            }),
        ];
    };

    const columns = useMemo<ColumnDef<ServicesTableRow, any>[]>(
    () => [
      columnHelper.accessor("name", {
        header: "Name",
        cell: ({ row, getValue }) => (
          <Link to={`/enclave/${enclaveShortUUID}/service/${row.original.serviceUUID}`}>
            <Button size={"sm"} variant={"ghost"}>
              {getValue()}
            </Button>
          </Link>
        ),
      }),
      columnHelper.accessor("status", {
        header: "Status",
        cell: (statusCell) => <ServiceStatusTag status={statusCell.getValue()} variant={"square"} />,
      }),
      columnHelper.accessor("image", {
        header: "Image",
        cell: (imageCell) => <ImageButton image={imageCell.getValue()} />,
      }),
      ...getSetImageColumn(toast, runStarlarkScript),
      columnHelper.accessor("ports", {
        header: "Ports",
        cell: (portsCell) => <PortsSummary ports={portsCell.getValue()} />,
        meta: { centerAligned: true },
      }),
      columnHelper.accessor("serviceUUID", {
        header: "Logs",
        cell: (portsCell) => (
          <Link to={`/enclave/${enclaveShortUUID}/service/${portsCell.getValue()}/logs`}>
            <Button size={"xs"} variant={"ghost"}>
              View
            </Button>
          </Link>
        ),
        enableSorting: false,
      }),
    ],
    [enclaveShortUUID, toast, runStarlarkScript],
  );

  return <DataTable columns={columns} data={services} defaultSorting={[{ id: "name", desc: true }]} />;
};
