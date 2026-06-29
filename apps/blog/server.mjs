import { createReadStream } from "node:fs";
import { access, readFile, stat } from "node:fs/promises";
import http from "node:http";
import path from "node:path";
import { fileURLToPath } from "node:url";

const host = process.env.HOST || "0.0.0.0";
const port = Number.parseInt(process.env.PORT || "3000", 10);
const basePath = (process.env.PUBLIC_BASE_PATH || "/blog").replace(/\/$/, "");
const rootDir = path.join(path.dirname(fileURLToPath(import.meta.url)), "dist");

const MIME_TYPES = {
  ".css": "text/css; charset=utf-8",
  ".html": "text/html; charset=utf-8",
  ".ico": "image/x-icon",
  ".jpg": "image/jpeg",
  ".js": "text/javascript; charset=utf-8",
  ".json": "application/json; charset=utf-8",
  ".png": "image/png",
  ".svg": "image/svg+xml",
  ".txt": "text/plain; charset=utf-8",
  ".webp": "image/webp",
  ".xml": "application/xml; charset=utf-8",
};

function sendJson(res, statusCode, payload) {
  res.writeHead(statusCode, { "Content-Type": "application/json; charset=utf-8" });
  res.end(JSON.stringify(payload));
}

function safeJoin(root, requestPath) {
  const normalizedPath = path
    .normalize(requestPath)
    .replace(/^([/\\])+/, "")
    .replace(/^(\.\.[/\\])+/, "");
  const resolvedRoot = path.resolve(root);
  const fullPath = path.resolve(root, normalizedPath);

  if (fullPath !== resolvedRoot && !fullPath.startsWith(`${resolvedRoot}${path.sep}`)) {
    return null;
  }

  return fullPath;
}

async function resolveFile(requestPath) {
  const decodedUrl = new URL(requestPath, "http://localhost");
  const decodedPath = decodeURIComponent(decodedUrl.pathname);
  const normalizedRequestPath =
    decodedPath === basePath || decodedPath.startsWith(`${basePath}/`)
      ? decodedPath.slice(basePath.length) || "/"
      : decodedPath;
  const candidate = safeJoin(rootDir, normalizedRequestPath);

  if (!candidate) {
    return null;
  }

  try {
    const stats = await stat(candidate);
    if (stats.isDirectory()) {
      const indexPath = path.join(candidate, "index.html");
      await access(indexPath);
      return indexPath;
    }
    return candidate;
  } catch {
    if (!path.extname(candidate)) {
      const htmlPath = `${candidate}.html`;
      try {
        await access(htmlPath);
        return htmlPath;
      } catch {
        const nestedIndexPath = path.join(candidate, "index.html");
        try {
          await access(nestedIndexPath);
          return nestedIndexPath;
        } catch {
          return null;
        }
      }
    }

    return null;
  }
}

const server = http.createServer(async (req, res) => {
  if (!req.url) {
    sendJson(res, 400, { error: "Missing request URL" });
    return;
  }

  const filePath = await resolveFile(req.url);
  if (!filePath) {
    const notFoundPath = path.join(rootDir, "404.html");
    try {
      const html = await readFile(notFoundPath);
      res.writeHead(404, { "Content-Type": "text/html; charset=utf-8" });
      res.end(html);
    } catch {
      sendJson(res, 404, { error: "Not found" });
    }
    return;
  }

  const ext = path.extname(filePath).toLowerCase();
  const contentType = MIME_TYPES[ext] || "application/octet-stream";
  res.writeHead(200, { "Content-Type": contentType });
  createReadStream(filePath).pipe(res);
});

server.listen(port, host, () => {
  console.log(`Serving static blog from ${rootDir} on http://${host}:${port}`);
});
