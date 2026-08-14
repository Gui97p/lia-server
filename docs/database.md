# Banco de Dados

PostgreSQL como banco principal. Suporta múltiplas conexões simultâneas e prepara para `pgvector` no futuro (ver [Memória](memory.md#mvp-injetar-tudo)).

```sql
users          (id, username, groq_api_key_encrypted, created_at, token_version)
groups         (id, name, created_at)
user_groups    (user_id, group_id)
voice_profiles (id, user_id, embedding, created_at)
messages       (id, user_id, role, content, created_at)
memories       (id, user_id, group_id, scope, fact, category, created_at)
behavior_rules (id, user_id, rule, created_at)
tasks          (id, user_id, state, workflow, trigger_type, authorized_trust_level, created_at, updated_at)
```

## Secrets em repouso

`groq_api_key_encrypted` é cifrada na camada da aplicação (AES-GCM, chave em `ENCRYPTION_KEY`, separada de `JWT_SECRET`) antes de ser gravada — ver [detalhes em Identidade e Secrets](identity-auth-and-secrets.md#groq_api_key-cifrada-em-repouso). Um dump do banco sozinho não deve ser suficiente para recuperar as keys.

Sem senha, e portanto sem `password_hash` — tokens são gerados via `lia-admin` (ver [Sem login, tokens gerados via `lia-admin`](identity-auth-and-secrets.md#sem-login-tokens-gerados-via-lia-admin)).

## Isolamento de escopo de memória

A tabela `memories` só protege PRIVATE/USER/GROUP de vazamento entre usuários se **toda query** filtrar corretamente por `scope`/`user_id`/`group_id`. Um único ponto de código que esqueça esse filtro vaza dado que não deveria ser visível.

Para o volume e número de pessoas mexendo no código hoje (Gui, possivelmente Yure depois), revisão disciplinada da camada de aplicação já reduz bastante o risco. Se o projeto crescer (mais devs, mais superfícies de código tocando `memories`), reforçar com Row-Level Security do Postgres passa a valer a complexidade adicional — não antes disso.

## `behavior_rules` é separada de `memories`

`behavior_rules` guarda como a Lia deve se comportar (tom, princípios de conduta) — ver [Memória: regras de comportamento não são memória](memory.md#regras-de-comportamento-não-são-memória). Só é editável via `lia-admin`/rota administrativa protegida, nunca pelas tools `saveMemory`/`updateMemory`. Separar essa tabela de `memories` no schema (em vez de só uma `category` diferente na mesma tabela) já impede, estruturalmente, que o mesmo código que escreve memória comum escreva regra de comportamento por engano.
