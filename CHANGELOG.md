# Changelog

## 1.0.1

- Update @grafana/* dependencies to 13.3.1.
- Update @emotion/css, @types/node, @testing-library/jest-dom and the eslint group.
- Pin eslint to 9.x: eslint 10 is incompatible with the eslint-plugin-react peer
  range required by @grafana/eslint-config (broke npm ci).
- CI: bump Grafana plugin actions (is-compatible, wait-for-grafana,
  create-plugin-update, bundle-size).

## 1.0.0

Initial release.
