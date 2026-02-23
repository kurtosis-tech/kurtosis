/*
 * Mocks — getPortTableRows is a pure function but lives in a file that
 * imports React components, protobuf SDKs, and ESM-only packages.
 * We mock everything that isn't directly used by getPortTableRows so the
 * module can be loaded in the Jest/jsdom environment.
 */
jest.mock("@bufbuild/protobuf", () => ({ Empty: {} }));
jest.mock("true-myth", () => ({ Result: {} }));
jest.mock("enclave-manager-sdk/build/api_container_service_pb", () => ({ Port: {} }));
jest.mock("kurtosis-ui-components", () => ({
  isDefined: (it: any) => it !== null && it !== undefined,
  DataTable: () => null,
}));
jest.mock("../../EnclavesContext", () => ({
  useEnclavesContext: () => ({ addAlias: jest.fn() }),
}));
jest.mock("../widgets/PortMaybeLink", () => ({ PortMaybeLink: () => null }));
jest.mock("../utils", () => ({
  transportProtocolToString: (protocol: number) => {
    switch (protocol) {
      case 0:
        return "TCP";
      case 1:
        return "SCTP";
      case 2:
        return "UDP";
      default:
        return "";
    }
  },
}));

import { getPortTableRows } from "./PortsTable";

const makePort = (overrides: Record<string, unknown> = {}): any => ({
  number: 8080,
  transportProtocol: 0 /* TCP */,
  maybeApplicationProtocol: "http",
  locked: false,
  alias: "",
  ...overrides,
});

describe("getPortTableRows", () => {
  it("handles ports with matching public ports", () => {
    const privatePorts: Record<string, any> = {
      rpc: makePort({ number: 8545 }),
    };
    const publicPorts: Record<string, any> = {
      rpc: makePort({ number: 32000 }),
    };

    const rows = getPortTableRows("enclave-uuid", "service-uuid", privatePorts, publicPorts, "127.0.0.1");

    expect(rows).toHaveLength(1);
    expect(rows[0].port.privatePort).toBe(8545);
    expect(rows[0].port.publicPort).toBe(32000);
    expect(rows[0].link).toContain("32000");
  });

  it("handles missing public ports without throwing (regression test)", () => {
    const privatePorts: Record<string, any> = {
      rpc: makePort({ number: 8545 }),
      metrics: makePort({ number: 9090, maybeApplicationProtocol: "" }),
    };
    const publicPorts: Record<string, any> = {};

    const rows = getPortTableRows("enclave-uuid", "service-uuid", privatePorts, publicPorts, "127.0.0.1");

    expect(rows).toHaveLength(2);
    expect(rows[0].port.publicPort).toBe(8545);
    expect(rows[0].link).toContain("8545");
    expect(rows[1].port.publicPort).toBe(9090);
    expect(rows[1].link).toContain("9090");
  });

  it("handles partial public ports", () => {
    const privatePorts: Record<string, any> = {
      rpc: makePort({ number: 8545, maybeApplicationProtocol: "http" }),
      udp: makePort({ number: 30303, transportProtocol: 2 /* UDP */, maybeApplicationProtocol: "" }),
    };
    const publicPorts: Record<string, any> = {
      rpc: makePort({ number: 32000 }),
    };

    const rows = getPortTableRows("enclave-uuid", "service-uuid", privatePorts, publicPorts, "127.0.0.1");

    expect(rows).toHaveLength(2);
    const rpcRow = rows.find((r) => r.port.name === "rpc")!;
    const udpRow = rows.find((r) => r.port.name === "udp")!;

    expect(rpcRow.port.publicPort).toBe(32000);
    expect(rpcRow.link).toContain("32000");

    expect(udpRow.port.publicPort).toBe(30303);
    expect(udpRow.link).toContain("30303");
  });
});
