# LLM Providers

logidoc supports any LLM provider. Most use the OpenAI-compatible API with a different base URL.

## Anthropic (Claude)

```env
LLM_PROVIDER=anthropic
LLM_MODEL=claude-haiku-4-5-20251001
LLM_API_KEY=sk-ant-...
```

Recommended for best indexation accuracy.

## OpenAI

```env
LLM_PROVIDER=openai
LLM_MODEL=gpt-4o-mini
LLM_API_KEY=sk-...
```

## Mistral

```env
LLM_PROVIDER=mistral
LLM_MODEL=mistral-large-latest
LLM_API_KEY=...
```

## xAI (Grok)

```env
LLM_PROVIDER=xai
LLM_MODEL=grok-2
LLM_API_KEY=xai-...
```

## Groq

```env
LLM_PROVIDER=groq
LLM_MODEL=llama-3.3-70b-versatile
LLM_API_KEY=gsk_...
```

## Ollama (local)

```env
LLM_PROVIDER=ollama
LLM_MODEL=llama3.2
LLM_API_KEY=ollama
```

::: tip
When running in Docker, Ollama's default URL is `http://host.docker.internal:11434/v1`. This is configured automatically.
:::

## Custom provider

Any OpenAI-compatible API:

```env
LLM_PROVIDER=custom
LLM_MODEL=my-model
LLM_API_KEY=...
LLM_BASE_URL=https://my-api.example.com/v1
```
