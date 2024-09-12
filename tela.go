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
	"unicode"

	"github.com/civilware/Gnomon/rwc"
	"github.com/civilware/tela/logger"
	"github.com/civilware/tela/shards"
	"github.com/creachadair/jrpc2"
	"github.com/creachadair/jrpc2/channel"
	"github.com/deroproject/derohe/cryptography/crypto"
	"github.com/deroproject/derohe/dvm"
	"github.com/deroproject/derohe/globals"
	"github.com/deroproject/derohe/rpc"
	"github.com/deroproject/derohe/walletapi"
	"github.com/gorilla/websocket"

	_ "embed"
)

// TELA-DOC-1 structure
type DOC struct {
	DocType   string   `json:"docType"`           // Language this document is using (ex: "TELA-HTML-1", "TELA-JS-1" or "TELA-CSS-1")
	Code      string   `json:"code"`              // The application code HTML, JS...
	SubDir    string   `json:"subDir"`            // Sub directory to place file in (always use / for further children, ex: "sub1" or "sub1/sub2/sub3")
	SCID      string   `json:"scid"`              // SCID of this DOC, only used after DOC has been installed on-chain
	Author    string   `json:"author"`            // Author of this DOC, only used after DOC has been installed on-chain
	DURL      string   `json:"dURL"`              // TELA dURL
	SCVersion *Version `json:"version,omitempty"` // Version of this DOC SC
	// Signature values of Code
	Signature `json:"signature"`
	// Standard headers
	Headers `json:"headers"`
}

// TELA-INDEX-1 structure
type INDEX struct {
	SCID      string            `json:"scid"`              // SCID of this INDEX, only used after INDEX has been installed on-chain
	Author    string            `json:"author"`            // Author of this INDEX, only used after INDEX has been installed on-chain
	DURL      string            `json:"dURL"`              // TELA dURL
	Mods      string            `json:"mods"`              // TELA modules string, stores addition functionality to be parsed when validating, module tags are separated by comma
	DOCs      []string          `json:"docs"`              // SCIDs of TELA DOCs embedded in this INDEX SC
	SCVersion *Version          `json:"version,omitempty"` // Version of this INDEX SC
	SC        dvm.SmartContract `json:"-"`                 // DVM smart contract is used for further parsing of installed INDEXs
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
	version struct {
		pkg   Version
		index []Version
		docs  []Version
	}
	client struct {
		WS  *websocket.Conn
		RPC *jrpc2.Client
	}
}

// Versioning structure used for package and contracts
type Version struct {
	Major int `json:"major"`
	Minor int `json:"minor"`
	Patch int `json:"patch"`
}

var tela TELA

const DOC_STATIC = "TELA-STATIC-1" // Generic docType for any file type
const DOC_HTML = "TELA-HTML-1"     // HTML docType
const DOC_JSON = "TELA-JSON-1"     // JSON docType
const DOC_CSS = "TELA-CSS-1"       // CSS docType
const DOC_JS = "TELA-JS-1"         // JavaScript docType
const DOC_MD = "TELA-MD-1"         // Markdown docType
const DOC_GO = "TELA-GO-1"         // Golang docType

const DEFAULT_MAX_SERVER = 20   // Default max amount of servers
const DEFAULT_PORT_START = 8082 // Default start port for servers
const DEFAULT_MIN_PORT = 1200   // Minimum port of possible serving range
const DEFAULT_MAX_PORT = 65535  // Maximum port of possible serving range

const MINIMUM_GAS_FEE = uint64(100) // Minimum gas fee used when making transfers

const TAG_LIBRARY = ".lib"       // A collection of standard DOCs embedded within an INDEX, each DOC is its own file
const TAG_DOC_SHARDS = ".shards" // A collection of DocShard DOCs embedded within an INDEX, when recreated this will be one file

// Accepted languages of this TELA package
var acceptedLanguages = []string{DOC_STATIC, DOC_HTML, DOC_JSON, DOC_CSS, DOC_JS, DOC_MD, DOC_GO}

// // Embed the standard TELA smart contracts

//go:embed */TELA-INDEX-1.bas
var TELA_INDEX_1 string

//go:embed */TELA-DOC-1.bas
var TELA_DOC_1 string

// Initialize the default storage path TELA will use, can be changed with SetShardPath if required
func init() {
	tela.version.pkg = Version{Major: 1, Minor: 0, Patch: 0}

	tela.version.index = []Version{
		{Major: 1, Minor: 0, Patch: 0},
		{Major: 1, Minor: 1, Patch: 0},
	}

	tela.version.docs = []Version{
		{Major: 1, Minor: 0, Patch: 0},
	}

	tela.path.main = shards.GetPath()

	initMods()
	initRatings()
	tela.port = DEFAULT_PORT_START
	tela.max = DEFAULT_MAX_SERVER

	// Cleanup any residual files before package is used
	os.RemoveAll(tela.path.tela())
}

// Returns semantic string
func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// LessThan returns true if v is less than ov
func (v *Version) LessThan(ov Version) bool {
	if v.Major < ov.Major {
		return true
	} else if v.Major > ov.Major {
		return false
	}

	if v.Minor < ov.Minor {
		return true
	} else if v.Minor > ov.Minor {
		return false
	}

	if v.Patch < ov.Patch {
		return true
	} else if v.Patch > ov.Patch {
		return false
	}

	return false
}

// Equal returns true if v is equal to ov
func (v *Version) Equal(ov Version) bool {
	return v.Major == ov.Major && v.Minor == ov.Minor && v.Patch == ov.Patch
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

// Parse a TELA DOC that has been formatted for DocShards and get its code shard
func parseDocShardCode(fileName, code string) (shard []byte, err error) {
	start := strings.Index(code, "/*")
	end := strings.Index(code, "*/")

	if start == -1 || end == -1 {
		err = fmt.Errorf("could not parse multiline comment from %s", fileName)
		return
	}

	comment := code[start+3:]
	comment = strings.TrimSuffix(comment, "\n*/")

	shard = []byte(comment)

	return
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

	case DOC_GO:

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

// Handle all the GetSC append errors to result.ValuesString
func getSCErrors(result string) bool {
	errStr := []string{
		"NOT AVAILABLE err:",
		"Unmarshal error",
		"UNKNOWN Data type",
	}

	for _, str := range errStr {
		if strings.Contains(result, str) {
			return true
		}
	}

	return false
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
	if len(res) < 1 || res[0] == "" || getSCErrors(res[0]) {
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
func getTXID(txid, endpoint string) (txidAsHex string, height int64, err error) {
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
	height = result.Txs[0].Block_Height

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

// Get the code of a smart contract at height from endpoint
func getContractCodeAtHeight(height int64, scid, endpoint string) (code string, err error) {
	var params = rpc.GetSC_Params{SCID: scid, Variables: false, Code: true, TopoHeight: height}
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

// Get a default DERO transfer address for the network defined by globals.Arguments --testnet and --simulator flags
func GetDefaultNetworkAddress() (network, destination string) {
	network = "mainnet"
	if b, ok := globals.Arguments["--testnet"].(bool); ok && b {
		network = "testnet"
		if b, ok := globals.Arguments["--simulator"].(bool); ok && b {
			network = "simulator"
		}
	}

	switch network {
	case "simulator":
		destination = "deto1qyvyeyzrcm2fzf6kyq7egkes2ufgny5xn77y6typhfx9s7w3mvyd5qqynr5hx"
	case "testnet":
		destination = "deto1qy0ehnqjpr0wxqnknyc66du2fsxyktppkr8m8e6jvplp954klfjz2qqdzcd8p"
	default:
		destination = "dero1qykyta6ntpd27nl0yq4xtzaf4ls6p5e9pqu0k2x4x3pqq5xavjsdxqgny8270"
	}

	return
}

// transfer0 is used for executing TELA smart contract functions without a DEROVALUE or ASSETVALUE, it creates a transfer of 0 to a default address for the network
func transfer0(wallet *walletapi.Wallet_Disk, ringsize uint64, args rpc.Arguments) (txid string, err error) {
	return Transfer(wallet, ringsize, nil, args)
}

// Transfer is used for executing TELA smart contract actions with DERO walletapi, if nil transfers is passed
// it initializes a transfer of 0 to a default address for the network using GetDefaultNetworkAddress()
func Transfer(wallet *walletapi.Wallet_Disk, ringsize uint64, transfers []rpc.Transfer, args rpc.Arguments) (txid string, err error) {
	if wallet == nil {
		err = fmt.Errorf("no wallet for transfer")
		return
	}

	if ringsize < 2 {
		ringsize = 2
	} else if ringsize > 128 {
		ringsize = 128
	}

	// Initialize a DERO transfer if none is provided
	if transfers == nil || len(transfers) < 1 {
		_, dest := GetDefaultNetworkAddress()
		transfers = []rpc.Transfer{{Destination: dest, Amount: 0}}
	}

	// Validate all transfer addresses
	for i, t := range transfers {
		_, err = globals.ParseValidateAddress(t.Destination)
		if err != nil {
			err = fmt.Errorf("invalid transfer address %d: %s", i, err)
			return
		}
	}

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

	if gasResult.GasStorage < MINIMUM_GAS_FEE {
		gasResult.GasStorage = MINIMUM_GAS_FEE
	}

	tx, err := wallet.TransferPayload0(transfers, ringsize, false, args, gasResult.GasStorage, false)
	if err != nil {
		err = fmt.Errorf("transfer build error: %s", err)
		return
	}

	if err = wallet.SendTransaction(tx); err != nil {
		err = fmt.Errorf("transfer dispatch error: %s", err)
		return
	}

	txid = tx.GetHash().String()

	return
}

// Clone a TELA-DOC SCID to path from endpoint
func cloneDOC(scid, docNum, path, endpoint string) (clone Cloning, err error) {
	if len(scid) != 64 {
		err = fmt.Errorf("invalid DOC SCID: %s", scid)
		return
	}

	var code string
	code, err = getContractCode(scid, endpoint)
	if err != nil {
		err = fmt.Errorf("could not get SC code from %s: %s", scid, err)
		return
	}

	_, _, err = ValidDOCVersion(code)
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

	err = parseAndSaveTELADoc(filePath, code, docType)
	if err != nil {
		err = fmt.Errorf("error saving %s: %s", fileName, err)
		return
	}

	return
}

// Clone a TELA-INDEX SCID to path from endpoint creating all DOCs embedded within the INDEX
func cloneINDEX(scid, dURL, path, endpoint string) (clone Cloning, err error) {
	if len(scid) != 64 {
		err = fmt.Errorf("invalid INDEX SCID: %s", scid)
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

	var modTag string // mods store can be empty so don't return error
	if storedMods, err := getContractVar(scid, "mods", endpoint); err == nil {
		modTag = storedMods
	}

	// Only clone contracts matching TELA standard
	sc, _, err := ValidINDEXVersion(code, modTag)
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

	// If INDEX contains DocShards to be constructed
	if strings.Contains(dURL, TAG_DOC_SHARDS) {
		err = cloneDocShards(sc, basePath, endpoint)
		if err != nil {
			err = fmt.Errorf("%s %s", tagErr, err)
			return
		}
	} else {
		// Parse INDEX SC for valid DOCs
		entrypoint, servePath, err = parseAndCloneINDEXForDOCs(sc, 0, basePath, endpoint)
		if err != nil {
			// If all of the files were not cloned successfully, any residual files are removed if they did not exist already
			err = fmt.Errorf("%s %s", tagErr, err)
			if !strings.Contains(err.Error(), "already exists") {
				os.RemoveAll(basePath)
			}
			return
		}
	}

	clone.DURL = dURL
	clone.BasePath = basePath
	clone.ServePath = servePath
	clone.Entrypoint = entrypoint

	return
}

// cloneDocShards takes a TELA-INDEX SC and parses its DOCs, creating them as DocShards which get recreated as a single file
func cloneDocShards(sc dvm.SmartContract, basePath, endpoint string) (err error) {
	docShards, recreate, err := parseDocShards(sc, basePath, endpoint)
	if err != nil {
		err = fmt.Errorf("could not clone DocShards: %s", err)
		return
	}

	err = ConstructFromShards(docShards, recreate, basePath)
	if err != nil {
		err = fmt.Errorf("could not construct DocShards: %s", err)
		return
	}

	return
}

// ConstructFromShards takes DocShards and recreates them as a file at basePath,
// CreateShardFiles can be used to create the shard files formatted for ConstructFromShards
func ConstructFromShards(docShards [][]byte, recreate, basePath string) (err error) {
	err = os.MkdirAll(basePath, os.ModePerm)
	if err != nil {
		return
	}

	filePath := filepath.Join(basePath, recreate)
	if _, err = os.Stat(filePath); !os.IsNotExist(err) {
		err = fmt.Errorf("file %s already exists", filePath)
		return
	}

	logger.Printf("[TELA] Constructing %s\n", recreate)

	var file *os.File
	file, err = os.Create(filePath)
	if err != nil {
		err = fmt.Errorf("failed to create %s: %s", recreate, err)
		return
	}
	defer file.Close()

	for i, code := range docShards {
		_, err = file.Write(code)
		if err != nil {
			err = fmt.Errorf("failed to write shard %d to %s: %s", i+1, recreate, err)
			return
		}
	}

	return
}

// CreateShardFiles takes a source file and creates DocShard files sized and formatted for installing as TELA DOCs,
// the package uses ConstructFromShards to re-build the DocShards as its original file when cloning,
// output files are formatted as "name-#.ext" in the source file's directory
func CreateShardFiles(filePath string) (err error) {
	var content []byte
	content, err = os.ReadFile(filePath)
	if err != nil {
		err = fmt.Errorf("failed to read file: %s", err)
		return
	}

	fileName := filepath.Base(filePath)
	for _, r := range string(content) {
		if r > unicode.MaxASCII {
			err = fmt.Errorf("cannot shard file %s: '%c'", fileName, r)
			return
		}
	}

	shardSize := int(MAX_DOC_CODE_SIZE*1000) - 500

	totalShards := (len(content) + shardSize - 1) / shardSize

	newFileName := func(i int, name, ext string) string {
		return fmt.Sprintf("%s-%d%s", strings.TrimSuffix(name, ext), i, ext)
	}

	fileDir := filepath.Dir(filePath)
	ext := filepath.Ext(fileName)

	// Check no shard files already exist
	for i := 1; i <= totalShards; i++ {
		name := newFileName(int(i), fileName, ext)
		newPath := filepath.Join(fileDir, name)
		if _, err = os.Stat(newPath); !os.IsNotExist(err) {
			err = fmt.Errorf("file %s already exists", newPath)
			return
		}
	}

	count := 0
	fileEnd := len(content)
	for start := 0; start < fileEnd; start += shardSize {
		end := start + shardSize
		if end > fileEnd {
			end = fileEnd
		}

		count++
		name := newFileName(count, fileName, ext)

		var shardFile *os.File
		shardFile, err = os.Create(filepath.Join(fileDir, name))
		if err != nil {
			err = fmt.Errorf("failed to create %s: %s", name, err)
			return
		}
		defer shardFile.Close()

		if _, err = shardFile.Write(content[start:end]); err != nil {
			err = fmt.Errorf("failed to write %s: %s", name, err)
			return
		}
	}

	return
}

// Clone a TELA-INDEX SCID at commit TXID to path from endpoint creating all DOCs embedded within the INDEX at the commit height
func cloneINDEXAtCommit(height int64, scid, txid, path, endpoint string) (clone Cloning, err error) {
	if len(scid) != 64 {
		err = fmt.Errorf("invalid INDEX SCID: %s", scid)
		return
	}

	// TXID only needed on first INDEX
	if height == 0 && len(txid) != 64 {
		err = fmt.Errorf("invalid INDEX commit TXID: %s", txid)
		return
	}

	dURL, err := getContractVar(scid, HEADER_DURL.Trim(), endpoint)
	if err != nil {
		err = fmt.Errorf("could not get dURL from %s: %s", scid, err)
		return
	}

	tagErr := fmt.Sprintf("cloning %s@%s was not successful:", dURL, txid)

	var code, modTag string
	if height > 0 {
		// If more then one INDEX embed, use height from commit TXID to get docCode at commit height
		code, err = getContractCodeAtHeight(height, scid, endpoint)
		if err != nil {
			return
		}

		modTag = extractModTagFromCode(code)
	} else {
		// First INDEX get commit height and code from TXID
		txidAsHex, commitHeight, errr := getTXID(txid, endpoint)
		if errr != nil {
			err = fmt.Errorf("%s could not get TXID: %s", tagErr, errr)
			return
		}

		height = commitHeight

		code, err = extractCodeFromTXID(txidAsHex)
		if err != nil {
			err = fmt.Errorf("%s could not get SC code: %s", tagErr, err)
			return
		}

		modTag = extractModTagFromCode(code)
	}

	// Only clone contracts matching TELA standard
	sc, _, err := ValidINDEXVersion(code, modTag)
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

	// If INDEX contains DocShards to be constructed
	if strings.Contains(dURL, TAG_DOC_SHARDS) {
		err = cloneDocShards(sc, basePath, endpoint)
		if err != nil {
			err = fmt.Errorf("%s %s", tagErr, err)
			return
		}
	} else {
		// Parse INDEX SC for valid DOCs
		entrypoint, servePath, err = parseAndCloneINDEXForDOCs(sc, height, basePath, endpoint)
		if err != nil {
			// If all of the files were not cloned successfully, any residual files are removed if they did not exist already
			err = fmt.Errorf("%s %s", tagErr, err)
			if !strings.Contains(err.Error(), "already exists") {
				os.RemoveAll(basePath)
			}
			return
		}
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

	dURL, err := getContractVar(scid, HEADER_DURL.Trim(), endpoint)
	if err != nil {
		err = fmt.Errorf("could not get dURL from %s: %s", scid, err)
		return
	}

	path := tela.path.clone()

	switch valid {
	case "INDEX":
		_, err = cloneINDEX(scid, dURL, path, endpoint)
	case "DOC":
		// Store DOCs in respective dURL directories
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

	_, err = cloneINDEXAtCommit(0, scid, txid, path, endpoint)

	return
}

// Before serving check if dURL has any known tags that indicate it should not be served
func checkIfAbleToServe(scid, endpoint string) (dURL string, err error) {
	dURL, err = getContractVar(scid, HEADER_DURL.Trim(), endpoint)
	if err != nil {
		err = fmt.Errorf("could not get INDEX dURL from %s: %s", scid, err)
		return
	}

	if strings.HasSuffix(dURL, TAG_LIBRARY) {
		err = fmt.Errorf("%q is a library and cannot be served", dURL)
		return
	}

	if strings.Contains(dURL, TAG_DOC_SHARDS) {
		err = fmt.Errorf("%q is DocShards and cannot be served", dURL)
		return
	}

	return
}

// serveTELA serves cloned TELA content returning a link to the running TELA server if successful
func serveTELA(scid string, clone Cloning) (link string, err error) {
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

	dURL, err := checkIfAbleToServe(scid, endpoint)
	if err != nil {
		return
	}

	clone, err := cloneINDEX(scid, dURL, tela.path.tela(), endpoint)
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

	_, err = checkIfAbleToServe(scid, endpoint)
	if err != nil {
		return
	}

	clone, err := cloneINDEXAtCommit(0, scid, txid, tela.path.tela(), endpoint)
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

// Get the current clone datashard storage path
func GetClonePath() string {
	tela.RLock()
	defer tela.RUnlock()

	return tela.path.clone()
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
		indexTemplate := TELA_INDEX_1
		if h.Mods != "" {
			_, indexTemplate, err = Mods.InjectMODs(h.Mods, indexTemplate)
			if err != nil {
				err = fmt.Errorf("could not inject MODs: %s", err)
				return
			}
		}

		code, err = ParseHeaders(indexTemplate, h)
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

	return transfer0(wallet, ringsize, args)
}

// Create arguments for INDEX SC UpdateCode call
func NewUpdateArgs(params interface{}) (args rpc.Arguments, err error) {
	var version *Version
	var code, scid, mods string
	switch h := params.(type) {
	case *INDEX:
		indexTemplate := TELA_INDEX_1
		if h.Mods != "" {
			_, indexTemplate, err = Mods.InjectMODs(h.Mods, indexTemplate)
			if err != nil {
				err = fmt.Errorf("could not inject MODs: %s", err)
				return
			}
		}

		code, err = ParseHeaders(indexTemplate, h)
		if err != nil {
			return
		}

		scid = h.SCID
		mods = h.Mods
		if h.SCVersion == nil {
			// Use latest version if not provided
			latestV := GetLatestContractVersion(false)
			version = &latestV
		} else {
			version = h.SCVersion
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

	// Handle any version specific params that need to be added
	switch {
	case version.Equal(Version{1, 1, 0}):
		args = append(args, rpc.Argument{Name: "mods", DataType: rpc.DataString, Value: mods})
	default:
		// nothing, use 1.0.0
	}

	return
}

// Update a TELA INDEX SC with DERO walletapi, requires wallet to be owner of SC
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

	return transfer0(wallet, 2, args)
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

// Rate a TELA SC positively (rating > 49) or negatively (rating < 50) with DERO walletapi, the transaction will use ringsize of 2
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

	return transfer0(wallet, 2, args)
}

// Check if key is stored in SCID at endpoint
func KeyExists(scid, key, endpoint string) (variable string, exists bool, err error) {
	var vars map[string]interface{}
	vars, err = getContractVars(scid, endpoint)
	if err != nil {
		return
	}

	for k, val := range vars {
		if k == key {
			exists = true
			switch v := val.(type) {
			case string:
				variable = decodeHexString(v)
			case uint64:
				variable = fmt.Sprintf("%d", v)
			case float64:
				variable = fmt.Sprintf("%.0f", v)
			default:
				variable = fmt.Sprintf("%v", v)
			}
			break
		}
	}

	return
}

// Check if prefixed key is stored in SCID at endpoint
func KeyPrefixExists(scid, prefix, endpoint string) (key, variable string, exists bool, err error) {
	var vars map[string]interface{}
	vars, err = getContractVars(scid, endpoint)
	if err != nil {
		return
	}

	for k, val := range vars {
		if strings.HasPrefix(k, prefix) {
			exists = true
			key = k
			switch v := val.(type) {
			case string:
				variable = decodeHexString(v)
			case uint64:
				variable = fmt.Sprintf("%d", v)
			case float64:
				variable = fmt.Sprintf("%.0f", v)
			default:
				variable = fmt.Sprintf("%v", v)
			}

			break
		}
	}

	return
}

// Create arguments for TELA SetVar SC call
func NewSetVarArgs(scid, key, value string) (args rpc.Arguments, err error) {
	if len(key) > 256 {
		err = fmt.Errorf("key cannot be larger than 256 characters")
		return
	}

	args = rpc.Arguments{
		rpc.Argument{Name: "entrypoint", DataType: rpc.DataString, Value: "SetVar"},
		rpc.Argument{Name: "k", DataType: rpc.DataString, Value: key},
		rpc.Argument{Name: "v", DataType: rpc.DataString, Value: value},
		rpc.Argument{Name: rpc.SCID, DataType: rpc.DataHash, Value: crypto.HashHexToHash(scid)},
		rpc.Argument{Name: rpc.SCACTION, DataType: rpc.DataUint64, Value: uint64(rpc.SC_CALL)},
	}

	return
}

// Set a K/V store on a TELA SC
func SetVar(wallet *walletapi.Wallet_Disk, scid, key, value string) (txid string, err error) {
	if wallet == nil {
		err = fmt.Errorf("no wallet for TELA SetVar")
		return
	}

	var args rpc.Arguments
	args, err = NewSetVarArgs(scid, key, value)
	if err != nil {
		return
	}

	return transfer0(wallet, 2, args)
}

// Create arguments for TELA DeleteVar SC call
func NewDeleteVarArgs(scid, key string) (args rpc.Arguments, err error) {
	if len(key) > 256 {
		err = fmt.Errorf("invalid key")
		return
	}

	args = rpc.Arguments{
		rpc.Argument{Name: "entrypoint", DataType: rpc.DataString, Value: "DeleteVar"},
		rpc.Argument{Name: "k", DataType: rpc.DataString, Value: key},
		rpc.Argument{Name: rpc.SCID, DataType: rpc.DataHash, Value: crypto.HashHexToHash(scid)},
		rpc.Argument{Name: rpc.SCACTION, DataType: rpc.DataUint64, Value: uint64(rpc.SC_CALL)},
	}

	return
}

// Delete a K/V store from a TELA SC, requires wallet to be owner of SC
func DeleteVar(wallet *walletapi.Wallet_Disk, scid, key string) (txid string, err error) {
	if wallet == nil {
		err = fmt.Errorf("no wallet for TELA DeleteVar")
		return
	}

	var args rpc.Arguments
	args, err = NewDeleteVarArgs(scid, key)
	if err != nil {
		return
	}

	return transfer0(wallet, 2, args)
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

	var modTag string
	storedMods, ok := vars["mods"].(string)
	if ok {
		modTag = decodeHexString(storedMods)
	}

	code := decodeHexString(c)
	_, _, err = ValidINDEXVersion(code, modTag)
	if err != nil {
		_, _, err = ValidDOCVersion(code)
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
	_, version, err := ValidDOCVersion(code)
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
		DocType:   docType,
		Code:      code,
		SubDir:    subDir,
		SCID:      scid,
		Author:    author,
		DURL:      dURL,
		SCVersion: &version,
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

	var modTag string
	storedMods, ok := vars["mods"].(string)
	if ok {
		modTag = decodeHexString(storedMods)
	}

	code := decodeHexString(c)
	sc, version, err := ValidINDEXVersion(code, modTag)
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
	docs := parseINDEXForDOCs(sc)

	index = INDEX{
		Mods:      modTag,
		SCID:      scid,
		Author:    author,
		DURL:      dURL,
		DOCs:      docs,
		SCVersion: &version,
		SC:        sc,
		Headers: Headers{
			NameHdr:  nameHdr,
			DescrHdr: descrHdr,
			IconHdr:  iconHdr,
		},
	}

	return
}
