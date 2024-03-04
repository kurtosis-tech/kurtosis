---
title: NixBuildSpec
sidebar_label: NixBuildSpec
---

You can provide Kurtosis just with source code and a Nix definition on how to build an image. Kurtosis will take care of building and deploying the image directly into the enclave without the need to upload or register the image beforehand.

For that, we use Nix flakes, which is a way to package build definitions and dependencies in a reproducible manner. Using Nix flakes, you can define your system configurations and dependencies in a single file (`flake.nix`), making it easier to manage and share.

Here's a basic explanation of how you can generate Docker images from services using Nix flakes:

1. **Install Nix**: Installing Nix isn't strictly necessary with Kurtosis, but it's recommended if you are creating or developing the package. You can install it by following the instructions on the Nix website: [https://nixos.org/download.html](https://nixos.org/download.html)

2. **Create a Nix Flake**: Go to your project root directory and initialize a Nix flake. You can do this by running:
   ```bash
   cd myproject
   nix flake init -t simple
   ```

3. **Define Your Services**: Inside the `flake.nix` file, you can define your services and their dependencies. For example:
   ```nix
   {
     description = "My project";

     inputs = {
       nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
     };

     outputs = { self, nixpkgs, myservice }: {
       defaultPackage.aarch64-darwin = nixpkgs.lib.dockerTools.buildImage {
         name = "myservice";
         tag = "latest";
         contents = [ myservice ];
         config.Cmd = [ "myservice-binary" ];
       };
     };
   }
   ```

4. **Add a service definition to Starlark**: Now just add a service to your Starlark configuration:
   ```python
    plan.add_service(
        name = "nix-example",
        config = ServiceConfig(
            image = NixBuildSpec(image_name = "myservice", flake_location_dir = ".", build_context_dir = "./"),
        ),
    )
   ```

5. **Build and Deploy with Kurtosis**: From your package folder, simply run `kurtosis run .` to get your cluster up and running.

This is just a basic example. Depending on your specific use case and requirements, you may need to adjust the configuration and dependencies in your `flake.nix` file accordingly. Additionally, you can add more services, configure networking, volumes, environment variables, etc., based on your needs.