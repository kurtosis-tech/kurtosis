{ pkgs ? (let
  inherit (builtins) fetchTree fromJSON readFile;
  inherit ((fromJSON (readFile ./flake.lock)).nodes) nixpkgs gomod2nix;
in import (fetchTree nixpkgs.locked) {
  overlays = [ (import "${fetchTree gomod2nix.locked}/overlay.nix") ];
}), gomod2nix ? pkgs.gomod2nix, rev ? "dirty" }:
let

in pkgs.mkShell {
  nativeBuildInputs = with pkgs;
    let
      node-devtools = import ./nix-pkgs/node-tools/. {
        inherit pkgs system;
        nodejs = pkgs.nodejs_20;
      };
      openapi-codegen-go =
        import ./nix-pkgs/openapi-codegen.nix { inherit pkgs; };
      grpc-tools-node = import ./nix-pkgs/grpc-tools-node.nix { inherit pkgs; };
    in [
      goreleaser
      go_1_20
      gopls
      golangci-lint
      delve
      enumer
      gomod2nix
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
      rustc
      cargo
      rustfmt
      rust-analyzer
      clippy
      libiconv
      bash-completion
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
    Starting Kurtosis dev shell. Setup the alias to local compiled Kurtosis cli command "ktdev" by running:
    \e[32m
    source ./scripts/set_ktdev.sh
    \e[0m
    '
  '';
}
