const fs = require('fs');
const path = require('path');

const SOURCE_DOCS_DIR = 'docs';
const SITE_URL = 'https://docs.kurtosis.com';

function walkDocs(dir, baseDir, out) {
  for (const entry of fs.readdirSync(dir, {withFileTypes: true})) {
    const full = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      walkDocs(full, baseDir, out);
    } else if (entry.isFile() && /\.(md|mdx)$/.test(entry.name)) {
      const rel = path.relative(baseDir, full);
      const urlPath = rel.replace(/\.(md|mdx)$/, '');
      out.push({sourcePath: full, urlPath});
    }
  }
}

module.exports = function rawMarkdownPlugin(context) {
  return {
    name: 'raw-markdown-plugin',
    async postBuild({outDir, siteConfig}) {
      const siteDir = context.siteDir;
      const docsDir = path.join(siteDir, SOURCE_DOCS_DIR);
      if (!fs.existsSync(docsDir)) {
        return;
      }

      const pages = [];
      walkDocs(docsDir, docsDir, pages);

      for (const {sourcePath, urlPath} of pages) {
        const destPath = path.join(outDir, `${urlPath}.md`);
        fs.mkdirSync(path.dirname(destPath), {recursive: true});
        fs.copyFileSync(sourcePath, destPath);
      }

      const baseUrl = (siteConfig && siteConfig.url ? siteConfig.url : SITE_URL).replace(/\/$/, '');
      const llmsLines = [
        '# Kurtosis Docs',
        '',
        '> Raw markdown source of every page on docs.kurtosis.com. Append `.md` to any docs URL to fetch the source.',
        '',
      ];
      for (const {urlPath} of pages.slice().sort((a, b) => a.urlPath.localeCompare(b.urlPath))) {
        llmsLines.push(`- [${urlPath}](${baseUrl}/${urlPath}.md)`);
      }
      fs.writeFileSync(path.join(outDir, 'llms.txt'), llmsLines.join('\n') + '\n');
    },
  };
};
