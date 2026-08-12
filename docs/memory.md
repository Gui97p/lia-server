# Memória

## Escopos

Memória não é apenas "fatos sobre o usuário". Existem quatro escopos:

```
GLOBAL   → conhecimento da Lia sobre o mundo (não pertence a ninguém)
USER     → fatos sobre um usuário específico
GROUP    → fatos compartilhados entre membros de um grupo
PRIVATE  → fatos que a Lia possui mas nenhum usuário pode acessar diretamente
```

> **Memória não é autorização.** A Lia pode saber algo sem nenhum usuário poder consultar esse dado.

## Injeção no contexto

No contexto de cada conversa, a Lia recebe:
- Memórias do usuário atual (USER)
- Memórias dos grupos do usuário (GROUP)
- Memórias globais relevantes (GLOBAL)

Memórias PRIVATE nunca são expostas diretamente.

### MVP: injetar tudo

Enquanto o volume de memórias for pequeno (fase de descoberta por casos de uso, antes de haver client externo testando), a estratégia é injetar todas as memórias USER + GROUP do usuário atual, sem filtro nem categorização. Não vale a pena construir um categorizador manual — isso é exatamente o problema que `pgvector` resolve de forma melhor.

`pgvector` entra em cena quando o servidor passar a ser usado de fato por um client em teste (fase beta, trabalhada em branch separada da `main`) — nesse ponto o volume de memórias já justifica busca semântica em vez de injeção total. Até lá, a interface `Store` (`internal/memory/store.go`) já isola quem consome memória de como ela é buscada, então trocar para busca semântica depois fica contido em `internal/memory`, sem afetar Planner/Executor.

## Tools de memória

```
saveMemory(fact, category, scope, target_id?)
updateMemory(id, fact)
deleteMemory(id)
```

IDs são expostos no system prompt para que a Lia saiba o que pode gerenciar.

## Proveniência e segurança contra prompt injection

Memórias GROUP e GLOBAL são injetadas no contexto de qualquer conversa dos membros relevantes. Se a Lia salvar memória a partir de conteúdo de terceiros (uma mensagem recebida, um documento, o resultado de uma tool), esse conteúdo pode conter instruções disfarçadas de fato, que depois voltam ao prompt do Planner como se fossem verdade estabelecida.

A defesa não é pedir confirmação do usuário toda vez que uma tool sensível roda — isso mata a experiência de uso principal (comando de voz direto, sem precisar ir ao computador confirmar). Em vez disso, cada instrução que chega ao Planner carrega uma **proveniência**:

- `comando_direto` — o usuário pediu isso agora, na mensagem/fala atual.
- `memoria_injetada` — a justificativa vem de uma memória GROUP/GLOBAL no contexto, não de um pedido direto do usuário nesta interação.

Regra de segurança: **tools com efeito físico/irreversível (ver [matriz de capability × trust level](identity-auth-and-secrets.md)) só executam sem confirmação quando a proveniência é `comando_direto`.** Se o Planner decidir executar algo sensível cuja justificativa é `memoria_injetada`, aí sim é exigida confirmação explícita — que é justamente o caso raro onde prompt injection se manifestaria, não o caminho comum de uso.

Isso ainda depende de definir, na implementação do Planner, como a proveniência é rastreada e anexada a cada step do Workflow (ver [Planner e Executor](planner-and-executor.md)).
