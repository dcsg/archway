# Invariant 001: Rules File Sizing Constraints

**Status:** active
**Date:** 2026-03-10

## Constraint

Generated rules files (from `archway guide`) MUST stay within these bounds:

- **15-25 instructions** maximum per file
- **500-1,500 tokens** per file
- Critical rules (NEVER/ALWAYS) placed at **start and end** of file
- Each instruction paired with **motivation** (why, not just what)
- **3-5 code examples** included (most reliable steering mechanism)

## Rationale

Research on LLM instruction following (2025-2026 frontier models) shows:

1. **Compliance = f(instruction_count)** — more instructions exponentially reduces full compliance. At 10 instructions, GPT-4o achieves 15% full compliance. Even Gemini 2.5 Pro only achieves 68.9% at 500 instructions. (Source: IFScale, NeurIPS 2025 Workshop — https://arxiv.org/abs/2507.11538)

2. **Position matters** — "Lost in the Middle" effect: 20%+ performance drop for instructions in the middle of context. Start and end positions get highest attention. (Source: Liu et al., TACL 2023 — https://arxiv.org/abs/2307.03172)

3. **Agentic scenarios are hardest** — best models follow <30% of instructions perfectly in agentic contexts (exactly where coding agents operate). (Source: AgentIF, NeurIPS 2025 — https://arxiv.org/abs/2505.16944)

4. **Rephrasing fragility** — even GPT-5 drops 18% with subtle prompt variations. Consistent phrasing matters. (Source: IFEval++ — https://arxiv.org/abs/2512.14754)

## Consequences of Violation

- Bloated rules files (50+ instructions) will have lower total compliance than focused ones
- Users will experience inconsistent AI agent behavior
- The value proposition of `archway guide` degrades

## Enforcement

- Review generated output token count in `internal/guide/guide.go`
- Count instructions in generated content during tests
- Add test: `TestGuideOutput_InstructionCount` that fails if >30 instructions generated
