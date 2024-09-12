package tela

import (
	"fmt"
	"strings"

	"github.com/civilware/tela/logger"
	"github.com/deroproject/derohe/dvm"

	_ "embed"
)

// TELA-MOD structure
type MOD struct {
	Name          string        // The name of the MOD
	Tag           string        // The identifying tag for the MOD
	Description   string        // Description of what the MOD does
	FunctionCode  func() string // The embedded DVM code
	FunctionNames []string      // The function names from the DVM code
}

// MODClass structure is used to group similarly functioned MODs
type MODClass struct {
	Name  string         // The name of the MODClass
	Tag   string         // The identifying tag for the MODClass
	Rules []MODClassRule // Any rules for the MODClass
}

// MODClassRule structure contains data for the rules used withing a MODClass
type MODClassRule struct {
	Name        string                         // Name of the rule
	Description string                         // Description of what the rule is enforcing
	Verify      func([]string, MODClass) error // Verification function checking that the parameters of the rule are met
}

// MODs contains the package's TELA-MOD data structures and access,
// all MODClasses are held in the mods variable and can be separated by its index
type MODs struct {
	mods    []MOD      // MOD data
	classes []MODClass // classes hold the MODClass info
	index   []int      // index is the index for the MODClasses: vs, tx...
	rules   []MODClassRule
}

// Access the package's initialized TELA-MOD data
var Mods MODs

// Tag returns a TELA-MOD tag string by index
func (m *MODs) Tag(i int) (tag string) {
	if i < len(m.mods) {
		tag = m.mods[i].Tag
	}

	return
}

// Functions returns the TELA-MOD function code and names for the given tag
func (m *MODs) Functions(tag string) (functionCode string, functionNames []string) {
	for _, m := range m.mods {
		if m.Tag == tag {
			return m.FunctionCode(), m.FunctionNames
		}
	}

	return
}

// GetAllMods returns all TELA-MODs
func (m *MODs) GetAllMods() (mods []MOD) {
	mods = m.mods
	return
}

// GetAllClasses returns all MODClasses
func (m *MODs) GetAllClasses() (classes []MODClass) {
	classes = m.classes
	return
}

// GetMod returns a TELA-MOD by its tag string
func (m *MODs) GetMod(tag string) (mod MOD) {
	for _, m := range m.mods {
		if m.Tag == tag {
			mod = m
			return
		}
	}

	return
}

// GetClass returns a MODClass from a MOD tag
func (m *MODs) GetClass(modTag string) (class MODClass) {
	for _, c := range m.classes {
		if strings.HasPrefix(modTag, c.Tag) {
			return c
		}
	}

	return
}

// GetRules returns all of the MODClassesRules
func (m *MODs) GetRules() (rules []MODClassRule) {
	rules = m.rules
	return
}

// Index returns the MODClass index
func (m *MODs) Index() (ci []int) {
	ci = m.index
	return
}

// Function HowToAddMODs() Uint64
// 10 Create the function set bas file in /TELA-MOD-1/
// 20 Embed the function set bas file below as a new variable
// 30 In initMods() create and append the new MOD in its MODClass's []MOD
// 35 If a new MODClass is required, it can be added similarly to the existing MODClasses
// 50 RETURN 0
// End Function

// // Embed the TELA-MOD smart contract function sets

//go:embed */vs/TELA-MOD-1-VSOO.bas
var TELA_MOD_1_VSOO string

//go:embed */vs/TELA-MOD-1-VSOOIM.bas
var TELA_MOD_1_VSOOIM string

//go:embed */vs/TELA-MOD-1-VSPUBSU.bas
var TELA_MOD_1_VSPUBSU string

//go:embed */vs/TELA-MOD-1-VSPUBOW.bas
var TELA_MOD_1_VSPUBOW string

//go:embed */vs/TELA-MOD-1-VSPUBIM.bas
var TELA_MOD_1_VSPUBIM string

//go:embed */tx/TELA-MOD-1-TXDWA.bas
var TELA_MOD_1_TXDWA string

//go:embed */tx/TELA-MOD-1-TXDWD.bas
var TELA_MOD_1_TXDWD string

//go:embed */tx/TELA-MOD-1-TXTO.bas
var TELA_MOD_1_TXTO string

// Initialize TELA-MOD data
func initMods() {
	// Initialize all the package's MODClass rules
	{
		rules := []MODClassRule{
			{
				Name:        "Single MOD",
				Description: "Only one MOD from this MODClass can be used at a time",
				Verify: func(tags []string, c MODClass) (err error) {
					if hasMultipleClassTags(tags, c.Tag) {
						err = fmt.Errorf("conflicting tags for %q MODClass", strings.ToLower(c.Name))
					}
					return
				},
			},
			{
				Name:        "Multi MOD",
				Description: "Multiple MODs from this MODClass can be used simultaneously",
				Verify: func(tags []string, c MODClass) (err error) {
					// All MODs are valid so return nil
					return
				},
			},
		}

		// When new rules are added, any conflict handling will go in verifyRules()
		Mods.rules = append(Mods.rules, rules...)
	}

	// // Variable Store MODClass
	{
		// Initialize the MODClass
		variableStoreMODClass := MODClass{
			Name:  "Variable store",
			Tag:   "vs",
			Rules: []MODClassRule{Mods.rules[0]},
		}

		variableStoreMODs := []MOD{
			{
				Name:          fmt.Sprintf("%s %s", variableStoreMODClass.Name, "owner only"),
				Tag:           variableStoreMODClass.NewTag("oo"),
				Description:   "Allows the owner of the smart contract to set and delete custom variable stores",
				FunctionCode:  func() string { return TELA_MOD_1_VSOO },
				FunctionNames: []string{"SetVar", "DeleteVar"},
			},
			{
				Name:          fmt.Sprintf("%s %s", variableStoreMODClass.Name, "owner only immutable"),
				Tag:           variableStoreMODClass.NewTag("ooim"),
				Description:   "Allows the owner of the smart contract to set custom variable stores which cannot be changed or deleted",
				FunctionCode:  func() string { return TELA_MOD_1_VSOOIM },
				FunctionNames: []string{"SetVar"},
			},
			{
				Name:          fmt.Sprintf("%s %s", variableStoreMODClass.Name, "public single use"),
				Tag:           variableStoreMODClass.NewTag("pubsu"),
				Description:   "Allows all wallets to store variables which the wallet cannot change, the owner of the smart contract can set and delete all variables",
				FunctionCode:  func() string { return TELA_MOD_1_VSPUBSU },
				FunctionNames: []string{"SetVar", "DeleteVar"},
			},
			{
				Name:          fmt.Sprintf("%s %s", variableStoreMODClass.Name, "public overwrite"),
				Tag:           variableStoreMODClass.NewTag("pubow"),
				Description:   "Allows all wallets to store variables which the wallet can overwrite, the owner of the smart contract can set and delete all variables",
				FunctionCode:  func() string { return TELA_MOD_1_VSPUBOW },
				FunctionNames: []string{"SetVar", "DeleteVar"},
			},
			{
				Name:          fmt.Sprintf("%s %s", variableStoreMODClass.Name, "public immutable"),
				Tag:           variableStoreMODClass.NewTag("pubim"),
				Description:   "Allows all wallets to store variables which cannot be changed or deleted from the smart contract",
				FunctionCode:  func() string { return TELA_MOD_1_VSPUBIM },
				FunctionNames: []string{"SetVar"},
			},
			// New MOD for variable store MODClass would go here
		}

		// Add the variable store MODClass and its MODs
		Mods.Add(variableStoreMODClass, variableStoreMODs)
	}

	// // Transfers MODClass
	{
		transferMODClass := MODClass{
			Name:  "Transfers",
			Tag:   "tx",
			Rules: []MODClassRule{Mods.rules[1]},
		}

		transferMODs := []MOD{
			{
				Name:          "Deposit and withdraw DERO",
				Tag:           transferMODClass.NewTag("dwd"),
				Description:   "Stores DERO deposits and owner can withdraw DERO from the smart contract",
				FunctionCode:  func() string { return TELA_MOD_1_TXDWD },
				FunctionNames: []string{"DepositDero", "WithdrawDero"},
			},
			{
				Name:          "Deposit and withdraw assets",
				Tag:           transferMODClass.NewTag("dwa"),
				Description:   "Stores asset deposits and owner can withdraw tokens from the smart contract",
				FunctionCode:  func() string { return TELA_MOD_1_TXDWA },
				FunctionNames: []string{"DepositAsset", "WithdrawAsset"},
			},
			{
				Name:          "Transfer ownership",
				Tag:           transferMODClass.NewTag("to"),
				Description:   "Transfer the ownership of the smart contract to another wallet",
				FunctionCode:  func() string { return TELA_MOD_1_TXTO },
				FunctionNames: []string{"TransferOwnership", "ClaimOwnership"},
			},
			// New MODs for transfers MODClass would go here
		}

		// Add the transfers MODClass and its MODs
		Mods.Add(transferMODClass, transferMODs)
	}

	// This is checked each Add but we should ensure again there is no conflicts with the initialized MODClasses and MODs
	err := Mods.Verify()
	if err != nil {
		logger.Fatalf("[TELA] MODs: %s\n", err)
	}
}

// verifyRules is used to check that the initialized MODClassRules do not conflict with each other
func (mc *MODClass) verifyRules() (err error) {
	var hasSingleMODRule bool
	for _, r := range mc.Rules {
		if r.Name == "Single MOD" {
			hasSingleMODRule = true
			continue
		}

		if hasSingleMODRule && r.Name == "Multi MOD" {
			err = fmt.Errorf("conflicting rules for MODClass %q", mc.Name)
			return
		}
	}

	return
}

// Verify that all MODs present have the appropriate data and that there are no conflicting MODs or MODClasses
func (m *MODs) Verify() (err error) {
	var classTags, modTags, modNames []string
	for _, c := range m.classes {
		err = c.verifyRules()
		if err != nil {
			return
		}

		classTags = append(classTags, c.Tag)
	}

	duplicate, found := hasDuplicateString(classTags)
	if found {
		err = fmt.Errorf("class tag %q cannot be duplicated", duplicate)
		return
	}

	if len(m.classes) != len(m.index) {
		err = fmt.Errorf("invalid MODClass index: %d/%d", len(m.classes), len(m.index))
		return
	}

	for _, mod := range m.mods {
		class := m.GetClass(mod.Tag)
		if class.Tag == "" {
			err = fmt.Errorf("could not get MODClass tag for %q", mod.Tag)
			return
		}

		if mod.FunctionCode == nil {
			err = fmt.Errorf("missing function code for %q", mod.Name)
			return
		}

		var sc dvm.SmartContract
		sc, _, err = dvm.ParseSmartContract(mod.FunctionCode())
		if err != nil {
			err = fmt.Errorf("function code for %q is invalid: %s", mod.Name, err)
			return
		}

		if len(mod.FunctionNames) != len(sc.Functions) {
			err = fmt.Errorf("missing function names for %q  %d/%d", mod.Name, len(mod.FunctionNames), len(sc.Functions))
			return
		}

		modTags = append(modTags, mod.Tag)
		modNames = append(modNames, mod.Name)
	}

	duplicate, found = hasDuplicateString(modTags)
	if found {
		err = fmt.Errorf("tag %q cannot be duplicated", duplicate)
		return
	}

	duplicate, found = hasDuplicateString(modNames)
	if found {
		err = fmt.Errorf("name %q cannot be duplicated", duplicate)
		return
	}

	return
}

// Add a new MODClass and its MODs if they are valid
func (m *MODs) Add(class MODClass, mods []MOD) (err error) {
	addMod := MODs{
		mods:    append(mods, m.mods...),
		classes: append([]MODClass{class}, m.classes...),
		index:   append(m.index, len(mods)+len(m.mods)),
	}

	err = addMod.Verify()
	if err != nil {
		return
	}

	// Add the new MODClass
	m.classes = append(m.classes, class)
	// Add the new MODs
	m.mods = append(m.mods, mods...)
	// Complete the MODClass by storing its index
	m.index = append(m.index, len(m.mods))

	return
}

// NewModTag returns a modTag string from the given tag elements
func NewModTag(tags []string) (modTag string) {
	modTag = strings.Join(tags, ",")
	return
}

// NewTag returns a MOD tag prefixed with the MODClass tag
func (mc *MODClass) NewTag(tag string) string {
	return fmt.Sprintf("%s%s", mc.Tag, tag)
}

// Check if any duplicate elements exists
func hasDuplicateString(check []string) (duplicate string, found bool) {
	have := map[string]bool{}
	for _, ele := range check {
		if have[ele] {
			found = true
			duplicate = ele
			return
		}

		have[ele] = true
	}

	return
}

// Check tag prefixes ensuring that only one tag from the MODClass is used within the tags
func hasMultipleClassTags(tags []string, prefix string) bool {
	have := 0
	for _, tag := range tags {
		if strings.HasPrefix(tag, prefix) {
			have++
		}

		if have > 1 {
			return true
		}
	}

	return false
}

// ModTagsAreValid parses a modTag string formatted as "tag,tag,tag" returning its tags if all tags are valid
// and there is no conflicting tags. If an empty modTag is passed it will return no error and nil tags
func (m *MODs) TagsAreValid(modTag string) (tags []string, err error) {
	checkTags := strings.Split(modTag, ",")
	if checkTags[0] == "" {
		return
	}

	duplicate, found := hasDuplicateString(checkTags)
	if found {
		err = fmt.Errorf("found duplicate tag %q", duplicate)
		return
	}

	var modExists bool
	for _, tag := range checkTags {
		modExists = false
		for _, m := range m.mods {
			if m.Tag == tag {
				modExists = true
				break
			}
		}

		if !modExists {
			err = fmt.Errorf("tag %q does not exist", tag)
			return
		}
	}

	// Go through all MODClasses ensuring that the tags abide by MODClassRules
	for _, c := range m.classes {
		for _, r := range c.Rules {
			err = r.Verify(checkTags, c)
			if err != nil {
				return
			}
		}
	}

	tags = checkTags

	return
}

// injectMOD injects a TELA-MOD's functions into the code of a smart contract and returns the new smart contract and new code
func (m *MODs) injectMOD(mod, code string) (modSC dvm.SmartContract, modCode string, err error) {
	modSC, _, err = dvm.ParseSmartContract(code)
	if err != nil {
		err = fmt.Errorf("could not parse MOD base code: %s", err)
		return
	}

	// Find the MOD code to be injected
	modCodeSet, functionNames := m.Functions(mod)
	if modCodeSet == "" {
		err = fmt.Errorf("could not find MOD %q code: %s", mod, err)
		return
	}

	sc, _, err := dvm.ParseSmartContract(modCodeSet)
	if err != nil {
		err = fmt.Errorf("could not parse MOD %q code: %s", mod, err)
		return
	}

	// Inject the new MOD functions into the smart contract
	for _, name := range functionNames {
		modSC.Functions[name] = sc.Functions[name]
	}

	modCode = fmt.Sprintf("%s\n%s", code, modCodeSet)

	return
}

// InjectMODs parses the modTag ensuring all tags are valid before injecting the TELA-MOD code for the given MODs in the modTag
func (m *MODs) InjectMODs(modTag, code string) (modSC dvm.SmartContract, modCode string, err error) {
	tags, err := m.TagsAreValid(modTag)
	if err != nil {
		err = fmt.Errorf("modTag has invalid MODs: %s", err)
		return
	}

	modCode = code // tags could be nil otherwise start injecting mods
	for _, mod := range tags {
		_, modCode, err = m.injectMOD(mod, modCode)
		if err != nil {
			return
		}
	}

	return
}
