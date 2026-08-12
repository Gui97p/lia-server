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

## Registry de capabilities

O Executor consulta um registry que mapeia capabilities para implementations disponíveis no momento. Clients anunciam suas capabilities ao conectar.

```json
{
  "capabilities": {
    "openApp": ["discord", "vscode", "spotify"],
    "activeWindows": ["discord", "vscode"],
    "default": ["exit", "setClock"]
  }
}
```

## Protocolo de anúncio: handshake da conexão

Capabilities são anunciadas em uma mensagem única no handshake da conexão WebSocket — não há atualização dinâmica em runtime. Se um client precisa mudar suas capabilities (ex: instalou um novo plugin), a forma de refletir isso é reconectar.

**Reconectar é o caminho simples e suficiente**: como os clients já precisam suportar reconexão (quedas de rede são esperadas), reiniciar o handshake do zero (re-anunciar todas as capabilities) é mais simples e barato do que construir eventos incrementais de adicionar/remover tools de uma sessão já estabelecida. Não há necessidade real de complexidade adicional aqui.

## Detecção de desconexão: heartbeat nativo do protocolo

Frames de controle (ping/pong) nativos do protocolo WebSocket já cobrem a detecção de desconexão silenciosa — não é necessário heartbeat de aplicação por cima disso.

## Em aberto

- **Comunicação entre serviços futuros** — hoje é um monólito Go. Se no futuro surgir mais de um serviço server-side (ex: um serviço dedicado a STT/TTS separado do core), como eles se comunicam é uma decisão prematura — só vale revisitar quando a necessidade aparecer de verdade.
