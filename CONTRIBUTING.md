# Contribuindo

PRs welcome! Esse é meu primeiro projeto em Go público, então feedback é ouro.

## 🛠 Setup

```bash
git clone https://github.com/guuszz/gh-streak.git
cd gh-streak
go build .
./gh-streak --help
```

## 🐛 Reportando bugs

Abre uma [issue](../../issues) com:
- Versão (`gh-streak --version`)
- Sistema operacional + arquitetura
- Comando exato que rodou
- Output (com `--no-color` se relevante)

## 💡 Sugerindo features

Ideias bem-vindas, especialmente:
- Novos modos de output (Slack, RSS, image PNG?)
- Cache local (Redis? sqlite? plain JSON?)
- Performance (a query pode ser otimizada pra `--year`)
- Internacionalização (PT-BR é o atual; EN/ES/JP em discussão)

Abre issue com label `enhancement` antes de PR pra evitar trabalho desperdiçado.

## 🛠 Submetendo Pull Requests

### Workflow

1. Fork + branch a partir de `main`: `git checkout -b feat/minha-feature`
2. Mantém commits atômicos e descritivos
3. Roda `go build ./...` + `go vet ./...` localmente — CI vai rodar o mesmo
4. Push pra tua fork
5. Abre PR contra `main`

### Padrões de commit

[Conventional Commits](https://www.conventionalcommits.org/):
- `feat:` nova feature
- `fix:` bug fix
- `chore:` mudança que não afeta funcionalidade
- `refactor:` refactor sem mudança de comportamento
- `docs:` mudança de docs
- `perf:` melhoria de performance

### Code style

- `gofmt -s` formatação default — não negociável
- Comentários em PT-BR são OK, código (identificadores) em EN
- Pacote único pra manter simples; não criar `internal/` se não necessário
- Testes em `*_test.go` quando aplicável

## 📝 Licença

Ao contribuir, você concorda que suas contribuições serão licenciadas sob MIT.

---

Made with ❤️ in Vitória da Conquista, BA · [Gustavo Oliveira](https://github.com/guuszz)
