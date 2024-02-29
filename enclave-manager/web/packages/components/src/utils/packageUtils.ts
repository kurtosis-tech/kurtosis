export function parsePackageUrl(packageUrl: string) {
  const components = packageUrl
    .replace(/https?:\/\//, "")
    .replace(/(?<=github\.com\/[^/]+\/[^/]+)\/tree\/([^/]+)/, "")
    .split("/");
  if (components.length < 3) {
    throw Error(`Illegal url, invalid number of components: ${packageUrl}`);
  }
  if (components[1].length < 1 || components[2].length < 1) {
    throw Error(`Illegal url, empty components: ${packageUrl}`);
  }

  const branchMatches = packageUrl.match(/(?<=github\.com\/[^/]+\/[^/]+)\/tree\/([^/]+)/);
  const defaultBranch = branchMatches ? branchMatches[1] : undefined;

  return {
    baseUrl: "github.com",
    owner: components[1],
    name: components[2],
    rootPath: components.filter((v, i) => i > 2 && v.length > 0).join("/") + "/",
    defaultBranch,
  };
}
