{

  inputs.flake-utils.url = "github:numtide/flake-utils";

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        vms = import ./vm_modules.nix { inherit nixpkgs; };
        architecture = builtins.head (builtins.match "(.*)-.*" system);
      in {
        packages.shellVM =
          vms.nixosConfigurations.${architecture}.VM.config.system.build.vm;
      });

}
