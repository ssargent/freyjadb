# Book Lore CLI — Minimal Spec

## Goal

A command-line tool for writers to **create, browse, and update** reference notes about:

- **Characters** (people, creatures)
- **Places** (locations, settings)
- **Groups** (factions, families, organizations)

Writers should be able to quickly answer “who/what/where” with short commands.

---

## Basics

- **Binary name:** `lore`
- **Framework:** Go + Cobra
- **Default output:** human-readable table
- **Machine output:** JSON via `-o json`
- **Global flags:**
  - `--project, -p <path>`: path to project (default: current dir)
  - `--format, -o <table|json>`: output format (default: `table`)
  - `--quiet, -q`: suppress non-essential messages
  - `--yes, -y`: assume “yes” for prompts (used by destructive ops)
  - `--editor`: open `$EDITOR` for multi-line fields (optional for create/update)
- **Exit codes:**
  - `0` success  
  - `1` generic error  
  - `2` not found / no results  
  - `3` invalid input  

---

## Entity Model (behavioral contract)

Each entity has:

- `id` (string; unique, stable; slug-like or UUID)
- `type` (`character` | `place` | `group`)
- `name` (string; required)
- `aka` (array of strings; optional)
- `summary` (string; 1–3 sentences)
- `details` (freeform text; optional)
- `tags` (array of strings; optional)
- `links` (array of `{type: "character|place|group", id: "<id>", relation: "<string>"}`) — optional
- `created_at`, `updated_at` (ISO 8601; maintained by app)

> The exact persistence format and indexing are **not specified** and can be chosen freely by the implementer.

---

## Command Tree
