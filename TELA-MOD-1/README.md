# TELA-MOD-1 - TELA Decentralized Web Standard Modules<!-- omit in toc -->

## Introduction <!-- omit in toc -->
TELA introduces a standard for decentralized browser-based applications that can be executed locally, eliminating the reliance on third-party servers.

This portion of the documentation focuses on `TELA-MOD-1`, which serve as additional building blocks that developers can use within their `TELA-INDEX-1` contracts to extend functionality for TELA applications.

Information on `TELA-INDEX-1` can be found [here](../TELA-INDEX-1/README.md). It is recommended to review it before proceeding with `TELA-MOD-1`.

## Getting Started with TELA-MOD-1<!-- omit in toc -->

### Table of Contents <!-- omit in toc -->
- [What Is TELA-MOD-1?](#what-is-tela-mod-1)
- [Utilization](#utilization)
- [MODs](#mods)
    - [MOD Classes](#mod-classes)
        - [Class Rules](#class-rules)
    - [MOD Tags](#mod-tags)
    - [Adding New MODs](#adding-new-mods)
- [Available MODs](#available-mods)
    - [Variable Store MODs](#variable-store-mods)
        - [Get Variable](#get-variable)
        - [Set Variable](#set-variable)
        - [Delete Variable](#delete-variable)
    - [Transfers MODs](#transfers-mods)
        - [Deposit and Withdraw DERO](#deposit-and-withdraw-dero)
        - [Deposit and Withdraw Assets](#deposit-and-withdraw-assets)
        - [Transfer Ownership](#transfer-ownership)
- [TELA-INDEX-1](../TELA-INDEX-1/README.md)
- [TELA-DOC-1](../TELA-DOC-1/README.md)
- [TELA-CLI](../cmd/tela-cli/README.md)

### What Is TELA-MOD-1?
`TELA-MOD-1` provides a set of predefined [DVM](https://docs.dero.io/developer/dvm.html) modules that can be embedded into `TELA-INDEX-1` smart contracts to extend their functionality. These modules help keep the base smart contract small and efficient while allowing developers to add extended features as needed. The `MODs` can be added or removed from a smart contract when it is installed or updated. TELA validates the injected `MOD` code, ensuring that any functions in the smart contract and their data points remain consistent with the TELA standard.

### Utilization
It is recommended to use a compliant host application such as [TELA-CLI](../cmd/tela-cli/README.md) when enabling `MODs`, which automates the process and helps to avoid errors during installation or updates. To embed a `MOD` manually, developers can include the `MOD` code within their `TELA-INDEX-1`, keeping in mind that TELA's parser will invalidate any contract with functions that do not match `TELA-MOD-1` or have conflicting `MODs`.

### MODs
The `civilware/tela` go package has established a foundational structure for `MODs`, enabling their continuous expansion as the standard evolves, without requiring changes to smart contract versioning. A `MOD` consists of:
- **Name**: The name of the `MOD`
- **Tag**: A short identifier for the `MOD`
- **Description**: What the `MOD` does
- **Function Code**: The DVM function code for the `MOD`
- **Function Names**: Names of any functions within the function code

#### MOD Classes
`MODs` are divided into classes based on their utility. A `MODClass` consists of:
- **Name**: The name of the `MODClass`, highlighting its utility
- **Tag**: A short identifier for the `MODClass`
- **MODs**: All of the `MODs` which can be enabled
- **Rules**: To maintain order and eliminate conflicts within themselves, `MODClasses` can define a set of rules that emphasize their best results and help steer innovative ideas towards successful `MOD` development.

##### Class Rules
A `MODClassRule` consists of:
- **Name**: The name of the `MODClassRule`
- **Description**: What the `MODClassRule` enforces
- **Verification Function**: Golang code verifying that the rule parameters are met

| Rule Name      | Description                                                     |
|----------------|-----------------------------------------------------------------|
| **Single MOD** | Only one `MOD` from a specific `MODClass` can be used at a time.|
| **Multi MOD**  | Multiple `MODs` from the same `MODClass` can be used simultaneously. |

Using these combined structures, the `civilware/tela` package has defined these available `MODClasses`:

| Class Name     | Tag | MODS           | Rules      |
|----------------|-----|----------------|------------|
| Variable Store | vs  | [Source](vs/)  | Single MOD |
| Transfers      | tx  | [Source](tx/)  | Multi MOD  |

#### MOD Tags
A `MOD` is most easily identified by its tag. A `TELA-INDEX-1` with `MODs` will have a variable stored containing a tag for each `MOD` that is enabled. The mod tag is formatted as `tag1,tag2,tag3`. When TELA parses a smart contract, it will look for any stored tags and use them in its validation process. The `MODClass` tag defines the prefix for all its members' tags. For example, all `MODs` within the *variable store* `MODClass` will prefix their tags with `vs`. Tags can be checked prior to use with:
```go
import (
	"github.com/civilware/tela"
)

func main() {
	_, err := tela.Mods.TagsAreValid("vsoo,vspubsu,vspubow")
	if err != nil {
		// Returns the conflict as an error
	}
}
```

#### Adding New MODs
New `MODs` can be added to the `civilware/tela` package using a new `MODClass`. The documentation will describe the process of adding new `MODs` in the style that the package implements them, aiming to aid any contributors who wish to expand the package's `MOD` library. If adding new `MODs` while using `civilware/tela` as an import, this same process can be followed. The file embed portion can be substituted for any code string.

- Create a DVM function or functions as a bas file. If a new `MODClass` is being started, create a directory for it inside `TELA-MOD-1/`, otherwise the file can be placed in its respective `MODClass` directory. For example, additions to the *variable store* `MODClass` should go in `TELA-MOD-1/vs/`.
```go
Function NewFunction1() Uint64
10 RETURN 0
End Function
```

- In the `mods.go` file, embed the new bas file as a variable alongside the existing `MODs`.
```go
//go:embed */nc/TELA-MOD-1-NCONE.bas
var TELA_MOD_1_NCONE string
```

- In `initMods()` create a new `MODClass` and all its `MODs`. Passing the newly created `MODClass` and `MODs` to the `Mods.Add` method will add them to the package if there are no conflicts caused by the new additions. 
```go 
newClass := MODClass{
    Name:  "New Class",
    Tag:   "nc",
    Rules: []MODClassRule{Mods.rules[0]},
}

newMODs := []MOD{
    {
        Name:          "New Class Function 1",
        Tag:           newClass.NewTag("one"), // NewTag() prefixes "one" with "nc" to create its MOD tag of "ncone"
        Description:   "The NewFunction1 function returns 0, it is part of a README example",
        FunctionCode:  func() string { return TELA_MOD_1_NCONE }, // The DVM function code for this MOD
        FunctionNames: []string{"NewFunction1"},
    },
    // If adding MODs to an existing MODClass directly in the package, MODs can simply be added to the target MODClass's []MOD in initMods()
}

err := Mods.Add(newClass, newMODs)
if err != nil {
    // Handle error
}
```

- The new `MOD` would now be available within its `MODClass` and can be accessed by its tag.
```go
Mods.GetMod("ncone")
Mods.GetClass("nc")
```

### Available MODs
The following documentation sections cover the specifics of the available `MODs`. Host applications can get data for the `civilware/tela` package's available `MODs` using:
```go
import (
	"fmt"

	"github.com/civilware/tela"
)

func main() {
	for _, mod := range tela.Mods.GetAllMods() {
		fmt.Println()
		fmt.Printf("Name: %s\n", mod.Name)
		fmt.Printf("Tag: %s\n", mod.Tag)
		fmt.Printf("Description: %s\n", mod.Description)
		fmt.Println()
		fmt.Println(mod.FunctionCode())
	}
}
```


The manual transaction examples will be split with a getgasestimate example call first, followed by a transaction with the respective fees present in that exact scenario. It's up to you to modify the fees parameter to reflect the 'gasstorage' return of getgasestimate.

#### Variable Store MODs
The variable store `MODClass` focuses on storing arbitrary data in TELA smart contracts.
- **Owner Only**: Allows the owner of the smart contract to set and delete custom variable stores.
- **Owner Only Immutable**: Allows the owner of the smart contract to set custom variable stores that cannot be changed or deleted.
- **Public Single Use**: Allows all wallets to store variables that the wallet cannot change, with the owner having the ability to set and delete all variables.
- **Public Overwrite**: Allows all wallets to store variables that can be overwritten by the wallet, with the owner having the ability to set and delete all variables.
- **Public Immutable**: Allows all wallets to store variables that cannot be changed or deleted.

The variable store `MODClass` has defined some common stores which can be universally interpreted.
| Key             | Value          | Usage Description                                              | Examples                                                  |
|-----------------|----------------|----------------------------------------------------------------|-----------------------------------------------------------|
| var_owners_note | String         | Commentary from the smart contract owner about the application | "Deprecated", "Offline", "Something I want users to know" |

All the variable store `MODs` use a common API for their functions and data retrieval.

##### Get Variable
Variables stored have their keys prefixed to avoid any conflicts with the TELA smart contract base variables. When the owner of the smart contract stores a variable, its key will be prefixed with `var_`. When a non-owner of a smart contract stores a variable (public variable stores MODs only), their keys will include the wallet address in the prefix, ex: `var_deto1qyre7td6x9r88y4c...`. Immutable variables will be prefixed similarly for owners and non-owners with `ivar_`. All stored variables can be retrieved from the smart contract with a `DERO.GetSC` call.

```json
curl -X POST\
  http://127.0.0.1:20000/json_rpc\
  -H 'content-type: application/json'\
  -d '{
    "jsonrpc": "2.0",
    "id": "1",
    "method": "DERO.GetSC",
    "params": {
        "scid": "f2815b442d62a055e4bb8913167e3dbce3208f300d7006aaa3a2f127b06de29d",
        "code": false,
        "variables": false,
        "keysstring": ["var_myKey"]
    }
}'
```

##### Set Variable
- SetVar GetGasEstimate
```json
curl -X POST\
  http://127.0.0.1:20000/json_rpc\
  -H 'content-type: application/json'\
  -d '{
    "jsonrpc": "2.0",
    "id": "1",
    "method": "DERO.GetGasEstimate",
    "params": {
        "transfers": [],
        "signer": "deto1qyre7td6x9r88y4cavdgpv6k7lvx6j39lfsx420hpvh3ydpcrtxrxqg8v8e3z",
        "sc_rpc": [
            {
                "name": "SC_ACTION",
                "datatype": "U",
                "value": 0
            },
            {
                "name": "SC_ID",
                "datatype": "H",
                "value": "f2815b442d62a055e4bb8913167e3dbce3208f300d7006aaa3a2f127b06de29d"
            },
            {
                "name": "entrypoint",
                "datatype": "S",
                "value": "SetVar"
            },
            {
                "name": "k",
                "datatype": "S",
                "value": "myKey"
            },
            {
                "name": "v",
                "datatype": "S",
                "value": "myValue"
            }
        ]
    }
}'
```

- SetVar Txn
```json
curl -X POST\
    http://127.0.0.1:30001/json_rpc\
    -H 'content-type: application/json'\
    -d '{
    "jsonrpc": "2.0",
    "id": "0",
    "method": "Transfer",
    "params": {
        "ringsize": 2,
        "fees": 120,
        "sc_rpc": [
            {
                "name": "entrypoint",
                "datatype": "S",
                "value": "SetVar"
            },
            {
                "name": "k",
                "datatype": "S",
                "value": "myKey"
            },
            {
                "name": "v",
                "datatype": "S",
                "value": "myValue"
            },
            {
                "name": "SC_ACTION",
                "datatype": "U",
                "value": 0
            },
            {
                "name": "SC_ID",
                "datatype": "H",
                "value": "f2815b442d62a055e4bb8913167e3dbce3208f300d7006aaa3a2f127b06de29d"
            }
        ]
    }
}'
```

##### Delete Variable
- DeleteVar GetGasEstimate
```json
curl -X POST\
  http://127.0.0.1:20000/json_rpc\
  -H 'content-type: application/json'\
  -d '{
    "jsonrpc": "2.0",
    "id": "1",
    "method": "DERO.GetGasEstimate",
    "params": {
        "transfers": [],
        "signer": "deto1qyre7td6x9r88y4cavdgpv6k7lvx6j39lfsx420hpvh3ydpcrtxrxqg8v8e3z",
        "sc_rpc": [
            {
                "name": "SC_ACTION",
                "datatype": "U",
                "value": 0
            },
            {
                "name": "SC_ID",
                "datatype": "H",
                "value": "f2815b442d62a055e4bb8913167e3dbce3208f300d7006aaa3a2f127b06de29d"
            },
            {
                "name": "entrypoint",
                "datatype": "S",
                "value": "DeleteVar"
            },
            {
                "name": "k",
                "datatype": "S",
                "value": "myKey"
            }
        ]
    }
}'
```

- DeleteVar Txn
```json
curl -X POST\
    http://127.0.0.1:30001/json_rpc\
    -H 'content-type: application/json'\
    -d '{
    "jsonrpc": "2.0",
    "id": "0",
    "method": "Transfer",
    "params": {
        "ringsize": 2,
        "fees": 100,
        "sc_rpc": [
            {
                "name": "entrypoint",
                "datatype": "S",
                "value": "DeleteVar"
            },
            {
                "name": "k",
                "datatype": "S",
                "value": "myKey"
            },
            {
                "name": "SC_ACTION",
                "datatype": "U",
                "value": 0
            },
            {
                "name": "SC_ID",
                "datatype": "H",
                "value": "f2815b442d62a055e4bb8913167e3dbce3208f300d7006aaa3a2f127b06de29d"
            }
        ]
    }
}'
```

#### Transfers MODs
The transfers `MODClass` focuses on general transferring. This includes DERO, assets or the smart contract itself.

##### Deposit and Withdraw DERO
Stores DERO deposits and allows the owner to withdraw DERO from the smart contract.

- DepositDero GetGasEstimate
```json
curl -X POST\
  http://127.0.0.1:20000/json_rpc\
  -H 'content-type: application/json'\
  -d '{
    "jsonrpc": "2.0",
    "id": "1",
    "method": "DERO.GetGasEstimate",
    "params": {
        "transfers": [
            {
                "destination": "deto1qyvyeyzrcm2fzf6kyq7egkes2ufgny5xn77y6typhfx9s7w3mvyd5qqynr5hx",
                "burn": 10000
            }
        ],
        "signer": "deto1qyre7td6x9r88y4cavdgpv6k7lvx6j39lfsx420hpvh3ydpcrtxrxqg8v8e3z",
        "sc_rpc": [
            {
                "name": "SC_ACTION",
                "datatype": "U",
                "value": 0
            },
            {
                "name": "SC_ID",
                "datatype": "H",
                "value": "d9e1972456dc7e39f0b63f43bfa310de8d1716b4ececcb1d1b5c81b96f4350dc"
            },
            {
                "name": "entrypoint",
                "datatype": "S",
                "value": "DepositDero"
            }
        ]
    }
}'
```

- DepositDero Txn
```json
curl -X POST\
    http://127.0.0.1:30001/json_rpc\
    -H 'content-type: application/json'\
    -d '{
    "jsonrpc": "2.0",
    "id": "0",
    "method": "Transfer",
    "params": {
        "transfers": [
            {
                "destination": "deto1qyvyeyzrcm2fzf6kyq7egkes2ufgny5xn77y6typhfx9s7w3mvyd5qqynr5hx",
                "burn": 10000
            }
        ],
        "ringsize": 2,
        "fees": 90,
        "sc_rpc": [
            {
                "name": "entrypoint",
                "datatype": "S",
                "value": "DepositDero"
            },
            {
                "name": "SC_ACTION",
                "datatype": "U",
                "value": 0
            },
            {
                "name": "SC_ID",
                "datatype": "H",
                "value": "d9e1972456dc7e39f0b63f43bfa310de8d1716b4ececcb1d1b5c81b96f4350dc"
            }
        ]
    }
}'
```

- WithdrawDero GetGasEstimate
```json
curl -X POST\
  http://127.0.0.1:20000/json_rpc\
  -H 'content-type: application/json'\
  -d '{
    "jsonrpc": "2.0",
    "id": "1",
    "method": "DERO.GetGasEstimate",
    "params": {
        "transfers": [],
        "signer": "deto1qyre7td6x9r88y4cavdgpv6k7lvx6j39lfsx420hpvh3ydpcrtxrxqg8v8e3z",
        "sc_rpc": [
            {
                "name": "SC_ACTION",
                "datatype": "U",
                "value": 0
            },
            {
                "name": "SC_ID",
                "datatype": "H",
                "value": "d9e1972456dc7e39f0b63f43bfa310de8d1716b4ececcb1d1b5c81b96f4350dc"
            },
            {
                "name": "entrypoint",
                "datatype": "S",
                "value": "WithdrawDero"
            },
            {
                "name": "amt",
                "datatype": "U",
                "value": 100
            }
        ]
    }
}'
```

- WithdrawDero Txn
```json
curl -X POST\
    http://127.0.0.1:30001/json_rpc\
    -H 'content-type: application/json'\
    -d '{
    "jsonrpc": "2.0",
    "id": "0",
    "method": "Transfer",
    "params": {
        "transfers": [],
        "ringsize": 2,
        "fees": 100,
        "sc_rpc": [
            {
                "name": "entrypoint",
                "datatype": "S",
                "value": "WithdrawDero"
            },
            {
                "name": "amt",
                "datatype": "U",
                "value": 100
            },
            {
                "name": "SC_ACTION",
                "datatype": "U",
                "value": 0
            },
            {
                "name": "SC_ID",
                "datatype": "H",
                "value": "d9e1972456dc7e39f0b63f43bfa310de8d1716b4ececcb1d1b5c81b96f4350dc"
            }
        ]
    }
}'
```

##### Deposit and Withdraw Assets
Stores asset deposits and allows the owner to withdraw assets from the smart contract.

- DepositAsset GetGasEstimate
```json
curl -X POST\
  http://127.0.0.1:20000/json_rpc\
  -H 'content-type: application/json'\
  -d '{
    "jsonrpc": "2.0",
    "id": "1",
    "method": "DERO.GetGasEstimate",
    "params": {
        "transfers": [
            {
                "scid": "d33b7f82d0f291f6347a049e7dd6a8471aa6a40e39cef443edce30b147494790",
                "burn": 10000
            }
        ],
        "signer": "deto1qyre7td6x9r88y4cavdgpv6k7lvx6j39lfsx420hpvh3ydpcrtxrxqg8v8e3z",
        "sc_rpc": [
            {
                "name": "SC_ACTION",
                "datatype": "U",
                "value": 0
            },
            {
                "name": "SC_ID",
                "datatype": "H",
                "value": "96523e98c16432f2ccb95767afc6853365e74980b08a618d65c40093e2f2b6c1"
            },
            {
                "name": "entrypoint",
                "datatype": "S",
                "value": "DepositAsset"
            },
            {
                "name": "scid",
                "datatype": "S",
                "value": "d33b7f82d0f291f6347a049e7dd6a8471aa6a40e39cef443edce30b147494790"
            }
        ]
    }
}'
```

- DepositAsset Txn
```json
curl -X POST\
    http://127.0.0.1:30001/json_rpc\
    -H 'content-type: application/json'\
    -d '{
    "jsonrpc": "2.0",
    "id": "0",
    "method": "Transfer",
    "params": {
        "transfers": [
            {
                "scid": "d33b7f82d0f291f6347a049e7dd6a8471aa6a40e39cef443edce30b147494790",
                "burn": 10000
            }
        ],
        "ringsize": 2,
        "fees": 170,
        "sc_rpc": [
            {
                "name": "entrypoint",
                "datatype": "S",
                "value": "DepositAsset"
            },
            {
                "name": "scid",
                "datatype": "S",
                "value": "d33b7f82d0f291f6347a049e7dd6a8471aa6a40e39cef443edce30b147494790"
            },
            {
                "name": "SC_ACTION",
                "datatype": "U",
                "value": 0
            },
            {
                "name": "SC_ID",
                "datatype": "H",
                "value": "96523e98c16432f2ccb95767afc6853365e74980b08a618d65c40093e2f2b6c1"
            }
        ]
    }
}'
```

- WithdrawAsset GetGasEstimate
```json
curl -X POST\
  http://127.0.0.1:20000/json_rpc\
  -H 'content-type: application/json'\
  -d '{
    "jsonrpc": "2.0",
    "id": "1",
    "method": "DERO.GetGasEstimate",
    "params": {
        "signer": "deto1qyre7td6x9r88y4cavdgpv6k7lvx6j39lfsx420hpvh3ydpcrtxrxqg8v8e3z",
        "sc_rpc": [
            {
                "name": "SC_ACTION",
                "datatype": "U",
                "value": 0
            },
            {
                "name": "SC_ID",
                "datatype": "H",
                "value": "96523e98c16432f2ccb95767afc6853365e74980b08a618d65c40093e2f2b6c1"
            },
            {
                "name": "entrypoint",
                "datatype": "S",
                "value": "WithdrawAsset"
            },
            {
                "name": "scid",
                "datatype": "S",
                "value": "d33b7f82d0f291f6347a049e7dd6a8471aa6a40e39cef443edce30b147494790"
            },
            {
                "name": "amt",
                "datatype": "U",
                "value": 100
            }
        ]
    }
}'
```

- WithdrawAsset Txn
```json
curl -X POST\
    http://127.0.0.1:30001/json_rpc\
    -H 'content-type: application/json'\
    -d '{
    "jsonrpc": "2.0",
    "id": "0",
    "method": "Transfer",
    "params": {
        "transfers": [],
        "ringsize": 2,
        "fees": 180,
        "sc_rpc": [
            {
                "name": "entrypoint",
                "datatype": "S",
                "value": "WithdrawAsset"
            },
            {
                "name": "scid",
                "datatype": "S",
                "value": "d33b7f82d0f291f6347a049e7dd6a8471aa6a40e39cef443edce30b147494790"
            },
            {
                "name": "amt",
                "datatype": "U",
                "value": 100
            },
            {
                "name": "SC_ACTION",
                "datatype": "U",
                "value": 0
            },
            {
                "name": "SC_ID",
                "datatype": "H",
                "value": "96523e98c16432f2ccb95767afc6853365e74980b08a618d65c40093e2f2b6c1"
            }
        ]
    }
}'
```

##### Transfer Ownership
Allows the transfer of smart contract ownership to another wallet.

- TransferOwnership GetGasEstimate
```json
curl -X POST\
  http://127.0.0.1:20000/json_rpc\
  -H 'content-type: application/json'\
  -d '{
    "jsonrpc": "2.0",
    "id": "1",
    "method": "DERO.GetGasEstimate",
    "params": {
        "transfers": [],
        "signer": "deto1qyre7td6x9r88y4cavdgpv6k7lvx6j39lfsx420hpvh3ydpcrtxrxqg8v8e3z",
        "sc_rpc": [
            {
                "name": "SC_ACTION",
                "datatype": "U",
                "value": 0
            },
            {
                "name": "SC_ID",
                "datatype": "H",
                "value": "96523e98c16432f2ccb95767afc6853365e74980b08a618d65c40093e2f2b6c1"
            },
            {
                "name": "entrypoint",
                "datatype": "S",
                "value": "TransferOwnership"
            },
            {
                "name": "addr",
                "datatype": "S",
                "value": "deto1qyre7td6x9r88y4cavdgpv6k7lvx6j39lfsx420hpvh3ydpcrtxrxqg8v8e3z"
            }
        ]
    }
}'
```

- TransferOwnership Txn
```json
curl -X POST\
    http://127.0.0.1:30001/json_rpc\
    -H 'content-type: application/json'\
    -d '{
    "jsonrpc": "2.0",
    "id": "0",
    "method": "Transfer",
    "params": {
        "ringsize": 2,
        "fees": 240,
        "sc_rpc": [
            {
                "name": "entrypoint",
                "datatype": "S",
                "value": "TransferOwnership"
            },
            {
                "name": "addr",
                "datatype": "S",
                "value": "deto1qyre7td6x9r88y4cavdgpv6k7lvx6j39lfsx420hpvh3ydpcrtxrxqg8v8e3z"
            },
            {
                "name": "SC_ACTION",
                "datatype": "U",
                "value": 0
            },
            {
                "name": "SC_ID",
                "datatype": "H",
                "value": "96523e98c16432f2ccb95767afc6853365e74980b08a618d65c40093e2f2b6c1"
            }
        ]
    }
}'
```

- ClaimOwnership GetGasEstimate
```json
curl -X POST\
  http://127.0.0.1:20000/json_rpc\
  -H 'content-type: application/json'\
  -d '{
    "jsonrpc": "2.0",
    "id": "1",
    "method": "DERO.GetGasEstimate",
    "params": {
        "transfers": [],
        "signer": "deto1qyre7td6x9r88y4cavdgpv6k7lvx6j39lfsx420hpvh3ydpcrtxrxqg8v8e3z",
        "sc_rpc": [
            {
                "name": "SC_ACTION",
                "datatype": "U",
                "value": 0
            },
            {
                "name": "SC_ID",
                "datatype": "H",
                "value": "96523e98c16432f2ccb95767afc6853365e74980b08a618d65c40093e2f2b6c1"
            },
            {
                "name": "entrypoint",
                "datatype": "S",
                "value": "ClaimOwnership"
            }
        ]
    }
}'
```

- ClaimOwnership Txn
```json
curl -X POST\
    http://127.0.0.1:30001/json_rpc\
    -H 'content-type: application/json'\
    -d '{
    "jsonrpc": "2.0",
    "id": "0",
    "method": "Transfer",
    "params": {
        "ringsize": 2,
        "fees": 170,
        "sc_rpc": [
            {
                "name": "entrypoint",
                "datatype": "S",
                "value": "ClaimOwnership"
            },
            {
                "name": "SC_ACTION",
                "datatype": "U",
                "value": 0
            },
            {
                "name": "SC_ID",
                "datatype": "H",
                "value": "96523e98c16432f2ccb95767afc6853365e74980b08a618d65c40093e2f2b6c1"
            }
        ]
    }
}'
```

### TELA-INDEX-1
[TELA-INDEX-1](../TELA-INDEX-1/README.md)

### TELA-DOC-1
[TELA-DOC-1](../TELA-DOC-1/README.md)

### TELA-CLI
[TELA-CLI](../cmd/tela-cli/README.md)