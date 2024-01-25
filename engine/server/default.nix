{ pkgs ? (let
  inherit (builtins) fetchTree fromJSON readFile;
  inherit ((fromJSON (readFile ./flake.lock)).nodes) nixpkgs gomod2nix;
in import (fetchTree nixpkgs.locked) {
  overlays = [ (import "${fetchTree gomod2nix.locked}/overlay.nix") ];
}), buildGoApplication ? pkgs.buildGoApplication, rev ? "dirty" }:
let
  kurtosis_version = (builtins.readFile ../../kurtosis_version.txt);
  pname = "engine";
  ldflags = pkgs.lib.concatStringsSep "\n" ([
    "-X github.com/kurtosis-tech/kurtosis/kurtosis_version.AppName=${pname}"
    "-X github.com/kurtosis-tech/kurtosis/kurtosis_version.Version=${kurtosis_version}"
    "-X github.com/kurtosis-tech/kurtosis/kurtosis_version.Commit=${rev}"
  ]);
in buildGoApplication {
  inherit pname ldflags;
  name = pname;
  pwd = ./.;
  src = ./.;
  modules = ./gomod2nix.toml;
  CGO_ENABLED = 0;
}
