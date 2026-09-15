# Go Rate Limiter

Rate limiter HTTP desenvolvido em Go, utilizando Redis como mecanismo de persistência e middleware compatível com o roteador Chi.

O projeto permite limitar requisições por endereço IP ou por token de acesso, com suporte a bloqueio temporário após o excesso do limite.

## Tecnologias

- Go
- Chi
- Redis
- Docker
- Docker Compose
- Testify
- Mockery

## Funcionalidades

- Limitação de requisições por IP.
- Limitação de requisições por token.
- Precedência do token sobre o IP.
- Bloqueio temporário após exceder o limite.
- Configuração por variáveis de ambiente.
- Persistência dos contadores no Redis.
- Estratégia de persistência desacoplada por interface.
- Testes unitários.
- Execução via Docker Compose.

## Como funciona

O middleware analisa cada requisição recebida:

1. Verifica se existe um token no header `API_KEY`.
2. Se houver token, utiliza o limite configurado para o token.
3. Se não houver token, utiliza o limite configurado para o IP.
4. Caso o limite seja excedido, retorna HTTP `429 Too Many Requests`.
5. O IP ou token excedido pode permanecer bloqueado durante o período configurado.

### Header de autenticação

As requisições podem utilizar o header:

```http
API_KEY: meu-token
```

Quando o token está presente, o limite do token tem prioridade sobre o limite do IP.

Por exemplo:

```text
Limite por IP:    3 requisições por segundo
Limite por token: 10 requisições por segundo
```

Uma requisição com token poderá realizar até 10 requisições por segundo, mesmo que o IP tenha um limite menor.

## Configuração

A configuração é feita por variáveis de ambiente.

| Variável | Descrição | Exemplo |
|---|---|---|
| `REDIS_ADDR` | Endereço do Redis | `redis:6379` |
| `RATE_LIMIT_WINDOW` | Janela de contagem | `1s` |
| `RATE_LIMIT_BLOCK_TIME` | Tempo de bloqueio após exceder o limite | `5m` |
| `RATE_LIMIT_IP` | Limite de requisições por IP | `3` |
| `RATE_LIMIT_TOKEN` | Limite de requisições por token | `10` |

### Exemplo de configuração

```env
REDIS_ADDR=redis:6379
RATE_LIMIT_WINDOW=1s
RATE_LIMIT_BLOCK_TIME=5m
RATE_LIMIT_IP=3
RATE_LIMIT_TOKEN=10
```

Os valores de duração utilizam o formato aceito pelo pacote `time` do Go.

Exemplos:

```text
1s  = 1 segundo
30s = 30 segundos
5m  = 5 minutos
1h  = 1 hora
```

### Configuração pelo Docker Compose

O arquivo `docker-compose.yml` pode definir os valores padrão:

```yaml
services:
  app:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: go-rate-limiter
    ports:
      - "8080:8080"
    environment:
      REDIS_ADDR: ${REDIS_ADDR:-redis:6379}
      RATE_LIMIT_WINDOW: ${RATE_LIMIT_WINDOW:-1s}
      RATE_LIMIT_BLOCK_TIME: ${RATE_LIMIT_BLOCK_TIME:-5m}
      RATE_LIMIT_IP: ${RATE_LIMIT_IP:-3}
      RATE_LIMIT_TOKEN: ${RATE_LIMIT_TOKEN:-10}
    depends_on:
      - redis

  redis:
    image: redis:7-alpine
    container_name: go-rate-limiter-redis
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    restart: unless-stopped

volumes:
  redis_data:
```

Também é possível sobrescrever os valores diretamente no terminal:

```bash
RATE_LIMIT_IP=5 RATE_LIMIT_TOKEN=20 docker compose up --build
```

## Executando o projeto

### Pré-requisitos

- Docker
- Docker Compose

### Subir a aplicação

```bash
docker compose up --build
```

A aplicação ficará disponível em:

```text
http://localhost:8080
```

O Redis será iniciado automaticamente como dependência da aplicação.

### Executar em segundo plano

```bash
docker compose up --build -d
```

### Visualizar os logs

```bash
docker compose logs -f app
```

### Parar os serviços

```bash
docker compose down
```

Para remover também o volume persistente do Redis:

```bash
docker compose down -v
```

> A remoção do volume apaga os dados persistidos no Redis.

## Testando a aplicação

Com a aplicação em execução, faça uma requisição:

```bash
curl -i http://localhost:8080
```

### Testando o limite por IP

Considerando a configuração:

```env
RATE_LIMIT_IP=3
RATE_LIMIT_WINDOW=1s
```

Execute:

```bash
for i in {1..5}; do
  curl -i http://localhost:8080
done
```

As primeiras três requisições devem ser permitidas. As próximas devem retornar:

```http
HTTP/1.1 429 Too Many Requests
```

Com o corpo:

```text
you have reached the maximum number of requests or actions allowed within a certain time frame
```

Após o término da janela configurada, o contador será renovado.

### Testando o limite por token

Execute:

```bash
for i in {1..12}; do
  curl -i \
    -H "API_KEY: meu-token" \
    http://localhost:8080
done
```

Nesse caso, será utilizado o valor de `RATE_LIMIT_TOKEN`, pois o token possui precedência sobre o IP.

### Testando a precedência do token

Por exemplo:

```env
RATE_LIMIT_IP=3
RATE_LIMIT_TOKEN=10
```

As requisições abaixo utilizam o limite do token:

```bash
for i in {1..10}; do
  curl -i \
    -H "API_KEY: meu-token" \
    http://localhost:8080
done
```

Mesmo que o limite do IP seja `3`, as dez requisições com o mesmo token devem ser permitidas, desde que estejam dentro da mesma janela e o token ainda não esteja bloqueado.

## Executando os testes

Para executar os testes localmente:

```bash
go test ./...
```

Para visualizar os detalhes:

```bash
go test -v ./...
```

Os testes cobrem:

- Permissão de requisições dentro do limite.
- Bloqueio após exceder o limite por IP.
- Limite por token.
- Precedência do token sobre o IP.
- Bloqueio temporário.
- Interação com a estratégia de persistência.

### Executando os testes pelo Docker

Caso exista um serviço `test` no `docker-compose.yml`, os testes podem ser executados com:

```bash
docker compose run --rm test
```

Exemplo de serviço:

```yaml
test:
  build:
    context: .
    dockerfile: Dockerfile
  command: go test ./...
  environment:
    REDIS_ADDR: redis:6379
  depends_on:
    - redis
```

## Arquitetura

A lógica do rate limiter é separada do middleware HTTP.

```text
HTTP Request
     |
     v
Middleware
     |
     v
Rate Limiter
     |
     v
Cache Interface
     |
     v
Redis Adapter
     |
     v
Redis
```

### Middleware

O middleware é responsável por:

- Extrair o IP da requisição.
- Ler o token do header `API_KEY`.
- Encaminhar os dados para o limiter.
- Retornar HTTP `429` quando a requisição for bloqueada.
- Encaminhar a requisição para o próximo handler quando permitida.

O middleware não contém a lógica de contagem ou persistência.

### Rate Limiter

O rate limiter é responsável por:

- Definir qual limite deve ser aplicado.
- Priorizar o token quando ele estiver presente.
- Incrementar os contadores.
- Verificar bloqueios existentes.
- Criar bloqueios temporários após o excesso do limite.

### Estratégia de persistência

A persistência é definida por uma interface:

```go
type Cache interface {
    Increment(ctx context.Context, key string) (int64, error)
    Expire(ctx context.Context, key string, expiration time.Duration) error
    Exists(ctx context.Context, key string) (bool, error)
    Set(ctx context.Context, key string, value string, expiration time.Duration) error
}
```

A implementação atual utiliza Redis, mas a lógica do limiter depende apenas dessa interface.

## Como alterar a estratégia de persistência

Para utilizar outra tecnologia de persistência, basta criar uma nova implementação da interface `Cache`.

Por exemplo:

```go
type InMemoryCache struct {
    // implementação
}

func (c *InMemoryCache) Increment(
    ctx context.Context,
    key string,
) (int64, error) {
    // implementação
}

func (c *InMemoryCache) Expire(
    ctx context.Context,
    key string,
    expiration time.Duration,
) error {
    // implementação
}

func (c *InMemoryCache) Exists(
    ctx context.Context,
    key string,
) (bool, error) {
    // implementação
}

func (c *InMemoryCache) Set(
    ctx context.Context,
    key string,
    value string,
    expiration time.Duration,
) error {
    // implementação
}
```

Depois, a nova implementação pode ser injetada no limiter:

```go
limiter, err := ratelimiter.NewRedisLimiter(
    cache,
    window,
    blockTime,
    ipLimit,
    tokenLimit,
)
```

Embora o construtor atual tenha o nome `NewRedisLimiter`, a lógica de contagem utiliza a interface `Cache`. Uma possível evolução seria renomeá-lo para `NewLimiter`, tornando o nome independente da tecnologia utilizada.

## Chaves utilizadas no Redis

As chaves de contagem seguem o padrão:

```text
rate_limit_<identificador>
```

Exemplo para IP:

```text
rate_limit_192.168.0.1
```

Exemplo para token:

```text
rate_limit_meu-token
```

A chave de bloqueio utiliza o sufixo `_blocked`:

```text
rate_limit_192.168.0.1_blocked
rate_limit_meu-token_blocked
```

As chaves de contagem possuem o TTL definido por `RATE_LIMIT_WINDOW`.

As chaves de bloqueio possuem o TTL definido por `RATE_LIMIT_BLOCK_TIME`.

## Resposta quando o limite é excedido

Quando o limite é excedido, a aplicação retorna:

```http
HTTP/1.1 429 Too Many Requests
```

Corpo da resposta:

```text
you have reached the maximum number of requests or actions allowed within a certain time frame
```

## Estrutura do projeto

```text
.
├── README.md
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── go.sum
├── cmd
│   └── main.go
└── internal
    ├── adapters
    │   └── http
    │       └── middleware
    │           └── rate_limit.go
    ├── infrastructure
    │   └── cache
    │       └── cache.go
    └── ratelimiter
        └── limiter.go
```