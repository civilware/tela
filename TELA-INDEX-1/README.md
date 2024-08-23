# TELA-INDEX-1 - TELA Decentralized Web Standard <!-- omit in toc -->

## Introduction <!-- omit in toc -->
TELA introduces a standard for decentralized browser-based applications that can be executed locally, eliminating the need for third-party servers. 

This portion of the documentation will focus on `TELA-INDEX-1`. This is the recommended starting point for learning about installing TELA applications and content.

## Getting started with TELA-INDEX-1<!-- omit in toc -->

### Table of Contents <!-- omit in toc -->
- [Initialization](#initialization)
- [TELA-INDEX-1 Template](#tela-index-1-template)
- [Utilization](#utilization)
    - [Install TELA-INDEX-1](#install-tela-index-1)
    - [Update TELA-INDEX-1](#update-tela-index-1)
    - [Rate TELA-INDEX-1](#rate-tela-index-1)
- [TELA-DOC-1](../TELA-DOC-1/README.md)

### Initialization
To deploy a `TELA-INDEX-1` manually, developers can fill out the input fields inside of `InitializePrivate()`. The minimum requirement for deployment is a valid `TELA-DOC-1` SCID, which must be input as the "DOC1" STORE value, and will serve as the entrypoint for the TELA application.

```go
Function InitializePrivate() Uint64
... 
// Input fields starts at line 30
30 STORE("nameHdr", "App Name") // nameHdr defines the name of the TELA application, following the ART-NFA headers standard.
31 STORE("descrHdr", "A TELA App") // descrHdr defines the description of the TELA application, following the ART-NFA headers standard.
32 STORE("iconURLHdr", "https://raw.githubusercontent.com/civilware/.github/main/CVLWR.png") // iconURLHdr defines the URL for the icon representing the TELA application, following the ART-NFA headers standard. This should be of size 100x100.
33 STORE("dURL", "app.tela") // dURL is unique identifier for the TELA application, ex myapp.tela.
40 STORE("DOC1", "a891299086d218840d1eb71ae759ddc08f1e85cbf35801cc34ef64b4b07939c9") // DOC#s are installed TELA-DOC-1 SCIDs to be used in this TELA application. DOC1 will be used as the entrypoint for the application, a valid DOC1 k/v store is the minimum requirement for a TELA application.
41 STORE("DOC2", "b891299086d218840d1eb71ae759ddc08f1e85cbf35801cc34ef64b4b07939c8") // Further DOCs can be added/removed as required. All DOC stores that are used on the contract must have a valid SCID value otherwise the INDEX will not be validated when serving.
// For any further DOCs needed, make sure the smart contract line number and DOC# increment++ 
// 42 STORE("DOC3", "c891299086d218840d1eb71ae759ddc08f1e85cbf35801cc34ef64b4b07939c7")
// 43 STORE("DOC4", "d891299086d218840d1eb71ae759ddc08f1e85cbf35801cc34ef64b4b07939c6")
// ..
// Input fields end at the last DOC#
...
1000 RETURN 0
End Function
```
It is recommended to keep `TELA-INDEX-1` contracts below 9KB total file size if updating is to be required at any point in the future. If staying within the 9KB size limit, developers are able to embed ~90 (<i>may vary slightly depending on header values</i>) `TELA-DOC-1` SCIDs into a single `TELA-INDEX-1` and maintain its ability to be updated with more DOCs than it was originally installed with. Current test show the maximum limit of DOCs that a `TELA-INDEX-1` can be successfully installed with is ~120 (<i>may vary slightly depending on header values</i>). Library usage details are documented within the `TELA-DOC-1` section and can increase the total capacity beyond these figures. 

### TELA-INDEX-1 Template
* [TELA-INDEX-1](TELA-INDEX-1.bas)

 <i>The majority of comments on the smart contract template have been removed to utilize the maximum allowable contract space, refer to the above [Initialization](#initialization) documentation for guidance</i>.

### Utilization
It is in your best interest to always run a getgasestimate ahead of any direct curls below. This will 1) reduce your errors and 2) ensure you are utilizing the proper amount for fees from your specific parameters and ringsize. 

#### Install TELA-INDEX-1
```json
curl --request POST --data-binary   @TELA-INDEX-1.bas http://127.0.0.1:30000/install_sc;
```

#### Update TELA-INDEX-1
The update portion will be split with a getgasestimate example call first, and then the follow-up with the respective fees that were present in that exact scenario. It's up to you to modify the fees parameter to reflect the 'gasstorage' return of getgasestimate.

- GetGasEstimate
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
        "sc_rpc": [{
            "name": "SC_ACTION",
            "datatype": "U",
            "value": 0
        },
        {
            "name": "SC_ID",
            "datatype": "H",
            "value": "ce25b92083f089357d72295f4cf51cc58fed7439500792b94c85244f1067279e"
        },
        {
            "name": "entrypoint",
            "datatype": "S",
            "value": "UpdateCode"
        },
        {
            "name": "code",
            "datatype": "S",
            "value": "NEW_TELA_INDEX_SC_CODE_GOES_HERE"
        }]
    }
}'
```

- Txn
```json
curl -X POST\
    http://127.0.0.1:30000/json_rpc\
    -H 'content-type: application/json'\
    -d '{
    "jsonrpc":"2.0",
    "id":"0",
    "method":"Transfer",
    "params":{
        "ringsize":2,
        "fees":194,
        "sc_rpc":[{
            "name":"entrypoint",
            "datatype":"S",
            "value":"UpdateCode"
        },
        {
            "name":"code",
            "datatype":"S",
            "value":"NEW_TELA_INDEX_SC_CODE_GOES_HERE"
        },
        {
            "name":"SC_ACTION",
            "datatype":"U",
            "value":0
        },
        {
            "name":"SC_ID",
            "datatype":"H",
            "value":"ce25b92083f089357d72295f4cf51cc58fed7439500792b94c85244f1067279e"
        }]
    }
}'
```

#### Rate TELA-INDEX-1
The following Rate command also applies to `TELA-DOC-1` contracts.

- GetGasEstimate
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
        "sc_rpc": [{
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
            "value": "Rate"
        },
        {
            "name": "r",
            "datatype": "U",
            "value": 0
        }]
    }
}'
```

- Txn
```json
curl -X POST\
    http://127.0.0.1:30002/json_rpc\
    -H 'content-type: application/json'\
    -d '{
    "jsonrpc":"2.0",
    "id":"0",
    "method":"Transfer",
    "params":{
        "ringsize":2,
        "fees":90,
        "sc_rpc":[{
            "name":"entrypoint",
            "datatype":"S",
            "value":"Rate"
        },
        {
            "name":"r",
            "datatype":"U",
            "value": 0
        },
        {
            "name":"SC_ACTION",
            "datatype":"U",
            "value":0
        },
        {
            "name":"SC_ID",
            "datatype":"H",
            "value":"f2815b442d62a055e4bb8913167e3dbce3208f300d7006aaa3a2f127b06de29d"
        }]
    }
}'
```

### TELA-DOC-1
* [TELA-DOC-1](../TELA-DOC-1/README.md)