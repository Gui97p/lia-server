# Visão e Princípios

## Visão

A Lia não é um chatbot. É uma plataforma de inteligência pessoal — uma única IA que acompanha o usuário em todos os dispositivos, mantém memória contínua, reconhece quem está falando, e é capaz de planejar e executar tarefas complexas com múltiplas etapas, falhas e recuperações.

A referência é o Jarvis: uma inteligência que age, não apenas responde.

```
                      LIA
                       │
          ┌────────────┼────────────┐
          │            │            │
       Usuário A    Usuário B    Usuário C
          │
    ┌─────┴─────┐
  Desktop     Mobile
```

A identidade da Lia é única. Usuários, sessões, permissões e memórias possuem seus próprios contextos.

Uso interno — número de usuários é mínimo (inicialmente Gui e Yure). Isso influencia várias decisões de escopo ao longo desta documentação (ex: sem rota pública de registro, sem necessidade de infraestrutura para múltiplas instâncias). Não influencia o padrão de segurança: a Lia executa ações no mundo real (trancar porta, backups, etc.), e isso justifica cuidado independente do tamanho da base de usuários.

## Princípios

- **Fluidez** — a experiência deve ser natural. A Lia não esquece, não recomeça, não perde contexto ao trocar de device.
- **Casos primeiro** — a arquitetura é descoberta através de casos de uso reais, do simples ao complexo. Não desenhamos o que não precisamos ainda.
- **Falhas são normais** — o sistema é projetado para lidar com falhas como parte do fluxo, não como exceções arquiteturais.
- **Separação clara** — servidor é responsável pelo significado e inteligência. Cliente é responsável pela experiência de interação em tempo real.

## Stack de Linguagens

A linguagem é escolhida pelo que faz mais sentido em cada contexto:

| Componente | Linguagem | Motivo |
|---|---|---|
| Server (core) | Go | Concorrência nativa, performance, binário único |
| Clients (desktop, CLI, Discord) | Rust | Performance, baixo consumo, nativo |
| IoT / hardware | C++ | Acesso direto ao hardware |
| Treinamento de modelos (wake word) | Python | Ecossistema de ML sem alternativa |

Este repositório contém apenas o server (Go). Clients, treinamento de wake word e demais componentes vivem em repositórios próprios — ver [Estrutura do Repositório](repo-structure.md).
