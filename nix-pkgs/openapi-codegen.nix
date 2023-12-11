{ pkgs ? import <nixpkgs> { } }:
with pkgs;
buildGoModule rec {
  pname = "oapi-codegen";
  version = "v1.16.2";
    
  src = fetchgit {
    url = "https://github.com/deepmap/oapi-codegen.git";
    rev = version;
    sha256 = "sha256-MY8rxCs03R/jbdu0AnOH2i8Cx3cnRVoY68lnj4ySe1U=";
  };

  subPackages = [ "cmd/oapi-codegen" ];
  vendorHash = "sha256-Q91naMiThWQr347Uj/O4sqRyGDYIV1FCWZYcAseuPqI=";
  proxyVendor = true;

  # Set version correctly instead of `(devel)`
  # Availble from v1.14. https://github.com/deepmap/oapi-codegen/pull/1163
  ldflags = ["-X main.noVCSVersionOverride=${version}"];

  # Tests use network
  doCheck = false;

  meta = with lib; {
    description = "OpenAPI Client and Server Code Generator for Go";
    homepage = "https://github.com/deepmap/oapi-codegen";
    license = licenses.asl20;
  };
}
