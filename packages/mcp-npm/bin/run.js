#!/usr/bin/env node
const { execFileSync } = require("child_process");
const path = require("path");

const isWindows = process.platform === "win32";
const binary = path.join(__dirname, "bitbottle" + (isWindows ? ".exe" : ""));

try {
  execFileSync(binary, process.argv.slice(2), { stdio: "inherit" });
} catch (e) {
  process.exit(e.status ?? 1);
}
