{
  description = "Kurtosis dev flake";

  inputs = {
    # This is workaround for enabling the NixOS VM tests on macOS. This is a WIP PR and should be removed once merged.
    # https://github.com/NixOS/nixpkgs/pull/282401
    nixpkgs_gabi.url = "github:Gabriella439/nixpkgs/gabriella/macOS_NixOS_test";

    nixpkgs.url = "github:NixOS/nixpkgs/nixos-23.11";
    unstable.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    gomod2nix.url = "github:nix-community/gomod2nix";
    gomod2nix.inputs.nixpkgs.follows = "nixpkgs";
    gomod2nix.inputs.flake-utils.follows = "flake-utils";
    flake-compat.url = "github:edolstra/flake-compat";
  };

  outputs =
    { self, nixpkgs, unstable, flake-utils, gomod2nix, nixpkgs_gabi, ... }:
    let
      utils = flake-utils;
      all_systems_output = utils.lib.eachDefaultSystem (system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
          unstable_pkgs = unstable.legacyPackages.${system};
          commit_hash = "${self.shortRev or self.dirtyShortRev or "dirty"}";
          kurtosis_version = let
            file_ver = (builtins.readFile ./kurtosis_version.txt);
            clean_ver =
              builtins.match "[[:space:]]*([^[:space:]]+)[[:space:]]*" file_ver;
          in if clean_ver == null then commit_hash else builtins.head clean_ver;
        in rec {
          formatter = pkgs.nixpkgs-fmt;

          devShells.default = pkgs.callPackage ./shell.nix {
            inherit commit_hash kurtosis_version;
            inherit (gomod2nix.legacyPackages.${system}) mkGoEnv gomod2nix;
          };

          packages.default = packages.cli;

          packages.cli = pkgs.callPackage ./cli/cli/. {
            inherit commit_hash kurtosis_version;
            inherit (gomod2nix.legacyPackages.${system}) buildGoApplication;
          };

          packages.engine = pkgs.callPackage ./engine/server/. {
            inherit commit_hash kurtosis_version;
            inherit (gomod2nix.legacyPackages.${system}) buildGoApplication;
          };

          packages.enclave-manager =
            pkgs.callPackage ./enclave-manager/server/. {
              inherit commit_hash kurtosis_version;
              inherit (gomod2nix.legacyPackages.${system}) buildGoApplication;
            };

          packages.core = pkgs.callPackage ./core/server/. {
            inherit commit_hash kurtosis_version;
            inherit (gomod2nix.legacyPackages.${system}) buildGoApplication;
          };

          packages.files-artifacts-expander =
            pkgs.callPackage ./core/files_artifacts_expander/. {
              inherit commit_hash kurtosis_version;
              inherit (gomod2nix.legacyPackages.${system}) buildGoApplication;
            };

          checks.cli = packages.cli;
          checks.core = packages.core;
          checks.engine = packages.engine;

          cross-packages.cli = let
            architectures = [ "amd64" "arm64" ];
            OSs = [ "linux" "darwin" "windows" ];
            all = pkgs.lib.lists.crossLists (arch: os: {
              "${toString arch}-${toString os}" = packages.cli.overrideAttrs
                (old:
                  old // {
                    GOOS = os;
                    GOARCH = arch;
                    # CGO_ENABLED = disabled breaks the CLI compilation 
                    # CGO_ENABLED = 0;
                    doCheck = false;
                  });
            }) [ architectures OSs ];
          in pkgs.lib.foldl' (set: acc: acc // set) { } all;

          containers = let
            architectures = [ "amd64" "arm64" ];
            service_names = [ "engine" "core" "files-artifacts-expander" ];
            os = "linux";
            all = pkgs.lib.lists.crossLists (arch: service_name: {
              "${service_name}" = {
                "${toString arch}" = let
                  # if running from linux no cross-compilation is needed to palce the service in a container
                  needsCrossCompilation = "${arch}-${os}"
                    != builtins.replaceStrings [ "aarch64" "x86_64" ] [
                      "arm64"
                      "amd64"
                    ] system;
                  service = if !needsCrossCompilation then
                    packages.${service_name}.overrideAttrs
                    (old: old // { doCheck = false; })
                  else
                    packages.${service_name}.overrideAttrs (old:
                      old // {
                        GOOS = os;
                        GOARCH = arch;
                        # CGO_ENABLED = disabled breaks the CLI compilation 
                        # CGO_ENABLED = 0;
                        doCheck = false;
                      });
                in pkgs.dockerTools.buildImage {
                  name = "kurtosistech/${service_name}";
                  tag = kurtosis_version;
                  created = "now";
                  copyToRoot = pkgs.buildEnv {
                    name = "image-root";
                    paths = [
                      service
                      pkgs.bashInteractive
                      pkgs.nettools
                      pkgs.gnugrep
                    ];
                    pathsToLink = [ "/bin" ];
                  };
                  architecture = arch;
                  config.Cmd = if !needsCrossCompilation then
                    [ "${service}/bin/${service.pname}" ]
                  else
                    [ "${service}/bin/${os}_${arch}/${service.pname}" ];
                };
              };
            }) [ architectures service_names ];
          in pkgs.lib.foldl' (set: acc: pkgs.lib.recursiveUpdate acc set) { }
          all;

          packages.integrationTest = import ./internal_testsuites/vm_tests.nix
            (self.inputs // {
              inherit pkgs nixpkgs;
              kurtosisPkgs = packages;
              kurtosisImages = containers;
            });
        });

      # This is workaround for enabling the NixOS VM tests on macOS. This is a WIP PR and should be removed once merged.
      # https://github.com/NixOS/nixpkgs/pull/282401
      macos_ete_test = {
        packages.aarch64-darwin.integrationTestMacOS =
          import ./internal_testsuites/vm_tests_macos.nix (self.inputs // {
            nixpkgs = nixpkgs_gabi;
            kurtosisPkgs = all_systems_output.packages.aarch64-linux;
            kurtosisImages = all_systems_output.containers.aarch64-darwin;
          });
      };

    in nixpkgs.lib.recursiveUpdate all_systems_output macos_ete_test;
}
