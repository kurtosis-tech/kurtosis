import { capitalize } from "../utils";

export function readablePackageName(packageName: string): string {
  const parts = packageName.replaceAll("-", " ").split("/");
  if (parts.length < 3) {
    return packageName;
  }
  return capitalize(`${parts[2]} ${parts.slice(3).join(" ")}`);
}
