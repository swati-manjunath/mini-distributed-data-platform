# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]
- Added `CHANGELOG.md` for project history tracking.
- Upgraded README with detailed mini key-value store description and usage.

## [2026-05-15]
- Initialized mini KV store server with JSON `POST /put` and `GET /get` API.
- Added write-ahead logging to `data.log`.
- Implemented thread-safe in-memory store with mutex.
- Added `handlers.go`, `store.go`, `wal.go`, and `README.md` documentation.
