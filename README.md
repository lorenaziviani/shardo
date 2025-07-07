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
Client → Gateway → Hash Ring → Node → Local Cache
```

- **Client**: Realiza requisições de leitura/escrita de dados.
- **Gateway**: Ponto de entrada, roteia requisições para o nó correto via gRPC.
- **Hash Ring**: Algoritmo de Consistent Hashing para balancear e localizar o nó responsável.
- **Node**: Instância do cache, responsável por armazenar parte dos dados e expor interface gRPC.
- **Local Cache**: Armazenamento em memória de cada nó.

Veja o diagrama detalhado em `docs/architecture.drawio`.

## Estrutura de Pastas

- `cmd/node`: Código do binário principal de cada nó.
- `pkg/cache`: Implementação do cache local.
- `pkg/hashring`: Algoritmo de Consistent Hashing.
- `internal/grpc`: Handlers e servidor gRPC.
- `infra/`: Scripts e arquivos de infraestrutura.
- `docs/`: Documentação e diagramas.

## Protocolos de Comunicação

- Toda comunicação entre gateway e nós será feita via **gRPC** (alta performance, tipagem forte, fácil integração entre linguagens).

## FAQ

**Q: Por que não usar um cache local simples?**
A: Em sistemas distribuídos, múltiplas instâncias precisam compartilhar o cache para garantir consistência e escalabilidade.

**Q: O sistema é pronto para produção?**
A: Não. O objetivo inicial é educacional e experimental.

---

Contribuições são bem-vindas!
