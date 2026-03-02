# =============================================================================
# STAGE 1: "builder" — Compila o código Go
# =============================================================================
# Usamos a imagem oficial do Go baseada em Alpine Linux (leve, ~300MB vs ~800MB da Debian).
# Este stage é temporário: só existe para compilar o binário. NÃO vai para a imagem final.
FROM golang:1.24-alpine AS builder

# Instala "ca-certificates" (necessário para chamadas HTTPS dentro do app)
# e "git" (necessário para o "go mod download" buscar dependências privadas, se houver).
RUN apk add --no-cache ca-certificates git

# Define o diretório de trabalho dentro do container. É como fazer "mkdir /app && cd /app".
WORKDIR /app

# Copia APENAS o go.mod primeiro (sem o go.sum).
# Técnica de cache de layers do Docker: se o go.mod não mudar entre builds,
# o Docker reutiliza o cache do "go mod download", acelerando muito os builds subsequentes.
COPY go.mod ./

# Pré-aquece o cache de dependências com base no go.mod.
RUN go mod download

# Copia todo o código-fonte do projeto para dentro do container.
COPY . .

# "go mod tidy" faz duas coisas importantes aqui:
#   1. Gera o go.sum com os hashes de segurança de cada dependência (necessário para compilar).
#   2. Remove dependências declaradas no go.mod que não são mais usadas no código.
# Ele precisa ser executado DEPOIS do "COPY . ." pois precisa ler os imports do código-fonte.
RUN go mod tidy

# Compila o binário Go com otimizações para produção:
#   CGO_ENABLED=0   → Desabilita CGo, produzindo um binário 100% estático (sem dependências de bibliotecas C do SO).
#   GOOS=linux      → Compila para Linux (necessário mesmo se você estiver desenvolvendo no macOS/Windows).
#   -ldflags        → Flags passadas ao linker:
#     -w              Omite a tabela de símbolos de debug (reduz o tamanho do binário).
#     -s              Omite o DWARF debug info (reduz ainda mais o tamanho).
#   -trimpath       → Remove caminhos absolutos do sistema de build do binário (melhora reprodutibilidade).
#   -o /app/server  → Define o nome e caminho do arquivo de saída.
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -trimpath \
    -o /app/server \
    ./main.go

# =============================================================================
# STAGE 2: "final" — Imagem de produção minimalista
# =============================================================================
# "scratch" é uma imagem completamente vazia — sem SO, sem shell, sem nada.
# Nossa imagem final terá APENAS o binário compilado e os certificados TLS.
# Isso resulta em imagens extremamente pequenas (~5-15MB) e seguras
# (sem binários desnecessários para exploração por atacantes).
FROM scratch

# Copia os certificados CA do stage builder para que chamadas HTTPS funcionem.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copia o binário compilado do stage builder para a imagem final.
COPY --from=builder /app/server /server

# Documenta que o container escuta na porta 8080 (não abre a porta, é apenas informativo).
EXPOSE 8080

# Define o comando a ser executado quando o container iniciar.
# Como não temos shell nesta imagem, usamos a forma "exec" (array JSON) em vez da forma "shell".
ENTRYPOINT ["/server"]
