# Wallet API

A RESTful service for managing digital wallets, supporting operations such as deposits, withdrawals, transfers, balance
inquiries, and transaction history retrieval. Built with [Go Fiber](https://gofiber.io/) for high performance and rapid
development.

---

## Table of Contents

- [API Endpoints](#api-endpoints)
    - [Wallet Operations](#wallet-operations)
    - [User Management](#user-management)
- [Design Overview](#design-overview)
- [Setup & Running](#setup--running)
- [Testing](#testing)
- [Known Limitations](#known-limitations)
- [Future Improvements](#future-improvements)
- [Project Structure Overview](#project-structure-overview)

---

## API Endpoints

### Wallet Operations

#### 1. Deposit

- **Endpoint:** `POST /wallet/v1/{wallet-uuid}/deposit`
- **Headers:**
    - `X-User-ID: <user-uuid>`
- **Sample request:**
  ```shell
  curl --request POST \
  --url http://localhost:8080/wallet/v1/7dbacf5d-3099-4a66-ad3d-2fee93970017/deposit \
  --header 'content-type: application/json' \
  --header 'x-user-id: 0a644be3-cdf9-4491-b4ba-1cd8974c0278' \
  --data '{
      "amount": 250,
      "idempotency_token": "unique-idempotency-token"
    }'
  ```
- **Responses:**
    - `200 OK` – Deposit successful.
    - `400 Bad Request` – Missing headers or payload attributes.
    - `401 Unauthorized` – Wallet does not belong to the user.
    - `409 Conflict` – Duplicate `idempotency_token`.
    - `500 Internal Server Error` - Service error

#### 2. Withdraw

- **Endpoint:** `POST /wallet/v1/{wallet-uuid}/withdraw`
- **Headers:**
    - `X-User-ID: <user-uuid>`
- **Sample request:**

```shell
curl --request POST \
--url http://localhost:8080/wallet/v1/7dbacf5d-3099-4a66-ad3d-2fee93970017/withdraw \
--header 'content-type: application/json' \
--header 'x-user-id: 0a644be3-cdf9-4491-b4ba-1cd8974c0278' \
--data '{
    "amount": 5,
    "idempotency_token": "unique-idempotency-token"
  }'
```

- **Responses:**
    - `200 OK` – Withdrawal successful.
    - `400 Bad Request` – Missing headers or payload attributes.
    - `401 Unauthorized` – Wallet does not belong to the user.
    - `409 Conflict` – Duplicate `idempotency_token`.
    - `500 Internal Server Error` - Service error.

#### 3. Transfer

- **Endpoint:** `POST /wallet/v1/{wallet-uuid}/transfer`
- **Headers:**
    - `X-User-ID: <user-uuid>`
- **Sample request:**
  ```shell
  curl --request POST \
  --url http://localhost:8080/wallet/v1/7dbacf5d-3099-4a66-ad3d-2fee93970017/transfer \
  --header 'content-type: application/json' \
  --header 'x-user-id: 0a644be3-cdf9-4491-b4ba-1cd8974c0278' \
  --data '{
      "amount":100,
      "recipient_wallet_id": "2cbcd158-56d2-4d45-8113-d51adf9ef57a",
      "idempotency_token": "unique-token"
    }'
  ```
- **Responses:**
    - `200 OK` – Transfer successful.
    - `400 Bad Request` – Missing headers or payload attributes, or invalid transfer details.
    - `401 Unauthorized` – Wallet does not belong to the user.
    - `409 Conflict` – Duplicate `idempotency_token`.
    - `500 Internal Server Error` - Server error.

#### 4. Get Balance

- **Endpoint:** `GET /wallet/v1/{wallet-uuid}/balance`
- **Headers:**
    - `X-User-ID: <user-uuid>`
- **Sample request:**
  ```shell
  curl --request GET \
  --url http://localhost:8080/wallet/v1/7dbacf5d-3099-4a66-ad3d-2fee93970017/balance \
  --header 'x-user-id: 0a644be3-cdf9-4491-b4ba-1cd8974c0278'
  ```
- **Responses:**
    - `200 OK` – Returns current wallet balance.
    - `400 Bad Request` – Missing `X-User-ID` header.
    - `401 Unauthorized` – Wallet does not belong to the user.
    - `500 Internal Server Error`

#### 5. Get Transactions

- **Endpoint:** `GET /wallet/v1/{wallet-uuid}/transactions`
- **Headers:**
    - `X-User-ID: <user-uuid>`
- **Sample request:**

```shell
curl --request GET \
  --url http://localhost:8080/wallet/v1/7dbacf5d-3099-4a66-ad3d-2fee93970017/transactions?limit=10&offset=0 \
  --header 'x-user-id: 0a644be3-cdf9-4491-b4ba-1cd8974c0278'
```

- **Responses:**
    - `200 OK` – Returns list of transactions in descending chronological order.
    - `400 Bad Request` – Missing `X-User-ID` header.
    - `401 Unauthorized` – Wallet does not belong to the user.
    - `500 Internal Server Error`

### User Management

> **Note:** These endpoints are auxiliary to help create users and their wallets easily. Please do not evaluate these
> endpoints.

#### 1. Create User and Wallet

- **Endpoint:** `POST /user-management/v1/`
- **Responses:**
    - `200 OK` – Returns newly created user and wallet UUIDs.
    - `500 Internal Server Error`

#### 2. Get Users and Wallets

- **Endpoint:** `GET /user-management/v1/`
- **Responses:**
    - `200 OK` – Returns list of users and their associated wallets.
    - `500 Internal Server Error`

---

## Effort

> This service design and implementation was done in 80 hours.

---

## Design Overview

> **Note:** Design concentration is purely for the functional and non-functional requirements of the wallet service.

- **Framework:** Utilizes [Go Fiber](https://gofiber.io/), inspired by Express.js, for building fast and scalable web
  applications.
- **Authentication:** Currently, no request authentication mechanism is implemented. User identification is based on the
  `X-User-ID` header.
- **Concurrency Control:**
    - Employs `SERIALIZABLE` isolation level to prevent concurrency issues when updating wallets and transaction related information.
    - Uses `SELECT ... FOR UPDATE` statements to lock rows during transactions, ensuring data consistency.
- **Idempotency:**
    - Implements idempotency tokens to prevent duplicate create and update operations.
    - Tokens are stored in Redis with a TTL to manage their lifecycle.
- **Architecture:**
    - Adheres to the [Dependency Inversion Principle](https://en.wikipedia.org/wiki/Dependency_inversion_principle),
      promoting decoupled and modular code. Low level modules such as databases and caches are exposed via interfaces
      and are decoupled from the application.
    - Uses Go best practises in project structure and maintaining modularity.
- **Validation and Error Handling:**
    - All request payloads and attributes undergo validation.
    - Errors are handled gracefully with appropriate HTTP status codes and messages.
- **Database Integrity:**
    - Triggers in the database ensure that wallet balances do not go below zero after transactions. See `./migration/init.sql` for schema related info.
- **Containerization:**
    - The application is containerized using Docker for consistent deployment across environments.
- **Logging and monitoring:**
    - Information and errors logged properly with appropriate log levels.
    - All the transaction specific information logged properly without exposing sensitive data to support reconciliation
      actions on database transaction mismatch.
- **Performance:**
    - All external connectors such as cache and database has proper connection and read timeouts set.
    - Idempotency keys stored in redis for faster retrieval and TTL based clean up and uses atomic `SETNX` redis
      operation.
    - Instead of setting database level transactional isolation level, query level isolation is used for better performance.

---

## Setup & Running

### Prerequisites

- [Docker](https://www.docker.com/) installed on your machine.

### Running the Service

1. **Start the application:**
   ```bash
   docker-compose up
   ```
2. **Accessing the API:**
    - Base URL: `http://localhost:8080/wallet/v1/`
    - If you are using [Bruno](https://www.usebruno.com/) api client, API collection can be found in
      `./bruno-api-collection` directory
3. **Stopping the application:**
   ```bash
   docker-compose down
   ```

### Seed Data

Upon initialization, the following users and wallets are created with a balance of 0:

- **User 01:**
    - User UUID: `0a644be3-cdf9-4491-b4ba-1cd8974c0278`
    - Wallet UUID: `7dbacf5d-3099-4a66-ad3d-2fee93970017`
- **User 02:**
    - User UUID: `abe1f04a-68df-4e13-bd0d-5365ca9fdb0e`
    - Wallet UUID: `2cbcd158-56d2-4d45-8113-d51adf9ef57a`

---

## Testing

- **Unit Tests:** Written using `gomock` and `testify`.
- **Mocks:** Mocks generated with `mockgen`
- **Integration Tests:** Use PostgreSQL test containers.
- **Run Tests:**
  ```bash
  make test
  ```

---

## Known Limitations

- No proper request authentication for endpoints.
- No caching on retrieval endpoints.
- Requires database lookups for wallet and user existence validation.
- Uses row locking with `SELECT ... FOR UPDATE` queries instead of optimistic locking.

---

## Future Improvements

- Add proper request authentication with cookies or JWT tokens.
- Implement write-through caching that can be leveraged in retrieval endpoints. Write through caching is used to ensure
  consistency with the database.
- Consider Bloom filters and optimistic locking for faster database operations on concurrency and lookups.
- Improve payload validation and observability.

---

## Local Development

```bash
make build
make run
```

## Project Structure Overview

- `api/routes.go` - API route handling and payload validation.
- `api/setup.go` - Service setup logic.
- `api/utils.go` - Utility functions used for validation and for functionality within `api` directory.
- `bruno-api-collection` - API collection used to test the service with Bruno API client.
- `cmd/server/main.go` - Environment variable loading and service startup.
- `commons/constants.go` - Constants used within the service.
- `commons/types.go` - Common types used within the service.
- `config/postgresql.go` - Configuration related to PostgreSQL database.
- `config/redis.go` - Configuration related to Redis.
- `db/mocks/*` - Mocks related to PostgreSQL and Redis.
- `db/postgres.go` - PostgreSQL client and relational queries.
- `db/redis.go` - Redis client and implementation.
- `db/types.go` - Interface containing methods for the database and cache.
- `migration/init.sql` - Initial database seed script and schema script.
- `models/requests/types.go` - Request types.
- `models/responses/types.go` - Response types.
- `models/transaction.go` - Transaction table model.
- `models/user.go` - User table model.
- `models/wallet.go` - Wallet table model.
- `services/mocks/*` - Mocks for services.
- `services/user-service.go` - User service implementation.
- `services/wallet-service.go` - Wallet service implementation.
- `services/types.go` - Types used for services.
- `services/helpers.go` - Helpers used within services.
- `.env.docker` - Environment file to use in docker.
- `.env.local` - Environment file to use for local development.
