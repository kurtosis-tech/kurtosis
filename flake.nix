{
  description = "Kurtosis dev flake";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-23.11";
    unstable.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { nixpkgs, unstable, flake-utils, ... }:
    let utils = flake-utils;
    in utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};

        unstable_pkgs = unstable.legacyPackages.${system};
        node-devtools = import ./nix-pkgs/node-tools/. {
          inherit pkgs system;
          nodejs = pkgs.nodejs_20;
        };
      in {
        formatter = pkgs.nixpkgs-fmt;

        devShell = pkgs.mkShell {
          nativeBuildInputs = with pkgs;
            let
              openapi-codegen-go =
                import ./nix-pkgs/openapi-codegen.nix { inherit pkgs; };
              grpc-tools-node =
                import ./nix-pkgs/grpc-tools-node.nix { inherit pkgs; };
            in [
              goreleaser
              go
              gopls
              golangci-lint
              delve
              enumer
              go-mockery
              nodejs_20
              node2nix
              yarn
              protobuf
              protoc-gen-go
              protoc-gen-go-grpc
              protoc-gen-connect-go
              protoc-gen-grpc-web
              grpc-tools
              grpcui
              unstable_pkgs.rustc
              unstable_pkgs.cargo
              unstable_pkgs.rustfmt
              unstable_pkgs.rust-analyzer
              unstable_pkgs.clippy
              libiconv
              bash-completion
              # Standard C libraries needed for linking
              glibc
              glibc.dev
              glibc.static
              gcc
              binutils
              stdenv.cc.cc.lib
              # local definition (see above)
              openapi-codegen-go
              grpc-tools-node
              node-devtools.nodeDependencies
            ];

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
            Starting Kurtosis dev shell. Setup the alias to local compiled Kurtosis cli command "ktdev" and "ktdebug" by running:
            \e[32m
            source ./scripts/set_kt_alias.sh
            \e[0m
            '
          '';
        };
      });
}
