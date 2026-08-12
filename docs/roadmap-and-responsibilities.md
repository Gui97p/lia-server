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

## Visão de Longo Prazo

- Reconhecimento de voz por perfil para identificação passiva
- Consciência social: quem está online, memórias de grupo, mensagens entre usuários
- Tools colaborativas entre usuários
- Suporte a IoT e dispositivos físicos (C++)
- Múltiplos assistentes especializados orquestrados pelo mesmo core
