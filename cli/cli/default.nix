{ pkgs ? (let
  inherit (builtins) fetchTree fromJSON readFile;
  inherit ((fromJSON (readFile ./flake.lock)).nodes) nixpkgs gomod2nix;
in import (fetchTree nixpkgs.locked) {
  overlays = [ (import "${fetchTree gomod2nix.locked}/overlay.nix") ];
}), buildGoApplication ? pkgs.buildGoApplication, kurtosis_version ? "dirty"
, commit_hash ? "dirty" }:
with pkgs;
let
  kurtosis_version = (builtins.readFile ../../kurtosis_version.txt);
  pname = "cli";

  # The CLI fails to compile as static using CGO_ENABLE (macOS and Linux). We need to manually use flags and add glibc
  # More info on: https://nixos.wiki/wiki/Go (also fails with musl!)
  static_linking_config = if stdenv.isLinux then {
    buildInputs = [ glibc.static ];
    nativeBuildInputs = [ stdenv ];
    CFLAGS = "-I${glibc.dev}/include";
    LDFLAGS = "-L${glibc}/lib";
  } else
    { };

  static_ldflag = if stdenv.isLinux then
    [ "-s -w -linkmode external -extldflags -static" ]
  else
    [ ];

  ldflags = lib.concatStringsSep "\n" (static_ldflag ++ [
    "-X github.com/kurtosis-tech/kurtosis/kurtosis_version.AppName=${pname}"
    "-X github.com/kurtosis-tech/kurtosis/kurtosis_version.Version=${kurtosis_version}"
    "-X github.com/kurtosis-tech/kurtosis/kurtosis_version.Commit=${commit_hash}"
  ]);

in buildGoApplication ({
  # pname has to match the location (folder) where the main function is or use
  # subPackges to specify the file (e.g. subPackages = ["some/folder/main.go"];)
  checkPhase = '''';
  inherit pname ldflags;
  version = "${kurtosis_version}";
  pwd = ./.;
  src = ./.;
  modules = ./gomod2nix.toml;
} // static_linking_config)
