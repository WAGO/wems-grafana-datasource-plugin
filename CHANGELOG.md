# Changelog

## Unreleased

Updated tooling to create-plugin 7.9.2:

- Grafana packages (`@grafana/data`, `runtime`, `schema`, `ui`) bumped to v13, added `@grafana/i18n`
- React 18 -> 19
- ESLint 8 -> 9, migrated to flat config (`eslint.config.mjs`)
- Go 1.24.6 -> 1.26.5, `grafana-plugin-sdk-go` v0.279.0 -> v0.296.3
- `@grafana/plugin-e2e` 2.2.3 -> 3.11.0, so e2e page objects match current Grafana UI selectors
- `grafanaDependency` kept at `>=10.4.0`, verified by e2e smoke tests against Grafana 10.4.19, 12.2.0 and 13.2.0

## 0.1.2

Updated vulnerable package in 

## 0.1.1 (not released)

Added go version to "release" github action.

## 0.1.0 (not released)

Initial release.
