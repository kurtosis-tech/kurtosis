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
              openapi-codegen-go =
                import ./nix-pkgs/openapi-codegen.nix { inherit pkgs; };
              grpc-tools-node =
                import ./nix-pkgs/grpc-tools-node.nix { inherit pkgs; };
              protoc-gen-ts =
                import ./nix-pkgs/protoc-gen-ts.nix { inherit pkgs; };
              openapi-typescript =
                # import ./nix-pkgs/openapi-typescript.nix { inherit pkgs; };
                import ./nix-pkgs/openapi-ts { inherit pkgs; };
            in [
              goreleaser
              go_1_20
              gopls
              golangci-lint
              delve
              enumer
              nodejs_20
              yarn
              nodePackages.prettier
              protobuf
              protoc-gen-go
              protoc-gen-go-grpc
              protoc-gen-connect-go
              protoc-gen-grpc-web
              grpc-tools
              grpcui
              openapi-codegen-go
              rustc
              cargo
              rustfmt
              rust-analyzer
              clippy
              libiconv
              bash-completion
              # local definition (see above)
              grpc-tools-node
              protoc-gen-ts
              openapi-typescript
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
            Starting Kurtosis dev shell. Setup the alias to local compiled Kurtosis cli command "ktdev" by running:
            \e[32m
            source ./scripts/set_ktdev.sh
            \e[0m
            '
          '';
        };
      });
}
