---
date: "2026-05-06"
doc_dir: "Z:/project/ai/agent-task-center/docs/features/agent-task-dashboard/ui/"
iteration: 2
target_score: 90
evaluator: Claude (automated, adversarial)
---

# UI Design Eval -- Iteration 2

**Score: 89/100** (target: 90)

```
+---------------------------------------------------------------+
|                    UI DESIGN QUALITY SCORECARD                  |
+-------------------------------+---------+---------+-----------+
| Dimension / Perspective       | Score   | Max     | Status    |
+-------------------------------+---------+---------+-----------+
| 1. Requirement Coverage (PM)  |  22     |  25     | :warning: |
|    UI function coverage       |  8/8    |         |           |
|    State requirement coverage |  7/8    |         |           |
|    Edge case handling         |  7/9    |         |           |
+-------------------------------+---------+---------+-----------+
| 2. User Experience (User)     |  22     |  25     | :warning: |
|    Information hierarchy      |  7/8    |         |           |
|    Interaction intuitiveness  |  7/8    |         |           |
|    Accessibility              |  8/9    |         |           |
+-------------------------------+---------+---------+-----------+
| 3. Design Integrity (Design)  |  23     |  25     | :warning: |
|    Design system adherence    |  8/8    |         |           |
|    Visual coherence           |  8/9    |         |           |
|    State completeness         |  7/8    |         |           |
+-------------------------------+---------+---------+-----------+
| 4. Implementability (Dev)     |  22     |  25     | :warning: |
|    Layout specificity         |  7/8    |         |           |
|    Data binding explicit      |  8/8    |         |           |
|    Interaction unambiguity    |  7/9    |         |           |
+-------------------------------+---------+---------+-----------+
| TOTAL                         |  89     |  100    |           |
+-------------------------------+---------+---------+-----------+
```

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| Component 1: States | PRD UF-1 Empty state says "Config file has zero project entries" but the design's empty state says "Add projects to ~/.task-dashboard.yaml" -- introduces a specific config file path not referenced in the PRD, and does not state the trigger condition from the PRD (zero project entries) explicitly | -1 pt (State requirement coverage) |
| All components | No edge case handling for slow network, partial failures, or concurrent actions. PRD UF-2 mentions "filter by status/priority" which implies task-level filtering -- now addressed, but no spec for what happens when the underlying index.json data changes while the user is viewing the page (stale data) | -2 pts (Edge case handling) |
| Component 1 header | Filter controls are absent from UF-1 (correct per PRD), but the header has only a dark mode toggle -- the header bar at h-14 is visually sparse with only two elements ("Task Dashboard" + toggle), creating unbalanced whitespace | -1 pt (Information hierarchy) |
| Component 2 / Component 3 | Dependency link click causes a triple sequential action (close panel + scroll swimlane + highlight task). The spec says "Close panel, scrolls swimlane to referenced task" but does not state whether these are sequential (close animation finishes, then scroll) or overlapping (close starts, scroll begins immediately). User may see a flash of un-highlighted task before highlight kicks in | -1 pt (Interaction intuitiveness) |
| Accessibility section | No specification for WCAG contrast ratios or color contrast compliance. The status colors use blue-600 on white (#2563eb on #ffffff = ~4.6:1, passes AA), but red-600 on white for blocked (#dc2626 on #ffffff = ~4.5:1, borderline) and the dark mode combinations are not validated. No explicit WCAG claim | -1 pt (Accessibility) |
| Component 2 | No responsive behavior specification below the card grid minmax(320px, 1fr). The swimlane page with w-80 sidebar + flex-1 swimlane will have a minimum content width of ~320px + 320px = 640px plus padding, but no breakpoint spec for viewports below that | -1 pt (Layout specificity) |
| Component 2 | Dependency link click in UF-3: the close-then-scroll sequence timing is ambiguous. Does the 200ms slide-out animation complete before scroll starts, or does scroll begin immediately? A developer must guess | -1 pt (Interaction unambiguity) |
| Component 2 | Row collapse animation says "200ms height animation" but no specification for how expanded-row content (task cards, dependency arrows) is handled during the transition -- are cards faded out first, then row collapses, or do cards get clipped by overflow:hidden during the height animation? | -1 pt (State completeness) |

---

## Attack Points

### Attack 1: PM -- No edge case handling for slow network or stale data

**Where**: Across all components -- no mention of slow network, partial failures, or data freshness.
**Why it's weak**: The document describes ideal-state transitions (Loading -> Populated, Loading -> Empty, Loading -> Error) but never addresses what happens when the network is slow (does the skeleton persist indefinitely? is there a timeout?). There is no stale-data handling: if a user has the swimlane page open and someone modifies index.json files on disk, the dashboard shows outdated data with no indication. The PRD's validation rules mention "dependency references to non-existent tasks are silently skipped" which is addressed, but the broader edge cases of data freshness and slow operations are not. These are realistic scenarios for a file-based dashboard.
**What must improve**: Add at least: (1) a timeout threshold for loading states (e.g., "if Loading persists > 10s, show a 'Taking longer than expected' inline message"), (2) a data refresh strategy (polling interval? manual refresh button?), (3) concurrent-action guidance (what happens if the user clicks two task cards quickly -- does the second click replace the first panel or queue?).

### Attack 2: User -- Dependency link click creates an ambiguous multi-step interaction

**Where**: Component 3 Interactions table, dependency link row: `"Close panel -> scroll to task -> highlight"` and Component 3 Layout Structure, dependency links: `"Click: closes panel, scrolls swimlane to referenced task, highlights it with ring animation."`
**Why it's weak**: The user clicks a dependency link. Three things happen: (1) the detail panel closes with a 200ms slide-out, (2) the swimlane scrolls to a target task, (3) the target task gets a ring highlight. But the sequencing is unclear. If scroll starts immediately while the panel is still closing, the user sees the swimlane shifting under the panel overlay -- jarring. If scroll waits for the panel to fully close, there is a perceptible delay where nothing happens for 200ms. Neither timing is specified. Additionally, the ring highlight on the target task has a 300ms duration -- but is that 300ms of glow-then-fade, or does the ring persist? The spec says "ring animation 300ms" which sounds like a one-shot, but a user navigating via dependency link needs the target to remain visually marked until they notice it.
**What must improve**: Specify the exact sequence and timing: (1) whether actions are sequential or overlapping, (2) whether the ring highlight persists for a minimum time (e.g., "ring persists for 2s then fades") or is momentary, (3) what happens if the target task is in a collapsed row (the spec says arrows in collapsed rows are hidden, but does not address navigating to a task in a collapsed row).

### Attack 3: Dev -- No responsive breakpoint spec for swimlane page

**Where**: Component 2 Layout Structure: `"Page layout: Flex row. Swimlane area: flex-1. Activity sidebar: w-80 (320px, collapsible to w-12)."` and Component 1: `"Grid: auto-fill, minmax(320px, 1fr), gap-6"`.
**Why it's weak**: The swimlane page layout is a flex row with a minimum of flex-1 + w-80 (320px). With padding and the phase columns needing at minimum 140px each (per task card spec "min 140px") times 5 columns = 700px, plus the 320px sidebar, the minimum viable viewport is approximately 1060px+. There is no breakpoint specification for what happens below this threshold. Does the sidebar auto-collapse? Do phase columns become scrollable? Does the layout switch to a column direction? A developer has no guidance for tablet or narrow-desktop viewports. The UF-1 landing page has a responsive grid via minmax(320px), but the swimlane page has no analogous responsive strategy.
**What must improve**: Add a responsive behavior section for the swimlane page with at least one breakpoint (e.g., "< 1024px: sidebar auto-collapses to w-12; < 768px: phase columns become horizontally scrollable within the swimlane area"). Define the minimum supported viewport width.

---

## Previous Issues Check

| Previous Attack | Addressed? | Evidence |
|----------------|------------|----------|
| Attack 1: Missing "Other" column for unrecognized phase numbers | YES | Phase header row now includes "Other" column: `"Phase 1" \| "Phase 2" \| "Phase 3+ \| "Testing" \| "Other"`. Conditional rendering spec added. Data binding row for phase column includes `"any unmatched prefix -> Other"`. |
| Attack 2: Zero accessibility specifications | YES | Dedicated "Accessibility" section added with: ARIA labels table for all icon-only buttons, sr-only spans for status dots, focus management with focus trap for UF-3, keyboard navigation order, keyboard shortcuts table, and full prefers-reduced-motion table. |
| Attack 3: Error state treatment mismatch with PRD | YES | Error state changed from "yellow-tinted card" to "Warning banner per errored project, rendered above the card grid" using `bg-secondary, border-l 4px var(--destructive)`. Colors now reference the defined palette (--secondary, --destructive). Banner is non-navigable and distinct from project cards. |
| Attack 4: State transitions undescribed | PARTIAL | State transition timings added to all components (e.g., "skeletons fade out 150ms opacity, cards fade in 150ms opacity"). Component 3 retry interaction added to Interactions table. However, row collapse animation does not specify how task cards and dependency arrows behave during the height transition -- cards are clipped? faded out first? |
| Attack 5: Filter semantics ambiguous | YES | Dedicated "Filter matching logic" section added with explicit rules: task-level matching, feature-row visibility granularity, display:none for non-matching rows, no individual task filtering. |
| Attack 6: Dependency arrow rendering under-specified | YES | Z-index layering added (SVG z-10, cards z-20 with pointer-events:none on SVG). Collapsed-row arrow behavior specified (hidden until re-expand). Dense graph simplification threshold added (50 arrows triggers badge-based summary). |

---

## Verdict

- **Score**: 89/100
- **Target**: 90/100
- **Gap**: 1 point
- **Action**: Continue to iteration 3. Priority fixes: (1) Add slow network / stale data edge case handling (+2 pts edge case coverage), (2) Specify dependency link click sequencing and highlight persistence (+1 pt intuitiveness, +1 pt unambiguity), (3) Add responsive breakpoint spec for swimlane page (+1 pt layout specificity), (4) Clarify row collapse content behavior during transition (+1 pt state completeness).
