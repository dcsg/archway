# ADR-002: OpenAI-Compatible API for LLM Integration

**Status:** accepted
**Date:** 2026-02-14
**Deciders:** Daniel Gomes

## Context

Archway needs optional LLM capabilities for semantic analysis, ADR generation, and invariant extraction. The tool must be LLM-agnostic, work offline (local models), and degrade gracefully when no LLM is available. The LLM layer enhances deterministic analysis — it never replaces it.

## Constraints

- Must support both local (Ollama) and cloud (OpenAI, Groq, etc.) providers
- Must not require any specific LLM vendor
- All core features must work without LLM (graceful degradation)
- Go implementation (no Python/Node dependencies)
- Auto-detection chain: Ollama → env API keys → config → prompt user

## Options Considered

### Option A: OpenAI-compatible API via sashabaranov/go-openai — CHOSEN

Use a single Go library that speaks the OpenAI chat completions format. Swap providers by changing base URL.

```go
// OpenAI
config := openai.DefaultConfig(apiKey)
// Ollama
config.BaseURL = "http://localhost:11434/v1"
// Groq
config.BaseURL = "https://api.groq.com/openai/v1"
```

**Pros:**
- De facto industry standard (90%+ of providers support it)
- Single library, single API surface
- Swap providers via config, not code
- Battle-tested library (sashabaranov/go-openai)
- Ollama speaks this natively — zero adapter code
- Simple mental model

**Cons:**
- Not all providers are 100% compatible (edge cases)
- Anthropic (Claude) requires proxy/gateway for this format
- Tied to chat completions paradigm

### Option B: LiteLLM proxy

Run a Python proxy that normalizes 100+ LLM APIs into OpenAI format.

**Pros:**
- True universal provider support
- Built-in cost tracking, fallbacks, load balancing

**Cons:**
- Requires separate Python service (Archway is Go)
- Operational complexity
- Latency overhead from proxy layer
- Dependency on external project

### Option C: MCP for LLM calls

Use MCP protocol to delegate LLM reasoning to the host.

**Pros:**
- No API key management in Archway
- Host LLM does reasoning

**Cons:**
- MCP is for tools/context, not LLM inference
- Doesn't solve standalone CLI mode
- Limited to MCP-compatible hosts

### Option D: Terraform-style plugin providers

Each LLM is a separate plugin binary.

**Pros:**
- True abstraction, community can add providers

**Cons:**
- Massive complexity for a simple interface (send prompt, get response)
- Multiple binaries to distribute
- Overkill — OpenAI format already achieves provider flexibility

### Option E: Exec-based delegation (llm CLI, aichat)

Shell out to user's existing LLM CLI tool.

**Pros:**
- Minimal code, leverages user's setup
- No API key management

**Cons:**
- Requires external tool installation
- Harder to parse structured output
- Different tools have different interfaces
- Process spawning overhead

## Decision

**Option A: OpenAI-compatible API** as primary, with Ollama as the recommended local provider. The auto-detection chain provides zero-config experience:

1. Check for Ollama at `localhost:11434`
2. Check `OPENAI_API_KEY` / `ARCHWAY_LLM_API_KEY` env vars
3. Check `~/.archway/config.yaml`
4. Prompt user interactively

This covers 95% of use cases with minimal code. Option E (exec-based) may be added as a P2 fallback for power users.

## Consequences

- Archway depends on sashabaranov/go-openai library
- Anthropic Claude access requires a proxy (OpenRouter, LiteLLM) — documented, not solved in code
- All LLM features must have a `--no-llm` escape hatch
- LLM responses must never be trusted without deterministic validation
- Token usage should be logged for cost awareness
