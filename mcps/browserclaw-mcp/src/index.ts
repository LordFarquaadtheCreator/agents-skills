#!/usr/bin/env node
// MCP server wrapping browserclaw. Snapshot + ref targeting, no vision model.
// One BrowserClaw instance per server lifetime. Pages tracked by integer id -> CrawlPage.
import { Server } from "@modelcontextprotocol/sdk/server/index.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import {
  CallToolRequestSchema,
  ListToolsRequestSchema,
} from "@modelcontextprotocol/sdk/types.js";
import { BrowserClaw } from "browserclaw";

type CrawlPage = Awaited<ReturnType<BrowserClaw["open"]>>;

let browser: BrowserClaw | null = null;
const pages = new Map<number, CrawlPage>();
let nextId = 1;

async function getBrowser(): Promise<BrowserClaw> {
  if (!browser) {
    // noSandbox for Docker/CI; headless default
    browser = await BrowserClaw.launch({ headless: true, noSandbox: true });
  }
  return browser;
}

function getPage(id: number): CrawlPage {
  const p = pages.get(id);
  if (!p) throw new Error(`page ${id} not found`);
  return p;
}

const tools = [
  {
    name: "new_page",
    description: "Open a new browser tab at url. Returns integer page id.",
    inputSchema: {
      type: "object",
      properties: { url: { type: "string" } },
      required: ["url"],
    },
  },
  {
    name: "list_pages",
    description: "List open page ids and their current urls.",
    inputSchema: { type: "object", properties: {} },
  },
  {
    name: "snapshot",
    description:
      "Capture accessibility tree of a page. Returns { snapshot: text, refs: {eN: {role,name}} }.",
    inputSchema: {
      type: "object",
      properties: {
        page: { type: "integer" },
        interactive: { type: "boolean", description: "only interactive elements" },
        compact: { type: "boolean", description: "drop empty structural containers" },
      },
      required: ["page"],
    },
  },
  {
    name: "click",
    description: "Click element by ref.",
    inputSchema: {
      type: "object",
      properties: {
        page: { type: "integer" },
        ref: { type: "string" },
      },
      required: ["page", "ref"],
    },
  },
  {
    name: "type",
    description:
      "Type text into element by ref. submit: true presses Enter after. slowly: true keystroke-by-keystroke.",
    inputSchema: {
      type: "object",
      properties: {
        page: { type: "integer" },
        ref: { type: "string" },
        text: { type: "string" },
        submit: { type: "boolean" },
        slowly: { type: "boolean" },
      },
      required: ["page", "ref", "text"],
    },
  },
  {
    name: "fill",
    description:
      "Fill multiple form fields at once. fields[]: { ref, type: 'text'|'checkbox'|'radio', value }.",
    inputSchema: {
      type: "object",
      properties: {
        page: { type: "integer" },
        fields: {
          type: "array",
          items: {
            type: "object",
            properties: {
              ref: { type: "string" },
              type: { type: "string" },
              value: {},
            },
            required: ["ref", "value"],
          },
        },
      },
      required: ["page", "fields"],
    },
  },
  {
    name: "press",
    description:
      "Press a key combo (e.g. 'Enter', 'Control+a'). Optional ref to focus first.",
    inputSchema: {
      type: "object",
      properties: {
        page: { type: "integer" },
        key: { type: "string" },
        ref: { type: "string" },
      },
      required: ["page", "key"],
    },
  },
  {
    name: "scroll",
    description:
      "Scroll the page viewport. direction: up|down|left|right. amount in pixels (default 600).",
    inputSchema: {
      type: "object",
      properties: {
        page: { type: "integer" },
        direction: { type: "string", enum: ["up", "down", "left", "right"] },
        amount: { type: "number" },
      },
      required: ["page", "direction"],
    },
  },
  {
    name: "navigate",
    description: "Navigate a page: url|back|forward|reload. url requires url field.",
    inputSchema: {
      type: "object",
      properties: {
        page: { type: "integer" },
        action: { type: "string", enum: ["url", "back", "forward", "reload"] },
        url: { type: "string" },
      },
      required: ["page", "action"],
    },
  },
  {
    name: "close_page",
    description: "Close a page by id.",
    inputSchema: {
      type: "object",
      properties: { page: { type: "integer" } },
      required: ["page"],
    },
  },
];

async function handleCall(name: string, args: any): Promise<any> {
  const b = await getBrowser();
  switch (name) {
    case "new_page": {
      const page = await b.open(args.url);
      const id = nextId++;
      pages.set(id, page);
      return { page: id, targetId: page.id, url: args.url };
    }
    case "list_pages": {
      const tabs = await b.tabs();
      const out = tabs.map((t: any) => ({
        id: t.targetId,
        url: t.url ?? null,
        title: t.title ?? null,
      }));
      // also include our integer-id map
      const ours: any[] = [];
      for (const [id, p] of pages) {
        ours.push({ id, targetId: p.id, url: await p.url() });
      }
      return { tabs: out, ours };
    }
    case "snapshot": {
      const p = getPage(args.page);
      const opts: any = {};
      if (args.interactive !== undefined) opts.interactive = args.interactive;
      if (args.compact !== undefined) opts.compact = args.compact;
      const { snapshot, refs } = await p.snapshot(opts);
      return { snapshot, refs };
    }
    case "click": {
      const p = getPage(args.page);
      await p.click(args.ref);
      return { clicked: args.ref };
    }
    case "type": {
      const p = getPage(args.page);
      const opts: any = {};
      if (args.submit !== undefined) opts.submit = args.submit;
      if (args.slowly !== undefined) opts.slowly = args.slowly;
      await p.type(args.ref, args.text, opts);
      return { typed: args.text.length };
    }
    case "fill": {
      const p = getPage(args.page);
      await p.fill(args.fields);
      return { filled: args.fields.length };
    }
    case "press": {
      const p = getPage(args.page);
      if (args.ref) await p.click(args.ref);
      await p.press(args.key);
      return { pressed: args.key };
    }
    case "scroll": {
      const p = getPage(args.page);
      const amt = typeof args.amount === "number" ? args.amount : 600;
      let dx = 0, dy = 0;
      switch (args.direction) {
        case "up": dy = -amt; break;
        case "down": dy = amt; break;
        case "left": dx = -amt; break;
        case "right": dx = amt; break;
      }
      // browserclaw has no directional scroll; use window.scrollBy via evaluate.
      await p.evaluate(`() => window.scrollBy(${dx}, ${dy})`);
      return { scrolled: args.direction, amount: amt };
    }
    case "navigate": {
      const p = getPage(args.page);
      if (args.action === "url") {
        const r = await p.goto(args.url);
        return { navigated: args.action, url: r.url };
      }
      if (args.action === "back") { await p.goBack(); return { navigated: "back" }; }
      if (args.action === "forward") { await p.goForward(); return { navigated: "forward" }; }
      if (args.action === "reload") { await p.reload(); return { navigated: "reload" }; }
      throw new Error(`unknown navigate action: ${args.action}`);
    }
    case "close_page": {
      const p = pages.get(args.page);
      if (p) {
        await b.close(p.id);
        pages.delete(args.page);
      }
      return { closed: args.page };
    }
    default:
      throw new Error(`unknown tool: ${name}`);
  }
}

const server = new Server(
  { name: "browserclaw-mcp", version: "0.1.0" },
  { capabilities: { tools: {} } }
);

server.setRequestHandler(ListToolsRequestSchema, async () => ({ tools }));
server.setRequestHandler(CallToolRequestSchema, async (req) => {
  try {
    const result = await handleCall(req.params.name, req.params.arguments ?? {});
    return { content: [{ type: "text", text: JSON.stringify(result) }] };
  } catch (err: any) {
    return {
      content: [{ type: "text", text: `error: ${err.message}` }],
      isError: true,
    };
  }
});

// graceful shutdown: stop browser on signal
async function shutdown() {
  if (browser) {
    try { await browser.stop("manual"); } catch {}
  }
  process.exit(0);
}
process.on("SIGINT", shutdown);
process.on("SIGTERM", shutdown);

const transport = new StdioServerTransport();
await server.connect(transport);
