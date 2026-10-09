# Search and discovery plan

Baseline recorded 2026-10-09. This plan improves accurate discovery of the Go Bitbucket CLI and measures whether the project site becomes visible for relevant searches. Search engines decide what to crawl, index, rank, and show; none of the steps below guarantees those outcomes.

## Verified baseline

- The intended site is the GitHub Pages project URL: <https://proggarapsody.github.io/bitbottle/>.
- The GitHub repository description is “Go CLI for Bitbucket Cloud and Self-hosted Bitbucket Server”; its homepage field was empty. Topics were `bitbucket`, `bitbucket-cli`, `bitbucket-server`, `cli`, `devtools`, `mcp`, and `self-hosted`. GitHub Pages was not enabled at baseline.
- A web search for `bitbottle` surfaced the npm package, Go package, and skills sites, while an unrelated Rust archive-format project with the same name dominated the results observed. This is a search-tool observation only: it does not establish Google ranking, complete indexing, or what other people see by location or time.
- This repository's README already distinguishes Cloud-only and Server/DC-only commands. Discovery copy must preserve those distinctions and must not imply feature parity for every command.

## Exact query basket

Record the date, search engine, country/language, device, query, result URL, and whether the official repository or project site appears. Use the same queries and settings at each checkpoint:

1. `bitbottle`
2. `bitbottle bitbucket`
3. `Bitbucket CLI`
4. `Bitbucket Cloud CLI`
5. `Bitbucket Data Center CLI`
6. `Bitbucket Server CLI`
7. `Bitbucket pull request CLI`
8. `Bitbucket MCP server`

Do not collapse these into one “rank.” Report observed positions and URLs per query, with “not found in first 100 results” when appropriate. Search results are location-, device-, and time-sensitive.

## Changes in this rollout

- The README title and opening sentence identify this as an independent, gh-style CLI for Cloud and self-hosted Server/Data Center. A reader-facing link near the top points to the project site; the maintainer search and discovery guide is linked under Contributing.
- The README FAQ explains that the tool is standalone, points out host-specific capabilities, and distinguishes this Go project from the unrelated Rust archive-format project.
- The npm package homepage points to the project site; keywords describe Bitbucket Cloud, Data Center, pull requests, CLI, and MCP without repeating variants.
- The README Cloud login example uses an Atlassian account email and an API token supplied through standard input. It links to Atlassian's token instructions. The README says host settings live in `hosts.yml` and tokens are stored in the OS keyring.

## Publish and indexing checks

After publishing the project site:

1. Open the site and every linked documentation page over HTTPS. Check that page titles, headings, links, and core copy are readable without executing client-side scripts. Confirm the README and site use the same product description and feature caveats.
2. Check the published sitemap URL, if the site build provides one, and ensure its URLs use the canonical `https://proggarapsody.github.io/bitbottle/` prefix. A sitemap helps discovery but does not guarantee indexing. Follow [Google's sitemap guidance](https://developers.google.com/search/docs/crawling-indexing/sitemaps/build-sitemap).
3. In Google Search Console, create and verify the URL-prefix property `https://proggarapsody.github.io/bitbottle/` using a verification method that works with a GitHub Pages project site. Submit the sitemap URL if one is published. Use URL Inspection to check the homepage and the most useful documentation page; record indexed status and any reported issue. Do not claim verification or indexing until the console shows it.
4. In Bing Webmaster Tools, import the verified Search Console property if offered, or verify the URL-prefix site and submit its sitemap there. Record the result shown by the tool. Submission is a discovery request, not a promise of crawling or ranking.
5. A GitHub Pages project site is under the shared `github.io` origin. A `robots.txt` placed only at `/bitbottle/robots.txt` is not the origin-level `/robots.txt`, so it cannot set origin-wide crawler policy. Do not claim project-level control of that file; confirm access and crawl behavior in each webmaster console.

Google says ordinary search best practices apply to its AI features, with no special AI markup or separate machine-readable file required; pages must be indexed and eligible for snippets, and inclusion is not guaranteed. Treat citations in AI answers as a separate qualitative observation, not a substitute for search indexing or ranking measurements. See [Google's guidance on AI features](https://developers.google.com/search/docs/appearance/ai-features). If referring to OpenAI crawler access, use the current [OpenAI bot documentation](https://platform.openai.com/docs/bots); crawler access alone does not guarantee inclusion in an answer.

Do not use Google's Indexing API as a general-purpose submission shortcut; use Search Console sitemap submission and URL Inspection for this site. Do not describe any indexing request as an indexing ping or guarantee.

## Measurement checkpoints

Save a dated baseline after deployment, then repeat the exact query basket at day 7 and day 28. Keep the engine, region/language, device, and method consistent. Capture the result page or a source export so observations can be checked later.

Report separately:

- **Indexing:** Search Console and Bing Webmaster status for the homepage, sitemap, and selected documentation URLs.
- **Search visibility:** observed result position and URL for every query in the basket. When available, also record Search Console impressions, clicks, and average position by query and page for the period; these are aggregate measurements and may not match a manual search.
- **AI answer visibility:** whether the project is cited in a small, dated set of relevant questions, with the question, product, date, and citation URL. Label this anecdotal because responses vary and no stable ranking measure is implied.

At day 7, report early crawl/indexing signals without calling them a ranking trend. At day 28, compare the same measurements with baseline and state missing data. Do not claim improvement where the evidence is only a changed result page or a single AI answer.

## Ongoing off-site discovery

Maintain accurate repository metadata and package links. Where maintainers approve, contribute factual corrections or comparisons to relevant Bitbucket CLI directories and comparison pages. Draft an optional factual announcement for maintainers to review; do not represent outreach or publication as completed until it is actually approved and posted. Avoid keyword-stuffed submissions, duplicate pages, and unverified claims about rankings, adoption, or feature parity.
