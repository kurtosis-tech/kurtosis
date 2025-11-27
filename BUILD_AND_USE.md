# Building and Using Your Forked Kurtosis

This guide explains how to build your forked Kurtosis repository and use it with your Ethereum package.

## Prerequisites

1. **Install Nix** (recommended for dev environment):
   ```bash
   sh <(curl -L https://nixos.org/nix/install)
   ```

2. **Or install dependencies manually**:
   - Go 1.23+
   - Docker
   - Git
   - Goreleaser (for CLI builds)

## Building Kurtosis

### Option 1: Full Build (Recommended)

From the root of your forked Kurtosis repository (`/Users/gmazzeo/Dev/Trillion/PoTE-kurtosis`):

```bash
# Enter Nix development environment (if using Nix)
nix develop

# Or manually ensure you have Go, Docker, etc. installed

# Run the main build script
./scripts/build.sh
```

This will:
- Build all components (CLI, engine, core, etc.)
- Create Docker images for the engine and core
- Build the CLI binary

### Option 2: Build Just the CLI

If you only need the CLI binary:

```bash
cd cli/cli
./scripts/build.sh
```

The CLI binary will be created at:
```
cli/cli/dist/kurtosis_<os>_<arch>/kurtosis
```

### Option 3: Quick CLI Build (Development)

For faster iteration during development:

```bash
cd cli/cli
go build -o kurtosis main.go
```

This creates a `kurtosis` binary in the current directory.

## Using Your Forked Kurtosis

### Method 1: Replace the System Kurtosis Binary

1. **Find your system's Kurtosis binary location:**
   ```bash
   which kurtosis
   # Usually: /usr/local/bin/kurtosis or ~/.local/bin/kurtosis
   ```

2. **Backup the original (optional):**
   ```bash
   sudo mv $(which kurtosis) $(which kurtosis).backup
   ```

3. **Copy your built binary:**
   ```bash
   # If built with goreleaser:
   sudo cp cli/cli/dist/kurtosis_<os>_<arch>/kurtosis $(which kurtosis)
   
   # Or if built with go build:
   sudo cp cli/cli/kurtosis $(which kurtosis)
   ```

4. **Verify:**
   ```bash
   kurtosis version
   ```

### Method 2: Use Direct Path (Recommended for Development)

Instead of replacing the system binary, use the full path:

```bash
# From your Ethereum package directory
/Users/gmazzeo/Dev/Trillion/PoTE-kurtosis/cli/cli/dist/kurtosis_darwin_amd64/kurtosis run . "{}"
```

Or create an alias:

```bash
alias kurtosis-dev="/Users/gmazzeo/Dev/Trillion/PoTE-kurtosis/cli/cli/dist/kurtosis_darwin_amd64/kurtosis"
kurtosis-dev run . "{}"
```

### Method 3: Add to PATH Temporarily

```bash
export PATH="/Users/gmazzeo/Dev/Trillion/PoTE-kurtosis/cli/cli/dist/kurtosis_darwin_amd64:$PATH"
kurtosis run . "{}"
```

## Building Docker Images

Your forked Kurtosis also needs to build the engine and core Docker images:

```bash
# Build engine image
cd engine
./scripts/build.sh

# Build core (APIC) image  
cd core
./scripts/build.sh
```

These images will be tagged with your local version. The CLI will automatically use them when running packages.

## Testing Your Changes

1. **Build Kurtosis:**
   ```bash
   cd /Users/gmazzeo/Dev/Trillion/PoTE-kurtosis
   ./scripts/build.sh
   ```

2. **Test with your Ethereum package:**
   ```bash
   cd /Users/gmazzeo/Dev/Trillion/PoTE-ethereum-package
   
   # Use your forked Kurtosis
   /Users/gmazzeo/Dev/Trillion/PoTE-kurtosis/cli/cli/dist/kurtosis_darwin_amd64/kurtosis run . '{"participants": [{"cl_type": "lighthouse", "cl_devices": ["/dev/tpm0"]}]}'
   ```

## Troubleshooting

### Issue: "Cannot construct 'ServiceConfig' from the provided arguments"

**Solution:** Make sure you've built the latest version of Kurtosis after your changes:
```bash
cd /Users/gmazzeo/Dev/Trillion/PoTE-kurtosis
./scripts/build.sh
```

### Issue: Old Docker images being used

**Solution:** Rebuild the Docker images:
```bash
cd /Users/gmazzeo/Dev/Trillion/PoTE-kurtosis
cd engine && ./scripts/build.sh
cd ../core && ./scripts/build.sh
```

### Issue: CLI binary not found

**Solution:** Check the build output directory. The path depends on your OS/arch:
- macOS (Intel): `cli/cli/dist/kurtosis_darwin_amd64_v1/kurtosis`
- macOS (Apple Silicon): `cli/cli/dist/kurtosis_darwin_arm64/kurtosis`
- Linux: `cli/cli/dist/kurtosis_linux_amd64_v1/kurtosis`

### Issue: Test failures

**Solution:** Update test files that call `CreateServiceConfig` to include the new `devices` parameter:
```go
service.CreateServiceConfig(..., true, false, []string{})
//                                                      ^^^^^^^^^^ add this
```

## Development Workflow

1. **Make changes** to Kurtosis code
2. **Build Kurtosis:**
   ```bash
   cd /Users/gmazzeo/Dev/Trillion/PoTE-kurtosis
   ./scripts/build.sh
   ```
3. **Test with your package:**
   ```bash
   cd /Users/gmazzeo/Dev/Trillion/PoTE-ethereum-package
   /path/to/your/kurtosis run . "{}"
   ```
4. **Iterate** as needed

## Understanding Go Module Paths

You may notice that the code references `github.com/kurtosis-tech/kurtosis` in imports, but **this is correct and expected**. Here's why:

### How Go Replace Directives Work

The `go.mod` files use `replace` directives to point to local code:

```go
replace (
    github.com/kurtosis-tech/kurtosis/container-engine-lib => ../../container-engine-lib
    github.com/kurtosis-tech/kurtosis/core/server => ../../core/server
    // ... etc
)
```

This means:
- ✅ **Your local code IS being used** - the `replace` directives tell Go to use local filesystem paths
- ✅ **No GitHub fetch happens** - Go uses `../../container-engine-lib` instead of fetching from GitHub
- ✅ **Your changes are included** - any modifications you make are immediately used

### Verifying Local Code is Used

1. **Check the replace directives:**
   ```bash
   cd /Users/gmazzeo/Dev/Trillion/PoTE-kurtosis/cli/cli
   grep -A 10 "replace" go.mod
   ```

2. **Verify Go sees local modules:**
   ```bash
   go list -m all | grep kurtosis
   ```

3. **Test your changes:**
   - Make a small change (like adding a comment) to a file you modified
   - Rebuild: `./scripts/build.sh`
   - If the change appears in the binary, you're using local code ✅

### Why Keep the Original Module Path?

- **Compatibility**: External dependencies expect `github.com/kurtosis-tech/kurtosis`
- **Standard Practice**: Go monorepos use `replace` directives for local development
- **No Impact**: The `replace` directives override the module path for local builds

### If You Want to Change Module Paths (Not Recommended)

If you really want to change the module paths to your fork, you'd need to:
1. Update all `module` declarations in every `go.mod` file
2. Update all import statements across the entire codebase
3. Update all `replace` directives
4. This is **not necessary** and **not recommended** - the current setup works perfectly

## Notes

- The Kurtosis CLI automatically pulls and uses the engine/core Docker images
- If you're using Kubernetes backend, make sure your cluster can access the Docker images
- For local Docker backend, images are built and tagged locally
- The version is determined from `version.txt` in the Kurtosis repo root
- **The `replace` directives ensure your local code is used, not the GitHub version**

