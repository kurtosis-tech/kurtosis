{ pkgs ? import <nixpkgs> { } }:
let
  ld_paths = pkgs.lib.makeLibraryPath [
    pkgs.stdenv.cc.cc
    pkgs.stdenv.cc.libc
    pkgs.glibc
  ];
  ld_interpreter =
    pkgs.lib.fileContents "${pkgs.stdenv.cc}/nix-support/dynamic-linker";
in pkgs.buildBazelPackage rec {
  bazel = pkgs.bazel_6;

  pname = "protoc-gen-ts";
  version = "0.8.7";

  src = pkgs.fetchFromGitHub {
    owner = "thesayyn";
    repo = pname;
    rev = version;
    hash = "sha256-PGprtSPMRTodt/SD6gpEr/n22jiNqB1/C6HJGlDndLg";
  };

  LD_LIBRARY_PATH = ld_paths;
  NIX_LD_LIBRARY_PATH = ld_paths;
  NIX_LD = ld_interpreter;
  LD = ld_interpreter;

  fetchAttrs = {
    sha256 = "sha256-gIeB/+GJiU9TMtMFB3RrmBwFSqEpSoM/HMMxT8A9AuM=";
  };

  nativeBuildInputs = [ pkgs.git pkgs.cacert pkgs.nodejs ];

  buildPhase = ''
    bazel build --nobuild //package:protoc-gen-ts || true

    NODEBIN=$(bazel info output_base)/external/nodejs_linux_arm64/bin/nodejs/bin/node
    echo ">>>>>>>>>>>>>>>>>>>>>"
    INTER="${ld_interpreter}"
    GCCLIB="${ld_paths}"
    echo "$NODEBIN"
    echo "$INTER"
    echo "$GCCLIB"
    echo ">>>>>>>>>>>>>>>>>>>>>"
    echo $(sha256sum "$NODEBIN")
    patchelf --set-rpath "$GCCLIB" "$NODEBIN"
    echo $(sha256sum "$NODEBIN")
    patchelf --set-interpreter "$INTER" "$NODEBIN"
    echo $(sha256sum "$NODEBIN")
    echo ">>>>>>>>>>>>>>>>>>>>>"
    ldd -v "$NODEBIN"
    file "$NODEBIN"
    echo ">>>>>>>>>>>>>>>>>>>>>"
    # bazel build //package:protoc-gen-ts --spawn_strategy=standalone --action_env=LD_LIBRARY_PATH=$LD_LIBRARY_PATH
    bazel build package --action_env=LD_LIBRARY_PATH=$LD_LIBRARY_PATH --action_env=PATH=$PATH --action_env=NIX_LD=$LD --spawn_strategy=local

    mkdir -p $out/bin
  '';
  bazelTargets = [ "//package:protoc-gen-ts" ];
  installPhase = ''
    ls -lah 
    ls -lah bazel-bin
    ls -lah bazel-bin/package
    cp bazel-bin/package/protoc-gen-ts.js $out/bin/protoc-gen-ts
  '';
  buildAttrs = {};
}
