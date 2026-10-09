#!/usr/bin/env python3
"""Validate static site metadata, local links, sitemap entries, and home JSON-LD."""
from html.parser import HTMLParser
from pathlib import Path
from urllib.parse import unquote, urlparse
from xml.etree import ElementTree
import json

ROOT = Path(__file__).resolve().parents[1] / "website"
BASE = "https://proggarapsody.github.io/bitbottle/"
PAGES = ["index.html", "install/index.html", "pull-requests/index.html", "ai-mcp/index.html"]
class Page(HTMLParser):
    def __init__(self):
        super().__init__(); self.meta={}; self.links=[]; self.ids=set(); self.jsonld=[]; self.in_json=False; self.json_buf=[]; self.in_title=False; self.title=[]
    def handle_starttag(self, tag, attrs):
        a=dict(attrs)
        if tag == "meta":
            key=a.get("name") or a.get("property")
            if key: self.meta[key]=a.get("content", "")
        if tag == "link" and a.get("rel") == "canonical": self.meta["canonical"]=a.get("href", "")
        if tag == "a" and a.get("href"): self.links.append(a["href"])
        if a.get("id"): self.ids.add(a["id"])
        if tag == "title": self.in_title=True
        if tag == "script" and a.get("type") == "application/ld+json": self.in_json=True
    def handle_endtag(self, tag):
        if tag == "title": self.in_title=False
        if tag == "script" and self.in_json:
            self.jsonld.append("".join(self.json_buf)); self.json_buf=[]; self.in_json=False
    def handle_data(self, data):
        if self.in_json: self.json_buf.append(data)
        if self.in_title: self.title.append(data.strip())
errors=[]; parsed={}; titles=[]; descriptions=[]
for name in PAGES:
    path=ROOT/name; p=Page(); p.feed(path.read_text()); parsed[name]=p
    for key in ("description", "og:title", "og:description", "og:url", "og:site_name"):
        if not p.meta.get(key): errors.append(f"{name}: missing {key}")
    if not p.meta.get("canonical", "").startswith(BASE): errors.append(f"{name}: canonical must use {BASE}")
    if not p.meta.get("og:title"): errors.append(f"{name}: missing Open Graph title")
    title="".join(p.title).strip()
    if not title: errors.append(f"{name}: missing title")
    titles.append(title)
    descriptions.append(p.meta.get("description", ""))
    if name == "index.html":
        if len(p.jsonld) != 1: errors.append("index.html: expected one JSON-LD block")
        else:
            try:
                data=json.loads(p.jsonld[0])
                if data.get("@type") != "SoftwareSourceCode" or data.get("codeRepository") != "https://github.com/proggarapsody/bitbottle": errors.append("index.html: inaccurate SoftwareSourceCode JSON-LD")
            except json.JSONDecodeError as e: errors.append(f"index.html: invalid JSON-LD: {e}")
if len(set(titles)) != len(PAGES): errors.append("page titles must be unique")
if len(set(descriptions)) != len(PAGES): errors.append("page descriptions must be unique")
for name,p in parsed.items():
    for href in p.links:
        u=urlparse(href)
        if u.scheme or u.netloc or href.startswith("mailto:"): continue
        if href.startswith("#"):
            if href[1:] not in p.ids: errors.append(f"{name}: missing anchor {href}")
            continue
        if href.startswith("/") and not href.startswith("/bitbottle/"): errors.append(f"{name}: local link escapes /bitbottle/: {href}"); continue
        if href.startswith("/bitbottle/"):
            target=unquote(href[len("/bitbottle/"):]).split("#",1)[0]
            if not target: target="index.html"
            elif target.endswith("/"): target += "index.html"
            elif not Path(target).suffix: target += "/index.html"
            if not (ROOT/target).is_file(): errors.append(f"{name}: broken local link {href}")
sitemap=ElementTree.parse(ROOT/"sitemap.xml").getroot()
urls={e.text for e in sitemap.iter() if e.tag.endswith("loc")}
expected={BASE,BASE+"install/",BASE+"pull-requests/",BASE+"ai-mcp/"}
if urls != expected: errors.append(f"sitemap URLs differ: missing={sorted(expected-urls)}, extra={sorted(urls-expected)}")
if not (ROOT/".nojekyll").exists(): errors.append("missing .nojekyll")
if errors:
    print("Website validation failed:"); print("\n".join(f"- {e}" for e in errors)); raise SystemExit(1)
print(f"Website validation passed: {len(PAGES)} pages, metadata, internal links, sitemap, and JSON-LD.")
