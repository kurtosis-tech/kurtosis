export function validateDockerLocator(value?: string) {
  if (typeof value !== "string") {
    return "Value should be a string";
  }
  if (value === "") {
    return;
  }

  if (
    !value.match(
      /^(?<repository>[\w.\-_]+((?::\d+|)(?=\/[a-z0-9._-]+\/[a-z0-9._-]+))|)(?:\/|)(?<image>[a-z0-9.\-_]+(?:\/[a-z0-9.\-_]+|))(:(?<tag>[\w.\-_]{1,127})|)$/gim,
    )
  ) {
    return "Value does not look like a docker image";
  }
}

export function validateDurationString(value?: string) {
  if (typeof value !== "string") {
    return "Value should be a string";
  }
  if (value === "") {
    return;
  }

  if (!value.match(/^\d+[msd]?$/)) {
    return "Value should be a custom wait duration with like '10s' or '3m'.";
  }
}
