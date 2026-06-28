#!/usr/bin/env node
/**
 * Build docs navigation from markdown files in docs/content/.
 *
 * Each .md file may include YAML frontmatter:
 *   title, group, order, icon, description, id (optional)
 *
 * Usage: node docs/build.mjs
 * Output: docs/nav.json
 */

import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const CONTENT_DIR = path.join(__dirname, "content");
const OUT_FILE = path.join(__dirname, "nav.json");

const SITE = {
  brand: "Atrisos",
    version: "Version 2.0",
  accent: "orange",
  defaultTheme: "light",
  github: "https://github.com/sonmezerekrem/atrisos",
  // GitHub Pages project site: https://<user>.github.io/atrisos/
  basePath: "/atrisos",
};

function walkMd(dir, base = "") {
  const entries = fs.readdirSync(dir, { withFileTypes: true });
  const files = [];
  for (const ent of entries) {
    const rel = base ? `${base}/${ent.name}` : ent.name;
    const full = path.join(dir, ent.name);
    if (ent.isDirectory()) files.push(...walkMd(full, rel));
    else if (ent.name.endsWith(".md") && ent.name.toLowerCase() !== "readme.md") files.push({ rel, full });
  }
  return files;
}

function parseFrontmatter(raw) {
  const match = /^---\r?\n([\s\S]*?)\r?\n---\r?\n([\s\S]*)$/.exec(raw);
  if (!match) return { meta: {}, body: raw };
  const meta = {};
  for (const line of match[1].split("\n")) {
    const m = /^([A-Za-z0-9_-]+):\s*(.*)$/.exec(line.trim());
    if (!m) continue;
    let val = m[2].trim();
    if (
      (val.startsWith('"') && val.endsWith('"')) ||
      (val.startsWith("'") && val.endsWith("'"))
    ) {
      val = val.slice(1, -1);
    }
    if (/^\d+$/.test(val)) val = Number(val);
    meta[m[1]] = val;
  }
  return { meta, body: match[2] };
}

function titleFromBody(body) {
  const m = /^#\s+(.+)$/m.exec(body);
  return m ? m[1].trim() : "Untitled";
}

function pageId(rel, meta) {
  if (meta.id) return String(meta.id);
  return rel.replace(/\.md$/i, "").replace(/\\/g, "/");
}

function main() {
  if (!fs.existsSync(CONTENT_DIR)) {
    console.error(`content directory not found: ${CONTENT_DIR}`);
    process.exit(1);
  }

  const mdFiles = walkMd(CONTENT_DIR);
  const pages = mdFiles.map(({ rel, full }) => {
    const raw = fs.readFileSync(full, "utf8");
    const { meta, body } = parseFrontmatter(raw);
    const id = pageId(rel, meta);
    return {
      id,
      path: rel.replace(/\\/g, "/"),
      label: meta.title || titleFromBody(body),
      group: meta.group || "Documentation",
      order: typeof meta.order === "number" ? meta.order : 999,
      icon: meta.icon || "info",
      description: meta.description || "",
    };
  });

  pages.sort((a, b) => {
    if (a.group !== b.group) return a.group.localeCompare(b.group);
    if (a.order !== b.order) return a.order - b.order;
    return a.label.localeCompare(b.label);
  });

  const GROUP_ORDER = ["Introduction", "Core Concepts", "Reference", "Agents", "Documentation"];
  function groupRank(label) {
    const i = GROUP_ORDER.indexOf(label);
    return i === -1 ? GROUP_ORDER.length : i;
  }

  const groupMap = new Map();
  for (const p of pages) {
    if (!groupMap.has(p.group)) groupMap.set(p.group, []);
    groupMap.get(p.group).push({
      id: p.id,
      label: p.label,
      icon: p.icon,
      description: p.description,
      path: p.path,
    });
  }

  const groups = [...groupMap.entries()]
    .sort(([a], [b]) => groupRank(a) - groupRank(b) || a.localeCompare(b))
    .map(([label, items]) => ({ label, items }));

  const defaultPage =
    pages.find((p) => p.id === "getting-started")?.id || pages[0]?.id || "getting-started";

  const nav = {
    site: SITE,
    defaultPage,
    groups,
  };

  fs.writeFileSync(OUT_FILE, JSON.stringify(nav, null, 2) + "\n");
  console.log(`Wrote ${OUT_FILE} (${pages.length} pages, ${groups.length} groups)`);
}

main();
