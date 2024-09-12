package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/chzyer/readline"
	"github.com/civilware/tela"
	"github.com/civilware/tela/logger"
	"github.com/civilware/tela/shards"
	"github.com/deroproject/derohe/globals"
	"github.com/deroproject/derohe/walletapi"
)

type tela_cli struct {
	sync.WaitGroup
	sync.RWMutex
	rli           *readline.Instance
	cancel        context.CancelFunc
	ctx           context.Context
	shutdown      func()
	endpoint      string
	os            string
	exclusions    []string
	pageSize      int
	minLikes      float64
	openInBrowser bool
	wait          bool
	wallet        struct {
		disk *walletapi.Wallet_Disk
		name string
	}
	local struct {
		server *http.Server
		tela.ServerInfo
	}
}

const searchDivider = "------------"
const appName = "TELA-CLI"

// Start arguments
const argsLine = `TELA DWEB CLI: Interact with with Tela content

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
  --num-parallel-blocks=<3>    Set the number of parallel blocks that Gnomon will index. (highly recommend to use local nodes if this is greater than 1-5)`

// Running commands
const commandLine string = `----------------
List of commands:

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
----------------`

func main() {
	var app tela_cli

	completer := readline.NewPrefixCompleter(
		readline.PcItem("exit"),
		readline.PcItem("quit"),
		readline.PcItem("help"),
		readline.PcItem("info"),
		readline.PcItem("endpoint",
			completerNetworks(true)...,
		),
		readline.PcItem("clone",
			readline.PcItem("github.com/"),
		),
		readline.PcItem("mv",
			completerFiles(".", completerFiles(".")),
		),
		readline.PcItem("rm",
			completerFiles(filepath.Join(filepath.Base(shards.GetPath()), "clone")),
		),
		readline.PcItem("file-info",
			completerFiles("."),
		),
		readline.PcItem("file-shard",
			completerFiles("."),
		),
		readline.PcItem("file-construct",
			completerFiles("."),
		),
		readline.PcItem("file-diff",
			completerFiles(".", completerFiles(".")),
		),
		readline.PcItem("scid-diff"),
		readline.PcItem("serve",
			readline.PcItem("local",
				completerFiles("."),
				readline.PcItem(filepath.Join("..", "..", "tela_tests", "app1")),
			),
		),
		readline.PcItem("list"),
		readline.PcItem("shutdown",
			readline.PcItemDynamic(app.completerServers()),
			readline.PcItem("all"),
			readline.PcItem("tela"),
			readline.PcItem("local"),
		),
		readline.PcItem("page-size"),
		readline.PcItem("port-start"),
		readline.PcItem("max-servers"),
		readline.PcItem("updates",
			completerTrueFalse()...,
		),
		readline.PcItem("browser",
			readline.PcItem("true"),
			readline.PcItem("false"),
		),
		readline.PcItem("colors",
			completerTrueFalse()...,
		),
		readline.PcItem("wallet",
			readline.PcItem("close"),
			completerFiles("."),
		),
		readline.PcItem("balance"),
		readline.PcItem("rate"),
		readline.PcItem("install-doc",
			completerFiles("."),
		),
		readline.PcItem("install-index"),
		readline.PcItem("update-index"),
		readline.PcItem("mods",
			completerMODClasses()...,
		),
		readline.PcItem("mod-info",
			completerMODs("")...,
		),
		readline.PcItem("set-var"),
		readline.PcItem("delete-var"),
		readline.PcItem("sc-execute"),
		readline.PcItem("gnomon",
			readline.PcItem("start"),
			readline.PcItem("stop"),
			readline.PcItem("resync"),
			readline.PcItem("add"),
			readline.PcItem("clean",
				completerNetworks(false)...,
			),
		),
		readline.PcItem("search",
			readline.PcItem("all"),
			readline.PcItem("key"),
			readline.PcItem("value"),
			readline.PcItem("scid",
				readline.PcItem("vars"),
			),
			readline.PcItem("docs",
				completerDocType()...,
			),
			readline.PcItem("indexes"),
			readline.PcItem("libs"),
			readline.PcItem("durl"),
			readline.PcItem("code"),
			readline.PcItem("line"),
			readline.PcItem("author"),
			readline.PcItem("min-likes"),
			readline.PcItem("ratings"),
			readline.PcItem("my",
				readline.PcItem("docs",
					completerDocType()...,
				),
				readline.PcItem("indexes"),
			),
			readline.PcItem("exclude",
				readline.PcItem("view"),
				readline.PcItem("clear"),
				readline.PcItem("add"),
				readline.PcItem("remove"),
			),
		),
	)

	// Readline config
	rlConfig := &readline.Config{
		EOFPrompt:           "back",
		InterruptPrompt:     "^C",
		HistorySearchFold:   true,
		AutoComplete:        completer,
		FuncFilterInputRune: filterInput,
	}

	// Initialize readline for TELA-CLI app
	var err error
	app.rli, err = readline.NewEx(rlConfig)
	if err != nil {
		logger.Fatalf("[%s] Readline: %s\n", appName, err)
	}
	defer app.rli.Close()

	// Default app vars
	app.endpoint = "127.0.0.1:10102"
	app.os = runtime.GOOS
	app.openInBrowser = true
	app.pageSize = 20
	app.minLikes = 30

	app.ctx, app.cancel = context.WithCancel(context.Background())
	defer app.cancel()

	// App shutdown func, closes all servers, wallet and gnomon
	app.shutdown = func() {
		logger.Printf("[%s] Closing...\n", appName)
		stopGnomon()
		tela.ShutdownTELA()
		app.shutdownLocalServer()
		app.closeWallet()
		app.Wait()
		logger.Printf("[%s] Closed\n", appName)
		os.Exit(0)
	}

	// Gnomon defaults
	gnomon.fastsync = true
	gnomon.parallelBlocks = 3

	logger.Printf("[%s] Run %q for list of commands\n", appName, "help")

	// Parse start flags and set stored preferences
	flags := app.parseFlags()

	// Initialize DERO network config
	globals.InitNetwork()

	// Set daemon endpoint if provided and connect
	if endpoint, ok := flags["--daemon"]; ok {
		if endpoint != "" { // no arg will connect to existing stored endpoint without overwriting it
			app.endpoint = endpoint
			logger.Printf("[%s] %s=%s\n", appName, "--daemon", app.endpoint)
		}

		err = app.connectEndpoint()
		if err != nil {
			logger.Errorf("[%s] Connect: %s\n", appName, networkError(err))
		} else {
			logger.Printf("[%s] Connected: %s\n", appName, app.endpoint)
			// Daemon is connected, start Gnomon if requested
			if _, ok := flags["--gnomon"]; ok {
				startGnomon(app.endpoint)
				// Wait here until Gnomon is synced, force fast sync will trigger when LastIndexedHeight < (Chain height - ForceFastSyncDiff)
				for gnomon.Indexer != nil {
					time.Sleep(time.Second)
					s := gnomon.Indexer.Status
					if s == "indexed" || s == "closing" {
						break
					}
				}
			}
		}
	}

	// Open a wallet file, prompting for password if not provided
	if file, ok := flags["--wallet"]; ok {
		if file != "" {
			password := ""
			if p, ok := flags["--password"]; ok {
				password = p
			}

			err := app.openWallet(file, password)
			if err != nil {
				if readError(err) {
					return
				}

				logger.Errorf("[%s] Opening wallet: %s\n", appName, networkError(err))
			}
		}
	}

	app.Add(1)

	// Refresh readline
	go func() {
		ticker := time.NewTicker(time.Second * 5)
		defer ticker.Stop()
		defer app.Done()
		for {
			select {
			case <-app.ctx.Done():
				return
			case <-time.After(time.Second):
				// Refresh the prompt
			case <-ticker.C:
				checkDaemonConnection()
			}

			if !app.wait {
				app.Lock()
				app.setPrompt("")
				app.Unlock()
			}
		}
	}()

	// Process inputs
	for {
		app.setPrompt("")
		app.wait = false
		line, err := app.rli.Readline()
		if err != nil {
			if readError(err) {
				app.cancel()
				app.shutdown()
				return
			} else {
				logger.Errorf("[%s] %s\n", appName, err)
				continue
			}
		}

		app.wait = true
		split := strings.Split(strings.TrimSpace(line), " ")

		// Parse input arguments
		var args []string
		for i, str := range split {
			if i == 0 {
				continue
			}

			if s := strings.TrimSpace(str); s != "" {
				args = append(args, s)
			}
		}

		// Commands
		switch strings.ToLower(split[0]) {
		case "exit", "quit":
			app.cancel()
			app.shutdown()
			return
		case "help":
			fmt.Println(commandLine)
		case "info":
			logger.ASCIIBlend(logger.ASCIISmall, app.getCLIInfo())
		case "endpoint":
			if args == nil {
				line, err := app.readLine("Set daemon endpoint", "")
				if err != nil {
					if readError(err) {
						return
					}
					continue
				}

				args = []string{line}
			}

			walletapi.Connected = false

			// Get the current network to see if it will be switched
			gnomonStopped := false
			currentNetwork := strings.ToLower(getNetworkInfo())

			// Default network addresses
			switch args[0] {
			case "mainnet":
				if currentNetwork != args[0] {
					gnomonStopped = stopGnomon()
				}
				app.endpoint = "127.0.0.1:10102"
				globals.Arguments["--testnet"] = false
				globals.Arguments["--simulator"] = false
			case "testnet":
				if currentNetwork != args[0] {
					gnomonStopped = stopGnomon()
				}
				app.endpoint = "127.0.0.1:40402"
				globals.Arguments["--testnet"] = true
				globals.Arguments["--simulator"] = false
			case "simulator":
				if currentNetwork != args[0] {
					gnomonStopped = stopGnomon()
				}
				app.endpoint = "127.0.0.1:20000"
				globals.Arguments["--testnet"] = true
				globals.Arguments["--simulator"] = true
			case "remote":
				if currentNetwork != "mainnet" {
					gnomonStopped = stopGnomon()
				}
				app.endpoint = "node.derofoundation.org:11012"
				globals.Arguments["--testnet"] = false
				globals.Arguments["--simulator"] = false
			case "close":
				stopGnomon()
				globals.Arguments["--testnet"] = false
				globals.Arguments["--simulator"] = false
				globals.InitNetwork()
				walletapi.Connect(" ")
				continue
			default:
				_, err := net.ResolveTCPAddr("tcp", args[0])
				if err != nil {
					logger.Errorf("[%s] Endpoint: %s\n", appName, err)
					continue
				}

				app.endpoint = args[0]
			}

			err = app.connectEndpoint()
			if err != nil {
				walletapi.Connect(" ")
				walletapi.Daemon_Endpoint_Active = ""
				logger.Errorf("[%s] Endpoint %s: %s\n", appName, app.endpoint, networkError(err))
				stopGnomon()

				continue
			}

			logger.Printf("[%s] Endpoint set to: %s\n", appName, app.endpoint)
			if gnomonStopped {
				startGnomon(app.endpoint)
				if gnomon.Indexer != nil {
					time.Sleep(time.Second * 5)
				}
			}
		case "clone":
			// Git clone
			if args != nil && strings.Contains(args[0], "/") {
				err := gitClone(args[0])
				if err != nil {
					logger.Errorf("[%s] Git clone: %s\n", appName, err)
				}

				continue
			}

			// TELA clone
			if !walletapi.IsDaemonOnline() {
				logger.Errorf("[%s] Daemon %s not online to clone\n", appName, app.endpoint)
				continue
			}

			if args == nil {
				line, err := app.readLine("Enter SCID", "")
				if err != nil {
					if readError(err) {
						return
					}
					continue
				}

				args = []string{line}
			}

			if len(args[0]) == 129 {
				// Clone content at commit
				split := strings.Split(args[0], "@")
				if len(split) < 2 {
					logger.Errorf("[%s] Cloning at commit requires scid@txid format\n", appName)
					continue
				}

				err := tela.CloneAtCommit(split[0], split[1], app.endpoint)
				if err != nil {
					logger.Errorf("[%s] Clone: %s\n", appName, err)
					continue
				}
			} else {
				if len(args[0]) != 64 {
					logger.Errorf("[%s] Invalid SCID: %q\n", appName, args[0])
					continue
				}

				// Standard clone at height
				err := tela.Clone(args[0], app.endpoint)
				if err != nil {
					logger.Errorf("[%s] Clone: %s\n", appName, err)
					continue
				}
			}
		case "mv":
			if args == nil {
				completer := readline.NewPrefixCompleter(completerFiles("."))
				line, err := app.readLineWithCompleter("Enter source", "", completer)
				if err != nil {
					if readError(err) {
						return
					}
					continue
				}

				args = []string{line}
			}

			args[0] = filepath.Clean(args[0])
			if _, err = os.Stat(args[0]); os.IsNotExist(err) {
				logger.Errorf("[%s] %q does not exists\n", appName, args[0])
				continue
			}

			if len(args) < 2 {
				completer := readline.NewPrefixCompleter(completerFiles("."))
				line, err := app.readLineWithCompleter("Enter destination", "", completer)
				if err != nil {
					if readError(err) {
						return
					}
					continue
				}

				args = append(args, line)
			}

			args[1] = filepath.Clean(args[1])
			if _, err = os.Stat(args[1]); !os.IsNotExist(err) {
				logger.Errorf("[%s] %q already exists\n", appName, args[1])
				continue
			}

			yes, err := app.readYesNo(fmt.Sprintf("Move %s%s%s to %s%s%s", logger.Color.Green(), args[0], logger.Color.End(), logger.Color.Red(), args[1], logger.Color.End()))
			if err != nil {
				if readError(err) {
					return
				}
				continue
			}

			if !yes {
				continue
			}

			err = os.MkdirAll(filepath.Dir(args[1]), os.ModePerm)
			if err != nil {
				logger.Errorf("[%s] Move: %s\n", appName, err)
				continue
			}

			err = os.Rename(args[0], args[1])
			if err != nil {
				logger.Errorf("[%s] Move: %s\n", appName, err)
			}
		case "rm":
			if args == nil {
				logger.Errorf("[%s] Remove requires a path\n", appName)
				continue
			}

			datashards := shards.GetPath()
			cloneDir := filepath.Join(filepath.Base(datashards), "clone")

			args[0] = filepath.Clean(args[0])
			if !strings.Contains(args[0], cloneDir) || args[0] == cloneDir {
				logger.Warnf("[%s] Remove will only target files within %s\n", appName, cloneDir)
				continue
			}

			if _, err = os.Stat(args[0]); os.IsNotExist(err) {
				logger.Errorf("[%s] %q does not exists\n", appName, args[0])
				continue
			}

			yes, err := app.readYesNo(fmt.Sprintf("Remove %s%s%s", logger.Color.Red(), args[0], logger.Color.End()))
			if err != nil {
				if readError(err) {
					return
				}
				continue
			}

			if !yes {
				continue
			}

			os.RemoveAll(args[0])
			logger.Printf("[%s] %s deleted\n", appName, args[0])
		case "file-info":
			if args == nil {
				completer := readline.NewPrefixCompleter(completerFiles("."))
				line, err := app.readLineWithCompleter("Enter source", "", completer)
				if err != nil {
					if readError(err) {
						return
					}
					continue
				}

				args = []string{line}
			}

			args[0] = filepath.Clean(args[0])

			var fileInfo fs.FileInfo
			if fileInfo, err = os.Stat(args[0]); err != nil {
				if os.IsNotExist(err) {
					logger.Errorf("[%s] %q does not exists\n", appName, args[0])
				} else {
					logger.Errorf("[%s] Could not access file: %s\n", appName, err)
				}

				continue
			}

			printFileInfo(fileInfo)
		case "file-shard":
			if args == nil {
				completer := readline.NewPrefixCompleter(completerFiles("."))
				line, err := app.readLineWithCompleter("Enter source", "", completer)
				if err != nil {
					if readError(err) {
						return
					}
					continue
				}

				args = []string{line}
			}

			args[0] = filepath.Clean(args[0])

			var fileInfo fs.FileInfo
			if fileInfo, err = os.Stat(args[0]); err != nil {
				if os.IsNotExist(err) {
					logger.Errorf("[%s] %q does not exists\n", appName, args[0])
				} else {
					logger.Errorf("[%s] Could not access file: %s\n", appName, err)
				}

				continue
			}

			if fileInfo.IsDir() {
				logger.Errorf("[%s] %q is a directory\n", appName, args[0])
				continue
			}

			totalChunks := (fileInfo.Size() + CHUNK_SIZE - 1) / CHUNK_SIZE
			if totalChunks < 2 {
				logger.Errorf("[%s] %q is smaller than chunk size\n", appName, args[0])
				continue
			}

			yes, err := app.readYesNo(fmt.Sprintf("Shard %s%s%s (%d)", logger.Color.Green(), args[0], logger.Color.End(), totalChunks))
			if err != nil {
				if readError(err) {
					return
				}
				continue
			}

			if !yes {
				continue
			}

			err = tela.CreateShardFiles(args[0])
			if err != nil {
				logger.Errorf("[%s] Shard: %s\n", appName, err)
				continue
			}

			logger.Printf("[%s] Shards created successfully\n", appName)
		case "file-construct":
			if args == nil {
				completer := readline.NewPrefixCompleter(completerFiles("."))
				line, err := app.readLineWithCompleter("Enter source", "", completer)
				if err != nil {
					if readError(err) {
						return
					}
					continue
				}

				args = []string{line}
			}

			args[0] = filepath.Clean(args[0])
			if _, err = os.Stat(args[0]); os.IsNotExist(err) {
				logger.Errorf("[%s] %q does not exists\n", appName, args[0])
				continue
			}

			docShards, recreate, err := findDocShardFiles(args[0])
			if err != nil {
				logger.Errorf("[%s] Find shards: %s\n", appName, err)
				continue
			}

			if len(docShards) < 2 {
				logger.Errorf("[%s] Not enough shards to construct\n", appName)
				continue
			}

			yes, err := app.readYesNo(fmt.Sprintf("Construct %s%s%s (%d)", logger.Color.Green(), filepath.Join(filepath.Dir(args[0]), recreate), logger.Color.End(), len(docShards)))
			if err != nil {
				if readError(err) {
					return
				}
				continue
			}

			if !yes {
				continue
			}

			err = tela.ConstructFromShards(docShards, recreate, filepath.Dir(args[0]))
			if err != nil {
				logger.Errorf("[%s] Construct: %s\n", appName, err)
			}
		case "file-diff":
			if args == nil {
				completer := readline.NewPrefixCompleter(completerFiles("."))
				line, err := app.readLineWithCompleter("Enter source", "", completer)
				if err != nil {
					if readError(err) {
						return
					}
					continue
				}

				args = []string{line}
			}

			args[0] = filepath.Clean(args[0])
			if _, err = os.Stat(args[0]); os.IsNotExist(err) {
				logger.Errorf("[%s] %q does not exists\n", appName, args[0])
				continue
			}

			if len(args) < 2 {
				completer := readline.NewPrefixCompleter(completerFiles("."))
				line, err := app.readLineWithCompleter("Enter comparison", "", completer)
				if err != nil {
					if readError(err) {
						return
					}
					continue
				}

				args = append(args, line)
			}

			args[1] = filepath.Clean(args[1])
			if _, err = os.Stat(args[1]); os.IsNotExist(err) {
				logger.Errorf("[%s] %q does not exists\n", appName, args[1])
				continue
			}

			diff, fileNames, err := getFileDiff(args[0], args[1])
			if err != nil {
				logger.Errorf("[%s] Get Diff: %s\n", appName, err)
				continue
			}

			err = app.printDiff(diff, fileNames)
			if err != nil {
				logger.Errorf("[%s] File Diff: %s\n", appName, err)
				continue
			}
		case "scid-diff":
			if gnomon.Indexer == nil {
				logger.Printf("[%s] Gnomon is not running\n", appName)
				continue
			}

			if args == nil {
				line, err := app.readLine("Enter SCID1", "")
				if err != nil {
					if readError(err) {
						return
					}
					continue
				}

				args = []string{line}
			}

			if len(args[0]) != 64 {
				logger.Errorf("[%s] Invalid SCID: %q\n", appName, args[0])
				continue
			}

			vars := gnomon.GetAllSCIDVariableDetails(args[0])
			if vars == nil {
				logger.Printf("[%s] SCID not found\n", appName)
				continue
			}

			var diff []string
			for _, h := range vars {
				switch k := h.Key.(type) {
				case string:
					if k == "C" {
						code, ok := h.Value.(string)
						if ok {
							diff = append(diff, code)
						}

						break
					}
				}
			}

			if len(args) < 2 {
				line, err := app.readLine("Enter SCID2", "")
				if err != nil {
					if readError(err) {
						return
					}
					continue
				}

				args = append(args, line)
			}

			if len(args[1]) != 64 {
				logger.Errorf("[%s] Invalid SCID: %q\n", appName, args[1])
				continue
			}

			vars = gnomon.GetAllSCIDVariableDetails(args[1])
			if vars == nil {
				logger.Printf("[%s] SCID2 not found\n", appName)
				continue
			}

			for _, h := range vars {
				switch k := h.Key.(type) {
				case string:
					if k == "C" {
						code, ok := h.Value.(string)
						if ok {
							diff = append(diff, code)
						}

						break
					}
				}
			}

			err = app.printDiff(diff, []string{args[0], args[1]})
			if err != nil {
				logger.Errorf("[%s] SCID Diff: %s\n", appName, err)
				continue
			}
		case "serve":
			if args == nil {
				line, err := app.readLine("Enter SCID to serve", "")
				if err != nil {
					if readError(err) {
						return
					}
					continue
				}

				args = []string{line}
			}

			if len(args[0]) != 64 {
				if args[0] == "local" {
					if len(args) < 2 {
						completer := readline.NewPrefixCompleter(
							completerFiles("."),
							readline.PcItem(filepath.Join("..", "..", "tela_tests", "app1")),
						)

						line, err := app.readLineWithCompleter("Enter path to local directory", "", completer)
						if err != nil {
							if readError(err) {
								return
							}
							continue
						}

						args = append(args, line)
					}

					path := args[1]
					if _, err = os.Stat(path); os.IsNotExist(err) {
						logger.Errorf("[%s] Directory %s does not exist\n", appName, path)
						continue
					}

					if err = app.serveLocal(path); err != nil {
						logger.Errorf("[%s] Serve local: %s\n", appName, err)
					}

					continue
				} else {
					logger.Errorf("[%s] Invalid SCID: %q\n", appName, args[0])
					continue
				}
			}

			// Serve TELA content from SCID arg
			link, err := tela.ServeTELA(args[0], app.endpoint)
			if err != nil {
				logger.Errorf("[%s] Serve: %s\n", appName, err)
				if !strings.Contains(err.Error(), "user defined no updates and content has been updated") {
					continue
				}

				// Prompt to allow serving the updated TELA content
				yes, err := app.readYesNo("Allow this updated content to be served")
				if err != nil {
					if readError(err) {
						return
					}
					continue
				}

				if !yes {
					continue
				}

				tela.AllowUpdates(true)
				link, err = tela.ServeTELA(args[0], app.endpoint)
				tela.AllowUpdates(false)
				if err != nil {
					logger.Errorf("[%s] Serve: %s\n", appName, err)
					continue
				}
			}

			if !app.openInBrowser {
				continue
			}

			// Open content in default browser if possible
			err = app.openBrowser(link)
			if err != nil {
				logger.Errorf("[%s] Open: %s\n", appName, err)
			}
		case "list":
			servers, total, local := app.getServerInfo()
			if total < 1 {
				logger.Printf("[%s] No active servers\n", appName)
				continue
			}

			sort.Slice(servers, func(i, j int) bool { return servers[i].Name < servers[j].Name })

			fmt.Println(searchDivider)
			fmt.Printf("Active servers:\n\n")
			for _, server := range servers {
				fmt.Printf("%-30s  Port: %-6s  SCID: %s\n", server.Name, server.Address, server.SCID)
			}

			if local {
				fmt.Println(searchDivider)
				fmt.Printf("Local  %-24s Port: %s\n", "", app.local.Address)
			}
			fmt.Println(searchDivider)
		case "shutdown":
			if args == nil {
				logger.Errorf("[%s] Missing shutdown argument\n", appName)
				continue
			}

			switch args[0] {
			case "all":
				tela.ShutdownTELA()
				app.shutdownLocalServer()
				continue
			case "local":
				app.shutdownLocalServer()
				continue
			case "tela":
				tela.ShutdownTELA()
				continue
			}

			if !tela.HasServer(args[0]) {
				logger.Errorf("[%s] %q is not a active server\n", appName, args[0])
				continue
			}

			// Shutdown a server
			tela.ShutdownServer(args[0])
		case "page-size":
			if args == nil {
				line, err := app.readLine("Set page size", "")
				if err != nil {
					if readError(err) {
						return
					}
					continue
				}

				args = []string{line}
			}

			u, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				logger.Errorf("[%s] %s\n", appName, err)
				continue
			}

			if u < 1 {
				logger.Errorf("[%s] Page size must be above 0\n", appName)
				continue
			}

			app.pageSize = int(u)
			logger.Printf("[%s] Page size set to: %d\n", appName, app.pageSize)
			err = shards.StoreSettingsValue(keys.pageSize, []byte(args[0]))
			if err != nil {
				logger.Debugf("[%s] Storing page size: %s\n", appName, err)
			}
		case "port-start":
			if args == nil {
				line, err := app.readLine("Set port start", "")
				if err != nil {
					if readError(err) {
						return
					}
					continue
				}

				args = []string{line}
			}

			u, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				logger.Errorf("[%s] %s\n", appName, err)
				continue
			}

			err = tela.SetPortStart(int(u))
			if err != nil {
				logger.Errorf("[%s] Set port start: %s\n", appName, err)
				continue
			}

			logger.Printf("[%s] TELA port start set to: %d\n", appName, tela.PortStart())
		case "max-servers":
			if args == nil {
				line, err := app.readLine("Set max servers", "")
				if err != nil {
					if readError(err) {
						return
					}
					continue
				}

				args = []string{line}
			}

			u, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				logger.Errorf("[%s] %s\n", appName, err)
				continue
			}

			tela.SetMaxServers(int(u))
			logger.Printf("[%s] Max TELA servers set to: %d\n", appName, tela.MaxServers())
		case "updates":
			if args == nil {
				completer := readline.NewPrefixCompleter(completerTrueFalse()...)
				line, err := app.readLineWithCompleter("Set updates (true/false)", "", completer)
				if err != nil {
					if readError(err) {
						return
					}
					continue
				}

				args = []string{line}
			}

			b, err := strconv.ParseBool(args[0])
			if err != nil {
				logger.Errorf("[%s] Updates: %s\n", appName, err)
				continue
			}

			tela.AllowUpdates(b)
			logger.Printf("[%s] Updates allowed: %t\n", appName, tela.UpdatesAllowed())
		case "browser":
			if args == nil {
				completer := readline.NewPrefixCompleter(completerTrueFalse()...)
				line, err := app.readLineWithCompleter("Set browser (true/false)", "", completer)
				if err != nil {
					if readError(err) {
						return
					}
					continue
				}

				args = []string{line}
			}

			b, err := strconv.ParseBool(args[0])
			if err != nil {
				logger.Errorf("[%s] Browser: %s\n", appName, err)
				continue
			}

			app.openInBrowser = b
			logger.Printf("[%s] Open in default browser: %t\n", appName, app.openInBrowser)
		case "colors":
			if args == nil {
				completer := readline.NewPrefixCompleter(completerTrueFalse()...)
				line, err := app.readLineWithCompleter("Set colors (true/false)", "", completer)
				if err != nil {
					if readError(err) {
						return
					}
					continue
				}

				args = []string{line}
			}

			b, err := strconv.ParseBool(args[0])
			if err != nil {
				logger.Errorf("[%s] Colors: %s\n", appName, err)
				continue
			}

			logger.EnableColors(b)
			logger.Printf("[%s] Colors: %t\n", appName, b)
		case "wallet":
			// Prompt for wallet file path or use arg
			var file string
			if args == nil {
				if app.wallet.disk != nil {
					logger.Errorf("[%s] Wallet %s is already open\n", appName, app.wallet.name)
					continue
				}

				completer := readline.NewPrefixCompleter(completerFiles("."))
				file, err = app.readLineWithCompleter("Enter wallet file path", "", completer)
				if err != nil {
					if readError(err) {
						return
					}
					continue
				}
			} else {
				// Close wallet if requested
				if args[0] == "close" {
					app.closeWallet()
					continue
				}

				if app.wallet.disk != nil {
					logger.Errorf("[%s] Wallet %s is already open\n", appName, app.wallet.name)
					continue
				}

				file = args[0]
			}

			if _, err := os.Stat(file); os.IsNotExist(err) {
				logger.Errorf("[%s] Wallet file %s does not exist\n", appName, file)
				continue
			}

			var password string
			if len(args) > 1 {
				password = args[1]

			}

			// Open wallet file with password
			err = app.openWallet(file, password)
			if err != nil {
				logger.Errorf("[%s] Opening wallet: %s\n", appName, err)
				continue
			}

			logger.Printf("[%s] Wallet connected: %s\n", appName, file)
			logger.Printf("[%s] Address: %s\n", appName, app.wallet.disk.GetAddress().String())
		case "balance":
			if app.wallet.disk == nil {
				logger.Errorf("[%s] Open a wallet file to view balance\n", appName)
				continue
			}

			app.wallet.disk.Sync_Wallet_Memory_With_Daemon()
			account := app.wallet.disk.GetAccount()

			logger.Printf("[%s] %sDERO%s %s\n", appName, logger.Color.Magenta(), logger.Color.End(), globals.FormatMoney(account.Balance_Mature))

			var balances []string
			for scid := range account.EntriesNative {
				if !scid.IsZero() {
					balanceMature, _ := app.wallet.disk.Get_Balance_scid(scid)
					balances = append(balances, fmt.Sprintf("%s%s:%s %s", logger.Color.Grey(), scid.String(), logger.Color.End(), globals.FormatMoney(balanceMature)))
				}
			}

			sort.Strings(balances)
			for _, balance := range balances {
				logger.Printf("[%s] %s\n", appName, balance)
			}
		case "rate":
			if app.wallet.disk == nil {
				logger.Errorf("[%s] Open a wallet file to rate SCID\n", appName)
				continue
			}

			pass, err := app.readWithPasswordPrompt("Confirm password")
			if err != nil {
				if readError(err) {
					return
				}
				continue
			}

			if !app.wallet.disk.Check_Password(string(pass)) {
				logger.Errorf("[%s] Invalid password\n", appName)
				continue
			}

			if args == nil {
				line, err := app.readLine("Enter SCID to rate", "")
				if err != nil {
					if readError(err) {
						return
					}
					continue
				}

				args = []string{line}
			}

			if len(args[0]) != 64 {
				logger.Errorf("[%s] Invalid SCID: %q\n", appName, args[0])
				continue
			}

			// Validate as TELA SC
			_, err = tela.GetRating(args[0], app.endpoint, 0)
			if err != nil {
				logger.Errorf("[%s] GetRating: %s\n", appName, err)
				continue
			}

			// Either user submits 0-99 number, or prompt for TELA category/detail rating
			if len(args) < 2 {
				categories := tela.Ratings.Categories()
				for i := 0; i < len(categories); i++ {
					fmt.Printf("%d: %s\n", i, categories[uint64(i)])
				}

				fP, err := app.readUint64("Enter rating number")
				if err != nil {
					if readError(err) {
						return
					}
					continue
				}

				if fP > 9 {
					logger.Errorf("[%s] Rating number must be less than 10\n", appName)
					continue
				}

				var details map[uint64]string
				if fP < 5 {
					details = tela.Ratings.NegativeDetails()
				} else {
					details = tela.Ratings.PositiveDetails()
				}

				for i := 0; i < len(details); i++ {
					fmt.Printf("%d: %s\n", i, details[uint64(i)])
				}

				sP, err := app.readUint64("Enter detail number")
				if err != nil {
					if readError(err) {
						return
					}
					continue
				}

				if sP > 9 {
					logger.Errorf("[%s] Detail number must be less than 10\n", appName)
					continue
				}

				line := fmt.Sprintf("%d", (fP*10)+sP)
				args = append(args, line)
			}

			rating, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				logger.Errorf("[%s] %s\n", appName, err)
				continue
			}

			category, _ := tela.Ratings.ParseString(rating)
			logger.Printf("[%s] Rating is: %s  %s\n", appName, args[1], category)

			yes, err := app.readYesNo("Confirm rating")
			if err != nil {
				if readError(err) {
					return
				}
				continue
			}

			if !yes {
				continue
			}

			txid, err := tela.Rate(app.wallet.disk, args[0], rating)
			if err != nil {
				logger.Errorf("[%s] Rate: %s\n", appName, err)
				continue
			}

			logger.Printf("[%s] Rate TXID: %s\n", appName, txid)
		case "install-doc":
			if app.wallet.disk == nil {
				logger.Errorf("[%s] Open a wallet file to install TELA-DOC\n", appName)
				continue
			}

			pass, err := app.readWithPasswordPrompt("Confirm password")
			if err != nil {
				if readError(err) {
					return
				}
				continue
			}

			if !app.wallet.disk.Check_Password(string(pass)) {
				logger.Errorf("[%s] Invalid password\n", appName)
				continue
			}

			// Prompt for DOC file path or use arg
			var filePath string
			if args == nil {
				completer := readline.NewPrefixCompleter(completerFiles("."))
				line, err := app.readLineWithCompleter("Enter DOC file path", "", completer)
				if err != nil {
					if readError(err) {
						return
					}
					continue
				}

				filePath = line
			} else {
				filePath = args[0]
				if _, err := os.Stat(filePath); os.IsNotExist(err) {
					logger.Errorf("[%s] DOC %q does not exists\n", appName, filePath)
					continue
				}
			}

			// Get DOC and install data
			fileName := filepath.Base(filePath)

			if !tela.IsAcceptedLanguage(tela.ParseDocType(fileName)) {
				logger.Errorf("[%s] %q is not a valid docType\n", appName, fileName)
				continue
			}

			data, err := os.ReadFile(filePath)
			if err != nil {
				logger.Errorf("[%s] Cannot read DOC data for %s: %s\n", appName, fileName, err)
				continue
			}

			docCode := string(data)
			if docCode == "" {
				logger.Errorf("[%s] DOC content is empty for %s\n", appName, fileName)
				continue
			}

			docType := tela.ParseDocType(fileName)
			if docType == "" {
				logger.Errorf("[%s] Invalid docType language for %s\n", appName, fileName)
				continue
			}

			headers, err := app.headersPrompt("DOC", nil)
			if err != nil {
				if readError(err) {
					return
				}
				continue
			}

			if headers[tela.HEADER_DURL] == "" {
				logger.Errorf("[%s] Missing %s header\n", appName, tela.HEADER_DURL)
				continue
			}

			var subDir string
			line, err := app.readLine("Enter DOC subDir", "")
			if err != nil {
				if readError(err) {
					return
				}
				continue
			}

			subDir = line

			ringsize, err := app.ringsizePrompt("DOC")
			if err != nil {
				if readError(err) {
					return
				}
				continue
			}

			// Sign docCode contents
			signature := app.wallet.disk.SignData(data)

			_, cStr, sStr, err := tela.ParseSignature(signature)
			if err != nil {
				logger.Errorf("[%s] Parse signature: %s\n", appName, err)
				continue
			}

			signatureStr := string(signature)
			b, ok := globals.Arguments["--debug"].(bool)
			if ok && b { // If debug enabled, print the full signature
				logger.Debugf("[%s] DOC signature:\n", appName)
				fmt.Println(signatureStr)
			} else {
				logger.Printf("[%s] DOC signature headers:\n", appName)
				for i, s := range strings.Split(signatureStr, "\n") {
					fmt.Println(s)
					if i > 2 {
						break
					}
				}
			}

			// Create TELA DOC
			doc := &tela.DOC{
				DocType: docType,
				Code:    docCode,
				SubDir:  subDir,
				DURL:    headers[tela.HEADER_DURL],
				Signature: tela.Signature{
					CheckC: cStr,
					CheckS: sStr,
				},
				Headers: tela.Headers{
					NameHdr:  fileName,
					DescrHdr: headers[tela.HEADER_DESCRIPTION],
					IconHdr:  headers[tela.HEADER_ICON_URL],
				},
			}

			yes, err := app.readYesNo("Confirm DOC install")
			if err != nil {
				if readError(err) {
					return
				}
				continue
			}

			if !yes {
				logger.Printf("[%s] DOC install cancelled\n", appName)
				continue
			}

			// Install TELA DOC
			txid, err := tela.Installer(app.wallet.disk, ringsize, doc)
			if err != nil {
				logger.Errorf("[%s] DOC install: %s\n", appName, err)
				continue
			}

			logger.Printf("[%s] DOC install TXID: %s\n", appName, txid)
		case "install-index":
			if app.wallet.disk == nil {
				logger.Errorf("[%s] Open a wallet file to install INDEX\n", appName)
				continue
			}

			pass, err := app.readWithPasswordPrompt("Confirm password")
			if err != nil {
				if readError(err) {
					return
				}
				continue
			}

			if !app.wallet.disk.Check_Password(string(pass)) {
				logger.Errorf("[%s] Invalid password\n", appName)
				continue
			}

			// Prompt for INDEX name or use arg
			var name string
			if args == nil {
				line, err := app.readLine("Enter INDEX name", "")
				if err != nil {
					if readError(err) {
						return
					}
					continue
				}

				name = line
			} else {
				name = args[0]
			}

			if name == "" {
				logger.Errorf("[%s] INDEX requires a name\n", appName)
				continue
			}

			// Get INDEX and install data
			index, err := app.indexPrompt(name, nil)
			if err != nil {
				if readError(err) {
					return
				}
				continue
			}

			modTag, err := app.modsPrompt()
			if err != nil {
				if readError(err) {
					return
				}
				continue
			}

			if modTag != "" {
				index.Mods = modTag
			}

			var ringsize uint64
			if index.Mods == "" {
				ringsize, err = app.ringsizePrompt("INDEX")
				if err != nil {
					if readError(err) {
						return
					}
					continue
				}
			} else {
				// MODs have no functionality for installs above RS 2
				ringsize = 2
				logger.Printf("[%s] MODs will use ringsize %d\n", appName, ringsize)
			}

			if ringsize > 2 {
				logger.Warnf("[%s] Ringsize is more than 2, this INDEX will be immutable\n", appName)
			}

			yes, err := app.readYesNo("Confirm INDEX install")
			if err != nil {
				if readError(err) {
					return
				}
				continue
			}

			if !yes {
				continue
			}

			// Install TELA-INDEX
			txid, err := tela.Installer(app.wallet.disk, ringsize, &index)
			if err != nil {
				logger.Errorf("[%s] INDEX install: %s\n", appName, err)
				continue
			}

			logger.Printf("[%s] INDEX install TXID: %s\n", appName, txid)
		case "update-index":
			if app.wallet.disk == nil {
				logger.Errorf("[%s] Open a wallet file to update INDEX\n", appName)
				continue
			}

			pass, err := app.readWithPasswordPrompt("Confirm password")
			if err != nil {
				if readError(err) {
					return
				}
				continue
			}

			if !app.wallet.disk.Check_Password(string(pass)) {
				logger.Errorf("[%s] Invalid password\n", appName)
				continue
			}

			if args == nil {
				line, err := app.readLine("Enter INDEX SCID to update", "")
				if err != nil {
					if readError(err) {
						return
					}
					continue
				}

				args = []string{line}
			}

			if len(args[0]) != 64 {
				logger.Errorf("[%s] Invalid SCID: %q\n", appName, args[0])
				continue
			}

			index, err := tela.GetINDEXInfo(args[0], app.endpoint)
			if err != nil {
				logger.Errorf("[%s] GetINDEXInfo: %s\n", appName, err)
				continue
			}

			if index.Author == "anon" {
				logger.Errorf("[%s] SCID %q cannot be updated\n", appName, args[0])
				continue
			} else if index.Author != app.wallet.disk.GetAddress().String() {
				logger.Errorf("[%s] Wallet address does not match author of SCID %q\n", appName, args[0])
				continue
			}

			// INDEX is behind latest version, it will update to latest anyways but ask to push current values with latest contract code
			latestVersion := tela.GetLatestContractVersion(false)
			if index.SCVersion.LessThan(latestVersion) {
				logger.Printf("[%s] INDEX is behind latest version v%s => v%s\n", appName, index.SCVersion.String(), latestVersion.String())
				yes, err := app.readYesNo("Update INDEX to latest using current values")
				if err != nil {
					if readError(err) {
						return
					}
					continue
				}

				if yes {
					txid, err := tela.Updater(app.wallet.disk, &index)
					if err != nil {
						logger.Errorf("[%s] INDEX update: %s\n", appName, err)
						continue
					}

					logger.Printf("[%s] INDEX update TXID: %s\n", appName, txid)
					continue
				}
			}

			// Prompt for INDEX name or use arg
			var name string
			if len(args) < 2 {
				line, err := app.readLine("Enter INDEX name", index.NameHdr)
				if err != nil {
					if readError(err) {
						return
					}
					continue
				}

				name = line
			} else {
				name = args[0]
			}

			index, err = app.indexPrompt(name, &index)
			if err != nil {
				if readError(err) {
					return
				}
				continue
			}

			if index.SCVersion != nil {
				// SC code is still behind TELA-MODs version so don't offer
				if !index.SCVersion.LessThan(tela.Version{Major: 1, Minor: 1, Patch: 0}) {
					_, err = tela.Mods.TagsAreValid(index.Mods)
					if err != nil {
						logger.Warnf("[%s] MODs are invalid, continue update to repair: %s\n", appName, err)
					}

					modTag, err := app.modsPrompt()
					if err != nil {
						if readError(err) {
							return
						}
						continue
					}

					index.Mods = modTag
				}
			}

			yes, err := app.readYesNo("Confirm INDEX update")
			if err != nil {
				if readError(err) {
					return
				}
				continue
			}

			if !yes {
				continue
			}

			txid, err := tela.Updater(app.wallet.disk, &index)
			if err != nil {
				logger.Errorf("[%s] INDEX update: %s\n", appName, err)
				continue
			}

			logger.Printf("[%s] INDEX update TXID: %s\n", appName, txid)
		case "mods":
			if args == nil || len(args) < 1 {
				// Print all MOD info
				allTelaMods := tela.Mods.GetAllMods()
				for _, m := range allTelaMods {
					printMODInfo(m, true)
				}
				fmt.Println(printDivider)

				continue
			}

			// Print MOD info by MODClass
			for _, c := range tela.Mods.GetAllClasses() {
				if args[0] == c.Tag {
					printMODClassInfo(c)
					allTelaMods := tela.Mods.GetAllMods()
					for _, m := range allTelaMods {
						if strings.HasPrefix(m.Tag, c.Tag) {
							printMODInfo(m, true)
						}
					}
					fmt.Println(printDivider)

					break
				}
			}
		case "mod-info":
			if args == nil {
				completer := readline.NewPrefixCompleter(completerMODs("")...)
				line, err := app.readLineWithCompleter("Enter modTag", "", completer)
				if err != nil {
					if readError(err) {
						return
					}
					continue
				}

				args = []string{line}
			}

			mod := tela.Mods.GetMod(args[0])
			if mod.Tag == "" {
				logger.Errorf("[%s] MOD %q not found\n", appName, args[0])
				continue
			}

			printMODInfo(mod, true)
			fmt.Println(printDivider)
		case "set-var":
			if app.wallet.disk == nil {
				logger.Errorf("[%s] Open a wallet file to set INDEX variable\n", appName)
				continue
			}

			pass, err := app.readWithPasswordPrompt("Confirm password")
			if err != nil {
				if readError(err) {
					return
				}
				continue
			}

			if !app.wallet.disk.Check_Password(string(pass)) {
				logger.Errorf("[%s] Invalid password\n", appName)
				continue
			}

			if args == nil {
				line, err := app.readLine("Enter INDEX SCID", "")
				if err != nil {
					if readError(err) {
						return
					}
					continue
				}

				args = []string{line}
			}

			if len(args[0]) != 64 {
				logger.Errorf("[%s] Invalid SCID: %q\n", appName, args[0])
				continue
			}

			index, err := tela.GetINDEXInfo(args[0], app.endpoint)
			if err != nil {
				logger.Errorf("[%s] GetINDEXInfo: %s\n", appName, err)
				continue
			}

			if index.Author == "anon" {
				logger.Errorf("[%s] SCID %q cannot set variables\n", appName, args[0])
				continue
			}

			modTags, err := tela.Mods.TagsAreValid(index.Mods)
			if err != nil {
				logger.Errorf("[%s] Invalid MOD tags: %q\n", appName, err)
				continue
			}

			// Handle MOD requirements
			var isOwner = index.Author == app.wallet.disk.GetAddress().String()
			var canSetVars, ownerOnly, singleUse, immutable bool
			for _, t := range modTags {
				if strings.HasPrefix(t, "vs") {
					canSetVars = true
					switch t {
					case tela.Mods.Tag(0):
						ownerOnly = true
					case tela.Mods.Tag(1):
						immutable = true
						ownerOnly = true
					case tela.Mods.Tag(2):
						singleUse = true
					case tela.Mods.Tag(4):
						immutable = true
					}
				}
			}

			if !canSetVars {
				logger.Errorf("[%s] SCID %q cannot set variables\n", appName, args[0])
				continue
			}

			if ownerOnly && !isOwner {
				logger.Errorf("[%s] Wallet address does not match author of SCID %q\n", appName, args[0])
				continue
			}

			if len(args) < 2 {
				line, err := app.readLine("Enter key to store", "")
				if err != nil {
					if readError(err) {
						return
					}
					continue
				}

				args = append(args, line)
			}

			var checkKey string
			switch {
			case singleUse && !isOwner:
				checkKey = fmt.Sprintf("var_%s_%s", app.wallet.disk.GetAddress().String(), args[1])
			case immutable:
				checkKey = fmt.Sprintf("ivar_%s", args[1])
				if !isOwner {
					checkKey = fmt.Sprintf("ivar_%s_%s", app.wallet.disk.GetAddress().String(), args[1])
				}
			}

			// MOD does not allow overwrites
			if checkKey != "" {
				value, exists, err := tela.KeyExists(args[0], checkKey, app.endpoint)
				if err != nil {
					logger.Errorf("[%s] KeyExists: %s\n,", appName, err)
					continue
				}

				if exists {
					logger.Errorf("[%s] Cannot change key %q with store: %s\n", appName, checkKey, value)
					continue
				}

				logger.Warnf("[%s] Key store %q will be immutable on SCID %s\n", appName, args[1], args[0])
			} else {
				// MOD allows overwrites or isOwner
				checkKey = fmt.Sprintf("var_%s", args[1])
				if !isOwner {
					checkKey = fmt.Sprintf("var_%s_%s", app.wallet.disk.GetAddress().String(), args[1])
				}

				_, exists, err := tela.KeyExists(args[0], checkKey, app.endpoint)
				if err != nil {
					logger.Errorf("[%s] KeyExists: %s\n", appName, err)
					continue
				}

				if exists {
					yes, err := app.readYesNo(fmt.Sprintf("Key %q already exists, overwrite", args[1]))
					if err != nil {
						if readError(err) {
							return
						}
						continue
					}

					if !yes {
						continue
					}
				}
			}

			if len(args) < 3 {
				line, err := app.readLine(fmt.Sprintf("Enter value for key %q", args[1]), "")
				if err != nil {
					if readError(err) {
						return
					}
					continue
				}

				args = append(args, line)
			}

			yes, err := app.readYesNo("Confirm SetVar")
			if err != nil {
				if readError(err) {
					return
				}
				continue
			}

			if !yes {
				continue
			}

			txid, err := tela.SetVar(app.wallet.disk, args[0], args[1], args[2])
			if err != nil {
				logger.Errorf("[%s] SetVar: %s\n", appName, err)
				continue
			}

			logger.Printf("[%s] SetVar TXID: %s\n", appName, txid)
		case "delete-var":
			if app.wallet.disk == nil {
				logger.Errorf("[%s] Open a wallet file to delete INDEX variable\n", appName)
				continue
			}

			pass, err := app.readWithPasswordPrompt("Confirm password")
			if err != nil {
				if readError(err) {
					return
				}
				continue
			}

			if !app.wallet.disk.Check_Password(string(pass)) {
				logger.Errorf("[%s] Invalid password\n", appName)
				continue
			}

			if args == nil {
				line, err := app.readLine("Enter INDEX SCID", "")
				if err != nil {
					if readError(err) {
						return
					}
					continue
				}

				args = []string{line}
			}

			if len(args[0]) != 64 {
				logger.Errorf("[%s] Invalid SCID: %q\n", appName, args[0])
				continue
			}

			index, err := tela.GetINDEXInfo(args[0], app.endpoint)
			if err != nil {
				logger.Errorf("[%s] GetINDEXInfo: %s\n", appName, err)
				continue
			}

			if index.Author == "anon" {
				logger.Errorf("[%s] SCID %q cannot delete variables\n", appName, args[0])
				continue
			}

			modTags, err := tela.Mods.TagsAreValid(index.Mods)
			if err != nil {
				logger.Errorf("[%s] Invalid MOD tags: %q\n", appName, err)
				continue
			}

			// Handle MOD requirements
			var canDeleteVars bool
			for _, t := range modTags {
				if strings.HasPrefix(t, "vs") {
					canDeleteVars = true
					switch t {
					case tela.Mods.Tag(1), tela.Mods.Tag(4):
						canDeleteVars = false
					}
				}
			}

			if !canDeleteVars {
				logger.Errorf("[%s] SCID %q cannot delete variables\n", appName, args[0])
				continue
			}

			if index.Author != app.wallet.disk.GetAddress().String() {
				logger.Errorf("[%s] Wallet address does not match author of SCID %q\n", appName, args[0])
				continue
			}

			if len(args) < 2 {
				line, err := app.readLine("Enter key to delete", "")
				if err != nil {
					if readError(err) {
						return
					}
					continue
				}

				args = append(args, line)
			}

			_, exists, err := tela.KeyExists(args[0], fmt.Sprintf("var_%s", args[1]), app.endpoint)
			if err != nil {
				logger.Errorf("[%s] KeyExists: %s\n", appName, err)
				continue
			}

			if !exists {
				logger.Errorf("[%s] Key %q does not exists on %s\n", appName, args[1], args[0])
				continue
			}

			yes, err := app.readYesNo("Confirm DeleteVar")
			if err != nil {
				if readError(err) {
					return
				}
				continue
			}

			if !yes {
				continue
			}

			txid, err := tela.DeleteVar(app.wallet.disk, args[0], args[1])
			if err != nil {
				logger.Errorf("[%s] DeleteVar: %s\n", appName, err)
				continue
			}

			logger.Printf("[%s] DeleteVar TXID: %s\n", appName, txid)
		case "sc-execute":
			if app.wallet.disk == nil {
				logger.Errorf("[%s] Open a wallet file to execute a smart contract\n", appName)
				continue
			}

			pass, err := app.readWithPasswordPrompt("Confirm password")
			if err != nil {
				if readError(err) {
					return
				}
				continue
			}

			if !app.wallet.disk.Check_Password(string(pass)) {
				logger.Errorf("[%s] Invalid password\n", appName)
				continue
			}

			if args == nil {
				line, err := app.readLine("Enter SCID", "")
				if err != nil {
					if readError(err) {
						return
					}
					continue
				}

				args = []string{line}
			}

			if len(args[0]) != 64 {
				logger.Errorf("[%s] Invalid SCID: %q\n", appName, args[0])
				continue
			}

			index, err := tela.GetINDEXInfo(args[0], app.endpoint)
			if err != nil {
				logger.Errorf("[%s] GetINDEXInfo: %s\n", appName, err)
				continue
			}

			if index.Author == "anon" {
				logger.Errorf("[%s] SCID %q cannot execute functions\n", appName, args[0])
				continue
			}

			isOwner := index.Author == app.wallet.disk.GetAddress().String()

			if len(args) < 2 {
				completer := readline.NewPrefixCompleter(completerSCFunctionNames(isOwner, index.SC)...)
				line, err := app.readLineWithCompleter("Enter function name", "", completer)
				if err != nil {
					if readError(err) {
						return
					}
					continue
				}

				args = append(args, line)
			}

			if args[1] == "" {
				logger.Errorf("[%s] Invalid function name: %q\n", appName, args[0])
				continue
			}

			transfers, arguments, err := app.executeContractPrompt(args[0], args[1], index.SC)
			if err != nil {
				if readError(err) {
					return
				}

				logger.Errorf("[%s] %s: %s\n", appName, args[1], err)
				continue
			}

			// Display the transfer and arguments for confirmation
			if indentTransfers, err := json.MarshalIndent(transfers, "", " "); err == nil {
				fmt.Printf("%s\n", string(indentTransfers))
			} else {
				fmt.Printf("%v\n", transfers)
			}

			if indentArguments, err := json.MarshalIndent(arguments, "", " "); err == nil {
				fmt.Printf("%s\n", string(indentArguments))
			} else {
				fmt.Printf("%v\n", arguments)
			}

			yes, err := app.readYesNo(fmt.Sprintf("Confirm %s", args[1]))
			if err != nil {
				if readError(err) {
					return
				}
				continue
			}

			if !yes {
				continue
			}

			txid, err := tela.Transfer(app.wallet.disk, 2, transfers, arguments)
			if err != nil {
				logger.Errorf("[%s] %s: %s\n", appName, args[1], err)
				continue
			}

			logger.Printf("[%s] %s TXID: %s\n", appName, args[1], txid)
		case "gnomon":
			if args == nil {
				logger.Errorf("[%s] Missing Gnomon argument\n", appName)
				continue
			}

			// Gnomon offline commands
			switch args[0] {
			case "clean":
				if gnomon.Indexer != nil {
					logger.Printf("[%s] Shut down Gnomon indexer before cleaning DB\n", appName)
					continue
				}

				if len(args) < 2 {
					line, err := app.readLine("Clean what?", "")
					if err != nil {
						if readError(err) {
							return
						}
						continue
					}

					args = append(args, line)
				}

				var subDir string
				switch args[1] {
				case "mainnet":
					subDir = "gnomon"
				case "testnet":
					subDir = "gnomon_testnet"
				case "simulator":
					subDir = "gnomon_simulator"
				default:
					logger.Errorf("[%s] Unknown clean path %q\n", appName, args[1])
					continue
				}

				yes, err := app.readYesNo(fmt.Sprintf("Delete %s%s%s DB", logger.Color.Red(), subDir, logger.Color.End()))
				if err != nil {
					if readError(err) {
						return
					}
					continue
				}

				if !yes {
					continue
				}

				os.RemoveAll(filepath.Join("datashards", subDir))
				logger.Printf("[%s] %s DB deleted\n", appName, subDir)
				continue
			}

			if !walletapi.IsDaemonOnline() {
				logger.Errorf("[%s] Daemon %s not online for Gnomon\n", appName, app.endpoint)
				continue
			}

			// Gnomon online commands
			switch args[0] {
			case "start":
				if gnomon.Indexer != nil {
					logger.Printf("[%s] Gnomon is already running\n", appName)
					continue
				}

				startGnomon(app.endpoint)
				if gnomon.Indexer != nil {
					time.Sleep(time.Second * 5)
				}
			case "stop":
				stopGnomon()
			case "resync":
				var subDir string
				switch getNetworkInfo() {
				case shards.Value.Network.Testnet():
					subDir = "gnomon_testnet"
				case shards.Value.Network.Simulator():
					subDir = "gnomon_simulator"
				default:
					subDir = "gnomon"
				}

				yes, err := app.readYesNo(fmt.Sprintf("Resync %s%s%s DB", logger.Color.Red(), subDir, logger.Color.End()))
				if err != nil {
					if readError(err) {
						return
					}
					continue
				}

				if !yes {
					continue
				}

				stopGnomon()
				os.RemoveAll(filepath.Join("datashards", subDir))
				logger.Printf("[%s] Resyncing %s%s%s DB\n", appName, logger.Color.Green(), subDir, logger.Color.End())
				time.Sleep(time.Second)
				startGnomon(app.endpoint)
				if gnomon.Indexer != nil {
					time.Sleep(time.Second * 5)
				}
			case "add":
				if len(args) < 2 {
					line, err := app.readLine("Enter SCID(s)", "")
					if err != nil {
						if readError(err) {
							return
						}
						continue
					}

					args = append(args, line)
				}

				if len(args[1]) != 64 {
					logger.Errorf("[%s] Invalid SCID: %q\n", appName, args[1])
					continue
				}

				scidsToAdd := args[1:]

				// Check that all entered SCIDs are valid
				var invalidSCID string
				for _, sc := range scidsToAdd {
					if len(strings.TrimSpace(sc)) != 64 {
						invalidSCID = sc
						break
					}
				}

				if invalidSCID != "" {
					logger.Errorf("[%s] Invalid SCID: %q\n", appName, invalidSCID)
					continue
				}

				err = app.addToIndex(scidsToAdd)
				if err != nil {
					logger.Errorf("[%s] Could not add SCID(s): %s\n", appName, err)
					continue
				}

				logger.Printf("[%s] SCID(s) added\n", appName)
			default:
				logger.Printf("[%s] Unknown Gnomon command: %q\n", appName, args[0])
			}
		case "search":
			if gnomon.Indexer == nil {
				logger.Errorf("[%s] Gnomon is not online\n", appName)
				continue
			}

			if args == nil {
				line, err := app.readLine("Enter search query", "")
				if err != nil {
					if readError(err) {
						return
					}
					continue
				}

				args = []string{line}
			}

			switch args[0] {
			case "scid":
				if len(args) < 2 {
					line, err := app.readLine("Enter SCID to search", "")
					if err != nil {
						if readError(err) {
							return
						}
						continue
					}

					args = append(args, line)
				}

				if len(args) == 3 {
					switch args[1] {
					case "vars":
						if len(args[2]) != 64 {
							logger.Errorf("[%s] Invalid SCID: %q\n", appName, args[2])
							continue
						}

						vars := gnomon.GetAllSCIDVariableDetails(args[2])
						if vars == nil {
							logger.Printf("[%s] SCID not found\n", appName)
							continue
						}

						var varResults []string
						for _, h := range vars {
							switch k := h.Key.(type) {
							case string:
								if k == "C" {
									continue
								}

								switch v := h.Value.(type) {
								case string:
									varResults = append(varResults, fmt.Sprintf("{%s%q%s, %q}", logger.Color.Cyan(), k, logger.Color.End(), v))
								case uint64:
									varResults = append(varResults, fmt.Sprintf("{%s%q%s, %d}", logger.Color.Cyan(), k, logger.Color.End(), v))
								default:
									varResults = append(varResults, fmt.Sprintf("{%s%q%s, %v}", logger.Color.Cyan(), k, logger.Color.End(), v))
								}
							case uint64:
								switch v := h.Value.(type) {
								case string:
									varResults = append(varResults, fmt.Sprintf("{%s%d%s, %q}", logger.Color.Cyan(), k, logger.Color.End(), v))
								case uint64:
									varResults = append(varResults, fmt.Sprintf("{%s%d%s, %d}", logger.Color.Cyan(), k, logger.Color.End(), v))
								default:
									varResults = append(varResults, fmt.Sprintf("{%s%d%s, %v}", logger.Color.Cyan(), k, logger.Color.End(), v))
								}
							default:
								switch v := h.Value.(type) {
								case string:
									varResults = append(varResults, fmt.Sprintf("{%s%v%s, %q}", logger.Color.Cyan(), k, logger.Color.End(), v))
								case uint64:
									varResults = append(varResults, fmt.Sprintf("{%s%v%s, %d}", logger.Color.Cyan(), k, logger.Color.End(), v))
								default:
									varResults = append(varResults, fmt.Sprintf("{%s%v%s, %v}", logger.Color.Cyan(), k, logger.Color.End(), v))
								}
							}
						}

						if len(varResults) < 1 {
							logger.Errorf("[%s] No variables found\n", appName)
							continue
						}

						sort.Strings(varResults)

						// Don't want the divider between variable prints from app.paging
						display := app.pageSize - 1
						resultLen := len(varResults)

						isPaged := false
						if resultLen > app.pageSize {
							isPaged = true
							logger.Printf("[%s] Showing %d of %d variables\n", appName, app.pageSize, resultLen)
						}

						for printed, vLine := range varResults {
							fmt.Println(vLine)

							end := printed == resultLen-1
							if printed >= display && !end {
								var yes bool
								yes, err = app.readYesNo(fmt.Sprintf("Show more variables? (%d)", (resultLen-1)-printed))
								if err != nil {
									return
								}

								if !yes {
									break
								}

								display = display + app.pageSize
							}

							if isPaged && end {
								logger.Printf("[%s] End of variables\n", appName)
							}
						}
					default:
						logger.Errorf("[%s] Unknown SCID command: %q\n", appName, args[1])
					}

					continue
				}

				if len(args[1]) != 64 {
					logger.Errorf("[%s] Invalid SCID: %q\n", appName, args[1])
					continue
				}

				all := gnomon.GetAllOwnersAndSCIDs()
				if len(all) < 1 {
					logger.Printf("[%s] No SCIDs found\n", appName)
					continue
				}

				var found bool
				var owner string
				for sc := range all {
					if sc == args[1] {
						found = true
						owner = gnomon.GetOwnerAddress(sc)
						break
					}
				}

				if !found {
					logger.Printf("[%s] SCID %q not found\n", appName, args[1])
					continue
				}

				dURL, likesRatio, err := app.getLikesRatio(args[1], true)
				if err != nil {
					logger.Printf("[%s] SCID search: %s\n", appName, err)
					continue
				}

				var resultLines [][]string
				resultLines = append(resultLines, parseSearchQuery(args[1], owner, dURL, likesRatio))

				err = app.paging(resultLines)
				if err != nil {
					if readError(err) {
						return
					}
				}
			case "value":
				all := gnomon.GetAllOwnersAndSCIDs()
				if len(all) < 1 {
					logger.Printf("[%s] No SCIDs found\n", appName)
					continue
				}

				if len(args) < 2 {
					line, err := app.readLine("Enter value to search", "")
					if err != nil {
						if readError(err) {
							return
						}
						continue
					}

					args = append(args, line)
				}

				var resultLines [][]string
				for sc := range all {
					dURL, likesRatio, err := app.getLikesRatio(sc, true)
					if err != nil {
						continue
					}

					vStr, _ := gnomon.GetSCIDKeysByValue(sc, args[1])
					if vStr == nil {
						continue
					}

					owner := gnomon.GetOwnerAddress(sc)
					resultLines = append(resultLines, parseSearchQuery(sc, owner, dURL, likesRatio))
				}

				err = app.paging(resultLines)
				if err != nil {
					if readError(err) {
						return
					}
				}
			case "key":
				all := gnomon.GetAllOwnersAndSCIDs()
				if len(all) < 1 {
					logger.Printf("[%s] No SCIDs found\n", appName)
					continue
				}

				if len(args) < 2 {
					line, err := app.readLine("Enter key to search", "")
					if err != nil {
						if readError(err) {
							return
						}
						continue
					}

					args = append(args, line)
				}

				var resultLines [][]string
				for sc := range all {
					dURL, likesRatio, err := app.getLikesRatio(sc, true)
					if err != nil {
						continue
					}

					vStr, _ := gnomon.GetSCIDValuesByKey(sc, args[1])
					if vStr == nil {
						continue
					}

					owner := gnomon.GetOwnerAddress(sc)
					resultLines = append(resultLines, parseSearchQuery(sc, owner, dURL, likesRatio))
				}

				err = app.paging(resultLines)
				if err != nil {
					if readError(err) {
						return
					}
				}
			case "all":
				all := gnomon.GetAllOwnersAndSCIDs()
				if len(all) < 1 {
					logger.Printf("[%s] No SCIDs found\n", appName)
					continue
				}

				var resultLines [][]string
				for sc := range all {
					dURL, likesRatio, err := app.getLikesRatio(sc, true)
					if err != nil {
						continue
					}

					owner := gnomon.GetOwnerAddress(sc)
					resultLines = append(resultLines, parseSearchQuery(sc, owner, dURL, likesRatio))
				}

				err := app.paging(resultLines)
				if err != nil {
					if readError(err) {
						return
					}
				}
			case "docs":
				all := gnomon.GetAllOwnersAndSCIDs()
				if len(all) < 1 {
					logger.Printf("[%s] No SCIDs found\n", appName)
					continue
				}

				err := app.paging(app.searchDOCInfo(all, false, args...))
				if err != nil {
					if readError(err) {
						return
					}
				}
			case "indexes":
				all := gnomon.GetAllOwnersAndSCIDs()
				if len(all) < 1 {
					logger.Printf("[%s] No SCIDs found\n", appName)
					continue
				}

				err := app.paging(app.searchINDEXInfo(all, false))
				if err != nil {
					if readError(err) {
						return
					}
				}
			case "libs":
				libKeys, libMap := app.getLibraries()
				if len(libKeys) < 1 {
					logger.Printf("[%s] No SCIDs found\n", appName)
					continue
				}

				var resultLines [][]string
				for _, lib := range libKeys {
					resultLines = append(resultLines, parseLibraryInfo(lib, libMap))
				}

				err := app.paging(resultLines)
				if err != nil {
					if readError(err) {
						return
					}
				}
			case "durl":
				if len(args) < 2 {
					line, err := app.readLine("Enter dURL to search", "")
					if err != nil {
						if readError(err) {
							return
						}
						continue
					}

					args = append(args, line)
				}

				all := gnomon.GetAllOwnersAndSCIDs()
				if len(all) < 1 {
					logger.Printf("[%s] No SCIDs found\n", appName)
					continue
				}

				var resultLines [][]string
				for sc := range all {
					dURL, likesRatio, err := app.getLikesRatio(sc, true)
					if err != nil {
						continue
					}

					owner := gnomon.GetOwnerAddress(sc)
					if strings.Contains(strings.ToLower(dURL), strings.ToLower(args[1])) {
						resultLines = append(resultLines, parseSearchQuery(sc, owner, dURL, likesRatio))
					}
				}

				err := app.paging(resultLines)
				if err != nil {
					if readError(err) {
						return
					}
				}
			case "code":
				if len(args) < 2 {
					line, err := app.readLine("Enter SCID", "")
					if err != nil {
						if readError(err) {
							return
						}
						continue
					}

					args = append(args, line)
				}

				if len(args[1]) != 64 {
					logger.Errorf("[%s] Invalid SCID: %q\n", appName, args[1])
					continue
				}

				vars := gnomon.GetAllSCIDVariableDetails(args[1])
				if vars == nil {
					logger.Printf("[%s] SCID not found\n", appName)
					continue
				}

				for _, h := range vars {
					switch k := h.Key.(type) {
					case string:
						if k == "C" {
							code, ok := h.Value.(string)
							if ok {
								fmt.Println(code)
							}

							break
						}
					}
				}
			case "line":
				if len(args) < 2 {
					line, err := app.readLine("Enter code line", "")
					if err != nil {
						if readError(err) {
							return
						}
						continue
					}

					args = append(args, line)
				}

				all := gnomon.GetAllOwnersAndSCIDs()
				if len(all) < 1 {
					logger.Printf("[%s] No SCIDs found\n", appName)
					continue
				}

				var resultLines [][]string
				for sc := range all {
					dURL, likesRatio, err := app.getLikesRatio(sc, true)
					if err != nil {
						continue
					}

					vars := gnomon.GetAllSCIDVariableDetails(sc)
					if vars == nil {
						continue
					}

					owner := gnomon.GetOwnerAddress(sc)

					for _, h := range vars {
						switch k := h.Key.(type) {
						case string:
							if k == "C" {
								code, ok := h.Value.(string)
								if ok {
									if strings.Contains(code, args[1]) {
										resultLines = append(resultLines, parseSearchQuery(sc, owner, dURL, likesRatio))
									}
								}

								break
							}
						}
					}
				}

				err := app.paging(resultLines)
				if err != nil {
					if readError(err) {
						return
					}
				}
			case "author":
				if len(args) < 2 {
					line, err := app.readLine("Enter address to search", "")
					if err != nil {
						if readError(err) {
							return
						}
						continue
					}

					args = append(args, line)
				}

				_, err := globals.ParseValidateAddress(args[1])
				if err != nil {
					logger.Errorf("[%s] %q is not a valid DERO address\n", appName, args[1])
					continue
				}

				all := gnomon.GetAllOwnersAndSCIDs()
				if len(all) < 1 {
					logger.Printf("[%s] No SCIDs found\n", appName)
					continue
				}

				var resultLines [][]string
				for sc := range all {
					dURL, likesRatio, err := app.getLikesRatio(sc, true)
					if err != nil {
						continue
					}

					owner := gnomon.GetOwnerAddress(sc)
					if strings.ToLower(owner) == args[1] {
						resultLines = append(resultLines, parseSearchQuery(sc, owner, dURL, likesRatio))
					}
				}

				err = app.paging(resultLines)
				if err != nil {
					if readError(err) {
						return
					}
				}
			case "min-likes":
				if len(args) < 2 {
					line, err := app.readLine("Set min likes %", "")
					if err != nil {
						if readError(err) {
							return
						}
						continue
					}

					args = append(args, line)
				}

				f, err := strconv.ParseFloat(args[1], 64)
				if err != nil {
					logger.Errorf("[%s] %s\n", appName, err)
					continue
				}

				if f > 100 {
					logger.Errorf("[%s] Minimum likes %% must be below 100%%\n", appName)
					continue
				}

				app.minLikes = f
				logger.Printf("[%s] Minimum likes for search queries set to: %.0f%%\n", appName, app.minLikes)
				err = shards.StoreSettingsValue(keys.minLikes, []byte(args[1]))
				if err != nil {
					logger.Debugf("[%s] Storing minimum likes: %s\n", appName, err)
				}
			case "ratings":
				if len(args) < 2 {
					line, err := app.readLine("Enter SCID", "")
					if err != nil {
						if readError(err) {
							return
						}
						continue
					}

					args = append(args, line)
				}

				if len(args[1]) != 64 {
					logger.Errorf("[%s] Invalid SCID: %q\n", appName, args[1])
					continue
				}

				height := uint64(0)
				if len(args) > 2 {
					if u, err := strconv.ParseUint(args[2], 10, 64); err == nil {
						height = u
					}
				}

				ratings, err := tela.GetRating(args[1], app.endpoint, height)
				if err != nil {
					logger.Errorf("[%s] GetRating: %s\n", appName, err)
					continue
				}

				nameHdr, _ := gnomon.GetSCIDValuesByKey(args[1], "nameHdr")
				if nameHdr == nil {
					nameHdr = append(nameHdr, "?")
				}

				fmt.Printf("Name: %s\n", nameHdr[0])
				fmt.Printf("Likes: %s%d%s\n", logger.Color.Green(), ratings.Likes, logger.Color.End())
				fmt.Printf("Dislikes: %s%d%s\n", logger.Color.Red(), ratings.Dislikes, logger.Color.End())

				if height > 0 && ratings.Likes+ratings.Dislikes > 0 {
					fmt.Printf("Omitting ratings below height %d\n", height)
				}

				if len(ratings.Ratings) > 0 {
					fmt.Printf("Average: %.1f/10   (%s)\n", ratings.Average, tela.Ratings.Category(uint64(ratings.Average)))
				}

				var resultLines [][]string
				for _, r := range ratings.Ratings {
					rating, err := tela.Ratings.ParseString(r.Rating)
					if err != nil {
						logger.Debugf("[%s] Ratings: %s\n", appName, err)
						continue
					}

					resultLines = append(resultLines, []string{fmt.Sprintf("Address: %s  Height: %-10d Rating: %-5s %s", r.Address, r.Height, fmt.Sprintf("[%d]", r.Rating), rating)})
				}

				err = app.paging(resultLines)
				if err != nil {
					if readError(err) {
						return
					}
				}
			case "my":
				if app.wallet.disk == nil {
					logger.Errorf("[%s] Open a wallet file to search %q queries\n", appName, "my")
					continue
				}

				if len(args) < 2 {
					completer := readline.NewPrefixCompleter(
						readline.PcItem("docs", completerDocType()...),
						readline.PcItem("indexes"),
					)

					line, err := app.readLineWithCompleter("Enter query", "", completer)
					if err != nil {
						if readError(err) {
							return
						}
						continue
					}

					split := strings.Split(strings.TrimSpace(line), " ")
					if len(split) > 1 {
						args = append(args, split...) // get docType arg if provided in prompt
					} else {
						args = append(args, line)
					}
				}

				all := gnomon.GetAllOwnersAndSCIDs()
				if len(all) < 1 {
					logger.Printf("[%s] No SCIDs found\n", appName)
					continue
				}

				var resultLines [][]string
				switch args[1] {
				case "docs":
					resultLines = app.searchDOCInfo(all, true, args[1:]...)
				case "indexes":
					resultLines = app.searchINDEXInfo(all, true)
				default:
					logger.Errorf("[%s] Unknown search query: %q\n", appName, fmt.Sprintf("%s %s", args[0], args[1]))
					continue
				}

				err := app.paging(resultLines)
				if err != nil {
					if readError(err) {
						return
					}
				}
			case "exclude":
				if len(args) < 2 {
					logger.Errorf("[%s] Missing search exclude argument\n", appName)
					continue
				}

				switch args[1] {
				case "view":
					var resultLines [][]string
					for _, exclude := range app.exclusions {
						resultLines = append(resultLines, []string{exclude})
					}

					if len(resultLines) < 1 {
						logger.Printf("[%s] No search exclusions enabled\n", appName)
						continue
					}

					err := app.paging(resultLines)
					if err != nil {
						if readError(err) {
							return
						}
					}
				case "clear":
					if len(app.exclusions) < 1 {
						logger.Printf("[%s] No search exclusions enabled\n", appName)
						continue
					}

					yes, err := app.readYesNo("Clear all search exclusions")
					if err != nil {
						if readError(err) {
							return
						}
						continue
					}

					if !yes {
						continue
					}

					app.exclusions = []string{}
					err = shards.DeleteSettingsKey(keys.exclude)
					if err != nil {
						logger.Debugf("[%s] Deleting stored search exclusions: %s\n", appName, err)
					}

					logger.Printf("[%s] Search exclusions cleared\n", appName)
				case "add":
					if len(args) < 3 {
						line, err := app.readLine("Enter search exclusion", "")
						if err != nil {
							if readError(err) {
								return
							}
							continue
						}

						args = append(args, line)
					}

					if args[2] == "" {
						logger.Errorf("[%s] Invalid search exclusion\n", appName)
						continue
					}

					var exists bool
					for _, exclude := range app.exclusions {
						if args[2] == exclude {
							exists = true
							break
						}
					}

					if exists {
						logger.Printf("[%s] Exclusion %q exists already\n", appName, args[2])
						continue
					}

					app.exclusions = append(app.exclusions, args[2])
					logger.Printf("[%s] Search exclusion added\n", appName)

					storeExclusions, err := json.Marshal(app.exclusions)
					if err != nil {
						logger.Errorf("[%s] Marshal search exclusions: %s\n", appName, err)
						continue
					}

					err = shards.StoreSettingsValue(keys.exclude, storeExclusions)
					if err != nil {
						logger.Debugf("[%s] Storing search exclusions: %s\n", appName, err)
					}
				case "remove":
					if len(app.exclusions) < 1 {
						logger.Printf("[%s] No search exclusions enabled\n", appName)
						continue
					}

					if len(args) < 3 {
						completer := readline.NewPrefixCompleter(app.completerSearchExclusions()...)
						line, err := app.readLineWithCompleter("Enter search exclusion", "", completer)
						if err != nil {
							if readError(err) {
								return
							}
							continue
						}

						args = append(args, line)
					}

					var removed bool
					var tempExclusions []string
					for i, exclude := range app.exclusions {
						if exclude == args[2] {
							tempExclusions = append(app.exclusions[:i], app.exclusions[i+1:]...)
							removed = true
							break
						}
					}

					if !removed {
						logger.Printf("[%s] Search exclusion %q not found\n", appName, args[2])
						continue
					}

					app.exclusions = tempExclusions
					logger.Printf("[%s] Search exclusion removed\n", appName)

					storeExclusions, err := json.Marshal(app.exclusions)
					if err != nil {
						logger.Errorf("[%s] Marshal search exclusions: %s\n", appName, err)
						continue
					}

					err = shards.StoreSettingsValue(keys.exclude, storeExclusions)
					if err != nil {
						logger.Debugf("[%s] Storing search exclusions: %s\n", appName, err)
					}
				default:
					logger.Errorf("[%s] Unknown search exclude query: %q\n", appName, args[1])
				}
			default:
				logger.Errorf("[%s] Unknown search query: %q\n", appName, args[0])
			}
		default:
			logger.Errorf("[%s] Unknown command: %q\n", appName, split[0])
		}
	}
}
