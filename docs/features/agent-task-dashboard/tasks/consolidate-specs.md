---
id: "T-test-5"
title: "Consolidate Specs"
priority: "P2"
estimated_time: "20min"
dependencies: ["T-test-4.5"]
status: pending
---

# Consolidate Specs

## Description

Call `/consolidate-specs` skill to extract business rules from PRD and technical specifications from design into `specs/` directory.

## Reference Files

- `docs/features/agent-task-dashboard/prd/prd-spec.md` — Source for business rules
- `docs/features/agent-task-dashboard/design/tech-design.md` — Source for technical specs
- `docs/features/agent-task-dashboard/design/api-handbook.md` — Source for API contracts

## Acceptance Criteria

- [ ] `docs/features/agent-task-dashboard/specs/biz-specs.md` exists with extracted business rules
- [ ] `docs/features/agent-task-dashboard/specs/tech-specs.md` exists with extracted technical specs
- [ ] If any `[CROSS]` items exist: `docs/features/agent-task-dashboard/specs/review-choices.md` exists

## User Stories

No direct user story mapping. This is a standard knowledge consolidation task.

## Implementation Notes

Run `/consolidate-specs` skill. Early exit if all items are `[LOCAL]`.
