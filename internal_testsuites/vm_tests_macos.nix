{ nixpkgs, kurtosisPkgs, kurtosisImages, ... }:
let
  vm_modules = import ./vm_modules.nix { inherit nixpkgs; };
  architecture = "aarch64";
  container_arch = "arm64";
in nixpkgs.legacyPackages.aarch64-darwin.nixosTest rec {
  name = "demo-test";

  nodes.machine = { ... }: {
    imports = vm_modules.nixosConfigurations.${architecture}.modules;
  };

  testScript = ''
    start_all()
    machine.wait_for_unit("k3s")
    machine.succeed("k3s kubectl cluster-info")

    machine.succeed("docker load < ${kurtosisImages.core.${container_arch}}")
    machine.succeed("docker load < ${kurtosisImages.engine.${container_arch}}")
    machine.succeed("docker load < ${
      kurtosisImages.files-artifacts-expander.${container_arch}
    }")

    with subtest("Check that if engine and companion containers start"):
        machine.succeed("${kurtosisPkgs.cli}/bin/cli engine restart --cli-log-level debug")
  '';
}
