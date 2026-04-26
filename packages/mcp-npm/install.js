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
});
