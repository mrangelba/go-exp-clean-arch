# Desafio Clean Architecture

## Como executar

### Executar docker compose

```sh
docker compose up
```
ou
```sh
make run
```

## Portas dos serviços

Todos os serviços (API REST, GraphQL e GRPC) estão utilizando a porta 8080

### API REST

Retorna a listagem de Orders

GET http://localhost:8080/order

Cria uma Order

POST http://localhost:8080/order

### GraphQL

Acesso ao Playground

http://localhost:8080/graphql

Query

http://localhost:8080/graphql/query


### GRPC

Host: localhost

Port: 8080

#### Executando com o Evans
```sh
evans --host localhost -p 8080

call ListOrders
```