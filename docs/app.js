"use strict";
(function () {
  const { createElement: E, Component, createRef } = React;
  const MONO = "'JetBrains Mono', ui-monospace, monospace";

  class DocsApp extends Component {
    constructor(props) {
      super(props);
      this.searchRef = createRef();
      let theme = "light";
      try {
        const s = localStorage.getItem("atrisos-docs-theme");
        if (s) theme = s;
      } catch (_) {}
      this.state = {
        theme,
        nav: null,
        section: "",
        current: "",
        pages: {},
        search: "",
        showResults: false,
        active: "",
      };
    }

    componentDidMount() {
      this._onHash = () => {
        const h = (location.hash || "").replace("#", "").replace(/^\//, "");
        if (h && this._allIds().includes(h) && h !== this.state.current) {
          const section = this._sectionForPage(h);
          this.setState({ current: h, section: section || this.state.section }, () => {
            window.scrollTo(0, 0);
            this._ensurePage(h);
            setTimeout(() => this._spy(), 60);
          });
        }
      };
      window.addEventListener("hashchange", this._onHash);
      this._onScroll = () => {
        if (this._raf) return;
        this._raf = requestAnimationFrame(() => {
          this._raf = null;
          this._spy();
        });
      };
      window.addEventListener("scroll", this._onScroll, { passive: true });
      this._onKey = (e) => {
        if ((e.metaKey || e.ctrlKey) && e.key === "k") {
          e.preventDefault();
          this.searchRef.current && this.searchRef.current.focus();
        }
      };
      window.addEventListener("keydown", this._onKey);
      this._boot();
    }

    componentWillUnmount() {
      window.removeEventListener("hashchange", this._onHash);
      window.removeEventListener("scroll", this._onScroll);
      window.removeEventListener("keydown", this._onKey);
    }

    async _boot() {
      try {
        const res = await fetch("./nav.json");
        if (!res.ok) throw new Error("nav.json not found");
        const nav = await res.json();
        const valid = this._allIdsFromNav(nav);
        const fromHash = (location.hash || "").replace("#", "").replace(/^\//, "");
        const current = valid.includes(fromHash) ? fromHash : nav.defaultPage;
        const section = this._sectionForPage(current, nav) || nav.defaultSection || "overview";
        this.setState({ nav, current, section }, () => {
          if (!fromHash || !valid.includes(fromHash)) location.hash = current;
          this._ensurePage(current);
          setTimeout(() => this._spy(), 60);
        });
      } catch (e) {
        console.error(e);
      }
    }

    _allIdsFromNav(nav) {
      const ids = [];
      (nav.sections || []).forEach((sec) => sec.groups.forEach((g) => g.items.forEach((i) => ids.push(i.id))));
      return ids;
    }

    _sectionForPage(id, nav) {
      nav = nav || this.state.nav;
      if (!nav) return null;
      for (const sec of nav.sections || [])
        for (const g of sec.groups)
          for (const i of g.items) if (i.id === id) return sec.id;
      return null;
    }

    _currentSection() {
      return this.state.section || this._sectionForPage(this.state.current) || this.state.nav?.defaultSection || "overview";
    }

    _nav() {
      if (!this.state.nav) return [];
      const sec = (this.state.nav.sections || []).find((s) => s.id === this._currentSection());
      return (sec?.groups || []).map((g) => ({
        label: g.label,
        items: g.items.map((it) => ({
          id: it.id,
          label: it.label,
          icon: it.icon || "info",
          path: it.path,
          description: it.description,
        })),
      }));
    }

    _allIds() {
      return this._allIdsFromNav(this.state.nav || { sections: [] });
    }

    _find(id) {
      for (const sec of this.state.nav?.sections || [])
        for (const g of sec.groups)
          for (const i of g.items) if (i.id === id) return { item: i, group: g.label, section: sec.id, sectionLabel: sec.label };
      return null;
    }

    _go(id) {
      location.hash = id;
    }

    _parseFrontmatter(raw) {
      const m = /^---\r?\n([\s\S]*?)\r?\n---\r?\n([\s\S]*)$/.exec(raw);
      if (!m) return { meta: {}, body: raw };
      const meta = {};
      m[1].split("\n").forEach((line) => {
        const p = /^([A-Za-z0-9_-]+):\s*(.*)$/.exec(line.trim());
        if (!p) return;
        let v = p[2].trim();
        if ((v.startsWith('"') && v.endsWith('"')) || (v.startsWith("'") && v.endsWith("'"))) v = v.slice(1, -1);
        meta[p[1]] = v;
      });
      return { meta, body: m[2] };
    }

    _slug(text) {
      return String(text).toLowerCase().replace(/[^\w\s-]/g, "").trim().replace(/\s+/g, "-");
    }

    _inline(text) {
      if (!text) return "";
      return String(text)
        .replace(/\*\*(.+?)\*\*/g, "$1")
        .replace(/\*(.+?)\*/g, "$1")
        .replace(/`(.+?)`/g, "$1")
        .replace(/\[(.+?)\]\(.+?\)/g, "$1");
    }

    _renderInline(tokens) {
      if (!tokens || !tokens.length) return "";
      return tokens.map((t, i) => {
        if (t.type === "text") return t.raw || t.text || "";
        if (t.type === "strong") return E("strong", { key: i }, this._renderInline(t.tokens));
        if (t.type === "em") return E("em", { key: i }, this._renderInline(t.tokens));
        if (t.type === "codespan") return E("code", { key: i, style: { fontFamily: MONO, fontSize: "0.9em", background: "var(--pill)", padding: "1px 5px", borderRadius: 4 } }, t.text);
        if (t.type === "link") return E("a", { key: i, href: t.href, style: { color: "var(--accent)" } }, this._renderInline(t.tokens));
        return t.raw || t.text || "";
      });
    }

    _paragraphNode(tok) {
      const children = tok.tokens ? this._renderInline(tok.tokens) : this._inline(tok.text);
      return E("p", { style: { fontSize: "16.5px", lineHeight: 1.72, color: "var(--muted)", margin: "0 0 18px", maxWidth: 760 } }, children);
    }

    _markdownToBlocks(body) {
      if (!window.marked) return [{ isP: true, text: "Markdown renderer failed to load." }];
      const blocks = [];
      const counts = {};
      const slugify = (t) => {
        const base = this._slug(t);
        counts[base] = (counts[base] || 0) + 1;
        return counts[base] > 1 ? base + "-" + counts[base] : base;
      };
      let codeIdx = 0;
      for (const tok of marked.lexer(body)) {
        if (tok.type === "heading") {
          if (tok.depth === 1) continue;
          const text = this._inline(tok.text);
          const id = slugify(text);
          if (tok.depth === 2) blocks.push({ isH2: true, text, id });
          else if (tok.depth === 3) blocks.push({ isH3: true, text, id });
          else blocks.push({ isNode: true, node: this._headingNode(Math.min(tok.depth, 4), text, id) });
        } else if (tok.type === "paragraph") {
          blocks.push({ isNode: true, node: this._paragraphNode(tok) });
        } else if (tok.type === "code") {
          const lines = tok.text.split("\n");
          const label = tok.lang || "snippet-" + ++codeIdx;
          blocks.push({ isNode: true, node: this.codeCard(label, lines, { tall: lines.length > 30 || tok.text.length > 1500 }) });
        } else if (tok.type === "blockquote") {
          const inner = (tok.tokens || []).filter((t) => t.type === "paragraph").map((t) => this._paragraphNode(t));
          if (inner.length) blocks.push({ isNode: true, node: E("div", { style: { margin: "0 0 18px", maxWidth: 760 } }, inner) });
        } else if (tok.type === "list") {
          blocks.push({ isNode: true, node: this._listNode(tok) });
        } else if (tok.type === "table") {
          blocks.push({ isNode: true, node: this._tableNode(tok) });
        } else if (tok.type === "hr") {
          blocks.push({ isNode: true, node: this._hrNode() });
        } else if (tok.type === "html") {
          blocks.push({ isP: true, text: tok.text.replace(/<[^>]+>/g, "") });
        }
      }
      return blocks.length ? blocks : [{ isP: true, text: "No content." }];
    }

    _headingNode(depth, text, id) {
      const tag = "h" + depth;
      return E(tag, {
        "data-anchor": id,
        id,
        style: { fontSize: "17px", fontWeight: 650, letterSpacing: "-.01em", margin: "24px 0 12px", color: "var(--text)", scrollMarginTop: "100px" },
      }, text);
    }

    _quoteNode(text) {
      return E("blockquote", {
        style: { borderLeft: "3px solid var(--accent)", padding: "4px 0 4px 18px", margin: "0 0 18px", fontSize: "16.5px", lineHeight: 1.72, color: "var(--muted)", maxWidth: 760 },
      }, text);
    }

    _listNode(tok) {
      const Tag = tok.ordered ? "ol" : "ul";
      return E(Tag, {
        style: { fontSize: "16.5px", lineHeight: 1.72, color: "var(--muted)", margin: "0 0 18px", paddingLeft: "1.4em", maxWidth: 760 },
      }, tok.items.map((it, i) => E("li", { key: i, style: { margin: "6px 0" } }, it.tokens ? this._renderInline(it.tokens) : this._inline(it.text))));
    }

    _tableNode(tok) {
      const cell = (t, i, head) => E(head ? "th" : "td", {
        key: i,
        style: { border: "1px solid var(--border-strong)", padding: "10px 14px", textAlign: "left", background: head ? "var(--pill)" : "transparent", color: head ? "var(--text)" : "var(--muted)", fontWeight: head ? 600 : 400 },
      }, this._inline(t.text || t));
      const row = (cells, i, head) => E("tr", { key: i }, cells.map((c, j) => cell(c, j, head)));
      return E("table", { style: { width: "100%", borderCollapse: "collapse", margin: "0 0 22px", fontSize: "15px", maxWidth: 760 } },
        E("thead", null, row(tok.header, 0, true)),
        E("tbody", null, tok.rows.map((r, i) => row(r, i, false))));
    }

    _hrNode() {
      return E("hr", { style: { border: "none", borderTop: "1px solid var(--border)", margin: "32px 0" } });
    }

    async _ensurePage(id) {
      if (this.state.pages[id]) return;
      const found = this._find(id);
      if (!found) return;
      try {
        const res = await fetch("./content/" + found.item.path);
        if (!res.ok) throw new Error("Failed to load " + found.item.path);
        const raw = await res.text();
        const { meta, body } = this._parseFrontmatter(raw);
        const page = {
          title: meta.title || found.item.label,
          subtitle: meta.description || found.item.description || "",
          blocks: this._markdownToBlocks(body),
        };
        this.setState({ pages: { ...this.state.pages, [id]: page } }, () => setTimeout(() => this._spy(), 60));
      } catch (e) {
        this.setState({
          pages: {
            ...this.state.pages,
            [id]: { title: found.item.label, subtitle: "", blocks: [{ isP: true, text: e.message }] },
          },
        });
      }
    }

    _content(id) {
      const found = this._find(id);
      const cached = this.state.pages[id];
      if (cached) return cached;
      return { title: found ? found.item.label : "Loading…", subtitle: "", blocks: [{ isP: true, text: "Loading…" }] };
    }

    _spy() {
      const heads = Array.from(document.querySelectorAll("[data-anchor]"));
      if (!heads.length) return;
      let act = heads[0].getAttribute("data-anchor");
      for (const h of heads) if (h.getBoundingClientRect().top <= 140) act = h.getAttribute("data-anchor");
      if (act !== this.state.active) this.setState({ active: act });
    }

    brandLogo(size) {
      size = size || 30;
      return E("img", { src: "./assets/logo.svg", alt: "", width: size, height: size, style: { display: "block" } });
    }

    icon(name, size) {
      size = size || 18;
      const MAP = {
        info: "hgi-information-circle",
        pin: "hgi-pin-location-01",
        grid: "hgi-grid-view",
        flag: "hgi-flag-01",
        users: "hgi-user-multiple-02",
        folder: "hgi-folder-01",
        search: "hgi-search-01",
        chev: "hgi-arrow-right-01",
        sun: "hgi-sun-01",
        moon: "hgi-moon-01",
        github: "hgi-github",
      };
      return E("i", { className: "hgi-stroke " + (MAP[name] || MAP.info), style: { fontSize: size, width: size, height: size } });
    }

    highlight(code) {
      const re = /(<\/?)|(>)|("(?:[^"\\]|\\.)*")|([A-Za-z_][\w.-]*)|(=)|(\s+)|([^\s])/g;
      const out = [];
      let m, expectTag = false, k = 0;
      while ((m = re.exec(code))) {
        const t = m[0];
        let c = "var(--code-text)";
        if (m[1]) { c = "var(--code-punct)"; expectTag = true; }
        else if (m[2]) { c = "var(--code-punct)"; expectTag = false; }
        else if (m[3]) c = "var(--code-string)";
        else if (m[4]) { c = expectTag ? "var(--code-tag)" : "var(--code-attr)"; expectTag = false; }
        else if (m[5]) c = "var(--code-punct)";
        out.push(E("span", { key: k++, style: { color: c } }, t));
      }
      return out;
    }

    codeCard(filename, lines, opts) {
      opts = opts || {};
      const wrap = {
        background: "var(--code-bg)",
        border: "1px solid var(--border-strong)",
        borderRadius: 12,
        padding: "18px 20px",
        overflow: "auto",
        fontFamily: MONO,
        margin: "0 0 22px",
        maxWidth: "100%",
      };
      if (opts.tall) wrap.maxHeight = "min(70vh, 720px)";
      return E("div", { style: wrap },
        E("div", { style: { fontFamily: MONO, fontSize: 13, fontWeight: 600, color: "var(--accent)" } }, filename),
        E("div", { style: { width: 34, height: 2, background: "var(--accent)", borderRadius: 2, margin: "9px 0 15px", opacity: 0.85 } }),
        lines.map((ln, i) => E("div", { key: i, style: { fontFamily: MONO, fontSize: 13.5, lineHeight: 1.85, whiteSpace: "pre" } }, ...this.highlight(ln))));
    }

    toggleTheme() {
      const t = this.state.theme === "dark" ? "light" : "dark";
      try { localStorage.setItem("atrisos-docs-theme", t); } catch (_) {}
      this.setState({ theme: t });
    }

    _renderBlock(b, i) {
      if (b.isH2) return E("h2", { key: i, "data-anchor": b.id, style: { fontSize: "27px", fontWeight: 700, letterSpacing: "-.02em", margin: "44px 0 16px", color: "var(--text)", scrollMarginTop: "100px" } }, b.text);
      if (b.isH3) return E("h3", { key: i, "data-anchor": b.id, style: { fontSize: "19px", fontWeight: 650, letterSpacing: "-.01em", margin: "30px 0 12px", color: "var(--text)", scrollMarginTop: "100px" } }, b.text);
      if (b.isP) return E("p", { key: i, style: { fontSize: "16.5px", lineHeight: 1.72, color: "var(--muted)", margin: "0 0 18px", maxWidth: 760 } }, b.text);
      if (b.isNode) return E("div", { key: i }, b.node);
      return null;
    }

    render() {
      const site = this.state.nav?.site || {};
      const accent = site.accent || "blue";
      const cur = this.state.current || "getting-started";
      const data = this._content(cur);
      const found = this._find(cur);
      const groupLabel = found ? found.group : "Documentation";
      const activeSection = this._currentSection();
      const sectionMeta = (this.state.nav?.sections || []).find((s) => s.id === activeSection);
      const brand = site.brand || "Atrisos";
      const github = site.github || "https://github.com/sonmezerekrem/atrisos";

      const tabActive = { fontSize: 14.5, fontWeight: 600, color: "var(--text)", background: "var(--pill)", padding: "7px 14px", borderRadius: 9, cursor: "pointer", textDecoration: "none" };
      const tabIdle = { fontSize: 14.5, fontWeight: 500, color: "var(--muted)", padding: "7px 12px", borderRadius: 9, cursor: "pointer", textDecoration: "none" };
      const crumbBase = { fontSize: 14, color: "var(--muted)", cursor: "default", whiteSpace: "nowrap" };

      const scrollTo = (idd) => {
        const el = document.querySelector('[data-anchor="' + idd + '"]');
        if (el) window.scrollTo({ top: el.getBoundingClientRect().top + window.scrollY - 92, behavior: "smooth" });
      };

      const q = this.state.search.trim().toLowerCase();
      let results = [];
      if (q) {
        (this.state.nav?.sections || []).forEach((sec) => sec.groups.forEach((g) => g.items.forEach((it) => {
          const hay = (it.label + " " + (it.description || "")).toLowerCase();
          if (hay.includes(q)) results.push({ label: it.label, group: sec.label, icon: it.icon || "info", id: it.id, section: sec.id });
        })));
        results = results.slice(0, 8);
      }

      const toc = [];
      data.blocks.forEach((b) => {
        if (b.isH2) toc.push({ id: b.id, title: b.text, items: [] });
        else if (b.isH3 && toc.length) toc[toc.length - 1].items.push({ id: b.id, title: b.text });
      });
      const active = this.state.active;

      return E("div", { "data-theme": this.state.theme, "data-accent": accent, style: { minHeight: "100vh", background: "var(--panel)", color: "var(--text)" } },
        E("div", { style: { maxWidth: 1664, margin: "0 auto", background: "var(--panel)", minHeight: "100vh" } },

          E("header", { style: { position: "sticky", top: 0, zIndex: 40, background: "var(--panel)", display: "flex", alignItems: "center", gap: 24, padding: "16px 40px", borderBottom: "1px solid var(--border)" } },
            E("a", {
              href: "#",
              className: "dfy-home",
              onClick: (e) => {
                e.preventDefault();
                const sec = this.state.nav?.defaultSection || "overview";
                const def = (this.state.nav?.sections || []).find((s) => s.id === sec)?.defaultPage || "getting-started";
                this.setState({ section: sec }, () => this._go(def));
              },
              style: { display: "flex", alignItems: "center", gap: 11, flexShrink: 0, cursor: "pointer", textDecoration: "none", color: "var(--text)" },
            }, E("span", { className: "brand-mark" }, this.brandLogo(30)), E("span", { style: { fontWeight: 700, fontSize: 20, letterSpacing: "-.02em" } }, brand)),

            E("div", { style: { flex: 1, display: "flex", justifyContent: "center", position: "relative" } },
              E("div", { style: { position: "relative", width: "100%", maxWidth: 560 } },
                E("div", { style: { display: "flex", alignItems: "center", gap: 10, height: 42, padding: "0 12px 0 14px", background: "var(--search-bg)", border: "1px solid var(--border-strong)", borderRadius: 11 } },
                  E("span", { style: { color: "var(--faint)", display: "flex", flexShrink: 0 } }, this.icon("search", 18)),
                  E("input", {
                    ref: this.searchRef,
                    value: this.state.search,
                    onChange: (e) => this.setState({ search: e.target.value, showResults: !!e.target.value }),
                    onFocus: () => { if (this.state.search) this.setState({ showResults: true }); },
                    onBlur: () => { setTimeout(() => this.setState({ showResults: false }), 160); },
                    placeholder: "Search",
                    style: { flex: 1, border: "none", outline: "none", background: "transparent", fontFamily: "inherit", fontSize: 15, color: "var(--text)" },
                  }),
                  E("span", { style: { fontSize: 12, color: "var(--faint)", background: "var(--kbd)", borderRadius: 6, padding: "2px 7px", flexShrink: 0 } }, "⌘ K")),
                this.state.showResults && results.length > 0 && E("div", { className: "dfy-scroll", style: { position: "absolute", top: 50, left: 0, right: 0, background: "var(--panel)", border: "1px solid var(--border-strong)", borderRadius: 12, boxShadow: "var(--shadow)", padding: 6, zIndex: 50, maxHeight: 340, overflowY: "auto" } },
                  results.map((r, i) => E("a", {
                    key: i,
                    className: "dfy-hover-row",
                    href: "#" + r.id,
                    onMouseDown: (e) => {
                      e.preventDefault();
                      this.setState({ section: r.section }, () => {
                        this._go(r.id);
                        this.setState({ search: "", showResults: false });
                      });
                    },
                    style: { display: "flex", alignItems: "center", gap: 11, padding: "9px 11px", borderRadius: 8, cursor: "pointer", textDecoration: "none" },
                  }, E("span", { style: { color: "var(--faint)", display: "flex" } }, this.icon(r.icon, 16)), E("span", { style: { fontSize: 14.5, color: "var(--text)", fontWeight: 500 } }, r.label), E("span", { style: { marginLeft: "auto", fontSize: 12, color: "var(--faint)" } }, r.group)))))),

            E("div", { style: { display: "flex", alignItems: "center", gap: 16, flexShrink: 0 } },
              E("button", {
                type: "button",
                className: "dfy-icon-btn",
                onClick: () => this.toggleTheme(),
                title: "Toggle theme",
                style: { display: "flex", alignItems: "center", justifyContent: "center", width: 36, height: 36, border: "none", background: "transparent", color: "var(--muted)", borderRadius: 9, cursor: "pointer" },
              }, this.icon(this.state.theme === "dark" ? "sun" : "moon", 18)))),

          E("nav", { style: { display: "flex", alignItems: "center", justifyContent: "space-between", padding: "9px 40px", borderBottom: "1px solid var(--border)" } },
            E("div", { style: { display: "flex", alignItems: "center", gap: 4 } },
              (this.state.nav?.sections || []).map((sec) => E("a", {
                key: sec.id,
                href: "#",
                className: "dfy-tab",
                onClick: (e) => { e.preventDefault(); this.setState({ section: sec.id }, () => this._go(sec.defaultPage)); },
                style: sec.id === activeSection ? tabActive : tabIdle,
              }, sec.label))),
            E("div", { style: { display: "flex", alignItems: "center", gap: 18, color: "var(--faint)" } },
              E("a", { href: github, target: "_blank", rel: "noopener", className: "dfy-social", style: { display: "flex", cursor: "pointer", color: "inherit" } }, this.icon("github", 17)))),

          E("div", { style: { display: "grid", gridTemplateColumns: "268px minmax(0, 1fr) 268px", alignItems: "start" } },

            E("aside", { className: "dfy-scroll", style: { position: "sticky", top: 73, maxHeight: "calc(100vh - 73px)", overflowY: "auto", padding: "30px 22px 60px 40px" } },
              this._nav().map((g) => E("div", { key: g.label },
                E("div", { style: { fontSize: 13, fontWeight: 700, color: "var(--text)", letterSpacing: ".01em", margin: "22px 0 12px 12px" } }, g.label),
                g.items.map((it) => {
                  const on = it.id === cur;
                  return E("a", {
                    key: it.id,
                    href: "#" + it.id,
                    className: on ? "" : "dfy-nav-item",
                    onClick: (e) => { e.preventDefault(); this._go(it.id); },
                    style: {
                      display: "flex", alignItems: "center", gap: 11, padding: "8px 12px", borderRadius: 9,
                      fontSize: 15, fontWeight: on ? 600 : 500, marginBottom: 1, cursor: "pointer", textDecoration: "none",
                      color: on ? "var(--accent)" : "var(--muted)",
                      background: on ? "var(--accent-soft)" : "transparent",
                    },
                  }, E("span", { style: { display: "flex", flexShrink: 0, color: on ? "var(--accent)" : "var(--faint)" } }, this.icon(it.icon, 18)), E("span", null, it.label));
                })))),

            E("main", { style: { padding: "38px 56px 120px", minWidth: 0 } },
              E("div", { style: { display: "flex", alignItems: "center", gap: 8, marginBottom: 22 } },
                E("span", { style: crumbBase }, sectionMeta?.label || "Overview"),
                E("span", { style: { color: "var(--faint)", display: "flex" } }, this.icon("chev", 15)),
                E("span", { style: crumbBase }, groupLabel),
                E("span", { style: { color: "var(--faint)", display: "flex" } }, this.icon("chev", 15)),
                E("span", { style: { fontSize: 14, color: "var(--accent)", fontWeight: 600, whiteSpace: "nowrap" } }, found ? found.item.label : data.title)),
              E("h1", { style: { fontSize: 42, fontWeight: 800, letterSpacing: "-.03em", lineHeight: 1.08, margin: "0 0 16px", color: "var(--text)" } }, data.title),
              E("p", { style: { fontSize: 18.5, lineHeight: 1.55, color: "var(--muted)", margin: "0 0 12px", maxWidth: 760 } }, data.subtitle),
              data.blocks.map((b, i) => this._renderBlock(b, i))),

            E("aside", { className: "dfy-scroll", style: { position: "sticky", top: 73, maxHeight: "calc(100vh - 73px)", overflowY: "auto", padding: "38px 40px 60px 24px" } },
              E("div", { style: { fontSize: 14, fontWeight: 700, color: "var(--text)", marginBottom: 18 } }, "On this page"),
              toc.map((t) => {
                const groupActive = t.id === active || t.items.some((s) => s.id === active);
                return E("div", { key: t.id },
                  E("a", {
                    href: "#" + t.id,
                    className: "dfy-toc",
                    onClick: (e) => { e.preventDefault(); scrollTo(t.id); },
                    style: { display: "block", fontSize: 14, fontWeight: 600, marginBottom: 10, cursor: "pointer", textDecoration: "none", color: groupActive ? "var(--accent)" : "var(--text)" },
                  }, t.title),
                  E("div", { style: { borderLeft: "1px solid var(--border-strong)", margin: "2px 0 18px" } },
                    t.items.map((s) => {
                      const on = s.id === active;
                      return E("a", {
                        key: s.id,
                        href: "#" + s.id,
                        className: "dfy-toc-sub",
                        onClick: (e) => { e.preventDefault(); scrollTo(s.id); },
                        style: {
                          display: "block", fontSize: 13.5, padding: "5px 0 5px 14px", marginLeft: -1, cursor: "pointer", textDecoration: "none",
                          borderLeft: "2px solid " + (on ? "var(--accent)" : "transparent"),
                          color: on ? "var(--accent)" : "var(--muted)", fontWeight: on ? 600 : 500,
                        },
                      }, s.title);
                    })
                  )
                );
              })
            )
          )
        )
      );
    }
  }

  const root = document.getElementById("root");
  ReactDOM.createRoot(root).render(E(DocsApp));
})();
