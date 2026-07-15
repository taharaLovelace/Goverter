# Plano de compatibilidade Linux e eficiência de saída

## Objetivo

Tornar o Goverter oficialmente compatível com Linux x86-64 e avaliar reduções
reais no tamanho dos arquivos convertidos, preservando o contrato atual da CLI,
a qualidade associada aos presets e a publicação segura dos arquivos.

Este plano não inclui funcionalidades especulativas. Cada ajuste de compressão
deve ser medido antes de entrar no produto; se não houver ganho consistente, o
código atual permanece.

## Estado atual verificado

- O binário já compila para `linux/amd64` com `CGO_ENABLED=0`. Uma compilação
  cruzada local produziu o executável sem dependências C.
- `internal/publish/replace_other.go` já usa `os.Rename`, e o arquivo temporário
  é criado no mesmo diretório da saída. Esse fluxo atende Linux e não precisa
  ser reescrito.
- `exec.CommandContext`, `filepath` e o tratamento de `os.Interrupt` usados pelo
  núcleo são portáveis.
- O uso de FFmpeg pelo `PATH` já funciona no Linux.
- `GOVERTER_FFMPEG_DIR` e o diretório `tools` ainda procuram apenas
  `ffmpeg.exe` e `ffprobe.exe`; esses dois caminhos não funcionam no Linux.
- `samePath` usa comparação sem diferenciar maiúsculas e minúsculas, embora
  caminhos Linux normalmente sejam sensíveis a caixa.
- A CI executa somente em Windows e a release gera somente o instalador
  Windows x64.
- O pacote de FFmpeg registrado em `tools.lock.json` é exclusivamente Windows.
- O preset `compact` já é o mecanismo público para gerar arquivos menores. Os
  ajustes atuais de CRF, qualidade e bitrate estão em
  `internal/convert/planner.go`.
- O README ainda informa Go 1.26.4, enquanto `go.mod` exige Go 1.26.5.

## Decisões de escopo

1. Suportar primeiro **Linux x86-64 (`linux/amd64`)**. ARM64 só deve ser
   adicionado quando houver ambiente de teste e demanda concreta.
2. Distribuir no Linux apenas o binário do Goverter e sua documentação. O
   usuário instala `ffmpeg` e `ffprobe` pelo sistema. Não escolher nem
   redistribuir agora um build Linux de terceiros.
3. Não adicionar codecs, aceleração por hardware, processamento paralelo,
   configuração arbitrária do FFmpeg ou novos presets nesta etapa.
4. Não alterar silenciosamente resolução, taxa de quadros, canais de áudio ou
   metadados para reduzir tamanho; isso mudaria o conteúdo, não apenas a
   eficiência da codificação.
5. Não usar UPX. O build de release já utiliza `-s -w` e `-trimpath`; outra
   camada de compactação do executável não resolve o tamanho da mídia gerada.

## Fase 1 — corrigir a portabilidade do núcleo

### 1.1 Resolver nomes nativos de FFmpeg

Arquivos: `internal/toolchain/resolver.go` e
`internal/toolchain/resolver_test.go`.

- Fazer `FFmpeg()` e `FFprobe()` trabalharem com o nome-base da ferramenta.
- Acrescentar `.exe` somente em Windows ao montar candidatos em
  `GOVERTER_FFMPEG_DIR` e no diretório `tools` ao lado do executável.
- Continuar usando `exec.LookPath("ffmpeg")` e `exec.LookPath("ffprobe")` como
  último fallback.
- Manter a ordem atual: variável de ambiente, `tools`, `PATH`.
- Adaptar o teste para criar o nome nativo do sistema em vez de codificar
  `.exe` no fixture.

Critérios de aceite:

- No Linux, um diretório configurado com arquivos `ffmpeg` e `ffprobe` é
  resolvido corretamente.
- No Windows, os testes existentes continuam encontrando os arquivos `.exe`.
- Uma variável configurada para um diretório inválido continua falhando sem
  cair silenciosamente para outro FFmpeg.

### 1.2 Respeitar caminhos sensíveis a caixa

Arquivos: `internal/convert/service.go` e um teste em
`internal/convert/service_edge_test.go`.

- Usar `strings.EqualFold` somente no Windows.
- Nos demais sistemas, comparar os caminhos limpos com igualdade normal.
- Não adicionar resolução de links simbólicos até existir um caso real que a
  exija.

Critério de aceite: no Linux, diretórios `converted` e `Converted` não são
tratados como o mesmo caminho; no Windows, o comportamento atual é preservado.

### 1.3 Verificar permissões da saída

O arquivo temporário de conversão é criado por `os.CreateTemp`, portanto deve
ser verificado em um teste de integração Linux se o arquivo publicado termina
com permissão adequada para leitura normal. Só aplicar `chmod` ou mudar a forma
de criação do temporário se o teste reproduzir uma permissão inadequada. Não
alterar esse fluxo por suposição.

## Fase 2 — comprovar suporte em CI

Arquivo: `.github/workflows/ci.yml`.

- Executar o job atual em uma matriz mínima com `windows-latest` e
  `ubuntu-latest`.
- Manter `gofmt`, `go vet`, `go mod verify`, `go test ./...` e
  `go build ./cmd/goverter` em ambos.
- No job Linux, instalar FFmpeg pelo gerenciador do runner e executar
  `go test -tags integration ./integration`.
- Não criar testes duplicados por sistema; adicionar apenas casos específicos
  para nome de executável, sensibilidade a caixa e permissões.

Critérios de aceite:

- CI verde nos dois sistemas.
- Conversões reais de WebM, MP3, FLAC e WebP passam com o FFmpeg do Linux.
- Criação e merge de PDF passam nativamente no Linux.
- `goverter --version`, `goverter formats` e `goverter completion bash`
  executam no runner Linux.

## Fase 3 — publicar um artefato Linux

Arquivo principal: `.github/workflows/release.yml`.

Artefato proposto:

```text
Goverter-<versao>-linux-x64.tar.gz
```

Conteúdo mínimo:

```text
goverter
README.md
LICENSE
THIRD_PARTY_NOTICES.md
CHANGELOG.md
```

Implementação:

- Criar o binário em `ubuntu-latest` com `CGO_ENABLED=0`, `GOOS=linux` e
  `GOARCH=amd64`, usando os mesmos `-trimpath` e `ldflags` da versão Windows.
- Preservar o bit executável dentro do `tar.gz`.
- Executar o binário e os testes de integração antes de empacotar.
- Gerar SHA-256 para o arquivo Linux.
- Fazer os jobs Windows e Linux enviarem seus resultados como artifacts do
  workflow; um job final baixa ambos e cria uma única GitHub Release.
- Manter o instalador Windows e seu teste de instalação sem mudanças de
  comportamento.

O compartilhamento por artifacts é necessário porque os pacotes são gerados
em sistemas diferentes. O GitHub documenta `upload-artifact` e
`download-artifact` para essa transferência entre jobs.

Critérios de aceite:

- A release contém instalador e checksum Windows, além de arquivo e checksum
  Linux.
- A extração do `tar.gz` mantém `goverter` executável.
- O binário informa a versão, commit e data injetados pela release.
- Nenhum FFmpeg é incluído no pacote Linux.

## Fase 4 — documentação de instalação

Arquivos: `README.md` e `THIRD_PARTY_NOTICES.md`.

- Separar instalação Windows e Linux.
- Informar que Linux requer `ffmpeg` e `ffprobe` disponíveis no `PATH` ou em
  `GOVERTER_FFMPEG_DIR`.
- Mostrar download, extração e instalação opcional do binário em um diretório
  do `PATH`.
- Documentar `goverter completion bash` e `goverter completion zsh` sem criar
  scripts próprios de completion; Cobra já gera ambos.
- Corrigir o requisito de desenvolvimento para Go 1.26.5.
- Explicar que os codecs disponíveis dependem da compilação do FFmpeg instalada
  no sistema.

## Fase 5 — medir e reduzir o tamanho das saídas

### 5.1 Criar uma linha de base antes de alterar argumentos

Usar o mesmo conjunto local de entradas para medir:

- tamanho em bytes;
- duração da codificação;
- formato e streams confirmados por `ffprobe`;
- cada combinação atual de formato e preset.

O conjunto deve conter ao menos um vídeo com movimento e áudio, um áudio e uma
imagem fotográfica. Esses arquivos não devem entrar no Git se não houver
licença e tamanho adequados. Os fixtures sintéticos atuais são suficientes
para corretude, mas pequenos demais para decidir eficiência de compressão.

Registrar os resultados em uma tabela no pull request. Não criar um framework
de benchmark permanente para uma medição isolada.

### 5.2 Avaliar somente ajustes suportados pelos encoders atuais

Cada candidato deve ser testado isoladamente:

1. **PNG:** comparar o `-compression_level 6` atual com o padrão/máximo do
   encoder. A documentação do FFmpeg informa faixa de 0 a 9 e padrão 9. Se o
   resultado confirmar arquivo menor sem mudança de pixels, remover o override
   `6` é preferível a adicionar configuração.
2. **FLAC:** comparar o nível atual 5 com níveis superiores suportados. FLAC
   continua lossless; adotar outro nível apenas se a redução justificar o
   aumento de tempo.
3. **WebP:** medir `-compression_level 6` contra o padrão 4 mantendo os valores
   atuais de `-quality`. O FFmpeg define esse parâmetro como troca entre esforço
   e eficiência; não presumir ganho sem medir.
4. **MP4/H.264:** comparar `-preset medium` com `slow` apenas para o preset
   `compact`, mantendo CRF e bitrate. Não mudar o padrão `balanced` sem ganho
   demonstrado e tempo aceitável.
5. **WebM/VP9:** comparar `-cpu-used 2` com um valor menor apenas para
   `compact`. Valores menores usam mais tempo; a documentação não garante que
   a troca compense para todas as entradas.

Regra de decisão: adotar uma mudança somente se produzir arquivos menores de
forma consistente no conjunto avaliado, preservar a validade do arquivo e não
causar custo de tempo considerado desproporcional. Se o ganho for instável ou
irrelevante, manter os argumentos atuais.

### 5.3 Testes após qualquer ajuste aceito

- Atualizar `internal/convert/planner_test.go` para verificar somente os
  argumentos decididos.
- Executar os testes de integração nos dois sistemas.
- Comparar novamente todos os formatos para evitar regressão fora do candidato.
- Manter `compact`, `balanced` e `quality` com o mesmo significado público.

## Melhorias explicitamente adiadas

- Conversões paralelas de diretório: FFmpeg já usa múltiplas threads e não há
  medição mostrando que concorrência adicional melhora o tempo total.
- HEVC/H.265 e AV1: seriam novos formatos com impacto de compatibilidade, não
  uma otimização interna.
- Duas passagens de vídeo: aumenta muito o tempo e exige outro fluxo de
  temporários; não é necessário para o modelo CRF atual.
- Recompressão automática de imagens em PDF: pode perder qualidade e alterar
  o documento. `pdf merge` deve continuar preservando os PDFs recebidos.
- Redução automática de resolução, FPS ou canais: requer opção explícita e uma
  decisão de produto separada.
- Pacotes `.deb`, `.rpm`, Snap, Flatpak, Homebrew e repositórios de distro: o
  `tar.gz` cobre a primeira distribuição Linux sem multiplicar manutenção.
- Linux ARM64 e macOS: somente após demanda e CI nativa.

## Ordem de implementação sugerida

1. Resolver nomes nativos e corrigir comparação de caminhos.
2. Adicionar Linux à CI e validar permissões reais.
3. Atualizar README e avisos.
4. Gerar e publicar o `tar.gz` Linux pela release.
5. Medir os argumentos atuais de compressão.
6. Aplicar apenas candidatos que passarem pela regra de decisão.

Cada item deve ser um commit pequeno e revisável. A compatibilidade Linux não
deve depender das otimizações de compressão; assim, a entrega pode ocorrer
mesmo que nenhum candidato reduza os arquivos de forma convincente.

## Definição de pronto

- CI Windows e Linux verde.
- Binários Windows e Linux publicados com checksums.
- Fluxos `convert`, `info`, `formats`, `pdf images`, `pdf merge` e completion
  validados no Linux.
- Busca de FFmpeg funciona por variável, diretório adjacente e `PATH` nos dois
  sistemas.
- Documentação descreve dependências e instalação sem prometer codecs ausentes.
- Toda mudança de compressão possui comparação registrada; nenhuma mudança é
  incluída apenas por expectativa.

## Referências técnicas

- [FFmpeg: opções de codecs](https://ffmpeg.org/ffmpeg-codecs.html)
- [GitHub Actions: workflow artifacts](https://docs.github.com/en/actions/concepts/workflows-and-actions/workflow-artifacts)
- [Go: build constraints e nomes por GOOS/GOARCH](https://go.dev/src/cmd/go/internal/help/helpdoc.go)
