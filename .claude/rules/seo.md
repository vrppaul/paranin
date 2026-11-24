---
paths:
  - "docs/SEO.md"
  - "internal/document/**"
  - "internal/compose/**"
  - "internal/face/html/**"
  - "internal/face/feed/**"
  - "internal/face/text/**"
  - "internal/hosting/web/**"
  - "internal/routes/**"
---

# SEO

What a search engine, an AI crawler and a link preview read of the site is `docs/SEO.md`, the owner's model: a change
to it is his yes first, marked there (`.claude/rules/prose.md`), then code.

- Whether a page is indexed is worked out once, where the page is composed; the `robots` line, `X-Robots-Tag`, the
  canonical address, the record and the sitemap each read that, never a rule of their own.
- A page not indexed sends `noindex` in every format and carries no canonical address or record, and stands in no
  sitemap; `robots.txt` never disallows it, or no crawler reads its `noindex`.
- A page indexed names one address, the one `routes` writes, whole: its canonical link, its `Link` header, `og:url`, its
  record's `url` and its sitemap entry.
- Every page indexed has a title and a description of its own; no two pages indexed share a title.
- A day a crawler reads, `lastmod`, `dateModified` or `article:modified_time`, is the day its text changed, from the
  content: never the build's, the load's or the clock's.
- A new kind of page or a new address says, in its chunk, whether it is indexed, its sitemap entry, its record and its
  languages.
- One function chooses a page's format (D21). No crawler is named in code: a terminal's client alone is told apart by
  its agent, from one list.
- A change to what a crawler reads is checked as `docs/SEO.md`'s "How it is checked" says, in the report, never in a
  test of how it looks.
