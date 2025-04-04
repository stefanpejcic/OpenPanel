/**
 * `react-live` does not support `export` statements in the code.
 * This function will remove the `export` statements from the code.
 */
export const replaceExports = (code: string): string => {
  const removeExportStatements = (line: string) =>
    !line.trim().startsWith("export default");

  const updatedCode = code
    .replace(/export\s+(const|let|var|type|interface|function|class)\s+(\w+)\s*(=|:)/g, "$1 $2 =")
    .split("\n")
    .filter(removeExportStatements)
    .join("\n");

  return updatedCode;
};
