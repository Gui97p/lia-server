# Observabilidade

Logs em toda parte. Em um sistema event-driven com múltiplos componentes, saber o que aconteceu e quando é crítico.

Todo evento tem: timestamp, tipo, `user_id`, `task_id` (quando aplicável), payload, resultado.

Logging estruturado via `log/slog` (stdlib) — sem dependência externa para isso.

## Redação de segredos em log

"Logar tudo" é fácil de errar: logar o payload inteiro de uma request de login, o JWT, a `groq_api_key`, ou o áudio transcrito com dados sensíveis do usuário. Em vez de deixar isso a critério de cada ponto de log individual, existe um helper de logging que redige automaticamente campos sensíveis (`password`, `token`, `*_key`, etc.) antes de escrever — mais barato de fazer uma vez agora do que auditar todo `logger.Info(...)` do projeto depois.

Fora esse conjunto de campos sempre redigidos, o conteúdo do que é logado fica a critério de quem escreve cada ponto de log.

## Traces

Uma execução completa (Task → Agent → Planner → Model → Tool → resultado) deve poder ser reconstruída a partir dos logs — isso não exige um sistema de Trace dedicado agora, só que os campos estruturados (`task_id`, `user_id`, timestamp, etc.) continuem consistentes em todo evento. Um armazenamento/visualização de Trace dedicado (para debugging, avaliação, melhoria de prompts) é uma evolução futura, não uma necessidade do MVP.
