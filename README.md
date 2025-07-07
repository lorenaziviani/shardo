# Shardo: Distributed Cache System

## Motivação

Por que não usar Redis?

- Redis é excelente, mas pode ser um ponto único de falha e requer configuração de cluster para alta disponibilidade.
- Em alguns cenários, você pode precisar de um sistema mais simples, customizável ou que rode em ambientes restritos.
- O objetivo do Shardo é ser educativo, leve e fácil de adaptar para casos de uso específicos.

Quando usar caching distribuído?

- Quando múltiplos serviços precisam compartilhar dados em memória com baixa latência.
- Para escalar horizontalmente o cache sem depender de um único nó.
- Quando a tolerância a falhas e a distribuição de carga são requisitos.

## Arquitetura

```
Client → Gateway (HTTP) → Hash Ring → Node (gRPC) → Local Cache
```

- **Client**: Realiza requisições HTTP para o gateway.
- **Gateway**: Recebe requisições HTTP (GET, SET, DELETE, benchmark), utiliza o hash ring para decidir o nó responsável e faz proxy via gRPC para o nó correto.
- **Hash Ring (Consistent Hashing)**: Responsável por balancear e localizar o nó responsável por cada chave, usando réplicas virtuais para melhor distribuição e resiliência.
- **Node**: Instância do cache, responsável por armazenar parte dos dados e expor interface gRPC.
- **Local Cache**: Armazenamento em memória de cada nó, com TTL, LRU e métricas.

Veja o diagrama detalhado em `docs/architecture.drawio`.

## Fluxo de Requisições

1. O cliente faz uma requisição HTTP para o gateway (`/get`, `/set`, `/delete`).
2. O gateway usa o hash ring para decidir qual nó é responsável pela chave.
3. O gateway faz uma chamada gRPC para o nó correto.
4. O nó executa a operação no cache local e retorna o resultado.

## Endpoints do Gateway

- `GET /get?key=foo` — Busca o valor da chave `foo`.
- `POST /set?key=foo&ttl=60` — Define o valor da chave `foo` com TTL de 60 segundos (valor no corpo da requisição).
- `DELETE /delete?key=foo` — Remove a chave `foo`.
- `GET /benchmark?keys=1000` — Executa um benchmark de latência e distribuição de carga entre os nós.

## Exemplo de Uso

```sh
# Set
curl -X POST "http://localhost:8080/set?key=foo&ttl=60" -d 'bar'
# Get
curl "http://localhost:8080/get?key=foo"
# Delete
curl -X DELETE "http://localhost:8080/delete?key=foo"
# Benchmark
curl "http://localhost:8080/benchmark?keys=1000"
```

## Estrutura de Pastas

- `cmd/node`: Código do binário principal de cada nó.
- `cmd/gateway`: Código do binário do gateway HTTP.
- `cmd/hashring-cli`: CLI para testar o hash ring.
- `pkg/cache`: Implementação do cache local.
- `pkg/hashring`: Algoritmo de Consistent Hashing.
- `internal/grpc`: Servidor gRPC do nó.
- `internal/gateway`: Lógica do gateway HTTP.
- `proto/cachepb`: Código gerado do gRPC/protobuf.
- `infra/`: Scripts e arquivos de infraestrutura.
- `docs/`: Documentação e diagramas.

## Protocolos de Comunicação

- **Client → Gateway**: HTTP
- **Gateway → Node**: gRPC

## FAQ

**Q: Por que não usar um cache local simples?**
A: Em sistemas distribuídos, múltiplas instâncias precisam compartilhar o cache para garantir consistência e escalabilidade.

**Q: O sistema é pronto para produção?**
A: Não. O objetivo inicial é educacional e experimental.

---

Contribuições são bem-vindas!
