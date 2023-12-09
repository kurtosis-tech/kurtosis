{ pkgs ? import <nixpkgs> { } }:
with pkgs;
let
  pname = "openapi-typescript";
  nodejs = pkgs.nodejs_18;

  openapi-ts = pkgs.buildNpmPackage {
    name = "${pname}_node_modules";

    # The packages required by the build process
    buildInputs = [ nodejs ];
    dontNpmBuild = true;

    # The code sources for the package
    src = ./.;
    npmDepsHash = "sha256-yZcriNNQnin0IHHypPq46gdMzdT5j3J2cs40NVldfwY=";

    # How the output of the build phase
    installPhase = ''
      mkdir $out
      cp -r node_modules/ $out
    '';
  };

  openapi-ts-bin = pkgs.writeScript "${pname}_wrapper" ''
    #! ${pkgs.stdenv.shell}
    ${nodejs}/bin/npm exec --prefix ${openapi-ts}/node_modules -- openapi-typescript "$@"
  '';

in stdenv.mkDerivation {
  name = pname;
  src = ./.;
  installPhase = ''
    mkdir -p $out/bin
    cp -r ${openapi-ts-bin} $out/bin/openapi-typescript
  '';
}
