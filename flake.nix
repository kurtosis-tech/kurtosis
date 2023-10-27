{
  description = "Kurtosis dev flake";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-23.05";
    unstable.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { nixpkgs, unstable, flake-utils, ... }:
    let utils = flake-utils;
    in utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        unstable_pkgs = unstable.legacyPackages.${system};
      in {
        formatter = pkgs.nixpkgs-fmt;

        devShell = pkgs.mkShell {
          nativeBuildInputs = with pkgs;
            let
              frameworks = darwin.apple_sdk.frameworks;
              inherit (lib) optional optionals;

              renamed_grpc_tools = stdenv.mkDerivation {
                name = "renamed-grpc-tools";
                version = "0.1";
                phases = [ "installPhase" ];
                installPhase = ''
                  mkdir -p $out/bin
                  cp -r ${protobuf}/include/ $out/bin/
                  cp "${grpc-tools}/bin/protoc" $out/bin/grpc_tools_node_protoc
                  cp "${grpc-tools}/bin/grpc_node_plugin" $out/bin/grpc_tools_node_protoc_plugin
                '';
              };

              ts_protoc = stdenv.mkDerivation rec {
                name = "protoc-gen-ts-wksp";
                src = fetchFromGitHub {
                  owner = "thesayyn";
                  repo = "protoc-gen-ts";
                  rev = "0.8.7";
                  hash = "sha256-PGprtSPMRTodt/SD6gpEr/n22jiNqB1/C6HJGlDndLg=";
                };
                buildInputs = [ git cacert nodejs bazel ];
                buildPhase = ''
                  export HOME=$(pwd)
                  mkdir -p $out/bin
                  npm ci
                  bazel build package
                '';
                installPhase = ''
                  cp bazel-bin/package/package/protoc-gen-ts.js $out/bin/protoc-gen-ts
                '';
              };

              elixir_pkgs = [
                unstable_pkgs.elixir
                unstable_pkgs.elixir_ls
                nodejs
                unstable_pkgs.erlang
                rebar3
              ] ++ optionals stdenv.isDarwin [
                # Dev environment
                flyctl
                postgresql
                docker
              ] ++ optionals stdenv.isLinux [
                # Docker build
                (python3.withPackages (ps: with ps; [ pip numpy ]))
                stdenv
                gcc
                gnumake
                bazel
                glibc
                gcc
                glibcLocales
              ] ++ optionals stdenv.isDarwin [
                # add macOS headers to build mac_listener and ELXA
                frameworks.CoreServices
                frameworks.CoreFoundation
                frameworks.Foundation
              ];

            in [
              goreleaser
              go_1_19
              gopls
              golangci-lint
              delve
              enumer
              nodejs_20
              yarn
              protobuf
              protoc-gen-go
              protoc-gen-go-grpc
              protoc-gen-connect-go
              protoc-gen-grpc-web
              grpc-tools
              rustc
              cargo
              rustfmt
              rust-analyzer
              clippy
              libiconv
              bash-completion
              # local definition (see above)
              renamed_grpc_tools
              ts_protoc
            ] ++ elixir_pkgs;

          shellHook = ''
            export CARGO_NET_GIT_FETCH_WITH_CLI=true
            printf '\u001b[32m
                                @@@@@@@@     
                  @@@ @@     @@@   @@@      
                @@@   @@    @@    @@        
                @   @@    @@    @@          
                  @@    @@    @@            
                @@    @@    @@              
                @   @@    @@    @@          
                @       @@  @@    @@        
                  @@   @@@    @@@    @@      
                    @@@         @@@@@@@@     
            \u001b[0m
            Starting Kurtosis dev shell. Setup the alias to local compiled Kurtosis cli command "ktdev" by running:
            \e[32m
            source ./scripts/set_ktdev.sh
            \e[0m
            '
          '';
        };
      });
}
