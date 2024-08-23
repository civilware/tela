package tela

import (
	"fmt"
	"strings"
)

// Standard SC header value stores
type Headers struct {
	NameHdr  string `json:"nameHdr"`  // On-chain name of SC. For TELA-DOCs, they are recreated using this as the file name, it should include the file extension
	DescrHdr string `json:"descrHdr"` // On-chain description of DOC, INDEX or Asset SC
	IconHdr  string `json:"iconHdr"`  // On-chain icon URL, (size 100x100)
}

// DERO signature
type Signature struct {
	CheckC string `json:"checkC"` // C signature value
	CheckS string `json:"checkS"` // S signature value
}

// Standard SC header keys
type Header string

const (
	HEADER_NAME         Header = `"nameHdr"`
	HEADER_DESCRIPTION  Header = `"descrHdr"`
	HEADER_ICON_URL     Header = `"iconURLHdr"`
	HEADER_CHECK_C      Header = `"fileCheckC"`
	HEADER_CHECK_S      Header = `"fileCheckS"`
	HEADER_DURL         Header = `"dURL"`
	HEADER_DOCUMENT     Header = `"DOC` // append with Number()
	HEADER_SUBDIR       Header = `"subDir"`
	HEADER_DOCTYPE      Header = `"docType"`
	HEADER_COLLECTION   Header = `"collection"`
	HEADER_TYPE         Header = `"typeHdr"`
	HEADER_TAGS         Header = `"tagsHdr"`
	HEADER_FILE_URL     Header = `"fileURL"`
	HEADER_SIGN_URL     Header = `"fileSignURL"`
	HEADER_COVER_URL    Header = `"coverURL"`
	HEADER_ART_FEE      Header = `"artificerFee"`
	HEADER_ROYALTY      Header = `"royalty"`
	HEADER_OWNER        Header = `"owner"`
	HEADER_OWNER_UPDATE Header = `"ownerCanUpdate"`
)

// Trim any `"` from Header and return string
func (h Header) Trim() string {
	return strings.Trim(string(h), `"`)
}

// Returns if Header can be appended. Headers ending in `"` or with len < 2 will return false
func (h Header) CanAppend() bool {
	i := len(h) - 1
	if i < 1 {
		return false
	}

	return h[i] != byte(0x22)
}

// Append a number to Header if applicable, otherwise return unchanged Header
func (h Header) Number(i int) Header {
	if !h.CanAppend() {
		return h
	}

	return Header(fmt.Sprintf(`%s%d"`, h, i))
}
