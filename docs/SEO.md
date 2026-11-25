# SEO

What a search engine, an AI crawler and a link preview read of the site, and why. D27 asks for good SEO and Open
Graph on every page (`docs/FOUNDATION.md`); this is its model (owner, 2026-11-24). The addresses and the headers each
answer carries are [WEB](WEB.md); this says what they tell a crawler. Each decision says whether the owner gave it his
yes (`.claude/rules/prose.md`). What of it stands and what is still to build is its last section.

## Who reads the site

- **Search engines.** Google indexes with its phone crawler, which asks for HTML first; its desktop crawler, Bing
  (whose index DuckDuckGo and Yahoo show) and others send `Accept: */*`.
- **AI crawlers.** GPTBot and ClaudeBot send `*/*`; PerplexityBot sends no `Accept` at all.
- **Link previews.** LinkedIn sends `*/*`; what X, Telegram, Slack and Facebook send is measured nowhere, and a preview
  reads Open Graph from HTML alone.

The headers are those [EdgeComet](https://edgecomet.com/blog/request-headers-what-googlebot-and-ai-bots-send-with-every-request/)
recorded in September 2026. Since most crawlers name no format, the format a page goes in decides what they see.

## Formats (D21)

A page's own address chooses its format by the `Accept` header's ranking: HTML, markdown or the manual.

- Where the header names none of them but by a wildcard, or is absent, the page goes as HTML, but to a terminal's
  HTTP client, known by its `User-Agent`'s product (curl, Wget, HTTPie, xh), which gets the manual, so
  `curl paran.in` gets the manual (owner, 2026-11-24).
- Every negotiated answer sends `Vary: Accept, User-Agent`.

The rest of the rule is [WEB](WEB.md), "Negotiation".

Why: a crawler or a preview that gets the manual reads no title, description, Open Graph or record, so Bing would
index bare text and LinkedIn would show no preview. Telling terminals apart, as
[wttr.in](https://github.com/chubin/wttr.in) and [ifconfig.co](https://github.com/mpolden/echoip) do, keeps a short
list of clients that rarely changes, where telling crawlers apart would keep a long one that does. Ways turned down:
HTML unless `text/plain` is named, which gives a bare `curl` the HTML; the manual for `*/*` with a list of crawlers sent
HTML, which every crawler not on it would miss. A client whose agent was changed to a browser's gets the HTML.

The first version guessed every request's format from its agent and so split a cache by every browser's version
(`docs/FOUNDATION.md` §3.1, 11). Here the agent is read only where `Accept` names no format, and the split costs
nothing: every negotiated answer is `no-cache`, and no CDN caches an address whose format `Accept` chooses
(`docs/REQUIREMENTS.md`, Cloudflare).

## What every page indexed says

- **A title of its own**, no two alike: the résumé its person's name; a tab its title and the person's name,
  `Experience — Pavel Paranin`; the list by series its own, `blog · series`, beside the list by year's `blog`, each in
  its language's words (owner, 2026-11-24); a project and a post their titles; a project's tab, whose title another
  project's may share, its title and the project's name, `How it works — semantic-code-mcp` (the author's choice, on
  the owner's leave, 2026-11-24).
- **A description:** the check requires one of a post.
- **Its markup:** one `h1`, `<article>`, `<time datetime>`, alt text required of every picture.
- **One address**, whole, as `routes` writes it: the canonical link in the head, the `Link` header every format sends
  (WEB, "Negotiation"), `og:url`, the record's `url` and the sitemap's entry.
- **Its languages:** a post read in more than one, `hreflang` to each, itself among them, and `x-default` to the
  English original. The blog's lists name no other: each lists the posts out in its language alone, so none is
  another's translation.
- **Open Graph:** `og:type`, `og:site_name` the site's name (Records), `og:locale` from its catalogue, `og:locale:alternate` for a post's other
  languages, `og:title`, `og:description`, `og:url`. A post is an `article` (owner, 2026-11-24), with
  `article:published_time`, `article:modified_time` its current version's, each written as its record's (Records),
  `article:author` the résumé's address and an `article:tag` for each tag ([ogp.me](https://ogp.me/#type_article));
  every other page a `website`.
- **X:** `twitter:card`, `summary`, `summary_large_image` once the page has a card (owner, 2026-11-24); X reads the
  rest from Open Graph ([X](https://developer.x.com/cards/markup)).
- **A card**, `og:image`, 1200×630, the page's own, a tab's its page's, which the build draws with Typst (ADR-0005)
  as `DESIGN.md`'s Social cards has it (owner, 2026-11-25), with its size and the page's title as its `alt` (owner,
  2026-11-25), and
  `max-image-preview:large` so Google may show it large. LinkedIn wants
  [at least 1200×627](https://www.linkedin.com/help/linkedin/answer/46687), and a post's record takes it as its
  `image`. Every page that shows a preview has one: an unlisted post among them. The cards come before hosting, as
  everything the site serves does (owner, 2026-11-24).
- No `fediverse:creator` (owner, 2026-11-24): the posts go where D13 names, and Mastodon is not among them.

## Records

Each page indexed carries its record as JSON-LD, in schema.org's words, through typed structs (`.claude/rules/go.md`).
One `Person` stands for the author, by the `@id` `https://paran.in/#person`, so every record names the same one.

- **The résumé, `/`:** `ProfilePage` about the person; beside it `WebSite`, its `name` the name Google shows for the
  site ([Google](https://developers.google.com/search/docs/appearance/site-names)): the person's, as the profile
  writes it, its host, `paran.in`, as its `alternateName` (owner, 2026-11-24).
- **A tab of the résumé:** `WebPage` about the person: the profile is `/` alone.
- **A project's page:** `SoftwareSourceCode`, its author the person.
- **A post:** `BlogPosting` with `headline`, `description`, `url`, `datePublished`, `dateModified` its current
  version's, its author the person with the résumé's address as `url`, `inLanguage`, `keywords` from its tags,
  `isPartOf` its series, and its card as `image` once cards come (owner, 2026-11-24). Each moment is written whole
  with its zone, as Google reads one; a version's stands at the start of its day in the post's zone, never before the
  post came out. Google requires none of them and recommends the author's `url` and both dates
  ([Google](https://developers.google.com/search/docs/appearance/structured-data/article)).
- **The lists:** none.
- **No `BreadcrumbList`** (owner, 2026-11-24): Google shows a page's crumbs on a desktop alone since
  [January 2025](https://developers.google.com/search/blog/2025/01/simplifying-breadcrumbs), and the site is two
  levels deep.

## What is indexed

Indexed: the résumé and its tabs, the projects' pages and theirs, the blog's lists by year and by series and each post
published and out, in each language. Not indexed: a filtered list, an older version of a post and an unlisted post
(owner, 2026-11-24), each sent with `noindex` in every format, with no canonical address or record, and out of the
sitemap. A filtered list and an older version carry no Open Graph either, since they have no address of their own to
show; an unlisted post keeps its preview, since it is read by whoever is sent its link. A draft is served to its author
alone. Whether a page is indexed is worked out once, where its page is composed, as `document.Head`'s `Indexed`: from
its own address (`document.Page`'s `OwnAddress`) and whether a post is listed. Every place that says it reads that.

## Crawling

- **`robots.txt`** lets every crawler in, those that train AI models among them (owner, 2026-11-24): a model that read
  the site knows its author. It names the sitemap, and `llms.txt` in a comment, since no page links it and no `rel`
  names it (owner, 2026-11-23). It disallows nothing, so a crawler reaches the `noindex` of every page that has one.

  ```
  # The site for a language model: https://paran.in/llms.txt
  User-agent: *
  Allow: /

  Sitemap: https://paran.in/sitemap.xml
  ```

- **`sitemap.xml`** is one file: every page indexed, by its address, with its `lastmod` where the text has a day
  (owner, 2026-11-24): a post's is its current version's day; a list's the latest of those among the posts it names,
  since a post's new version changes the title and lead the list shows. The résumé and the
  projects have none: a wrong day is worse than none, and Google reads `lastmod` only where it is
  [true](https://developers.google.com/search/blog/2023/06/sitemaps-lastmod-ping). No `priority` or `changefreq`,
  which Google reads not at all, and no `hreflang`, which the pages carry and Yandex reads from none.
- **The feeds** stand in for a sitemap of new posts: Google takes an Atom feed as one
  ([Google](https://developers.google.com/search/docs/crawling-indexing/sitemaps/build-sitemap)).
- **`llms.txt`** stays for agents that read it, coding agents among them; Google says its Search reads none.
- **IndexNow** names each post to Bing, Yandex, Seznam and Naver as it comes out, which Google does not read, by a
  key file at the site's root (owner, 2026-11-24): built before hosting, it sends once the site is served with its key
  (`PARANIN_INDEXNOW_KEY`). It sends what changed since it last looked, as a reload takes or a post dated later comes
  out; a start sends nothing, so a deploy announces nothing, and a post that came out while the site was down reaches
  a search engine by the sitemap alone.

## Addresses

- `/blog/<number>` answers `301` to `/blog/<slug>`.
- A page's path with a slash at its end, `/blog/`, answers `301` to the path without it (owner, 2026-11-24); a path
  that is no page with one is `404`, as without. Google counts the two as two addresses
  ([Google](https://developers.google.com/search/blog/2010/04/to-slash-or-not-to-slash)).

## At hosting

Everything above is built before hosting (owner, 2026-11-24). Hosting does only what needs the site served at its
address: the site verified in Google's and Bing's consoles by a DNS record, the sitemap and the feeds submitted
there, and the previews looked at in LinkedIn's Post Inspector.

## How it is checked

- The head of each kind of page, read off the server with a browser's `Accept`, with `*/*`, with none, and with curl's
  agent.
- A record, pasted into [the Schema Markup Validator](https://validator.schema.org) and
  [the Rich Results Test](https://search.google.com/test/rich-results).
- A preview, once the site is served, in [LinkedIn's Post Inspector](https://www.linkedin.com/post-inspector/).
- Each kind of page by Lighthouse's SEO audit, against the server run locally.
- Where it is at hand, a second look by [Front-End Checklist](https://frontendchecklist.io/rules/seo)'s MCP (owner,
  2026-11-24), `review_code` over a golden page: a finding is weighed against this model, since some are not faults
  here (a page with no `robots` line is indexed, as it should be).

## What stands and what is to build

Stands: the formats of D21; each page's title, a tab's and the list by series' among them, its description and
canonical address; Open Graph, a post as an `article`, with the site's name, and `twitter:card` `summary`; a post's
languages; every record naming the one `Person`, the résumé's `ProfilePage` beside `WebSite`, a tab's `WebPage`, a
project's `SoftwareSourceCode` and a post's `BlogPosting`; `noindex` on a filtered list, an older version and an
unlisted post; `/blog/<number>`'s `301`, and that of a path with a slash at its end; `robots.txt` and `sitemap.xml`;
IndexNow and its key file; the feeds; `llms.txt`.

To build: the cards, as the SEO item of `TODO.md` lists them; "At hosting", with hosting.
