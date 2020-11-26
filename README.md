# key-vault

## How to use this project?

  1. Build the images and run the containers:

        ```sh
        $ docker-compose up -d --build
        ```

  2. Execute the container

        ```sh
        $ docker-compose exec vault bash
        ```

  3. Read the root token
 
        ```sh
        $ docker-compose exec -T vault cat /data/keys/vault.root.token
        ```

## Endpoints 


### LIST ACCOUNTS

This endpoint will list all accounts of key-vault.

| Method  | Path | Produces |
| ------------- | ------------- | ------------- |
| `LIST`  | `:mount-path/:network/accounts`  | `200 application/json` |


#### Sample Response

The example below shows output for a query path of `/ethereum/accounts` when there is 1 account.

```
{
    "request_id": "489790dc-b4bd-54e5-be6e-95a894ffc48c",
    "lease_id": "",
    "renewable": false,
    "lease_duration": 0,
    "data": {
        "accounts": [
            {
                "id": "9676ef06-d238-49f3-ab50-b3fe9930db0f",
                "name": "account-0",
                "validationPubKey": "8a5df36be5f89f9fe19cabadcbb17babc8c518bcd7fe0095c89f83915ea943343fa7dd3c26d8fb6096bce11fbc1ec7d3",
                "withdrawalPubKey": "887abb059075160ce2556a8bfef745898ee3a11b2b6521b09077d422c164929dea277ac8afcacd5b6d729198238f8f6c"
            }
        ]
    },
    "wrap_info": null,
    "warnings": null,
    "auth": null
}
```

### UPDATE STORAGE

This endpoint will update the storage.

| Method  | Path | Produces |
| ------------- | ------------- | ------------- |
| `POST`  | `:mount-path/:network/storage`  | `200 application/json` |


#### Sample Response

The example below shows output for a query path of `/ethereum/storage`.

```
{
    {
        "request_id": "d53d5075-6a3b-2642-ffde-0714beb595f5",
        "lease_id": "",
        "renewable": false,
        "lease_duration": 0,
        "data": {
            "status": true
        },
        "wrap_info": null,
        "warnings": null,
        "auth": null
    }
}
```

### UPDATE SLASHING STORAGE

This endpoint will update the storage.

| Method  | Path | Produces |
| ------------- | ------------- | ------------- |
| `POST`  | `:mount-path/:network/storage/slashing`  | `200 application/json` |


#### Sample Request

The example below shows input for a query path of `/ethereum/storage/slashing`.

```
{
    "<public_key>": "<hex_encoded_slashing_storage>"
}
```


#### Sample Response

The example below shows output for a query path of `/ethereum/storage/slashing`.

```
{
    "request_id": "d53d5075-6a3b-2642-ffde-0714beb595f5",
    "lease_id": "",
    "renewable": false,
    "lease_duration": 0,
    "data": {
        "status": true
    },
    "wrap_info": null,
    "warnings": null,
    "auth": null
}
```

### READ SLASHING STORAGE

This endpoint will update the storage.

| Method  | Path | Produces |
| ------------- | ------------- | ------------- |
| `GET`  | `:mount-path/:network/storage/slashing`  | `200 application/json` |


#### Sample Response

The example below shows output for a query path of `/ethereum/storage/slashing`.

```
{
    {
        "request_id": "d53d5075-6a3b-2642-ffde-0714beb595f5",
        "lease_id": "",
        "renewable": false,
        "lease_duration": 0,
        "data": {
            "<public_key>": "<hex_encoded_slashing_storage>"
        },
        "wrap_info": null,
        "warnings": null,
        "auth": null
    }
}
```

### SIGN ATTESTATION

This endpoint will sign attestation for specific account at a path.

| Method  | Path | Produces |
| ------------- | ------------- | ------------- |
| `POST`  | `:mount-path/:network/accounts/sign-attestation`  | `200 application/json` |

#### Parameters

* `public_key` (`string: <required>`) - Specifies the public key of the account to sign.
* `domain` (`string: <required>`) - Specifies the domain.
* `slot` (`int: <required>`) - Specifies the slot.
* `committeeIndex` (`int: <required>`) - Specifies the committeeIndex.
* `beaconBlockRoot` (`string: <required>`) - Specifies the beaconBlockRoot.
* `sourceEpoch` (`int: <required>`) - Specifies the sourceEpoch.
* `sourceRoot` (`string: <required>`) - Specifies the sourceRoot.
* `targetEpoch` (`int: <required>`) - Specifies the targetEpoch.
* `targetRoot` (`string: <required>`) - Specifies the targetRoot.

#### Sample Response

The example below shows output for the successful sign of `/ethereum/accounts/account1/sign-attestation`.

```
{
    "request_id": "b767dcca-5b10-4a52-1d9a-0a9b81b378ae",
    "lease_id": "",
    "renewable": false,
    "lease_duration": 0,
    "data": {
        "signature": "kEEOMxNkouz7EOSULfrG6hXzZbIOvRCVVK+lfBofj3U49/PHm7YHji8ac9Gf9vgEFVEmbPp+lhO3OpAElt3yaBajTKaJBWocgXuv64Ojq44tfxLJo6jrzMU5yoP78dYW"
    },
    "wrap_info": null,
    "warnings": null,
    "auth": null
}
```

### SIGN PROPOSAL

This endpoint will sign attestation for specific account at a path.

| Method  | Path | Produces |
| ------------- | ------------- | ------------- |
| `POST`  | `:mount-path/<pyrmont|mainnet>/accounts/sign-proposal`  | `200 application/json` |

#### Parameters

* `public_key` (`string: <required>`) - Specifies the public key of the account to sign.
* `domain` (`string: <required>`) - Specifies the domain.
* `slot` (`int: <required>`) - Specifies the slot.
* `proposerIndex` (`int: <required>`) - Specifies the proposerIndex.
* `parentRoot` (`string: <required>`) - Specifies the parentRoot.
* `stateRoot` (`string: <required>`) - Specifies the stateRoot.
* `bodyRoot` (`string: <required>`) - Specifies the bodyRoot.

#### Sample Response

The example below shows output for the successful sign of `/ethereum/accounts/account1/sign-proposal`.

```
{
    "request_id": "b767dcca-5b10-4a52-1d9a-0a9b81b378ae",
    "lease_id": "",
    "renewable": false,
    "lease_duration": 0,
    "data": {
        "signature": "kPyCp8ID44ceUB3KSp+7QsxqTlGSP2u6/cytr04qJyxkkIKIO/FW57qwH9E7/c48D1PgHsyb8hgoT8/jOLMD7Y/Jt06Qiw80ZRtoS78CzMFYRut/OQot+FzAJcW7Jk0U"
    },
    "wrap_info": null,
    "warnings": null,
    "auth": null
}
```

### SIGN AGGREGATION

This endpoint will sign attestation for specific account at a path.

| Method  | Path | Produces |
| ------------- | ------------- | ------------- |
| `POST`  | `:mount-path/:network/accounts/sign-aggregation`  | `200 application/json` |

#### Parameters

* `public_key` (`string: <required>`) - Specifies the public key of the account to sign.
* `domain` (`string: <required>`) - Specifies the domain.
* `dataToSign` (`string: <required>`) - Specifies the slot.

#### Sample Response

The example below shows output for the successful creation of `/ethereum/accounts/account1/sign-aggregation`.

```
{
    "request_id": "b767dcca-5b10-4a52-1d9a-0a9b81b378ae",
    "lease_id": "",
    "renewable": false,
    "lease_duration": 0,
    "data": {
        "signature": "kPyCp8ID44ceUB3KSp+7QsxqTlGSP2u6/cytr04qJyxkkIKIO/FW57qwH9E7/c48D1PgHsyb8hgoT8/jOLMD7Y/Jt06Qiw80ZRtoS78CzMFYRut/OQot+FzAJcW7Jk0U"
    },
    "wrap_info": null,
    "warnings": null,
    "auth": null
}
```

## Access Policies
The plugin's endpoint paths are designed such that admin-level access policies vs. signer-level access policies can be easily separated.

### Sample Signer Level Policy:
Use the following policy to assign to a signer level access token, with the abilities to list accounts and sign transactions.

```
# Ability to list existing wallet accounts ("list")
path "ethereum/pyrmont/accounts" {
  capabilities = ["list"]
}
path "ethereum/mainnet/accounts" {
  capabilities = ["list"]
}

# Ability to sign data ("create")
path "ethereum/pyrmont/accounts/sign-*" {
  capabilities = ["create"]
}
path "ethereum/mainnet/accounts/sign-*" {
  capabilities = ["create"]
}
```

### Sample Admin Level Policy:
Use the following policy to assign to a admin level access token, with the full ability to update storage, list accounts and sign transactions.

```
# Ability to list existing wallet accounts ("list")
path "ethereum/pyrmont/accounts" {
  capabilities = ["list"]
}
path "ethereum/mainnet/accounts" {
  capabilities = ["list"]
}

# Ability to sign data ("create")
path "ethereum/pyrmont/accounts/sign-*" {
  capabilities = ["create"]
}
path "ethereum/mainnet/accounts/sign-*" {
  capabilities = ["create"]
}

# Ability to update storage ("create")
path "ethereum/pyrmont/storage" {
  capabilities = ["create"]
}
path "ethereum/mainnet/storage" {
  capabilities = ["create"]
}
```

## How to use policies?

  1. Create a new policy named admin:

        ```sh
        $ vault policy write admin policies/admin-policy.hcl
        ```

  2. Create a token attached to admin policy:

        ```sh
        $ vault token create -policy="admin"
        ```

  3. Create a new policy named signer:
 
        ```sh
        $ vault policy write signer policies/signer-policy.hcl
        ```

  4. Create a token attached to signer policy:

        ```sh
        $ vault token create -policy="signer"
        ```

## About testing

There are 2 types of tests in the project: end-to-end and unit ones. 
In order to run all tests including e2e ones you will need to do the following command:
```bash
$ make test
``` 

New e2e tests should be placed in `./e2e/tests` directory and implement `E2E` interface.
Use the current format to add new tests. 


## Release Version

versions are published to dockerhub based on tags.
before publishing a tag update docker compose image to the to be pushed tag 

## Multinetworks

The plugin supports multiple Ethereum networks. All available networks are defined in `./config/vault-plugin.sh`.
New networks could be defined by the following steps:

1. Enable secrets for a new network in `./config/vault-plugin.sh`. 
    Example
    ```bash
    $ vault secrets enable \
        -path=ethereum/pyrmont \
        -description="Eth Signing Wallet - Pyrmont Test Network" \
        -plugin-name=ethsign plugin > /dev/null 2>&1
    ```

2. Update policies `./policies/admin-policy.hcl` and `./policies/signer-policy.hcl` by adding a definition with a new network in the path.
