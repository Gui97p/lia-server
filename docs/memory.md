# Memória

## Escopos

Memória não é apenas "fatos sobre o usuário". Existem três escopos:

```
GLOBAL   → conhecimento da Lia sobre o mundo (não pertence a ninguém)
USER     → fatos sobre um usuário específico
PRIVATE  → fatos que a Lia possui mas nenhum usuário pode acessar diretamente
```

Um escopo `GROUP` (fatos compartilhados entre usuários) foi cogitado, mas descartado por enquanto — a Lia é de uso interno, com poucas pessoas, e não existe nenhuma feature de grupo implementada. Revisitar só quando isso virar uma necessidade real, não especulativa.

> **Memória não é autorização.** A Lia pode saber algo sem nenhum usuário poder consultar esse dado.

## Injeção no contexto

No contexto de cada conversa, a Lia recebe:
- Memórias do usuário atual (USER)
- Memórias globais relevantes (GLOBAL)

Memórias PRIVATE nunca são expostas diretamente.

### MVP: injetar tudo

Enquanto o volume de memórias for pequeno (fase de descoberta por casos de uso, antes de haver client externo testando), a estratégia é injetar todas as memórias USER do usuário atual (+ GLOBAL relevante), sem filtro nem categorização. Não vale a pena construir um categorizador manual — isso é exatamente o problema que `pgvector` resolve de forma melhor.

`pgvector` entra em cena quando o servidor passar a ser usado de fato por um client em teste (fase beta, trabalhada em branch separada da `main`) — nesse ponto o volume de memórias já justifica busca semântica em vez de injeção total. Até lá, a interface `Store` (`internal/memory/store.go`) já isola quem consome memória de como ela é buscada, então trocar para busca semântica depois fica contido em `internal/memory`, sem afetar Planner/Executor.

## Tools de memória

```
saveMemory(fact, category, scope)
updateMemory(id, fact)
deleteMemory(id)
```

IDs são expostos no system prompt para que a Lia saiba o que pode gerenciar.

## Proveniência e segurança contra prompt injection

Memórias GLOBAL são injetadas no contexto de qualquer conversa. Se a Lia salvar memória a partir de conteúdo de terceiros (uma mensagem recebida, um documento, o resultado de uma tool), esse conteúdo pode conter instruções disfarçadas de fato, que depois voltam ao prompt do Planner como se fossem verdade estabelecida.

A defesa não é pedir confirmação do usuário toda vez que uma tool sensível roda — isso mata a experiência de uso principal (comando de voz direto, sem precisar ir ao computador confirmar). Em vez disso, cada instrução que chega ao Planner carrega uma **proveniência**:

- `comando_direto` — o usuário pediu isso agora, na mensagem/fala atual.
- `memoria_injetada` — a justificativa vem de uma memória GLOBAL no contexto, não de um pedido direto do usuário nesta interação.
- `agendado` — a Task foi configurada com antecedência pelo próprio usuário (ver [Tasks e Eventos](tasks-and-events.md#autorização-de-tasks-sem-sessão-viva)) para rodar sem sessão viva.
- `evento` — a Task foi disparada por uma condição externa (sensor, estado do ambiente) que o usuário configurou, mas sem comando ao vivo no momento do disparo.

Regra de segurança: **tools com efeito físico/irreversível (ver [matriz de capability × trust level](identity-auth-and-secrets.md)) só executam sem confirmação quando a proveniência é `comando_direto` ou `agendado`** — nos dois casos, o usuário já deu autorização explícita (agora, ou com antecedência ao configurar). `memoria_injetada` e `evento` exigem confirmação explícita antes de executar algo sensível — são os casos onde a justificativa não veio de uma ação direta e deliberada do usuário neste momento.

Isso ainda depende de definir, na implementação do Planner, como a proveniência é rastreada e anexada a cada step do Workflow (ver [Planner e Executor](planner-and-executor.md)).

## Regras de comportamento não são memória

Regras que ditam como a Lia deve se comportar (tom, tamanho de resposta, princípios de conduta) **não são fatos editáveis via `saveMemory`/`updateMemory`** — se fossem, uma regra maliciosa injetada por uma memória GLOBAL teria ainda mais garantia de influenciar o Planner do que um fato comum, já que regras de comportamento tendem a ser sempre carregadas no contexto.

Essas regras vivem numa tabela separada, `behavior_rules` (ver [Banco de Dados](database.md)), editável **apenas** via `lia-admin`/rota administrativa protegida — nunca pelas tools de memória, nunca influenciada por conteúdo injetado. `memories` continua cobrindo fatos e preferências (USER/GLOBAL/PRIVATE); `behavior_rules` cobre só a configuração de comportamento da própria Lia, que só o usuário edita diretamente.
