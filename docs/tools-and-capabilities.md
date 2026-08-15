# Tools e Capabilities

## Ciclo de Vida de Tools

Tools têm ciclo de vida próprio — cancelar uma Task não cancela automaticamente uma tool já em execução no client.

```
tool.request
     │
     ▼
tool.accepted
     │
     ▼
tool.started
     │
     ▼
tool.progress (opcional)
     │
     ▼
tool.completed
```

## Separação de capabilities

```
Capability    →  o que precisa ser feito ("abrir aplicativo")
Tool          →  abstração que representa essa capability
Implementation →  como é feito em cada plataforma (PowerShell, AppleScript, etc.)
```

## Catálogo de capabilities (Postgres) vs Registry em runtime (in-memory)

Duas coisas que parecem a mesma coisa, mas não são:

- **Catálogo** — a definição de cada capability: nome, descrição (usada para explicar a tool ao LLM), schema de parâmetros, trust level exigido (ver [Identidade e Secrets](identity-auth-and-secrets.md)), retry policy default. É metadado relativamente estável, então mora no **Postgres**. É isso que torna "criar uma capability nova" trivial via `lia-admin` (`lia-admin tools register --name moveWindow --description "..." --trust authenticated`), sem precisar de deploy novo do server para o Planner passar a *conhecer* a tool.
- **Registry em runtime** — quais capabilities um client conectado *agora de fato* implementa. Isso é estado efêmero, atrelado à conexão WebSocket viva (`Session.Capabilities` no `Hub` em `internal/session`), e vive **em memória**, nunca no Postgres — muda a cada connect/disconnect, e persistir isso em banco seria trabalho de escrita constante para algo que só existe enquanto a conexão existe. Com multi-device, a pergunta do Executor é “qual **conn** deste user tem a capability?”, não só “o user tem?” — ver [Sessões e multi-device](transport-and-audio.md#sessões-e-multi-device).

O catálogo no Postgres elimina a necessidade de deploy só para *documentar* uma capability nova, mas não elimina a necessidade de código real implementando o comportamento no client (Rust) — isso continua sendo código, não dado. "Trivial via `lia-admin`" se aplica ao registro/descrição da capability, não a inventar comportamento sem escrever nada.

## Registry de capabilities

O Executor consulta o registry em runtime, que mapeia capabilities para implementations disponíveis no momento. No connect, o client só anuncia **quais** capabilities ele implementa; **como** cada uma se parece (descrição, schema de params, `source`, trust) vem do catálogo no Postgres. O server cruza os dois: nomes anunciados ∩ catálogo → o que o Planner/Executor podem usar nesta sessão.

Capability declarada no handshake significa **"eu sei fazer isso"** — não schema, não enum de valores, não snapshot de estado. Valores que mudam (janelas abertas, apps instalados, etc.) nunca vão no handshake: ou são resolvidos no client na hora de executar a tool, ou são consultados sob demanda via ciclo de vida normal (`tool.request` → `tool.completed`).

### Contrato do handshake

O anúncio entra no mesmo evento `auth` da conexão WebSocket (mensagem única — ver [Protocolo de anúncio](#protocolo-de-anúncio-handshake-da-conexão)). `capabilities` é um **array de nomes** (strings). Nada de objetos com `params`, enums de apps, nem metadado de catálogo repetido pelo client.

```json
{
  "event": "auth",
  "payload": {
    "token": "<jwt>",
    "capabilities": ["openApp", "activeWindows", "moveWindow", "exit", "setClock"]
  }
}
```

Por que não reenviar o contrato da tool no handshake: o schema já está no Postgres. Repeti-lo no client cria duas fontes de verdade e diverge (client desatualizado vs catálogo). Enums locais (ex.: apps instalados) também não cabem aqui — ficam stale na hora em que o usuário instala ou remove algo; o client resolve na execução (ou falha a tool se o app não existir).

Nomes desconhecidos do catálogo (typo, capability ainda não registrada via `lia-admin`) devem ser rejeitados ou ignorados de forma explícita no server — não silenciar sem regra. Lista vazia é válida (client autenticado sem tools locais).

O formato que o Planner/LLM vê (function-calling, descrição, params) é montado **no server** a partir do catálogo, filtrado pelo registry da sessão — não é o payload do handshake.

### Dependência entre capabilities (`source`)

Alguns parâmetros não têm enum estático nem estado consultável isoladamente — o valor válido depende do resultado de **outra** capability (ex: `moveWindow` recebe `window`, mas os nomes de janela válidos só existem chamando `activeWindows` primeiro). Isso é indicado **no catálogo** (Postgres) com `"source": "<nomeDaCapability>"` no lugar do enum — nunca no handshake.

Isso não exige nenhum mecanismo novo de execução — é o mesmo problema que o caso de uso 3 do roadmap já força a resolver ("baixe o relatório, compare com o anterior"): um step do Workflow usa o resultado de outro step. `moveWindow` vira dois steps no plano do Planner: um step chamando `activeWindows` (sem parâmetro), e um segundo step chamando `moveWindow` com `window` preenchido a partir do resultado do primeiro, e `DependsOn` apontando para ele (ver [`$fromStep` em Planner e Executor](planner-and-executor.md#referência-a-resultado-de-outro-step-fromstep)). O `source` no catálogo é só a pista que evita o Planner alucinar um valor em vez de gerar o step de descoberta.

### Identificação de janelas: nomenclatura canônica

O resultado de `activeWindows` precisa expor um campo `app` estruturado (não só o título cru da janela), para que o `match` do `$fromStep` seja uma igualdade exata em vez de uma correspondência aproximada de texto — nomes de processo variam por plataforma e por versão (`Discord.exe`, `DiscordUpdater`, `discordCanary`, etc.), e correspondência aproximada arrisca casar o processo errado silenciosamente.

O client já mantém, para implementar `openApp`, uma tabela local de nome canônico → identificador real do processo/bundle (é o mínimo necessário para saber o que executar quando pedem `openApp("discord")`). Essa tabela é detalhe de implementação do client — não é anunciada no handshake nem é a fonte de schema do server. `activeWindows` reaproveita a mesma tabela na direção inversa: para cada janela do sistema, verifica se o processo bate com algum identificador conhecido e, se sim, popula `app` com o nome canônico correspondente. Processos que não correspondem a nenhum app conhecido simplesmente não recebem `app` — não são adivinhados por aproximação.

Consequência: o universo de apps identificáveis em `activeWindows` é o mesmo conjunto finito que o client já sabe abrir via `openApp` — não é necessário reconhecer todo processo do sistema operacional.

## `speak` como capability

A comunicação da Lia (fala, e futuramente `display`/`notify`) é uma capability como qualquer outra, executada pelo Executor através do mesmo ciclo de vida de tool já documentado (`tool.request` → `tool.completed`) — não um efeito implícito da resposta da LLM. Isso permite intercalar fala com outras ações no mesmo Workflow (ex: `speak("vou verificar a disponibilidade")` → `makePhoneCall(...)` → `speak("consegui às 20h")`), representando explicitamente a ordem das ações em vez de um texto único no fim.

Sem streaming por agora — `speak(text)` é uma chamada única, bloqueante, igual a qualquer outra tool (`tool.request`/`tool.completed`, sem eventos `speak.chunk`), consistente com a decisão já registrada em [Transporte e Áudio](transport-and-audio.md#sem-streaming-de-áudio-por-agora). Se streaming vier a ser necessário, só é viável quebrando por frase/oração completa (o TTS precisa da oração inteira pra acertar entonação) — não por token. Fica anotado, sem implementar agora.

### `speak` é uma capability "core", nunca sujeita a discovery

Conforme o catálogo de capabilities crescer, pode fazer sentido não injetar tudo no prompt do Planner de uma vez — descobrir capabilities relevantes sob demanda (o mesmo princípio já aplicado a memória em [MVP: injetar tudo](memory.md#mvp-injetar-tudo): injeta tudo até o volume doer, só então adiciona descoberta/filtragem). Isso não é necessário agora (poucas capabilities), mas quando existir, `speak` — e outras capabilities fundamentais a qualquer turno do Agent — não podem correr o risco de ficar de fora de um filtro de relevância. O catálogo já nasce com um flag (`core: true`) para essas capabilities, que qualquer mecanismo futuro de descoberta é obrigado a sempre incluir, independente de pontuação de relevância.

## Protocolo de anúncio: handshake da conexão

Capabilities (só os nomes) são anunciadas na mensagem única de `auth` do WebSocket — não há atualização dinâmica em runtime nem eventos incrementais de add/remove. Mudança no *conjunto* de capabilities que o client implementa (ex: novo plugin que passa a expor tools) → reconectar e refazer o handshake.

Instalar/remover um app no SO **não** exige reconectar: apps não entram no anúncio; o client resolve na execução de `openApp` (ou equivalente).

**Reconectar é o caminho simples e suficiente** para mudanças no conjunto de capabilities: os clients já precisam suportar reconexão (quedas de rede são esperadas), e reiniciar o handshake do zero é mais barato do que sincronizar diffs numa sessão viva.

## Detecção de desconexão: heartbeat nativo do protocolo

Frames de controle (ping/pong) nativos do protocolo WebSocket já cobrem a detecção de desconexão silenciosa — não é necessário heartbeat de aplicação por cima disso.

## Em aberto

- **Comunicação entre serviços futuros** — hoje é um monólito Go. Se no futuro surgir mais de um serviço server-side (ex: um serviço dedicado a STT/TTS separado do core), como eles se comunicam é uma decisão prematura — só vale revisitar quando a necessidade aparecer de verdade.
- **Skills e Capability Discovery** — conforme o número de capabilities crescer, agrupar Tools relacionadas em "Skills" de nível mais alto (ex: uma Skill `reservarRestaurante` orquestrando várias Tools) e descobrir capabilities relevantes sob demanda, em vez de injetar o catálogo inteiro no prompt do Planner. Ideia validada (é essencialmente o mesmo mecanismo usado por ferramentas de busca de tool sob demanda em agentes de LLM hoje), mas deliberadamente fora de escopo do MVP — poucas capabilities existem ainda para essa dor se manifestar de verdade.
