const DEFAULT_FILE_TYPE = "text/plain";

export const saveTextAsFile = (
  text: string | Blob,
  fileName: string,
  options: { elementName?: string; fileType?: string } = {},
) => {
  const fileType = options.fileType || DEFAULT_FILE_TYPE;

  const blob = typeof text === "string" ? new Blob([text], { type: fileType }) : text;

  const a = document.createElement("a");

  a.href = URL.createObjectURL(blob);
  a.download = fileName;
  a.click();

  URL.revokeObjectURL(a.href);
};
