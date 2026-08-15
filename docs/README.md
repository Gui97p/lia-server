# Documentação — Lia Server

Documentação viva do server da Lia. Cada arquivo cobre um domínio específico do sistema. Onde ainda não há decisão fechada, o próprio arquivo tem uma seção **Em aberto** ao final — não existe mais uma pasta separada de "decisões pendentes": o que está decidido e o que ainda não está vivem juntos, no lugar onde fazem sentido.

## Índice

- [Visão e Princípios](vision-and-principles.md) — o que é a Lia, princípios de design, stack de linguagens
- [Identidade, Autenticação e Secrets](identity-auth-and-secrets.md) — níveis de confiança, JWT, hashing de senha, gestão de usuários, matriz de capability × trust level
- [Memória](memory.md) — escopos, injeção de contexto, tools de memória, proveniência e segurança
- [Planner e Executor](planner-and-executor.md) — separação planejamento/execução, estrutura de Workflow, replanning
- [Tasks e Eventos](tasks-and-events.md) — máquina de estados, arquitetura de eventos, retries, recuperação após reboot
- [Tools e Capabilities](tools-and-capabilities.md) — ciclo de vida de tools, registry, protocolo de reconexão
- [Transporte e Áudio](transport-and-audio.md) — WebSocket, sessões/Hub multi-device, rotas HTTP, fluxo de áudio, wake word
- [Banco de Dados](database.md) — schema Postgres, criptografia de secrets em repouso
- [Observabilidade](observability.md) — logging estruturado, redação de segredos
- [Estrutura do Repositório](repo-structure.md) — organização de pastas; `transport/` vs `session/`
- [Roadmap e Responsabilidades](roadmap-and-responsibilities.md) — divisão Gui/Yure, casos de uso, visão de longo prazo
