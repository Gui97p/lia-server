# Transporte e Áudio

## Cliente vs Servidor

### Cliente

Responsável pela experiência de interação em tempo real:

- Microfone e speaker
- VAD (Voice Activity Detection)
- Wake word (inferência local do modelo .onnx)
- Captura e reprodução de áudio
- Buffering e streaming
- Estado da conexão
- Interrupção de fala
- Latência percebida
- Tools locais (openApp, windowManager, etc.)

### Servidor

Responsável pelo significado e inteligência:

- STT (Whisper via Groq)
- TTS (Edge TTS)
- LLM (Groq)
- Planner
- Executor
- Memory
- Tool registry
- Orquestração de workflows
- Autenticação e autorização

STT/TTS no servidor não significa que o cliente precisa esperar respostas completas. A comunicação pode ser streaming quando necessário.

## Transporte

Protocolo definido: **WebSocket puro** (`nhooyr.io/websocket` no server), com mensagens em JSON (`event`/`payload`) definidas pelo próprio projeto. A comunicação é bidirecional e persistente — não HTTP puro.

Socket.IO foi avaliado e descartado: suporte de terceiros imaturo tanto em Go (server, via `googollee/go-socket.io`, manutenção irregular) quanto em Rust (clients) — risco alto para dependência de longo prazo. WebSocket puro ganha o determinismo do event-driven que o projeto já quer, sem depender de uma lib de compatibilidade duvidosa entre as duas linguagens.

Rotas HTTP permanecem para operações pontuais:
- `POST /auth/login`
- `POST /audio/transcribe`
- `POST /audio/speak`

## Áudio

### Fluxo completo

```
Microfone → VAD → Wake Word (local) → captura áudio
→ POST /audio/transcribe (STT no server)
→ socket: mensagem de texto
→ Planner/Executor
→ POST /audio/speak (TTS no server)
→ client recebe bytes e reproduz
```

### Wake Word

Modelo `hey_lia.onnx` treinado via OpenWakeWord. Inferência local no client — zero latência de rede para detecção.

### Sem streaming de áudio por agora

TTS recebe texto completo e gera áudio de uma vez. Streaming de texto conflita com entonação natural do TTS — prioridade para qualidade de voz. Revisitar quando necessário.
