# Change Log

## [20.7.0] - 2020-07-01
- Disable http server by default
- Cleanup
- Added env variable `LOG_JSON` and `VERBOSE`
- Added various Docker image builds (alpine, debian, ubuntu)
- Switch to `ENTRYPOINT` instead of `CMD` for Dockerfiles

## [20.6.0] - 2020-06-29
- Switch to [Calendar Versioning](https://calver.org/) with `YY.MM.MICRO format` (year, month, increment)
- Implemented new build system
    - inject version info from git
    - linting
- Add prometheus metric support
- Switch logger to logrus
- Replace `--no-auto` with `--auto` (system crontab needs to be enabled now)

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
