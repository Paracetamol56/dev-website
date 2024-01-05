
type Page = {
  title: string;
};

type Link = {
  source: string;
  target: string;
};

function extractLinks(html: string): Page[] {
  const pages: Page[] = [];
  // Wikipedia link regex (eg. <a rel="mw:WikiLink" href="./Cookbook" title="Cookbook" id="mw_w">cook-book</a>)
  const linkRegex = /<a rel="mw:WikiLink" href="\.\/([^"]+)" title="([^"]+)"[^>]*>[^<]+<\/a>/g;
  let match: RegExpExecArray | null;
  while ((match = linkRegex.exec(html)) !== null) {
    // extract the page title from the link
    const [, target, title] = match;
    // add the link to the list
    pages.push({ title });
  }

  return pages;
}

export type { Page, Link };
export { extractLinks };