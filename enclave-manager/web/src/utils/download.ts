const DEFAULT_FILE_TYPE = "text/plain";

export const saveTextAsFile = (
  text: string,
  fileName: string,
  options: { elementName?: string; fileType?: string } = {},
) => {
  const fileType = options.fileType || DEFAULT_FILE_TYPE;

  const blob = new Blob([text], { type: fileType });

  const a = document.createElement("a");

  a.href = URL.createObjectURL(blob);
  a.download = fileName;
  a.click();

  URL.revokeObjectURL(a.href);
};
