package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/chzyer/readline"
	"github.com/civilware/tela"
	"github.com/civilware/tela/logger"
	"github.com/civilware/tela/shards"
	"github.com/deroproject/derohe/cryptography/crypto"
	"github.com/deroproject/derohe/dvm"
	"github.com/deroproject/derohe/globals"
	"github.com/deroproject/derohe/rpc"
	"github.com/deroproject/derohe/walletapi"
)

var altSet = [][]string{
	{"▼", "▲"},
	{"□", "■"},
	{"✘", "✔"},
}

type ratioDOC struct {
	tela.DOC
	LikesRatio float64
}

type shardKeys struct {
	pageSize []byte
	minLikes []byte
	exclude  []byte
	gnomon   struct {
		fastsync       []byte
		parallelBlocks []byte
	}
}

var keys shardKeys

const CHUNK_SIZE = 17500

const printDivider = "------------------------------"

// Initialize TELA-CLI
func init() {
	parseHelpFlag() // Parse help argument
	logger.ASCIIPrint(false)
	fmt.Printf("TELA: Decentralized Web Standard  v%s\n", tela.GetVersion().String())
	fmt.Print("Initializing...")
	keys.pageSize = []byte("tela-cli.page-size")
	keys.minLikes = []byte("tela-cli.search-min-likes")
	keys.exclude = []byte("tela-cli.search-exclusions")
	keys.gnomon.fastsync = []byte("tela-cli.gnomon.fastsync")
	keys.gnomon.parallelBlocks = []byte("tela-cli.gnomon.parallel-blocks")
	done := make(chan struct{})
	// Initialize walletapi lookup table
	go func() {
		if os.Getenv("USE_BIG_TABLE") != "" {
			walletapi.Initialize_LookupTable(1, 1<<24)
		} else {
			walletapi.Initialize_LookupTable(1, 1<<21)
		}

		done <- struct{}{}
	}()

	for {
		select {
		case <-done:
			fmt.Print("\n")
			return
		case <-time.After(time.Second):
			fmt.Print(".")
		}
	}
}

// Print start usage if -h, --h, -help or --help flag is received
func parseHelpFlag() {
	for i, a := range os.Args {
		if i == 0 {
			continue
		}

		trimmed := strings.Trim(a, "elp")
		if trimmed == "-h" || trimmed == "--h" {
			fmt.Println(argsLine)
			os.Exit(0)
		}
	}
}

// Parse and return start flags
func (t *tela_cli) parseFlags() (arguments map[string]string) {
	var networkIsSet bool

	// Parse start flags
	arguments = map[string]string{}
	for i, a := range os.Args {
		if i == 0 {
			continue
		}

		flag, arg, _ := strings.Cut(a, "=")
		arguments[flag] = strings.TrimSpace(arg)
		switch flag {
		case "--testnet", "--simulator":
			networkIsSet = true
			globals.Arguments[flag] = true
			logger.Printf("[%s] %s enabled\n", appName, flag)
			if flag == "--simulator" {
				globals.Arguments["--testnet"] = true
			}
		case "--debug":
			globals.Arguments["--debug"] = true
		}
	}

	if _, ok := arguments["--mainnet"]; ok {
		networkIsSet = true
		globals.Arguments["--testnet"] = false
		globals.Arguments["--simulator"] = false
		logger.Printf("[%s] --mainnet enabled\n", appName)
	}

	if arg, ok := arguments["--db-type"]; ok {
		if arg != "" {
			if err := shards.SetDBType(arg); err != nil {
				logger.Warnf("[%s] %s: %s\n", appName, "--db-type", err)
			} else {
				logger.Printf("[%s] %s=%s\n", appName, "--db-type", shards.GetDBType())
			}
		}
	}

	if arg, ok := arguments["--fastsync"]; ok && arg != "" {
		b, err := strconv.ParseBool(arg)
		if err != nil {
			logger.Warnf("[%s] %s: %s\n", appName, "--fastsync", err)
		} else {
			logger.Printf("[%s] %s=%t\n", appName, "--fastsync", b)
			err = shards.StoreSettingsValue(keys.gnomon.fastsync, []byte(arg))
			if err != nil {
				logger.Debugf("[%s] Storing fastsync: %s\n", appName, err)
			}

			gnomon.fastsync = b
		}
	} else {
		fastsync, err := shards.GetSettingsValue(keys.gnomon.fastsync)
		if err != nil {
			logger.Debugf("[%s] Getting fastsync: %s\n", appName, err)
		} else {
			if b, err := strconv.ParseBool(string(fastsync)); err != nil {
				logger.Debugf("[%s] Setting fastsync: %s\n", appName, err)
			} else {
				gnomon.fastsync = b
			}
		}
	}

	if arg, ok := arguments["--num-parallel-blocks"]; ok && arg != "" {
		if i, err := strconv.Atoi(arg); err != nil {
			logger.Warnf("[%s] %s: %s\n", appName, "--num-parallel-blocks", err)
		} else if i > 0 {
			if i > maxParallelBlocks {
				logger.Printf("[%s] Setting number of parallel blocks to maximum of %d\n", appName, maxParallelBlocks)
				i = maxParallelBlocks
			}

			logger.Printf("[%s] %s=%d\n", appName, "--num-parallel-blocks", i)
			err = shards.StoreSettingsValue(keys.gnomon.parallelBlocks, []byte(fmt.Sprintf("%d", i)))
			if err != nil {
				logger.Debugf("[%s] Storing number of parallel blocks: %s\n", appName, err)
			}

			gnomon.parallelBlocks = i
		}
	} else {
		parallelBlocks, err := shards.GetSettingsValue(keys.gnomon.parallelBlocks)
		if err != nil {
			logger.Debugf("[%s] Getting number of parallel blocks: %s\n", appName, err)
		} else {
			if i, err := strconv.Atoi(string(parallelBlocks)); err != nil {
				logger.Debugf("[%s] Setting number of parallel blocks: %s\n", appName, err)
			} else {
				gnomon.parallelBlocks = i
			}
		}
	}

	if !networkIsSet {
		network, err := shards.GetNetwork()
		if err != nil {
			logger.Debugf("[%s] Getting network: %s\n", appName, err)
		}

		switch network {
		case shards.Value.Network.Testnet():
			globals.Arguments["--testnet"] = true
			globals.Arguments["--simulator"] = false
		case shards.Value.Network.Simulator():
			globals.Arguments["--testnet"] = true
			globals.Arguments["--simulator"] = true
		default:
			globals.Arguments["--testnet"] = false
			globals.Arguments["--simulator"] = false
		}
	}

	endpoint, err := shards.GetEndpoint()
	if err != nil {
		logger.Debugf("[%s] Getting endpoint: %s\n", appName, err)
	} else {
		t.endpoint = endpoint
	}

	pageSize, err := shards.GetSettingsValue(keys.pageSize)
	if err != nil {
		logger.Debugf("[%s] Getting page size: %s\n", appName, err)
	} else {
		if u, err := strconv.ParseUint(string(pageSize), 10, 64); err != nil {
			logger.Debugf("[%s] Setting page size: %s\n", appName, err)
		} else {
			t.pageSize = int(u)
		}
	}

	minLikes, err := shards.GetSettingsValue(keys.minLikes)
	if err != nil {
		logger.Debugf("[%s] Getting minimum likes: %s\n", appName, err)
	} else {
		if f, err := strconv.ParseFloat(string(minLikes), 64); err != nil {
			logger.Debugf("[%s] Setting minimum likes: %s\n", appName, err)
		} else {
			t.minLikes = f
		}
	}

	searchExclusions, err := shards.GetSettingsValue(keys.exclude)
	if err != nil {
		logger.Debugf("[%s] Getting search exclusions: %s\n", appName, err)
	} else {
		err = json.Unmarshal(searchExclusions, &t.exclusions)
		if err != nil {
			logger.Debugf("[%s] Setting search exclusions: %s\n", appName, err)
		}
	}

	return
}

// Filter out specific readline inputs from input processing
func filterInput(r rune) (rune, bool) {
	switch r {
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}

// Read errors for close from readline
func readError(err error) bool {
	switch err.Error() {
	case "Interrupt":
		return true
	default:
		return false
	}
}

// Return network command information if Connect error is due to network settings
func networkError(err error) error {
	if strings.Contains(err.Error(), "Mainnet/TestNet") {
		err = fmt.Errorf("invalid network settings, run command %q to change network settings", "endpoint <network>")
	}

	return err
}

// Get current network info, mainnet/testnet/simulator
func getNetworkInfo() (network string) {
	network = shards.Value.Network.Mainnet()
	if !globals.IsMainnet() {
		network = shards.Value.Network.Testnet()
		if b, ok := globals.Arguments["--simulator"].(bool); ok && b {
			network = shards.Value.Network.Simulator()
		}
	}

	return
}

// No options for auto completer
func completerNothing() (options []readline.PrefixCompleterInterface) {
	return
}

// List of current search exclusions for auto completer
func (t *tela_cli) completerSearchExclusions() (options []readline.PrefixCompleterInterface) {
	for _, filter := range t.exclusions {
		options = append(options, readline.PcItem(filter))
	}

	return
}

// List of DERO networks for auto completer
func completerNetworks(ep bool) (options []readline.PrefixCompleterInterface) {
	options = append(options, readline.PcItem("mainnet"))
	options = append(options, readline.PcItem("testnet"))
	options = append(options, readline.PcItem("simulator"))
	if ep {
		options = append(options, readline.PcItem("remote"))
		options = append(options, readline.PcItem("close"))
	}

	return
}

// True/false options for auto completer
func completerTrueFalse() (options []readline.PrefixCompleterInterface) {
	options = append(options, readline.PcItem("false"))
	options = append(options, readline.PcItem("true"))

	return
}

// Yes/no options for auto completer
func completerYesNo() (options []readline.PrefixCompleterInterface) {
	options = append(options, readline.PcItem("n"))
	options = append(options, readline.PcItem("y"))

	return
}

// List of TELA docType options for auto completer
func completerDocType() (options []readline.PrefixCompleterInterface) {
	options = append(options, readline.PcItem("tela-static-1"))
	options = append(options, readline.PcItem("tela-html-1"))
	options = append(options, readline.PcItem("tela-css-1"))
	options = append(options, readline.PcItem("tela-js-1"))
	options = append(options, readline.PcItem("tela-md-1"))

	return
}

// List of TELA-MOD tags for auto completer
func completerMODs(prefix string) (options []readline.PrefixCompleterInterface) {
	for _, m := range tela.Mods.GetAllMods() {
		if prefix != "" {
			if !strings.HasPrefix(m.Tag, prefix) {
				continue
			}
		}

		options = append(options, readline.PcItem(m.Tag))
	}

	return
}

// List of TELA-MODClass tags for auto completer
func completerMODClasses() (options []readline.PrefixCompleterInterface) {
	for _, c := range tela.Mods.GetAllClasses() {
		options = append(options, readline.PcItem(c.Tag))
	}

	return
}

// List of smart contract function names for auto completer
func completerSCFunctionNames(isOwner bool, sc dvm.SmartContract) (options []readline.PrefixCompleterInterface) {
	var names []string
	for name := range sc.Functions {
		// Functions that don't need to be offered as options
		if strings.ToLower(name) == name || name == "InitializePrivate" || name == "Initialize" || name == "UpdateCode" || name == "Rate" ||
			(!isOwner && strings.HasPrefix(name, "Withdraw")) || (!isOwner && name == "TransferOwnership") || (!isOwner && name == "DeleteVar") {
			continue
		}

		names = append(names, name)
	}

	sort.Strings(names)

	for _, n := range names {
		options = append(options, readline.PcItem(n))
	}

	return
}

// List files in directory for auto completer
func completerFiles(dir string, pc ...readline.PrefixCompleterInterface) *readline.PrefixCompleter {
	completer := func(line string) (names []string) {
		filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}

			add := path
			if info.IsDir() {
				add = add + string(filepath.Separator)
			}

			names = append(names, add)

			return nil
		})

		return
	}

	return readline.PcItemDynamic(completer, pc...)
}

// List active TELA servers for auto completer
func (t *tela_cli) completerServers() func(string) (names []string) {
	return func(string) (names []string) {
		telas, _, _ := t.getServerInfo()
		for _, t := range telas {
			names = append(names, t.Name)
		}

		return
	}
}

// Connect to walletapi endpoint, storing endpoint and network in preferences if successful
func (t *tela_cli) connectEndpoint() (err error) {
	globals.InitNetwork()
	if err = walletapi.Connect(t.endpoint); err == nil {
		if errr := shards.StoreEndpoint(t.endpoint); errr != nil {
			logger.Debugf("[%s] Storing connect endpoint: %s\n", appName, errr)
		}

		network := getNetworkInfo()
		if errr := shards.StoreNetwork(network); errr != nil {
			logger.Debugf("[%s] Storing connect network: %s\n", appName, errr)
		}
	}

	return
}

// Open wallet file with password
func (t *tela_cli) openWallet(file, password string) (err error) {
	if _, err = os.Stat(file); os.IsNotExist(err) {
		err = fmt.Errorf("wallet file %s does not exist", file)
		return
	}

	for password == "" {
		var line []byte
		line, err = t.readWithPasswordPrompt(fmt.Sprintf("Enter %s password", filepath.Base(file)))
		if err != nil {
			return
		}

		password = string(line)
	}

	t.wallet.disk, err = walletapi.Open_Encrypted_Wallet(file, password)
	if err != nil {
		return
	}

	t.wallet.name = file

	// Connect wallet
	t.wallet.disk.SetNetwork(globals.IsMainnet())
	t.wallet.disk.SetOnlineMode()
	if !walletapi.Connected {
		err = walletapi.Connect(t.endpoint)
		if err != nil {
			walletapi.Daemon_Endpoint_Active = ""
			t.closeWallet()
		}
	}

	return
}

// Close the wallet if open
func (t *tela_cli) closeWallet() {
	if t.wallet.disk != nil {
		t.wallet.disk.Close_Encrypted_Wallet()
		logger.Printf("[%s] Closed wallet %s\n", appName, t.wallet.name)
		t.wallet.disk = nil
		t.wallet.name = ""
	}
}

// Set the prompt text
func (t *tela_cli) setPrompt(text string) {
	t.rli.SetPrompt(t.formatPrompt(text))
	t.rli.Refresh()
}

func onlineStatus(b bool) string {
	if b {
		return fmt.Sprintf("%s%s%s", logger.Color.Green(), altSet[0][1], logger.Color.End())
	}

	return fmt.Sprintf("%s%s%s", logger.Color.Red(), altSet[0][0], logger.Color.End())
}

// Format prompt text with data
func (t *tela_cli) formatPrompt(text string) string {
	count := 0
	_, servers, _ := t.getServerInfo()
	if servers > 0 {
		count = servers
	}

	status := logger.Color.Red()
	height := int64(0)
	if t.wallet.disk != nil && walletapi.IsDaemonOnline() {
		status = logger.Color.Green()
		height = walletapi.Get_Daemon_Height()
	}

	var d, g bool
	if walletapi.IsDaemonOnline() {
		d = true
	}

	if gnomon.Indexer != nil {
		g = true
	}

	daemonStatus := fmt.Sprintf("[%sD%s:%s]", logger.Color.Grey(), logger.Color.End(), onlineStatus(d))
	gnomonStatus := fmt.Sprintf("[%sG%s:%s]", logger.Color.Grey(), logger.Color.End(), onlineStatus(g))

	braille := " ⠞⠑⠇⠁ " // TELA, might not work on some os
	if t.os == "windows" {
		braille = " •••• "
	}

	wHeight := fmt.Sprintf("[%sW%s:%s%d%s]", logger.Color.Grey(), logger.Color.End(), status, height, logger.Color.End())
	sCount := fmt.Sprintf("[%s%d%s/%s%d%s]", logger.Color.Grey(), count, logger.Color.End(), logger.Color.Grey(), tela.MaxServers()+1, logger.Color.End())
	prompt := fmt.Sprintf("%s%s%s%s: %s %s %s %s", braille, logger.Color.Cyan(), "TELA-CLI", logger.Color.End(), daemonStatus, gnomonStatus, wHeight, sCount)
	if text != "" {
		prompt = fmt.Sprintf("%s %s", prompt, text)
	}

	return fmt.Sprintf("%s %s » ", logger.Timestamp(), prompt)
}

// Check if the daemon connection is active and handle accordingly when disconnected
func checkDaemonConnection() {
	if walletapi.IsDaemonOnline() {
		var result string
		if err := walletapi.GetRPCClient().Call("DERO.Ping", nil, &result); err != nil {
			stopGnomon()
			walletapi.Connect(" ")
		}
	}
}

// Handle error signals inside read funcs before returning the error
func (t *tela_cli) lineError(err error) error {
	if err != nil {
		if err.Error() == "Interrupt" {
			t.cancel()
			t.shutdown()
			return err
		} else if err.Error() == "EOF" {
			return err
		}
	}

	return nil
}

// Read password input
func (t *tela_cli) readWithPasswordPrompt(text string) (line []byte, err error) {
	t.Lock()
	defer t.Unlock()

	passCfg := t.rli.GenPasswordConfig()
	passCfg.SetListener(func(line []rune, pos int, key rune) (newLine []rune, newPos int, ok bool) {
		t.setPrompt(fmt.Sprintf("%s (%d)", text, len(line)))
		return nil, 0, false
	})

	for {
		line, err = t.rli.ReadPasswordWithConfig(passCfg)
		if err != nil {
			if t.lineError(err) != nil {
				return
			} else {
				logger.Errorf("[%s] %s\n", appName, err)
				continue
			}
		}

		return
	}
}

// Read line input
func (t *tela_cli) readLine(text, buffer string) (line string, err error) {
	completer := readline.NewPrefixCompleter(completerNothing()...)

	return t.readLineWithCompleter(text, buffer, completer)
}

// Read line input with auto completer specified
func (t *tela_cli) readLineWithCompleter(text, buffer string, completer readline.AutoCompleter) (line string, err error) {
	resetCompleter := t.rli.Config.AutoComplete
	t.rli.Config.AutoComplete = completer

	t.Lock()
	defer func() {
		t.rli.Config.AutoComplete = resetCompleter
		t.Unlock()
	}()

	t.setPrompt(text)
	if buffer != "" {
		t.rli.Operation.SetBuffer(buffer)
	}

	for {
		line, err = t.rli.Readline()
		if err != nil {
			if t.lineError(err) != nil {
				return
			} else {
				logger.Errorf("[%s] %s\n", appName, err)
				continue
			}
		}

		return strings.TrimSpace(string(line)), nil
	}
}

// Read uint64 input
func (t *tela_cli) readUint64(text string) (u uint64, err error) {
	for {
		var line string
		line, err = t.readLine(text, "")
		if err != nil {
			return
		}

		u, err = strconv.ParseUint(line, 10, 64)
		if err != nil {
			logger.Errorf("[%s] %s\n", appName, err)
			continue
		}

		return
	}
}

// Read yes or no input
func (t *tela_cli) readYesNo(text string) (confirmed bool, err error) {
	completer := readline.NewPrefixCompleter(completerYesNo()...)

	for {
		var line string
		line, err = t.readLineWithCompleter(fmt.Sprintf("%s (y/n)", text), "", completer)
		if err != nil {
			return
		}

		switch strings.ToLower(line) {
		case "n":
			return
		case "y":
			confirmed = true
			return
		default:
			continue
		}
	}
}

// Read inputs for setting common headers
func (t *tela_cli) headersPrompt(text string, index *tela.INDEX) (headers map[tela.Header]string, err error) {
	headers = map[tela.Header]string{}

	var descrHdr, iconHdr, durl string
	if index != nil {
		if index.DescrHdr != "" {
			descrHdr = index.DescrHdr
		}

		if index.IconHdr != "" {
			iconHdr = index.IconHdr
		}

		if index.DURL != "" {
			durl = index.DURL
		}
	}

	prompt := fmt.Sprintf("Enter %s description", text)
	headers[tela.HEADER_DESCRIPTION], err = t.readLine(prompt, descrHdr)
	if err != nil {
		return
	}

	prompt = fmt.Sprintf("Enter %s icon", text)
	headers[tela.HEADER_ICON_URL], err = t.readLine(prompt, iconHdr)
	if err != nil {
		return
	}

	var dURL string
	for dURL == "" {
		prompt = fmt.Sprintf("Enter %s dURL", text)
		dURL, err = t.readLine(prompt, durl)
		if err != nil {
			return
		}

		headers[tela.HEADER_DURL] = dURL
		if headers[tela.HEADER_DURL] == "" {
			logger.Errorf("[%s] dURL is required\n", appName)
		}
	}

	return
}

// Read ringsize input
func (t *tela_cli) ringsizePrompt(text string) (ringsize uint64, err error) {
	prompt := fmt.Sprintf("Enter %s install ringsize", text)
	completer := readline.NewPrefixCompleter(readline.PcItem("2"), readline.PcItem("16"))

	for ringsize < 2 {
		var line string
		line, err = t.readLineWithCompleter(prompt, "", completer)
		if err != nil {
			return
		}

		ringsize, err = strconv.ParseUint(line, 10, 64)
		if err != nil {
			logger.Errorf("[%s] %s\n", appName, err)
			continue
		}

		if ringsize < 2 {
			ringsize = 2
			logger.Printf("[%s] Applying minimum ringsize of 2\n", appName)
		} else if ringsize > 128 {
			ringsize = 128
			logger.Printf("[%s] Applying maximum ringsize of 128\n", appName)
		}
	}

	return
}

// Read INDEX input for installer/updater
func (t *tela_cli) indexPrompt(nameHdr string, previousIndex *tela.INDEX) (index tela.INDEX, err error) {
	// Common headers
	headers, err := t.headersPrompt("INDEX", previousIndex)
	if err != nil {
		return
	}

	var line string
	var total uint64
	for total == 0 {
		line, err = t.readLine("How many total documents are embedded in this INDEX?", "")
		if err != nil {
			return
		}

		total, err = strconv.ParseUint(line, 10, 0)
		if err != nil {
			logger.Errorf("[%s] Total: %s\n", appName, err)
			continue
		}

		if total == 0 {
			logger.Errorf("[%s] Minimum of one DOC is required\n", appName)
		}
	}

	var scids, paths []string
	for i := 0; i < int(total); {
		var lastSCID string
		if previousIndex != nil && i < len(previousIndex.DOCs) {
			lastSCID = previousIndex.DOCs[i]
		}

		line, err = t.readLine(fmt.Sprintf("Enter DOC%d SCID", i+1), lastSCID)
		if err != nil {
			return
		}

		if len(line) != 64 {
			logger.Errorf("[%s] Invalid DOC SCID: %q\n", appName, line)
			continue
		}

		doc, err := tela.GetDOCInfo(line, t.endpoint)
		if err != nil {
			ind, errr := tela.GetINDEXInfo(line, t.endpoint)
			if errr != nil {
				logger.Errorf("[%s] GetDOCInfo: %s\n", appName, err)
				logger.Errorf("[%s] GetINDEXInfo: %s\n", appName, errr)
				continue
			}

			if i == 0 {
				logger.Errorf("[%s] INDEX %s cannot be used as DOC1\n", appName, ind.DURL)
				continue
			}

			if !strings.HasSuffix(ind.DURL, tela.TAG_LIBRARY) && !strings.Contains(ind.DURL, tela.TAG_DOC_SHARDS) {
				logger.Errorf("[%s] INDEX %s is not a library or shards\n", appName, ind.DURL)
				continue
			}

			// SCID is INDEX library, ensure its DOC paths do not exist already in this INDEX
			var indexErr string
			for _, d := range ind.DOCs {
				doc, err = tela.GetDOCInfo(d, t.endpoint)
				if err != nil {
					indexErr = fmt.Sprintf("Failed to validate %q from %s", ind.NameHdr, d)
					logger.Errorf("[%s] GetDOCInfo: %s\n", appName, err)
					break
				}

				filePath := filepath.Join(doc.SubDir, doc.NameHdr)
				for _, p := range paths {
					if p == filePath {
						indexErr = fmt.Sprintf("Import from %s INDEX already contains a DOC with path %q", d, filePath)
						break
					}
				}

				if indexErr != "" {
					break
				}

				logger.Printf("[%s] File: %s\n", appName, filePath)
				logger.Printf("[%s] Author: %s\n", appName, doc.Author)
				paths = append(paths, filePath)
			}

			if indexErr != "" {
				logger.Errorf("[%s] %s\n", appName, indexErr)
				continue
			}
		} else {
			// Ensure DOC path does not exists already in this INDEX
			var have bool
			filePath := filepath.Join(doc.SubDir, doc.NameHdr)
			for _, p := range paths {
				if p == filePath {
					have = true
					break
				}
			}

			if have {
				logger.Errorf("[%s] INDEX already contains a DOC with path %q\n", appName, filePath)
				continue
			}

			logger.Printf("[%s] File: %s\n", appName, filePath)
			logger.Printf("[%s] Author: %s\n", appName, doc.Author)
			paths = append(paths, filePath)
		}

		i++
		scids = append(scids, line)
	}

	scid := ""
	var version *tela.Version
	if previousIndex != nil {
		version = previousIndex.SCVersion
		scid = previousIndex.SCID
	}

	// Create TELA INDEX
	index = tela.INDEX{
		SCVersion: version,
		SCID:      scid,
		DURL:      headers[tela.HEADER_DURL],
		DOCs:      scids,
		Headers: tela.Headers{
			NameHdr:  nameHdr,
			DescrHdr: headers[tela.HEADER_DESCRIPTION],
			IconHdr:  headers[tela.HEADER_ICON_URL],
		},
	}

	return
}

// Read inputs for setting TELA-MODs
func (t *tela_cli) modsPrompt() (modTag string, err error) {
	// Handle any TELA-MODs to be added
	yes, err := t.readYesNo("Add SC TELA-MODs")
	if err != nil {
		return
	}

	if yes {
		yes, err = t.readYesNo("Add variable store MOD")
		if err != nil {
			return
		}

		var modTags []string
		allTelaMods := tela.Mods.GetAllMods()

		if yes {
			// Display MOD info
			for i, m := range allTelaMods {
				printMODInfo(m, false)
				if i == tela.Mods.Index()[0]-1 {
					break
				}
			}
			fmt.Println(printDivider)

			for len(modTags) < 1 {
				var line string
				completer := readline.NewPrefixCompleter(completerMODs("vs")...)
				line, err = t.readLineWithCompleter("Enter a variable store MOD tag", "", completer)
				if err != nil {
					return
				}

				// Bypass the loop
				if line == "none" || line == "" {
					break
				}

				_, err = tela.Mods.TagsAreValid(line)
				if err != nil {
					logger.Errorf("[%s] Invalid MOD tag: %s\n", appName, err)
					continue
				}

				if line != "" {
					modTags = append(modTags, line)
					logger.Printf("[%s] VS MOD: %q\n", appName, line)
					break
				}
			}
		}

		yes, err = t.readYesNo("Add transfer MODs")
		if err != nil {
			return
		}

		if yes {
			for i, m := range allTelaMods {
				if i < tela.Mods.Index()[0] {
					// Already did variable store mods
					continue
				}

				printMODInfo(m, false)
				if i == tela.Mods.Index()[1]-1 {
					break
				}
			}
			fmt.Println(printDivider)

			startLen := len(modTags)

			for len(modTags) <= startLen {
				var line string
				completer := readline.NewPrefixCompleter(completerMODs("tx")...)
				line, err = t.readLineWithCompleter("Enter transfers MOD tags", "", completer)
				if err != nil {
					return
				}

				// Bypass the loop
				if line == "none" || line == "" {
					break
				}

				// Accepts two entry formats, space separator or comma separator
				var split []string
				if strings.Contains(line, ",") {
					split = strings.Split(line, ",")
				} else {
					split = strings.Split(line, " ")
				}

				_, err = tela.Mods.TagsAreValid(tela.NewModTag(split))
				if err != nil {
					logger.Errorf("[%s] Invalid MOD tag: %s\n", appName, err)
					continue
				}

				for _, s := range split {
					tag := strings.TrimSpace(s)
					if tag != "" {
						modTags = append(modTags, tag)
						logger.Printf("[%s] TX MOD: %q\n", appName, tag)
					}
				}
			}
		}

		modTag = tela.NewModTag(modTags)
		_, err = tela.Mods.TagsAreValid(modTag)
		if err != nil {
			modTag = ""
			err = fmt.Errorf("invalid MOD tags: %s", err)
			return
		}

		if modTag != "" {
			logger.Printf("[%s] Adding MODs: %q\n", appName, modTag)
		}
	}

	return
}

// Prompt for the information needed to execute TELA-MOD smart contract functions
func (t *tela_cli) executeContractPrompt(scid, functionName string, sc dvm.SmartContract) (transfers []rpc.Transfer, args rpc.Arguments, err error) {
	function, ok := sc.Functions[functionName]
	if !ok {
		err = fmt.Errorf("function %q does not exist", functionName)
		return
	}

	_, destination := tela.GetDefaultNetworkAddress()
	for _, line := range function.Lines {
		if slices.Contains(line, "DEROVALUE") {
			var derovalue uint64
			derovalue, err = t.readUint64("Enter DEROVALUE")
			if err != nil {
				return
			}

			transfers = append(transfers, rpc.Transfer{Destination: destination, Amount: 0, Burn: derovalue})
			break
		} else if slices.Contains(line, "ASSETVALUE") {
			var assetSCID string
			assetSCID, err = t.readLine("Enter ASSET SCID", "")
			if err != nil {
				return
			}

			if len(assetSCID) != 64 {
				err = fmt.Errorf("invalid ASSET SCID: %q", assetSCID)
				return
			}

			var assetvalue uint64
			assetvalue, err = t.readUint64("Enter ASSETVALUE")
			if err != nil {
				return
			}

			transfers = append(transfers, rpc.Transfer{Destination: destination, SCID: crypto.HashHexToHash(assetSCID), Amount: 0, Burn: assetvalue})
			break
		}
	}

	// tela.Transfer would do this anyways but cli should display a transfer for the user instead of nil
	if len(transfers) < 1 {
		transfers = []rpc.Transfer{{Destination: destination, Amount: 0}}
	}

	args = rpc.Arguments{
		rpc.Argument{Name: "entrypoint", DataType: rpc.DataString, Value: functionName},
		rpc.Argument{Name: rpc.SCID, DataType: rpc.DataHash, Value: crypto.HashHexToHash(scid)},
		rpc.Argument{Name: rpc.SCACTION, DataType: rpc.DataUint64, Value: uint64(rpc.SC_CALL)},
	}

	for _, p := range function.Params {
		switch p.Type {
		case dvm.String:
			var sParam string
			sParam, err = t.readLine(fmt.Sprintf("Enter %q param string value", p.Name), "")
			if err != nil {
				return
			}

			args = append(args, rpc.Argument{Name: p.Name, DataType: rpc.DataString, Value: sParam})
		case dvm.Uint64:
			var uParam uint64
			uParam, err = t.readUint64(fmt.Sprintf("Enter %q param uint64 value", p.Name))
			if err != nil {
				return
			}

			args = append(args, rpc.Argument{Name: p.Name, DataType: rpc.DataUint64, Value: uParam})
		default:
			err = fmt.Errorf("found unknown param type: %s %v", p.Name, p.Type)
			return
		}
	}

	return
}

// Shut down the additional server
func (t *tela_cli) shutdownLocalServer() {
	if t.local.server != nil {
		err := t.local.server.Shutdown(context.Background())
		if err != nil {
			logger.Errorf("[%s] Shutdown: %s\n", appName, err)
		}
		t.local.server = nil
	}
}

// Additional server for cli to serve files from local directory for TELA tests and development
func (t *tela_cli) serveLocal(path string) (err error) {
	if t.local.server != nil {
		err = fmt.Errorf("server already exists")
		return
	}

	server, found := tela.FindOpenPort()
	if !found {
		err = fmt.Errorf("could not find open port to serve %s", path)
		return
	}

	// Set the directory to serve files from
	fs := http.FileServer(http.Dir(path))

	// Handle all requests to server
	server.Handler = fs

	entrypoint := "index.html"

	// Serve on this address:port
	link := fmt.Sprintf("http://localhost%s/%s", server.Addr, entrypoint)

	t.local.server = server
	t.local.Address = server.Addr

	// Serve content
	go func() {
		logger.Printf("[%s] Serving %s at %s\n", appName, path, link)
		err := server.ListenAndServe()
		if err != nil {
			if err == http.ErrServerClosed {
				logger.Printf("[%s] Closed local %s %s\n", appName, server.Addr, path)
			} else {
				logger.Errorf("[%s] Listen %s %s %s\n", appName, server.Addr, path, err)
			}
		}
	}()

	if t.openInBrowser {
		if err = t.openBrowser(link); err != nil {
			return
		}
	}

	return
}

// Open link in default browser for common OS
func (t *tela_cli) openBrowser(link string) (err error) {
	switch t.os {
	case "linux", "freebsd", "netbsd", "openbsd":
		err = exec.Command("xdg-open", link).Run()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", link).Run()
	case "darwin":
		err = exec.Command("open", link).Run()
	default:
		err = fmt.Errorf("could not open default browser for %s", t.os)
	}

	return
}

// Get TELA and local server info
func (t *tela_cli) getServerInfo() (telas []tela.ServerInfo, total int, local bool) {
	telas = tela.GetServerInfo()
	local = t.local.server != nil
	total = len(telas)
	if local {
		total++
	}

	return
}

// Get TELA-CLI app info
func (t *tela_cli) getCLIInfo() (allInfo []string) {
	allInfo = append(allInfo, printDivider)

	allInfo = append(allInfo, fmt.Sprintf("OS: %s", t.os))
	walletName := "offline"
	if t.wallet.disk != nil {
		walletName = t.wallet.name
	}

	allInfo = append(allInfo, fmt.Sprintf("DOC: v%s", tela.GetLatestContractVersion(true).String()))
	allInfo = append(allInfo, fmt.Sprintf("INDEX: v%s", tela.GetLatestContractVersion(false).String()))

	allInfo = append(allInfo, fmt.Sprintf("Wallet: %s", filepath.Base(walletName)))
	allInfo = append(allInfo, fmt.Sprintf("Network: %s", strings.ToLower(getNetworkInfo())))

	if walletapi.IsDaemonOnline() {
		allInfo = append(allInfo, fmt.Sprintf("Height: %d", walletapi.Get_Daemon_Height()))
	}
	allInfo = append(allInfo, fmt.Sprintf("Endpoint: %s", t.endpoint))

	indexer := "sleeping"
	if gnomon.Indexer != nil {
		indexer = "active"
	}
	allInfo = append(allInfo, fmt.Sprintf("Gnomon: %s", indexer))
	allInfo = append(allInfo, fmt.Sprintf("DB Type: %s", shards.GetDBType()))
	allInfo = append(allInfo, fmt.Sprintf("Fastsync: %t", gnomon.fastsync))
	if gnomon.Indexer != nil {
		allInfo = append(allInfo, fmt.Sprintf("Indexed SCs: %d", len(gnomon.GetAllOwnersAndSCIDs())))
		allInfo = append(allInfo, fmt.Sprintf("Indexed Height: %d", gnomon.Indexer.LastIndexedHeight))
		allInfo = append(allInfo, fmt.Sprintf("Parallel Blocks: %d", gnomon.parallelBlocks))
	}

	allInfo = append(allInfo, fmt.Sprintf("Page size: %d", t.pageSize))
	allInfo = append(allInfo, fmt.Sprintf("Port start: %d", tela.PortStart()))
	allInfo = append(allInfo, fmt.Sprintf("Available MODs: %d", len(tela.Mods.GetAllMods())))

	_, servers, local := t.getServerInfo()
	allInfo = append(allInfo, fmt.Sprintf("Active servers: %d/%d", servers, tela.MaxServers()+1))
	if servers > 0 {
		allInfo = append(allInfo, fmt.Sprintf("Local server running: %t", local))
	}

	allInfo = append(allInfo, fmt.Sprintf("Search minimum likes: %.0f%%", t.minLikes))
	allInfo = append(allInfo, fmt.Sprintf("Open content in browser: %t", t.openInBrowser))
	allInfo = append(allInfo, fmt.Sprintf("Updated content allowed: %t", tela.UpdatesAllowed()))

	allInfo = append(allInfo, printDivider)

	return
}

// Green yellow or red string returned from ratio
func colorLikesRatio(ratio float64) string {
	color := logger.Color.Red()
	if ratio > 65 {
		color = logger.Color.Green()
	} else if ratio > 32 {
		color = logger.Color.Yellow()
	}

	return fmt.Sprintf("%s%0.f%%%s", color, ratio, logger.Color.End())
}

// Get the ratio of likes for a TELA SCID, if required and ratio < minLines an error will be returned
func (t *tela_cli) getLikesRatio(scid string, required bool) (dURL string, ratio float64, err error) {
	if gnomon.Indexer == nil {
		err = fmt.Errorf("gnomon is not online")
		return
	}

	d, _ := gnomon.GetSCIDValuesByKey(scid, "dURL")
	if d == nil {
		err = fmt.Errorf("could not get %s dURL", scid)
		return
	}

	dURL = d[0]
	for _, filter := range t.exclusions {
		if strings.Contains(dURL, filter) {
			err = fmt.Errorf("found dURL exclusion filter %q in %s", filter, scid)
			return
		}
	}

	_, up := gnomon.GetSCIDValuesByKey(scid, "likes")
	if up == nil {
		err = fmt.Errorf("could not get %s likes", scid)
		return
	}

	_, down := gnomon.GetSCIDValuesByKey(scid, "dislikes")
	if down == nil {
		err = fmt.Errorf("could not get %s dislikes", scid)
		return
	}

	total := float64(up[0] + down[0])
	if total == 0 {
		ratio = 50
	} else {
		ratio = (float64(up[0]) / total) * 100
	}

	if required && ratio < t.minLikes {
		err = fmt.Errorf("%s is below min rating setting", scid)
	}

	return
}

// Page results
func (t *tela_cli) paging(resultLines [][]string) (err error) {
	if len(resultLines) < 1 {
		logger.Printf("[%s] No results found\n", appName)
		return
	}

	sort.Slice(resultLines, func(i, j int) bool { return resultLines[i][0] < resultLines[j][0] })

	display := t.pageSize - 1

	resultLen := len(resultLines)

	isPaged := false
	if resultLen > t.pageSize {
		isPaged = true
		logger.Printf("[%s] Showing %d of %d results\n", appName, t.pageSize, resultLen)
	}

	fmt.Println(searchDivider)
	for printed, lines := range resultLines {
		for i := range lines {
			fmt.Println(lines[i])
		}

		fmt.Println(searchDivider)

		end := printed == resultLen-1
		if printed >= display && !end {
			var yes bool
			yes, err = t.readYesNo(fmt.Sprintf("Show more results? (%d)", (resultLen-1)-printed))
			if err != nil {
				return
			}

			if !yes {
				break
			}

			display = display + t.pageSize
			fmt.Println(searchDivider)
		}

		if isPaged && end {
			logger.Printf("[%s] End of results\n", appName)
		}
	}

	return
}

// Get the type of TELA contract from scid
func getSCType(scid string) (scType string) {
	scType = "?"
	ind, _ := gnomon.GetSCIDValuesByKey(scid, "DOC1")
	if ind != nil {
		scType = "TELA-INDEX-1"
	} else {
		docType, _ := gnomon.GetSCIDValuesByKey(scid, "docType")
		if docType != nil {
			if tela.IsAcceptedLanguage(docType[0]) {
				scType = docType[0]
			}
		}
	}

	return
}

// Get the type of TELA DOC from scid
func getDOCType(scid string) (docType []string) {
	docType, _ = gnomon.GetSCIDValuesByKey(scid, "docType")
	if docType == nil {
		return nil
	}

	if !tela.IsAcceptedLanguage(docType[0]) {
		return nil
	}

	return
}

// Parse generic info (non tela) from search queries and return resulting lines to print
func parseBASInfo(scid, owner string) (lines []string) {
	nameHdr, _ := gnomon.GetSCIDValuesByKey(scid, "nameHdr")
	if nameHdr == nil {
		nameHdr = append(nameHdr, "?")
	}

	lines = append(lines, fmt.Sprintf("%sNon-TELA%s %-62s Author: %s", logger.Color.Yellow(), logger.Color.End(), "", owner))
	lines = append(lines, fmt.Sprintf("SCID: %s  Name: %-35s", scid, nameHdr[0]))

	return
}

// Parse INDEX info from search queries and return resulting lines to print
func parseINDEXInfo(scid, owner, dURL, modTag string, ratio float64) (lines []string) {
	nameHdr, _ := gnomon.GetSCIDValuesByKey(scid, "nameHdr")
	if nameHdr == nil {
		nameHdr = append(nameHdr, "?")
	}

	lines = append(lines, fmt.Sprintf("%sdURL:%s %-65s Author: %s", logger.Color.Grey(), logger.Color.End(), dURL, owner))
	lines = append(lines, fmt.Sprintf("SCID: %s  Type: %-16s  Name: %-33s  Likes: %s", scid, "TELA-INDEX-1", nameHdr[0], colorLikesRatio(ratio)))
	if modTag != "" {
		lines = append(lines, fmt.Sprintf("MODs: %s", modTag))
	}

	return
}

// Search for INDEX info from Gnomon DB filtering for wallet owner
func (t *tela_cli) searchINDEXInfo(all map[string]string, owned bool) (resultLines [][]string) {
	if owned && t.wallet.disk == nil {
		return
	}

	for sc := range all {
		owner := gnomon.GetOwnerAddress(sc)
		if owned && owner != t.wallet.disk.GetAddress().String() {
			continue
		}

		dURL, likesRatio, err := t.getLikesRatio(sc, !owned)
		if err != nil {
			continue
		}

		var modTag string
		mods, _ := gnomon.GetSCIDValuesByKey(sc, "mods")
		if mods != nil {
			modTag = mods[0]
		}

		ind, _ := gnomon.GetSCIDValuesByKey(sc, "DOC1")
		if ind != nil {
			resultLines = append(resultLines, parseINDEXInfo(sc, owner, dURL, modTag, likesRatio))
		}
	}

	return
}

// Parse DOC info from search queries and return resulting lines to print
func parseDOCInfo(scid, owner, docType, dURL string, ratio float64) (lines []string) {
	nameHdr, _ := gnomon.GetSCIDValuesByKey(scid, "nameHdr")
	if nameHdr == nil {
		nameHdr = append(nameHdr, "?")
	}

	lines = append(lines, fmt.Sprintf("%sdURL:%s %-65s Author: %s", logger.Color.Grey(), logger.Color.End(), dURL, owner))
	lines = append(lines, fmt.Sprintf("SCID: %s  DocType: %-13s  Name: %-33s  Likes: %s", scid, docType, nameHdr[0], colorLikesRatio(ratio)))

	return
}

// Search for DOC info from Gnomon DB filtering for wallet owner and docType
func (t *tela_cli) searchDOCInfo(all map[string]string, owned bool, args ...string) (resultLines [][]string) {
	if owned && t.wallet.disk == nil {
		return
	}

	if len(args) > 1 {
		// docType search
		for sc := range all {
			owner := gnomon.GetOwnerAddress(sc)
			if owned && owner != t.wallet.disk.GetAddress().String() {
				continue
			}

			dURL, likesRatio, err := t.getLikesRatio(sc, !owned)
			if err != nil {
				continue
			}

			docType := getDOCType(sc)
			if docType != nil {
				if strings.ToLower(docType[0]) == args[1] {
					resultLines = append(resultLines, parseDOCInfo(sc, owner, docType[0], dURL, likesRatio))
				}
			}
		}
	} else {
		// search all DOCs
		for sc := range all {
			owner := gnomon.GetOwnerAddress(sc)
			if owned && owner != t.wallet.disk.GetAddress().String() {
				continue
			}

			dURL, likesRatio, err := t.getLikesRatio(sc, !owned)
			if err != nil {
				continue
			}

			docType := getDOCType(sc)
			if docType != nil {
				resultLines = append(resultLines, parseDOCInfo(sc, owner, docType[0], dURL, likesRatio))
			}
		}
	}

	return
}

// Create a DOC from scid stored in Gnomon DB
func createDOC(scid, dURL string) (doc tela.DOC) {
	docType, _ := gnomon.GetSCIDValuesByKey(scid, "docType")
	if docType == nil {
		docType = append(docType, "?")
	}

	subDir, _ := gnomon.GetSCIDValuesByKey(scid, "subDir")
	if subDir == nil {
		subDir = append(subDir, "")
	}

	nameHdr, _ := gnomon.GetSCIDValuesByKey(scid, "nameHdr")
	if nameHdr == nil {
		nameHdr = append(nameHdr, "?")
	}

	return tela.DOC{
		DocType: docType[0],
		SubDir:  subDir[0],
		SCID:    scid,
		DURL:    dURL,
		Headers: tela.Headers{
			NameHdr: nameHdr[0],
		},
	}
}

// Get TELA libraries from Gnomon DB
func (t *tela_cli) getLibraries() (libKeys []tela.Library, libMap map[tela.Library][]ratioDOC) {
	all := gnomon.GetAllOwnersAndSCIDs()
	if len(all) < 1 {
		return
	}

	libMap = map[tela.Library][]ratioDOC{}
	for sc := range all {
		dURL, likesRatio, err := t.getLikesRatio(sc, true)
		if err != nil {
			continue
		}

		owner := gnomon.GetOwnerAddress(sc)

		if strings.HasSuffix(dURL, tela.TAG_LIBRARY) {
			tLib := tela.Library{DURL: dURL, Author: owner}
			if libType := getSCType(sc); libType == "TELA-INDEX-1" {
				// INDEX tagged as lib, parse code for DOCs that make up this library
				code, _ := gnomon.GetSCIDValuesByKey(sc, "C")
				if code == nil {
					continue
				}

				scids, err := tela.ParseINDEXForDOCs(code[0])
				if err != nil {
					continue
				}

				// INDEX libraries are grouped by scid
				tLib.SCID = sc

				for _, scid := range scids {
					doc := createDOC(scid, dURL)
					libMap[tLib] = append(libMap[tLib], ratioDOC{DOC: doc, LikesRatio: likesRatio})
				}
			} else {
				// Single DOCs tagged as lib will be grouped by dURL and owner
				doc := createDOC(sc, dURL)
				libMap[tLib] = append(libMap[tLib], ratioDOC{DOC: doc, LikesRatio: likesRatio})
			}

			if !slices.Contains(libKeys, tLib) {
				libKeys = append(libKeys, tLib)
			}

			// TODO, sort these so they can be printed with hierarchy
			sort.Slice(libMap[tLib], func(i, j int) bool { return libMap[tLib][i].NameHdr < libMap[tLib][j].NameHdr })
		}
	}

	sort.Slice(libKeys, func(i, j int) bool { return libKeys[i].DURL < libKeys[j].DURL })

	return
}

// Parse library info from search queries and return resulting lines to print
func parseLibraryInfo(lib tela.Library, docs map[tela.Library][]ratioDOC) (lines []string) {
	identifier := fmt.Sprintf("Author: %s", lib.Author)
	if lib.SCID != "" {
		identifier = fmt.Sprintf("SCID: %s", lib.SCID)
	}

	lines = append(lines, fmt.Sprintf("%sdURL:%s %-65s %s", logger.Color.Grey(), logger.Color.End(), lib.DURL, identifier))
	for _, doc := range docs[lib] {
		lines = append(lines, fmt.Sprintf("SCID: %s  DocType: %-13s  Name: %-33s  Likes: %s", doc.SCID, doc.DocType, doc.NameHdr, colorLikesRatio(doc.LikesRatio)))
	}

	return
}

// Parse info for search queries switching between DOC, INDEX or generic and return resulting lines to print
func parseSearchQuery(scid, owner, dURL string, ratio float64) (lines []string) {
	scType := getSCType(scid)
	if scType == "TELA-INDEX-1" {
		var modTag string
		mods, _ := gnomon.GetSCIDValuesByKey(scid, "mods")
		if mods != nil {
			modTag = mods[0]
		}

		return parseINDEXInfo(scid, owner, dURL, modTag, ratio)
	} else if strings.Contains(scType, "TELA-") {
		return parseDOCInfo(scid, owner, scType, dURL, ratio)
	} else {
		return parseBASInfo(scid, owner)
	}
}

// Prints info for the given TELA-MOD
func printMODInfo(mod tela.MOD, printCode bool) {
	fmt.Println(printDivider)
	fmt.Printf("%sClass:%s %s\n", logger.Color.Grey(), logger.Color.End(), tela.Mods.GetClass(mod.Tag).Name)
	fmt.Printf("%sName:%s %s\n", logger.Color.Grey(), logger.Color.End(), mod.Name)
	fmt.Printf("%sTag:%s %s\n", logger.Color.Grey(), logger.Color.End(), mod.Tag)
	fmt.Printf("%sDescription:%s %s\n", logger.Color.Grey(), logger.Color.End(), mod.Description)
	if printCode {
		fmt.Println()
		fmt.Println(mod.FunctionCode())
	}
}

// Prints info for the given TELA-MODClass
func printMODClassInfo(class tela.MODClass) {
	fmt.Println(printDivider)
	fmt.Printf("%sClass Name:%s %s\n", logger.Color.Grey(), logger.Color.End(), class.Name)
	fmt.Printf("%sTag:%s %s\n", logger.Color.Grey(), logger.Color.End(), class.Tag)
	for i, r := range class.Rules {
		if i == 0 {
			fmt.Printf("%sRules:%s\n", logger.Color.Grey(), logger.Color.End())
		}
		fmt.Printf("#%d: %s - %s\n", i+1, r.Name, r.Description)
	}
}

// Print file information
func printFileInfo(info fs.FileInfo) {
	fmt.Printf("%sName:%s %s\n", logger.Color.Cyan(), logger.Color.End(), info.Name())
	fmt.Printf("%sSize:%s %0.2f KB\n", logger.Color.Cyan(), logger.Color.End(), float64(info.Size())/1000)
	fmt.Printf("%sDocType:%s %s\n", logger.Color.Cyan(), logger.Color.End(), tela.ParseDocType(info.Name()))
	fmt.Printf("%sModified:%s %s\n", logger.Color.Cyan(), logger.Color.End(), info.ModTime().Format("01/02/2006 15:04:05"))
	fmt.Printf("%sDirectory:%s %t\n", logger.Color.Cyan(), logger.Color.End(), info.IsDir())
}

// Find all DocShard files associated with the given filePath
func findDocShardFiles(filePath string) (docShards [][]byte, recreate string, err error) {
	fileName := filepath.Base(filePath)
	split := strings.Split(fileName, "-")
	if len(split) < 2 {
		err = fmt.Errorf("%q is not a DocShard file", filePath)
		return
	}

	fileDir := filepath.Dir(filePath)
	ext := filepath.Ext(fileName)

	prefix := fmt.Sprintf("%s-", split[0])

	files, err := os.ReadDir(fileDir)
	if err != nil {
		return
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		shardFileName := file.Name()
		if strings.HasPrefix(shardFileName, prefix) && filepath.Ext(shardFileName) == ext {
			shardFilePath := filepath.Join(fileDir, shardFileName)
			shard, errr := os.ReadFile(shardFilePath)
			if errr != nil {
				err = fmt.Errorf("could not read DocShard file %q", shardFilePath)
				return
			}

			docShards = append(docShards, shard)
		}
	}

	recreate = fmt.Sprintf("%s%s", split[0], ext)

	return
}

// Clone a repo with Git HTTPS, it will return error if Git executable cannot be found in the system's $PATH.
// The repo format follows go convention, github.com/civilware/tela and will accept branches with github.com/civilware/tela@branch
func gitClone(repo string) (err error) {
	expectedFormat := "domain/user/repo"
	if strings.HasPrefix(repo, "https://") || strings.HasSuffix(repo, ".git") {
		err = fmt.Errorf("expecting URL format %q or %q", expectedFormat, fmt.Sprintf("%s@branch", expectedFormat))
		return
	}

	parts := strings.Split(repo, "@")
	source := parts[0]

	split := strings.Split(source, "/")
	if len(split) < 3 {
		err = fmt.Errorf("invalid URL format, expecting %q", expectedFormat)
		return
	}

	for _, part := range split {
		if part == "" {
			err = fmt.Errorf("invalid URL %q", source)
			return
		}
	}

	_, err = exec.LookPath("git")
	if err != nil {
		return
	}

	url := fmt.Sprintf("https://%s.git", source)
	path := filepath.Join(tela.GetClonePath(), filepath.Base(source))

	args := []string{"clone", url, path}
	if len(parts) > 1 {
		args = append(args, "-b", parts[1])
	}

	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// Get the contents of two files and prepare for diff
func getFileDiff(file1, file2 string) (diff []string, fileNames []string, err error) {
	fileNames = []string{file1, file2}
	for _, fileName := range fileNames {
		var file []byte
		file, err = os.ReadFile(fileName)
		if err != nil {
			err = fmt.Errorf("read file %s: %s", fileName, err)
			return
		}

		content := string(file)
		for _, r := range content {
			if r > unicode.MaxASCII {
				err = fmt.Errorf("cannot diff file %s: '%c'", fileName, r)
				return
			}
		}

		diff = append(diff, content)
	}

	return
}

// Print the line diffs between two sources
func (t *tela_cli) printDiff(diff []string, names []string) (err error) {
	if len(diff) < 2 || len(names) < 2 {
		err = fmt.Errorf("invalid diff format")
		return
	}

	var lines [][]string
	lines = append(lines, strings.Split(diff[0], "\n"))
	lines = append(lines, strings.Split(diff[1], "\n"))

	lines1Len := len(lines[0])
	lines2Len := len(lines[1])

	var l1, l2 int
	var context bool
	var diffLines []string

	for l1 < lines1Len && l2 < lines2Len {
		if lines[0][l1] == lines[1][l2] {
			l1++
			l2++
			if context {
				lastLineIndex := len(diffLines) - 1
				diffLines[lastLineIndex] = fmt.Sprintf("%s\n", diffLines[lastLineIndex])
				context = false
			}
		} else {
			// Line diffs to print
			rmv := fmt.Sprintf("%d %s- %s%s", l1+1, logger.Color.Red(), lines[0][l1], logger.Color.End())
			add := fmt.Sprintf("%s%d%s %s+ %s%s", logger.Color.Grey(), l2+1, logger.Color.End(), logger.Color.Green(), lines[1][l2], logger.Color.End())
			diffLines = append(diffLines, fmt.Sprintf("%s\n%s", rmv, add))

			l1++
			l2++
			if l1 >= lines1Len {
				continue
			}

			// Rest of file is the same
			l1Index := strings.Index(diff[0], lines[0][l1])
			l2Index := strings.Index(diff[1], lines[0][l1])
			if l1Index != -1 && l2Index != -1 {
				if diff[0][l1Index:] == diff[1][l2Index:] {
					l1 = lines1Len
					if lines2Len <= lines1Len {
						l2 = lines2Len
					}
					l2++
					break
				}
			}

			context = true
		}
	}

	// Get the remaining lines from file1 if removed
	for l1 < lines1Len {
		rmv := fmt.Sprintf("%d %s- %s%s", l1+1, logger.Color.Red(), lines[0][l1], logger.Color.End())
		l1++
		diffLines = append(diffLines, rmv)
	}

	// Get the remaining lines from file2 if added
	for l2 < lines2Len {
		add := fmt.Sprintf("%d %s+ %s%s", l2+1, logger.Color.Green(), lines[1][l2], logger.Color.End())
		l2++
		diffLines = append(diffLines, add)
	}

	diffLinesLen := len(diffLines)
	display := t.pageSize - 1

	isPaged := false
	if diffLinesLen > t.pageSize {
		isPaged = true
		logger.Printf("[%s] Showing %d of %d diffs\n", appName, t.pageSize, diffLinesLen)
	}

	if len(diffLines) > 0 {
		fmt.Printf("--- a/%s\n+++ b/%s\n", names[0], names[1])
	} else {
		logger.Printf("No diffs found\n")
		return
	}

	for printed, p := range diffLines {
		fmt.Println(p)

		end := printed == diffLinesLen-1
		if printed >= display && !end {
			var yes bool
			yes, err = t.readYesNo(fmt.Sprintf("Show more? (%d)", (diffLinesLen-1)-printed))
			if err != nil {
				return
			}

			if !yes {
				break
			}

			display = display + t.pageSize
		}

		if isPaged && end {
			logger.Printf("[%s] End of diffs\n", appName)
		}
	}

	return
}
