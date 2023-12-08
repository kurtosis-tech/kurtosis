{ pkgs ? import <nixpkgs> { } }:
with pkgs;
buildGoModule rec {
  pname = "oapi-codegen";
  powner = "deepmap";
  version = "v1.12.4";

  src = fetchgit rec {
    url = "https://github.com/deepmap/oapi-codegen";
    rev = version;
    sha256 = "sha256-ME6RnHZxX9yb4fXKJ1JjvmHLqdwKddca7VTaDIMP/cE=";
    leaveDotGit = true;
    postFetch = ''
      set -x
      cd $out
      git fetch -vv --tags ${url}
      git reset --hard tags/${version} 
      git checkout tags/${version} -b ${version}-branch
      set +x
    '';
  };

  vendorSha256 = "sha256-o9pEeM8WgGVopnfBccWZHwFR420mQAA4K/HV2RcU2wU=";

  nativeBuildInputs = [ git ];

    ldflags = [
        "-X github.com/${powner}/${pname}.BuildVersion=${version}"
        "-X github.com/${powner}/${pname}/cmd.BuildVersion=${version}"
        "-X github.com/${powner}/${pname}/cmd/${pname}.BuildVersion=${version}"
        "-X main.BuildVersion=${version}"
        "-s"
        "-w"
    ];


  installPhase = ''
    mkdir -p $out/bin
    cp $GOPATH/bin/oapi-codegen $out/bin/
  '';

  GO111MODULE = "on";
  CGO_ENABLED = 0;

  meta = with lib; {
    description = "OpenAPI Client and Server Code Generator for Go";
    homepage = "https://github.com/deepmap/oapi-codegen";
    license = licenses.asl20;
  };
}
