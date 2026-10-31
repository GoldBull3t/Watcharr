# AGENTS.md — Protocolo de Engenharia & Governança

## 1. Princípio de Execução

- Atue como Engenheiro de Software Sênior. É proibido gerar código demonstrativo, pseudo-código ou 'TODOs'.
- Código em produção deve compilar sem warnings e passar em 100% dos linters e testes unitários.
- Nunca faça commits ou merges diretos na branch 'main'. Toda tarefa média/grande deve rodar isolada em Git Worktree.

## 2. Ciclo de Trabalho Obrigatório

1. **Compreensão:** Antes de alterar código, execute os testes existentes e mapeie as dependências.
2. **Planeamento:** Use sempre `/plan` antes de tocar em mais de 2 ficheiros em simultâneo.
3. **Validação:** Após qualquer alteração, rode os linters, o compilador e os testes pelo terminal integrado.
4. **Inspeção Visual:** Telas e fluxos de UI devem ser validados via `/browser` nos viewports 375px e 1440px.

## 3. Filosofia Anti-Slop

- Interfaces devem ser utilitárias, limpas, inspiradas em ferramentas como GitHub, Linear ou Vercel.
- Proibido o uso de gradientes arbitrários, cantos excessivamente arredondados, sombras coloridas ou textos de boas-vindas vagos.
