{ nixpkgs, pkgs, containers, ... }:
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

    machine.succeed("docker load < ${containers.core.${container_arch}}")
    machine.succeed("docker load < ${containers.engine.${container_arch}}")
    machine.succeed("docker load < ${
      containers.files_artifacts_expander.${container_arch}
    }")
    with subtest("Check that module fails the right way"):
        images, stdout = machine.execute("docker images 2>&1")
        assert 1 == 2, f"Expected: `images` but got {stdout}"

  '';
}
