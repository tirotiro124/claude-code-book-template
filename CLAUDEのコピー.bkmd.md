# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A browser-based Breakout (ブロック崩し) game built with vanilla HTML/CSS/JavaScript. No build step or package manager is required — open `claude-code-book-template/index.html` directly in a browser.

## Running the Game

```bash
# Serve locally (any static file server works)
npx serve claude-code-book-template
# or
python3 -m http.server 8080 --directory claude-code-book-template
```

Then open `http://localhost:8080` (or the port shown).

## Architecture

Everything lives in `claude-code-book-template/`:

- **`index.html`** — Layout, HUD (score/lives/level), canvas element, and all CSS. Loads `main.js` at the end of `<body>`.
- **`main.js`** — Single-file game engine. Structure:
  - *Config constants* (paddle/ball/brick dimensions, colors, points per row)
  - *State variables* (`score`, `lives`, `level`, `ball`, `paddle`, `bricks`)
  - `initLevel()` — resets ball/paddle/bricks; ball speed scales with level
  - `startGame()` — resets full game state, hides start button, calls `loop()`
  - `update()` — per-frame physics: paddle movement, ball movement, wall/paddle/brick collision detection, life loss, level-clear check
  - `draw()` — renders bricks (with glow + highlight), paddle, ball, and launch hint text
  - `loop()` — `requestAnimationFrame` game loop calling `update()` + `draw()`
  - Input: keyboard (`ArrowLeft`/`ArrowRight`/`a`/`d` to move, `Space`/`ArrowUp` to launch), mouse (`mousemove` to move, `click` to launch), touch events

## Dev Environment

The devcontainer installs Node.js, GitHub CLI, and Playwright's Chromium dependencies (`post_create.sh`). Playwright can be used for automated browser testing if needed.
