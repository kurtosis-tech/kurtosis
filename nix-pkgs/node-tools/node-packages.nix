# This file has been generated by node2nix 1.11.1. Do not edit!

{ nodeEnv, fetchurl, fetchgit, nix-gitignore, stdenv, lib
, globalBuildInputs ? [ ] }:

let
  sources = {
    "@bufbuild/protobuf-1.7.1" = {
      name = "_at_bufbuild_slash_protobuf";
      packageName = "@bufbuild/protobuf";
      version = "1.7.1";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/@bufbuild/protobuf/-/protobuf-1.7.1.tgz";
        sha512 =
          "UlI3lKLFBjZQJ0cHf47YUH6DzZxZYWk3sf6dKYyPUaXrfXq4z+zZqNO3q0lPUzyJgh14s6VscjcNFBaQBhYd9Q==";
      };
    };
    "@bufbuild/protoc-gen-es-1.7.1" = {
      name = "_at_bufbuild_slash_protoc-gen-es";
      packageName = "@bufbuild/protoc-gen-es";
      version = "1.7.1";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/@bufbuild/protoc-gen-es/-/protoc-gen-es-1.7.1.tgz";
        sha512 =
          "N1diiVcDkTTNX+b9rDY8EVgOXu0W8kRmf2w3nbYi8q/hfM6vBg4zry0m4v3ARSgKp60bCey1WUDBuiynm5+PqQ==";
      };
    };
    "@bufbuild/protoplugin-1.7.1" = {
      name = "_at_bufbuild_slash_protoplugin";
      packageName = "@bufbuild/protoplugin";
      version = "1.7.1";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/@bufbuild/protoplugin/-/protoplugin-1.7.1.tgz";
        sha512 =
          "bnPFXs38IXjL2EdpkthkCa/+SXOxERnXyV///rQj1wyidJmw21wOvqpucuIh25YnPtdrUItcIFFDVCoKPkuCPQ==";
      };
    };
    "@connectrpc/connect-1.3.0" = {
      name = "_at_connectrpc_slash_connect";
      packageName = "@connectrpc/connect";
      version = "1.3.0";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/@connectrpc/connect/-/connect-1.3.0.tgz";
        sha512 =
          "kTeWxJnLLtxKc2ZSDN0rIBgwfP8RwcLknthX4AKlIAmN9ZC4gGnCbwp+3BKcP/WH5c8zGBAWqSY3zeqCM+ah7w==";
      };
    };
    "@connectrpc/protoc-gen-connect-es-1.3.0" = {
      name = "_at_connectrpc_slash_protoc-gen-connect-es";
      packageName = "@connectrpc/protoc-gen-connect-es";
      version = "1.3.0";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/@connectrpc/protoc-gen-connect-es/-/protoc-gen-connect-es-1.3.0.tgz";
        sha512 =
          "UbQN48c0zafo5EFSsh3POIJP6ofYiAgKE1aFOZ2Er4W3flUYihydZdM6TQauPkn7jDj4w9jjLSTTZ9//ecUbPA==";
      };
    };
    "@typescript/vfs-1.5.0" = {
      name = "_at_typescript_slash_vfs";
      packageName = "@typescript/vfs";
      version = "1.5.0";
      src = fetchurl {
        url = "https://registry.npmjs.org/@typescript/vfs/-/vfs-1.5.0.tgz";
        sha512 =
          "AJS307bPgbsZZ9ggCT3wwpg3VbTKMFNHfaY/uF0ahSkYYrPF2dSSKDNIDIQAHm9qJqbLvCsSJH7yN4Vs/CsMMg==";
      };
    };
    "debug-4.3.4" = {
      name = "debug";
      packageName = "debug";
      version = "4.3.4";
      src = fetchurl {
        url = "https://registry.npmjs.org/debug/-/debug-4.3.4.tgz";
        sha512 =
          "PRWFHuSU3eDtQJPvnNY7Jcket1j0t5OuOsFzPPzsekD52Zl8qUfFIPEiswXqIvHWGVHOgX+7G/vCNNhehwxfkQ==";
      };
    };
    "ms-2.1.2" = {
      name = "ms";
      packageName = "ms";
      version = "2.1.2";
      src = fetchurl {
        url = "https://registry.npmjs.org/ms/-/ms-2.1.2.tgz";
        sha512 =
          "sGkPx+VjMtmA6MX27oA4FBFELFCZZ4S4XqeGOXCv68tT+jb3vk/RyaKWP0PTKyWtmLSM0b+adUTEvbs1PEaH2w==";
      };
    };
    "typescript-4.5.2" = {
      name = "typescript";
      packageName = "typescript";
      version = "4.5.2";
      src = fetchurl {
        url = "https://registry.npmjs.org/typescript/-/typescript-4.5.2.tgz";
        sha512 =
          "5BlMof9H1yGt0P8/WF+wPNw6GfctgGjXp5hkblpyT+8rkASSmkUKMXrxR0Xg8ThVCi/JnHQiKXeBaEwCeQwMFw==";
      };
    };
  };
  args = {
    name = "kurtosis-node-tools";
    packageName = "kurtosis-node-tools";
    version = "1.0.0";
    src = ./.;
    dependencies = [
      sources."@bufbuild/protobuf-1.7.1"
      sources."@bufbuild/protoc-gen-es-1.7.1"
      sources."@bufbuild/protoplugin-1.7.1"
      sources."@connectrpc/connect-1.3.0"
      sources."@connectrpc/protoc-gen-connect-es-1.3.0"
      sources."@typescript/vfs-1.5.0"
      sources."debug-4.3.4"
      sources."ms-2.1.2"
      sources."typescript-4.5.2"
    ];
    buildInputs = globalBuildInputs;
    meta = {
      description = "NodeJS dev tools used in the development";
      license = "ISC";
    };
    production = true;
    bypassCache = true;
    reconstructLock = true;
  };
in {
  args = args;
  sources = sources;
  tarball = nodeEnv.buildNodeSourceDist args;
  package = nodeEnv.buildNodePackage args;
  shell = nodeEnv.buildNodeShell args;
  nodeDependencies = nodeEnv.buildNodeDependencies (lib.overrideExisting args {
    src = stdenv.mkDerivation {
      name = args.name + "-package-json";
      src = nix-gitignore.gitignoreSourcePure [
        "*"
        "!package.json"
        "!package-lock.json"
      ] args.src;
      dontBuild = true;
      installPhase = "mkdir -p $out; cp -r ./* $out;";
    };
  });
}
