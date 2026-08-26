# Roadmap e Responsabilidades

## Divisão de Responsabilidades

**Gui (server Go)**
- Core da IA (Planner, Executor, LLM, memória)
- PostgreSQL e modelos de dados
- Protocolo de comunicação e sessões
- Autenticação, autorização e níveis de confiança
- Áudio (STT e TTS)
- CLI de gestão de usuários
- Observabilidade e logs

**Yure (clients Rust)**
- Desktop App (wake word, áudio local, tools locais)
- CLI
- Discord Bot
- SDK de conexão com o server

## Roadmap

### Fase atual — descoberta por casos

Desenvolver a arquitetura através de casos reais, do simples ao complexo:

1. "Abra o Spotify."
2. "Abra o Spotify e coloque aquela playlist que eu estava ouvindo ontem."
3. "Baixe o relatório, analise, compare com o anterior e mande um resumo se houver mudança."
4. "Saí de casa. Apague as luzes, tranque a porta, faça backup, e me avise quando terminar. Se der erro, tente resolver."

Cada caso revela uma necessidade arquitetural real. Vários pontos "Em aberto" espalhados por esta documentação dependem justamente de casos como o 3 e o 4 aparecerem de verdade antes de serem fechados (ex: DAG no [Planner](planner-and-executor.md), busca semântica de memória).

### Fase beta

Quando o servidor estiver pronto para ser usado por um client externo em teste, o trabalho passa a acontecer em branches separadas da `main`, preservando-a estável. Esse é o ponto natural para revisitar decisões deliberadamente adiadas até então (ex: `pgvector` em [Memória](memory.md)).

## Deploy (ainda não decidido)

Direção discutida, não implementada: self-host num notebook, acessível de qualquer lugar via [Tailscale](https://tailscale.com/) (WireGuard, sem precisar expor porta/lidar com IP dinâmico). Postgres roda na própria máquina (mais simples que um serviço gerenciado tipo Supabase, sem round-trip de rede por query, funciona offline da internet) — backup via `pg_dump` agendado, copiado pra outra máquina.

Opções de nuvem avaliadas e descartadas por ora:
- **Render (free tier)**: spin-down depois de ociosidade (~15min) com cold start de 30-60s+ — ruim pra algo que devia responder na hora a um "hey Lia".

Se um dia migrar pra nuvem de verdade (ex: Fly.io/Railway pro binário + Supabase pro Postgres, dado free tier de 500MB), o problema de "instância sempre ligada sem pagar" volta a aparecer — self-host resolve isso de graça reaproveitando hardware parado.

Máquina rodando Linux headless (Ubuntu Server ou Debian, sem ambiente gráfico) em vez de Windows: menos processo competindo por recurso num hardware fraco, e principalmente sem reboot surpresa de Windows Update numa máquina que precisa ficar de pé 24/7 sem supervisão. Consumo de energia nesse cenário é baixo o bastante (poucos reais por mês) pra não ser o critério de decisão.

## Visão de Longo Prazo

- Reconhecimento de voz por perfil para identificação passiva
- Consciência social: quem está online, memórias de grupo, mensagens entre usuários
- Tools colaborativas entre usuários
- Suporte a IoT e dispositivos físicos (C++)
- Múltiplos assistentes especializados orquestrados pelo mesmo core
