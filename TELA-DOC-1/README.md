# TELA-DOC-1 - TELA Decentralized Web Standard Document <!-- omit in toc -->

## Introduction <!-- omit in toc -->
TELA introduces a standard for decentralized browser-based applications that can be executed locally, eliminating the reliance on third-party servers.

This portion of the documentation will focus on `TELA-DOC-1`.

`TELA-INDEX-1` info can be found [here](../TELA-INDEX-1/README.md) and it is recommended to have reviewed it prior to `TELA-DOC-1`.

## Creating TELA-DOC-1's<!-- omit in toc -->

### Table of Contents <!-- omit in toc -->
- [Languages](#languages)
- [Preparation](#preparation)
    - [Guidelines](#guidelines)
        - [JavaScript](#javascript)
    - [docTypes](#doctypes)    
- [Initialization](#initialization)
- [TELA Libraries](#tela-libraries)
    - [Library Usage](#library-usage)
    - [Library Creation](#library-creation)
- [DocShards](#docshards)
    - [DocShard Usage](#docshard-usage)
    - [DocShard Creation](#docshard-creation)
- [TELA-DOC-1 Template](#tela-doc-1-template)
- [Utilization](#utilization)
    - [Install TELA-DOC-1](#install-tela-doc-1)
    - [Rate TELA-DOC-1](#rate-tela-doc-1)
- [TELA-INDEX-1](#tela-index-1)

### Languages
The `civilware/tela` go package's accepted languages are:

- <b>HTML, JSON, JavaScript, CSS</b> and <b>Markdown</b>.
- A <b>Static</b> type is also defined to encompass desired files that are not listed as an accepted language such as asset or build files.

### Preparation
Prepare all the code (or text) that will be required for the document. Ensure you are using a language accepted by the host application that will be serving this document. There are some minimal guidelines to be followed when preparing the code to be added to the `TELA-DOC-1` smart contract which will help reduce errors. Outside of these guidelines, syntax and formatting can be defined by the author of the document. References and execution can be assumed to perform as if code were being served from any public server.

#### Guidelines
- One `TELA-DOC-1` cannot exceed 20KB in total size.
- Do not use any `/* */` multiline comments within any document or smart contract code as it may cause errors during contract installation.
- Example application code can be found in [tela_tests](../tela_tests/).
- The dURL can be used to help indexers query details beyond the defined contract stores, examples:
    - Appending `.lib` to a dURL will mark it as a library for indexes such as in [TELA-CLI](../cmd/tela-cli/README.md).

##### JavaScript
- Accurate origin URLs for web socket connections can be generated using:

```javascript
const applicationData = {
    "url": "http://localhost:" + location.port,
};
```
#### docTypes
```go
"TELA-STATIC-1"
"TELA-HTML-1"
"TELA-JSON-1"
"TELA-CSS-1"
"TELA-JS-1"
"TELA-MD-1"
```

### Initialization
It is recommended to use a compliant host application such as [TELA-CLI](../cmd/tela-cli/README.md) when installing a `TELA-DOC-1`, which will automate the process and help to avoid errors during installation. To deploy a `TELA-DOC-1` manually, developers can fill out the input fields inside of `InitializePrivate()` and input the document code in the designated multiline comment section at the bottom of the smart contract template.

```go
Function InitializePrivate() Uint64
...
// Input fields starts at line 30
30 STORE("nameHdr", "index.html") // nameHdr defines the name of the TELA document, following the ART-NFA headers standard.
31 STORE("descrHdr", "A HTML index") // descrHdr defines the description of the TELA document, following the ART-NFA headers standard.
32 STORE("iconURLHdr", "https://raw.githubusercontent.com/civilware/.github/main/CVLWR.png")  // iconURLHdr defines the url for the icon representing the TELA document, following the ART-NFA headers standard. This should be of size 100x100.
33 STORE("dURL", "app.tela") // dURL is a unique identifier for the TELA document likely linking to the TELA-INDEX-1 where this document is being used or to its corresponding library. 
34 STORE("docType", "TELA-HTML-1") // docType is the language or file type being used, ex TELA-JS-1, TELA-CSS-1... see docTypes list for all store values
35 STORE("subDir", "") // subDir adds this file to a sub directory, it can be left empty if file location is in root directory, separators should always be / ex: sub1/sub2/sub3
36 STORE("fileCheckC", "1c37f9e61f15a9526ba680dce0baa567e642ca2cd0ddea71649dab415dad8cb2") // C and S from DERO signature
37 STORE("fileCheckS", "7e642ca2cd0ddea71649dab415dad8cb21c37f9e61f15a9526ba680dce0baa56") // signature is of the docType code in the multiline comment section
// Input fields ends at line 37
100 RETURN 0
End Function
... 
/*
<!DOCTYPE html>
<html>

<head>
    <meta charset="UTF-8">
    <title>TELA-DOC-1 Template</title>
    <link rel="stylesheet" type="text/css" href="styles.css">
</head>

<body>
    <script src="functions.js"></script>
</body>

</html>
*/
```

### TELA Libraries
TELA libraries are in essence any development library that is installed in TELA format. A TELA library consists of `TELA-DOC-1` contracts that have been designed for universal use. Once installed, these libraries are intended to be application-agnostic, allowing their functionality to be leveraged by any TELA application. This promotes code reuse for faster development, helps to reduce chain bloat, drives community-tested solutions, and helps maintain consistency across different projects. To assist developers in discovering and utilizing installed libraries, some indexes like [TELA-CLI](../cmd/tela-cli/README.md) provide specific queries to make it easier to find and propagate universal libraries within the TELA ecosystem.

#### Library Usage
At its core, a TELA library is a `TELA-DOC-1` contract or a collection of them. This means that using a library follows the same principles as using any `TELA-DOC-1`.

To consume a library:
- Identify the SCID of the library you want to use in your application.
- While in the development stages, libraries can be cloned to get a local copy of the files so its functions and variables can be referenced and tested in the application before it is installed on-chain. 
- When development is finished, input any `TELA-INDEX-1` or `TELA-DOC-1` library SCID(s) used while developing, into your applications `TELA-INDEX-1` contract before it is installed. The application will now contain those libraries when served.

#### Library Creation
The codebase being installed as a TELA library might exceed the total size of a single `TELA-DOC-1` smart contract. In this case multiple `TELA-DOC-1`'s can be deployed and embedded within a `TELA-INDEX-1` to create a multi-part TELA library which can be referenced by a single SCID for further use.

To create a library:
- Write all the docType code needed for the `TELA-DOC-1` smart contract or contracts.
- Install the `TELA-DOC-1` contracts, ensuring that all the dURLs match and have a `.lib` suffix.
- Optionally, all the `TELA-DOC-1` contracts can then be embedded within a `TELA-INDEX-1` to reference it with a single SCID.

### DocShards
DocShards provide developers with an alternative method for packaging TELA content. They are similar to libraries in their construction; however, embedded DocShards are recreated as a single file when cloned or served. This allows a single piece of TELA content to exceed the smart contract installation maximum size. Additionally, DocShards can be embedded into libraries to enhance their utility.

#### DocShard Usage
Like TELA libraries, DocShards consist of a collection of `TELA-DOC-1` contracts. It is important to note that a DocShard cannot serve as the entrypoint of a `TELA-INDEX-1`.

To consume a DocShard:
- Identify the SCID of the DocShard you want to use in your application.
- While in the development stages, DocShards can be cloned to get a local copy of the file so its functions and variables can be referenced and tested in the application before it is installed on-chain. DocShards are cloned to `dURL/nameHdr` in the target directory.
- When development is finished, input any `TELA-INDEX-1` DocShard SCID(s) used while developing, into your applications `TELA-INDEX-1` contract before it is installed. The application will now contain those DocShards when served.

#### DocShard Creation
To avoid formatting errors during the cloning or serving of content, it is recommended to use the `civilware/tela` go package when creating DocShards. `TELA-CLI` has extended the package's tooling to simplify the creation of DocShards. Creation is limited to files containing only ASCII characters.

To create a DocShard:
- Write the docType code that will be recreated as a single source file. 
- Using the appropriate tools, create the DocShard smart contracts from the source file. 
- Install the `TELA-DOC-1` DocShard contracts. The `.shard` tag can be contained within the DOC dURL's to signify they are DocShards.
- Embed all installed `TELA-DOC-1` DocShard contracts into a `TELA-INDEX-1` and include `.shards` in its dURL to signify it requires reconstruction.

### TELA-DOC-1 Template
* [TELA-DOC-1](TELA-DOC-1.bas)

<i>The majority of comments on the smart contract template have been removed to utilize the maximum allowable contract space, refer to the above [Initialization](#initialization) documentation for guidance</i>.

### Utilization
It is in your best interest to always run a getgasestimate ahead of any direct curls below. This will 1) reduce your errors and 2) ensure you are utilizing the proper amount for fees from your specific parameters and ringsize. 

#### Install TELA-DOC-1
```json
curl --request POST --data-binary   @TELA-DOC-1.bas http://127.0.0.1:30000/install_sc;
```

#### Rate TELA-DOC-1
Rating `TELA-DOC-1` contracts can be done the same as `TELA-INDEX-1`'s. Manual rating procedures can be found [here](../TELA-INDEX-1/README.md#rate-tela-index-1).

### TELA-INDEX-1
* [TELA-INDEX-1](../TELA-INDEX-1/README.md)