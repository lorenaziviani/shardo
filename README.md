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
Client → Gateway → Hash Ring (Consistent Hashing) → Node → Local Cache
```

- **Client**: Realiza requisições de leitura/escrita de dados.
- **Gateway**: Ponto de entrada, roteia requisições para o nó correto via gRPC.
- **Hash Ring (Consistent Hashing)**: Responsável por balancear e localizar o nó responsável por cada chave, usando réplicas virtuais para melhor distribuição e resiliência.
- **Node**: Instância do cache, responsável por armazenar parte dos dados e expor interface gRPC.
- **Local Cache**: Armazenamento em memória de cada nó.

Veja o diagrama detalhado em `docs/architecture.drawio`.

## Consistent Hashing

O Shardo utiliza o algoritmo de Consistent Hashing para distribuir as chaves entre os nós do cluster. As principais características:

- **Réplicas Virtuais**: Cada nó é representado múltiplas vezes no anel de hash, melhorando a distribuição das chaves e reduzindo hotspots.
- **Adição/Remoção Dinâmica de Nós**: É possível adicionar ou remover nós do cluster com impacto mínimo na redistribuição das chaves.
- **Mapeamento de Chaves**: Uma função eficiente mapeia cada chave para o nó responsável, garantindo balanceamento e resiliência.

## CLI de Teste

Inclui uma CLI para testar a distribuição de chaves entre os nós. Exemplo de uso:

```sh
# Simula a distribuição de 1000 chaves entre 3 nós
shardo-hashring-cli --nodes node1,node2,node3 --keys 1000
```

A CLI mostra como as chaves são distribuídas e o impacto ao adicionar/remover nós.

## Estrutura de Pastas

- `cmd/node`: Código do binário principal de cada nó.
- `cmd/hashring-cli`: CLI para testar o hash ring.
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
