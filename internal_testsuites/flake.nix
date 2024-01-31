{
  outputs = { self, nixpkgs, }: {
    nixosModules.vm = { ... }: {
      # Make VM output to the terminal instead of a separate window
      virtualisation.vmVariant.virtualisation.graphics = false;
    };

    nixosModules.base = { pkgs, ... }: {
      system.stateVersion = "23.11";

      # Configure networking
      networking.useDHCP = false;
      networking.interfaces.eth0.useDHCP = true;

      # Create user "test"
      services.getty.autologinUser = "tester";
      users.users.test.isNormalUser = true;

      # Enable passwordless ‘sudo’ for the "test" user
      users.users.test.extraGroups = [ "wheel" ];
      security.sudo.wheelNeedsPassword = false;
    };

    nixosConfigurations.linuxVM = nixpkgs.lib.nixosSystem {
      system = "x86_64-linux";
      modules = [ self.nixosModules.base self.nixosModules.vm ];
    };
    packages.x86_64-linux.linuxVM =
      self.nixosConfigurations.linuxVM.config.system.build.vm;

    nixosConfigurations.darwinVM = nixpkgs.lib.nixosSystem {
      system = "aarch64-linux";
      modules = [
        self.nixosModules.base
        self.nixosModules.vm
        {
          virtualisation.vmVariant.virtualisation.host.pkgs =
            nixpkgs.legacyPackages.aarch64-darwin;
        }
      ];
    };
    packages.aarch64-darwin.darwinVM =
      self.nixosConfigurations.darwinVM.config.system.build.vm;

  };
}
