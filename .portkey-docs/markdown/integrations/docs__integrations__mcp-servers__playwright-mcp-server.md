> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Playwright MCP server

> The Playwright MCP server provides browser automation capabilities using Playwright through the Model Context Protocol (MCP), enabling deterministic web interaction via accessibility trees rather than screenshots.

The Playwright MCP server provides browser automation capabilities using Playwright through the Model Context Protocol (MCP). Instead of relying on screenshots or visually-tuned models, it operates on Playwright's **accessibility tree**, allowing LLMs to interact deterministically with web pages using structured data.

## When should you use it

Use the Playwright MCP server when you want an agent to:

* Automate browser actions.
* Extract structured page context via **snapshots**, avoiding ambiguity of pixel-based methods.
* Test and verify UI elements, text, or values without vision models.
* Capture **console logs, network requests, PDFs, or traces** during automated workflows.
* Manage tabs, dialogs, and file uploads in real-time web automation.

## Requirements

* **Requirements**:
  * Node.js 18 or newer
  * An MCP-compatible client (VS Code, Cursor, Windsurf, Claude Desktop, Goose, etc.)
* **Installation**:\
  Install the Playwright MCP server with your MCP client.
* **Optional capabilities** (enabled via `--caps`):
  * `vision` → coordinate-based mouse actions
  * `pdf` → save pages as PDF
  * `verify` → element/text/value verification
  * `tracing` → start/stop browser tracing

## Tools

### Core interaction

* **browser\_click** — Click (or double click) on an element.
* **browser\_hover** — Hover over an element.
* **browser\_type** — Type text into an editable element, with optional submit/slow typing.
* **browser\_fill\_form** — Fill multiple form fields at once.
* **browser\_select\_option** — Select one or more dropdown values.
* **browser\_press\_key** — Press a keyboard key.
* **browser\_drag** — Perform drag-and-drop between elements.
* **browser\_file\_upload** — Upload one or multiple files.

### Navigation & tabs

* **browser\_navigate** — Navigate to a specific URL.
* **browser\_navigate\_back** — Go back to the previous page.
* **browser\_tabs** — List, create, close, or select tabs.
* **browser\_close** — Close the current page.

### Page context & capture

* **browser\_snapshot** — Capture structured accessibility snapshot (preferred for automation).
* **browser\_take\_screenshot** — Take a screenshot of viewport, full page, or element.
* **browser\_pdf\_save** *(opt-in via `--caps=pdf`)* — Save page as PDF.

### Evaluation & debugging

* **browser\_evaluate** — Run JavaScript in page context.
* **browser\_console\_messages** — Return console messages.
* **browser\_network\_requests** — Return all network requests since load.
* **browser\_resize** — Resize the browser window.
* **browser\_handle\_dialog** — Accept/decline modal dialogs or prompts.

### Verification *(opt-in via `--caps=verify`)*

* **browser\_verify\_element\_visible** — Verify an element is visible by role + accessible name.
* **browser\_verify\_text\_visible** — Verify a text string is visible.
* **browser\_verify\_list\_visible** — Verify a list with expected items is visible.
* **browser\_verify\_value** — Verify element values (e.g., checkbox state, input value).

### Coordinate-based *(opt-in via `--caps=vision`)*

* **browser\_mouse\_click\_xy** — Click at a coordinate.
* **browser\_mouse\_drag\_xy** — Drag mouse between coordinates.
* **browser\_mouse\_move\_xy** — Move mouse to a coordinate.

### Tracing *(opt-in via `--caps=tracing`)*

* **browser\_start\_tracing** — Start trace recording.
* **browser\_stop\_tracing** — Stop trace recording.

### Installation

* **browser\_install** — Install the required browser binaries if not already present.

## Notes

* Prefer **`browser_snapshot`** over screenshots for interaction—it's structured, fast, and deterministic.
* Many tools require both a **human-readable element description** and a **ref** from the snapshot for safety and determinism.
* Optional capabilities (`vision`, `pdf`, `verify`, `tracing`) must be explicitly enabled when starting the server.
