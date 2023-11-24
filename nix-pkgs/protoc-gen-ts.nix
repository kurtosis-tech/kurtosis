{ pkgs ? import <nixpkgs> { } }:
pkgs.buildBazelPackage rec {
  bazel = pkgs.bazel_6;

  pname = "protoc-gen-ts";
  version = "0.8.7";

  src = pkgs.fetchFromGitHub {
    owner = "thesayyn";
    repo = pname;
    rev = version;
    hash = "sha256-PGprtSPMRTodt/SD6gpEr/n22jiNqB1/C6HJGlDndLg";
  };
  
  fetchAttrs = {
    sha256 = "sha256-VF9KyQl1UUPlJP4JoTIfPoJO3U+GGvtCQ4rj/2joGGs=";
  };

  nativeBuildInputs = [ pkgs.git pkgs.cacert pkgs.nodejs ];

  bazelTargets = [ "//package:protoc-gen-ts" ];
  buildAttrs = {
    preBuild = ''
      mkdir -p $out/bin
    '';
    installPhase = ''
      cp bazel-bin/package/protoc-gen-ts.js $out/bin/protoc-gen-ts
    '';
  };

}
