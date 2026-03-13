# Changelog

## 0.2.1

- Add panic recovery in job handler goroutine — panicking handlers no longer leave jobs permanently stuck
- Add test for handler panic recovery

## 0.2.0

- Fix day-of-month / day-of-week matching to use OR logic per POSIX cron standard
- Fix `Stop()` to be safe to call multiple times (no double-close panic)
- Add comprehensive test suite for parser and scheduler

## 0.1.0

- Initial release
