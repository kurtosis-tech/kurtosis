{
  description = "Kurtosis dev flake";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-23.11";
    unstable.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    gomod2nix.url = "github:nix-community/gomod2nix";
    gomod2nix.inputs.nixpkgs.follows = "nixpkgs";
    gomod2nix.inputs.flake-utils.follows = "flake-utils";
  };

  outputs = { self, nixpkgs, unstable, flake-utils, gomod2nix, ... }:
    let utils = flake-utils;
    in utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        unstable_pkgs = unstable.legacyPackages.${system};
        rev = "${self.shortRev or self.dirtyRev or "dirty"}";
      in rec {
        formatter = pkgs.nixpkgs-fmt;

        devShells.default = pkgs.callPackage ./shell.nix {
          inherit rev;
          inherit (gomod2nix.legacyPackages.${system}) mkGoEnv gomod2nix;
        };

        packages.default = packages.cli;

        packages.cli = pkgs.callPackage ./cli/cli/. {
          inherit (gomod2nix.legacyPackages.${system}) buildGoApplication;
        };

        packages.engine = pkgs.callPackage ./engine/server/. {
          inherit (gomod2nix.legacyPackages.${system}) buildGoApplication;
        };

        packages.enclave-manager = pkgs.callPackage ./enclave-manager/server/. {
          inherit rev;
          inherit (gomod2nix.legacyPackages.${system}) buildGoApplication;
        };

        packages.core = pkgs.callPackage ./core/server/. {
          inherit rev;
          inherit (gomod2nix.legacyPackages.${system}) buildGoApplication;
        };

        packages.files-artifacts-expander =
          pkgs.callPackage ./core/files_artifacts_expander/. {
            inherit rev;
            inherit (gomod2nix.legacyPackages.${system}) buildGoApplication;
          };

        checks.cli = packages.cli;
        checks.core = packages.core;
        checks.engine = packages.engine;

        packages.cli-binaries = let
          architectures = [ "amd64" "arm64" ];
          OSs = [ "linux" "darwin" "windows" ];
          all = pkgs.lib.lists.crossLists (arch: os: {
            "${toString arch}-${toString os}" = packages.cli.overrideAttrs (old:
              old // {
                GOOS = os;
                GOARCH = arch;
                # CGO_ENABLED = disabled breaks the CLI compilation 
                # CGO_ENABLED = 0;
                doCheck = false;
              });
          }) [ architectures OSs ];
        in pkgs.lib.foldl' (set: acc: acc // set) { } all;

        packages.containers = let
          architectures = [ "amd64" "arm64" ];
          service_names = [ "engine" "core" "files-artifacts-expander" ];
          os = "linux";
          all = pkgs.lib.lists.crossLists (arch: service_name: {
            "${service_name}" = {
              "${toString arch}" = let
                tag = "${self.shortRev or "dirty"}";
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
                tag = tag;
                created = "now";
                copyToRoot = pkgs.buildEnv {
                  name = "image-root";
                  paths =
                    [ service pkgs.bashInteractive pkgs.nettools pkgs.gnugrep ];
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
        in pkgs.lib.foldl' (set: acc: acc // set) { } all;

        packages.integrationTest = import ./internal_testsuites/vm_tests.nix
          (self.inputs // {
            inherit pkgs nixpkgs;
            kurtosis = packages;
          });
      });
}
