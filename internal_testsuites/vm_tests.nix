{ nixpkgs, pkgs, ... }:
let nixos-lib = import (nixpkgs + "/nixos/lib") { };
in nixos-lib.runTest rec {
  name = "demo-test";

  hostPkgs = import nixpkgs { system = pkgs.stdenv.hostPlatform.system; };

  node = pkgs.lib.optionalAttrs pkgs.stdenv.isDarwin {
    pkgs = import nixpkgs {
      system = builtins.replaceStrings [ "darwin" ] [ "linux" ]
        pkgs.stdenv.hostPlatform.system;
    };
  };

  nodes.machine = { config, pkgs, ... }: {
    users.users.alice = {
      isNormalUser = true;
      extraGroups = [ "wheel" ];
      packages = with pkgs; [ tree ];
    };
    system.stateVersion = "23.11";
    virtualisation.vmVariant.virtualisation.graphics = false;
  };

  testScript = ''
    machine.wait_for_unit("default.target")
    machine.succeed("su -- alice -c 'which firefox'")
    machine.fail("su -- root -c 'which firefox'")
  '';
}
