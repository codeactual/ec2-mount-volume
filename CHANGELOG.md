# Change Log

## v0.1.3

> This release updates several first/third-party dependencies.

- feat
  - --version now prints details about the build's paths and modules.
- notable dependency changes
  - Bump github.com/aws/aws-sdk-go to v1.29.8.
  - Bump github.com/pkg/errors to v0.9.1.
  - Bump internal/cage/... to latest from monorepo.
- refactor
  - Migrate to latest cage/cli/handler API (e.g. handler.Session and handler.Input) and conventions (e.g. "func NewCommand").

## v0.1.2

- fix: CLI environment variable prefix is `CAGE_` instead of conventional `EC2_`

## v0.1.1

- refactor: migrate to `./cmd/<project name>` convention

## v0.1.0

- feat: initial project export from private monorepo
