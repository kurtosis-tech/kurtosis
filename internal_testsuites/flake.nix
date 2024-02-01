{
  outputs = { self, nixpkgs, }:
    let vms = import ./vm_modules.nix { inherit nixpkgs; };
    in {
      packages.x86_64-linux.linuxVM =
        vms.nixosConfigurations.linuxVM.config.system.build.vm;

      packages.aarch64-darwin.darwinVM =
        vms.nixosConfigurations.darwinVM.config.system.build.vm;
    };
}
