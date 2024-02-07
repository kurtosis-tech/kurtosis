{ nixpkgs, pkgs, kurtosis, ... }:
let
  nixos-lib = import (nixpkgs + "/nixos/lib") { };
  vm_modules = import ./vm_modules.nix { inherit nixpkgs; };
  architecture =
    builtins.head (builtins.match "(.*)-.*" pkgs.stdenv.hostPlatform.system);
  container_arch =
    builtins.replaceStrings [ "aarch64" "x86_64" ] [ "arm64" "amd64" ]
    architecture;
in nixos-lib.runTest rec {
  name = "demo-test";

  hostPkgs = import nixpkgs { system = pkgs.stdenv.hostPlatform.system; };

  node = pkgs.lib.optionalAttrs pkgs.stdenv.isDarwin {
    pkgs = import nixpkgs {
      system = builtins.replaceStrings [ "darwin" ] [ "linux" ]
        pkgs.stdenv.hostPlatform.system;
    };
  };

  nodes.machine = { config, pkgs, ... }: {
    imports = vm_modules.nixosConfigurations.${architecture}.modules;
  };

  testScript = ''
    start_all()
    machine.wait_for_unit("k3s")
    machine.succeed("k3s kubectl cluster-info")

    machine.succeed("docker load < ${
      kurtosis.containers.core.${container_arch}
    }")
    machine.succeed("docker load < ${
      kurtosis.containers.engine.${container_arch}
    }")
    machine.succeed("docker load < ${
      kurtosis.containers.files-artifacts-expander.${container_arch}
    }")

    with subtest("Check that if engine and companion containers start"):
        machine.succeed("${kurtosis.cli}/bin/cli engine restart --cli-log-level debug")
  '';
}
