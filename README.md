![](https://github.com/civilware/tela/blob/main/tela.png?raw=true)

## TELA: Decentralized Web Standard <!-- omit in toc -->

### Table of Contents <!-- omit in toc -->
- [What is TELA?](#what-is-tela)
- [How it works](#how-it-works)
- [Additional features](#additional-features)
- [Get started](#get-started)
    - [TELA-INDEX-1](TELA-INDEX-1/README.md)
    - [TELA-DOC-1](TELA-DOC-1/README.md)
- [Accessing TELA content](#accessing-tela-content)
    - [Compliant Host Applications](#compliant-host-applications)
- [Content rating system](#content-rating-system)
- [Package use](#package-use)
	- [Serving](#serving)
	- [Installing](#installing)
	- [Updating](#updating)
	- [Rating](#rating)
	- [Parse](#parse)
- [TELA-CLI](cmd/tela-cli/README.md)
- [Changelog](CHANGELOG.md)
- [License](LICENSE)

### What is TELA?
TELA enables the secure and decentralized storage of application files on [DERO's](https://dero.io) blockchain using smart contracts. This innovative standard ensures the integrity and authenticity of stored files, allowing them to be retrieved from the blockchain and executed locally in a browser. By doing so, TELA enhances user privacy and control over browser-based applications, eliminating the reliance on third-party servers.

### How it works
TELA applications are built on two key smart contract components:

`TELA-INDEX-1` Contract: This serves as the entrypoint for TELA applications. Users can simply input the SCID (Smart Contract ID) of any deployed `TELA-INDEX-1` contract to retrieve the necessary files to run the application locally. This process can be demonstrated using tools like the `civilware/tela` [go package](https://pkg.go.dev/github.com/Civilware/tela) or similar methods to parse a `TELA-INDEX-1` according to this standard, creating a <b>host app</b> for all TELA content.

`TELA-DOC-1` Contract: This contains the essential code required by the application. TELA supports common programming languages such as, but not limited to:

- <b>HTML
- JSON
- JavaScript
- CSS
- Markdown</b>

Multiple `TELA-DOC-1` contracts can be installed and embedded within a `TELA-INDEX-1` application, allowing the use of extensive codebases beyond DERO’s [DVM-BASIC](https://docs.dero.io/developer/dvm.html) smart contract language. These contracts can also install necessary libraries and tools on the blockchain, facilitating modular development through reusable code.

### Additional features
- File Management: TELA ensures application code remains in an unalterable state using a combination of mutable (`TELA-INDEX-1`) and immutable (`TELA-DOC-1`) contracts. This structure provides a commit-based system that allows for code updates, verification, and retrieval of previous contract states.

- Connectivity: TELA supports DERO's XSWD protocol, enabling permissioned web socket interactions with DERO wallets, enhancing connectivity and user interaction.

- Contract Identification: TELA contracts utilize the header structure from the [ART-NFA Standard](https://github.com/civilware/artificer-nfa-standard/blob/main/Headers/README.md), making them easily identifiable and ensuring consistent integration within the DERO ecosystem.

By leveraging these components and features, TELA represents a significant advancement in secure, decentralized web applications, fostering an environment where privacy, security, and user autonomy are paramount.

### Get started
See the following for more information on how to get started creating and installing your TELA application:

* [TELA-INDEX-1](TELA-INDEX-1/README.md)

  Entrypoint and manifest for TELA applications
  
* [TELA-DOC-1](TELA-DOC-1/README.md)

  TELA application files and libraries

### Accessing TELA content
The minimum requirements to access TELA content are:
* Connection to a DERO node
* Any <b>host application</b> that can serve TELA content

The `civilware/tela` go package aims to be a simple entrypoint for hosting and managing TELA content and can be used as the foundation for creating compliant <b>host applications</b>.

#### Compliant Host Applications
* [Engram](https://github.com/DEROFDN/Engram/)
* [TELA-CLI](cmd/tela-cli/README.md)

### Content rating system
TELA smart contracts have a rating system integrated into each contract to help developers and users navigate content. Ratings can be interpreted for a quick judgment, or can be used to gather further details about content. A guide to the content rating system is as follows:
- One rating per DERO account, per contract.
- A rating is a positive number < 100.
- Ringsize of 2 is required for a Rate transaction.
- The smart contract stores the rating number and the address of the rater.
- The smart contract also adds to a <i>likes</i> or <i>dislikes</i> value store based on the given rating, dislike is a rating < 50.
- Standard gas fees apply when making a Rate transaction.

A quick tally of the like and dislike values or gathering the average rating number can give a brief overview of how the content has been received by others. For a more detailed review of content, TELA has provided a structure for the 0-99 rating numbers, attaching them to a wider range of subjective comments while still supporting the basic "higher number is better" mentality for overall ease of use. The `civilware/tela` package can generate rating strings using the following interpretation:

Numbers are broken down into place values. 
- Ex, 24 = first place 2 and second place 4.
- Ex, 8 = first place 0 and second place 8.

Each place is given different values. The first place represents the <i>rating category</i>. 

<b>Rating categories</b>

| First Place | Category           |
|-------------|--------------------|
| 0           | Do not use         |
| 1           | Broken             |
| 2           | Major issues       |
| 3           | Minor issues       |
| 4           | Should be improved |
| 5           | Could be improved  |
| 6           | Average            |
| 7           | Good               |
| 8           | Very good          |
| 9           | Exceptional        |

The second place represents a <i>detail tag</i>. Positive and negative rating categories each have their own set of detail tags, sharing some common ones between them.

<b>Detail tags</b>
| Second Place | Negative Detail Tags | Positive Detail Tags |
|--------------|----------------------|----------------------|
| 0            | Nothing              | Nothing              |
| 1            | Needs review         | Needs review         |
| 2            | Needs improvement    | Needs improvement    |
| 3            | Bugs                 | Bugs                 |
| 4            | Errors               | Errors               |
| 5            | Inappropriate        | Visually appealing   |
| 6            | Incomplete           | In depth             |
| 7            | Corrupted            | Works well           |
| 8            | Plagiarized          | Unique               |
| 9            | Malicious            | Benevolent           |

This format would produce the following strings given some example rating numbers:

| Rating       | String                      |
|--------------|-----------------------------|
| 80           | Very good                   |
| 77           | Good, Works well            |
| 43           | Should be improved, Bugs    |
| 7            | Do not use, Corrupted       |

### Package use
The main usage of the `civilware/tela` go package is to query a deployed `TELA-INDEX-1` SCID from a connected node and serve the content on a URL such as `localhost:8081/tela/`

#### Serving
```go
import (
	"github.com/civilware/tela"
)

func main() {
	endpoint := "127.0.0.1:20000"
	scid := "a842dac04587000b019a7aeee55d7e3e5df40f959b0bd36a474cda67936e9399"
	url, err := tela.ServeTELA(scid, endpoint)
	if err != nil {
		// Handle error
	}
	// Code to open url in local browser
	// ..
	// Shutdown all TELA servers when done
	tela.ShutdownTELA()
}
```
#### Installing
TELA content can be installed in a manner of ways. The `civilware/tela` package takes the necessary data for the type of smart contract and creates the applicable transfer arguments and code to easily install the new contract. For manual installation see [here](TELA-DOC-1/README.md#install-tela-doc-1).
```go
import (
	"fmt"

	"github.com/civilware/tela"
	"github.com/deroproject/derohe/rpc"
	"github.com/deroproject/derohe/walletapi"
)

func main() {
	// Create DOC (or INDEX) with all relevant data
	doc := &tela.DOC{
		DocType: tela.DOC_HTML,
		Code:    "<HTML_CODE_HERE>",
		SubDir:  "",
		DURL:    "app.tela",
		Headers: tela.Headers{
			NameHdr:  "index.html",
			DescrHdr: "HTML index file",
			IconHdr:  "ICON_URL",
		},
		Signature: tela.Signature{
			CheckC: "c4d7bbdaaf9344f4c351e72d0b2145b4235402c89510101e0500f43969fd1387",
			CheckS: "b879b0ff01d78841d61e9770fd18436d8b9afce59302c77a786272e7422c15f6",
		},
	}

	// Pass doc to NewInstallArgs() to create smart contract and transfer arguments
	args, err := tela.NewInstallArgs(doc)
	if err != nil {
		// Handle error
	}
	fmt.Printf("SC Code: %s\n", args.Value(rpc.SCCODE, rpc.DataString))
	// Code to make transfer call with WS/RPC and install args

	// // //
	// //
	// Alternatively, Installer() takes a DOC or INDEX and installs it with the given walletapi
	ringsize := uint64(2)
	// ringsize 2 allows installed INDEX contracts to be updated,
	// ringsize > 2 will make the installed INDEX contract immutable
	txid, err := tela.Installer(&walletapi.Wallet_Disk{}, ringsize, doc)
	if err != nil {
		// Handle error
	}
	fmt.Printf("Installed TELA SCID: %s\n", txid)
}
```

#### Updating
Updating `TELA-INDEX-1`'s can be managed similarly to new installs. The values provided for updating the smart contract will be embedded into its new code making them available when the code is parsed post update, while the original variable stores for those values will remain unchanged preserving the contract's origin. The TXID generated by each update execution is stored in the smart contract, allowing for reference to the code changes that have taken place. For manual update procedures see [here](TELA-INDEX-1/README.md#update-tela-index-1).
```go
import (
	"fmt"

	"github.com/civilware/tela"
	"github.com/deroproject/derohe/walletapi"
)

func main() {
	// Create the new INDEX with all relevant data
	scid := "8dd839608e584f75b64c0ca7ff2c274879677ac3aaf60159c78797ee518946c2"

	index := &tela.INDEX{
		SCID: scid,
		DURL: "app.tela",
		DOCs: []string{"<scid>", "<scid>"},
		Headers: tela.Headers{
			NameHdr:  "TELA App",
			DescrHdr: "A TELA Application",
			IconHdr:  "ICON_URL",
		},
	}

	// Pass index to NewUpdateArgs() to create transfer arguments for update call
	args, err := tela.NewUpdateArgs(index)
	if err != nil {
		// Handle error
	}
	// Code to make transfer call with WS/RPC and update args

	// // //
	// //
	// Alternatively, Updater() takes a INDEX and updates it with the given walletapi
	txid, err := tela.Updater(&walletapi.Wallet_Disk{}, index)
	if err != nil {
		// Handle error
	}
	fmt.Printf("Update TXID: %s %s\n", txid, err)

	// // //
	// //
	// GetINDEXInfo() can be used to preform INDEX data from an existing SCID
	endpoint := "127.0.0.1:20000"
	liveIndex, _ := tela.GetINDEXInfo(scid, endpoint)
	liveIndex.DOCs = []string{"<scid>", "<scid>"}
	args, _ = tela.NewUpdateArgs(&liveIndex)
}
```

#### Rating
TELA content can be rated easily. The extended content rating system components are exported to make integrating the same TELA interpretations an easy process. For manual rating procedures see [here](TELA-INDEX-1/README.md#rate-tela-index-1).
```go
import (
	"fmt"

	"github.com/civilware/tela"
	"github.com/deroproject/derohe/walletapi"
)

func main() {
	// Define rating and scid
	rating := uint64(0)
	scid := "c4d7bbdaaf9344f4c351e72d0b2145b4235402c89510101e0500f43969fd1387"

	// Pass params to NewRateArgs() to create transfer arguments for Rate call
	args, err := tela.NewRateArgs(scid, rating)
	if err != nil {
		// Handle error
	}
	// Code to make transfer call with WS/RPC and rate args

	// // //
	// //
	// Alternatively, Rate() takes a TELA scid and rating and rates the contract with the given walletapi
	txid, err := tela.Rate(&walletapi.Wallet_Disk{}, scid, rating)
	if err != nil {
		// Handle error
	}
	fmt.Printf("Rate TXID: %s\n", txid)

	// // //
	// //
	// GetRating() gets the rating results from a scid
	endpoint := "127.0.0.1:20000"
	height := uint64(0) // Results can be filtered by height showing ratings that occurred >= height
	result, err := tela.GetRating(scid, endpoint, height)
	if err != nil {
		// Handle error
	}
	fmt.Printf("Likes: %d, Dislikes: %d Average: %d\n", result.Likes, result.Dislikes, result.Average)

	// // //
	// //
	// The package's rating structures can be accessed using the Rating variable
	category, detail, _ := tela.Ratings.Parse(rating)

	// Get a rating string formatted as "Category (detail)"
	ratingString, _ := tela.Ratings.ParseString(rating)

	// Get the category of a rating
	category = tela.Ratings.Category(rating)

	// Get the detail tag from a rating
	detail = tela.Ratings.Detail(rating, false)

	// Get all TELA rating categories
	categories := tela.Ratings.Categories()

	// Get all TELA negative details
	negativeDetails := tela.Ratings.NegativeDetails()

	// Get all TELA positive details
	positiveDetails := tela.Ratings.PositiveDetails()
}
```

#### Parse
The `civilware/tela` package has exported much of the functionality used in the necessary components to create TELA content stored on the DERO blockchain. The parsing and header tools can be of great use to developers working with any DVM smart contract.
```go
package main

import (
	"github.com/civilware/tela"
)

func main() {
	// Parse a file name for its respective TELA docType
	fileName := "index.html"
	docType := tela.ParseDocType(fileName)

	// Parse a TELA-INDEX-1 for any embedded DOCs
	scCode := tela.TELA_INDEX_1
	docSCID, _ := tela.ParseINDEXForDOCs(scCode)

	// Parse a DERO signature for its address, C and S values
	signature := []byte("-----BEGIN DERO SIGNED MESSAGE-----")
	address, c, s, _ := tela.ParseSignature(signature)

	// Get all function names from DERO smart contract code
	functionNames := tela.GetSmartContractFuncNames(scCode)

	// Compare equality between two DERO smart contracts
	dvmCode, _ := tela.EqualSmartContracts(scCode, scCode)

	// Format DERO smart contract code removing whitespace and comments
	formattedCode, _ := tela.FormatSmartContract(dvmCode, scCode)

	// Parse and inject headers into a DERO smart contract
	headers1 := &tela.Headers{
		NameHdr:  "myNameHdr",
		DescrHdr: "myDescrHdr",
		IconHdr:  "myIconURL",
	}
	formattedCode, _ = tela.ParseHeaders(scCode, headers1)

	// ParseHeaders takes various input formats for a wide range of use
	headers2 := &tela.INDEX{
		DURL:    "",
		DOCs:    []string{"<scid>"},
		Headers: *headers1,
	}
	formattedCode, _ = tela.ParseHeaders(scCode, headers2)

	// Including any custom headers
	headers3 := map[tela.Header]interface{}{
		tela.HEADER_NAME:           "myNameHdr",
		tela.HEADER_ICON_URL:       "myIconURL",
		tela.Header(`"customHdr"`): "myCustomHdr",
	}
	formattedCode, _ = tela.ParseHeaders(scCode, headers3)

	headers4 := map[string]interface{}{
		`"nameHdr"`:    "myNameHdr",
		`"iconURLHdr"`: "myIconURL",
		`"customHdr"`:  "myCustomHdr",
	}
	formattedCode, _ = tela.ParseHeaders(scCode, headers4)
}
```
### TELA-CLI
* [TELA-CLI](cmd/tela-cli/README.md)

### Changelog
* [Changelog](CHANGELOG.md)

### License
* [License](LICENSE)
