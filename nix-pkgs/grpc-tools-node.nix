{ pkgs ? import <nixpkgs> { } }:
with pkgs;
stdenv.mkDerivation {
  name = "grpc-tools-node";
  version = "0.1";
  phases = [ "installPhase" ];
  # For some reason, Node gRPC has its own 'protoc' binary
  installPhase = ''
    mkdir -p $out/bin
    cp -r ${protobuf}/include/ $out/bin/
    cp "${grpc-tools}/bin/protoc" $out/bin/grpc_tools_node_protoc
    cp "${grpc-tools}/bin/grpc_node_plugin" $out/bin/grpc_tools_node_protoc_plugin
  '';
}
