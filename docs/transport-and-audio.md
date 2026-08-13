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
- `POST /audio/transcribe`
- `POST /audio/speak`

Sem `/auth/login` — não existe senha nem fluxo de login em rede; tokens são gerados via `lia-admin` (ver [Identidade, Autenticação e Secrets](identity-auth-and-secrets.md#sem-login-tokens-gerados-via-lia-admin)).

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

Esse fluxo mostra um `speak` no fim, mas a fala é uma capability executada pelo Executor como qualquer outra (ver [`speak` como capability](tools-and-capabilities.md#speak-como-capability)) — pode aparecer intercalada com outras tools dentro do mesmo Workflow, não só como resposta final única.

### Wake Word

Modelo `hey_lia.onnx` treinado via OpenWakeWord. Inferência local no client — zero latência de rede para detecção.

### Sem streaming de áudio por agora

TTS recebe texto completo e gera áudio de uma vez. Streaming de texto conflita com entonação natural do TTS — prioridade para qualidade de voz. Revisitar quando necessário.

### Interação sem repetir a wake word

A wake word não precisa ser repetida a cada turno — ela inicia uma sessão de interação (`IDLE` → `ACTIVE` → `IDLE`), e turnos seguintes dentro da janela ativa não exigem reativação. Isso é majoritariamente responsabilidade do Client (VAD, estado de sessão, proximidade do dispositivo já são responsabilidades dele nesta mesma tabela) — fica registrado aqui para quando esse trabalho começar, sem desenho detalhado ainda.

Vale reforçar o mesmo princípio já estabelecido em [Identidade, Autenticação e Secrets](identity-auth-and-secrets.md): detectar que alguém está falando com a Lia (perceber a sessão ativa) não é o mesmo que autorizar uma ação — a camada de percepção (sessão de interação) permanece independente da camada de autorização.
