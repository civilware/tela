package main

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/chzyer/readline"
	"github.com/civilware/tela"
	"github.com/civilware/tela/logger"
	"github.com/civilware/tela/shards"
	"github.com/deroproject/derohe/globals"
	"github.com/deroproject/derohe/walletapi"
)

var altSet = [][]string{
	{"▼", "▲"},
	{"□", "■"},
	{"✘", "✔"},
}

type shardKeys struct {
	pageSize []byte
	minLikes []byte
	gnomon   struct {
		fastsync       []byte
		parallelBlocks []byte
	}
}

var keys shardKeys

// Initialize TELA-CLI
func init() {
	parseHelpFlag() // Parse help argument
	logger.ASCIIPrint(false)
	fmt.Println("TELA: Decentralized Web Standard")
	fmt.Print("Initializing...")
	keys.pageSize = []byte("tela-cli.page-size")
	keys.minLikes = []byte("tela-cli.search-min-likes")
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
func parseFlags() (arguments map[string]string) {
	arguments = map[string]string{}
	for i, a := range os.Args {
		if i == 0 {
			continue
		}

		flag, arg, _ := strings.Cut(a, "=")
		arguments[flag] = strings.TrimSpace(arg)
		switch flag {
		case "--testnet", "--simulator":
			globals.Arguments[flag] = true
			logger.Printf("[%s] %s enabled\n", appName, flag)
			if flag == "--simulator" {
				if !globals.Arguments["--testnet"].(bool) {
					globals.Arguments["--testnet"] = true
				}
			}
		}
	}

	globals.Arguments["--debug"] = false
	if _, ok := arguments["--debug"]; ok {
		globals.Arguments["--debug"] = true
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

	if arg, ok := arguments["--fastsync"]; ok {
		if arg != "" {
			b, err := strconv.ParseBool(arg)
			if err != nil {
				logger.Warnf("[%s] %s: %s\n", appName, "--fastsync", err)
			} else {
				logger.Printf("[%s] %s=%t\n", appName, "--fastsync", b)
				err = shards.StoreSettingsValue(keys.gnomon.fastsync, []byte(arg))
				if err != nil {
					logger.Debugf("[%s] Storing fastsync: %s\n", appName, err)
				}
			}
		}
	}

	if arg, ok := arguments["--num-parallel-blocks"]; ok {
		if arg != "" {
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
			}
		}
	}

	return
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

// Get stored preferences
func (t *tela_cli) getStoredPreferences() {
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

// Open wallet file with password
func (t *tela_cli) openWallet(file, password string) (err error) {
	if _, err = os.Stat(file); os.IsNotExist(err) {
		err = fmt.Errorf("wallet file %s does not exist", file)
		return
	}

	for password == "" {
		var line []byte
		line, err = t.readWithPasswordPrompt(fmt.Sprintf("Enter %s password", path.Base(file)))
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
		line, err = t.readLine(fmt.Sprintf("Enter DOC%d SCID", i+1), "")
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

			if !strings.HasSuffix(ind.DURL, tela.TAG_LIBRARY) {
				logger.Errorf("[%s] INDEX %s is not a library\n", appName, ind.DURL)
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
						indexErr = fmt.Sprintf("Import from %s INDEX already contains a DOC with path %q\n", d, filePath)
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

	// Create TELA INDEX
	index = tela.INDEX{
		DURL: headers[tela.HEADER_DURL],
		DOCs: scids,
		Headers: tela.Headers{
			NameHdr:  nameHdr,
			DescrHdr: headers[tela.HEADER_DESCRIPTION],
			IconHdr:  headers[tela.HEADER_ICON_URL],
		},
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
	margin := "------------------------------"
	allInfo = append(allInfo, margin)

	allInfo = append(allInfo, fmt.Sprintf("OS: %s", t.os))
	walletName := "offline"
	if t.wallet.disk != nil {
		walletName = t.wallet.name
	}

	allInfo = append(allInfo, fmt.Sprintf("Wallet: %s", path.Base(walletName)))
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

	_, servers, local := t.getServerInfo()
	allInfo = append(allInfo, fmt.Sprintf("Active servers: %d/%d", servers, tela.MaxServers()+1))
	if servers > 0 {
		allInfo = append(allInfo, fmt.Sprintf("Local server running: %t", local))

	}

	allInfo = append(allInfo, fmt.Sprintf("Search minimum likes: %.0f%%", t.minLikes))
	allInfo = append(allInfo, fmt.Sprintf("Open content in browser: %t", t.openInBrowser))
	allInfo = append(allInfo, fmt.Sprintf("Updated content allowed: %t", tela.UpdatesAllowed()))

	allInfo = append(allInfo, margin)

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

	dURL = d[0]

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
func parseINDEXInfo(scid, owner, dURL string, ratio float64) (lines []string) {
	nameHdr, _ := gnomon.GetSCIDValuesByKey(scid, "nameHdr")
	if nameHdr == nil {
		nameHdr = append(nameHdr, "?")
	}

	lines = append(lines, fmt.Sprintf("%sdURL:%s %-65s Author: %s", logger.Color.Grey(), logger.Color.End(), dURL, owner))
	lines = append(lines, fmt.Sprintf("SCID: %s  Type: %-16s  Name: %-33s  Likes: %s", scid, "TELA-INDEX-1", nameHdr[0], colorLikesRatio(ratio)))

	return
}

// Search for INDEX info from Gnomon DB filtering for wallet owner
func (t *tela_cli) searchINDEXInfo(all map[string]string, owned bool) (resultLines [][]string) {
	if owned && t.wallet.disk == nil {
		return
	}

	for sc, owner := range all {
		if owned && owner != t.wallet.disk.GetAddress().String() {
			continue
		}

		dURL, likesRatio, err := t.getLikesRatio(sc, !owned)
		if err != nil {
			continue
		}

		ind, _ := gnomon.GetSCIDValuesByKey(sc, "DOC1")
		if ind != nil {
			resultLines = append(resultLines, parseINDEXInfo(sc, owner, dURL, likesRatio))
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
		for sc, owner := range all {
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
		for sc, owner := range all {
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
func (t *tela_cli) getLibraries() (libKeys []tela.Library, libMap map[tela.Library][]tela.DOC) {
	all := gnomon.GetAllOwnersAndSCIDs()
	if len(all) < 1 {
		return
	}

	libMap = map[tela.Library][]tela.DOC{}
	for sc, owner := range all {
		dURL, _, err := t.getLikesRatio(sc, true)
		if err != nil {
			continue
		}

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
					libMap[tLib] = append(libMap[tLib], doc)
				}
			} else {
				// Single DOCs tagged as lib will be grouped by dURL and owner
				doc := createDOC(sc, dURL)
				libMap[tLib] = append(libMap[tLib], doc)
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
func parseLibraryInfo(lib tela.Library, docs map[tela.Library][]tela.DOC) (lines []string) {
	identifier := fmt.Sprintf("Author: %s", lib.Author)
	if lib.SCID != "" {
		identifier = fmt.Sprintf("SCID: %s", lib.SCID)
	}

	lines = append(lines, fmt.Sprintf("%sdURL:%s %-65s %s", logger.Color.Grey(), logger.Color.End(), lib.DURL, identifier))
	for _, doc := range docs[lib] {
		lines = append(lines, fmt.Sprintf("SCID: %s  DocType: %-13s  Name: %-33s  Likes: %s", doc.SCID, doc.DocType, doc.NameHdr, colorLikesRatio(lib.LikesRatio)))
	}

	return
}

// Parse info for search queries switching between DOC, INDEX or generic and return resulting lines to print
func parseSearchQuery(scid, owner, dURL string, ratio float64) (lines []string) {
	scType := getSCType(scid)
	if scType == "TELA-INDEX-1" {
		return parseINDEXInfo(scid, owner, dURL, ratio)
	} else if strings.Contains(scType, "TELA-") {
		return parseDOCInfo(scid, owner, scType, dURL, ratio)
	} else {
		return parseBASInfo(scid, owner)
	}
}
