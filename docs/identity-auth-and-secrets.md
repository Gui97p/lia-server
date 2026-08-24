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
users        (id, username, created_at, token_version)
providers    (user_id, provider, encrypted_key, created_at, updated_at)  -- PK (user_id, provider)
voice_profiles (id, user_id, embedding, created_at)
```

Sem senha — ver [Sem login, tokens gerados via `lia-admin`](#sem-login-tokens-gerados-via-lia-admin) abaixo.

### `groq_api_key` cifrada em repouso

A API key de cada provider (Groq, Gemini, …) configurado por um usuário é cifrada antes de ser gravada no banco — um dump do banco sozinho não deve expor as keys. Cifragem na camada da aplicação (AES-GCM) usando uma chave separada do `JWT_SECRET` (ex: `ENCRYPTION_KEY` no `.env`), não `pgcrypto` do banco — assim, mesmo com acesso ao banco, é preciso também o segredo da aplicação para decifrar. As keys vivem na tabela `providers`, uma linha por `(user_id, provider)` — ver [Banco de Dados](database.md#providers-uma-linha-por-provider-não-uma-coluna-por-provider).

Nenhuma key é embutida no payload do JWT (ver seção JWT abaixo). No handshake do WebSocket o server busca todas as keys cifradas do usuário, decifra as que existirem e guarda o plaintext (`providers.Providers`, um mapa) na `Session` da conexão (em memória de processo, só enquanto a conn vive) — uma key que falha ao decifrar é descartada silenciosamente, sem derrubar a autenticação; ela só falha de fato se **nenhuma** key sobrar utilizável. Threat model: dump do Postgres sozinho não basta; quem já tem RAM + `ENCRYPTION_KEY` do processo já venceu de qualquer forma. Não logar `Session` / o campo da key. Se uma key for trocada via `lia-admin`, a sessão ativa pode ficar com a key antiga até reconnect (aceitável no MVP).

## Gestão de usuários

Operações via CLI Go local — sem rota pública de registro:

```bash
./lia-admin users create --username yure
./lia-admin users delete --username yure
./lia-admin users set-key --username yure --provider groq --key gsk_...   # cifra antes de gravar
./lia-admin users set-key --username yure --provider groq --reset        # remove a key desse provider

./lia-admin tokens generate --username yure   # gera o JWT do usuário
```

## JWT

### Sem login, tokens gerados via `lia-admin`

Não existe senha, nem rota de login. Como não há registro público (gestão de usuários já é só via `lia-admin` local), estender esse mesmo princípio para autenticação elimina uma superfície de ataque inteira: sem senha, não há o que forçar por brute-force, phishing, nem rota de login pra proteger na rede.

`lia-admin` gera o JWT diretamente (com acesso direto ao banco, já sabe o `token_version` vigente) e o token é distribuído manualmente, fora de banda — não existe `POST /auth/login`. A credencial forte que define `AUTHENTICATED` na matriz não é mais "digitou a senha certa", é "possui um JWT pré-provisionado pelo `lia-admin`" — um JWT gerado por máquina tem entropia bem maior que qualquer senha memorizável.

**Um token por usuário, não por dispositivo.** O mesmo JWT é usado em todos os dispositivos daquele usuário — é a identidade dele, não a identidade de um device específico. Considerou-se ter um token por dispositivo (revogação mais granular, só invalidar o aparelho comprometido), mas para 2 usuários com poucos dispositivos cada, isso é complexidade que não se paga: reemitir um token novo pros outros dispositivos depois de um incidente raro não é um fardo real, e a alternativa (gerenciar token por dispositivo, uma tabela `devices`, comandos extras de `lia-admin`) é custo certo por um benefício que praticamente nunca vai ser usado.

Distribuição do token na prática: usar um canal que não retém texto plano indefinidamente (ex: mensagem descartável, entrega local) em vez de colar num chat persistente (um DM de Discord, por exemplo, fica no histórico do servidor deles para sempre). Trade-off aceito: gerar/revogar um token exige acesso direto ao `lia-admin` — o mesmo já era verdade para gestão de usuários, só se estende à autenticação.

Token de longa duração (1 ano), sem refresh tokens — decisão final, com a mitigação abaixo.

```json
{
  "user_id": "uuid",
  "username": "gui",
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

### Onde o token mora no client

Sem senha pra digitar de novo caso o token vaze de um arquivo qualquer, o armazenamento do JWT no client vira o ponto sensível — precisa ser tratado como segredo, nunca como config comum:

- **Desktop (Rust)** — armazenamento nativo do SO, não arquivo de config em texto plano: Keychain no macOS, Credential Manager no Windows, Secret Service no Linux. A crate `keyring` (Rust) abstrai os três.
- **CLI** — mesma abordagem (`keyring`) quando possível; se não for viável no contexto de uso, um arquivo com permissão restrita (`chmod 600`, fora de qualquer pasta versionada) é o mínimo aceitável.
- **Bot de Discord** — roda como processo sem sessão de usuário/GUI, então armazenamento nativo do SO não se aplica da mesma forma. Trata igual a qualquer outro segredo de servidor já documentado aqui: variável de ambiente, nunca em arquivo versionado.

Isso é majoritariamente decisão do lado client (Rust) — fica registrado aqui como requisito de segurança que o client precisa cumprir, não como algo a implementar neste repositório.

## Canal como sinal de risco (adiado)

Nem todo canal de origem merece o mesmo teto de confiança — um comando vindo do app desktop (WebSocket autenticado, dispositivo que o usuário fisicamente possui) é mais confiável que o mesmo comando chegando via Discord (mensagem em canal de grupo, mais fácil de falsificar se a conta de alguém for comprometida). A ideia é a matriz de capability × trust level também considerar o canal de origem da conexão — por exemplo, tools com efeito físico exigindo um canal específico, independente do trust level da sessão.

Isso fica deliberadamente **fora de escopo por enquanto** (não depende de um cadastro de dispositivos — o token já é por usuário, ver acima) — só vale desenhar de verdade quando o caso de uso 4 do roadmap (ações físicas: trancar porta, backup) estiver sendo implementado de fato, com um canal real de menor confiança (Discord) para testar contra.
