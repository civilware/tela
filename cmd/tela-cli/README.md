## TELA-CLI <!-- omit in toc -->

### Table of Contents <!-- omit in toc -->
- [What is TELA-CLI?](#what-is-tela-cli)
- [Get started](#get-started)
- [Flags](#flags)
- [Usage](#usage)
    - [Commands](#commands)
    - [Readline](#readline)
    - [Connect daemon](#connect-daemon)
    - [Connect wallet](#connect-wallet)
    - [Gnomon](#gnomon)
    - [Search TELA content](#search-tela-content)
    - [Serve TELA content](#serve-tela-content)
    - [Serve local directory](#serve-local-directory)
    - [Clone TELA content](#clone-tela-content)
    - [Shutdown servers](#shutdown-servers)
    - [Install TELA-DOC](#install-tela-doc)
    - [Install TELA-INDEX](#install-tela-index)
    - [Update TELA-INDEX](#update-tela-index)
    - [Rate TELA content](#rate-tela-content)
- [TELA](../../README.md)

### What is TELA-CLI?
TELA-CLI is an implementation of the `civilware/tela` [go package](https://pkg.go.dev/github.com/Civilware/tela) as a command line interface. With TELA-CLI, users can view and manage existing TELA content, along with creating and installing new TELA content.

### Get started
TELA-CLI can be built or installed from source in a number or ways. The following instructions assume that [go](https://go.dev/doc/install) is already installed on the system.

Run TELA-CLI from source. Leaves no build files after use.
```
git clone https://github.com/civilware/tela.git
cd tela/cmd/tela-cli
go mod tidy
go run .
```

Build TELA-CLI from source. Creates an executable to use.
```
git clone https://github.com/civilware/tela.git
cd tela/cmd/tela-cli
go mod tidy
go build .
./tela-cli
```

Install TELA-CLI from source. Installing allows TELA-CLI to be run from shell.
```
go install github.com/civilware/tela/cmd/tela-cli@latest
tela-cli
```

#### Flags
When starting TELA-CLI, a `--flag` may be provided to initialize various settings and values. If TELA-CLI has been previously connected on the device it will use the stored settings from the previous session as its defaults if no flags are provided.
```
Usage:
  tela-cli [options]
  tela-cli -h | --help

Options:
  -h --help     Show this screen
  --debug       Enable debug mode
  --mainnet     Set the network to mainnet
  --testnet     Set the network to testnet
  --simulator   Set the network to simulator

  --db-type=<gravdb>           Set DB type to use for preferences and encrypted storage, either gravdb or boltdb
  --wallet=<file.db>           Open a DERO wallet file
  --password=<password>        Use this password to open --wallet file
  --daemon=<127.0.0.1:10102>   Set and connect to daemon endpoint, if used with no endpoint arg it will connect using endpoint from stored preferences
  --gnomon                     Start Gnomon indexer, this must be used in in tandem with --daemon flag other wise Gnomon will not start
  --fastsync=<true>            Set Gnomon fastsync true/false to define loading smart contracts at chain height or continuing sync from last indexed height
  --num-parallel-blocks=<3>    Set the number of parallel blocks that Gnomon will index. (highly recommend to use local nodes if this is greater than 1-5)
```

### Usage
TELA-CLI can be used with or without a [DERO](https://dero.io) wallet. It follows a familiar file structure that aligns with other applications in the ecosystem, providing users with a sense of consistency and ease of navigation. Once started, use the `help` command to see the list of available commands. At any point `ctrl+d` can be used to go back to the main prompt. Use `exit`, `quit` or `ctrl+c` to close TELA-CLI. 

#### Commands 
```
exit, quit                   - Closes the app
help                         - Shows this menu
info                         - Application info and settings
list                         - List all running TELA servers

endpoint <127.0.0.1:10102>   - Set the daemon endpoint to fetch TELA content from
endpoint mainnet             - Set the network to mainnet and daemon endpoint to 127.0.0.1:10102
endpoint testnet             - Set the network to testnet and daemon endpoint to 127.0.0.1:40402
endpoint simulator           - Set the network to simulator and daemon endpoint to 127.0.0.1:20000
endpoint remote              - Set the network to mainnet and daemon endpoint to node.derofoundation.org:11012
endpoint close               - Close connection with current daemon endpoint

clone <scid>                 - Clone TELA content from SCID
clone <scid>@<txid>          - Clone TELA content from SCID at commit TXID if the TX data is available from the daemon

mv <source> <destination>    - Move a file or directory
rm <source>                  - Remove a file or directory, it will only remove from within the datashards/clone directory

file-info <source>           - Get source file information
file-shard <source>          - Take a source file and create DocShards files intended to be embedded in an INDEX and recreated as source when served
file-construct <source>      - Take a DocShard file and construct the original source file using the matching DocShards in the directory
file-diff <source> <compare> - View the line differences between a source and comparison file
scid-diff <scid1> <scid2>    - View the line differences between two smart contracts

serve <scid>                 - Serve TELA content from SCID
serve local <directory>      - Serve content from local directory, useful for testing TELA content pre install

shutdown <name>              - Shutdown a server by name
shutdown all                 - Shutdown all running servers
shutdown tela                - Shutdown all TELA servers
shutdown local               - Shutdown local directory server only

page-size <20>               - Set the max page size when displaying search results
port-start <8082>            - Set the port to start serving TELA servers from
max-servers <20>             - Set the maximum amount of TELA servers which can be active at once
updates <false>              - Set updates true/false to allow or deny updated TELA content when cloning or serving
browser <true>               - Set browser true/false to open content in default browser
colors <true>                - Set colors true/false to enable terminal colors

wallet <file.db>             - Open a DERO wallet file at path
wallet close                 - Close wallet file if active 
balance                      - View the connected wallet's balances

rate <scid>                  - Rate a TELA smart contract
install-doc <file.html>      - Start guided TELA-DOC smart contract install
install-index <name>         - Start guided TELA-INDEX smart contract install
update-index <scid>          - Start guided TELA-INDEX smart contract update

mods                         - Print all available TELA-MOD info
mods <class>                 - Print available info on a TELA-MODClass
mod-info <tag>               - Print info on a TELA-MOD by its tag

set-var <scid>               - Set a key/value store on a SCID
delete-var <scid>            - Delete a key/value store on a SCID, this is a owners only function
sc-execute <function>        - Execute a TELA-MOD smart contract function

gnomon start                 - Start Gnomon indexer
gnomon stop                  - Stop Gnomon indexer
gnomon resync                - Stop Gnomon indexer if running, delete Gnomon DB for current network and restart Gnomon indexer
gnomon add <scid>            - Add TELA SCID(s) to the local Gnomon DB
gnomon clean <network>       - Delete Gnomon indexer DB for network

search all                   - Search all TELA SCIDs in Gnomon DB
search key <key>             - Search all TELA SCIDs in Gnomon DB that contain a key store
search value <value>         - Search all TELA SCIDs in Gnomon DB that contain a value store
search scid <scid>           - Search by SCID
search scid vars <scid>      - Search all variables stored in a SCID
search docs                  - Search all TELA DOCs
search docs <docType>        - Search TELA DOCs by type
search indexes               - Search all TELA INDEXs
search libs                  - Search all dURLs tagged as libraries
search durl <dURL>           - Search by dURL
search code <scid>           - Search for SC code by SCID
search line <line>           - Search for a line of code in all SCs
search author <address>      - Search by author address
search min-likes <30>        - Sets the minimum likes % required to be a valid search result, 0 will not filter any content
search ratings <scid>        - Search ratings for a SCID, <height> can be added to filter results (min-likes will not apply to rating search results)
search my docs               - Search all DOCs installed for connected wallet (min-likes will not apply to any of the my search results)
search my docs <docType>     - Search all Docs by type for connected wallet
search my indexes            - Search all INDEXs for connected wallet
search exclude view          - View all current search exclusions
search exclude clear         - Clear all current search exclusions
search exclude add <text>    - Add a filter to exclude SCIDs containing <text> in their dURL
search exclude remove <text> - Remove a specific search exclusion
```

#### Readline
The realine displays the current status of TELA-CLI. Further status info can be viewed using the `info` command.
- `D:` displays the daemon status, either online (▲) or offline (▼).
- `G:` displays Gnomon's status, either online (▲) or offline (▼).
- `W:` displays the wallet height if connected.
- `0/21` displays the current <i>active/max</i> servers.
```
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▼] [G:▼] [W:0] [0/21] »
```

#### Connect daemon
To connect to a custom daemon, use `endpoint <127.0.0.1:10102>`. Alternatively, you can use `endpoint simulator` or `endpoint remote` commands to switch to the corresponding network and connect to its default daemon address. The hardcoded default addresses for each network can be viewed using the `help` command.
```
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▼] [G:▼] [W:0] [0/21] » endpoint simulator 
[01/02/2006 15:04:05]  INFO  TELA-CLI: Endpoint set to: 127.0.0.1:20000
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▲] [G:▼] [W:0] [0/21] » 
```

#### Connect wallet
Use the `wallet` command to enter guided wallet connection, or use `wallet <path/to/file.db>` to open a wallet file from path.
```
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▲] [G:▼] [W:0] [0/21] » wallet 
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▲] [G:▼] [W:0] [0/21] Enter wallet file path » testnet_simulator/wallet.db
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▲] [G:▼] [W:0] [0/21] Enter wallet.db password (7) » 
[01/02/2006 15:04:05]  INFO  TELA-CLI: Wallet connected: testnet_simulator/1.db
[01/02/2006 15:04:05]  INFO  TELA-CLI: Address: deto1qyre7td6x9r88y4cavdgpv6k7lvx6j39lfsx420hpvh3ydpcrtxrxqg8v8e3z
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▲] [G:▼] [W:1337] [0/21] »  
```

#### Gnomon
[Gnomon](https://github.com/civilware/Gnomon) is the decentralized indexer used for querying and storing TELA smart contract data in a local DB. Gnomon indexer can be started using `gnomon start` and stopped with `gnomon stop`.
```
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▲] [G:▼] [W:1337] [0/21] » gnomon start
[01/02/2006 15:04:05]  INFO  GetLastIndexHeight: No stored last index height. Starting from 0 or latest if fastsync is enabled
[01/02/2006 15:04:05]  INFO  Gnomon: Scan Status: [0 / 1337]
[01/02/2006 15:04:05]  INFO  StartDaemonMode: Trying to connect...
[01/02/2006 15:04:05]  INFO  StartDaemonMode: Fastsync Configuration: &{true false true 100 false}
[01/02/2006 15:04:05]  INFO  GetLastIndexHeight: No stored last index height. Starting from 0 or latest if fastsync is enabled
[01/02/2006 15:04:05]  INFO  StartDaemonMode: Fastsync initiated, setting to chainheight (1337)
[01/02/2006 15:04:05]  INFO  StartDaemonMode-storedIndex: Continuing from last indexed height 1337
[01/02/2006 15:04:05]  INFO  StartDaemonMode: Set number of parallel blocks to index to '3'. Starting index routine...
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▲] [G:▲] [W:1337] [0/21] »  
```

#### Search TELA content
Once Gnomon is online, installed TELA content can be searched with a variety of queries like `search docs` to find all TELA documents or `search indexes` to find all TELA indexes. Search results can be filtered by ratings using `search min-likes <50>` to omit any results below that percentage.
```
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▲] [G:▲] [W:0] [0/21] » search indexes 
[01/02/2006 15:04:05]  INFO  TELA-CLI: Showing 2 of 8 results
------------
dURL: app.tela                                                          Author: deto1qyre7td6x9r88y4cavdgpv6k7lvx6j39lfsx420hpvh3ydpcrtxrxqg8v8e3z
SCID: f0dfcb506fe313bfdd1b5f6ceaeaa01ef2725c81848418f6a8590743a14920f0  Type: TELA-INDEX-1      Name: TELA Application                   Likes: 50%
------------
dURL: dero.lib                                                          Author: deto1qy87ghfeeh6n6tdxtgh7yuvtp6wh2uwfzvz7vjq0krjy4smmx62j5qgqht7t3
SCID: adb26e75d411c8d27f3f2e1ee06bc0dd2d41a8087bf316ef3f446214ca0656ef  Type: TELA-INDEX-1      Name: A DERO library                     Likes: 80%
------------
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▲] [G:▲] [W:0] [0/21] Show more results? (6) (y/n) »   
```

#### Serve TELA content
Use `serve <scid>` to serve the TELA content from that SCID. TELA-CLI's default setting is to open served content in the devices default browser. Content that has been updated since its deployment will be restricted, running command `updates true` will allow updated content to be served.
```
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▲] [G:▼] [W:0] [0/21] » serve f0dfcb506fe313bfdd1b5f6ceaeaa01ef2725c81848418f6a8590743a14920f0
[01/02/2006 15:04:05]  INFO  TELA: Creating main.js
[01/02/2006 15:04:05]  INFO  TELA: Creating style.css
[01/02/2006 15:04:05]  INFO  TELA: Creating index.html
[01/02/2006 15:04:05]  INFO  TELA: Serving app.tela at http://localhost:8082/index.html
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▲] [G:▼] [W:0] [1/21] »  
```

#### Serve local directory
Use `serve local <path/to/directory>` to serve content from local files.
```
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▲] [G:▼] [W:0] [1/21] » serve local ../../tela_tests/app1 
[01/02/2006 15:04:05]  INFO  TELA-CLI: Serving ../../tela_tests/app1 at http://localhost:8083/index.html
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▲] [G:▼] [W:0] [2/21] »  
```
#### Clone TELA content
TELA content can be cloned by SCID using `clone <scid>`. Cloned files are stored separately from ones served and will persist after TELA-CLI is closed. Specifying clone with a TXID `clone <scid>@<txid>` will clone TELA-INDEX's at that commit TX, if the endpoint has the TX data.  
```
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▲] [G:▼] [W:0] [0/21] » clone f0dfcb506fe313bfdd1b5f6ceaeaa01ef2725c81848418f6a8590743a14920f0
[01/02/2006 15:04:05]  INFO  TELA: Creating index.html
[01/02/2006 15:04:05]  INFO  TELA: Creating main.js
[01/02/2006 15:04:05]  INFO  TELA: Creating style.css
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▲] [G:▼] [W:0] [0/21] »  
```

#### Shutdown servers
Use `shutdown <name>` to shutdown a server by name. Use `shutdown all` to shutdown all running servers. Using `shutdown local` will shutdown only the local directory server. When TELA-CLI closes, all servers will be shutdown.
```
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▲] [G:▼] [W:1337] [2/21] » shutdown all
[01/02/2006 15:04:05]  INFO  TELA: Shutdown
[01/02/2006 15:04:05]  INFO  TELA: Closed :8082 test1.tela
[01/02/2006 15:04:05]  INFO  TELA-CLI: Closed local :8083 ../../tela_tests/app1
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▲] [G:▼] [W:1337] [0/21] »  
```

#### Install TELA-DOC
Use `install-doc` to enter guided install. Using `install-doc <file.html>` will start guided install from a file path. It is recommended to have Gnomon running when installing so that the installed contract will be immediately added to your local DB.
```
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▲] [G:▲] [W:1337] [0/21] » install-doc README.md 
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▲] [G:▲] [W:1337] [0/21] Confirm password (7) » 
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▲] [G:▲] [W:1337] [0/21] Enter DOC description » Doc description (can be empty)
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▲] [G:▲] [W:1337] [0/21] Enter DOC icon » https://iconurl.com (can be empty)
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▲] [G:▲] [W:1337] [0/21] Enter DOC dURL » readme.tela
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▲] [G:▲] [W:1337] [0/21] Enter DOC subDir » 
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▲] [G:▲] [W:1337] [0/21] Enter DOC install ringsize » 2
[01/02/2006 15:04:05]  INFO  TELA-CLI: DOC signature headers:
-----BEGIN DERO SIGNED MESSAGE-----
Address: deto1qyre7td6x9r88y4cavdgpv6k7lvx6j39lfsx420hpvh3ydpcrtxrxqg8v8e3z
C: 29e2863d6b7bcb340a4ea4e8740dcda7bbc887c73347d8fbad7773bc395fd770
S: 18c12ed2e8ca950d5cc14c2725751230dea855c20b28cd2c7869b910b70e0578
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▲] [G:▲] [W:1337] [0/21] Confirm DOC install (y/n) » y
[01/02/2006 15:04:05]  INFO  TELA-CLI: DOC install TXID: 8232dd8e909bdc095ab213a035a70a94b64b2e4763a959f4fe3065f8f9fbc2df
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▲] [G:▲] [W:1337] [0/21] » 
```

#### Install TELA-INDEX
Use `install-index` to enter guided install. It is recommended to have Gnomon running when installing so that the installed contract will be immediately added to your local DB. `TELA-MODs` can be added during the installation process to enable custom smart contract functionalities. More information about `TELA-MODs` can be found [here](../../TELA-MOD-1/README.md). Installing an INDEX with ringsize above 2 will make it immutable.
```
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▲] [G:▲] [W:1337] [0/21] » install-index
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▲] [G:▲] [W:1337] [0/21] Confirm password (7) » 
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▲] [G:▲] [W:1337] [0/21] Enter INDEX name » myApp 
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▲] [G:▲] [W:1337] [0/21] Enter INDEX description » Index description (can be empty)
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▲] [G:▲] [W:1337] [0/21] Enter INDEX icon » https://iconurl.com (can be empty)
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▲] [G:▲] [W:1337] [0/21] Enter INDEX dURL » app.tela
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▲] [G:▲] [W:1337] [0/21] How many total documents are embedded in this INDEX? » 1
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▲] [G:▲] [W:1337] [0/21] Enter DOC1 SCID » 8232dd8e909bdc095ab213a035a70a94b64b2e4763a959f4fe3065f8f9fbc2df
[01/02/2006 15:04:05]  INFO  TELA-CLI: File: index.html
[01/02/2006 15:04:05]  INFO  TELA-CLI: Author: deto1qyre7td6x9r88y4cavdgpv6k7lvx6j39lfsx420hpvh3ydpcrtxrxqg8v8e3z
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▲] [G:▲] [W:1337] [0/21] Add SC TELA-MODs (y/n) » n
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▲] [G:▲] [W:1337] [0/21] Enter INDEX install ringsize » 2
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▲] [G:▲] [W:1337] [0/21] Confirm INDEX install (y/n) » y
[01/02/2006 15:04:05]  INFO  TELA-CLI: INDEX install TXID: daab713d2c0ee0d0efacb5990dcc5df227b847ab3f064fe28ae8e3d946f903cb
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▲] [G:▲] [W:1337] [0/21] »  
```

#### Update TELA-INDEX
Owners can update their INDEX smart contracts by using `update-index <scid>` to enter a guided update for the given SCID. If an INDEX's version is behind the latest, updating will pull in any required smart contract changes to update it to the latest version, matching the TELA standard. When updating a smart contract, existing `TELA-MODs` can be removed or new `TELA-MODs` can be added to introduce custom smart contract functionalities. More information about `TELA-MODs` can be found [here](../../TELA-MOD-1/README.md). 
```
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▲] [G:▲] [W:1337] [0/21] » update-index b0f24cac2045e2b21eb2885f4da1f2771cfdf88863f5f602b47ea897ae40fa83
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▲] [G:▲] [W:1337] [0/21] Confirm password (7) » 
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▲] [G:▲] [W:1337] [0/21] Enter INDEX name » myApp
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▲] [G:▲] [W:1337] [0/21] Enter INDEX description » Index description (can be empty)
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▲] [G:▲] [W:1337] [0/21] Enter INDEX icon » https://iconurl.com (can be empty)
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▲] [G:▲] [W:1337] [0/21] Enter INDEX dURL » app.tela
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▲] [G:▲] [W:1337] [0/21] How many total documents are embedded in this INDEX? » 2
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▲] [G:▲] [W:1337] [0/21] Enter DOC1 SCID » 8232dd8e909bdc095ab213a035a70a94b64b2e4763a959f4fe3065f8f9fbc2df
[01/02/2006 15:04:05]  INFO  TELA-CLI: File: index.html
[01/02/2006 15:04:05]  INFO  TELA-CLI: Author: deto1qyre7td6x9r88y4cavdgpv6k7lvx6j39lfsx420hpvh3ydpcrtxrxqg8v8e3z
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▲] [G:▲] [W:1337] [0/21] Enter DOC1 SCID » a82e92b0d1210f7b161e8c9f5a1aa1a7277b7852450d488c47059724a62cf0dc
[01/02/2006 15:04:05]  INFO  TELA-CLI: File: main.js
[01/02/2006 15:04:05]  INFO  TELA-CLI: Author: deto1qyre7td6x9r88y4cavdgpv6k7lvx6j39lfsx420hpvh3ydpcrtxrxqg8v8e3z
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▲] [G:▲] [W:1337] [0/21] Add SC TELA-MODs (y/n) » n
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▲] [G:▲] [W:1337] [0/21] Confirm INDEX update (y/n) » y
[01/02/2006 15:04:05]  INFO  TELA-CLI: INDEX update TXID: 2694052c159b3d31bfaf9d562a6bc2f283429c1d334a03bd44765c893edbfbc9
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▲] [G:▲] [W:1337] [0/21] »  
```

#### Rate TELA content
All TELA content can be rated by users. TELA-CLI follows the `civilware/tela` go packages [content rating system](../../README.md#content-rating-system) which is broken down into category and detail.
```
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▲] [G:▼] [W:1337] [0/21] » rate daab713d2c0ee0d0efacb5990dcc5df227b847ab3f064fe28ae8e3d946f903cb
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▲] [G:▼] [W:1337] [0/21] Confirm password (7) » 
0: Do not use
1: Broken
2: Major issues
3: Minor issues
4: Should be improved
5: Could be improved
6: Average
7: Good
8: Very good
9: Exceptional
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▲] [G:▼] [W:1337] [0/21] Enter rating number » 6
0: Nothing
1: Needs review
2: Needs improvement
3: Bugs
4: Errors
5: Visually appealing
6: In depth
7: Works well
8: Unique
9: Benevolent
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▲] [G:▼] [W:1337] [0/21] Enter detail number » 0
[01/02/2006 15:04:05]  INFO  TELA-CLI: Rating is: 60  Average
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▲] [G:▼] [W:1337] [0/21] Confirm rating (y/n) » y
[01/02/2006 15:04:05]  INFO  TELA-CLI: Rate TXID: 54c5c6b44ebb494160baa0f1d9071cea4caa342c109d70857ccfa55272589798
[01/02/2006 15:04:05]  ⠞⠑⠇⠁ TELA-CLI: [D:▲] [G:▼] [W:1337] [0/21] »  
```

### TELA
More information on TELA Decentralized Web Standard and its components can be found in the `civilware/tela` [package repo](https://github.com/civilware/tela).