# Memory

`memory.Memory` stores conversation history across turns, independent of any [LLM provider](llm.md) or [agent](agents.md). It's a small, standalone interface:

```go
type Memory interface {
	Messages() []llm.Message
	Add(msgs ...llm.Message)
	Clear()
}
```

## Buffer

`memory.Buffer` is the in-process implementation — a thread-safe slice of messages, optionally capped with a sliding window:

```go
import "github.com/singhJasvinder101/agentic-go/memory"

mem := memory.NewBuffer()
mem.Add(llm.UserMessage(llm.TextPart("hi")))
mem.Add(llm.AssistantMessage(llm.TextPart("hello!")))

history := mem.Messages() // returns a copy; safe to mutate
```

### Bounding memory size

```go
mem := memory.NewBuffer(memory.WithMaxMessages(20))
```

Once the buffer exceeds 20 messages, the oldest are dropped first (FIFO). Omit `WithMaxMessages` for an unbounded buffer.

### Clearing

```go
mem.Clear()
```

## Using Memory with a plain provider call

```go
mem.Add(llm.UserMessage(llm.TextPart(userInput)))

resp, err := provider.Generate(ctx, &llm.GenerateRequest{Messages: mem.Messages()})
if err != nil {
	log.Fatal(err)
}
mem.Add(resp.AssistantMessage())
```

## Using Memory with an Agent

`agent.Agent` reads prior history from `Memory` at the start of `Run` and appends the new turn once it produces a final answer — you don't manage the append/read cycle yourself:

```go
a := agent.New(provider, agent.WithMemory(mem))
result, _ := a.Run(ctx, "What's the weather in Paris?") // mem now includes this turn
```

See [Agents](agents.md) for the full loop.

## Implementing your own Memory

Anything satisfying the three-method interface works — for example a summarizing buffer, or one backed by Redis/a database for persistence across process restarts. `Buffer` is intentionally minimal; there's no built-in persistence or summarization.

## See also

- [Agents](agents.md) — the primary consumer of `Memory`
- [LLM providers](llm.md) — `llm.Message` is the type `Memory` stores
