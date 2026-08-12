# Identidade, Autenticação e Secrets

Três conceitos distintos que não devem ser confundidos:

```
Identificação      →  "Provavelmente é Guilherme"
Autenticação       →  "Essa sessão pertence a Guilherme"
Autorização        →  "Guilherme pode fazer isso?"
```

## Identificação por voz

A Lia reconhece quem está falando através do perfil de voz. Isso é identificação — não autenticação. Mais de uma pessoa pode usar o mesmo dispositivo.

```
Áudio recebido
      │
      ▼
Reconhecimento de voz
      │
      ├── Reconheceu → usa perfil do usuário identificado
      │
      └── Não reconheceu → usuário anônimo (permissões mínimas)
```

O reconhecimento de voz contribui para um **nível de confiança**, não para um booleano `authenticated = true/false`. Reconhecimento de voz pode ser enganado por replay de áudio gravado ou voice cloning — por isso ele nunca é suficiente, sozinho, para autorizar ações sensíveis (ver matriz abaixo).

## Níveis de confiança

Em vez de autenticado/não autenticado, o sistema opera com níveis:

```
ANONYMOUS     → voz não reconhecida ou sem voz
IDENTIFIED    → voz reconhecida (provavelmente é X)
AUTHENTICATED → sessão autenticada por credencial forte
TRUSTED       → autenticado + contexto conhecido + histórico
```

### Matriz de capability × trust level (rascunho)

Primeira versão para validar contra casos de uso reais — não é definitiva:

| Capability | ANONYMOUS | IDENTIFIED | AUTHENTICATED | TRUSTED |
|---|---|---|---|---|
| Conversa simples / perguntas gerais | ✅ | ✅ | ✅ | ✅ |
| Ler memória USER própria | ❌ | ❌ | ✅ | ✅ |
| Ler memória GROUP | ❌ | ❌ | ✅ | ✅ |
| Salvar/editar memória | ❌ | ❌ | ✅ | ✅ |
| Tools sem efeito no mundo real (abrir Spotify, consultar algo) | ❌ | ✅ (com aviso) | ✅ | ✅ |
| Tools com efeito físico/irreversível (trancar porta, deletar arquivo) | ❌ | ❌ | ✅ | ✅ |
| Gestão de usuários/admin | ❌ | ❌ | ❌ | apenas via `lia-admin` local, nunca via API |

A linha mais sensível é a de tools com efeito físico — merece revisão constante conforme tools reais forem implementadas.

## Modelo de dados

```sql
users        (id, username, password_hash, groq_api_key_encrypted, created_at)
groups       (id, name, created_at)
user_groups  (user_id, group_id)
voice_profiles (id, user_id, embedding, created_at)
devices      (id, user_id, device_type, token_version, trusted, created_at)
```

### Hashing de senha

`bcrypt` (`golang.org/x/crypto/bcrypt`), custo 12+. Simples, madura, suficiente para este projeto — não há necessidade de `argon2id` na escala e no modelo de ameaça atuais.

### `groq_api_key` cifrada em repouso

A API key do Groq de cada usuário é cifrada antes de ser gravada no banco — um dump do banco sozinho não deve expor as keys. Cifragem na camada da aplicação (AES-GCM) usando uma chave separada do `JWT_SECRET` (ex: `ENCRYPTION_KEY` no `.env`), não `pgcrypto` do banco — assim, mesmo com acesso ao banco, é preciso também o segredo da aplicação para decifrar.

A key **não** é embutida no payload do JWT (ver seção JWT abaixo) — o servidor busca a key cifrada do banco por `user_id` sempre que precisa chamar o Groq em nome do usuário, e decifra em memória no momento do uso.

## Gestão de usuários

Operações via CLI Go local — sem rota pública de registro:

```bash
./lia-admin users create --username yure
./lia-admin users delete --username yure
./lia-admin users set-key --username yure --key gsk_...   # cifra antes de gravar

./lia-admin devices list --username yure
./lia-admin devices revoke --device-id <uuid>              # bump no token_version só desse device
```

## JWT

Login via HTTP retorna JWT de longa duração (1 ano). Sem refresh tokens — decisão final, com a mitigação abaixo.

```json
{
  "user_id": "uuid",
  "username": "gui",
  "device_id": "uuid",
  "group_ids": ["amigos"],
  "trust_level": "authenticated",
  "token_version": 3
}
```

Note que `groq_api_key` **não** está no payload — só `user_id`, que o servidor usa para buscar a key cifrada quando necessário.

### Revogação via `token_version`

Cada usuário tem uma coluna `token_version` em `users`. Todo JWT emitido carrega o `token_version` vigente no momento do login. A validação do token, além da assinatura, checa se `token_version` do JWT bate com o valor atual no banco.

Revogar todas as sessões de um usuário (ex: suspeita de token roubado) é um único `UPDATE users SET token_version = token_version + 1 WHERE id = ...` via `lia-admin` — instantâneo, sem precisar de denylist de tokens que cresce sem limite. Isso resolve o cenário "peguei o token de madrugada e só consigo agir de manhã": a revogação acontece assim que alguém perceber, não depende de o token expirar.

Por que não passkeys/WebAuthn: WebAuthn é pensado para navegador conversando com um authenticator de plataforma (TouchID, Windows Hello, chave física). Os clients daqui (desktop Rust, CLI, bot de Discord) não têm acesso natural à API de WebAuthn do browser — seria complexidade nova sem ganho real. JWT + `token_version` é suficiente para o modelo de ameaça e a base de usuários deste projeto.

## Autorização por dispositivo

`token_version` no usuário revoga *todas* as sessões dele de uma vez — todos os dispositivos. Isso é grosso demais: perder o notebook não deveria exigir deslogar celular, CLI e bot de Discord também.

Por isso, `token_version` também existe por dispositivo (tabela `devices`, coluna própria), e cada JWT carrega `device_id` além de `user_id`. A validação do token passa a checar duas coisas: `token_version` do usuário **e** `token_version` do dispositivo. Revogar um dispositivo específico é um único `UPDATE devices SET token_version = token_version + 1 WHERE id = ...` via `lia-admin devices revoke` — sem afetar os demais dispositivos do mesmo usuário. É essencialmente a mesma mecânica do `token_version` de usuário, só que com granularidade menor, e por isso quase não adiciona custo de implementação.

Isso também abre espaço para "contexto conhecido" (o que hoje diferencia `AUTHENTICATED` de `TRUSTED`) incluir o dispositivo: um dispositivo marcado `trusted: true` (ex: o próprio desktop de casa) pode elevar a sessão a `TRUSTED`, enquanto um dispositivo novo/não reconhecido fica em `AUTHENTICATED` mesmo com credencial válida, até acumular histórico.

### Canal como sinal de risco (adiado)

Nem todo canal de origem merece o mesmo teto de confiança — um comando vindo do app desktop (WebSocket autenticado, dispositivo que o usuário fisicamente possui) é mais confiável que o mesmo comando chegando via Discord (mensagem em canal de grupo, mais fácil de falsificar se a conta de alguém for comprometida). A ideia é a matriz de capability × trust level também considerar o dispositivo/canal de origem — por exemplo, tools com efeito físico só executarem se o comando vier de um dispositivo pré-registrado como confiável, independente do trust level da sessão.

Isso fica deliberadamente **fora de escopo por enquanto** — só vale desenhar de verdade quando o caso de uso 4 do roadmap (ações físicas: trancar porta, backup) estiver sendo implementado de fato, com um canal real de menor confiança (Discord) para testar contra.
