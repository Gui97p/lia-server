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
- **Registry em runtime** — quais capabilities um client conectado *agora de fato* implementa. Isso é estado efêmero, atrelado à conexão WebSocket viva, e vive **em memória**, nunca no Postgres — muda a cada connect/disconnect, e persistir isso em banco seria trabalho de escrita constante para algo que só existe enquanto a conexão existe.

O catálogo no Postgres elimina a necessidade de deploy só para *documentar* uma capability nova, mas não elimina a necessidade de código real implementando o comportamento no client (Rust) — isso continua sendo código, não dado. "Trivial via `lia-admin`" se aplica ao registro/descrição da capability, não a inventar comportamento sem escrever nada.

## Registry de capabilities

O Executor consulta o registry em runtime, que mapeia capabilities para implementations disponíveis no momento. Clients anunciam suas capabilities ao conectar, usando o formato do catálogo.

Capability declarada no handshake significa **"eu sei fazer isso"**, não "o valor disso agora". Isso importa para distinguir dois tipos de capability:

- **Estáticas por sessão** — o enum de valores é estável durante toda a conexão (ex: `openApp`, a lista de apps instalados não muda a cada segundo). Pode vir com o enum completo já no handshake.
- **Estado dinâmico** — o "valor" muda constantemente (ex: `activeWindows`, quais janelas estão abertas agora). Não faz sentido anunciar um snapshot no handshake, porque ele fica desatualizado imediatamente. O handshake só declara que o client **suporta** essa capability; o valor real é sempre consultado sob demanda, no momento em que o Executor precisa dele, usando o ciclo de vida normal de tool (`tool.request` → `tool.completed`) — nunca cacheado.

### Contrato do handshake

Uma versão anterior deste projeto (protótipo em Python, HTTP) usava um contrato onde cada chave do objeto `capabilities` era um enum de valores (`"openApp": [...]`), exceto uma chave especial `"default"` que guardava uma lista de nomes de capability sem enum. Isso mistura dois formatos incompatíveis sob o mesmo tipo (array de string) e não generaliza — quebra na primeira capability que precisar de mais de um parâmetro com enum.

Contrato atual: `capabilities` é uma lista de objetos, cada um com `name` e, opcionalmente, `params` (parâmetros com enum de valores válidos). Sem capability especial "default" — uma capability sem enum simplesmente não declara `params`.

```json
{
  "capabilities": [
    { "name": "openApp", "params": { "app": ["discord", "vscode", "spotify"] } },
    { "name": "activeWindows" },
    { "name": "moveWindow", "params": { "window": { "source": "activeWindows" } } },
    { "name": "exit" },
    { "name": "setClock" }
  ]
}
```

Esse formato é próximo do que APIs de function-calling de LLM (incluindo a Groq) já esperam para descrever tools — reaproveitar essa forma evita tradução estranha entre o schema do registry e o schema que vai para o Planner.

### Dependência entre capabilities (`source`)

Alguns parâmetros não têm enum estático nem estado consultável isoladamente — o valor válido depende do resultado de **outra** capability (ex: `moveWindow` recebe `window`, mas os nomes de janela válidos só existem chamando `activeWindows` primeiro). Isso é indicado no catálogo com `"source": "<nomeDaCapability>"` no lugar do enum.

Isso não exige nenhum mecanismo novo de execução — é o mesmo problema que o caso de uso 3 do roadmap já força a resolver ("baixe o relatório, compare com o anterior"): um step do Workflow usa o resultado de outro step. `moveWindow` vira dois steps no plano do Planner: um step chamando `activeWindows` (sem parâmetro), e um segundo step chamando `moveWindow` com `window` preenchido a partir do resultado do primeiro, e `DependsOn` apontando para ele (ver [`$fromStep` em Planner e Executor](planner-and-executor.md#referência-a-resultado-de-outro-step-fromstep)). O `source` no catálogo é só a pista que evita o Planner alucinar um valor em vez de gerar o step de descoberta.

### Identificação de janelas: nomenclatura canônica

O resultado de `activeWindows` precisa expor um campo `app` estruturado (não só o título cru da janela), para que o `match` do `$fromStep` seja uma igualdade exata em vez de uma correspondência aproximada de texto — nomes de processo variam por plataforma e por versão (`Discord.exe`, `DiscordUpdater`, `discordCanary`, etc.), e correspondência aproximada arrisca casar o processo errado silenciosamente.

O client já mantém, para implementar `openApp`, uma tabela de nome canônico → identificador real do processo/bundle para cada app do enum (é o mínimo necessário para saber o que executar quando pedem `openApp("discord")`). `activeWindows` reaproveita essa mesma tabela na direção inversa: para cada janela do sistema, verifica se o processo bate com algum identificador conhecido e, se sim, popula `app` com o nome canônico correspondente (o mesmo nome usado no enum de `openApp`). Processos que não correspondem a nenhum app conhecido simplesmente não recebem `app` — não são adivinhados por aproximação.

Consequência: o universo de apps identificáveis em `activeWindows` é o mesmo (finito, já curado) enum de `openApp` — não é necessário reconhecer todo processo do sistema operacional, só os que o client já sabe abrir.

## Protocolo de anúncio: handshake da conexão

Capabilities são anunciadas em uma mensagem única no handshake da conexão WebSocket — não há atualização dinâmica em runtime. Se um client precisa mudar suas capabilities (ex: instalou um novo plugin), a forma de refletir isso é reconectar.

**Reconectar é o caminho simples e suficiente**: como os clients já precisam suportar reconexão (quedas de rede são esperadas), reiniciar o handshake do zero (re-anunciar todas as capabilities) é mais simples e barato do que construir eventos incrementais de adicionar/remover tools de uma sessão já estabelecida. Não há necessidade real de complexidade adicional aqui.

## Detecção de desconexão: heartbeat nativo do protocolo

Frames de controle (ping/pong) nativos do protocolo WebSocket já cobrem a detecção de desconexão silenciosa — não é necessário heartbeat de aplicação por cima disso.

## Em aberto

- **Comunicação entre serviços futuros** — hoje é um monólito Go. Se no futuro surgir mais de um serviço server-side (ex: um serviço dedicado a STT/TTS separado do core), como eles se comunicam é uma decisão prematura — só vale revisitar quando a necessidade aparecer de verdade.
