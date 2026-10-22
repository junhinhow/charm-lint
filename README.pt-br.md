# charm-lint

Análise estática para as armadilhas de terminal que quebram TUIs no estilo Charm: padding contado em bytes, UTF-8 cortado ao meio, offset de byte usado como índice de rune, iteração de map vazando para a saída renderizada e falhas de comando que somem sem deixar rastro.

## Instalação

```bash
go install github.com/junhinhow/charm-lint/cmd/charm-lint@latest
```

Como biblioteca:

```bash
go get github.com/junhinhow/charm-lint
```

## Uso

Rode todas as regras no módulo atual:

```bash
charm-lint ./...
```

O processo sai com status 3 quando encontra problemas (a convenção usual dos drivers `go/analysis`), então dá para usar como gate de CI.

### Como vettool

```bash
go vet -vettool=$(which charm-lint) ./...
```

### A partir de Go

```go
import (
	"github.com/junhinhow/charm-lint/lint"
	"golang.org/x/tools/go/analysis/unitchecker"
)

func main() {
	unitchecker.Main(lint.Analyzers()...)
}
```

## Regras

| Regra | Detecta |
| --- | --- |
| `bytelength-padding` | padding com `strings.Repeat` medido por `len` ou `utf8.RuneCountInString` |
| `byteslice` | `s[:n]` / `s[n:]` com limite variável que chega à saída renderizada |
| `range-index-arith` | aritmética com o índice de `for i := range s` quando `s` é string |
| `nondeterministic-render` | `range` sobre map dentro de `Render*`, `View*` ou `String` |
| `silent-exec` | `exec.Command(...).Run/Output/CombinedOutput` com erro descartado ou sem `cmd.Stderr` |
| `discarded-stat-error` | `fi, _ := os.Stdin.Stat()` seguido de `fi.Mode()` / `fi.Size()` |

### bytelength-padding

`len` conta bytes: `ação` tem 6 bytes e 4 células. `utf8.RuneCountInString` conta runes, mas CJK e emoji ocupam duas células.

```go
// errado: a borda fica mais larga que o rótulo
borda := strings.Repeat("─", len(rotulo))

// certo: medir a largura de exibição
borda := strings.Repeat("─", max(0, 20-ansi.StringWidth(rotulo)))
```

### byteslice

Cortar uma string por byte pode partir um rune e produzir UTF-8 inválido, que o terminal mostra como `�`.

```go
// errado
fmt.Println(s[:largura])

// certo
fmt.Println(ansi.Truncate(s, largura, "…"))
```

A regra segue atribuições e retornos simples (dentro da mesma função) e sinaliza slices cujos limites não são constantes de compilação.

### range-index-arith

Em `for i := range s` com string, `i` é offset de byte.

```go
// errado: i+1 pode apontar para o meio de um rune multibyte
for i := range s {
	proximo := s[i+1]
	_ = proximo
}

// certo
for _, r := range s {
	_ = r
}
```

### nondeterministic-render

Go randomiza a ordem de iteração de map. Renderizar o map direto deixa snapshots e testes instáveis.

```go
// errado
for k, v := range m {
	fmt.Fprintf(&b, "%s=%d\n", k, v)
}

// certo
for _, k := range slices.Sorted(maps.Keys(m)) {
	fmt.Fprintf(&b, "%s=%d\n", k, m[k])
}
```

### silent-exec

Um comando que falha sem que o erro ou o stderr apareçam deixa a interface parecendo funcionar.

```go
// errado: qualquer falha desaparece
_ = exec.Command("git", "status").Run()

// certo
cmd := exec.Command("git", "status")
cmd.Stderr = os.Stderr
if err := cmd.Run(); err != nil {
	return fmt.Errorf("git status: %w", err)
}
```

### discarded-stat-error

`Stat` devolve um `FileInfo` nil quando falha, então usar o valor depois provoca panic.

```go
// errado: nil pointer dereference quando Stat falha
info, _ := os.Stdin.Stat()
largura := info.Size()

// certo
info, err := os.Stdin.Stat()
if err != nil {
	return err
}
largura := info.Size()
```

## API

| Símbolo | Descrição |
| --- | --- |
| `Analyzers() []*analysis.Analyzer` | As seis regras, prontas para `unitchecker` |
| `Charmlint` | Analisador agregado usado pelo binário; roda todas as regras no mesmo pass |
| `ByteLengthPadding` | padding com `strings.Repeat` medido em bytes ou runes |
| `ByteSlice` | slice de string por byte que chega à saída renderizada |
| `RangeIndexArith` | offset de byte de um range sobre string usado em aritmética |
| `NondeterministicRender` | iteração de map dentro de função de renderização |
| `SilentExec` | comando externo sem tratamento de erro ou sem `cmd.Stderr` |
| `DiscardedStatError` | erro de `os.Stdin/Stdout/Stderr.Stat` descartado antes de `Mode`/`Size` |

## Como usar no CI

GitHub Actions:

```yaml
- name: charm-lint
  run: |
    go install github.com/junhinhow/charm-lint/cmd/charm-lint@latest
    charm-lint ./...
```

Ou via `go vet`, que se encaixa em um passo de vet já existente:

```bash
go install github.com/junhinhow/charm-lint/cmd/charm-lint@latest
go vet -vettool=$(which charm-lint) ./...
```

## Notas de design

- É `go/analysis` puro; ainda não há `SuggestedFix`. Cada diagnóstico nomeia a API que corrige o problema, então a correção fica explícita na mensagem.
- `byteslice` é intraprocedural: segue parênteses, concatenação de strings, atribuições e retornos dentro do mesmo arquivo, com profundidade limitada. Valores que atravessam fronteiras de função não são rastreados.
- `nondeterministic-render` só sinaliza loops sobre map que realmente produzem saída (chamadas Print/Sprintf/Render/Write/Format, concatenação de strings ou `+=`). Um loop que apenas coleta chaves para ordenar depois não é sinalizado; um loop que só acrescenta valores a um slice sem formatar é um falso negativo conhecido.
- `silent-exec` verifica `cmd.Stderr` no escopo da função, não por comando. `CombinedOutput` também dispara a checagem de stderr, conforme a regra especifica.
- Arquivos de teste (`_test.go`) são analisados como qualquer outro.
- Nomes de analisador não podem ter hífen (`analysis.Validate` exige identificadores), então os IDs com hífen ficam em `Diagnostic.Category` e nesta tabela.
- A ferramenta precisa que os pacotes alvo compilem; código gerado e arquivos específicos de build tags seguem as regras usuais do `go/packages`.

## Licença

MIT
