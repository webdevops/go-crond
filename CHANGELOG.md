# Change Log

## [Unreleased]
- Implemented new build system
    - inject version info from git
    - linting
- Add prometheus support

## [0.6.0] - 2017-06-01
- Switching of current working directory to / (root) when running cronjobs
- Add cronjob elapsed time in log
- Improve distribution detection and automatic fallback
- Switch from dynamic to static linking by default (dynamic is still available with `-dynamic` suffix)
- Replaced and reversed ` --system-defaults` to `--no-auto` (will now disable including of system default crontabs)
- Remove `--threads` (use env var `GOMAXPROCS` instead)

## [0.5.1] - 2017-05-26
- Fix crosscompiled binaries

## [0.5.0] - 2017-05-25
- Add daemon reload on SIGHUP
- Add `--allow-unprivileged` for running without root rights and without user switching capability

## [0.4.0] - 2017-05-13
- Replace `--processes` with `--threads`
- Fix argument parsing (segfault if invalid argument/option is passed)
- Include `/etc/cron.d/` as system defaults
- Improve logging

## [0.3.0] - 2017-04-28
- Add `--dumpversion`
- Add user support for `--run-parts`

## [0.2.0] - 2017-04-25
*Development release*

## [0.1.0] - 2017-04-25
*Development release*
