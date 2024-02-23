# NodeJS devtools

This Nix derivation get all nodeJS needed in the dev environment.

## Adding/Updating packages

On this same folder:
```bash
npm install <package>@<version>
```
This will update both `package.json` and `package-lock.json` files.

Then run this command to regenerate to Nix definitions.
```bash
node2nix
```