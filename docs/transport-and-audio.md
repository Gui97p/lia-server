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

Protocolo definido: **WebSocket puro** (`github.com/coder/websocket` no server — fork mantido pela Coder desde ago/2024, quando o autor original do `nhooyr.io/websocket` passou o projeto adiante; mesma API, só muda o import path), com mensagens em JSON (`event`/`payload`) definidas pelo próprio projeto. A comunicação é bidirecional e persistente — não HTTP puro.

Primeira mensagem da conexão: evento `auth`, com `token` (JWT) e `capabilities` (array de nomes das tools que este client implementa). Schema e metadado das tools vêm do catálogo no Postgres — o handshake só declara suporte. Detalhes em [Tools e Capabilities](tools-and-capabilities.md#contrato-do-handshake).

Resposta de sucesso: `auth.ok` com `{ "conn_id": "<uuid>" }` — identificador da **conexão** (não do user). Útil para debug e, no futuro, targeting explícito; no fluxo feliz de `tool.request` o server escolhe a conn (ver [multi-device](#sessões-e-multi-device)).

### Sessões e multi-device

Unidade de runtime = **conexão** (`ConnID`), não o user. O mesmo `userId` pode ter várias conns ativas (PC, notebook, bot, etc.).

- `Session` — estado autenticado da conn: user, trust, keys de provider decifradas (`providers.Providers`, um mapa — não uma key fixa, já que o usuário pode ter Groq, Gemini, ambos ou nenhum), capabilities anunciadas, e um `Writer` para enviar eventos (a borda WS captura `conn` + mutex de escrita; o resto do sistema só vê `Session`). O handshake decifra o que existir e descarta silenciosamente qualquer key que falhe ao decifrar — só falha a autenticação se **nenhuma** key sobrar utilizável.
- `Hub` — registry em memória `ConnID → *Session` (`Register` / `Unregister` / `FindByUser` / `FindByID`). Mutex só no mapa; não segurar durante I/O.
- Pacote: `internal/session`. Ciclo de vida do WS (Accept → handshake → Register → loop → Unregister) fica em `internal/transport`.

Roteamento de tool (combinado, ainda não codificado): não é “manda pro user” — é “manda pra uma conn desse user que anunciou a capability”. MVP: preferir a conn de origem do turno; senão qualquer conn do user com a cap.

Timeout só no handshake (ex.: 5s); o loop da sessão não usa esse deadline.

### Client vs Device

Duas coisas que hoje o handshake trata igual, mas são conceitualmente diferentes:

- **Client** — quem *fala com* a Lia. Tem microfone/speaker, sessão de conversa, recebe `message.reply`. Ex: o app desktop/mobile, o cliente Rust.
- **Device** — quem *recebe ordens* da Lia, mas nunca conversa. Ex: um ESP32 controlando uma luz/tomada. Nunca deveria receber `message`/`message.reply` — só existe pro ciclo `tool.request`/`tool.completed`.

Hoje o `auth` não distingue os dois — qualquer conexão que anuncia `capabilities` é tratada como uma sessão completa. O desenho pra separar:

- O handshake ganha um campo de papel (`role: "client" | "device"`), mudando como o server trata a conexão dali pra frente (Device nunca é origem de um turno de conversa).
- Device precisa de **nome persistente**, não só `ConnID` efêmero — pra "liga a luz da sala" ser roteável pro dispositivo certo entre reconexões. Caminho preferido: um catálogo em Postgres (`devices(id, user_id, name, description)`), registrado via `lia-admin` — mesmo padrão do catálogo de capabilities (ver [Tools e Capabilities](tools-and-capabilities.md#catálogo-de-capabilities-postgres-vs-registry-em-runtime-in-memory)), o handshake só manda um `device_id` que o server resolve.
- O `Planner` precisa saber quais devices existem — mesmo padrão já usado pra "outras tasks em andamento" (`extraContext`, mensagem de sistema separada, nunca no histórico direto): uma lista "dispositivos disponíveis: nome + capabilities" injetada por turno.
- Roteamento pra um device nomeado usa um campo `target_device` no `Step` (ver [estrutura do Workflow](planner-and-executor.md#estrutura-do-workflow)), preenchido pelo Planner quando o usuário nomeia um device explicitamente. Sem nome e só um device com a capability → resolve sozinho. Ambíguo (dois devices, mesma capability, sem nome dito) → mesmo tratamento que `$fromStep` já tem pra zero/múltiplos resultados: falha sem retry, escala pra replanning, pode virar pergunta de desambiguação ao usuário.
- `Hub` ganha um lookup a mais (`FindByUserAndDevice(userID, deviceName)`) — extensão pequena, não redesenho.

**Server fala direto com o Device, não via um Client intermediário.** Considerado e descartado: Device parear localmente com um Client específico (ex: via Bluetooth/rede local) que repassa o comando pro server. O problema é que isso amarra a disponibilidade do device a um Client específico estar online — quebra a premissa de "qualquer Client, qualquer device" que já é central no projeto. Device conectar direto ao server (mesmo mecanismo de `Session`/`Hub`/`tool.request` que já existe pra Client, só com `role: "device"`) desacopla os dois ciclos de vida e reaproveita tudo que já está construído. Custo real: firmware do device fica um pouco mais pesado (WS+TLS num ESP32 em vez de MQTT puro), mas é caminho batido (ESPHome faz algo parecido).

**Autenticação do Device**: um ESP32 não tem como digitar/guardar um fluxo de JWT com a mesma facilidade que um Client. Ideia (depende de o deploy ser mesmo via Tailscale, ver [Deploy](roadmap-and-responsibilities.md#deploy-ainda-não-decidido)): dar ao Device sua própria identidade Tailscale (chave de dispositivo, não conta de usuário) e o server confiar na conexão vir de um IP conhecido da tailnet — evita ter que gerenciar segredo de longa duração numa flash de poucos MB.

Rotas atuais do transport: `GET /health`, `GET /ws`, `POST /audio/transcribe`, `POST /audio/speak`.

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

### Modo `wait` do `speak` e o evento `message.done`

O servidor não sabe, por si só, quando o cliente termina de reproduzir o áudio de uma fala — `message.reply` (WS) e `POST /audio/speak` (HTTP) são desacoplados, o cliente pode pedir o áudio quando quiser. Pra o modo `wait` do `speak` (que deve bloquear o próximo step do plano até a fala terminar) funcionar de verdade, o protocolo tem um evento extra:

- Todo `message.reply` carrega um `step_id`.
- Quando o `speak` está em modo `wait`, o `Executor` bloqueia (`session.WaitForSpeechDone`) esperando o cliente mandar `message.done` com esse mesmo `step_id` — ou um fallback estimado por tamanho do texto (~15 chars/segundo, piso 1s, teto 20s), pra clients que ainda não implementam o ack real.
- Mandar `message.done` fora desse contexto (modo `fire_and_forget`, ou um `step_id` já resolvido) é inofensivo — o server só ignora, não tem `step_id` nenhum esperando por ele.

Ver o schema completo em `docs/asyncapi/asyncapi.yaml` (mensagens `messageReply`/`messageDone`).

### Wake Word

Modelo `hey_lia.onnx` treinado via OpenWakeWord. Inferência local no client — zero latência de rede para detecção.

### Sem streaming de áudio por agora

TTS recebe texto completo e gera áudio de uma vez. Streaming de texto conflita com entonação natural do TTS — prioridade para qualidade de voz. Revisitar quando necessário.

### Interação sem repetir a wake word

A wake word não precisa ser repetida a cada turno — ela inicia uma sessão de interação (`IDLE` → `ACTIVE` → `IDLE`), e turnos seguintes dentro da janela ativa não exigem reativação. Isso é majoritariamente responsabilidade do Client (VAD, estado de sessão, proximidade do dispositivo já são responsabilidades dele nesta mesma tabela) — fica registrado aqui para quando esse trabalho começar, sem desenho detalhado ainda.

Vale reforçar o mesmo princípio já estabelecido em [Identidade, Autenticação e Secrets](identity-auth-and-secrets.md): detectar que alguém está falando com a Lia (perceber a sessão ativa) não é o mesmo que autorizar uma ação — a camada de percepção (sessão de interação) permanece independente da camada de autorização.
