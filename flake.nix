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
      in {
        formatter = pkgs.nixpkgs-fmt;

        packages.default = pkgs.callPackage ./cli/cli/. {
          inherit (gomod2nix.legacyPackages.${system}) buildGoApplication;
        };

        packages.engine = pkgs.callPackage ./engine/server/. {
          inherit (gomod2nix.legacyPackages.${system}) buildGoApplication;
        };

        packages.enclave-manager = pkgs.callPackage ./enclave-manager/server/. {
          inherit rev;
          inherit (gomod2nix.legacyPackages.${system}) buildGoApplication;
        };

        devShells.default = pkgs.callPackage ./shell.nix {
          inherit rev;
          inherit (gomod2nix.legacyPackages.${system}) gomod2nix;
        };

      });
}
