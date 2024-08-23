package tela

import (
	"encoding/pem"
	"fmt"
	"math/big"
	"path/filepath"
	"sort"
	"strings"

	"github.com/deroproject/derohe/dvm"
	"github.com/deroproject/derohe/rpc"
)

const (
	DVM_FUNC_INIT_PRIVATE = "InitializePrivate"
	DVM_FUNC_INIT         = "Initialize"

	MAX_DOC_CODE_SIZE      = float64(18)    // DOC SC template file size is +1.2KB with headers
	MAX_DOC_INSTALL_SIZE   = float64(19.2)  // DOC SC total file size (including docCode) should be below this
	MAX_INDEX_INSTALL_SIZE = float64(11.64) // INDEX SC file size should be below this
)

// Append docCode to TELA-DOC-1 smart contract
func appendDocCode(code, docCode string) (newCode string, err error) {
	docSize := GetCodeSizeInKB(docCode)
	if docSize > MAX_DOC_CODE_SIZE {
		err = fmt.Errorf("docCode size is to large, max %.2fKB (%.5f)", MAX_DOC_CODE_SIZE, docSize)
		return
	}

	scSize := GetCodeSizeInKB(code)
	if scSize+docSize > MAX_DOC_INSTALL_SIZE {
		err = fmt.Errorf("DOC SC size is to large, max %.2fKB (%.5f)", MAX_DOC_INSTALL_SIZE, scSize+docSize)
		return
	}

	newCode = fmt.Sprintf("%s\n\n/*\n%s\n*/", code, docCode)

	return
}

// Check if Header key requires a value STORE that is not empty
func requiredHeader(value string, key Header) bool {
	return value == `""` && (key == HEADER_NAME || key == HEADER_CHECK_C || key == HEADER_CHECK_S)
}

// Format a string or uint64 value to be used on a DVM SC, default case assumes value to be string
func formatValue(value interface{}) string {
	switch v := value.(type) {
	case uint64:
		return fmt.Sprintf("%d", v)
	case string:
		return fmt.Sprintf(`"%s"`, v)
	case int:
		return fmt.Sprintf("%d", uint64(v))
	default:
		return fmt.Sprintf(`"%s"`, strings.ReplaceAll(fmt.Sprintf("%v", v), "\n", " "))
	}
}

// Determines if operator requires spacing before and after when formatting smart contract code to string
func isOperator(operator string) (string, bool) {
	operators := map[string]bool{
		// "+":  true, // no spacing for these
		// "-":  true,
		// "/":  true,
		// "*":  true,
		"=":  true,
		">":  true,
		"<":  true,
		"!":  true,
		"==": true,
		"!=": true,
		"<=": true,
		">=": true,
		"&&": true,
		"||": true,
		// "&":  true,
		// "|":  true,
	}

	return operator, operators[operator]
}

// Find how many KB code string is, counting for new lines in total size
func GetCodeSizeInKB(code string) float64 {
	newLines := strings.Count(code, "\n")
	return float64(len([]byte(code))+newLines) / 1024
}

// Parse a file for its TELA docType language
func ParseDocType(fileName string) (language string) {
	ext := filepath.Ext(strings.ToLower(fileName))
	switch ext {
	case ".html":
		language = DOC_HTML
	case ".json":
		language = DOC_JSON
	case ".js":
		language = DOC_JS
	case ".css":
		language = DOC_CSS
	case ".md":
		language = DOC_MD
	case "":
		// nameHdr does not have a file extension
		if fileName == "LICENSE" {
			language = DOC_STATIC
		}
	default:
		language = DOC_STATIC
	}

	return
}

// Parse a INDEX contract for its DOC SCIDs
func ParseINDEXForDOCs(code string) (scids []string, err error) {
	sc, err := EqualSmartContracts(TELA_INDEX_1, code)
	if err != nil {
		err = fmt.Errorf("scid does not parse as TELA-INDEX-1: %s", err)
		return
	}

	var docKeys []string
	docMap := map[string]string{}
	for name, function := range sc.Functions {
		// Find initialize function and parse lines
		if name == DVM_FUNC_INIT_PRIVATE {
			for _, line := range function.Lines {
				// Parse the contents of the line
				for i, parts := range line {
					if strings.Contains(parts, string(HEADER_DOCUMENT)) {
						// Line STORE is a DOC#, find its scid
						scid := strings.Trim(line[i+2], `"`)
						docKeys = append(docKeys, parts)
						docMap[parts] = scid
					}
				}
			}
		}
	}

	// Sort DOC scids by DOC#
	sort.Strings(docKeys)
	for _, v := range docKeys {
		scids = append(scids, docMap[v])
	}

	return
}

func parseAndCloneINDEXForDOCs(sc dvm.SmartContract, basePath, endpoint string) (entrypoint, servePath string, err error) {
	// Parse INDEX SC for valid DOCs
	for name, function := range sc.Functions {
		// Find initialize function and parse lines
		if name == DVM_FUNC_INIT_PRIVATE {
			for _, line := range function.Lines {
				// Parse the contents of the line
				for i, parts := range line {
					if strings.Contains(parts, string(HEADER_DOCUMENT)) {
						// Line STORE is a DOC#, find scid and get its document data
						scid := strings.Trim(line[i+2], `"`)
						isDOC1 := Header(parts) == HEADER_DOCUMENT.Number(1)

						// Check if scid is INDEX or DOC and handle accordingly
						var c Cloning
						var telaVersion string
						telaVersion, err = getContractVar(scid, "telaVersion", endpoint)
						if err != nil {
							c, err = cloneDOC(scid, parts, basePath, endpoint)
							if err != nil {
								return
							}

							// If DOC is entrypoint set it, and if serving from subDir point to it
							if isDOC1 {
								entrypoint = c.Entrypoint
								servePath = c.ServePath
							}
						} else {
							if telaVersion != TELA_VERSION {
								err = fmt.Errorf("cannot use TELA-INDEX v%s when package is v%s", telaVersion, TELA_VERSION)
								return
							}

							if isDOC1 {
								err = fmt.Errorf("cannot use TELA-INDEX as entrypoint for TELA-INDEX")
								return
							}

							var libCheck string
							libCheck, err = getContractVar(scid, "dURL", endpoint)
							if err != nil {
								err = fmt.Errorf("could not verify TELA-INDEX dURL for library embed: %s", err)
								return
							}

							if !strings.HasSuffix(libCheck, TAG_LIBRARY) {
								err = fmt.Errorf("cannot embed TELA-INDEX without %q tag", TAG_LIBRARY)
								return
							}

							c, err = cloneINDEX(scid, basePath, endpoint)
							if err != nil {
								return
							}
						}
					}
				}
			}
		}
	}

	return
}

// Decode a TXID as hex and parse it for SC code and return the result
func extractCodeFromTXID(txidAsHex string) (code string, err error) {
	var codeBlocks []string
	decodedTXID := decodeHexString(txidAsHex)
	startMarker := "Function "
	endMarker := "End Function"

	for {
		startIndex := strings.Index(decodedTXID, startMarker)
		if startIndex == -1 {
			break
		}

		decodedTXID = decodedTXID[startIndex:]

		endIndex := strings.Index(decodedTXID, endMarker)
		if endIndex == -1 {
			break
		}

		endIndex += len(endMarker)

		codeBlock := decodedTXID[:endIndex]
		codeBlocks = append(codeBlocks, codeBlock)

		decodedTXID = decodedTXID[endIndex:]
	}

	if len(codeBlocks) < 1 {
		err = fmt.Errorf("could not extract any SC code from TXID")
		return
	}

	code = strings.Join(codeBlocks, "\n\n")

	return
}

// Parse a DERO signature for address and C, S values
func ParseSignature(input []byte) (address, c, s string, err error) {
	p, _ := pem.Decode(input)
	if p == nil {
		err = fmt.Errorf("unknown format")
		return
	}

	aStr := p.Headers["Address"]
	cStr := p.Headers["C"]
	sStr := p.Headers["S"]

	addr, err := rpc.NewAddress(aStr)
	if err != nil {
		return
	}

	_, ok := new(big.Int).SetString(cStr, 16)
	if !ok {
		err = fmt.Errorf("unknown C format")
		return
	}

	_, ok = new(big.Int).SetString(sStr, 16)
	if !ok {
		err = fmt.Errorf("unknown S format")
		return
	}

	address = addr.String()
	c = cStr
	s = sStr

	return
}

// ParseHeaders takes a headerType and SC code string then returns a formatted SC string with those header values
// See ART-NFA and TELA docs for detailed header info
func ParseHeaders(code string, headerType interface{}) (formatted string, err error) {
	sc, _, err := dvm.ParseSmartContract(code)
	if err != nil {
		err = fmt.Errorf("error parsing code: %s", err)
		return
	}

	var headers map[Header]string

	switch h := headerType.(type) {
	case *INDEX:
		headers = map[Header]string{
			HEADER_NAME:        formatValue(h.NameHdr),
			HEADER_DESCRIPTION: formatValue(h.DescrHdr),
			HEADER_ICON_URL:    formatValue(h.IconHdr),
			HEADER_DURL:        formatValue(h.DURL),
		}

		for i, scid := range h.DOCs {
			doc := HEADER_DOCUMENT.Number(i + 1)
			if _, ok := headers[doc]; !ok {
				headers[doc] = formatValue(scid)
			} else {
				err = fmt.Errorf("conflicting %s document", doc)
				return
			}
		}
	case *DOC:
		headers = map[Header]string{
			HEADER_NAME:        formatValue(h.NameHdr),
			HEADER_DESCRIPTION: formatValue(h.DescrHdr),
			HEADER_ICON_URL:    formatValue(h.IconHdr),
			HEADER_DURL:        formatValue(h.DURL),
			HEADER_SUBDIR:      formatValue(h.SubDir),
			HEADER_DOCTYPE:     formatValue(h.DocType),
			HEADER_CHECK_C:     formatValue(h.CheckC),
			HEADER_CHECK_S:     formatValue(h.CheckS),
		}
	case *Headers:
		headers = map[Header]string{
			HEADER_NAME:        formatValue(h.NameHdr),
			HEADER_DESCRIPTION: formatValue(h.DescrHdr),
			HEADER_ICON_URL:    formatValue(h.IconHdr),
		}
	case map[Header]interface{}:
		headers = map[Header]string{}
		for k, v := range h {
			headers[k] = formatValue(v)
		}
	case map[string]interface{}:
		headers = map[Header]string{}
		for k, v := range h {
			headers[Header(k)] = formatValue(v)
		}
	default:
		err = fmt.Errorf("expecting to parse *INDEX, *DOC, *Headers, map[Header]interface{} or map[string]interface{} and got: %T", headerType)
		return
	}

	added := 0
	want := len(headers)
	for name, function := range sc.Functions {
		if name == DVM_FUNC_INIT_PRIVATE || name == DVM_FUNC_INIT {
			for _, line := range function.Lines {
				if len(line) < 6 {
					// Line is to short to be a STORE
					continue
				}

				for i, parts := range line {
					// Find if there is a existing STORE for this header and update the line with new value
					if parts == "STORE" {
						key := Header(line[i+2])
						value, ok := headers[key]
						if !ok {
							continue
						}

						if requiredHeader(value, key) {
							err = fmt.Errorf("header key %s is empty string", key)
							return
						}

						line[i+4] = value
						delete(headers, key)
						added++
					}
				}
			}

			// Inject further header STORE lines not in the given code
			if added < want {
				// Get remaining headers
				var inject []Header
				for key := range headers {
					inject = append(inject, key)
				}

				sort.Slice(inject, func(i, j int) bool {
					return inject[i] < inject[j]
				})

				// Create the new STORE lines
				var newLines [][]string
				for _, key := range inject {
					line := []string{"STORE", "(", string(key), ",", headers[key], ")"}
					newLines = append(newLines, line)
				}

				// Nothing to add
				if len(newLines) < 1 {
					break
				}

				// Get line numbers and see if there is enough room to add remaining headers
				l := len(function.LineNumbers)
				last := function.LineNumbers[l-1]
				second := uint64(0)
				if l > 1 {
					second = function.LineNumbers[l-2]
				}

				diff := last - second
				if diff < uint64(want-added)+1 {
					err = fmt.Errorf("not enough room to add %d headers", want)
					return
				}

				// Add the new lines to contract
				for i, new := range newLines {
					added++
					u := uint64(i)
					function.Lines[second+1+u] = new
					function.LineNumbers = append(function.LineNumbers, second+1+u)
				}

				// Sort and add new line number index
				sort.Slice(function.LineNumbers, func(i, j int) bool {
					return function.LineNumbers[i] < function.LineNumbers[j]
				})

				for u, ln := range function.LineNumbers {
					function.LinesNumberIndex[ln] = uint64(u)
				}

				sc.Functions[name] = function
			}
		}
	}

	if added != want {
		err = fmt.Errorf("could not add all entries, missing %d headers", want-added)
		return
	}

	return FormatSmartContract(sc, code)
}

// ParseTELALink takes a TELA link and parses it for its target and args. Host applications can use TELA links
// in combination with custom websocket methods to set up and perform their own client specific actions. Usage examples:
//   - target://<arg>/<arg>/<arg>...
//   - tela://open/<scid>/subDir/../..     		      Use like a hyperlink to open external TELA content from a existing page, see OpenTelaLink()
//   - client://module/explorer/<scid>     			  Tell a client to open a specific module or page with args
func ParseTELALink(telaLink string) (target string, args []string, err error) {
	split := strings.Split(telaLink, "://")
	if len(split) < 2 {
		err = fmt.Errorf("link %q missing target", telaLink)
		return
	}

	target = split[0]
	args = strings.Split(split[1], "/")

	return
}

// Formats dvm.SmartContract to string, removes whitespace and comments from original
func FormatSmartContract(sc dvm.SmartContract, code string) (formatted string, err error) {
	var sb strings.Builder

	// Get function names to maintain code order
	ordered := GetSmartContractFuncNames(code)
	indexEnd := len(ordered) - 1

	for i, name := range ordered {
		// Write new line after each function
		if i > 0 {
			sb.WriteString("\n")
		}

		function, ok := sc.Functions[name]
		if !ok {
			err = fmt.Errorf("function %q does not exist in map", name)
			return
		}

		// Write the function signature
		sb.WriteString(fmt.Sprintf("Function %s(", name))
		for i, param := range function.Params {
			if i > 0 {
				sb.WriteString(", ")
			}

			sb.WriteString(fmt.Sprintf("%s ", param.Name))
			switch param.Type {
			case 0x4:
				sb.WriteString("Uint64")
			case 0x5:
				sb.WriteString("String")
			default:
				err = fmt.Errorf("invalid param type")
				return
			}
		}

		// Write the return type
		sb.WriteString(") ")
		switch function.ReturnValue.Type {
		case 0x3:
			sb.WriteString("")
		case 0x4:
			sb.WriteString("Uint64")
		case 0x5:
			sb.WriteString("String")
		default:
			err = fmt.Errorf("invalid return type")
			return
		}
		sb.WriteString("\n")

		// Write the function body
		for _, lineNumber := range function.LineNumbers {
			sb.WriteString(fmt.Sprintf("%d ", lineNumber))

			line := function.Lines[lineNumber]
			lineEnd := len(line) - 1

			var skip bool
			for i, part := range line {
				// Skip if we wrote part in last iteration
				if skip {
					skip = false
					continue
				}

				p := strings.TrimSpace(part)

				// Add double operator
				if i+1 <= lineEnd {
					operator, isOp := isOperator(p + strings.TrimSpace(line[i+1]))
					if isOp {
						sb.WriteString(fmt.Sprintf(" %s ", operator))
						skip = true
						continue
					}
				}

				if i > 0 {
					// Add a space after specific tokens
					last := strings.TrimSpace(line[i-1])
					lastLower := strings.ToLower(last)
					if last == "RETURN" || last == "," || last == "IF" || last == "GOTO" ||
						lastLower == "dim" || lastLower == "as" || lastLower == "let" {
						sb.WriteString(" ")
					}

					// Add single operator
					operator, isOp := isOperator(p)
					if isOp {
						sb.WriteString(fmt.Sprintf(" %s ", operator))
						continue
					}

					// Add a space before specific tokens
					if p == "IF" || p == "THEN" || p == "ELSE" || p == "GOTO" ||
						strings.ToLower(p) == "as" {
						sb.WriteString(" ")
					}
				}

				sb.WriteString(p)
			}

			sb.WriteString("\n")
		}

		// End of function
		sb.WriteString("End Function")

		// Write new line after all function ends except last one
		if i != indexEnd {
			sb.WriteString("\n")
		}
	}

	return sb.String(), nil
}

// Parse DVM code string and return its functions names in written order
func GetSmartContractFuncNames(code string) (names []string) {
	multilineComment := false
	for _, line := range strings.Split(code, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Skip commented lines
		if strings.HasPrefix(trimmed, "//") {
			continue
		}

		if strings.HasPrefix(trimmed, "/*") {
			multilineComment = true
			continue
		}

		if strings.HasSuffix(trimmed, "*/") {
			multilineComment = false
			continue
		}

		if multilineComment {
			continue
		}

		if strings.Contains(trimmed, "Function") {
			key := strings.Split(trimmed, " ")
			if len(key) > 1 && strings.ToLower(strings.TrimSpace(key[0])) != "end" {
				name := strings.Split(strings.TrimSpace(key[1]), "(")
				if name[0] != "" {
					names = append(names, name[0])
				}
			}
		}
	}

	return
}

// EqualSmartContract compares if c is equal to v by parsing function lines and parts,
// it compares all functions other than InitializePrivate/Initialize,
// contract returned is dvm.SmartContract of v when equal
func EqualSmartContracts(c, v string) (contract dvm.SmartContract, err error) {
	sc1, _, err := dvm.ParseSmartContract(c)
	if err != nil {
		err = fmt.Errorf("could not parse c contract")
		return
	}

	sc2, _, err := dvm.ParseSmartContract(v)
	if err != nil {
		err = fmt.Errorf("could not parse v contract")
		return
	}

	if len(sc1.Functions) != len(sc2.Functions) {
		err = fmt.Errorf("functions are not equal")
		return
	}

	for name, function := range sc1.Functions {
		if _, ok := sc2.Functions[name]; !ok {
			err = fmt.Errorf("missing function name: %s", name)
			return
		}

		// Skip Initialize funcs as they have custom defined fields
		if name != DVM_FUNC_INIT_PRIVATE && name != DVM_FUNC_INIT {
			for li, line := range function.Lines {
				if _, ok := sc2.Functions[name].Lines[li]; !ok {
					err = fmt.Errorf("line index missing: %d", li)
					return
				}

				if len(line) != len(sc2.Functions[name].Lines[li]) {
					err = fmt.Errorf("lines are different: %d", li)
					return
				}

				for pi, part := range line {
					if part != sc2.Functions[name].Lines[li][pi] {
						err = fmt.Errorf("line parts are different: %d", li)
						return
					}
				}
			}
		}
	}

	contract = sc2

	return
}
