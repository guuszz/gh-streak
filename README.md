<h1 align="center">🔥 gh-streak</h1>

<p align="center">
  <b>GitHub contribution streak no terminal.</b><br/>
  <sub>Ano inteiro de commits em ASCII colorido, sem sair do shell.</sub>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.22-00ADD8?style=flat-square&logo=go&logoColor=white"/>
  <a href="https://github.com/guuszz/gh-streak/actions/workflows/ci.yml"><img src="https://github.com/guuszz/gh-streak/actions/workflows/ci.yml/badge.svg" alt="CI"/></a>
  <a href="LICENSE"><img src="https://img.shields.io/github/license/guuszz/gh-streak?style=flat-square" alt="MIT License"/></a>
  <a href="https://github.com/guuszz/gh-streak/releases"><img src="https://img.shields.io/github/v/release/guuszz/gh-streak?style=flat-square&logo=github" alt="Latest release"/></a>
</p>

---

## ✨ Preview

```
@guuszz · Gustavo Oliveira
May 27, 2025 — May 27, 2026 · 1,247 contributions

         Jan       Fev   Mar       Abr   Mai       Jun
     ██  ██  ██████  ██  ██  ██████████████████  ██  ██
Seg  ████████  ████  ██  ████████████████████  ██████
     ██████  ██████████████  ██████████████████  ██
Qua  ████████  ██████  ██████████████████████████████
     ██  ████████████████████  ██████████  ████████
Sex  ████████████  ██  ██████████████████████████████
     ██  ██  ████████  ██████████████████  ██████

Less ░░▒▒▓▓██ More

Current streak: 12 days 🔥
Longest streak: 42 days (Mar 14 → Apr 24, 2026)
Active days:    287 / 365 (79%)
```

> _Cores reais no teu terminal — paleta GitHub (5 níveis de verde)._

---

## 💡 Por que existe

GitHub mostra o teu heatmap **no perfil web**. Mas quando você tá no terminal codando, ver progresso requer abrir browser, esperar carregar, scrollar. **`gh-streak`** traz isso pra dentro do shell — útil pra:

- ✅ Adicionar no teu **prompt do shell** (`PROMPT_COMMAND`) — vê streak antes de cada comando
- ✅ Pôr no **MOTD** de uma VPS — motiva self-hosters
- ✅ Compartilhar em screenshots de termo (`asciinema`, `freeze`)
- ✅ Usar em scripts (`gh-streak --json | jq .streak.current`)

## 📦 Instalação

### Via Go (recomendado)

```bash
go install github.com/guuszz/gh-streak@latest
```

Garanta que `$GOPATH/bin` está no `$PATH`.

### Via binary release

Baixe do [latest release](https://github.com/guuszz/gh-streak/releases/latest) o arquivo pra teu OS/arch:

```bash
# macOS arm64
curl -L https://github.com/guuszz/gh-streak/releases/latest/download/gh-streak_Darwin_arm64.tar.gz | tar xz
sudo mv gh-streak /usr/local/bin/

# Linux x86_64
curl -L https://github.com/guuszz/gh-streak/releases/latest/download/gh-streak_Linux_x86_64.tar.gz | tar xz
sudo mv gh-streak /usr/local/bin/
```

### Via Homebrew (planejado)

```bash
brew install guuszz/tap/gh-streak  # em breve
```

## 🚀 Uso

```bash
# Tua própria conta (auto-detecta via gh CLI ou git config)
gh-streak

# Outro usuário
gh-streak torvalds

# Ano específico
gh-streak --year 2024 guuszz

# Sem cores (pra piping)
gh-streak --no-color | tee streak.txt

# JSON pra script
gh-streak --json | jq '.streak.current'
```

## 🔑 Token

`gh-streak` precisa de um token do GitHub pra ler o contribution calendar (a REST API não expõe — só GraphQL). Tenta as fontes nesta ordem:

1. **`GH_TOKEN`** env var
2. **`GITHUB_TOKEN`** env var
3. **`gh auth token`** (se GitHub CLI tiver autenticado)

Scope mínimo: `read:user`. [Crie um aqui](https://github.com/settings/tokens/new?scopes=read:user&description=gh-streak).

## 🛠 Como funciona

### A query GraphQL

O REST API do GitHub **não** expõe o contribution graph. Só o GraphQL via `contributionsCollection.contributionCalendar`:

```graphql
query($login: String!, $from: DateTime!, $to: DateTime!) {
  user(login: $login) {
    contributionsCollection(from: $from, to: $to) {
      contributionCalendar {
        totalContributions
        weeks {
          contributionDays {
            contributionCount
            date
            contributionLevel
          }
        }
      }
    }
  }
}
```

O `contributionLevel` é um enum (`NONE`, `FIRST_QUARTILE`, ..., `FOURTH_QUARTILE`) que o GitHub já calculou — mapeamos pra 0..4 e usamos como índice na paleta de cores.

### Cálculo do streak

**Current streak** conta de hoje pra trás, parando no primeiro dia com 0 contributions. Tolerância: se hoje ainda tá vazio, não quebra (mesma regra do perfil do GitHub).

**Longest streak** é um single pass O(n) pelo array de days, mantendo o run atual e o max.

### Render

Usa [`lipgloss`](https://github.com/charmbracelet/lipgloss) pra cores true-color (com fallback ANSI 256 automático). Cada dia vira 2 chars (`██`) pra dar aparência de "quadrado" em fontes monoespaçadas.

Layout: 7 linhas (dias da semana, Dom→Sáb) × ~53 colunas (semanas). Labels Seg/Qua/Sex à esquerda, mês no rodapé.

## 📁 Estrutura

```
gh-streak/
├── main.go      # Flag parsing, CLI orchestration, JSON output
├── github.go    # GraphQL client + token discovery
├── streak.go    # Streak calc (current + longest)
├── render.go    # Heatmap + stats + colors (lipgloss)
├── go.mod
├── .github/
│   ├── workflows/ci.yml         # build + vet em cada push
│   └── workflows/release.yml    # GoReleaser em cada tag
└── .goreleaser.yaml             # binaries Linux/macOS/Windows
```

## 🤔 Por que Go (e não outra linguagem)

- **Single binary** — `go build` gera 1 arquivo executável sem runtime
- **Cross-compile trivial** — `GOOS=darwin GOARCH=arm64 go build` from any machine
- **stdlib generoso** — `net/http`, `encoding/json`, `time`, `flag` cobrem 90% do escopo
- **Startup instantâneo** — CLIs Node leva 100ms+ pra inicializar V8. Go = 10ms.
- **Match com `gh` CLI ecosystem** — o próprio GitHub CLI é Go

## 🗺 Roadmap

- [x] Current + longest streak
- [x] ASCII heatmap colorido (5 níveis)
- [x] JSON output pra scripts
- [x] Token discovery (env + gh CLI)
- [ ] `gh-streak compare <user1> <user2>` — lado a lado
- [ ] `gh-streak --shell-prompt` — modo "1 linha" pra PS1
- [ ] Cache local (1h TTL) — evita hammer na API
- [ ] Homebrew tap
- [ ] gh CLI extension (`gh extension install guuszz/gh-streak`)

## 🤝 Contribuindo

PRs welcome! Lê o [CONTRIBUTING.md](CONTRIBUTING.md) e abre issue antes de feature grande.

## 📝 Licença

MIT © [Gustavo Oliveira](https://github.com/guuszz)
