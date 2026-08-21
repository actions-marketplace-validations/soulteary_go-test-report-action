# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- `.gitignore`, `.editorconfig`, Dependabot config, issue/PR templates,
  `CONTRIBUTING.md`, `CODE_OF_CONDUCT.md`, and a `.golangci.yml` lint config.
- Full CLI reference in the README (`run` flags and the `validate-paths`
  subcommand), plus a note about the differing default `cover-mode` between the
  CLI (`set`) and the Action (`atomic`).
- `golangci-lint` step in CI and a `make lint` target.

## [1.0.0]

### Added
- Initial release: single `go test -json` run producing a deterministic
  Markdown report, self-contained SVG coverage badge, and JSON output, with a
  GitHub Actions Job Summary, coverage gates, and optional in-repo write-back.

[Unreleased]: https://github.com/soulteary/go-test-report-action/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/soulteary/go-test-report-action/releases/tag/v1.0.0
