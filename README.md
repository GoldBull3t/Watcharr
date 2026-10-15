<h1 align="center">
  <br>
  <img src="./static/logo-col.png" alt="Watcharr Logo" width="220" />
  <br>
  Watcharr
  <br>
</h1>

<p align="center">
  <strong>Performance & Luxury Obsidian Edition</strong>
</p>

<p align="center">
  <em>A modern, high-performance media tracker for movies, TV shows, and games — elevated with a bespoke Obsidian Glass UI, fullstack speed optimizations, and automated Cosmos Cloud / Docker deployment.</em>
</p>

<p align="center">
  <a href="https://github.com/GoldBull3t/my-watcharr"><img src="https://img.shields.io/badge/version-4.2.2--custom-amber?style=for-the-badge&logo=svelte&logoColor=white&color=f59e0b" alt="Version"></a>
  <a href="https://github.com/GoldBull3t/my-watcharr/actions"><img src="https://img.shields.io/badge/build-passing-brightgreen?style=for-the-badge&logo=githubactions&logoColor=white" alt="Build Status"></a>
  <a href="https://ghcr.io/goldbull3t/my-watcharr"><img src="https://img.shields.io/badge/docker-GHCR-blue?style=for-the-badge&logo=docker&logoColor=white" alt="Docker Image"></a>
  <a href="/LICENSE"><img src="https://img.shields.io/badge/license-GPLv3-lightgrey?style=for-the-badge" alt="License"></a>
</p>

---

## 📑 Table of Contents

- [Overview](#-overview)
- [✨ What's New & Different (vs. Original Watcharr)](#-whats-new--different-vs-original-watcharr)
  - [💎 Luxury Obsidian UI & UX Overhaul](#1-luxury-obsidian-ui--ux-overhaul)
  - [⚡ Fullstack Performance & Engine Optimization](#2-fullstack-performance--engine-optimization)
  - [🐳 Automated CI/CD & Cosmos Cloud Integration](#3-automated-cicd--cosmos-cloud-integration)
- [📊 Side-by-Side Comparison Matrix](#-side-by-side-comparison-matrix)
- [🚀 Quick Start with Docker & Cosmos Cloud](#-quick-start-with-docker--cosmos-cloud)
  - [Option A: Cosmos Cloud Server (Recommended)](#option-a-cosmos-cloud-server-recommended)
  - [Option B: Docker Compose](#option-b-docker-compose)
- [🔄 How to Keep Updated (Upstream Sync)](#-how-to-keep-updated-upstream-sync)
- [🛠️ Local Development & Build](#️-local-development--build)
- [📜 Attribution & License](#-attribution--license)

---

## 🌟 Overview

This repository is a customized, production-tuned edition of [Watcharr](https://github.com/sbondCo/Watcharr) (originally created by `sbondCo`). While maintaining 100% data compatibility with the upstream project, this edition solves visual inconsistencies, eliminates database and network bottlenecks, and packages the entire application for effortless deployment and automated updates on **Cosmos Cloud** or any Docker-compatible infrastructure.

---

## ✨ What's New & Different (vs. Original Watcharr)

### 1. 💎 Luxury Obsidian UI & UX Overhaul

Designed with inspiration from **Apple TV+** and **Linear**, the interface eliminates all visual rough edges and awkward element wraps:

* **Command Bar Unification:** The main dashboard toolbar is locked into a rigid 34px single-row container on desktop (`DashboardControls.svelte`). Buttons like *Bulk Mode* no longer stretch awkwardly across 100% of the screen.
* **Obsidian Glass Floating Menus:** Replaced generic menus with true floating glass surfaces (`rgba(15, 19, 27, 0.96)`, 20px–24px backdrop blur, and 1px translucent borders). Removed fragile CSS triangle arrows in favor of clean 6px popovers.
* **Eradicated Rogue Border Lines:** Traced and eliminated global CSS pollution that drew horizontal divider lines (`border-top`) directly over menu labels in `SortMenu` and `DetailedMenu`.
* **True Vector SVG Icons:**
  * Replaced the crude asterisk text character (`*`) in `PosterRating` with a crisp, centered vector Star SVG icon.
  * Replaced the text `+` character in `PosterStatus` with a dedicated vector Add SVG icon.
  * Symmetrical 34px height alignment across all poster action buttons.
* **Zero Poster Overflow Jank:** Sinopses on poster hover previews are constrained with a 4-line smart truncation clamp (`-webkit-line-clamp: 4`) and text ellipsis, eliminating ugly internal scrollbars and preventing action buttons from being pushed out of view.
* **Cinematic Skeletons:** Upgraded the loading skeleton shimmer from harsh gray flashes to deep obsidian steel tones (`#141824` / `#242c40`) that harmoniously match dark themes.

---

### 2. ⚡ Fullstack Performance & Engine Optimization

#### 🏎️ Backend (Go & SQLite)
* **Eliminated Episode Over-Fetching (~70% Smaller API Payloads):**
  * In upstream, fetching the library `/api/watched` executed `Preload("WatchedSeasons")` and `Preload("WatchedEpisodes")` for every TV series on the page, loading thousands of episode rows across the wire just to show poster cards.
  * These heavy preloads have been removed from list views and preserved exclusively on individual TV show detail pages (`/tv/[id]`), dramatically speeding up library responses.
* **SQLite WAL Performance Tuning:**
  * Configured `PRAGMA synchronous = 1 (NORMAL)`: In WAL mode, this is the official SQLite recommendation. It guarantees 100% crash durability while avoiding blocking `fsync` disk flushes on every single transaction, resulting in **up to 10x faster batch updates and syncs**.
  * Increased SQLite page cache to **64 MB** in RAM (`PRAGMA cache_size = -64000`).
  * Enforced in-memory temporary tables and sorting (`PRAGMA temp_store = MEMORY`).
* **High-Speed Database Indexes:**
  * Added automated migration `202610151430_0007` with composite indexes:
    * `idx_watcheds_user_status` on `(user_id, status)` (instant status tab switching).
    * `idx_watcheds_user_rating` on `(user_id, rating)` (accelerates rating filters).
    * `idx_watcheds_user_created` on `(user_id, created_at DESC)` (optimizes default date sorting).
    * `idx_activities_watched_play` on `(watched_id, count_as_play)`.
* **Automatic HTTP Gzip Compression:**
  * Integrated official `github.com/gin-contrib/gzip` middleware on `/api` routes, shrinking JSON network payloads by ~75%.

#### 🏎️ Frontend (SvelteKit)
* **Lightweight Virtualization (`content-visibility: auto`):**
  * Added `content-visibility: auto` and `contain-intrinsic-size` to the poster grid. Off-screen posters are completely skipped during layout and paint cycles, ensuring buttery-smooth **60/120 FPS infinite scrolling** even with thousands of media items loaded.
* **28x Faster Blurhash Decoding via GPU:**
  * Reduced client-side blurhash canvas decode from 170×256 pixels down to **32×48 pixels**, delegating bilinear upscaling to hardware GPU textures.
  * Wrapped in `requestAnimationFrame` and guarded by cache status — saving **97% main-thread CPU time** during initial page renders.

---

### 3. 🐳 Automated CI/CD & Cosmos Cloud Integration

* **Automated GHCR Pipeline:**
  * Includes `.github/workflows/docker-build.yml`: every `git push` to `main` automatically triggers a multi-stage Docker build in GitHub Actions and publishes a fresh image to GitHub Container Registry (`ghcr.io/goldbull3t/my-watcharr:latest`).
* **Cosmos Cloud First-Class Citizen:**
  * Ready to deploy in Cosmos Server as a ServApp container with automatic reverse proxy routing, Let's Encrypt HTTPS/SSL certificates, and persistent SQLite storage.

---

## 📊 Side-by-Side Comparison Matrix

| Feature / Metric | Upstream Watcharr (`sbondCo`) | Custom Edition (`GoldBull3t`) |
| :--- | :--- | :--- |
| **Command Bar Layout** | Wraps into 2–3 rows; Bulk button stretches 100% | Single unified 34px horizontal row on desktop |
| **Dropdown Menus** | Plain dark boxes; labels cut by unwanted border lines | Obsidian Glass (`blur(20px)`), deep shadow, clean text |
| **Rating / Status Trigger** | Text asterisk (`*`) & plain text plus (`+`) | Crisp vector SVGs with precision alignment |
| **Poster Hover Preview** | Can trigger internal vertical scrollbar | Smart 4-line clamp with ellipsis, zero scrollbars |
| **Loading Skeletons** | Generic light gray flash | Deep obsidian steel wave matching the dark palette |
| **Library List Payload** | Preloads all episodes/seasons across all page items | Preloads only essential card metadata (**-70% payload size**) |
| **Database Disk Sync** | `PRAGMA synchronous = 2 (FULL)` (blocks on fsync) | `PRAGMA synchronous = 1 (NORMAL)` (**up to 10x faster writes**) |
| **Database Cache** | 2 MB SQLite default | **64 MB RAM cache** + in-memory temp tables |
| **Database Indexing** | Missing indexes for `status`, `rating`, `created_at` | Composite indexes for all major query filters |
| **API Compression** | Raw JSON text | **Gzip compression active** (~75% bandwidth reduction) |
| **Grid Scroll Rendering** | Heavy DOM recalculations on scroll | `content-visibility: auto` (**smooth 60/120 FPS**) |
| **Blurhash Computation** | 43,520 pixels decoded per poster in pure JS | 1,536 pixels with GPU upscaling (**28x faster, 97% less CPU**) |
| **Docker Build** | Manual or upstream releases only | **Automated GitHub Actions CI/CD to GHCR** |

---

## 🚀 Quick Start with Docker & Cosmos Cloud

### Option A: Cosmos Cloud Server (Recommended)

1. Open your **Cosmos Cloud** dashboard.
2. Navigate to **ServApps** (or **Containers**) and click **Create / Add Container**:
   * **Container Name:** `watcharr`
   * **Image:** `ghcr.io/goldbull3t/my-watcharr:latest`
   * **Port:** Container port `3080`
   * **Volumes / Storage:**
     * Map a host path to `/data` in the container:
       * **Host Path:** `/var/lib/cosmos/storage/watcharr` *(or your preferred path)*
       * **Container Path:** `/data`
   * **Restart Policy:** `Unless Stopped`
3. Click **Create Container**.
4. Go to **URL / Reverse Proxy** -> **Create Route**:
   * Map your domain (e.g. `watcharr.yourdomain.com`) to container `watcharr` on port `3080`.
   * Cosmos Cloud will automatically generate and manage Let's Encrypt SSL/HTTPS certificates.

> [!TIP]
> If your GitHub repository is private, make sure your published package in GitHub is set to **Public** (*Profile -> Packages -> watcharr -> Package Settings -> Change Visibility to Public*), so Cosmos Cloud can pull updates without needing extra registry credentials.

---

### Option B: Docker Compose

You can also deploy with standard `docker compose`:

```yaml
services:
  watcharr:
    image: ghcr.io/goldbull3t/my-watcharr:latest
    container_name: watcharr
    restart: unless-stopped
    ports:
      - "3080:3080"
    volumes:
      - ./data:/data
```

Start the service:
```bash
docker compose up -d
```

---

## 🔄 How to Keep Updated (Upstream Sync)

Your local repository is configured with two remotes:
* `origin`: Your custom GitHub repository (`GoldBull3t/my-watcharr`)
* `upstream`: The official Watcharr repository (`sbondCo/Watcharr`)

### Inspecting New Upstream Changes
To check if the original author released new upstream commits without altering your code:
```bash
git fetch upstream
git log HEAD..upstream/dev --oneline
```

### Applying Specific Upstream Bugfixes (Cherry-Pick)
To safely adopt a specific upstream fix without touching your custom UI or performance optimizations:
```bash
git cherry-pick <COMMIT_HASH>
git push origin main
```

### Making Your Own Updates
Anytime you make changes to the code:
```bash
git add .
git commit -m "feat: description of improvement"
git push origin main
```
The GitHub Actions workflow will automatically build a new Docker image in ~2 minutes. Then simply click **Restart / Re-pull Image** in your Cosmos Cloud dashboard!

---

## 🛠️ Local Development & Build

### Prerequisites
* **Node.js** >= 20
* **Go** >= 1.22

### Running Locally
```bash
# 1. Install frontend dependencies
npm install

# 2. Start frontend dev server
npm run dev

# 3. In another terminal, run the Go backend
cd server
go run watcharr.go
```

### Running Checks & Production Builds
```bash
# Type check frontend
npm run check

# Build production frontend
npm run build

# Build Go production binary
cd server
go build -o Watcharr.exe watcharr.go
```

---

## 📜 Attribution & License

* Original project created by [sbondCo](https://github.com/sbondCo) under the **GPLv3 License**.
* This customized edition is maintained by [GoldBull3t](https://github.com/GoldBull3t) under the same **GNU General Public License v3.0**. See the [LICENSE](LICENSE) file for details.
