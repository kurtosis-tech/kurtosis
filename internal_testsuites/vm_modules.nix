{ nixpkgs, ... }:
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

    # Enable passwordless ‘sudo’ for the "tester" user
    users.users.tester.extraGroups = [ "wheel" ];
    security.sudo.wheelNeedsPassword = false;
  };

  nixosModulesK3s = { pkgs, ... }: {
    # setup k3s
    services.k3s.enable = true;
    services.k3s.role = "server";
    environment.systemPackages = [ pkgs.k3s ];
  };

  nixosModulesDocker = { pkgs, ... }: {
    # setup docker
    virtualisation.docker.enable = true;
    users.users.tester.extraGroups = [ "docker" ];
  };

in rec {

  nixosConfigurations.x86_64.modules =
    [ nixosModulesBase nixosModulesK3s nixosModulesDocker nixosModulesVM ];

  nixosConfigurations.aarch64.modules = [
    nixosModulesBase
    nixosModulesK3s
    nixosModulesDocker
    nixosModulesVM
    {
      virtualisation.vmVariant.virtualisation.host.pkgs =
        nixpkgs.legacyPackages.aarch64-darwin;
    }
  ];

  nixosConfigurations.x86_64.VM = nixpkgs.lib.nixosSystem {
    system = "x86_64-linux";
    modules = nixosConfigurations.x86_64.modules;
  };

  nixosConfigurations.aarch64.VM = nixpkgs.lib.nixosSystem {
    system = "aarch64-linux";
    modules = nixosConfigurations.aarch64.modules;
  };

}
