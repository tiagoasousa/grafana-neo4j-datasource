# Changelog

## 1.0.3

- Config editor redesigned to the modern Grafana layout: a sticky left sidebar
  ("Connect data source") with scroll-to-section navigation beside titled,
  described sections (Connection, Authentication, Private data source connect).
- Plugin display name shortened to "Neo4j".

## 1.0.2

- Modernize the data source configuration editor to the current Grafana UI
  pattern (`Stack` + `Field` with label/description, section headers), replacing
  the legacy `InlineField` form layout. Same settings and PDC toggle; no backend
  or configuration-schema changes.

## 1.0.1

- Update @grafana/* dependencies to 13.3.1.
- Update @emotion/css, @types/node, @testing-library/jest-dom and the eslint group.
- Pin eslint to 9.x: eslint 10 is incompatible with the eslint-plugin-react peer
  range required by @grafana/eslint-config (broke npm ci).
- CI: bump Grafana plugin actions (is-compatible, wait-for-grafana,
  create-plugin-update, bundle-size).

## 1.0.0

Initial release.
