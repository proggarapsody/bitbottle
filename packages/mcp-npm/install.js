#!/usr/bin/env node
// Downloads the correct platform binary during npm postinstall.
const https = require("https");
const fs = require("fs");
const path = require("path");
const { execSync } = require("child_process");

const pkg = require("./package.json");
// Strip npm wrapper suffix; binary version matches npm version exactly.
const VERSION = pkg.version;
const REPO = "proggarapsody/bitbottle";

const PLATFORM_MAP = {
  darwin: "darwin",
  linux: "linux",
  win32: "windows",
};

const ARCH_MAP = {
  x64: "amd64",
  arm64: "arm64",
};

const platform = PLATFORM_MAP[process.platform];
const arch = ARCH_MAP[process.arch];

if (!platform || !arch) {
  console.error(
    `Unsupported platform: ${process.platform}/${process.arch}. ` +
      "Download the binary manually from https://github.com/" + REPO + "/releases"
  );
  process.exit(1);
}

const ext = platform === "windows" ? ".zip" : ".tar.gz";
const archiveName = `bitbottle_${platform}_${arch}${ext}`;
const url = `https://github.com/${REPO}/releases/download/v${VERSION}/${archiveName}`;
const binDir = path.join(__dirname, "bin");
const dest = path.join(binDir, "bitbottle" + (platform === "windows" ? ".exe" : ""));

if (!fs.existsSync(binDir)) fs.mkdirSync(binDir, { recursive: true });

console.log(`Downloading bitbottle v${VERSION} for ${platform}/${arch}…`);

function downloadFile(url, dest, cb) {
  const file = fs.createWriteStream(dest);
  function get(url) {
    https.get(url, (res) => {
      if (res.statusCode === 301 || res.statusCode === 302) {
        return get(res.headers.location);
      }
      if (res.statusCode !== 200) {
        cb(new Error(`HTTP ${res.statusCode} for ${url}`));
        return;
      }
      res.pipe(file);
      file.on("finish", () => file.close(cb));
    }).on("error", cb);
  }
  get(url);
}

const archiveDest = path.join(binDir, archiveName);
downloadFile(url, archiveDest, (err) => {
  if (err) {
    console.error("Download failed:", err.message);
    process.exit(1);
  }

  try {
    if (ext === ".tar.gz") {
      execSync(`tar -xzf "${archiveDest}" -C "${binDir}" bitbottle`);
    } else {
      execSync(`unzip -o "${archiveDest}" bitbottle.exe -d "${binDir}"`);
    }
    fs.unlinkSync(archiveDest);
    if (platform !== "windows") fs.chmodSync(dest, 0o755);
    console.log("bitbottle installed successfully.");
  } catch (e) {
    console.error("Extraction failed:", e.message);
    process.exit(1);
  }

  // Best-effort: register the bitbottle SKILL.md with any AI agent
  // runtimes the user has installed (Claude Code, Cursor, Codex, etc.).
  // Uses vercel-labs/skills which knows the canonical install path for
  // each runtime and symlinks/copies SKILL.md from the GitHub repo.
  //
  // This is fire-and-forget: failures (no network, no agent runtimes,
  // sudo install on root's HOME) must NOT break the binary install.
  installSkill();
});

function installSkill() {
  // Opt-out via env var.
  if (process.env.BITBOTTLE_SKIP_SKILL_INSTALL === "1") {
    return;
  }
  // Skip in CI — most CI environments install bitbottle as a build
  // dependency, not for interactive agent use. Avoids hanging the
  // pipeline on a no-op skill registration.
  if (process.env.CI || process.env.GITHUB_ACTIONS) {
    return;
  }
  // Skip when running under sudo: postinstall runs as root, so the
  // skill would land in /root/.claude/skills (or similar) instead of
  // the user's home, which silently does nothing useful. Detect via
  // SUDO_USER and bail with a hint.
  if (process.env.SUDO_USER && process.getuid && process.getuid() === 0) {
    console.log(
      "Skipped agent skill registration (running under sudo). " +
        "Re-run as your user: npx -y skills add proggarapsody/bitbottle --global -y"
    );
    return;
  }

  console.log("Registering bitbottle agent skill (Claude Code, Cursor, Codex, …)…");
  const env = { ...process.env, npm_config_yes: "true" };
  const skillsCmd = (cmd) =>
    execSync(cmd, { stdio: "ignore", timeout: 60_000, env });
  try {
    // Remove first so reinstalls actually pick up new SKILL.md content.
    // `skills add` is idempotent and skips when the skill already exists,
    // which means a fresh `npm i -g` after a release would NOT refresh
    // the skill — defeating the whole point of postinstall. Removing
    // first guarantees `add` writes the latest content.
    try {
      skillsCmd("npx -y skills remove bitbottle -g -y");
    } catch (_) {
      // not installed yet; that's fine
    }
    skillsCmd("npx -y skills add proggarapsody/bitbottle --global -y");
    console.log(
      "Agent skill registered. Set BITBOTTLE_SKIP_SKILL_INSTALL=1 to disable on next install."
    );
  } catch (e) {
    // Non-fatal. Common causes: offline, no agent runtimes installed,
    // npx blocked by corporate proxy, timeout. The CLI still works.
    console.log(
      "Skipped agent skill registration: " +
        (e.message || "unknown error") +
        ". " +
        "Install manually with: npx -y skills add proggarapsody/bitbottle --global -y"
    );
  }
}
