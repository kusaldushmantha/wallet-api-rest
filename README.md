# Wallet API

## APIs

### Wallet related APIs

* Deposit API
    * Description: Deposits a provided amount to the given wallet uuid.
    * Endpoint: `POST http://localhost:8080/wallet/v1/{wallet-uuid}/deposit`
    * Mandatory headers: `X-User-ID = <user-uuid>`
    * Responses:
        * 200 - Success.
        * 400 - Bad Request - Mandatory header `X-User-ID` not provided.
        * 400 - Bad Request - Mandatory `amount` and `idempotency_token` json attributes not found in the payload.
          `amount` should be positive and `idempotency_token` should be a unique string per request.
        * 401 - Unauthorized - Provided `wallet-id` in the path does not belong to the provided `X-User-ID` user.
        * 409 - Conflict - Provided `idempotency_token` is already available indicating the request is a duplicate
          request.
        * 500 - Internal Server Error.
    * Sample curl:

```
curl --request POST \
  --url http://localhost:8080/wallet/v1/7dbacf5d-3099-4a66-ad3d-2fee93970017/deposit \
  --header 'content-type: application/json' \
  --header 'x-user-id: 0a644be3-cdf9-4491-b4ba-1cd8974c0278' \
  --data '{
  "amount": 250,
  "idempotency_token": "ccc"
}'
```

* Withdraw API
    * Description: Withdraw an amount from the given wallet uuid if there are enough funds in the wallet.
    * Endpoint: `POST http://localhost:8080/wallet/v1/{wallet-uuid}/withdraw`
    * Mandatory headers: `X-User-ID = <user-uuid>`
    * Responses:
        * 200 - Success.
        * 400 - Bad Request - Mandatory header `X-User-ID` not provided.
        * 400 - Bad Request - Mandatory `amount` and `idempotency_token` json attributes not found in the payload.
          `amount` should be positive and `idempotency_token` should be a unique string per request.
        * 401 - Unauthorized - Provided `wallet-id` in the path does not belong to the provided `X-User-ID` user.
        * 409 - Conflict - Provided `idempotency_token` is already available indicating the request is a duplicate
          request.
        * 500 - Internal Server Error.
    * Sample curl:

```
curl --request POST \
  --url http://localhost:8080/wallet/v1/7dbacf5d-3099-4a66-ad3d-2fee93970017/withdraw \
  --header 'content-type: application/json' \
  --header 'user-id: 0a644be3-cdf9-4491-b4ba-1cd8974c0278' \
  --data '{
  "amount": 50.75,
  "idempotency_token": "ccc"
}'
```

* Transfer API
    * Description: Transfer an amount from a wallet uuid to a given recipient wallet id given the source wallet has
      enough funds and the recipient wallet exists.
    * Endpoint: `POST http://localhost:8080/wallet/v1/{wallet-uuid}/transfer`
    * Mandatory headers: `X-User-ID = <user-uuid>`
    * Responses:
        * 200 - Success.
        * 400 - Bad Request - Mandatory header `X-User-ID` not provided.
        * 400 - Bad Request - Mandatory `amount`, `idempotency_token`, `recipient_wallet_id` json attributes not found
          in the payload. `amount` should be positive and `idempotency_token` should be a unique string per request.
          source account should have enough funds. `recipient_wallet_id` should be valid and existing.
        * 400 - Bad Request - Source and destination wallets are same.
        * 401 - Unauthorized - Provided `wallet-id` in the path does not belong to the provided `X-User-ID` user.
        * 409 - Conflict - Provided `idempotency_token` is already available indicating the request is a duplicate
          request.
        * 500 - Internal Server Error.
    * Sample curl:

```
curl --request POST \
  --url http://localhost:8080/wallet/v1/7dbacf5d-3099-4a66-ad3d-2fee93970017/transfer \
  --header 'content-type: application/json' \
  --header 'x-user-id: 0a644be3-cdf9-4491-b4ba-1cd8974c0278' \
  --data '{
  "amount":100,
  "recipient_wallet_id": "2cbcd158-56d2-4d45-8113-d51adf9ef57a",
  "idempotency_token": "gggg"
}'
```

* Get Balance API
    * Description: Retrieves the wallet balance.
    * Endpoint: `GET http://localhost:8080/wallet/v1/{wallet-uuid}/balance`
    * Mandatory headers: `X-User-ID = <user-uuid>`
    * Responses:
        * 200 - Success.
        * 400 - Bad Request - Mandatory header `X-User-ID` not provided.
        * 401 - Unauthorized - Provided `wallet-id` in the path does not belong to the provided `X-User-ID` user.
        * 500 - Internal Server Error.
    * Sample curl:

```
curl --request GET \
  --url http://localhost:8080/wallet/v1/2cbcd158-56d2-4d45-8113-d51adf9ef57a/balance \
  --header 'x-user-id: abe1f04a-68df-4e13-bd0d-5365ca9fdb0e'
```

* Get Transactions API
    * Description: Retrieves the transactions for the provided wallet in chronological descending order.
    * Endpoint: `GET http://localhost:8080/wallet/v1/{wallet-uuid}/transactions`
    * Mandatory headers: `X-User-ID = <user-uuid>`
    * Responses:
        * 200 - Success.
        * 400 - Bad Request - Mandatory header `X-User-ID` not provided.
        * 401 - Unauthorized - Provided `wallet-id` in the path does not belong to the provided `X-User-ID` user.
        * 500 - Internal Server Error.
    * Sample curl:

```
curl --request GET \
  --url http://localhost:8080/wallet/v1/7dbacf5d-3099-4a66-ad3d-2fee93970017/transactions \
  --header 'x-user-id: 0a644be3-cdf9-4491-b4ba-1cd8974c0278'
```

### User management APIs

**NOTE:** User management APIs are helper APIs to create users, associated wallets, and retrieve them. This is not for
evaluation.

* Create Users and Wallets API
    * Description: Creates and user and a wallet for that user and provides the user uuid and wallet uuid
    * Endpoint: `POST http://localhost:8080/user-management/v1`
    * Mandatory headers: N/A
    * Responses:
        * 200 - Success.
        * 500 - Internal Server Error.
    * Sample curl:

```
curl --request POST \
  --url http://localhost:8080/user-management/v1/
```

* Get Users and Wallets
    * Description: Retrieve user uuid and their associated wallet uuids.
    * Endpoint: `GET http://localhost:8080/user-management/v1`
    * Mandatory headers: N/A
    * Responses:
        * 200 - Success.
        * 500 - Internal Server Error.
    * Sample curl:

```
curl --request GET \
  --url http://localhost:8080/user-management/v1/
```

## Design Decisions
* API authentication not implemented. Therefore, wallet and user association is performed based on the `X-User-ID` header and the `waller-id` UUIDs provided.
* Transactional isolation level read-commited is used
