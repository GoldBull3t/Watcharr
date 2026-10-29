# Watcharr (Enhanced Edition) 🎬🍿🎮

[![Docker Pulls](https://img.shields.io/badge/docker-ghcr.io%2Fgoldbull3t%2Fwatcharr-blue?logo=docker&logoColor=white)](https://ghcr.io/goldbull3t/watcharr)
[![Go Version](https://img.shields.io/badge/go-1.22+-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![SvelteKit](https://img.shields.io/badge/frontend-SvelteKit%202%20%2F%20Svelte%205-FF3E00?logo=svelte&logoColor=white)](https://svelte.dev/)
[![License](https://img.shields.io/badge/license-Apache%202.0-green.svg)](LICENSE)

An ultra-performant, beautifully modernized, feature-packed edition of **[Watcharr](https://github.com/sbondco/Watcharr)** — the self-hosted media tracker for movies, TV shows, and video games.

---

> [!IMPORTANT]
> **100% Backward Compatible with Upstream Watcharr.**  
> All API contracts, database entities, authentication mechanisms, and endpoints remain strictly compatible with upstream [Watcharr](https://github.com/sbondco/Watcharr). Existing mobile apps, third-party extensions, and database backups work seamlessly out of the box with zero migration headaches or breaking changes.

---

## 💡 Why Switch to this Fork?

While upstream Watcharr provides an excellent foundation, this fork turns Watcharr into the **ultimate self-hosted media & gaming headquarters**. It brings together high-speed backend optimizations, cutting-edge Svelte 5 UX, comprehensive video game tracking, interactive analytics, and native mobile gestures:

| Feature | Upstream Watcharr | This Enhanced Fork |
| :--- | :---: | :---: |
| **API & Database Compatibility** | ✅ | ✅ **100% Drop-in Replacement** |
| **Idle Memory Consumption** | ~60–80 MB | ⚡ **~23 MB RAM** (Calibrated Go runtime) |
| **Database Engines** | SQLite only | 🗄️ **PostgreSQL, MySQL & SQLite (WAL Mode)** |
| **Gaming 2.0 Tracker** | Basic status & rating | 🎮 **Hours Counter, Platforms, Completion States & Time-to-Beat** |
| **Dashboard Navigation** | 1500px+ cluttered bar | ✨ **Compact 860px Bar, Smart Menus & Status Grouping** |
| **Quick Trailers** | Deep inside detail pages | 🎬 **1-Click Modal Lightbox directly from any Poster** |
| **Visual Analytics** | Static text lists | 📊 **Watcharr Wrapped, Donut Charts & Studio/Dev Stats** |
| **Mobile & PWA Experience** | Tap-only web interface | 📱 **Touch Swipe Gestures (Watched/Planned) + Haptic Feedback** |
| **Episode Progression** | Manual season browsing | 📺 **"Up Next" Episode Engine Rail + Cliffhanger Alerts** |
| **Smart Recommendation Picker** | None | 🎲 **Watchlist Roulette (Duration, Mood, Quality)** |
| **Automated Backups** | Manual export only | 🛡️ **Zero-Downtime Scheduled Gzip Dumps with Retention** |
| **Homelab Webhooks** | Limited | 🏠 **Discord, Slack, Home Assistant, Ntfy + Emby Support** |
| **Request Integrations** | None | 🍿 **Overseerr & Jellyseerr 1-Click Requests** |

---

## 🖼️ Showcase & Interface

<img width="1913" height="913" alt="Watcharr Dashboard" src="https://github.com/user-attachments/assets/c7c07bae-be86-4054-a999-4968ea22a9cd" />
<img width="1906" height="914" alt="Watcharr Details & Sagas" src="https://github.com/user-attachments/assets/b1927bdc-5285-4e03-aa63-fd7930a2c9d7" />
<img width="1912" height="916" alt="Watcharr Media Table & Grid" src="https://github.com/user-attachments/assets/d43ad80e-6024-42cd-ad5b-9d87507bbe54" />

---

## 🌟 Key Highlights & Enhancements

### 🎮 Gaming 2.0 Tracker
* **Hours Played Counter:** Keep track of exact game times with convenient quick-increment steppers ($\pm 1\text{h}$, $\pm 5\text{h}$) or direct manual entry.
* **Platform Badges:** Tag games by platform (PC, PlayStation, Xbox, Nintendo Switch, Steam Deck, Retro, etc.).
* **4 Completion Modes:** Track whether you finished the *Main Story*, completed *Main + Extras*, achieved *100% Completionist*, or *Dropped* the game.
* **HowLongToBeat (HLTB) Comparison Bar:** Real-time visual comparison of your hours played versus community averages.
* **Dashboard Badges:** Display gaming hours and completion states seamlessly on posters and library tables.

### 🎛️ Streamlined & Compact Command Bar
* **Clean Status Architecture:** Core tabs (`All`, `Watching`, `Planned`, `Completed`) remain front and center, while secondary states are neatly tucked into `+ More ▾` (`On Hold`, `Dropped`, `Cancelled`).
* **✨ Discover Menu:** Replaces redundant buttons with a consolidated dropdown for *Watch Together*, *Franchises / Sagas*, and *Smart Roulette*.
* **Auto-Expanding In-Library Search:** Compact search input that expands smoothly upon focus with live debounced filtering.
* **Zero Horizontal Clipping:** Reduced minimum width requirement from over ~1530px down to ~860px.

### 🎬 1-Click Quick Trailers & Media Lightbox
* **Play from Any Poster:** Watch official HD trailers and teasers instantly by hovering or tapping the `▶ Trailer` button on any media card.
* **Distraction-Free Lightbox:** Clean modal player powered by `youtube-nocookie.com` with zero popups or invasive tracking.

### 📊 Visual Statistics & "Watcharr Wrapped"
* **Annual & Lifetime Wrapped Showcase:** Summary highlight card featuring total days and hours watched, total gaming hours logged, completed media counts, and average personal rating.
* **Dynamic Gamification Badges:** Automatically unlock milestone badges like *Active Cinephile*, *Binge Master*, *Veteran Gamer*, *Discerning Critic*, and *Master Collector*.
* **Interactive SVG Donut Chart:** Zero-dependency, responsive donut chart for genre distribution with slice hover states and percentage breakdown.
* **Top Studios & Game Developers:** Discover your most-watched networks (HBO, Netflix, Warner Bros.) and favorite game studios (Nintendo, Rockstar, FromSoftware).

### 📱 Mobile PWA Gestures & Haptics
* **Fluid Swipe Actions:**
  * Swipe **Right** $\rightarrow$ Instantly mark as **Completed / Watched** with an animated green confirmation badge.
  * Swipe **Left** $\rightarrow$ Instantly queue into **Planned** with a blue bookmark badge.
* **Haptic Vibration:** Tactile confirmation on mobile devices via the native Vibration API (`navigator.vibrate`).
* **Scroll-Aware:** Sophisticated directional detection ensures vertical page scrolling remains 100% natural and fluid.

### 📺 "Up Next" Episode Engine & Status Badges
* **Next Unwatched Episode Rail:** Automatically resolves the exact next unwatched episode across all your in-progress series with TMDB still backdrops and 1-click watched toggles.
* **Cliffhanger & Cancellation Alerts:** Unobtrusive tags on posters highlighting cancelled series or abrupt cliffhanger endings.

### 🎲 Smart Watchlist Roulette ("What to Watch?")
* **No More Decision Paralysis:** Spin an interactive slot-machine picker to draw a random title from your backlog.
* **Custom Filter Chips:** Filter recommendations by duration ($\le 90\text{min}$, $\le 120\text{min}$), media type, status, or minimum rating.

### 👥 Watch Together & Shareable Lists
* **Cross-Watchlist Matcher:** Compare watchlists with friends or family members to find titles you both want to watch or haven't seen yet.
* **Public & Unauthenticated Lists:** Share customized collections (`/tag/:id`) with non-registered guests via direct links, with granular privacy toggles.

### 🛡️ Automated Backups & Disaster Recovery
* **Zero-Downtime Atomic Snapshots:** Background Gzip-compressed dumps (`.db.gz` and `.json.gz`) with automatic WAL checkpointing.
* **Retention Policies:** Configurable backup scheduler (12h, 24h, 7 days) with automated pruning of older snapshots.
* **1-Click Web Management:** Download, upload, or restore backups directly from the Profile settings with pre-restore safety snapshots.

### ⚡ Extreme Performance & Multi-Database Engine
* **Ultra-Low Memory Footprint:** Idles at just **~23 MB RAM** under Docker.
* **Multi-Database Support:** Choose between **SQLite** (calibrated WAL mode and 16 MB memory cache), **PostgreSQL**, or **MySQL** via `DB_TYPE` / `DATABASE_URL`.
* **High-Speed Fuzzy Search:** Trigram indexing, Dice coefficient similarity, and Damerau-Levenshtein typo tolerance.

---

## 🐳 Quick Start with Docker (Recommended)

Pre-built multi-architecture container images (`amd64`, `arm64`) are automatically published to the GitHub Container Registry.

### Using Docker Run
```bash
docker run -d \
  --name watcharr \
  -p 3077:3077 \
  -v ./data:/data \
  --restart unless-stopped \
  ghcr.io/goldbull3t/watcharr:latest
```

### Using Docker Compose
Create a `docker-compose.yml` file:

```yaml
services:
  watcharr:
    image: ghcr.io/goldbull3t/watcharr:latest
    container_name: watcharr
    restart: unless-stopped
    ports:
      - "3077:3077"
    volumes:
      - ./data:/data
    environment:
      - PORT=3077
      # Optional external database (defaults to SQLite in /data):
      # - DB_TYPE=postgres
      # - DATABASE_URL=postgres://user:password@localhost:5432/watcharr?sslmode=disable
```

Run with:
```bash
docker compose up -d
```
Access the web UI at: **`http://localhost:3077`**.

---

## 🛠️ Local Development Setup

### Prerequisites
- [Go 1.22+](https://go.dev/)
- [Node.js 20+](https://nodejs.org/)

### 1. Backend (Go)
```bash
cd server
go run .
```
> The Go backend starts on: **`http://localhost:3080`** (Health endpoint: `http://localhost:3080/api/health`).

### 2. Frontend (SvelteKit)
```bash
# In the repository root
npm install
npm run dev
```
> The SvelteKit frontend will be running at: **`http://localhost:5173`**.

---

## 💬 Feedback, Contributions & Community

This enhanced edition is actively maintained and evolved. We warmly welcome community involvement:
- 🐛 Found a visual bug or edge case? Open an [Issue](https://github.com/GoldBull3t/Watcharr/issues).
- 💡 Have an idea for a homelab integration or feature? Start a [Discussion](https://github.com/GoldBull3t/Watcharr/discussions).
- 🚀 Want to contribute code? Pull requests are always appreciated!

---

## ⚖️ License
Distributed under the [Apache License 2.0](LICENSE).  
Based on the original [Watcharr](https://github.com/sbondco/Watcharr) project by Sam Bond and all upstream contributors.
