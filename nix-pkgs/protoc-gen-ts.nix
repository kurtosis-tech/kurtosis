{ pkgs ? import <nixpkgs> { } }:
with pkgs;
let
  pkg_name = "protoc-gen-ts";
  pkg_version = "0.8.7";

  # Using fixed output derivation so we can fetch things from outside
  # and to avoid messing with Bazel's hermetic deps (mostly nodejs)
  # we'll build it direclty using using Nix's nodejs. Very very hack solution
  # after countless hours fighting Bazel deps vs nix sandbox. Maybe there a way easer
  # way to tell bazel to use external nodejs but I couldn't figure it out.
  # It seems the upstream repo is moving away from Bazel into Cargo. 
  deps = stdenv.mkDerivation rec {
    name = "${pkg_name}-build";

    src = fetchFromGitHub {
      owner = "thesayyn";
      repo = pkg_name;
      rev = pkg_version;
      hash = "sha256-PGprtSPMRTodt/SD6gpEr/n22jiNqB1/C6HJGlDndLg=";
    };

    buildInputs = [ git cacert nodejs ];

    # Setup TS/Node config files directly into the src repo
    tsconfig_json = pkgs.writeText "tsconfig_json" ''
      {
            "compilerOptions": {
                "target": "ES2020",
                "module": "EsNext",
                "moduleResolution": "node",
                "outDir": "./.build",
                "noImplicitAny": true
            }
        }
    '';

    package_json = pkgs.writeText "package_json" ''
      {
          "name": "example-node-nix",
          "version": "1.0.0",
          "main": "dist/index.js",
          "bin": {
              "example-node-nix": "dist/index.js"
          },
          "scripts": {
              "build": "tsc -p tsconfig.json && rollup -c rollup.config.js",
              "start": "node dist/index.js"
          },
          "dependencies": {
              "@grpc/grpc-js": "^1.7.3",
              "@types/google-protobuf": "^3.15.5",
              "google-protobuf": "^3.19.1"
          },
          "devDependencies": {
              "@rollup/plugin-commonjs": "^23.0.2",
              "@rollup/plugin-node-resolve": "^15.0.1",
              "@rollup/plugin-typescript": "^11.1.5",
              "@tsconfig/node16": "^1.0.1",
              "@tsconfig/node16-strictest": "^1.0.0",
              "@types/node": "^18.7.14",
              "rollup": "<3.0.0",
              "typescript": "^4.8.2",
              "tslib": "^2.6.2"
          }
      }
    '';

    rollup_config_js = pkgs.writeText "rollup_config_js" ''
      import nodeResolve from '@rollup/plugin-node-resolve';
      import commonjs from '@rollup/plugin-commonjs';
      import typescript from '@rollup/plugin-typescript';
      import fs from "node:fs";

      const executable = () => {
      return {
          name: 'executable',
          writeBundle: (options) => {
          fs.chmodSync(options.file, '755');
          },
      };
      };

      export default {
      input: "./index.ts",
      output: {
          file: '.bin/protoc-gen-ts.js',
          format: 'cjs'
      },
      plugins: [
          typescript(),
          nodeResolve(),
          commonjs(),
          executable()
      ],
      onwarn(message, warn) {
          if (message.code === "EVAL" || message.code == "THIS_IS_UNDEFINED") {
          return;
          }
          warn(message);
      }
      }
    '';

    buildPhase = ''
      cp ${package_json} src/package.json
      cp ${rollup_config_js} src/rollup.config.js
      cp ${tsconfig_json} src/tsconfig.json
      cd src
      
      echo ">> Create .tmp folder as Nix sandbox are homeless and nodejs needs one"
      mkdir -p .tmp
      export HOME=$(pwd)/.tmp
      
      echo ">> Intsalling TS/JS/nodejs deps"
      npm install
      
      echo ">> Patching npm's TS and RollUp scripts to work on Nix"
      patchShebangs --build -- node_modules/typescript/bin/tsc
      patchShebangs --build -- node_modules/rollup/dist/bin/rollup
      
      echo ">> Building"
      npm run build
    '';

    installPhase = ''
      mkdir -p $out/bin
      cp .bin/protoc-gen-ts.js $out/bin/protoc-gen-ts
    '';

    outputHashAlgo = "sha256";
    outputHashMode = "recursive";
    outputHash = "sha256-cIIFpEXlg/UogK4QskIeWFDuXReFDBKokddTB6juTZo=";
  };

in stdenv.mkDerivation rec {
  name = pkg_name;
  version = pkg_version;
  phases = [ "installPhase" ];

  # Fixed Output derivation can't have Nix references (paths /nix/...) on it
  # and we need to add it on this second (and normal) derivation. We're adding
  # a shebang to call nodejs in the compiled script.
  installPhase = ''
    mkdir -p $out/bin
    echo "$(echo '#!${nodejs}/bin/node' && cat '${deps}/bin/protoc-gen-ts')" > $out/bin/protoc-gen-ts
  '';
}
