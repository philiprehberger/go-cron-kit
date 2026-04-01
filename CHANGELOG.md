# Changelog

## 0.2.4

- Standardize README to 3-badge format with emoji Support section
- Update CI checkout action to v5 for Node.js 24 compatibility
- Add GitHub issue templates, dependabot config, and PR template

## 0.2.3

- Consolidate README badges onto single line

## 0.2.2

- Add badges and Development section to README

## 0.2.1

- Add panic recovery in job handler goroutine — panicking handlers no longer leave jobs permanently stuck
- Add test for handler panic recovery

## 0.2.0

- Fix day-of-month / day-of-week matching to use OR logic per POSIX cron standard
- Fix `Stop()` to be safe to call multiple times (no double-close panic)
- Add comprehensive test suite for parser and scheduler

## 0.1.0

- Initial release
