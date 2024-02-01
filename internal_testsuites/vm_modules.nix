{ nixpkgs, engine_image, ... }:
let
  nixosModulesVM = { ... }: {
    # Make VM output to the terminal instead of a separate window
    virtualisation.vmVariant.virtualisation.graphics = false;
  };

  nixosModulesBase = { pkgs, ... }: {
    system.stateVersion = "23.11";

    # Configure networking
    networking.useDHCP = false;
    networking.interfaces.eth0.useDHCP = true;

    # Create user "tester"
    services.getty.autologinUser = "tester";
    users.users.tester.isNormalUser = true;

    # setup k3s
    services.k3s.enable = true;
    services.k3s.role = "server";
    environment.systemPackages = [ pkgs.k3s ];

    # setup docker
    virtualisation.docker.enable = true;

    # Enable passwordless ‘sudo’ for the "tester" user
    users.users.tester.extraGroups = [ "wheel" "docker" ];
    security.sudo.wheelNeedsPassword = false;
  };

in {
  nixosConfigurations.linuxVM = nixpkgs.lib.nixosSystem {
    system = "x86_64-linux";
    modules = [ nixosModulesBase nixosModulesVM ];
  };

  nixosConfigurations.darwinVM = nixpkgs.lib.nixosSystem {
    system = "aarch64-linux";
    modules = [
      nixosModulesBase
      nixosModulesVM
      {
        virtualisation.vmVariant.virtualisation.host.pkgs =
          nixpkgs.legacyPackages.aarch64-darwin;
      }
    ];
  };

}
