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
  # pname has to match the location (folder) where the main function is or use
  # subPackges to specify the file (e.g. subPackages = ["some/folder/main.go"];)
  inherit pname rev ldflags;
  version = "${rev}";
  pwd = ./.;
  src = ./.;
  modules = ./gomod2nix.toml;
  CGO_ENABLED = 0;
}
