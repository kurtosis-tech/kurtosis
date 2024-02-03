{ pkgs ? (let
  inherit (builtins) fetchTree fromJSON readFile;
  inherit ((fromJSON (readFile ./flake.lock)).nodes) nixpkgs gomod2nix;
in import (fetchTree nixpkgs.locked) {
  overlays = [ (import "${fetchTree gomod2nix.locked}/overlay.nix") ];
}), buildGoApplication ? pkgs.buildGoApplication, rev ? "dirty" }:
with pkgs;
let
  kurtosis_version = (builtins.readFile ../../kurtosis_version.txt);
  pname = "kurtosis";
  ldflags = lib.concatStringsSep "\n" ([
    "-X github.com/kurtosis-tech/kurtosis/kurtosis_version.AppName=${pname}"
    "-X github.com/kurtosis-tech/kurtosis/kurtosis_version.Version=${kurtosis_version}"
    "-X github.com/kurtosis-tech/kurtosis/kurtosis_version.Commit=${rev}"
  ] ++ lib.optionalAttrs stdenv.isLinux
    [ "-s -w -linkmode external -extldflags -static" ]);
in buildGoApplication {
  # pname has to match the location (folder) where the main function is or use
  # subPackges to specify the file (e.g. subPackages = ["some/folder/main.go"];)
  inherit pname rev ldflags;
  version = "${rev}";
  pwd = ./.;
  src = ./.;
  modules = ./gomod2nix.toml;
  # The CLI fails to compile as static using CGO_ENABLE. We need to manually use flags and add glibc
  # More info on: https://nixos.wiki/wiki/Go (also fails with musl!)
  CGO_ENABLED = if stdenv.isLinux then "" else "0";
  buildInputs = lib.optionalAttrs stdenv.isLinux [ glibc.static ];
  nativeBuildInputs = lib.optionalAttrs stdenv.isLinux [ stdenv ];
  CFLAGS = "-I${glibc.dev}/include";
  LDFLAGS = "-L${glibc}/lib";
}
