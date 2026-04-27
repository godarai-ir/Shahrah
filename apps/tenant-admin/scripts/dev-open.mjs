import { spawn } from "node:child_process";

const APP_URL = process.env.TENANT_ADMIN_OPEN_URL ?? "http://tenant1.localhost:3000/";
const HEALTH_URL = process.env.TENANT_ADMIN_HEALTH_URL ?? "http://127.0.0.1:3000/";
const child = spawnDevServer();

let opened = false;

const timer = setInterval(async () => {
  if (opened) {
    return;
  }
  const isReady = await isServerReady(HEALTH_URL);
  if (!isReady) {
    return;
  }
  opened = true;
  clearInterval(timer);
  openBrowser(APP_URL);
}, 800);

timer.unref();

child.on("exit", (code) => {
  clearInterval(timer);
  process.exit(code ?? 0);
});

child.on("error", (error) => {
  clearInterval(timer);
  console.error("Failed to start tenant-admin dev server:", error);
  process.exit(1);
});

for (const signal of ["SIGINT", "SIGTERM"]) {
  process.on(signal, () => {
    clearInterval(timer);
    if (!child.killed) {
      try {
        child.kill(signal);
      } catch {
        child.kill();
      }
    }
  });
}

function spawnDevServer() {
  if (process.platform === "win32") {
    return spawn("cmd.exe", ["/d", "/s", "/c", "pnpm run dev:server"], {
      stdio: "inherit",
      shell: false,
      cwd: process.cwd(),
    });
  }

  return spawn("pnpm", ["run", "dev:server"], {
    stdio: "inherit",
    shell: false,
    cwd: process.cwd(),
  });
}

async function isServerReady(url) {
  try {
    const response = await fetch(url, {
      method: "GET",
      cache: "no-store",
    });
    return response.ok || response.status >= 300;
  } catch {
    return false;
  }
}

function openBrowser(url) {
  const options = {
    stdio: "ignore",
    detached: true,
    shell: false,
  };

  if (process.platform === "win32") {
    spawn("cmd", ["/c", "start", "", url], options).unref();
    return;
  }
  if (process.platform === "darwin") {
    spawn("open", [url], options).unref();
    return;
  }
  spawn("xdg-open", [url], options).unref();
}
