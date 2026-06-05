# Feature Baseline Specifications

Specification type: baseline current-state feature catalog.

This directory is a baseline catalog only. It does not contain implementation specs for individual feature work.

Do not create `docs/specs/features/<feature-name>/` until a feature is selected for future implementation.

## Files

- `feature-inventory.md`
- `domain-feature-matrix.md`
- `implemented-features.md`
- `partially-implemented-features.md`
- `stubbed-features.md`
- `planned-features.md`
- `unknown-features.md`

## Status Values

- `Implemented`: current code, APIs, persistence, and behavior are verified.
- `Partially Implemented`: current code verifies part of the feature, but verified gaps remain.
- `Stubbed`: API/function exists but core behavior is no-op or fake success.
- `Deprecated`: explicitly deprecated by current evidence.
- `Planned`: described in historical/planned docs but not verified as current implementation.
- `UNKNOWN`: insufficient evidence.

## Evidence Rules

- Current source code and migrations are source of truth for implemented behavior.
- `docs/features-develop-problematico/` is historical/planned evidence only unless current code also verifies the behavior.
