# Watcharr (Custom Edition)

Uma edição personalizada, ultra-leve e de alto desempenho do **Watcharr**, desenhada para registo pessoal e acompanhamento de filmes, séries e videojogos.

---

## ✨ O que foi Melhorado & Implantado

### 🗺️ Franchise Roadmaps (Linhas Cronológicas & Universos)
* **Visualização Completa de Franquias:** Acesso direto a roadmaps de sagas como *Marvel Cinematic Universe*, *Star Wars*, *Harry Potter / Wizarding World*, *Alien & Predator*, *Tolkien*, *MonsterVerse*, entre outros.
* **Comutador de Ordem:** Alternância imediata entre **Ordem de Lançamento** (datas de estreia oficiais) e **Ordem Cronológica** (linha temporal da história).
* **Progresso da Saga:** Indicador visual de conclusão, destaque automático para o próximo título a ver (*Next Up*) e badges de ligação direta nos detalhes de cada produção.

### 📡 Widgets & Integrações Domésticas (Homelab)
* **Endpoints Otimizados:** Respostas prontas em formato JSON para dashboards populares como **Homepage**, **Dashy** e sensores REST do **Home Assistant**.
* **Cache em Memória RAM:** Endpoints de polling frequente mantidos em cache curta na RAM para prevenir leituras redundantes na base de dados SQLite.

### 🔔 Notificações e Webhooks de Calendário
* **Alertas de Estreia:** Envio automatizado de resumos diários para Discord ou webhooks compatíveis com os episódios que estreiam no próprio dia.
* **Teste com 1 Clique:** Botão de teste imediato integrado no painel de configurações para validação do webhook.

### 🎨 Experiência de Utilização & Fluidez
* **Header Fixo Inteligente:** Navegação e barra de ferramentas da biblioteca com fixação suave no topo (*sticky* com *backdrop-blur*), sem qualquer sobreposição indesejada sobre os cartões durante o scroll.
* **Hero Spotlight & Continue Watching:** Destaque dinâmico e calha de conteúdos em progresso posicionados de forma intuitiva no topo da biblioteca.

### ⚡ Motor Backend & Performance Extrema
* **Consumo Mínimo de Recursos:** Servidor Go nativo com consumo de memória reduzido para apenas **~23 MB de RAM** em repouso.
* **SQLite em Modo WAL:** Pragmas calibrados para 16 MB de cache em RAM e escrita não bloqueante para leituras concorrentes instantâneas.
* **Índices de Base de Dados Dedicados:** Adicionados índices compostos (`idx_contents_title`, `idx_games_name`, `idx_watcheds_user_status`) para pesquisas em milissegundos.
* **Blindagem e Robustez:** Correção de potenciais *nil-pointer panics*, proteção contra fugas de descritores de rede (*sockets*) e suporte completo para reposição de avaliações a zero.

---

## 🚀 Como Executar Localmente

### 1. Backend (Go)
Certifica-te de que tens o [Go](https://go.dev/) instalado:

```bash
cd server
go run .
```

> **Dica:** Podes também executar diretamente o binário compilado `Watcharr.exe` localizado na pasta `server/`.  
> O backend estará disponível em: **`http://localhost:3080`** (Health check em `/api/health`).

### 2. Frontend (SvelteKit)
Certifica-te de que tens o [Node.js](https://nodejs.org/) instalado:

```bash
# Na raiz do projeto:
npm install
npm run dev
```

> O frontend estará disponível em: **`http://localhost:5173`**.

---

## 🔌 Portas e Serviços

| Serviço | Porta Local | Descrição |
| :--- | :--- | :--- |
| **Backend API** | `3080` | Servidor Go, base de dados SQLite e integrações |
| **Frontend UI** | `5173` | Aplicação web interativa SvelteKit com Vite |

---

## 🛠️ Tecnologias Principais
* **Frontend:** SvelteKit, TypeScript, TailwindCSS / CSS Scoped.
* **Backend:** Go (Gin Framework, GORM).
* **Base de Dados:** SQLite em modo WAL (armazenamento local rápido e sem dependências externas).
