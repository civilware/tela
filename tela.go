package tela

import (
	"context"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/civilware/Gnomon/rwc"
	"github.com/civilware/tela/logger"
	"github.com/civilware/tela/shards"
	"github.com/creachadair/jrpc2"
	"github.com/creachadair/jrpc2/channel"
	"github.com/deroproject/derohe/cryptography/crypto"
	"github.com/deroproject/derohe/globals"
	"github.com/deroproject/derohe/rpc"
	"github.com/deroproject/derohe/walletapi"
	"github.com/gorilla/websocket"

	_ "embed"
)

// TELA-DOC-1 structure
type DOC struct {
	DocType string `json:"docType"` // Language this document is using (ex: "TELA-HTML-1", "TELA-JS-1" or "TELA-CSS-1")
	Code    string `json:"code"`    // The application code HTML, JS...
	SubDir  string `json:"subDir"`  // Sub directory to place file in (always use / for further children, ex: "sub1" or "sub1/sub2/sub3")
	SCID    string `json:"scid"`    // SCID of this DOC, only used after DOC has been installed on-chain
	Author  string `json:"author"`  // Author of this DOC, only used after DOC has been installed on-chain
	DURL    string `json:"dURL"`    // TELA dURL
	// Signature values of Code
	Signature `json:"signature"`
	// Standard headers
	Headers `json:"headers"`
}

// TELA-INDEX-1 structure
type INDEX struct {
	SCID   string   `json:"scid"`   // SCID of this INDEX, only used after INDEX has been installed on-chain
	Author string   `json:"author"` // Author of this INDEX, only used after INDEX has been installed on-chain
	DURL   string   `json:"dURL"`   // TELA dURL
	DOCs   []string `json:"docs"`   // SCIDs of TELA DOCs embedded in this INDEX SC
	// Standard headers
	Headers `json:"headers"`
}

// Cloning structure for creating DOC/INDEX
type Cloning struct {
	BasePath   string `json:"basePath"`   // Main directory path for TELA files
	ServePath  string `json:"servePath"`  // URL serve path
	Entrypoint string `json:"entrypoint"` // INDEX entrypoint
	DURL       string `json:"dURL"`       // TELA dURL
	Hash       string `json:"hash"`       // Commit hash of INDEX
}

// Library structure for search queries
type Library struct {
	DURL       string  `json:"dURL"`       // TELA library dURL
	Author     string  `json:"author"`     // Author of the library
	SCID       string  `json:"scid"`       // SCID of the library DOC or INDEX
	LikesRatio float64 `json:"likesRatio"` // Likes to dislike ratio of the library
}

// Local TELA server info
type ServerInfo struct {
	Name       string
	Address    string
	SCID       string
	Entrypoint string
}

// Datashards structure
type ds struct {
	main string
}

// TELA core components for serving content from TELA-INDEX-1 smart contracts
type TELA struct {
	sync.RWMutex
	servers map[ServerInfo]*http.Server
	path    ds   // Access datashard paths
	updates bool // Allow updated content
	port    int  // Start port to range servers from
	max     int  // Max amount of TELA servers
	client  struct {
		WS  *websocket.Conn
		RPC *jrpc2.Client
	}
}

var tela TELA

const DOC_STATIC = "TELA-STATIC-1" // Generic docType for any file type
const DOC_HTML = "TELA-HTML-1"     // HTML docType
const DOC_JSON = "TELA-JSON-1"     // JSON docType
const DOC_CSS = "TELA-CSS-1"       // CSS docType
const DOC_JS = "TELA-JS-1"         // JavaScript docType
const DOC_MD = "TELA-MD-1"         // Markdown docType

const DEFAULT_MAX_SERVER = 20   // Default max amount of servers
const DEFAULT_PORT_START = 8082 // Default start port for servers
const DEFAULT_MIN_PORT = 1200   // Minimum port of possible serving range
const DEFAULT_MAX_PORT = 65535  // Maximum port of possible serving range

const TAG_LIBRARY = ".lib"

const TELA_VERSION = "1.0.0"

// Accepted languages of this TELA package
var acceptedLanguages = []string{DOC_STATIC, DOC_HTML, DOC_JSON, DOC_CSS, DOC_JS, DOC_MD}

// // Embed the standard TELA smart contracts

//go:embed */TELA-INDEX-1.bas
var TELA_INDEX_1 string

//go:embed */TELA-DOC-1.bas
var TELA_DOC_1 string

// Initialize the default storage path TELA will use, can be changed with SetShardPath if required
func init() {
	tela.path.main = shards.GetPath()

	initRatings()
	tela.port = DEFAULT_PORT_START
	tela.max = DEFAULT_MAX_SERVER

	// Cleanup any residual files before package is used
	os.RemoveAll(tela.path.tela())
}

// Returns TELA datashard path
func (s ds) tela() string {
	return filepath.Join(s.main, "tela")
}

// Returns TELA clone path
func (s ds) clone() string {
	return filepath.Join(s.main, "clone")
}

// Find if port is within valid range
func isValidPort(port int) bool {
	if port < DEFAULT_MIN_PORT || port > DEFAULT_MAX_PORT-tela.max {
		return false
	}
	return true
}

// Listen for open ports and returns http server for TELA content on open port if found
func FindOpenPort() (server *http.Server, found bool) {
	max := tela.port + tela.max
	port := tela.port // Start on tela.port and try +20
	server = &http.Server{Addr: fmt.Sprintf(":%d", port)}
	for !found && port < max {
		li, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			logger.Debugf("[TELA] Finding port: %s\n", err)
			port++ // Not found, try next port
			server.Addr = fmt.Sprintf(":%d", port)
			time.Sleep(time.Millisecond * 50)
			continue
		}

		li.Close()
		found = true

	}

	if !found {
		server = nil
	}

	return
}

// Check if language used is accepted by TELA, see acceptedLanguages for full list
func IsAcceptedLanguage(language string) bool {
	for _, l := range acceptedLanguages {
		if language == l {
			return true
		}
	}

	return false
}

// Parse a TELA DOC for useable code and write file if IsAcceptedLanguage
func parseAndSaveTELADoc(filePath, code, doctype string) (err error) {
	start := strings.Index(code, "/*")
	end := strings.Index(code, "*/")

	if start == -1 || end == -1 {
		err = fmt.Errorf("could not parse multiline comment")
		return
	}

	comment := code[start+2:]
	comment = strings.TrimSpace(strings.TrimSuffix(comment, "*/"))

	// TODO any further DOC parsing for docTypes
	switch doctype {
	case DOC_HTML:

	case DOC_JSON:

	case DOC_CSS:

	case DOC_JS:

	case DOC_MD:

	case DOC_STATIC:

	default:
		err = fmt.Errorf("invalid docType")
		return
	}

	err = os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
	if err != nil {
		return
	}

	logger.Printf("[TELA] Creating %s\n", filepath.Base(filePath))

	return os.WriteFile(filePath, []byte(comment), 0644)
}

// Decode a hex string if possible otherwise return it
func decodeHexString(hexStr string) string {
	if decode, err := hex.DecodeString(hexStr); err == nil {
		return string(decode)
	}

	return hexStr
}

// Get a string key from smart contract at endpoint
func getContractVar(scid, key, endpoint string) (variable string, err error) {
	var params = rpc.GetSC_Params{SCID: scid, Variables: false, Code: false, KeysString: []string{key}}
	var result rpc.GetSC_Result

	tela.client.WS, _, err = websocket.DefaultDialer.Dial("ws://"+endpoint+"/ws", nil)
	if err != nil {
		return
	}

	input_output := rwc.New(tela.client.WS)
	tela.client.RPC = jrpc2.NewClient(channel.RawJSON(input_output, input_output), nil)

	err = tela.client.RPC.CallResult(context.Background(), "DERO.GetSC", params, &result)
	if err != nil {
		return
	}

	res := result.ValuesString
	if len(res) < 1 || res[0] == "" || strings.Contains(res[0], "NOT AVAILABLE err:") {
		err = fmt.Errorf("invalid string value for %q", key)
		return
	}

	// uint values don't need to be decoded
	if key == "likes" || key == "dislikes" {
		variable = res[0]
		return
	}

	variable = decodeHexString(res[0])

	return
}

// Get a TXID as hex from daemon endpoint
func getTXID(txid, endpoint string) (txidAsHex string, err error) {
	var params = rpc.GetTransaction_Params{Tx_Hashes: []string{txid}}
	var result rpc.GetTransaction_Result

	tela.client.WS, _, err = websocket.DefaultDialer.Dial("ws://"+endpoint+"/ws", nil)
	if err != nil {
		return
	}

	input_output := rwc.New(tela.client.WS)
	tela.client.RPC = jrpc2.NewClient(channel.RawJSON(input_output, input_output), nil)

	err = tela.client.RPC.CallResult(context.Background(), "DERO.GetTransaction", params, &result)
	if err != nil {
		return
	}

	res := result.Txs_as_hex
	if len(res) < 1 || res[0] == "" {
		err = fmt.Errorf("no data found for TXID %s", txid)
		return
	}

	txidAsHex = res[0]

	return
}

// Get the current state of all string keys in a smart contract
func getContractVars(scid, endpoint string) (vars map[string]interface{}, err error) {
	var params = rpc.GetSC_Params{SCID: scid, Variables: true, Code: false}
	var result rpc.GetSC_Result

	tela.client.WS, _, err = websocket.DefaultDialer.Dial("ws://"+endpoint+"/ws", nil)
	if err != nil {
		return
	}

	input_output := rwc.New(tela.client.WS)
	tela.client.RPC = jrpc2.NewClient(channel.RawJSON(input_output, input_output), nil)

	err = tela.client.RPC.CallResult(context.Background(), "DERO.GetSC", params, &result)
	if err != nil {
		return
	}

	vars = result.VariableStringKeys

	return
}

// Get the current code of a smart contract at endpoint
func getContractCode(scid, endpoint string) (code string, err error) {
	var params = rpc.GetSC_Params{SCID: scid, Variables: false, Code: true}
	var result rpc.GetSC_Result

	tela.client.WS, _, err = websocket.DefaultDialer.Dial("ws://"+endpoint+"/ws", nil)
	if err != nil {
		return
	}

	input_output := rwc.New(tela.client.WS)
	tela.client.RPC = jrpc2.NewClient(channel.RawJSON(input_output, input_output), nil)

	err = tela.client.RPC.CallResult(context.Background(), "DERO.GetSC", params, &result)
	if err != nil {
		return
	}

	if result.Code == "" {
		err = fmt.Errorf("code is empty string")
		return
	}

	code = result.Code

	return
}

// Transfer for executing TELA smart contract actions with DERO walletapi
func transfer(wallet *walletapi.Wallet_Disk, ringsize uint64, args rpc.Arguments) (txid string, err error) {
	if wallet == nil {
		err = fmt.Errorf("no wallet for transfer")
		return
	}

	network := "mainnet"
	if b, ok := globals.Arguments["--testnet"].(bool); ok && b {
		network = "testnet"
		if b, ok := globals.Arguments["--simulator"].(bool); ok && b {
			network = "simulator"
		}
	}

	if ringsize < 2 {
		ringsize = 2
	} else if ringsize > 128 {
		ringsize = 128
	}

	// Initialize a DERO transfer
	var dest string
	switch network {
	case "simulator":
		dest = "deto1qyvyeyzrcm2fzf6kyq7egkes2ufgny5xn77y6typhfx9s7w3mvyd5qqynr5hx"
	case "testnet":
		dest = "deto1qy0ehnqjpr0wxqnknyc66du2fsxyktppkr8m8e6jvplp954klfjz2qqdzcd8p"
	default:
		dest = "dero1qykyta6ntpd27nl0yq4xtzaf4ls6p5e9pqu0k2x4x3pqq5xavjsdxqgny8270"
	}

	transfers := []rpc.Transfer{{Destination: dest, Amount: 0}}

	var code string
	if c, ok := args.Value(rpc.SCCODE, rpc.DataString).(string); ok {
		code = c
	}

	// Get gas estimate for transfer
	gasParams := rpc.GasEstimate_Params{
		Transfers: transfers,
		SC_Code:   code,
		SC_Value:  0,
		SC_RPC:    args,
		Ringsize:  ringsize,
	}

	if ringsize == 2 {
		gasParams.Signer = wallet.GetAddress().String()
	}

	endpoint := walletapi.Daemon_Endpoint_Active
	tela.client.WS, _, err = websocket.DefaultDialer.Dial("ws://"+endpoint+"/ws", nil)
	if err != nil {
		err = fmt.Errorf("could not dial daemon endpoint %s: %s", endpoint, err)
		return
	}

	input_output := rwc.New(tela.client.WS)
	tela.client.RPC = jrpc2.NewClient(channel.RawJSON(input_output, input_output), nil)

	var gasResult rpc.GasEstimate_Result
	if err = tela.client.RPC.CallResult(context.Background(), "DERO.GetGasEstimate", gasParams, &gasResult); err != nil {
		err = fmt.Errorf("could not estimate install fees: %s", err)
		return
	}

	if gasResult.GasStorage < 90 {
		gasResult.GasStorage = 90
	}

	tx, err := wallet.TransferPayload0(transfers, ringsize, false, args, gasResult.GasStorage, false)
	if err != nil {
		err = fmt.Errorf("contract install build error: %s", err)
		return
	}

	if err = wallet.SendTransaction(tx); err != nil {
		err = fmt.Errorf("contract install dispatch error: %s", err)
		return
	}

	txid = tx.GetHash().String()

	return
}

// Clone a TELA-DOC scid to path from endpoint
func cloneDOC(scid, docNum, path, endpoint string) (clone Cloning, err error) {
	if len(scid) != 64 {
		err = fmt.Errorf("invalid DOC SCID: %s", scid)
		return
	}

	var scCode string
	scCode, err = getContractCode(scid, endpoint)
	if err != nil {
		err = fmt.Errorf("could not get SC code from %s: %s", scid, err)
		return
	}

	_, err = EqualSmartContracts(TELA_DOC_1, scCode)
	if err != nil {
		err = fmt.Errorf("scid does not parse as TELA-DOC-1: %s", err)
		return
	}

	var docType string
	docType, err = getContractVar(scid, HEADER_DOCTYPE.Trim(), endpoint)
	if err != nil {
		err = fmt.Errorf("could not get docType from %s: %s", scid, err)
		return
	}

	var fileName string
	fileName, err = getContractVar(scid, HEADER_NAME.Trim(), endpoint)
	if err != nil {
		err = fmt.Errorf("could not get nameHdr from %s", scid)
		return
	}

	// Set entrypoint DOC
	isDOC1 := Header(docNum) == HEADER_DOCUMENT.Number(1)
	if isDOC1 {
		clone.Entrypoint = fileName
	}

	// Check if DOC is to be placed in subDir
	var subDir string
	subDir, err = getContractVar(scid, HEADER_SUBDIR.Trim(), endpoint)
	if err != nil && !strings.Contains(err.Error(), "invalid string value for") { // only return on RPC error
		err = fmt.Errorf("could not get subDir for %s: %s", fileName, err)
		return
	}

	// If a valid subDir was decoded add it to path for this DOC
	if subDir != "" {
		// Split all subDir to create path
		split := strings.Split(subDir, "/")
		for _, s := range split {
			path = filepath.Join(path, s)
		}

		// If serving from subDir point to it
		if isDOC1 {
			clone.ServePath = fmt.Sprintf("/%s", subDir)
		}
	}

	filePath := filepath.Join(path, fileName)
	if _, err = os.Stat(filePath); !os.IsNotExist(err) {
		err = fmt.Errorf("file %s already exists", filePath)
		return
	}

	if !IsAcceptedLanguage(docType) {
		err = fmt.Errorf("%s is not an accepted language for DOC %s", docType, fileName)
		return
	}

	err = parseAndSaveTELADoc(filePath, scCode, docType)
	if err != nil {
		err = fmt.Errorf("error saving %s: %s", fileName, err)
		return
	}

	return
}

// Clone a TELA-INDEX SCID to path from endpoint creating all DOCs embedded within the INDEX
func cloneINDEX(scid, path, endpoint string) (clone Cloning, err error) {
	if len(scid) != 64 {
		err = fmt.Errorf("invalid INDEX SCID: %s", scid)
		return
	}

	dURL, err := getContractVar(scid, HEADER_DURL.Trim(), endpoint)
	if err != nil {
		err = fmt.Errorf("could not get dURL from %s: %s", scid, err)
		return
	}

	tagErr := fmt.Sprintf("cloning %s@%s was not successful:", dURL, scid)

	hash, err := getContractVar(scid, "hash", endpoint)
	if err != nil {
		err = fmt.Errorf("%s could not get commit hash: %s", tagErr, err)
		return
	}

	tagCommit := fmt.Sprintf("%s@%s", dURL, hash)

	// If the user does not want updated content
	if !tela.updates && scid != hash {
		err = fmt.Errorf("%s user defined no updates and content has been updated to %s", tagErr, tagCommit)
		return
	}

	code, err := getContractCode(scid, endpoint)
	if err != nil {
		err = fmt.Errorf("%s could not get SC code: %s", tagErr, err)
		return
	}

	// Only clone contracts matching TELA standard
	sc, err := EqualSmartContracts(TELA_INDEX_1, code)
	if err != nil {
		err = fmt.Errorf("%s does not parse as TELA-INDEX-1: %s", tagErr, err)
		return
	}

	// TELA-INDEX entrypoint, this will be nameHdr of DOC1
	entrypoint := ""
	// Path where file will be stored
	basePath := filepath.Join(path, dURL)
	// Path to entrypoint
	servePath := ""

	// Parse INDEX SC for valid DOCs
	entrypoint, servePath, err = parseAndCloneINDEXForDOCs(sc, basePath, endpoint)
	if err != nil {
		// If all of the files were not cloned successfully, any residual files are removed if they did not exist already
		err = fmt.Errorf("%s %s", tagErr, err)
		if !strings.Contains(err.Error(), "already exists") {
			os.RemoveAll(basePath)
		}
		return
	}

	clone.DURL = dURL
	clone.BasePath = basePath
	clone.ServePath = servePath
	clone.Entrypoint = entrypoint

	return
}

// Clone a TELA-INDEX SCID at commit TXID to path from endpoint creating all DOCs embedded within the INDEX at that commit
func cloneINDEXAtCommit(scid, txid, path, endpoint string) (clone Cloning, err error) {
	if len(scid) != 64 {
		err = fmt.Errorf("invalid INDEX SCID: %s", scid)
		return
	}

	if len(txid) != 64 {
		err = fmt.Errorf("invalid INDEX commit TXID: %s", txid)
		return
	}

	dURL, err := getContractVar(scid, HEADER_DURL.Trim(), endpoint)
	if err != nil {
		err = fmt.Errorf("could not get dURL from %s: %s", scid, err)
		return
	}

	tagErr := fmt.Sprintf("cloning %s@%s was not successful:", dURL, txid)

	txidAsHex, err := getTXID(txid, endpoint)
	if err != nil {
		err = fmt.Errorf("%s could not get TXID: %s", tagErr, err)
		return
	}

	code, err := extractCodeFromTXID(txidAsHex)
	if err != nil {
		err = fmt.Errorf("%s could not get SC code: %s", tagErr, err)
		return
	}

	// Only clone contracts matching TELA standard
	sc, err := EqualSmartContracts(TELA_INDEX_1, code)
	if err != nil {
		err = fmt.Errorf("%s does not parse as TELA-INDEX-1: %s", tagErr, err)
		return
	}

	// TELA-INDEX entrypoint, this will be nameHdr of DOC1
	entrypoint := ""
	// Path when file will be stored
	basePath := filepath.Join(path, dURL)
	// Path to entrypoint
	servePath := ""

	// Parse INDEX SC for valid DOCs
	entrypoint, servePath, err = parseAndCloneINDEXForDOCs(sc, basePath, endpoint)
	if err != nil {
		// If all of the files were not cloned successfully, any residual files are removed if they did not exist already
		err = fmt.Errorf("%s %s", tagErr, err)
		if !strings.Contains(err.Error(), "already exists") {
			os.RemoveAll(basePath)
		}
		return
	}

	clone.DURL = dURL
	clone.BasePath = basePath
	clone.ServePath = servePath
	clone.Entrypoint = entrypoint

	return
}

// Clone TELA content at SCID from endpoint
func Clone(scid, endpoint string) (err error) {
	var valid string
	_, err = getContractVar(scid, HEADER_DOCTYPE.Trim(), endpoint)
	if err == nil {
		valid = "DOC"
	}

	if valid == "" {
		_, err = getContractVar(scid, HEADER_DOCUMENT.Number(1).Trim(), endpoint)
		if err == nil {
			valid = "INDEX"
		}
	}

	path := tela.path.clone()

	switch valid {
	case "INDEX":
		_, err = cloneINDEX(scid, path, endpoint)
	case "DOC":
		// Store DOCs in respective dURL directories
		dURL, errr := getContractVar(scid, HEADER_DURL.Trim(), endpoint)
		if errr != nil {
			err = fmt.Errorf("could not get DOC dURL from %s: %s", scid, errr)
			return
		}
		_, err = cloneDOC(scid, "", filepath.Join(path, dURL), endpoint)
	default:
		err = fmt.Errorf("could not validate %s as TELA INDEX or DOC", scid)
	}

	return
}

// Clone a TELA-INDEX SC at a commit TXID from endpoint
func CloneAtCommit(scid, txid, endpoint string) (err error) {
	_, err = getContractVar(scid, HEADER_DOCUMENT.Number(1).Trim(), endpoint)
	if err != nil {
		return
	}

	path := tela.path.clone()

	_, err = cloneINDEXAtCommit(scid, txid, path, endpoint)

	return
}

// serveTELA serves cloned TELA content returning a link to the running TELA server if successful
func serveTELA(scid string, clone Cloning) (link string, err error) {
	if strings.HasSuffix(clone.DURL, TAG_LIBRARY) {
		os.RemoveAll(clone.BasePath)
		err = fmt.Errorf("%s is a library", clone.DURL)
		return
	}

	// INDEX and DOCs are valid, get ready to serve
	server, found := FindOpenPort()
	if !found {
		os.RemoveAll(clone.BasePath)
		err = fmt.Errorf("could not find open port to serve %s", clone.DURL)
		return
	}

	// Set the directory to serve files from
	fs := http.FileServer(http.Dir(clone.BasePath))

	// Handle all requests to server
	server.Handler = fs

	// Serve on this address:port
	link = fmt.Sprintf("http://localhost%s/%s", server.Addr+clone.ServePath, clone.Entrypoint)

	if tela.servers == nil {
		tela.servers = make(map[ServerInfo]*http.Server)
	}

	// Add server to TELA
	info := ServerInfo{Name: clone.DURL, Address: server.Addr, SCID: scid, Entrypoint: clone.Entrypoint}
	tela.servers[info] = server

	// Serve content
	go func() {
		logger.Printf("[TELA] Serving %s at %s\n", clone.DURL, link)
		err := server.ListenAndServe()
		if err != nil {
			if err == http.ErrServerClosed {
				logger.Printf("[TELA] Closed %s %s\n", server.Addr, clone.DURL)
			} else {
				logger.Errorf("[TELA] Listen %s %s %s\n", server.Addr, clone.DURL, err)
			}
			os.RemoveAll(clone.BasePath)
		}
	}()

	return
}

// ServeTELA clones and serves a TELA-INDEX-1 SC from endpoint and returns a link to the running TELA server if successful
func ServeTELA(scid, endpoint string) (link string, err error) {
	tela.Lock()
	defer tela.Unlock()

	clone, err := cloneINDEX(scid, tela.path.tela(), endpoint)
	if err != nil {
		os.RemoveAll(clone.BasePath)
		return
	}

	return serveTELA(scid, clone)
}

// ServeAtCommit clones and serves a TELA-INDEX-1 SC from endpoint at commit TXID if the SC code from that commit can be decoded,
// ensure AllowUpdates is set true prior to calling ServeAtCommit otherwise it will return error
func ServeAtCommit(scid, txid, endpoint string) (link string, err error) {
	tela.Lock()
	defer tela.Unlock()

	if !tela.updates {
		err = fmt.Errorf("cannot serve %s at commit as AllowUpdates is set false", scid)
		return
	}

	clone, err := cloneINDEXAtCommit(scid, txid, tela.path.tela(), endpoint)
	if err != nil {
		os.RemoveAll(clone.BasePath)
		return
	}

	return serveTELA(scid, clone)
}

// OpenTELALink will open content from a telaLink formatted as tela://open/<scid>/subDir/../..
// if no server exists for that content it will try starting one using ServeTELA()
func OpenTELALink(telaLink, endpoint string) (link string, err error) {
	target, args, err := ParseTELALink(telaLink)
	if err != nil {
		err = fmt.Errorf("could not parse tela link: %s", err)
		return
	}

	if target != "tela" {
		err = fmt.Errorf("%q target required for OpenTELALink", "tela")
		return
	}

	if len(args) < 2 || args[0] != "open" {
		err = fmt.Errorf("%q is invalid tela link format for OpenTELALink", telaLink)
		return
	}

	var exists bool
	link, err = ServeTELA(args[1], endpoint)
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			err = fmt.Errorf("could not serve tela link: %s", err)
			return
		}

		// Find the server that already exists
		for _, s := range GetServerInfo() {
			if s.SCID == args[1] {
				link = fmt.Sprintf("http://localhost%s", s.Address)
				break
			}
		}

		if link == "" {
			err = fmt.Errorf("could not find active server to create tela link")
			return
		}

		err = nil
		exists = true
	}

	// TELA will serve with entrypoint if server did not exist
	if !exists && len(args) > 2 {
		var entrypoint string
		for _, s := range GetServerInfo() {
			if s.SCID == args[1] {
				entrypoint = fmt.Sprintf("/%s", s.Entrypoint)
				break
			}
		}

		link = strings.TrimSuffix(link, entrypoint)
	}

	// Add link path
	for i, a := range args {
		if i < 2 {
			continue
		}

		link = fmt.Sprintf("%s/%s", link, a)
	}

	return
}

// ShutdownTELA shuts down all TELA servers and cleans up directory
func ShutdownTELA() {
	tela.Lock()
	defer tela.Unlock()

	if tela.servers == nil {
		return
	}

	logger.Printf("[TELA] Shutdown\n")
	for i, s := range tela.servers {
		err := s.Shutdown(context.Background())
		if err != nil {
			logger.Errorf("[TELA] Shutdown: %s\n", err)
		}
		tela.servers[i] = nil
	}

	tela.servers = nil

	if tela.client.WS != nil {
		tela.client.WS.Close()
		tela.client.WS = nil
	}

	if tela.client.RPC != nil {
		tela.client.RPC.Close()
		tela.client.RPC = nil
	}

	// All files removed when servers are shutdown
	os.RemoveAll(tela.path.tela())
}

// ShutdownTELA shuts down running TELA servers by name, if two servers with same name exist both will shutdown
func ShutdownServer(name string) {
	tela.Lock()
	defer tela.Unlock()

	if tela.servers == nil {
		return
	}

	logger.Printf("[TELA] Shutdown %s\n", name)
	for i, s := range tela.servers {
		if i.Name == name {
			err := s.Shutdown(context.Background())
			if err != nil {
				logger.Errorf("[TELA] Shutdown: %s\n", err)
			}
			delete(tela.servers, i)
		}
	}
}

// Get the current TELA datashard storage path
func GetPath() string {
	tela.RLock()
	defer tela.RUnlock()

	return tela.path.tela()
}

// SetShardPath can be used to set a custom path for TELA DOC storage,
// TELA will remove all its files from the /tela directory when servers are Shutdown
func SetShardPath(path string) (err error) {
	tela.Lock()
	if path, err = shards.SetPath(path); err == nil {
		tela.path.main = path
	}
	tela.Unlock()

	return
}

// Get running TELA server info
func GetServerInfo() []ServerInfo {
	tela.RLock()
	defer tela.RUnlock()

	servers := make([]ServerInfo, 0, len(tela.servers))
	for info := range tela.servers {
		servers = append(servers, info)
	}

	return servers
}

// Check if TELA has existing server by name
func HasServer(name string) bool {
	tela.RLock()
	defer tela.RUnlock()

	for info := range tela.servers {
		if strings.EqualFold(info.Name, name) {
			return true
		}
	}
	return false
}

// AllowUpdates default is false and will not allow TELA content to be served that has been updated since its original install
func AllowUpdates(b bool) {
	tela.Lock()
	tela.updates = b
	tela.Unlock()
}

// Check if TELA server is allowed to serve TELA content that has been updated since its original install
func UpdatesAllowed() bool {
	tela.RLock()
	defer tela.RUnlock()

	return tela.updates
}

// Set the initial port to start serving TELA content from if isValidPort
func SetPortStart(port int) (err error) {
	if isValidPort(port) {
		tela.Lock()
		tela.port = port
		tela.Unlock()
	} else {
		err = fmt.Errorf("invalid port %d", port)
	}

	return
}

// Check the initial port that TELA content will be served from
func PortStart() int {
	tela.RLock()
	defer tela.RUnlock()

	return tela.port
}

// Set the maximum amount of TELA servers which can be active
func SetMaxServers(i int) {
	tela.Lock()
	max := DEFAULT_MAX_PORT - tela.port
	if i < 1 {
		tela.max = 1
	} else if i > max { // This would exceed all possible ports within serving range
		tela.max = max
	} else {
		tela.max = i
	}
	tela.Unlock()
}

// Check the maximum amount of TELA servers
func MaxServers() int {
	tela.RLock()
	defer tela.RUnlock()

	return tela.max
}

// Create arguments for INDEX or DOC SC install
func NewInstallArgs(params interface{}) (args rpc.Arguments, err error) {
	var code string
	switch h := params.(type) {
	case *INDEX:
		code, err = ParseHeaders(TELA_INDEX_1, h)
		if err != nil {
			return
		}

		kbSize := GetCodeSizeInKB(code)
		if kbSize > MAX_INDEX_INSTALL_SIZE {
			err = fmt.Errorf("contract exceeds max INDEX install size by %.2fKB", kbSize-MAX_INDEX_INSTALL_SIZE)
			return
		}
	case *DOC:
		code, err = ParseHeaders(TELA_DOC_1, h)
		if err != nil {
			return
		}

		code, err = appendDocCode(code, h.Code)
		if err != nil {
			return
		}
	default:
		err = fmt.Errorf("expecting params to be *INDEX or *DOC for install and got: %T", params)

		return
	}

	args = rpc.Arguments{
		rpc.Argument{Name: rpc.SCACTION, DataType: rpc.DataUint64, Value: uint64(rpc.SC_INSTALL)},
		rpc.Argument{Name: rpc.SCCODE, DataType: rpc.DataString, Value: code},
	}

	return
}

// Install TELA smart contracts with DERO walletapi
func Installer(wallet *walletapi.Wallet_Disk, ringsize uint64, params interface{}) (txid string, err error) {
	if wallet == nil {
		err = fmt.Errorf("no wallet for TELA Installer")
		return
	}

	var args rpc.Arguments
	args, err = NewInstallArgs(params)
	if err != nil {
		return
	}

	return transfer(wallet, ringsize, args)
}

// Create arguments for INDEX SC UpdateCode call
func NewUpdateArgs(params interface{}) (args rpc.Arguments, err error) {
	var code, scid string
	switch h := params.(type) {
	case *INDEX:
		scid = h.SCID
		code, err = ParseHeaders(TELA_INDEX_1, h)
		if err != nil {
			return
		}
	default:
		err = fmt.Errorf("expecting params to be *INDEX for update and got: %T", params)

		return
	}

	args = rpc.Arguments{
		rpc.Argument{Name: "entrypoint", DataType: rpc.DataString, Value: "UpdateCode"},
		rpc.Argument{Name: "code", DataType: rpc.DataString, Value: code},
		rpc.Argument{Name: rpc.SCID, DataType: rpc.DataHash, Value: crypto.HashHexToHash(scid)},
		rpc.Argument{Name: rpc.SCACTION, DataType: rpc.DataUint64, Value: uint64(rpc.SC_CALL)},
	}

	return
}

// Update a TELA INDEX SC with DERO walletapi
func Updater(wallet *walletapi.Wallet_Disk, params interface{}) (txid string, err error) {
	if wallet == nil {
		err = fmt.Errorf("no wallet for TELA Updater")
		return
	}

	var args rpc.Arguments
	args, err = NewUpdateArgs(params)
	if err != nil {
		return
	}

	return transfer(wallet, 2, args)
}

// Create arguments for TELA Rate SC call
func NewRateArgs(scid string, rating uint64) (args rpc.Arguments, err error) {
	if rating > 99 {
		err = fmt.Errorf("invalid TELA rating, must be less than 100")
		return
	}

	// TODO if scid not TELA

	args = rpc.Arguments{
		rpc.Argument{Name: "entrypoint", DataType: rpc.DataString, Value: "Rate"},
		rpc.Argument{Name: "r", DataType: rpc.DataUint64, Value: rating},
		rpc.Argument{Name: rpc.SCID, DataType: rpc.DataHash, Value: crypto.HashHexToHash(scid)},
		rpc.Argument{Name: rpc.SCACTION, DataType: rpc.DataUint64, Value: uint64(rpc.SC_CALL)},
	}

	return
}

// Rate a TELA SC positively (rating > 49) or negatively (rating < 50) with DERO walletapi
func Rate(wallet *walletapi.Wallet_Disk, scid string, rating uint64) (txid string, err error) {
	if wallet == nil {
		err = fmt.Errorf("no wallet for TELA Rate")
		return
	}

	var args rpc.Arguments
	args, err = NewRateArgs(scid, rating)
	if err != nil {
		return
	}

	return transfer(wallet, 2, args)
}

// Get the rating of a TELA scid from endpoint. Result is all individual ratings, likes and dislikes and the average rating category.
// Using height will filter the individual ratings (including only >= height) this will not effect like and dislike results
func GetRating(scid, endpoint string, height uint64) (ratings Rating_Result, err error) {
	var vars map[string]interface{}
	vars, err = getContractVars(scid, endpoint)
	if err != nil {
		return
	}

	c, ok := vars["C"].(string)
	if !ok {
		err = fmt.Errorf("could not get TELA SC code for rating")
		return
	}

	code := decodeHexString(c)
	_, err = EqualSmartContracts(TELA_INDEX_1, code)
	if err != nil {
		_, err = EqualSmartContracts(TELA_DOC_1, code)
		if err != nil {
			err = fmt.Errorf("scid does not parse as a TELA SC: %s", err)
			return
		}
	}

	for k, v := range vars {
		switch k {
		case "likes":
			if f, ok := v.(float64); ok {
				ratings.Likes = uint64(f)
			}
		case "dislikes":
			if f, ok := v.(float64); ok {
				ratings.Dislikes = uint64(f)
			}
		default:
			_, err := globals.ParseValidateAddress(k)
			if err == nil {
				if rStr, ok := v.(string); ok {
					split := strings.Split(decodeHexString(rStr), "_")
					if len(split) < 2 {
						continue // not a valid rating string
					}

					h, err := strconv.ParseUint(split[1], 10, 64)
					if err != nil {
						continue // not a valid rating height
					}

					if h < height {
						continue // filter by height
					}

					r, err := strconv.ParseUint(split[0], 10, 64)
					if err != nil {
						continue // not a valid rating number
					}

					ratings.Ratings = append(ratings.Ratings, Rating{Address: k, Rating: r, Height: h})
				}
			}
		}
	}

	sort.Slice(ratings.Ratings, func(i, j int) bool { return ratings.Ratings[i].Height > ratings.Ratings[j].Height })

	// Gather average rating from the category sum only
	var sum uint64
	for _, num := range ratings.Ratings {
		sum += num.Rating / 10
	}

	if sum > 0 {
		ratings.Average = float64(sum) / float64(len(ratings.Ratings))
	}

	return
}

// Get TELA-DOC info from scid at endpoint
func GetDOCInfo(scid, endpoint string) (doc DOC, err error) {
	vars, err := getContractVars(scid, endpoint)
	if err != nil {
		return
	}

	// SC code, dURL and docType are required, otherwise values can be empty
	c, ok := vars["C"].(string)
	if !ok {
		err = fmt.Errorf("could not get SC code from %s", scid)
		return
	}

	code := decodeHexString(c)
	_, err = EqualSmartContracts(TELA_DOC_1, code)
	if err != nil {
		err = fmt.Errorf("scid does not parse as TELA-DOC-1: %s", err)
		return
	}

	dT, ok := vars[HEADER_DOCTYPE.Trim()].(string)
	if !ok {
		err = fmt.Errorf("could not get docType from %s", scid)
		return
	}

	docType := decodeHexString(dT)
	if !IsAcceptedLanguage(docType) {
		err = fmt.Errorf("could not validate docType %q", docType)
		return
	}

	d, ok := vars[HEADER_DURL.Trim()].(string)
	if !ok {
		err = fmt.Errorf("could not get dURL from %s", scid)
		return
	}

	dURL := decodeHexString(d)

	var nameHdr, descrHdr, iconHdr, subDir, checkC, checkS string
	name, ok := vars[HEADER_NAME.Trim()].(string)
	if ok {
		nameHdr = decodeHexString(name)
	}

	desc, ok := vars[HEADER_DESCRIPTION.Trim()].(string)
	if ok {
		descrHdr = decodeHexString(desc)
	}

	ic, ok := vars[HEADER_ICON_URL.Trim()].(string)
	if ok {
		iconHdr = decodeHexString(ic)
	}

	sd, ok := vars[HEADER_SUBDIR.Trim()].(string)
	if ok {
		subDir = decodeHexString(sd)
	}

	author := "anon"
	addr, ok := vars[HEADER_OWNER.Trim()].(string)
	if ok {
		author = decodeHexString(addr)
	}

	fC, ok := vars[HEADER_CHECK_C.Trim()].(string)
	if ok {
		checkC = decodeHexString(fC)
	}

	fS, ok := vars[HEADER_CHECK_S.Trim()].(string)
	if ok {
		checkS = decodeHexString(fS)
	}

	doc = DOC{
		DocType: docType,
		Code:    code,
		SubDir:  subDir,
		SCID:    scid,
		Author:  author,
		DURL:    dURL,
		Signature: Signature{
			CheckC: checkC,
			CheckS: checkS,
		},
		Headers: Headers{
			NameHdr:  nameHdr,
			DescrHdr: descrHdr,
			IconHdr:  iconHdr,
		},
	}

	return
}

// Get TELA-INDEX info from scid at endpoint
func GetINDEXInfo(scid, endpoint string) (index INDEX, err error) {
	vars, err := getContractVars(scid, endpoint)
	if err != nil {
		return
	}

	// SC code and dURL are required, otherwise values can be empty
	c, ok := vars["C"].(string)
	if !ok {
		err = fmt.Errorf("could not get SC code from %s", scid)
		return
	}

	code := decodeHexString(c)
	_, err = EqualSmartContracts(TELA_INDEX_1, code)
	if err != nil {
		err = fmt.Errorf("scid does not parse as TELA-INDEX-1: %s", err)
		return
	}

	d, ok := vars[HEADER_DURL.Trim()].(string)
	if !ok {
		err = fmt.Errorf("could not get dURL from %s", scid)
		return
	}

	dURL := decodeHexString(d)

	var nameHdr, descrHdr, iconHdr string
	name, ok := vars[HEADER_NAME.Trim()].(string)
	if ok {
		nameHdr = decodeHexString(name)
	}

	desc, ok := vars[HEADER_DESCRIPTION.Trim()].(string)
	if ok {
		descrHdr = decodeHexString(desc)
	}

	ic, ok := vars[HEADER_ICON_URL.Trim()].(string)
	if ok {
		iconHdr = decodeHexString(ic)
	}

	author := "anon"
	addr, ok := vars[HEADER_OWNER.Trim()].(string)
	if ok {
		author = decodeHexString(addr)
	}

	// Get all DOCs from contract code
	docs, err := ParseINDEXForDOCs(code)
	if err != nil {
		return
	}

	index = INDEX{
		SCID:   scid,
		Author: author,
		DURL:   dURL,
		DOCs:   docs,
		Headers: Headers{
			NameHdr:  nameHdr,
			DescrHdr: descrHdr,
			IconHdr:  iconHdr,
		},
	}

	return
}
