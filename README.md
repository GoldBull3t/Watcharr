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
<img width="1913" height="913" alt="watcharr2" src="https://github.com/user-attachments/assets/12f87ebf-deed-40bb-8f33-8d2138d7233b" />
<img width="1912" height="916" alt="watcharr3" src="https://github.com/user-attachments/assets/62ee1519-a2f4-44f3-b645-81f4c7c70057" />
<img width="1906" height="914" alt="watcharr4" src="https://github.com/user-attachments/assets/15539fe3-d943-4e73-b8d9-e9229b77133d" />

---

## ✨ Principais Melhorias Implementadas

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
