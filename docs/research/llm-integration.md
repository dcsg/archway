# LLM Integration Research for Archway CLI Tool

**Research Date:** February 2026
**Purpose:** Evaluate integration approaches for LLM capabilities in an open-source, LLM-agnostic Go CLI tool

## Executive Summary

Based on comprehensive research of the current LLM integration landscape (February 2026), this document evaluates six distinct approaches for integrating AI capabilities into the Archway CLI tool for brownfield code analysis, migration planning, and semantic code understanding.

**MVP Recommendation:** OpenAI-compatible API with sashabaranov/go-openai library + optional Ollama for local development

**Long-term Architecture:** Hybrid approach combining OpenAI-compatible APIs (primary) with MCP support (emerging standard) and optional exec-based delegation for advanced use cases

---

## 1. OpenAI-Compatible API Integration

### Overview
The OpenAI chat completions format has become the de facto standard API interface, supported natively or via compatibility layers by virtually all major LLM providers.

### Provider Ecosystem (2026)

**Native OpenAI-Compatible Support:**
- **OpenAI** (GPT-4, GPT-4o, o1) - Original implementer
- **Groq** - 18x faster inference for open models like Llama 3 70B
- **Together AI** - Full OpenAI compatibility with custom model arguments
- **Ollama** - Local models with OpenAI API compatibility layer
- **Mistral AI** - Direct OpenAI-format endpoints
- **Hugging Face Inference** - Largest model selection
- **SiliconFlow** - Unified high-performance inference

**Via Gateway/Proxy:**
- **Anthropic Claude** - Via OpenRouter, LiteLLM, or custom proxies
- **Google Gemini** - Via compatibility layers
- **AWS Bedrock** - Via AWS SDK with OpenAI format
- **Azure OpenAI** - Native support through Azure endpoints

### Go Libraries

#### sashabaranov/go-openai (Recommended)

**Repository:** https://github.com/sashabaranov/go-openai
**Statistics (2026):**
- Most widely adopted unofficial OpenAI Go client
- Active maintenance with regular updates
- Community-driven with excellent documentation

**Key Features:**
- Supports ChatGPT 4o, o1, GPT-3, GPT-4, DALL·E, and Whisper
- Streaming responses
- Function calling support
- Embeddings for semantic analysis
- Requires Go 1.18+

**Code Example:**
```go
import (
    "context"
    "github.com/sashabaranov/go-openai"
)

client := openai.NewClient(apiKey)
resp, err := client.CreateChatCompletion(
    context.Background(),
    openai.ChatCompletionRequest{
        Model: openai.GPT4o,
        Messages: []openai.ChatCompletionMessage{
            {
                Role:    openai.ChatMessageRoleUser,
                Content: "Analyze this code pattern...",
            },
        },
    },
)
```

**Provider Switching:**
```go
// For Ollama local models
config := openai.DefaultConfig(apiKey)
config.BaseURL = "http://localhost:11434/v1"
client := openai.NewClientWithConfig(config)

// For Groq
config := openai.DefaultConfig(groqAPIKey)
config.BaseURL = "https://api.groq.com/openai/v1"
client := openai.NewClientWithConfig(config)
```

#### openai-go (Official)

**Repository:** https://github.com/openai/openai-go
**Notes:** Official library from OpenAI, but less flexible for multi-provider scenarios. More tightly coupled to OpenAI-specific features.

#### tmc/langchaingo (Framework)

**Repository:** https://github.com/tmc/langchaingo
**Use Case:** Complex multi-stage AI workflows

**Provider Support (10+ integrations):**
- OpenAI (GPT-4, GPT-3.5)
- Anthropic (Claude)
- Google (Gemini, PaLM)
- AWS Bedrock
- Cohere
- Mistral AI
- Ollama (local)
- Hugging Face

**When to Use:**
- Need document loaders for code analysis
- Require embeddings + vector stores for semantic search
- Building RAG (Retrieval-Augmented Generation) for codebase context
- Multi-step reasoning chains

**Trade-offs:**
- More dependencies
- Higher complexity for simple use cases
- Better for advanced semantic analysis

### Evaluation

| Criterion | Rating | Notes |
|-----------|--------|-------|
| **Developer Experience** | ⭐⭐⭐⭐⭐ | Extremely simple: set API key, make requests |
| **Portability** | ⭐⭐⭐⭐ | Works anywhere with internet; local via Ollama |
| **Cost** | ⭐⭐⭐ | API costs variable; free with Ollama local |
| **Quality** | ⭐⭐⭐⭐⭐ | Access to best models (GPT-4o, Claude Opus 4.6) |
| **OSS Friendliness** | ⭐⭐⭐⭐ | No vendor lock-in due to format standardization |
| **Future-proofing** | ⭐⭐⭐⭐⭐ | Industry standard; broad adoption |

**Pros:**
- Zero infrastructure requirements for users
- Swap providers by changing base URL + API key
- Battle-tested Go libraries available
- Access to cutting-edge models
- Simple mental model

**Cons:**
- Requires API keys (friction for first-time users)
- Internet dependency (unless using Ollama)
- Costs can accumulate with heavy usage
- Rate limits vary by provider

**Best For:**
- MVP development
- Users who want "just works" experience
- Access to latest/best models
- Teams already using cloud LLM APIs

---

## 2. LiteLLM / Universal Proxy

### Overview
LiteLLM is a Python-based proxy server that normalizes 100+ LLM APIs into a single OpenAI-compatible interface, handling authentication, retries, fallbacks, and load balancing.

**Repository:** https://github.com/BerriAI/litellm

### How It Works

```
┌─────────────┐         ┌──────────────┐         ┌────────────────┐
│   Archway   │ ────▶   │   LiteLLM    │ ────▶   │  OpenAI        │
│   (Go CLI)  │         │   Proxy      │         │  Anthropic     │
│             │ ◀────   │  (Python)    │ ◀────   │  Bedrock       │
└─────────────┘         └──────────────┘         │  100+ others   │
                                                  └────────────────┘
```

**Archway makes OpenAI-format requests to:**
- `http://localhost:4000/chat/completions`

**LiteLLM routes to configured providers:**
- Handles provider-specific authentication
- Normalizes responses to OpenAI format
- Manages fallbacks and retries
- Tracks costs across providers

### Configuration Example

```yaml
# litellm-config.yaml
model_list:
  - model_name: gpt-4
    litellm_params:
      model: openai/gpt-4
      api_key: os.environ/OPENAI_API_KEY

  - model_name: claude-opus
    litellm_params:
      model: anthropic/claude-opus-4-6
      api_key: os.environ/ANTHROPIC_API_KEY

  - model_name: local-llama
    litellm_params:
      model: ollama/llama3.1
      api_base: http://localhost:11434

general_settings:
  master_key: sk-1234  # API key for Archway to use

router_settings:
  enable_fallbacks: true
  fallback_models: ["gpt-4", "claude-opus"]
```

### Deployment Options

**Local Development:**
```bash
pip install litellm
litellm --config litellm-config.yaml
```

**Docker:**
```bash
docker run -p 4000:4000 ghcr.io/berriai/litellm:latest \
  --config /config.yaml
```

**Production:**
- Deploy to cloud infrastructure
- Use as team-shared service
- Configure load balancing and failover

### Evaluation

| Criterion | Rating | Notes |
|-----------|--------|-------|
| **Developer Experience** | ⭐⭐⭐ | Requires separate proxy deployment |
| **Portability** | ⭐⭐⭐ | Works everywhere but adds dependency |
| **Cost** | ⭐⭐⭐⭐ | Smart routing can optimize costs |
| **Quality** | ⭐⭐⭐⭐⭐ | Access to all providers |
| **OSS Friendliness** | ⭐⭐⭐⭐ | Prevents vendor lock-in completely |
| **Future-proofing** | ⭐⭐⭐⭐ | Proxy pattern is stable |

**Pros:**
- True provider abstraction - swap without code changes
- Built-in cost tracking across providers
- Automatic fallbacks and load balancing
- Centralized configuration for teams
- Guardrails and safety features

**Cons:**
- Requires running separate Python service
- Additional operational complexity
- Dependency on LiteLLM project maintenance
- Latency overhead from proxy layer
- Python runtime requirement (archway is Go)

**Best For:**
- Enterprise deployments
- Teams wanting centralized LLM management
- Multi-provider fallback strategies
- Cost optimization across providers
- Organizations with ops resources

**Not Recommended For:**
- Solo developers
- Simple CLI tools
- MVP/early development
- Users who want zero dependencies

---

## 3. Model Context Protocol (MCP)

### Overview
MCP is an open protocol standardizing how applications provide context to LLMs. Created by Anthropic in 2024, it was donated to the Linux Foundation's Agentic AI Foundation in December 2025, with OpenAI, Block, AWS, Google, Microsoft joining as members.

**Specification:** https://github.com/modelcontextprotocol/modelcontextprotocol

### Current State (2026)

**Adoption Status:**
- 2025: Year of initial adoption and experimentation
- 2026: Predicted year of enterprise expansion
- Gartner predicts 40% of enterprise applications will include AI agents by end of 2026 (up from <5%)
- Major tech companies (OpenAI, AWS, Google, Microsoft) now backing the standard
- Moving toward full standardization

**Industry Momentum:**
- Growing connector ecosystem
- Security frameworks maturing
- Enterprise-grade implementations emerging
- Still early in adoption curve but accelerating

### Architecture Options for Archway

#### Option A: Archway as MCP Server

```
┌─────────────────┐         ┌──────────────────┐
│  Claude Desktop │         │   Custom AI      │
│  (MCP Client)   │ ────▶   │   Client         │
└─────────────────┘         └──────────────────┘
                                     │
                                     ▼
                            ┌──────────────────┐
                            │  Archway         │
                            │  (MCP Server)    │
                            │                  │
                            │  - Tools:        │
                            │    • analyze_code│
                            │    • extract_adr │
                            │    • suggest_migration│
                            │  - Resources:    │
                            │    • codebase    │
                            │    • test_results│
                            └──────────────────┘
```

**Use Case:** AI clients (Claude Desktop, custom tools) can access Archway's code analysis capabilities

**Implementation:**
```go
// Using official Go SDK
import "github.com/modelcontextprotocol/go-sdk/mcp"

server := mcp.NewServer()

// Register code analysis tool
server.AddTool(mcp.Tool{
    Name:        "analyze_codebase",
    Description: "Analyze existing codebase patterns",
    InputSchema: schema,
}, analyzeCodeHandler)

// Register resources
server.AddResource(mcp.Resource{
    URI:         "codebase://current",
    Name:        "Current Codebase",
    Description: "Access to analyzed code",
}, codeLister)

// Start server
transport := mcp.NewStdioTransport()
server.Connect(transport)
```

**Value:**
- Archway becomes infrastructure for AI tools
- Other developers can build on Archway's analysis
- Standardized interface for code analysis

**Challenges:**
- Users need MCP-compatible AI client
- Limited existing MCP client ecosystem
- Doesn't solve Archway's need for LLM access

#### Option B: Archway as MCP Client

```
┌──────────────────┐
│  Archway         │
│  (MCP Client)    │
└────────┬─────────┘
         │
         ├─────▶ MCP Server: Code Search
         ├─────▶ MCP Server: Git Operations
         ├─────▶ MCP Server: Test Runner
         └─────▶ MCP Server: Documentation
```

**Use Case:** Archway consumes specialized MCP servers for functionality

**Go Implementation Options:**

**Official SDK:**
```go
import "github.com/modelcontextprotocol/go-sdk/mcp"

client := mcp.NewClient()
transport := mcp.NewStdioTransport()
client.Connect(transport)

// List available tools from server
tools, err := client.ListTools(context.Background())

// Call tool
result, err := client.CallTool(context.Background(), mcp.CallToolRequest{
    Name: "search_code",
    Arguments: map[string]any{
        "pattern": "TODO",
    },
})
```

**Community Libraries:**
- `mark3labs/mcp-go` - High-level, easy to use
- `metoro-io/mcp-golang` - Simple API, sane defaults
- `dstotijn/go-mcp` - Lightweight implementation

**Value:**
- Leverage MCP ecosystem of tools
- Standardized integrations
- Community-driven extensions

**Challenges:**
- Doesn't solve LLM provider integration (MCP is for tools/context, not LLM calls)
- MCP server ecosystem still developing
- Additional complexity vs. direct integrations

#### Option C: Hybrid (Most Practical)

```
┌──────────────────────────────────────────┐
│  Archway CLI                             │
│                                          │
│  ┌─────────────────┐  ┌────────────────┐│
│  │ Core Analysis   │  │ LLM Layer      ││
│  │ (Go native)     │  │ (OpenAI API)   ││
│  └─────────────────┘  └────────────────┘│
│           │                   │          │
│           ▼                   ▼          │
│  ┌─────────────────┐  ┌────────────────┐│
│  │ MCP Server      │  │ MCP Clients    ││
│  │ (Expose tools)  │  │ (Use tools)    ││
│  └─────────────────┘  └────────────────┘│
└──────────────────────────────────────────┘
```

**Strategy:**
1. **Primary:** Use OpenAI-compatible APIs for LLM calls (OpenAI, Claude via proxy, Ollama)
2. **Optional:** Expose Archway analysis as MCP server for other tools
3. **Future:** Consume MCP servers as ecosystem matures

### Go SDK Status (2026)

**Official SDK:** https://github.com/modelcontextprotocol/go-sdk
- Maintained in collaboration with Google
- Full MCP spec implementation
- Client and server support
- Production-ready standard library usage

**Performance:**
- Go implementations leverage standard library
- Excellent performance (see multi-language benchmarks)
- Production-grade reliability

### Evaluation

| Criterion | Rating | Notes |
|-----------|--------|-------|
| **Developer Experience** | ⭐⭐⭐ | More complex than OpenAI API |
| **Portability** | ⭐⭐⭐⭐ | Standard protocol, broad support growing |
| **Cost** | ⭐⭐⭐⭐⭐ | Protocol itself is free; costs are provider-dependent |
| **Quality** | ⭐⭐⭐ | Depends on server implementations |
| **OSS Friendliness** | ⭐⭐⭐⭐⭐ | Open standard, Linux Foundation backed |
| **Future-proofing** | ⭐⭐⭐⭐⭐ | Industry backing suggests long-term viability |

**Pros:**
- Open standard backed by major companies
- Strong future trajectory
- Excellent for exposing Archway as infrastructure
- Growing ecosystem
- True vendor neutrality

**Cons:**
- Still maturing (2026 is expansion year)
- Limited client ecosystem today
- Doesn't replace need for LLM API integration
- More complex than simple API calls
- Documentation still evolving

**Best For:**
- Long-term architectural planning
- Building reusable infrastructure
- Enterprise standardization
- Complementing (not replacing) OpenAI APIs

**Not For:**
- MVP development
- Direct LLM inference
- Solo developer projects (yet)
- Time-sensitive launches

### Recommendation
**Monitor and adopt incrementally:** Start with OpenAI APIs, design with MCP in mind, add MCP server capability in v2 as ecosystem matures.

---

## 4. Plugin-Based Providers (Terraform-Style)

### Concept
Each LLM provider is implemented as a separate plugin binary that Archway discovers and loads at runtime, similar to Terraform's provider model.

### Architecture

```
archway
├── archway (core binary)
├── plugins/
│   ├── archway-provider-openai
│   ├── archway-provider-anthropic
│   ├── archway-provider-ollama
│   └── archway-provider-bedrock
```

### Technical Implementation (Go)

**HashiCorp go-plugin System:**

**Repository:** https://github.com/hashicorp/go-plugin

**Key Features:**
- RPC over subprocess (net/rpc or gRPC)
- Plugins can't crash host process
- Language-agnostic (via gRPC)
- Battle-tested (Terraform, Vault, Nomad)

**Implementation Pattern:**

```go
// Define provider interface
type LLMProvider interface {
    Analyze(ctx context.Context, code string, prompt string) (string, error)
    GetCapabilities() []string
}

// Plugin handshake for security
var handshakeConfig = plugin.HandshakeConfig{
    ProtocolVersion:  1,
    MagicCookieKey:   "ARCHWAY_PLUGIN",
    MagicCookieValue: "archway-llm-provider",
}

// Main archway binary - loads plugins
pluginMap := plugin.NewPluginMap()
client := plugin.NewClient(&plugin.ClientConfig{
    HandshakeConfig: handshakeConfig,
    Plugins:         pluginMap,
    Cmd:            exec.Command("./plugins/archway-provider-openai"),
    AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
})

rpcClient, err := client.Client()
raw, err := rpcClient.Dispense("llm_provider")
provider := raw.(LLMProvider)

// Use provider
result, err := provider.Analyze(ctx, codeSnippet, analysisPrompt)
```

**Plugin Implementation (OpenAI provider):**
```go
// plugins/archway-provider-openai/main.go
type OpenAIProvider struct {
    client *openai.Client
}

func (p *OpenAIProvider) Analyze(ctx context.Context, code, prompt string) (string, error) {
    resp, err := p.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
        Model: openai.GPT4o,
        Messages: []openai.ChatCompletionMessage{
            {Role: "system", Content: prompt},
            {Role: "user", Content: code},
        },
    })
    return resp.Choices[0].Message.Content, err
}

func main() {
    plugin.Serve(&plugin.ServeConfig{
        HandshakeConfig: handshakeConfig,
        Plugins: map[string]plugin.Plugin{
            "llm_provider": &LLMProviderPlugin{Impl: &OpenAIProvider{}},
        },
        GRPCServer: plugin.DefaultGRPCServer,
    })
}
```

### Provider Discovery

**Option 1: Explicit Configuration**
```yaml
# ~/.archway/config.yaml
providers:
  - name: openai
    path: /usr/local/bin/archway-provider-openai
    priority: 1
  - name: claude
    path: /usr/local/bin/archway-provider-anthropic
    priority: 2
```

**Option 2: Auto-discovery**
```go
// Search standard paths
pluginDirs := []string{
    filepath.Join(homeDir, ".archway", "plugins"),
    "/usr/local/lib/archway/plugins",
    "./plugins",
}

// Scan for archway-provider-* binaries
```

### Distribution

**Package managers:**
```bash
# Homebrew
brew install archway-provider-openai

# Go install
go install github.com/archway/providers/openai@latest
```

**Single binary distribution:**
```bash
# Download plugin
archway plugin install openai
# Downloads to ~/.archway/plugins/archway-provider-openai
```

### Evaluation

| Criterion | Rating | Notes |
|-----------|--------|-------|
| **Developer Experience** | ⭐⭐ | Complex for users and maintainers |
| **Portability** | ⭐⭐⭐ | Works anywhere, but plugin distribution challenges |
| **Cost** | ⭐⭐⭐⭐⭐ | Provider-independent architecture |
| **Quality** | ⭐⭐⭐⭐ | Each provider can be optimized |
| **OSS Friendliness** | ⭐⭐⭐⭐⭐ | Community can contribute providers |
| **Future-proofing** | ⭐⭐⭐⭐ | Extensible but complex |

**Pros:**
- True provider abstraction
- Community can build providers
- Core stays lean
- Language-agnostic plugins possible (gRPC)
- Strong isolation (crashes don't affect core)
- Proven pattern (Terraform)

**Cons:**
- Significant implementation complexity
- Plugin distribution challenges
- Version compatibility matrix issues
- Debugging is harder (multiple processes)
- Slower than in-process calls
- Overkill for simple use cases

**Best For:**
- Large-scale enterprise tools
- When expecting dozens of providers
- Teams with plugin ecosystem vision
- Products with dedicated plugin developers

**Not For:**
- MVP development
- Small teams
- Simple integration needs
- Solo maintainers

### Recommendation
**Too complex for Archway's current needs.** The OpenAI-compatible API approach achieves similar provider flexibility with vastly lower complexity. Consider only if Archway becomes a major platform with community provider contributions.

---

## 5. Exec-Based / CLI Delegation

### Concept
Shell out to existing CLI tools that handle LLM interactions, using stdin/stdout for communication.

### Available CLI Tools (2026)

#### Simon Willison's `llm`

**Repository:** https://github.com/simonw/llm
**Homepage:** https://llm.datasette.io/

**Features:**
- Python-based CLI and library
- Plugin system for providers
- Tool calling support (OpenAI, Anthropic, Gemini, Ollama)
- Multi-modal support (images, audio, video)
- Local model support via plugins

**Installation:**
```bash
pip install llm
# or
brew install llm
```

**Basic Usage:**
```bash
llm "Analyze this code pattern" < code.go

# With specific model
llm -m gpt-4o "Extract architectural decisions" < src/

# With Ollama local
llm -m ollama/qwen2.5-coder "Find invariants" < domain/
```

**Plugin System:**
```bash
# Install provider plugins
llm install llm-claude-3
llm install llm-ollama

# List available models
llm models

# Configure API keys
llm keys set openai
```

#### `aichat`

**Repository:** https://github.com/sigoden/aichat
**Language:** Rust

**Features:**
- All-in-one LLM CLI tool
- 100+ LLMs across 20+ platforms
- Shell assistant (natural language → commands)
- Chat REPL mode
- RAG capabilities
- Function calling
- Role-based interactions

**Installation:**
```bash
cargo install aichat
# or
brew install aichat
```

**Usage:**
```bash
# Direct prompt
aichat "Analyze this code" < code.go

# Interactive mode
aichat

# Shell assistant
aichat -e "find all TODO comments in Go files"
# Outputs: find . -name "*.go" -exec grep -H "TODO" {} \;

# With specific provider
aichat --model openai:gpt-4o "analyze..." < code.go
```

### Integration Pattern

```go
package llm

import (
    "bytes"
    "context"
    "os/exec"
)

type ExecProvider struct {
    command string // "llm" or "aichat"
    model   string // optional model override
}

func (p *ExecProvider) Analyze(ctx context.Context, code, prompt string) (string, error) {
    var args []string

    switch p.command {
    case "llm":
        if p.model != "" {
            args = []string{"-m", p.model, prompt}
        } else {
            args = []string{prompt}
        }
    case "aichat":
        if p.model != "" {
            args = []string{"--model", p.model, prompt}
        } else {
            args = []string{prompt}
        }
    }

    cmd := exec.CommandContext(ctx, p.command, args...)
    cmd.Stdin = bytes.NewBufferString(code)

    output, err := cmd.CombinedOutput()
    return string(output), err
}

// Detection helper
func DetectAvailableCLI() []string {
    var available []string
    for _, tool := range []string{"llm", "aichat"} {
        if _, err := exec.LookPath(tool); err == nil {
            available = append(available, tool)
        }
    }
    return available
}
```

### Configuration Example

```yaml
# ~/.archway/config.yaml
llm:
  provider: exec
  exec:
    command: llm  # or aichat
    model: gpt-4o  # optional override
    fallback: true  # try exec first, fallback to API
```

### Advanced: Streaming Output

```go
func (p *ExecProvider) AnalyzeStream(ctx context.Context, code, prompt string, output io.Writer) error {
    cmd := exec.CommandContext(ctx, p.command, prompt)
    cmd.Stdin = bytes.NewBufferString(code)
    cmd.Stdout = output
    cmd.Stderr = os.Stderr
    return cmd.Run()
}

// Usage in Archway
provider.AnalyzeStream(ctx, code, prompt, os.Stdout)
// User sees real-time streaming output
```

### Evaluation

| Criterion | Rating | Notes |
|-----------|--------|-------|
| **Developer Experience** | ⭐⭐⭐⭐ | Simple implementation; users need to install tools |
| **Portability** | ⭐⭐⭐ | Requires external tool installation |
| **Cost** | ⭐⭐⭐⭐⭐ | Leverages user's existing LLM setup |
| **Quality** | ⭐⭐⭐⭐ | Depends on CLI tool and model |
| **OSS Friendliness** | ⭐⭐⭐⭐⭐ | Completely vendor-neutral |
| **Future-proofing** | ⭐⭐⭐ | Depends on maintenance of CLI tools |

**Pros:**
- Minimal code in Archway
- Leverage existing, well-maintained tools
- Users control their LLM configuration
- Respects user's provider preferences
- No API key management in Archway
- Works with user's existing setup (API keys, local models)
- Streaming support

**Cons:**
- Requires external dependency installation
- Less control over LLM interaction details
- Harder to parse structured output
- Performance overhead from process spawning
- Error handling is coarser
- Different tools have different interfaces

**Best For:**
- Prototyping and MVP
- Users who already use `llm` or `aichat`
- Minimal maintenance burden
- Developer-centric CLI tools
- When you trust user's LLM setup

**Not For:**
- Non-technical end users
- Fine-grained control needs
- Structured output requirements
- Enterprise environments with strict dependencies

### Hybrid Approach

```go
type ProviderChain struct {
    providers []Provider
}

func NewSmartProvider() *ProviderChain {
    chain := &ProviderChain{}

    // Try exec-based first if available
    if available := DetectAvailableCLI(); len(available) > 0 {
        chain.providers = append(chain.providers, &ExecProvider{command: available[0]})
    }

    // Fallback to direct API if configured
    if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
        chain.providers = append(chain.providers, &OpenAIProvider{apiKey: apiKey})
    }

    return chain
}

func (c *ProviderChain) Analyze(ctx context.Context, code, prompt string) (string, error) {
    var lastErr error
    for _, p := range c.providers {
        result, err := p.Analyze(ctx, code, prompt)
        if err == nil {
            return result, nil
        }
        lastErr = err
    }
    return "", lastErr
}
```

### Recommendation
**Excellent for developer-focused MVP.** Implement as optional provider with fallback to direct API. Very low maintenance, respects user autonomy.

---

## 6. Embedded / Local Models

### Overview
Bundle or integrate local LLM capabilities directly, eliminating external dependencies and API costs.

### Approaches

#### A. Ollama Integration (Recommended Local Approach)

**What is Ollama:**
- Local LLM runtime (like Docker for models)
- OpenAI-compatible API
- Simple model management
- Cross-platform (macOS, Linux, Windows)

**Website:** https://ollama.com

**Installation (User):**
```bash
# macOS
brew install ollama

# Linux
curl -fsSL https://ollama.com/install.sh | sh

# Start service
ollama serve
```

**Model Management:**
```bash
# Download model
ollama pull qwen2.5-coder:7b

# List installed models
ollama list

# Run model
ollama run qwen2.5-coder:7b
```

**Integration in Archway (OpenAI-Compatible):**
```go
import "github.com/sashabaranov/go-openai"

func NewOllamaClient() *openai.Client {
    config := openai.DefaultConfig("")
    config.BaseURL = "http://localhost:11434/v1"
    return openai.NewClientWithConfig(config)
}

// Use exactly like OpenAI
client := NewOllamaClient()
resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
    Model: "qwen2.5-coder:7b",
    Messages: []openai.ChatCompletionMessage{
        {Role: "user", Content: "Analyze this code..."},
    },
})
```

**Auto-detection:**
```go
func DetectOllama() bool {
    resp, err := http.Get("http://localhost:11434/api/tags")
    return err == nil && resp.StatusCode == 200
}

// Smart provider selection
if DetectOllama() {
    client = NewOllamaClient()
} else {
    client = NewOpenAIClient(apiKey)
}
```

#### B. Embedded Models (Go Native)

**Not Recommended for 2026** - Go native inference is still immature compared to Python ecosystem.

**Potential Options:**
- `ollama/ollama` as library (not designed for embedding)
- GGML bindings (experimental)
- ONNX Runtime Go bindings (limited LLM support)

**Challenges:**
- Large binary sizes (models are GB-sized)
- Performance worse than optimized runtimes
- Memory requirements
- Limited model selection
- Maintenance burden

### Best Local Models for Code Analysis (2026)

#### Qwen 2.5 Coder (Recommended)

**Models:** https://ollama.com/library/qwen2.5-coder

**Sizes:** 0.5B, 1.5B, 3B, 7B, 14B, 32B

**Performance:**
- **7B variant:** 88.4% HumanEval (beats CodeStral-22B's 81.1%)
- **32B variant:** 92.7% HumanEval, 73.7 Aider benchmark (comparable to GPT-4o)
- **32B on MdEval:** 75.2 (first among all open-source models)

**Hardware Requirements:**
- 7B: 8GB VRAM/RAM (minimum)
- 14B: 16GB VRAM/RAM
- 32B: 24GB VRAM/RAM

**Strengths:**
- Code generation
- Code reasoning
- Code fixing
- Multi-language support

**Usage:**
```bash
ollama pull qwen2.5-coder:7b
ollama run qwen2.5-coder:7b
```

#### DeepSeek Coder V2

**Models:** https://ollama.com/library/deepseek-coder

**Key Variant:** DeepSeek-Coder-V2 Lite (16B MoE)
- Only activates 2.4B parameters per inference
- Fast and memory-efficient
- Trained on 87% code, 13% natural language

**Strengths:**
- Memory efficiency
- Fast inference
- Good code understanding

**Usage:**
```bash
ollama pull deepseek-coder-v2
```

#### Qwen3-Coder-Next

**Optimized for:**
- Agentic coding workflows
- Local development
- Code completion

**Still emerging, monitor for maturity**

### Local vs. Cloud Comparison

| Aspect | Local (Ollama) | Cloud APIs |
|--------|----------------|------------|
| **Setup Time** | 5-10 minutes (download model) | 30 seconds (get API key) |
| **Cost** | Free (hardware required) | $0.01-$0.10 per request |
| **Quality (7B local)** | 88% HumanEval | GPT-4o: ~92% |
| **Quality (32B local)** | 92% HumanEval (GPT-4o level) | Claude Opus: ~95% |
| **Speed (7B)** | Fast on modern CPU/GPU | <2s typical |
| **Speed (32B)** | Requires good GPU | <2s typical |
| **Privacy** | 100% local | Data sent to provider |
| **Internet** | Not required | Required |
| **Hardware** | 8-24GB RAM | Any |

### Hybrid Strategy

```go
type LLMConfig struct {
    PreferLocal bool
    FallbackToAPI bool
}

func NewProvider(cfg LLMConfig) Provider {
    if cfg.PreferLocal && DetectOllama() {
        log.Info("Using local Ollama")
        return NewOllamaProvider()
    }

    if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
        log.Info("Using OpenAI API")
        return NewOpenAIProvider(apiKey)
    }

    if cfg.FallbackToAPI {
        log.Warn("No local model or API key found, prompting for API key")
        return PromptForAPIKey()
    }

    return nil, errors.New("no LLM provider available")
}
```

### Evaluation

| Criterion | Rating | Notes |
|-----------|--------|-------|
| **Developer Experience** | ⭐⭐⭐⭐ | Easy with Ollama; complex with embedded |
| **Portability** | ⭐⭐⭐⭐⭐ | Works offline; no API keys needed |
| **Cost** | ⭐⭐⭐⭐⭐ | Free (after hardware investment) |
| **Quality** | ⭐⭐⭐⭐ | 7B-32B models competitive with cloud |
| **OSS Friendliness** | ⭐⭐⭐⭐⭐ | Perfect - no external dependencies |
| **Future-proofing** | ⭐⭐⭐⭐ | Local models improving rapidly |

**Pros:**
- Zero recurring costs
- Complete privacy (code never leaves machine)
- Works offline
- No rate limits
- Transparent and auditable
- Good enough for many tasks (esp. 32B models)

**Cons:**
- Requires local installation (Ollama)
- Hardware requirements (8-24GB RAM)
- Quality ceiling lower than best cloud models
- Slower on CPU-only machines
- Model downloads are large (4-20GB)
- Setup friction for non-technical users

**Best For:**
- Privacy-sensitive codebases
- Offline development
- High-volume usage (cost savings)
- Developer tools with technical users
- Open-source first philosophy

**Not For:**
- Absolute best quality requirements
- Users with low-end hardware
- Non-technical end users
- When minimal setup is critical

### Recommendation
**Excellent for Archway's target audience.** Developers are comfortable installing Ollama. Offer as first-class option with graceful fallback to APIs.

---

## Comprehensive Evaluation Matrix

| Approach | DX | Portability | Cost | Quality | OSS-Friendly | Future-Proof | Complexity |
|----------|----|----|------|---------|--------------|--------------|------------|
| **OpenAI-Compatible API** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | Low |
| **LiteLLM Proxy** | ⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ | Medium |
| **MCP** | ⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | Medium |
| **Plugin System** | ⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | High |
| **Exec/CLI Delegation** | ⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | Low |
| **Local Models (Ollama)** | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | Low |

---

## Recommendations

### MVP (Immediate - Q1 2026)

**Primary Approach: OpenAI-Compatible API + Ollama Support**

```
Phase 1: Core Implementation (Week 1-2)
├── Use sashabaranov/go-openai library
├── Support provider via base URL + API key
├── Environment variables: ARCHWAY_LLM_API_KEY, ARCHWAY_LLM_BASE_URL
├── Default to OpenAI if only API key provided
└── Document setup for OpenAI, Groq, Together, Ollama

Phase 2: Enhanced UX (Week 3)
├── Auto-detect Ollama (http://localhost:11434)
├── Prompt for provider selection on first run
├── Store config in ~/.archway/config.yaml
└── Support multiple providers in config (primary + fallback)

Phase 3: Developer Experience (Week 4)
├── Command: archway configure llm
├── Interactive provider selection
├── API key management
├── Model selection per task type
└── Cost tracking (token usage)
```

**Implementation:**

```go
// pkg/llm/provider.go
type Provider interface {
    Analyze(ctx context.Context, req AnalysisRequest) (*AnalysisResponse, error)
    Capabilities() []Capability
}

type AnalysisRequest struct {
    Code       string
    Task       TaskType // BrownfieldAnalysis, MigrationPlanning, CodeUnderstanding
    Context    map[string]string
    MaxTokens  int
}

// pkg/llm/openai.go
type OpenAIProvider struct {
    client *openai.Client
    model  string
}

func NewFromConfig(cfg Config) (Provider, error) {
    switch cfg.Type {
    case "openai":
        return NewOpenAIProvider(cfg.APIKey, cfg.Model)
    case "ollama":
        return NewOllamaProvider(cfg.Model)
    default:
        return nil, fmt.Errorf("unknown provider: %s", cfg.Type)
    }
}

// Auto-detection
func AutoDetectProvider() (Provider, error) {
    // 1. Check for Ollama
    if isOllamaAvailable() {
        return NewOllamaProvider("qwen2.5-coder:7b"), nil
    }

    // 2. Check environment variables
    if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
        return NewOpenAIProvider(apiKey, "gpt-4o"), nil
    }

    // 3. Check config file
    if cfg, err := LoadConfig(); err == nil {
        return NewFromConfig(cfg)
    }

    // 4. Prompt user
    return PromptForProvider()
}
```

**Configuration File:**

```yaml
# ~/.archway/config.yaml
llm:
  # Primary provider
  provider: ollama
  model: qwen2.5-coder:7b

  # Fallback providers (optional)
  fallback:
    - provider: openai
      model: gpt-4o
      api_key: ${OPENAI_API_KEY}
    - provider: anthropic
      model: claude-opus-4-6
      api_key: ${ANTHROPIC_API_KEY}

  # Task-specific models (optional)
  tasks:
    brownfield_analysis:
      model: qwen2.5-coder:32b  # Use larger model for complex analysis
    migration_planning:
      model: gpt-4o  # Use best model for critical planning
    code_understanding:
      model: qwen2.5-coder:7b  # Smaller model is sufficient
```

**Why This Approach:**
1. **Fastest time to value** - Users with API keys work immediately
2. **Developer-friendly** - Ollama is easy for technical users
3. **No lock-in** - Swap providers via config
4. **Battle-tested** - OpenAI format is de facto standard
5. **Low maintenance** - Minimal code, leverages existing libraries

### V2 (Medium-Term - Q2-Q3 2026)

**Add: MCP Server Capability**

Expose Archway's analysis capabilities as an MCP server so other AI tools can use them.

```
archway mcp serve
# Starts MCP server on stdio or HTTP
# Other tools can connect and use:
# - analyze_codebase tool
# - extract_adr tool
# - suggest_migration tool
# - codebase resources
```

**Why:**
- MCP adoption accelerating (2026 is expansion year)
- Positions Archway as infrastructure
- Enables AI-powered development workflows
- Low effort with official Go SDK

**Implementation:**
```go
import "github.com/modelcontextprotocol/go-sdk/mcp"

func ServeMCP() {
    server := mcp.NewServer()

    server.AddTool(mcp.Tool{
        Name: "analyze_codebase",
        Description: "Analyze existing codebase patterns and extract architectural decisions",
        InputSchema: analyzeSchema,
    }, handleAnalyze)

    transport := mcp.NewStdioTransport()
    server.Connect(transport)
}
```

**Add: Optional Exec Provider**

Support delegation to `llm` or `aichat` for users who prefer those tools.

```yaml
llm:
  provider: exec
  exec:
    command: llm
    args: ["-m", "gpt-4o"]
```

**Add: Semantic Search / Embeddings**

For deeper codebase understanding.

```go
// Use embeddings for similarity search
type EmbeddingProvider interface {
    Embed(ctx context.Context, text string) ([]float64, error)
}

// Find similar code patterns
func FindSimilarCode(target string, codebase []CodeSnippet) []CodeSnippet {
    targetEmbed := provider.Embed(ctx, target)

    similarities := make([]Similarity, len(codebase))
    for i, snippet := range codebase {
        snippetEmbed := provider.Embed(ctx, snippet.Code)
        similarities[i] = Similarity{
            Snippet: snippet,
            Score:   cosineSimilarity(targetEmbed, snippetEmbed),
        }
    }

    sort.Slice(similarities, func(i, j int) {
        return similarities[i].Score > similarities[j].Score
    })

    return similarities[:10] // Top 10
}
```

### Long-Term Architecture (2027+)

**Hybrid Multi-Protocol System:**

```
┌────────────────────────────────────────────────┐
│              Archway Core                       │
├────────────────────────────────────────────────┤
│  LLM Abstraction Layer                         │
│  ├── Provider Registry                         │
│  ├── Fallback Chain                            │
│  ├── Cost Tracking                             │
│  └── Quality Monitoring                        │
├────────────────────────────────────────────────┤
│  Protocol Adapters                             │
│  ├── OpenAI API (primary)                      │
│  ├── MCP Client (for specialized tools)        │
│  ├── MCP Server (expose Archway tools)         │
│  └── Exec Provider (for user preference)       │
├────────────────────────────────────────────────┤
│  Providers                                     │
│  ├── OpenAI (GPT-4o, o1)                       │
│  ├── Anthropic (Claude Opus)                   │
│  ├── Ollama (Local: Qwen, DeepSeek)           │
│  ├── Groq (Fast inference)                     │
│  └── Together (Open models)                    │
└────────────────────────────────────────────────┘
```

**Smart Routing:**
```go
type Router struct {
    providers map[TaskType][]Provider
    costLimit float64
    qualityThreshold float64
}

func (r *Router) Route(task TaskType, req Request) Provider {
    candidates := r.providers[task]

    // Filter by cost constraint
    if r.costLimit > 0 {
        candidates = filter(candidates, func(p Provider) bool {
            return p.EstimateCost(req) <= r.costLimit
        })
    }

    // Sort by quality/speed/cost
    sort.Slice(candidates, r.scoringFunction)

    return candidates[0]
}
```

Example routing logic:
- **Simple code understanding:** Local 7B model (fast, free)
- **Complex migration planning:** GPT-4o or Claude Opus (high quality)
- **Bulk analysis:** Groq (fast, cheap)
- **Sensitive code:** Ollama local (privacy)

---

## Implementation Checklist

### MVP (Required)

- [ ] **Provider Interface** - Abstract LLM operations
- [ ] **OpenAI Provider** - Using sashabaranov/go-openai
- [ ] **Ollama Provider** - Same library, different base URL
- [ ] **Configuration** - YAML + environment variables
- [ ] **Auto-detection** - Ollama → env vars → config → prompt
- [ ] **Interactive Setup** - `archway configure llm`
- [ ] **Error Handling** - Graceful fallbacks, clear messages
- [ ] **Documentation** - Setup guides for each provider
- [ ] **Examples** - Brownfield analysis, ADR extraction, migration planning

### V2 (Nice to Have)

- [ ] **MCP Server** - Expose Archway tools via MCP
- [ ] **Exec Provider** - Support `llm` and `aichat` CLI tools
- [ ] **Embeddings** - Semantic code search
- [ ] **Cost Tracking** - Token usage and estimates
- [ ] **Multiple Providers** - Fallback chains
- [ ] **Task Routing** - Match tasks to optimal providers
- [ ] **Streaming** - Real-time output for long operations
- [ ] **Caching** - Cache LLM responses for identical requests

### V3 (Future)

- [ ] **MCP Client** - Consume specialized MCP servers
- [ ] **Plugin System** - If community demands custom providers
- [ ] **Smart Routing** - Cost/quality/speed optimization
- [ ] **Local Embeddings** - Fully offline semantic search
- [ ] **Fine-tuned Models** - Archway-specific model training
- [ ] **Feedback Loop** - Learn from user corrections
- [ ] **Multi-modal** - Diagrams, screenshots for analysis

---

## Key Decisions & Rationale

### Why OpenAI API Format?

1. **Ubiquity** - Supported by 90%+ of providers
2. **Simplicity** - Well-documented, battle-tested
3. **Go Libraries** - Excellent sashabaranov/go-openai
4. **No Lock-in** - Swap providers via base URL
5. **Future-proof** - Industry standard

### Why Ollama for Local?

1. **Developer-friendly** - Easy installation (Homebrew)
2. **OpenAI-compatible** - Reuse same code
3. **Model quality** - Qwen 2.5 Coder 32B matches GPT-4o
4. **Active development** - Rapidly improving
5. **Community** - Large model library

### Why Not Plugin System?

1. **Overkill** - OpenAI format achieves same flexibility
2. **Complexity** - Significant implementation burden
3. **Distribution** - Plugin versioning challenges
4. **Maintenance** - Multiple binaries to support
5. **Not needed** - No evidence of demand

### Why MCP in V2, Not MVP?

1. **Still maturing** - 2026 is expansion year
2. **Use case unclear** - MVP doesn't need MCP protocol
3. **Client ecosystem limited** - Few MCP clients today
4. **Server makes sense** - Expose Archway to AI tools
5. **Monitor adoption** - Add when ecosystem proves value

---

## Cost Analysis

### Typical Archway Usage Scenarios

**Scenario 1: Brownfield Analysis (Medium Codebase)**
- Input: 50 files, ~10K lines of code
- Prompts: 10 analysis requests (one per subsystem)
- Tokens per request: ~4K input + 2K output = 6K total
- Total: 60K tokens

**Costs:**
- **OpenAI GPT-4o:** $0.30 (60K tokens @ $5/1M)
- **Anthropic Claude Opus:** $0.90 (60K tokens @ $15/1M)
- **Groq (Llama 3 70B):** $0.02 (60K tokens @ $0.32/1M)
- **Ollama Local (Qwen 32B):** $0.00 (free)

**Scenario 2: Migration Planning (Large Codebase)**
- Input: 200 files, ~50K lines of code
- Prompts: 30 analysis + 10 planning requests
- Tokens per request: ~8K input + 3K output = 11K total
- Total: 440K tokens

**Costs:**
- **OpenAI GPT-4o:** $2.20
- **Anthropic Claude Opus:** $6.60
- **Groq:** $0.14
- **Ollama Local:** $0.00

**Scenario 3: Continuous Analysis (Daily Use)**
- 10 analysis requests per day
- 5K tokens per request
- 50K tokens/day = 1.5M tokens/month

**Monthly Costs:**
- **OpenAI GPT-4o:** $7.50/month
- **Anthropic Claude Opus:** $22.50/month
- **Groq:** $0.48/month
- **Ollama Local:** $0.00

### Cost Optimization Strategies

1. **Default to Local** - Use Ollama for routine analysis
2. **Cloud for Complex** - Reserve GPT-4o/Claude for hard problems
3. **Groq for Bulk** - Fast, cheap for high-volume
4. **Caching** - Cache identical analyses
5. **Smart Routing** - Match task complexity to model capability

---

## Testing Strategy

### Provider Interface Tests

```go
func TestProviderInterface(t *testing.T) {
    providers := []Provider{
        NewOpenAIProvider("test-key", "gpt-4o"),
        NewOllamaProvider("qwen2.5-coder:7b"),
        NewGroqProvider("test-key", "llama3-70b"),
    }

    for _, p := range providers {
        t.Run(fmt.Sprintf("%T", p), func(t *testing.T) {
            resp, err := p.Analyze(ctx, AnalysisRequest{
                Code: testCode,
                Task: BrownfieldAnalysis,
            })

            assert.NoError(t, err)
            assert.NotEmpty(t, resp.Analysis)
        })
    }
}
```

### Mock Provider for Testing

```go
type MockProvider struct {
    responses map[string]string
}

func (m *MockProvider) Analyze(ctx context.Context, req AnalysisRequest) (*AnalysisResponse, error) {
    if response, ok := m.responses[req.Code]; ok {
        return &AnalysisResponse{Analysis: response}, nil
    }
    return nil, errors.New("no mock response")
}

// Usage in tests
func TestMigrationPlanning(t *testing.T) {
    mock := &MockProvider{
        responses: map[string]string{
            testCode: "Analysis: This code uses pattern X...",
        },
    }

    planner := NewPlanner(mock)
    plan, err := planner.SuggestMigration(testCode)

    assert.NoError(t, err)
    assert.Contains(t, plan, "pattern X")
}
```

### Integration Tests (Optional)

```go
// +build integration

func TestRealOllama(t *testing.T) {
    if !isOllamaAvailable() {
        t.Skip("Ollama not available")
    }

    provider := NewOllamaProvider("qwen2.5-coder:7b")
    resp, err := provider.Analyze(ctx, AnalysisRequest{
        Code: sampleGoCode,
        Task: CodeUnderstanding,
    })

    assert.NoError(t, err)
    assert.NotEmpty(t, resp.Analysis)
}
```

---

## Security Considerations

### API Key Management

```go
// DON'T: Hardcode keys
provider := NewOpenAIProvider("sk-1234567890", "gpt-4o")

// DO: Environment variables
apiKey := os.Getenv("ARCHWAY_OPENAI_KEY")

// DO: Secure config file
// ~/.archway/config.yaml (chmod 600)
```

### Code Privacy

**Local Models (Ollama):**
- ✅ Code never leaves machine
- ✅ No telemetry
- ✅ Perfect for sensitive codebases

**Cloud APIs:**
- ⚠️ Code sent to provider
- ⚠️ Check provider data policies
- ⚠️ Consider data residency requirements

**Recommendations:**
1. **Default to local** for sensitive code
2. **Warn users** when sending to cloud
3. **Opt-in** for cloud providers
4. **Document** data handling

### Prompt Injection

User code might contain malicious instructions:

```go
// Sanitize user code before sending to LLM
func sanitize(code string) string {
    // Remove potential prompt injection attempts
    // This is a simple example; real implementation would be more sophisticated
    return code
}

// Structured prompts
func buildPrompt(code string, task TaskType) []Message {
    return []Message{
        {Role: "system", Content: systemPrompt[task]},
        {Role: "user", Content: fmt.Sprintf("```\n%s\n```", sanitize(code))},
    }
}
```

---

## Monitoring & Observability

### Metrics to Track

```go
type Metrics struct {
    RequestCount    int64
    TokensUsed      int64
    ErrorRate       float64
    AvgLatency      time.Duration
    CostEstimate    float64
    ProviderUsage   map[string]int64
}

// Prometheus example
var (
    llmRequests = prometheus.NewCounterVec(
        prometheus.CounterOpts{Name: "archway_llm_requests_total"},
        []string{"provider", "task", "status"},
    )

    llmTokens = prometheus.NewCounterVec(
        prometheus.CounterOpts{Name: "archway_llm_tokens_total"},
        []string{"provider", "type"},
    )

    llmLatency = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{Name: "archway_llm_latency_seconds"},
        []string{"provider", "task"},
    )
)
```

### Logging

```go
func (p *OpenAIProvider) Analyze(ctx context.Context, req AnalysisRequest) (*AnalysisResponse, error) {
    start := time.Now()

    log := log.WithFields(log.Fields{
        "provider": "openai",
        "model": p.model,
        "task": req.Task,
    })

    log.Debug("Starting LLM analysis")

    resp, err := p.client.CreateChatCompletion(ctx, ...)

    duration := time.Since(start)

    if err != nil {
        log.WithError(err).Error("LLM analysis failed")
        return nil, err
    }

    log.WithFields(log.Fields{
        "duration_ms": duration.Milliseconds(),
        "tokens_used": resp.Usage.TotalTokens,
        "cost_estimate": estimateCost(resp.Usage),
    }).Info("LLM analysis completed")

    return toAnalysisResponse(resp), nil
}
```

---

## Documentation Requirements

### User Documentation

1. **Quick Start Guide**
   - Installing Archway
   - Setting up Ollama (recommended)
   - Alternative: API keys (OpenAI, Anthropic, etc.)

2. **Provider Configuration**
   - Supported providers matrix
   - Configuration examples
   - Cost comparison
   - Privacy considerations

3. **Troubleshooting**
   - "No provider found" → setup instructions
   - Rate limit errors → fallback configuration
   - Model not found (Ollama) → `ollama pull` command
   - API errors → check API key validity

### Developer Documentation

1. **Architecture**
   - Provider interface design
   - Adding new providers
   - Testing with mock providers

2. **Best Practices**
   - When to use local vs. cloud
   - Cost optimization
   - Prompt engineering for code analysis

---

## Migration Path

### Existing Archway Users (If Any)

```yaml
# Old config (hypothetical)
analysis:
  method: static_only

# New config (backward compatible)
analysis:
  method: hybrid  # static + LLM
  llm:
    provider: ollama
    model: qwen2.5-coder:7b
```

Archway should:
1. Detect old config format
2. Auto-migrate to new format
3. Prompt user to configure LLM (optional)
4. Work without LLM (static analysis only)

---

## Conclusion

### Final Recommendations

**MVP (Immediate):**
```
✅ OpenAI-Compatible API via sashabaranov/go-openai
✅ Ollama support for local models
✅ Environment variable + config file setup
✅ Auto-detection with graceful fallbacks
✅ Clear documentation for each provider
```

**V2 (6-12 months):**
```
✅ MCP server capability (expose Archway tools)
✅ Exec provider (llm/aichat CLI delegation)
✅ Embeddings for semantic search
✅ Cost tracking and optimization
```

**V3 (Future):**
```
✅ MCP client (consume specialized tools)
✅ Smart routing (cost/quality optimization)
✅ Fine-tuned models (Archway-specific)
```

### Why This Approach Wins

1. **Fastest MVP** - Working in days, not months
2. **Best DX** - Users choose their preference (local/cloud)
3. **Future-proof** - OpenAI format + MCP = industry standards
4. **OSS-friendly** - No vendor lock-in, works offline
5. **Low maintenance** - Leverage existing libraries
6. **Scalable** - Easy to add providers/capabilities

### Success Metrics

**Technical:**
- Time to first analysis: <5 minutes
- Provider switching: <1 minute (config change)
- Analysis accuracy: >80% useful insights
- Cost per analysis: <$0.10 (cloud) or $0 (local)

**User Satisfaction:**
- Setup friction: Minimal (Ollama = `brew install`)
- Privacy: Excellent (local option available)
- Flexibility: Maximum (any provider)
- Performance: Good (Qwen 32B ~ GPT-4o quality)

---

## Research Sources

### OpenAI-Compatible API Providers
1. [Providers | OpenCode](https://opencode.ai/docs/providers/)
2. [Ultimate Guide – The Best API Providers of Open Source LLM of 2026](https://www.siliconflow.com/articles/en/The-best-API-providers-of-Open-Source-LLM)
3. [Ollama OpenAI Compatibility](https://docs.ollama.com/api/openai-compatibility)
4. [Configure LLM Provider | goose](https://block.github.io/goose/docs/getting-started/providers/)

### Go Libraries
5. [GitHub - sashabaranov/go-openai](https://github.com/sashabaranov/go-openai)
6. [How to Connect an LLM in Go](https://news-tech-io.medium.com/how-to-connect-an-llm-in-go-for-everyone-b78d30021830)
7. [Golang and LLM Integration](https://www.bacancytechnology.com/blog/golang-and-llm)
8. [GitHub - tmc/langchaingo](https://github.com/tmc/langchaingo)
9. [Top 7 Best Golang AI Agent Frameworks](https://reliasoftware.com/blog/golang-ai-agent-frameworks)

### LiteLLM
10. [LiteLLM - Getting Started](https://docs.litellm.ai/docs/)
11. [GitHub - BerriAI/litellm](https://github.com/BerriAI/litellm)
12. [LiteLLM proxy: Unified API](https://www.statsig.com/perspectives/lite-llm-proxy-api)

### Model Context Protocol (MCP)
13. [2026: The Year for Enterprise-Ready MCP Adoption](https://www.cdata.com/blog/2026-year-enterprise-ready-mcp-adoption)
14. [The Model Context Protocol's impact on 2025](https://www.thoughtworks.com/en-us/insights/blog/generative-ai/model-context-protocol-mcp-impact-2025)
15. [A Year of MCP: 2025 Review](https://www.pento.ai/blog/a-year-of-mcp-2025-review)
16. [MCP Explained: Why It Matters in 2026](https://robomotion.io/blog/mcp-explained-why-model-context-protocol-matters-in-2026)
17. [Anthropic Donates MCP to Agentic AI Foundation](https://www.anthropic.com/news/donating-the-model-context-protocol-and-establishing-of-the-agentic-ai-foundation)
18. [GitHub - modelcontextprotocol/go-sdk](https://github.com/modelcontextprotocol/go-sdk)
19. [GitHub - mark3labs/mcp-go](https://github.com/mark3labs/mcp-go)
20. [Code Analysis MCP Server](https://mcpservers.org/servers/zeocax/code-mcp)

### CLI Tools
21. [GitHub - simonw/llm](https://github.com/simonw/llm)
22. [LLM CLI Tool Documentation](https://llm.datasette.io/)
23. [GitHub - sigoden/aichat](https://github.com/sigoden/aichat)
24. [Top 5 CLI coding agents in 2026](https://pinggy.io/blog/top_cli_based_ai_coding_agents/)

### Local Models
25. [qwen2.5-coder | Ollama](https://ollama.com/library/qwen2.5-coder)
26. [deepseek-coder | Ollama](https://ollama.com/library/deepseek-coder)
27. [Best Models for Coding Locally in 2026](https://www.insiderllm.com/guides/best-local-coding-models-2026/)
28. [Best Ollama Models for Coding in 2025](https://www.codegpt.co/blog/best-ollama-model-for-coding)
29. [Choosing the Best Ollama Model for Coding](https://www.codegpt.co/blog/choosing-best-ollama-model)

### Terraform/Plugin Systems
30. [GitHub - hashicorp/go-plugin](https://github.com/hashicorp/go-plugin)
31. [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework)
32. [HashiCorp Plugin System Design](https://zerofruit-web3.medium.com/hashicorp-plugin-system-design-and-implementation-5f939f09e3b3)

### Code Analysis & Embeddings
33. [Building AI Search Using Embeddings](https://www.moderne.ai/blog/building-search-with-ai-embeddings-to-assist-automated-code-refactoring)
34. [Together We Are Better: LLM, IDE and Semantic Embedding](https://arxiv.org/html/2503.20934v2)
35. [10 Best Embedding Models 2026](https://www.openxcell.com/blog/best-embedding-models/)

### Best Practices
36. [Best Practices for LLM Integration](https://www.firecrawl.dev/blog/best-llm-observability-tools)
37. [LLM Orchestration in 2026](https://research.aimultiple.com/llm-orchestration/)
38. [OpenCode CLI: Open-Source Agentic Coding](https://yuv.ai/learn/opencode-cli)

---

**End of Research Document**

*This research was conducted in February 2026 for the Archway CLI project.*
