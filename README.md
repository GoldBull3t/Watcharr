# Watcharr (Custom Edition) 🎬🍿

Uma edição personalizada, ultra-leve e de alto desempenho do **Watcharr**, desenhada para registo pessoal e acompanhamento de filmes, séries e videojogos.

---

> [!IMPORTANT]
> **Todas as rotas de API, modelos de dados e endpoints são rigorosamente os mesmos do [Watcharr](https://github.com/sbondco/Watcharr).**  
> Não existem *breaking changes* de protocolo ou migrações forçadas. Qualquer cliente, webhook, ferramenta de backup ou integração externa existente funciona diretamente sem necessidade de qualquer alteração.
>
> **O que mudou?**  
> Foram realizadas exclusivamente **melhorias estruturais no backend** (eficiência de concorrência, queries otimizadas, SQLite em modo WAL calibrado e prevenção de memory leaks) e **melhorias modernas no frontend** (SvelteKit, maior fluidez visual, melhor responsividade, controlos táteis aperfeiçoados e componentes com estilo refinado).

---

## 💬 À Procura de Feedback!

Este projeto é ativamente mantido e estou ativamente à procura de opiniões, críticas e sugestões da comunidade:
- Experimentaste a aplicação? Partilha o que achaste da velocidade e da experiência de utilização!
- Encontraste algum comportamento inesperado ou bug visual?
- Tens ideias de novas integrações ou melhorias?

👉 **Abre uma [Issue](https://github.com/GoldBull3t/Watcharr/issues) ou inicia uma discussão! Todo o feedback é muito bem-vindo.**

---

## 🖼️ Showcase & Galeria
<img width="1913" height="913" alt="watcharr2" src="https://github.com/user-attachments/assets/c7c07bae-be86-4054-a999-4968ea22a9cd" />
<img width="1906" height="914" alt="watcharr4" src="https://github.com/user-attachments/assets/b1927bdc-5285-4e03-aa63-fd7930a2c9d7" />
<img width="1912" height="916" alt="watcharr3" src="https://github.com/user-attachments/assets/d43ad80e-6024-42cd-ad5b-9d87507bbe54" />

---

## ✨ Principais Melhorias Implementadas

### 🌐 Listas Temáticas Personalizadas & Partilháveis
* **Partilha Desautenticada:** Cria listas personalizadas de filmes, séries e videojogos e partilha-as através de um link direto público (`/tag/:id`) sem exigir conta aos visitantes.
* **Privacidade Granular:** Comuta instantaneamente a visibilidade entre pública e privada com garantia de isolamento (`403 Forbidden` nas privadas).

### 👥 Watch Together (Cruzar Watchlists)
* **Recomendações Colaborativas:** Cruza watchlists com amigos para descobrir o que ver em conjunto (modos: *Em Comum*, *Não Vistos* ou *Recomendações Inteligentes*).

### 🎬 Trailers TMDB com 1 Clique & Hub de Sagas
* **Trailers Sem Publicidade:** Assiste a trailers e teasers oficiais diretamente num modal lightbox sem anúncios nem cookies invasivos (`youtube-nocookie.com`).
* **Hub de Coleções e Franquias:** Vista agregada de sagas (Marvel, Star Wars, Dune, etc.) com cálculo dinâmico de progresso visual por utilizador.

### 🏠 Integrações Homelab em Tempo Real (Event-Driven)
* **Outbound Webhooks:** Motor assíncrono para automações instantâneas no Home Assistant, Discord, Slack e Ntfy com proteção anti-SSRF.
* **Overseerr & Jellyseerr:** Cliente nativo integrado para consulta de disponibilidade e submissão de pedidos com 1 clique.
* **Suporte Completo ao Emby:** Ingestão nativa de webhooks Emby (`playback.stop`, `playback.start`, `item.markplayed`) e compatibilidade de tokens.

### 📱 PWA & Notificações Web Push
* **Instalação Nativa:** Transforma o Watcharr numa Progressive Web App (PWA) instalável em smartphone e desktop com funcionamento offline.
* **Alertas Push VAPID:** Subscrição direta e notificações push sem intermediários no browser para novos episódios e atividade.

### 🗄️ Multi-Base de Dados & Pesquisa Inteligente
* **Motores de Dados:** Suporte nativo para **PostgreSQL**, **MySQL** e **SQLite** com connection pooling automático (`DB_TYPE` / `DATABASE_URL`).
* **Pesquisa Fuzzy de Alta Performance:** Algoritmo híbrido com indexação por trigramas, similaridade de Dice e tolerância a erros tipográficos (Damerau-Levenshtein).

### 🛡️ Disaster Recovery & Backups Automáticos Agendados
* **Snapshots Sem Downtime:** Dumps atómicos e comprimidos em Gzip (`.db.gz` e `.json.gz`) com truncagem de checkpoint WAL e zero risco de corrupção.
* **Agendamento com Retenção:** Motor de tarefas configurável (12h, 24h, 7 dias) com limpeza automática de snapshots antigos (retenção de últimos N backups).
* **Painel de Gestão & Restauro com 1 Clique:** Listagem, download, upload externo e restauro com salvaguarda automática prévia contra perda de dados acidental.

### 🗺️ Franchise Roadmaps (Linhas Cronológicas)
* **Comutador de Ordem:** Alternância imediata entre **Ordem de Lançamento** (datas de estreia oficiais) e **Ordem Cronológica** (linha temporal da história).

### 🏠 Widgets & Integrações Domésticas (Homelab)
* **Endpoints Otimizados:** Respostas prontas em formato JSON para dashboards populares como **Homepage**, **Dashy** e sensores REST do **Home Assistant**.
* **Cache em Memória RAM:** Endpoints de polling frequente mantidos em cache curta na RAM para prevenir leituras redundantes na base de dados SQLite.

### 🔔 Notificações e Webhooks de Calendário
* **Alertas de Estreia:** Envio automatizado de resumos diários para Discord ou webhooks compatíveis com os episódios que estreiam no próprio dia.
* **Teste com 1 Clique:** Botão de teste imediato integrado no painel de configurações para validação do webhook.

### ⚡ Motor Backend & Performance Extrema
* **Consumo Mínimo de Recursos:** Servidor Go nativo com consumo de memória reduzido para apenas **~23 MB de RAM** em repouso.
* **SQLite em Modo WAL:** Pragmas calibrados para 16 MB de cache em RAM e escrita não bloqueante para leituras concorrentes instantâneas.
* **Índices de Base de Dados Dedicados:** Adicionados índices compostos (`idx_contents_title`, `idx_games_name`, `idx_watcheds_user_status`) para pesquisas em milissegundos.
* **Blindagem e Robustez:** Correção de potenciais *nil-pointer panics*, proteção contra fugas de descritores de rede (*sockets*) e suporte completo para reposição de avaliações a zero.

---

## 🐳 Executar via Docker (Recomendado)

A imagem oficial pré-compilada desta versão está disponível no GitHub Container Registry:

```bash
docker run -d \
  --name watcharr \
  -p 3077:3077 \
  -v ./data:/data \
  --restart unless-stopped \
  ghcr.io/goldbull3t/watcharr:latest
```

### Docker Compose
```yaml
services:
  watcharr:
    image: ghcr.io/goldbull3t/watcharr:latest
    container_name: watcharr
    restart: unless-stopped
    ports:
      - "3077:3077"
    volumes:
      - ./data:/data
```

---

## 🛠️ Como Executar em Desenvolvimento Local

### 1. Backend (Go)
Certifica-te de que tens o [Go](https://go.dev/) instalado:

```bash
cd server
go run .
```
> O backend estará disponível em: **`http://localhost:3080`** (Health check em `/api/health`).

### 2. Frontend (SvelteKit)
Certifica-te de que tens o [Node.js](https://nodejs.org/) instalado:

```bash
npm install
npm run dev
```
> O frontend estará disponível em: **`http://localhost:5173`**.

---


## ⚖️ Licença
Distribuído sob a licença [Apache 2.0](LICENSE). Baseado no projeto original [Watcharr](https://github.com/sbondco/Watcharr) de Sam Bond.
