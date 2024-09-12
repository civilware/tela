package tela

import (
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/civilware/tela/logger"
	"github.com/civilware/tela/shards"
	"github.com/deroproject/derohe/cryptography/crypto"
	"github.com/deroproject/derohe/dvm"
	"github.com/deroproject/derohe/globals"
	"github.com/deroproject/derohe/rpc"
	"github.com/deroproject/derohe/walletapi"
	"github.com/stretchr/testify/assert"
)

// This package test requires the SIMULATOR to be running to test installed TELA apps,
// if validSCIDs is left empty the test will attempt to create and install apps using
// the templates contracts and app code found in tela_tests/

// Valid TELA application contracts installed on SIMULATOR, can be manually input, see inline comments for what the testers have it targeting
var validSCIDs = [5]string{
	"",
	"",
	"",
	"", // Has subDirs
	"", // Has subDirs
}

// Invalid TELA application contracts installed on SIMULATOR, inline comments should be what it is targeting
var invalidSCIDS = [6]string{
	"", // Invalid docType
	"", // No docType

	// These below should not need to change
	"",     // s == ""
	"scid", // len(s) != 64
	"c4d7bbdaaf9344f4c351e72d0b2145b4235402c89510101e0500f43969fd1387", // Does not exists
	"b879b0ff01d78841d61e9770fd18436d8b9afce59302c77a786272e7422c15f6", // Does not exists
}

var telaApps = []INDEX{
	{
		DURL: "test0.tela",
		Headers: Headers{
			NameHdr:  "TELA Test App0",
			DescrHdr: "Zero TELA test app with wallet connection",
			IconHdr:  "icon.url",
		},
	},
	{
		DURL: "test1.tela",
		Headers: Headers{
			NameHdr:  "TELA Test App1",
			DescrHdr: "One TELA test app with wallet connection",
			IconHdr:  "icon.url",
		},
	},
	{
		DURL: "test2.tela",
		Headers: Headers{
			NameHdr:  "TELA Test App2",
			DescrHdr: "Two TELA test app with wallet connection",
			IconHdr:  "icon.url",
		},
	},
	{
		DURL: "test3.tela",
		Headers: Headers{
			NameHdr:  "TELA Test App3",
			DescrHdr: "Three TELA test app with wallet connection",
			IconHdr:  "icon.url",
		},
	},
	{
		DURL: "test4.tela",
		Headers: Headers{
			NameHdr:  "TELA Test App4",
			DescrHdr: "Four TELA test app with wallet connection and subDir",
			IconHdr:  "icon.url",
		},
	},
	{
		DURL: "test5.tela",
		Headers: Headers{
			NameHdr:  "TELA Test App5",
			DescrHdr: "Five TELA test app, invalid language",
		},
	},
	{
		DURL: "test6.tela",
		Headers: Headers{
			NameHdr:  "TELA Test App6",
			DescrHdr: "Six TELA test app, no docType STORE",
		},
	},
}

// TELA document data
var telaDocs = []struct {
	filePath string // Where the test docType code file is stored
	DOC
}{
	// // DOCs for app1, serves everything from root at http://localhost:port
	// datashards
	// |-- tela
	// |   |-- dURL
	// |       |-- index.html
	// |       |-- style.css
	// |       |-- main.js
	{
		filePath: filepath.Join(testDir, "app1", "index.html"),
		DOC: DOC{
			DocType: DOC_HTML,
			SubDir:  "",
			DURL:    "test.tela",
			Headers: Headers{
				NameHdr:  "index.html",
				DescrHdr: "TELA test HTML index file",
				IconHdr:  "icon.url",
			},
			Signature: Signature{
				CheckC: "c4d7bbdaaf9344f4c351e72d0b2145b4235402c89510101e0500f43969fd1387",
				CheckS: "b879b0ff01d78841d61e9770fd18436d8b9afce59302c77a786272e7422c15f6",
			},
		},
	},
	{
		filePath: filepath.Join(testDir, "app1", "main.js"),
		DOC: DOC{
			DocType: DOC_JS,
			SubDir:  "",
			DURL:    "test.tela",
			Headers: Headers{
				NameHdr:  "main.js",
				DescrHdr: "TELA test JavaScript file",
				IconHdr:  "icon.url",
			},
			Signature: Signature{
				CheckC: "c4d7bbdaaf9344f4c351e72d0b2145b4235402c89510101e0500f43969fd1387",
				CheckS: "b879b0ff01d78841d61e9770fd18436d8b9afce59302c77a786272e7422c15f6",
			},
		},
	},
	{
		filePath: filepath.Join(testDir, "app1", "style.css"),
		DOC: DOC{
			DocType: DOC_CSS,
			SubDir:  "",
			DURL:    "test.tela",
			Headers: Headers{
				NameHdr:  "style.css",
				DescrHdr: "TELA test style sheet",
				IconHdr:  "icon.url",
			},
			Signature: Signature{
				CheckC: "c4d7bbdaaf9344f4c351e72d0b2145b4235402c89510101e0500f43969fd1387",
				CheckS: "b879b0ff01d78841d61e9770fd18436d8b9afce59302c77a786272e7422c15f6",
			},
		},
	},
	// // DOCs for app2, serves from a subDir at http://localhost:port/subDir/
	// datashards
	// |-- tela
	// |   |-- dURL
	// |       |-- subDir
	// |          |-- index.html
	// |          |-- js
	// |              |-- main.js
	// |          |-- css
	// |              |-- style.css
	{
		filePath: filepath.Join(testDir, "app2", "index.html"),
		DOC: DOC{
			DocType: DOC_HTML,
			SubDir:  "telaSub",
			DURL:    "test.tela",
			Headers: Headers{
				NameHdr:  "index.html",
				DescrHdr: "TELA test HTML index subDir file",
				IconHdr:  "icon.url",
			},
			Signature: Signature{
				CheckC: "c4d7bbdaaf9344f4c351e72d0b2145b4235402c89510101e0500f43969fd1387",
				CheckS: "b879b0ff01d78841d61e9770fd18436d8b9afce59302c77a786272e7422c15f6",
			},
		},
	},
	{
		filePath: filepath.Join(testDir, "app1", "main.js"), // Can use app1 file here
		DOC: DOC{
			DocType: DOC_JS,
			SubDir:  "telaSub/js",
			DURL:    "test.tela",
			Headers: Headers{
				NameHdr:  "main.js",
				DescrHdr: "TELA test JavaScript subDir file",
				IconHdr:  "icon.url",
			},
			Signature: Signature{
				CheckC: "c4d7bbdaaf9344f4c351e72d0b2145b4235402c89510101e0500f43969fd1387",
				CheckS: "b879b0ff01d78841d61e9770fd18436d8b9afce59302c77a786272e7422c15f6",
			},
		},
	},
	{
		filePath: filepath.Join(testDir, "app1", "style.css"), // Can use app1 file here
		DOC: DOC{
			DocType: DOC_CSS,
			SubDir:  "telaSub/css",
			DURL:    "test.tela",
			Headers: Headers{
				NameHdr:  "style.css",
				DescrHdr: "TELA test subDir style sheet",
				IconHdr:  "icon.url",
			},
			Signature: Signature{
				CheckC: "c4d7bbdaaf9344f4c351e72d0b2145b4235402c89510101e0500f43969fd1387",
				CheckS: "b879b0ff01d78841d61e9770fd18436d8b9afce59302c77a786272e7422c15f6",
			},
		},
	},
	// // DOC to test mark down install
	{
		filePath: filepath.Join("README.md"),
		DOC: DOC{
			DocType: DOC_MD,
			SubDir:  "",
			DURL:    "test.tela",
			Headers: Headers{
				NameHdr:  "README.md",
				DescrHdr: "TELA README file",
				IconHdr:  "icon.url",
			},
			Signature: Signature{
				CheckC: "c4d7bbdaaf9344f4c351e72d0b2145b4235402c89510101e0500f43969fd1387",
				CheckS: "b879b0ff01d78841d61e9770fd18436d8b9afce59302c77a786272e7422c15f6",
			},
		},
	},
	// // DOC for invalid app, language store invalid
	{
		DOC: DOC{
			DocType: "TELA-html-1", // Case sensitive
			SubDir:  "",
			DURL:    "test.tela",
			Headers: Headers{
				NameHdr:  "index.html",
				DescrHdr: "TELA languages are cases sensitive",
				IconHdr:  "icon.url",
			},
			Signature: Signature{
				CheckC: "c4d7bbdaaf9344f4c351e72d0b2145b4235402c89510101e0500f43969fd1387",
				CheckS: "b879b0ff01d78841d61e9770fd18436d8b9afce59302c77a786272e7422c15f6",
			},
		},
	},
	// // DOC for invalid app, no docType store
	{
		DOC: DOC{
			SubDir: "",
			DURL:   "test.tela",
			Headers: Headers{
				NameHdr:  "index.html",
				DescrHdr: "TELA contract missing docType STORE",
				IconHdr:  "icon.url",
			},
			Signature: Signature{
				CheckC: "c4d7bbdaaf9344f4c351e72d0b2145b4235402c89510101e0500f43969fd1387",
				CheckS: "b879b0ff01d78841d61e9770fd18436d8b9afce59302c77a786272e7422c15f6",
			},
		},
	},
}

// Simulator wallet seeds
var walletSeeds = []string{
	"2af6630905d73ee40864bd48339f297908a0731a6c4c6fa0a27ea574ac4e4733",
	"193faf64d79e9feca5fce8b992b4bb59b86c50f491e2dc475522764ca6666b6b",
	"2e49383ac5c938c268921666bccfcb5f0c4d43cd3ed125c6c9e72fc5620bc79b",
	"1c8ee58431e21d1ef022ccf1f53fec36f5e5851d662a3dd96ced3fc155445120",
	"19182604625563f3ff913bb8fb53b0ade2e0271ca71926edb98c8e39f057d557",
	"2a3beb8a57baa096512e85902bb5f1833f1f37e79f75227bbf57c4687bfbb002",
	"055e43ebff20efff612ba6f8128caf990f2bf89aeea91584e63179b9d43cd3ab",
	"2ccb7fc12e867796dd96e246aceff3fea1fdf78a28253c583017350034c31c81",
	"279533d87cc4c637bf853e630480da4ee9d4390a282270d340eac52a391fd83d",
	"03bae8b71519fe8ac3137a3c77d2b6a164672c8691f67bd97548cb6c6f868c67",
	"2b9022d0c5ee922439b0d67864faeced65ebce5f35d26e0ee0746554d395eb88",
	"1a63d5cf9955e8f3d6cecde4c9ecbd538089e608741019397824dc6a2e0bfcc1",
	"10900d25e7dc0cec35fcca9161831a02cb7ed513800368529ba8944eeca6e949",
	"2ac9a8984c988fcb54b261d15bc90b5961d673bffa5ff41c8250c7e262cbd606",
	"040572cec23e6df4f686192b776c197a50591836a3dd02ba2e4a7b7474382ccd",
	"2b2b029cfbc5d08b5d661e6fa444102d387780bec088f4dd41a4a537bf9762af",
}

var mainPath string
var sleepFor = time.Millisecond * 1000
var testDir = "tela_tests"
var nameservice = "0000000000000000000000000000000000000000000000000000000000000001"
var scDoesNotExist = "c4d7bbdaaf9344f4c351e72d0b2145b4235402c89510101e0500f43969fd1387"

func TestMain(m *testing.M) {
	dir, err := os.Getwd()
	if err != nil {
		logger.Fatalf("[TELA] %s\n", err)
	}

	var debugFlag bool
	flag.BoolVar(&debugFlag, "debug", false, "enables debug mode for tests")

	for _, arg := range os.Args {
		if arg == "--debug" {
			globals.Arguments["--debug"] = true
			logger.Debugf("[TELA] Debug mode enabled\n")
		}
	}

	mainPath = dir

	m.Run()
}

func TestTELA(t *testing.T) {
	endpoint, datashards, wallets := createTestEnvironment(t)

	var err error
	var docs [3][]string       // Organize DOC scids
	var shardSCIDs [2][]string // Stitch their docCode together
	var librarySCIDs []string  // Embedded library SCIDs
	var commitTXIDs []string   // TXIDs of updates
	var noCodeTXIDs []string   // TXIDs without any SC code

	successful := true
	scidsPredefined := true
	// If no user defined SCs have been installed and set before tests, install them here
	if len(validSCIDs[0]) != 64 {
		t.Run("Install", func(t *testing.T) {
			scidsPredefined = false

			for i, doc := range telaDocs {
				var code string
				if i < 7 {
					var err error
					// Use template code for valid apps
					code, err = readFile(doc.filePath)
					if err != nil {
						t.Fatalf("Could not read %s file: %s", doc.NameHdr, err)
					}
				}

				// Install some DOCs as anon and some with owner
				ringsize := uint64(2)
				if i > 2 {
					ringsize = 16
				}

				doc.Code = code
				tx, err := retry(t, fmt.Sprintf("DOC %d install", i), func() (string, error) {
					return Installer(wallets[i], ringsize, &doc.DOC)
				})

				assert.NoError(t, err, "Install %d %s should not error: %s", i, doc.NameHdr, err)

				_, err = retry(t, fmt.Sprintf("confirming install TX %s", tx), func() (string, error) {
					return getContractCode(tx, endpoint)
				})
				if err != nil {
					t.Fatalf("Could not confirm DOC %d TX %s: %s", i, tx, err)
				}

				t.Logf("Simulator DOC %d SC installed %s: %s", i, doc.NameHdr, tx)

				if i < 3 {
					docs[0] = append(docs[0], tx)
				} else if i < 6 {
					docs[1] = append(docs[1], tx)
				} else {
					docs[2] = append(docs[2], tx)
				}

				time.Sleep(sleepFor / 4)
			}

			if !successful {
				t.Fatalf("Could not create all DOCs")
			}

			for i, app := range telaApps {
				doc := docs[0]
				if i == 5 {
					d := docs[2][1]
					doc = []string{d, d, d}
				} else if i == 6 {
					d := docs[2][2]
					doc = []string{d, d, d}
				} else if i > 2 {
					doc = docs[1]
				}

				// Go through some different MODs when installing
				cycleModTags := 1
				if i%2 == 0 {
					cycleModTags = 2
				}

				install := &INDEX{
					Mods: Mods.Tag(cycleModTags),
					DOCs: doc,
					DURL: app.DURL,
					Headers: Headers{
						NameHdr:  app.NameHdr,
						DescrHdr: app.DescrHdr,
						IconHdr:  app.IconHdr,
					},
				}

				// Install some INDEXs as anon, the first one should have owner (RS2) and will be tested for successful update later
				ringsize := uint64(2)
				if i > 2 {
					ringsize = 16
				}

				tx, err := retry(t, fmt.Sprintf("INDEX %d install", i), func() (string, error) {
					return Installer(wallets[i+9], ringsize, install) // +9 to skip all wallets used in DOC installs
				})

				assert.NoError(t, err, "Install %d %s should not have error: %s", i, app.NameHdr, err)

				_, err = retry(t, fmt.Sprintf("confirming install TX %s", tx), func() (string, error) {
					return getContractCode(tx, endpoint)
				})
				if err != nil {
					t.Fatalf("Could not confirm INDEX %d TX %s: %s", i, tx, err)
				}

				t.Logf("Simulator INDEX %d SC installed %s: %s", i, app.NameHdr, tx)

				if i < 5 {
					validSCIDs[i] = tx
				} else if i == 5 {
					invalidSCIDS[0] = tx
				} else if i == 6 {
					invalidSCIDS[1] = tx
				}

				time.Sleep(sleepFor / 4)
			}

			// Test size limits of smart contract installs and updates
			indexLimits := &INDEX{
				DURL: "update-limit.tela",
				Headers: Headers{
					NameHdr:  "A name",
					DescrHdr: "A description",
					IconHdr:  "A icon url goes here",
				},
			}

			// SC will install and can be updated with at least this amount of DOCs
			// GetCodeSizeInKB() returns SC code size of 9.115234375 KB
			for i := 0; i < 90; i++ {
				indexLimits.DOCs = append(indexLimits.DOCs, scDoesNotExist)
			}

			tx, err := Installer(wallets[0], 2, indexLimits)
			assert.NoError(t, err, "Installing INDEX under size limit should not error: %s", err)
			t.Logf("Simulator INDEX 7 SC installed %s: %s", indexLimits.DURL, tx)

			indexLimits.SCID = tx
			time.Sleep(sleepFor)

			// Add more DOCs before updating
			for i := 0; i < 5; i++ {
				indexLimits.DOCs = append(indexLimits.DOCs, scDoesNotExist)
			}

			// Confirm installed scid can be updated
			tx, err = Updater(wallets[0], indexLimits)
			assert.NoError(t, err, "Updating INDEX under size limit should not error: %s", err)

			hash, err := varChanged(indexLimits.SCID, "hash", tx, endpoint)
			assert.NoError(t, err, "Getting hash variable should not error: %s", err)
			assert.Equal(t, tx, hash, "Contract hash should be txid of update")

			// SC will install but will not be able to be updated with this amount of DOCs
			// GetCodeSizeInKB() returns SC code size of 11.6240234375 KB
			indexLimits.DURL = "max-limit.tela"
			for i := 0; i < 24; i++ {
				indexLimits.DOCs = append(indexLimits.DOCs, scDoesNotExist)
			}

			tx, err = Installer(wallets[1], 2, indexLimits)
			assert.NoError(t, err, "Installing INDEX at max size limit should not error: %s", err)
			t.Logf("Simulator INDEX 8 SC installed %s: %s", indexLimits.DURL, tx)

			v, err := retry(t, "confirming contract size limit install", func() (string, error) {
				return getContractVar(tx, "likes", endpoint)
			})
			assert.NoError(t, err, "Confirming INDEX install at max size limit install should not error: %s", err)
			assert.NotEmpty(t, v, "Likes value should not be empty")

			// This should error as installing SC with this many DOCs will not be successful
			// GetCodeSizeInKB() returns SC code size of 11.80078125 KB
			indexLimits.DURL = "exceed-limit.tela"
			for i := 0; i < 2; i++ {
				indexLimits.DOCs = append(indexLimits.DOCs, scDoesNotExist)
			}

			tx, err = Installer(wallets[2], 2, indexLimits)
			assert.Error(t, err, "Installing exceeding max limit should error")
		})
	} else {
		t.Logf("Testing %d user defined SCIDs", len(validSCIDs))
	}

	if len(validSCIDs[0]) != 64 {
		t.Fatalf("Should have SCIDs to test")
	}

	// // Test serving TELA servers
	t.Run("ServeTELA", func(t *testing.T) {
		tela.servers = nil
		// Set the max amount of servers for number of apps being tested
		SetMaxServers(5)
		max := MaxServers()
		for _, scid := range validSCIDs {
			_, err := ServeTELA(scid, endpoint)
			assert.NoError(t, err, "TELA should not error: %s", err)
			assert.NotNil(t, tela.servers, "TELA should be running")
		}

		time.Sleep(sleepFor / 2)

		// If another port in range is in use these will fail
		assert.Equal(t, max, len(tela.servers), "All valid SCID should have been served")
		assert.Equal(t, max, len(GetServerInfo()), "All valid SCID should have info")

		assert.True(t, HasServer("test1.tela"), "This server name should be present and is not")
		assert.False(t, HasServer("none"), "This server name should not be present and is")
		assert.False(t, HasServer(""), "This server name should not be present and is")

		// All ports in use, try to start more
		os.RemoveAll(filepath.Join(datashards, "tela", telaApps[0].DURL)) // Remove some of the existing dirs to hit port not found
		os.RemoveAll(filepath.Join(datashards, "tela", telaApps[3].DURL))
		for _, scid := range validSCIDs {
			_, err := ServeTELA(scid, endpoint)
			assert.Error(t, err, "SCID %s should have error when all ports in use: %s", scid, err)
		}

		count := len(tela.servers)
		ShutdownServer(telaApps[3].DURL)
		assert.Less(t, len(tela.servers), count, "Server should have been closed")

		// Shutdown all servers
		ShutdownTELA()
		time.Sleep(sleepFor)
		assert.Nil(t, tela.servers, "Servers should be nil after shutdown")

		// Ensure all TELA files have been removed after Shutdown, test datashards dir should be present here
		files, err := os.ReadDir(datashards)
		if err != nil {
			t.Fatalf("Read directory error: %s", err)
		}

		assert.Empty(t, files, "There should be no TELA files in datashards after shutdown")
		assert.Nil(t, tela.servers, "TELA should not be running")

		// Sim wallets likely won't work as expected if valid scids have been defined before test
		if scidsPredefined {
			t.Skipf("SCIDs have been predefined, skipping Updater and Rate tests")
		}

		// Server should not exists
		telaLink := fmt.Sprintf("tela://open/%s/%s", validSCIDs[0], "main.js")
		expectedLink := "http://localhost:8082/main.js"
		link, err := OpenTELALink(telaLink, endpoint)
		assert.NoError(t, err, "OpenTELALink should not error: %s", err)
		assert.Equal(t, expectedLink, link, "Link should be the same")
		// Server already exists
		link, err = OpenTELALink(telaLink, endpoint)
		assert.NoError(t, err, "OpenTELALink should not error: %s", err)
		assert.Equal(t, expectedLink, link, "Link should be the same")
		ShutdownTELA()

		telaLink = ""
		_, err = OpenTELALink(telaLink, endpoint)
		assert.Error(t, err, "OpenTELALink with no target should error")
		telaLink = "nottela://"
		_, err = OpenTELALink(telaLink, endpoint)
		assert.Error(t, err, "OpenTELALink with invalid target should error")
		telaLink = "tela://invalid"
		_, err = OpenTELALink(telaLink, endpoint)
		assert.Error(t, err, "OpenTELALink with invalid argument should error")
		telaLink = "tela://open/" + nameservice
		_, err = OpenTELALink(telaLink, endpoint)
		assert.Error(t, err, "OpenTELALink with invalid scid should error")

		_, err = GetDOCInfo(docs[0][0], endpoint)
		assert.NoError(t, err, "Getting valid DOC should not error: %s", err)
		_, err = GetDOCInfo(docs[2][2], endpoint)
		assert.Error(t, err, "Getting DOC with invalid docType should error and did not")
		_, err = GetDOCInfo(nameservice, endpoint)
		assert.Error(t, err, "Getting invalid DOC should error and did not")
		_, err = GetDOCInfo(validSCIDs[1], "")
		assert.Error(t, err, "Getting DOC should error with invalid daemon")

		// Test Updater and Rate
		scid := validSCIDs[0]

		// Add a vs and all tx MODs
		modTags := []string{Mods.Tag(0)}
		for i := Mods.Index()[0]; i < len(Mods.GetAllMods()); i++ {
			modTags = append(modTags, Mods.Tag(i))
		}

		update := &INDEX{
			Mods: NewModTag(modTags),
			SCID: scid,
			DURL: "app.tela",
			DOCs: []string{"<scid>", "<scid>"},
			Headers: Headers{
				NameHdr:  "TELA App",
				DescrHdr: "A TELA Application",
				IconHdr:  "ICON_URL",
			},
		}

		// Contract should successfully update
		tx, err := Updater(wallets[9], update)
		assert.NoError(t, err, "Update should not have error: %s", err)
		commitTXIDs = append(commitTXIDs, tx)
		time.Sleep(sleepFor)
		hash, err := varChanged(scid, "hash", tx, endpoint)
		assert.NoError(t, err, "Getting hash variable should not error: %s", err)
		assert.Equal(t, tx, hash, "Contract hash should be txid of update")
		// Should not serve with invalid DOC <scid> STORE
		_, err = ServeTELA(scid, endpoint)
		assert.Error(t, err, "Invalid DOC was served")
		// Update to valid DOCs
		update.DOCs = []string{docs[0][0], docs[0][1], docs[0][2]}
		tx, err = Updater(wallets[9], update)
		assert.NoError(t, err, "Update should not have error: %s", err)
		commitTXIDs = append(commitTXIDs, tx)
		time.Sleep(sleepFor)
		hash, err = varChanged(scid, "hash", tx, endpoint)
		assert.NoError(t, err, "Getting hash variable should not error: %s", err)
		assert.Equal(t, tx, hash, "Contract hash should be txid of update")
		// Should not serve if !updates
		_, err = ServeTELA(scid, endpoint)
		assert.Error(t, err, "User defined no updates and SCID was served")

		// Anon contact should not update
		update.SCID = validSCIDs[3]
		_, err = Updater(wallets[10], update)
		assert.Error(t, err, "Update on anon contract should error")

		t.Logf("Setting Variables")
		// Test SetVar and DeleteVar
		for i := 0; i < 3; i++ {
			kv := fmt.Sprintf("%d", i)
			varK := fmt.Sprintf("var_%s", kv)

			_, err := SetVar(wallets[9], validSCIDs[0], kv, kv)
			assert.NoError(t, err, "Calling SetVar should not error: %s", err)
			_, err = SetVar(wallets[10], validSCIDs[0], kv, kv)
			assert.Error(t, err, "Calling SetVar when not owner should error")
			_, err = SetVar(wallets[11], validSCIDs[3], kv, kv)
			assert.Error(t, err, "Calling SetVar on anon SC should error")
			time.Sleep(sleepFor)

			v, err := retry(t, fmt.Sprintf("confirming set var %d", i), func() (string, error) {
				return getContractVar(validSCIDs[0], varK, endpoint)
			})
			assert.NoError(t, err, "SetVar variable should not error: %s", err)

			// Test KeyExists
			_, exists, err := KeyExists(validSCIDs[0], varK, endpoint)
			assert.NoError(t, err, "KeyExists string should not error: %s", err)
			assert.True(t, exists, "Key %s should exist", varK)
			_, exists, err = KeyExists(validSCIDs[0], "likes", endpoint)
			assert.NoError(t, err, "KeyExists uint64 should not error: %s", err)
			assert.True(t, exists, "Key %s should exist", varK)
			_, exists, err = KeyExists(validSCIDs[0], "notHere", endpoint)
			assert.NoError(t, err, "KeyExists should not error: %s", err)
			assert.False(t, exists, "Key should not exist")
			assert.NotEmpty(t, v, "SetVar Value should not be empty")
			_, _, err = KeyExists(validSCIDs[0], "", "")
			assert.Error(t, err, "KeyExists with invalid endpoint should error")

			// Test KeyPrefixExists
			_, _, exists, err = KeyPrefixExists(validSCIDs[0], "lik", endpoint)
			assert.NoError(t, err, "KeyExists uint64 should not error: %s", err)
			assert.True(t, exists, "Key %s should exist", kv)
			_, _, exists, err = KeyPrefixExists(validSCIDs[0], strings.Split(varK, "_")[0], endpoint)
			assert.NoError(t, err, "KeyExists string should not error: %s", err)
			assert.True(t, exists, "Key %s should exist", kv)
			_, _, _, err = KeyPrefixExists(validSCIDs[0], "", "")
			assert.Error(t, err, "KeyPrefixExists with invalid endpoint should error")

			// Leave the var store on SC 0
			if i != 0 {
				continue
			}

			_, err = DeleteVar(wallets[9], validSCIDs[0], kv)
			assert.NoError(t, err, "Calling DeleteVar should not error: %s", err)
			_, err = DeleteVar(wallets[10], validSCIDs[0], kv)
			assert.Error(t, err, "Calling DeleteVar when not owner should error")
			_, err = DeleteVar(wallets[11], validSCIDs[3], kv)
			assert.Error(t, err, "Calling DeleteVar on anon SC should error")
			time.Sleep(sleepFor)
			_, err = getContractVar(validSCIDs[0], varK, endpoint)
			assert.Error(t, err, "Variable %s should be deleted and error")
		}

		t.Logf("Rating contracts")
		// Test down rating contracts and getting rating result
		for i := 0; i < 4; i++ {
			txid, err := Rate(wallets[i], validSCIDs[i], uint64(i))
			assert.NoError(t, err, "Calling Rate should not error: %s", err)
			time.Sleep(sleepFor)

			noCodeTXIDs = append(noCodeTXIDs, txid)

			v, err := retry(t, fmt.Sprintf("confirming down %d rating", i), func() (string, error) {
				return getContractVar(validSCIDs[i], wallets[i].GetAddress().String(), endpoint)
			})
			assert.NoError(t, err, "Getting rating variable should not error: %s", err)
			assert.NotEmpty(t, v, "Value should not be empty")
			ratings, err := GetRating(validSCIDs[i], endpoint, 0)
			assert.NoError(t, err, "Getting ratings should not error: %s", err)
			assert.Equal(t, uint64(1), ratings.Dislikes, "Dislikes are not equal and should be")
			assert.Equal(t, float64(0), ratings.Average, "Rating average category should be the same")
		}

		// This wallet already rated
		_, err = Rate(wallets[0], validSCIDs[0], 0)
		assert.Error(t, err, "Wallet already rated and should not be allowed again")
		// Should return nothing
		emptyRatings, err := GetRating(validSCIDs[0], endpoint, 999999999999999)
		assert.NoError(t, err, "GetRating should not error: %s", err)
		assert.Empty(t, emptyRatings.Ratings, "Ratings should be empty when filtering with height")

		// Test up rating contracts and getting rating results
		offset := 4 // offset to next four wallets
		for i := 0; i < 4; i++ {
			_, err = Rate(wallets[i+offset], validSCIDs[i], uint64(i)+50) // >49 is up rating
			assert.NoError(t, err, "Calling Rate should not error: %s", err)
			time.Sleep(sleepFor)

			v, err := retry(t, fmt.Sprintf("confirming up %d rating", i), func() (string, error) {
				return getContractVar(validSCIDs[i], wallets[i+offset].GetAddress().String(), endpoint)
			})
			assert.NoError(t, err, "Getting rating variable should not error: %s", err)
			assert.NotEmpty(t, v, "Value should not be empty")
			ratings, err := GetRating(validSCIDs[i], endpoint, 0)
			assert.NoError(t, err, "Getting up rating should not error: %s", err)

			assert.Equal(t, uint64(1), ratings.Likes, "Likes are not equal and should be")
			assert.Equal(t, uint64(1), ratings.Dislikes, "Dislikes should have remained the same")
			assert.Equal(t, float64(2.5), ratings.Average, "Rating average category should be the same")
		}

		// This wallet already rated
		_, err = Rate(wallets[offset], validSCIDs[0], 0)
		assert.Error(t, err, "Wallet already rated and should not be allowed again")
		// Test Rate errors and nil values
		_, err = Rate(wallets[offset], validSCIDs[0], 1000)
		assert.Error(t, err, "Rating is beyond range")
		_, err = Rate(nil, "", 0)
		assert.Error(t, err, "Rate should error with nil wallet and did not")
		_, err = GetRating(nameservice, "", 0)
		assert.Error(t, err, "GetRating should error with invalid daemon")
		_, err = GetRating(nameservice, endpoint, 0)
		assert.Error(t, err, "GetRating should error on nameservice")
	})

	// // Test installing INDEXs as libraries
	t.Run("INDEXLibrary", func(t *testing.T) {
		// Tag INDEX as library
		libraryEmbeds := []*INDEX{
			{
				DURL: "zero.lib",
				DOCs: docs[0],
				Headers: Headers{
					NameHdr:  "Library0",
					DescrHdr: "Zero lib",
					IconHdr:  "icon.url",
				},
			},
			{
				DURL: "one.lib",
				DOCs: docs[1],
				Headers: Headers{
					NameHdr:  "Library1",
					DescrHdr: "One lib",
					IconHdr:  "icon.url",
				},
			},
		}

		for i, il := range libraryEmbeds {
			tx, err := retry(t, fmt.Sprintf("INDEX %d library install", i), func() (string, error) {
				return Installer(wallets[i], 2, il)
			})
			assert.NoError(t, err, "Installing INDEX %d library should not have error: %s", i, err)

			_, err = retry(t, fmt.Sprintf("confirming install TX %s", tx), func() (string, error) {
				return getContractCode(tx, endpoint)
			})
			if err != nil {
				t.Fatalf("Could not confirm INDEX %d library TX %s: %s", i, tx, err)
			}

			t.Logf("Simulator INDEX %d Library SC installed %s: %s", i, il.NameHdr, tx)
			librarySCIDs = append(librarySCIDs, tx)
		}

		// Update a lib
		libraryEmbeds[0].SCID = librarySCIDs[0]
		tx, err := Updater(wallets[0], libraryEmbeds[0])
		assert.NoError(t, err, "Update should not have error: %s", err)
		time.Sleep(sleepFor)
		hash, err := varChanged(librarySCIDs[0], "hash", tx, endpoint)
		assert.NoError(t, err, "Getting hash variable should not error: %s", err)
		assert.Equal(t, tx, hash, "Contract hash should be txid of update")

		// Embed INDEX into INDEXs
		indexEmbeds := []*INDEX{
			// Valid
			{
				DOCs: []string{docs[2][0], librarySCIDs[0]},
				DURL: "test7.tela",
				Headers: Headers{
					NameHdr:  "TELA Test App7",
					DescrHdr: "Seven TELA test app, valid INDEX embed",
				},
			},
			// Embed is not a lib
			{
				DOCs: []string{docs[2][0], validSCIDs[1]},
				DURL: "test8.tela",
				Headers: Headers{
					NameHdr:  "TELA Test App8",
					DescrHdr: "Am not a library",
				},
			},
			// Non TELA embed
			{
				DOCs: []string{docs[2][0], nameservice},
				DURL: "test9.tela",
				Headers: Headers{
					NameHdr:  "TELA Test App9",
					DescrHdr: "Have no dURL",
				},
			},
			// INDEX as entrypoint
			{
				DOCs: []string{librarySCIDs[0], librarySCIDs[1]},
				DURL: "test10",
				Headers: Headers{
					NameHdr:  "TELA Test App10",
					DescrHdr: " INDEX into INDEX",
				},
			},
		}

		var embedSCIDs []string
		for i, ie := range indexEmbeds {
			tx, err := retry(t, fmt.Sprintf("INDEX %d embed install", i), func() (string, error) {
				return Installer(wallets[1], 2, ie)
			})
			assert.NoError(t, err, "Installing INDEX %d embed should not have error: %s", i, err)

			_, err = retry(t, fmt.Sprintf("confirming install TX %s", tx), func() (string, error) {
				return getContractCode(tx, endpoint)
			})
			if err != nil {
				t.Fatalf("Could not confirm INDEX %d embed TX %s: %s", i, tx, err)
			}

			t.Logf("Simulator INDEX %d Embed SC installed %s: %s", i, ie.NameHdr, tx)
			embedSCIDs = append(embedSCIDs, tx)
		}

		// Has error first because it was updated, then it will clone successfully after AllowUpdates is set true
		err = Clone(embedSCIDs[0], endpoint)
		assert.Error(t, err, "Cloning INDEX embed with commits while updates set false should error: %s", err)
		AllowUpdates(true)
		err = Clone(embedSCIDs[0], endpoint)
		assert.NoError(t, err, "Cloning INDEX embed should not have error: %s", err)
		AllowUpdates(false)

		err = Clone(embedSCIDs[1], endpoint)
		assert.Error(t, err, "Cloning non library INDEX embed should error: %s", err)

		err = Clone(embedSCIDs[2], endpoint)
		assert.Error(t, err, "Cloning embed with invalid TELA SCID should error: %s", err)

		err = Clone(embedSCIDs[3], endpoint)
		assert.Error(t, err, "Cloning INDEX with INDEX entrypoint should error: %s", err)

		// Serve
		_, err = ServeTELA(librarySCIDs[1], endpoint)
		assert.Error(t, err, "Serving a library should error: %s", err)
	})

	// // Test creating, install and recreating shard INDEXs
	t.Run("DocShards", func(t *testing.T) {
		ringsize := uint64(2)

		CHUNK_SIZE := int64(17500)

		moveTo := filepath.Join(testDir, "datashards", "clone", "shard")

		var installedShardINDEXs []string
		var docShardFiles = []struct {
			Name      string
			Source    string
			Path      string
			Content   string
			ClonePath string
		}{
			{
				Name:   "splitTela.go",
				Source: "tela.go",
				Path:   filepath.Join(moveTo, "splitTela.go"),
			},
			{
				Name:   "splitParse.go",
				Source: "parse.go",
				Path:   filepath.Join(moveTo, "splitParse.go"),
			},
		}

		for si, shardFile := range docShardFiles {
			goFile, _ := readFile(shardFile.Source)
			// Prep the file content removing any multi line comments
			goFile = strings.ReplaceAll(goFile, "/*", "")
			goFile = strings.ReplaceAll(goFile, "*/", "")
			docShardFiles[si].Content = goFile

			content := []byte(goFile)

			err = os.MkdirAll(moveTo, os.ModePerm)
			assert.NoError(t, err, "Creating directories should not error: %s", err)

			file, err := os.Create(shardFile.Path)
			assert.NoError(t, err, "Creating new shard file should not error: %s", err)

			_, err = file.Write(content)
			assert.NoError(t, err, "Writing shard file should not error: %s", err)

			err = CreateShardFiles(shardFile.Path)
			assert.NoError(t, err, "CreateShardFiles should not error: %s", err)
			err = CreateShardFiles(shardFile.Path)
			assert.Error(t, err, "CreateShardFiles should error when exists already")
			// Hit already exists
			err = ConstructFromShards(nil, shardFile.Name, moveTo)
			assert.Error(t, err, "ConstructFromShards should error when exists already")

			shardDURL := fmt.Sprintf("%s%s", docShardFiles[si].Name, TAG_DOC_SHARDS)

			doc := &DOC{
				DocType: "TELA-GO-1",
				DURL:    shardDURL,
				Headers: Headers{
					NameHdr:  "",
					DescrHdr: "A DocShard",
				},
				Signature: Signature{
					CheckC: "c4d7bbdaaf9344f4c351e72d0b2145b4235402c89510101e0500f43969fd1387",
					CheckS: "b879b0ff01d78841d61e9770fd18436d8b9afce59302c77a786272e7422c15f6",
				},
			}

			fileInfo, _ := os.Stat(shardFile.Path)
			totalShards := (fileInfo.Size() + CHUNK_SIZE - 1) / CHUNK_SIZE

			// Install all DocShards created from original file
			for i := int64(1); i <= totalShards; i++ {
				// Name of file created by CreateShardFiles
				newName := fmt.Sprintf("%s-%d.go", strings.TrimSuffix(shardFile.Name, ".go"), i)
				code, err := readFile(strings.ReplaceAll(shardFile.Path, shardFile.Name, newName))
				if err != nil {
					t.Fatalf("Could not read %s file: %s", newName, err)
				}

				// Match nameHdr with DocShard file names
				doc.NameHdr = newName
				doc.Code = code

				tx, err := retry(t, fmt.Sprintf("DOC DocShard %d install", i), func() (string, error) {
					return Installer(wallets[i], ringsize, doc)
				})
				assert.NoError(t, err, "Install %d %s should not error: %s", i, doc.NameHdr, err)

				_, err = retry(t, fmt.Sprintf("confirming install TX %s", tx), func() (string, error) {
					return getContractCode(tx, endpoint)
				})
				if err != nil {
					t.Fatalf("Could not confirm DocShard %d TX %s: %s", i, tx, err)
				}

				shardSCIDs[si] = append(shardSCIDs[si], tx)
				t.Logf("Simulator DocShard %d SC installed %s: %s", i, doc.NameHdr, tx)
			}

			// Embed all the DocShards SCIDs into INDEX
			shardIndex := &INDEX{
				DURL: shardDURL,
				DOCs: shardSCIDs[si],
				Headers: Headers{
					NameHdr:  fmt.Sprintf("DocShards%d", si),
					DescrHdr: "Put back together",
				},
			}

			docShardFiles[si].ClonePath = filepath.Join(testDir, "datashards", "clone", shardIndex.DURL, docShardFiles[si].Name)

			tx, err := retry(t, "INDEX DocShards install", func() (string, error) {
				return Installer(wallets[9], ringsize, shardIndex)
			})

			assert.NoError(t, err, "Install %s should not error: %s", shardIndex.NameHdr, err)

			_, err = retry(t, fmt.Sprintf("confirming DocShards install TX %s", tx), func() (string, error) {
				return getContractCode(tx, endpoint)
			})
			if err != nil {
				t.Fatalf("Could not confirm INDEX DocShards TX %s: %s", tx, err)
			}

			t.Logf("Simulator INDEX DocShards SC installed %s: %s", shardIndex.NameHdr, tx)
			installedShardINDEXs = append(installedShardINDEXs, tx)
		}

		err = Clone(installedShardINDEXs[0], endpoint)
		assert.NoError(t, err, "Cloning DocShards INDEX should not error: %s", err)
		// Ensure recreated content matches original
		recreatedFile, err := readFile(docShardFiles[0].ClonePath)
		assert.NoError(t, err, "Reading recreated %s should not error: %s", docShardFiles[0].Name, err)
		assert.Equal(t, docShardFiles[0].Content, recreatedFile, "Recreated DocShard should match original")
		err = Clone(installedShardINDEXs[0], endpoint)
		assert.Error(t, err, "Cloning DocShards INDEX should error when exists already")

		err = Clone(installedShardINDEXs[1], endpoint)
		assert.NoError(t, err, "Cloning DocShards INDEX should not error: %s", err)
		recreatedFile, err = readFile(docShardFiles[1].Path)
		assert.NoError(t, err, "Reading recreated %s should not error: %s", docShardFiles[1].Name, err)
		assert.Equal(t, docShardFiles[1].Content, recreatedFile, "Recreated DocShard should match original")
		assert.NotEqual(t, docShardFiles[0].Content, docShardFiles[1].Content, "Test DocShard content should not be matching")

		AllowUpdates(true)
		_, err = ServeAtCommit(installedShardINDEXs[1], "", endpoint)
		assert.Error(t, err, "Serving with %s INDEX tag should error", TAG_DOC_SHARDS)
		AllowUpdates(false)

		// Now embed those two shard INDEXs into another INDEX
		embedShardIndex := &INDEX{
			DURL: "embedded.docshards.tela",
			DOCs: append(docs[1], installedShardINDEXs...),
			Headers: Headers{
				NameHdr:  "Embed DocShards",
				DescrHdr: "Put back together",
			},
		}

		tx, err := retry(t, "INDEX embed DocShards install", func() (string, error) {
			return Installer(wallets[10], ringsize, embedShardIndex)
		})
		assert.NoError(t, err, "Install embed DocShards should not error: %s", err)

		_, err = retry(t, fmt.Sprintf("confirming embed DocShards install TX %s", tx), func() (string, error) {
			return getContractCode(tx, endpoint)
		})
		if err != nil {
			t.Fatalf("Could not confirm INDEX embed DocShards TX %s: %s", tx, err)
		}

		t.Logf("Simulator INDEX SC installed %s: %s", embedShardIndex.NameHdr, tx)

		err = Clone(tx, endpoint)
		assert.NoError(t, err, "Cloning INDEX embed DocShards should not error: %s", err)

		// Test parseDocShards
		invalidDocShardSCs := []string{
			// No error, DOC has subDir
			`Function InitializePrivate() Uint64
	10 IF init() == 0 THEN GOTO 30
	20 RETURN 1
	30 STORE("nameHdr", "<nameHdr>")
	31 STORE("descrHdr", "<descrHdr>")
	32 STORE("iconURLHdr", "<iconURLHdr>")
	33 STORE("dURL", "<dURL>")
	40 STORE("DOC1", "` + docs[1][0] + `")
	1000 RETURN 0
	End Function`,
			// scid != 64
			`Function InitializePrivate() Uint64
	10 IF init() == 0 THEN GOTO 30
	20 RETURN 1
	30 STORE("nameHdr", "<nameHdr>")
	31 STORE("descrHdr", "<descrHdr>")
	32 STORE("iconURLHdr", "<iconURLHdr>")
	33 STORE("dURL", "<dURL>")
	40 STORE("DOC1", "<scid>")
	1000 RETURN 0
	End Function`,

			// No SC code
			`Function InitializePrivate() Uint64
	10 IF init() == 0 THEN GOTO 30
	20 RETURN 1
	30 STORE("nameHdr", "<nameHdr>")
	31 STORE("descrHdr", "<descrHdr>")
	32 STORE("iconURLHdr", "<iconURLHdr>")
	33 STORE("dURL", "<dURL>")
	40 STORE("DOC1", "c4d7bbdaaf9344f4c351e72d0b2145b4235402c89510101e0500f43969fd1387")
	1000 RETURN 0
	End Function`,

			// Invalid DOC
			`Function InitializePrivate() Uint64
	10 IF init() == 0 THEN GOTO 30
	20 RETURN 1
	30 STORE("nameHdr", "<nameHdr>")
	31 STORE("descrHdr", "<descrHdr>")
	32 STORE("iconURLHdr", "<iconURLHdr>")
	33 STORE("dURL", "<dURL>")
	40 STORE("DOC1", "0000000000000000000000000000000000000000000000000000000000000001")
	1000 RETURN 0
	End Function`,
		}

		for i, code := range invalidDocShardSCs {
			sc, _, err := dvm.ParseSmartContract(code)
			assert.NoError(t, err, "Parsing DocShard code %d should not error: %s", i, err)
			_, recreate, err := parseDocShards(sc, "", endpoint)
			if i == 0 {
				assert.NoError(t, err, "Parsing DocShard %d should not error: %s", i, err)
				assert.Equal(t, filepath.Join(telaDocs[3].SubDir, telaDocs[3].NameHdr), recreate, "Recreated file should have subDir prefix")
			} else {
				assert.Error(t, err, "Parsing DocShard %d should error", i)
			}
		}

		// The master embed INDEX contains every type of embed, DOC, Lib, DocShards
		masterEmbedIndex := &INDEX{
			DURL: "embedded.master.tela",
			DOCs: docs[0],
			Headers: Headers{
				NameHdr:  "Embed master",
				DescrHdr: "Has all",
			},
		}

		masterSCID, err := retry(t, "INDEX embed master install", func() (string, error) {
			return Installer(wallets[11], ringsize, masterEmbedIndex)
		})
		assert.NoError(t, err, "Install embed master should not error: %s", err)

		_, err = retry(t, fmt.Sprintf("confirming embed master install TX %s", tx), func() (string, error) {
			return getContractCode(tx, endpoint)
		})
		if err != nil {
			t.Fatalf("Could not confirm INDEX embed master TX %s: %s", tx, err)
		}

		// Update the master embed so it can be cloned from commit
		masterEmbedIndex.SCID = masterSCID
		masterEmbedIndex.DOCs = append(masterEmbedIndex.DOCs, librarySCIDs[0], installedShardINDEXs[0])
		masterTXID, err := retry(t, "confirming embed master update", func() (string, error) {
			return Updater(wallets[11], masterEmbedIndex)
		})
		assert.NoError(t, err, "Install embed master should not error: %s", err)

		t.Logf("Simulator INDEX master embed SC installed: %s", masterSCID)

		// Daemon might not have tx data yet
		time.Sleep(sleepFor * 3)
		err = CloneAtCommit(masterSCID, masterTXID, endpoint)
		assert.NoError(t, err, "Cloning embed master at commit should not error: %s", err)
		err = CloneAtCommit(masterSCID, masterTXID, endpoint)
		assert.Error(t, err, "Cloning embed master at commit should error when already exists")
	})

	// // Test downward compatibility
	t.Run("Versions", func(t *testing.T) {
		// Test LessThan
		for i := 0; i < len(tela.version.index)-1; i++ {
			old := tela.version.index[i]
			new := tela.version.index[i+1]
			assert.True(t, old.LessThan(new), "Version %v should be less than %v", old, new)
		}

		versionStringsTrue := [][]Version{
			{Version{0, 0, 0}, Version{1, 0, 0}},
			{Version{0, 0, 0}, Version{0, 1, 0}},
			{Version{0, 0, 0}, Version{0, 0, 1}},
		}

		for _, v := range versionStringsTrue {
			assert.True(t, v[0].LessThan(v[1]), "Version %s should be less than %s", v[0], v[1])
		}

		versionStringsFalse := [][]Version{
			{Version{1, 0, 0}, Version{0, 0, 0}},
			{Version{0, 1, 0}, Version{0, 0, 0}},
			{Version{0, 0, 1}, Version{0, 0, 0}},
			{Version{0, 0, 0}, Version{0, 0, 0}},
		}

		for _, v := range versionStringsFalse {
			assert.False(t, v[0].LessThan(v[1]), "Version %s should not be less than %s", v[0], v[1])
		}

		// Test GetVersion
		assert.Equal(t, tela.version.pkg, GetVersion(), "TELA versions are not the same")

		// Test GetContractVersions
		assert.Equal(t, tela.version.index, GetContractVersions(false), "INDEX versions are not the same")
		assert.Equal(t, tela.version.docs, GetContractVersions(true), "DOC versions are not the same")

		// Test GetLatestContractVersion
		assert.Equal(t, tela.version.docs[len(tela.version.docs)-1], GetLatestContractVersion(true), "Latest DOC version should be the same")

		// Test ParseVersion
		versionStrings := []string{
			"0.0.0",
			"1.1.1",
			"7.8.9",
			"1",
			"1.1",
			"a",
			"a.a.a",
			"1.a.a",
			"1.1.a",
			"",
		}

		for i, str := range versionStrings {
			_, err := ParseVersion(str)
			if i > 2 {
				assert.Error(t, err, "%s should not a valid a valid version", str)
			} else {
				assert.NoError(t, err, "%s should be a valid a valid version", str)
			}
		}

		// INDEXs
		indexVersions, indexCode := createContractVersions(false, "")
		for i := 0; i < len(indexVersions)-1; i++ {
			versionInstaller := func() (tx string, err error) {
				vIndex := &INDEX{
					DURL: fmt.Sprintf("v%s.tela", indexVersions[i].String()),
					Headers: Headers{
						NameHdr: indexVersions[i].String(),
					},
				}

				testVersionCode, err := ParseHeaders(indexCode[i], vIndex)
				assert.NoError(t, err, "Adding headers to v%s should not error: %s", indexVersions[i].String(), err)

				args := rpc.Arguments{
					rpc.Argument{Name: rpc.SCACTION, DataType: rpc.DataUint64, Value: uint64(rpc.SC_INSTALL)},
					rpc.Argument{Name: rpc.SCCODE, DataType: rpc.DataString, Value: testVersionCode},
				}

				return transfer0(wallets[i+9], 2, args)
			}

			tx, err := retry(t, fmt.Sprintf("INDEX v%s install", indexVersions[i].String()), func() (string, error) {
				return versionInstaller()
			})
			assert.NoError(t, err, "Installing INDEX v%s should not have error: %s", indexVersions[i].String(), err)

			_, err = retry(t, fmt.Sprintf("confirming install TX %s", tx), func() (string, error) {
				return getContractCode(tx, endpoint)
			})
			if err != nil {
				t.Fatalf("Could not confirm INDEX v%s TX %s: %s", indexVersions[i].String(), tx, err)
			}

			t.Logf("Simulator INDEX v%s SC installed %s: %s", indexVersions[i].String(), "Test App 11", tx)

			scid := tx
			update := &INDEX{
				SCVersion: &indexVersions[i],
				SCID:      scid,
				DURL:      fmt.Sprintf("v%s.tela", indexVersions[len(indexVersions)-1].String()), // the latest version number
				DOCs:      []string{docs[0][0], docs[0][1], docs[0][2]},
				Headers: Headers{
					NameHdr:  indexVersions[len(indexVersions)-1].String(),
					DescrHdr: "A TELA Application",
					IconHdr:  "ICON_URL",
				},
			}

			// Contract should successfully update to latest version
			tx, err = Updater(wallets[i+9], update)
			assert.NoError(t, err, "Update to latest should not have error: %s", err)
			time.Sleep(sleepFor)
			hash, err := varChanged(scid, "hash", tx, endpoint)
			assert.NoError(t, err, "Getting hash variable should not error: %s", err)
			assert.Equal(t, tx, hash, "Contract hash should be txid of update")
			AllowUpdates(true)
			_, err = ServeTELA(scid, endpoint)
			assert.NoError(t, err, "Update to latest and serve should not error: %s", err)
			AllowUpdates(false)
		}

		ShutdownTELA()
	})

	// // Test internal functions
	t.Run("Internal", func(t *testing.T) {
		// Invalid docType language
		assert.False(t, IsAcceptedLanguage("TELA-NotAcceptedLanguage-1"), "Should not be a accepted language")

		// Languages are case sensitive
		assert.False(t, IsAcceptedLanguage("TELA-html-1"), "Language should be case sensitive")

		// Parse empty document shard with no multiline comment
		_, err := parseDocShardCode("", "")
		assert.Error(t, err, "Invalid shard code should error")

		// Parse empty document with no multiline comment
		assert.Error(t, parseAndSaveTELADoc("", "", ""), "Should not be able to parse this document")
		// Parse invalid docType
		assert.Error(t, parseAndSaveTELADoc("", "/*\n*/", "invalid"), "DocType should be invalid")
		// Save to invalid path
		assert.Error(t, parseAndSaveTELADoc(filepath.Join(mainPath, testDir, "app2", "index.html", "filename.html"), TELA_DOC_1, DOC_HTML), "Path should be invalid")
		// Parse types not installed in tests
		assert.NoError(t, parseAndSaveTELADoc(filepath.Join(datashards, "test_test.go"), TELA_DOC_1, DOC_GO), "DOC_GO should be valid")
		assert.NoError(t, parseAndSaveTELADoc(filepath.Join(datashards, "json.json"), TELA_DOC_1, DOC_JSON), "DOC_JSON should be valid")
		assert.NoError(t, parseAndSaveTELADoc(filepath.Join(datashards, "markdown.md"), TELA_DOC_1, DOC_MD), "DOC_MD should be valid")
		assert.NoError(t, parseAndSaveTELADoc(filepath.Join(datashards, "other.jsx"), TELA_DOC_1, DOC_STATIC), "DOC_STATIC should be valid")

		// decodeHexString return non hex
		expectedAddress := "deto1qy87ghfeeh6n6tdxtgh7yuvtp6wh2uwfzvz7vjq0krjy4smmx62j5qgqht7t3"
		assert.Equal(t, expectedAddress, decodeHexString(expectedAddress), "decodeHexString should return a DERO address when passed")

		// Test cloneDOC
		_, err = cloneDOC(scDoesNotExist, "", "", endpoint)
		assert.Error(t, err, "cloneDOC should error with invalid scid")
		_, err = cloneDOC(nameservice, "", "", endpoint)
		assert.Error(t, err, "cloneDOC with NON TELA should error")

		// scid != 64
		_, err = cloneINDEX("", "", "", endpoint)
		assert.Error(t, err, "cloneINDEX with invalid SCID should error")
		_, err = cloneINDEX(nameservice, "", "", endpoint)
		assert.Error(t, err, "cloneINDEX should error when cloning a invalid SCID")
		_, err = cloneINDEX(docs[0][1], "", "", endpoint)
		assert.Error(t, err, "cloneINDEX should error when cloning a DOC")

		// Test getTXID
		_, _, err = getTXID(scDoesNotExist, endpoint)
		assert.Error(t, err, "Getting invalid TX should error")
		_, _, err = getTXID(scDoesNotExist, "")
		assert.Error(t, err, "Getting invalid TX with invalid endpoint should error")

		// Test cloneINDEXAtCommit
		_, err = cloneINDEXAtCommit(0, "", "", tela.path.clone(), endpoint) // invalid scid
		assert.Error(t, err, "cloneINDEXAtCommit with invalid SCID should error")
		_, err = cloneINDEXAtCommit(0, nameservice, commitTXIDs[0], tela.path.clone(), endpoint) // NON TELA
		assert.Error(t, err, "cloneINDEXAtCommit with NON TELA should error")
		_, err = cloneINDEXAtCommit(0, docs[0][1], docs[0][1], "", endpoint)
		assert.Error(t, err, "cloneINDEXAtCommit should error when cloning a DOC")
		_, err = cloneINDEXAtCommit(1, scDoesNotExist, scDoesNotExist, "", endpoint)
		assert.Error(t, err, "cloneINDEXAtCommit should error when cloning empty code")

		// Test extractCodeFromTXID
		_, err = extractCodeFromTXID(hex.EncodeToString([]byte("someHex"))) // random hex
		assert.Error(t, err, "Extracting hex without SC code should error")
		_, err = extractCodeFromTXID(hex.EncodeToString([]byte("Function "))) // malformed function in hex
		assert.Error(t, err, "Extracting hex malformed SC code should error")

		// Test extractModTagFromCode
		expectedModTag := "tag1,tag2,tag3"
		assert.Equal(t, "<modTags>", extractModTagFromCode(TELA_INDEX_1), "Extracting modTag from template should be equal")
		assert.Equal(t, expectedModTag, extractModTagFromCode(fmt.Sprintf(`%s"tag1,tag2,tag3")`, LINE_MODS_STORE)), "Extracting modTag should be equal")
		assert.Equal(t, expectedModTag, extractModTagFromCode(fmt.Sprintf(`%s "   tag1,tag2,tag3    "  )   `, LINE_MODS_STORE)), "Extracting spaced modTag should be equal")
		assert.Empty(t, extractModTagFromCode(LINE_MODS_STORE), "Invalid mods line should return empty modTag")
		assert.Empty(t, extractModTagFromCode(""), "Invalid mods line should return empty modTag")
	})

	// // Test other exported functions
	t.Run("Exported", func(t *testing.T) {
		// Test Clone
		err := Clone(docs[0][0], endpoint)
		assert.NoError(t, err, "Cloning DOC should not error: %s", err)
		err = Clone(validSCIDs[1], endpoint)
		assert.NoError(t, err, "Cloning INDEX should not error: %s", err)
		err = Clone(nameservice, endpoint)
		assert.Error(t, err, "Cloning NON-TELA should error")

		// Test CloneAtCommit
		err = CloneAtCommit(validSCIDs[0], commitTXIDs[0], endpoint) // invalid DOCs
		assert.Error(t, err, "Clone at commit is an invalid INDEX and should error")
		err = CloneAtCommit(validSCIDs[0], commitTXIDs[1], endpoint)
		assert.NoError(t, err, "Cloning valid SCID at commit should not error: %s", err)
		err = CloneAtCommit("", "", endpoint) // invalid scid
		assert.Error(t, err, "Cloning with invalid SCID should error")
		err = CloneAtCommit(validSCIDs[0], "", endpoint) // empty txid
		assert.Error(t, err, "Cloning with empty TXID should error")
		err = CloneAtCommit(nameservice, commitTXIDs[0], endpoint) // NON TELA
		assert.Error(t, err, "Cloning NON TELA should error")
		err = CloneAtCommit(validSCIDs[0], scDoesNotExist, endpoint) // good scid with invalid tx
		assert.Error(t, err, "Cloning valid INDEX with invalid commit should error")
		err = CloneAtCommit(validSCIDs[0], noCodeTXIDs[0], endpoint) // good scid with no code tx
		assert.Error(t, err, "Cloning valid INDEX with no code TXID should error")

		// Test ServeAtCommit
		_, err = ServeAtCommit(validSCIDs[0], commitTXIDs[1], endpoint) // error due to no update
		assert.Error(t, err, "Serving at commit when updates is set false should error")
		AllowUpdates(true)
		_, err = ServeAtCommit(validSCIDs[0], commitTXIDs[1], endpoint) // valid serve
		assert.NoError(t, err, "Serving valid SCID with valid commit should not error: %s", err)
		_, err = ServeAtCommit(validSCIDs[0], "txid", endpoint) // error
		assert.Error(t, err, "Serving valid SCID with invalid commit should error")
		AllowUpdates(false)
		ShutdownTELA()

		// Test GetPath
		expectedPath := filepath.Join(shards.GetPath(), "tela")
		assert.Equal(t, expectedPath, GetPath(), "TELA path should be equal")

		// Test GetClonePath
		expectedClonePath := filepath.Join(shards.GetPath(), "clone")
		assert.Equal(t, expectedClonePath, GetClonePath(), "Clone path should be equal")

		// Test SetShardPath
		testPath := "likely/A/invalid/Path"
		err = SetShardPath(testPath)
		assert.Error(t, err, "Invalid shard path should not be set")

		assert.Equal(t, expectedPath, GetPath(), "Datashard path should have remained as default")
		err = SetShardPath(filepath.Dir(datashards)) // set test path
		assert.NoError(t, err, "Valid shard path should not error: %s", err)

		// Test AllowUpdates
		AllowUpdates(true)
		assert.True(t, UpdatesAllowed(), "Updates should be allowed and are not")
		AllowUpdates(false)
		assert.False(t, UpdatesAllowed(), "Updates should not be allowed and are")

		// Test SetPortStart
		assert.Equal(t, DEFAULT_PORT_START, PortStart(), "Port start should be equal")
		setTo := 1200
		err = SetPortStart(setTo)
		assert.NoError(t, err, "Set port should not error setting to %d", setTo)
		assert.Equal(t, setTo, PortStart(), "Port start should be equal")
		err = SetPortStart(DEFAULT_PORT_START)
		assert.NoError(t, err, "Set port should not error setting to %d", DEFAULT_PORT_START)
		assert.Equal(t, DEFAULT_PORT_START, PortStart(), "Port start should be equal")
		setTo = 99999999999
		err = SetPortStart(setTo)
		assert.Error(t, err, "Set port should error setting to %d", setTo)
		assert.Equal(t, DEFAULT_PORT_START, PortStart(), "Port start should be equal")

		// Test SetMaxServers
		setTo = 30
		SetMaxServers(setTo)
		assert.Equal(t, setTo, MaxServers(), "Max servers should have been set to %d", setTo)
		SetMaxServers(-1)
		assert.Equal(t, 1, MaxServers(), "Max servers should default to minimum 1")
		SetMaxServers(9999999999)
		maxPossibleServers := DEFAULT_MAX_PORT - tela.port
		assert.Equal(t, maxPossibleServers, MaxServers(), "Max servers should default to maximum of %d-%d", DEFAULT_MAX_PORT, tela.port)
		SetMaxServers(DEFAULT_MAX_SERVER)
		assert.Equal(t, DEFAULT_MAX_SERVER, MaxServers(), "Max servers should have been set to default")

		// Pass nil values to Installer
		_, err = Installer(nil, 0, nil)
		assert.Error(t, err, "Installer should error with nil wallet and did not")
		_, err = Installer(wallets[0], 0, nil)
		assert.Error(t, err, "Installer should error with nil params and did not")

		// Pass nil values to Updater
		_, err = Updater(nil, nil)
		assert.Error(t, err, "Updater should error with nil wallet and did not")
		_, err = Updater(wallets[0], nil)
		assert.Error(t, err, "Updater should error with nil params and did not")

		indInfo, err := GetINDEXInfo(validSCIDs[1], endpoint)
		assert.NoError(t, err, "Getting valid INDEX should not error: %s", err)
		assert.Equal(t, docs[0], indInfo.DOCs, "INDEX DOCs should be equal and are not")
		expectedAuthor := "deto1qy5afru8r3rryk357gh002l77yssljsh3x6drrq2c3acf2u4w63zjqg7sqz9t"
		assert.Equal(t, expectedAuthor, indInfo.Author, "INDEX author should be equal and is not")
		assert.Equal(t, telaApps[1].NameHdr, indInfo.NameHdr, "INDEX nameHdr should be equal and is not")
		assert.Equal(t, telaApps[1].DURL, indInfo.DURL, "INDEX dURL should be equal and is not")
		_, err = GetINDEXInfo(nameservice, endpoint)
		assert.Error(t, err, "Getting invalid INDEX should error and did not")
		_, err = GetINDEXInfo(validSCIDs[1], "")
		assert.Error(t, err, "Getting INDEX should error with invalid daemon")
	})

	// // Test header methods
	t.Run("Header", func(t *testing.T) {
		// Trim()
		assert.Equal(t, "nameHdr", HEADER_NAME.Trim(), "HEADER_NAME should be equal and was not")
		assert.Equal(t, "DOC", HEADER_DOCUMENT.Trim(), "HEADER_DOCUMENT should be equal and was not")

		// Number()
		assert.Equal(t, Header(`"DOC1"`), HEADER_DOCUMENT.Number(1), "HEADER_DOCUMENT.Number() should be equal and was not")
		assert.Equal(t, Header(`"DOC22"`), HEADER_DOCUMENT.Number(22), "HEADER_DOCUMENT.Number() should be equal and was not")
		assert.Equal(t, HEADER_NAME, HEADER_NAME.Number(1), "HEADER_NAME.Number() should return the same and did not")
		assert.False(t, Header(`"`).CanAppend(), "Empty Header should not append")
	})

	// // Test parse functions
	t.Run("Parse", func(t *testing.T) {
		// Test formatValue()
		assert.Equal(t, "64", formatValue(uint64(64)), "Uint64 value format should be equal and was not")
		assert.Equal(t, `"11"`, formatValue(float64(11)), "Default value format should be equal and was not")
		assert.Equal(t, "1", formatValue(1), "Int value format should be equal and was not")

		// Equal contracts
		_, err := EqualSmartContracts(TELA_INDEX_1, TELA_DOC_1)
		assert.Error(t, err, "These contracts should not be equal")
		_, err = EqualSmartContracts(TELA_INDEX_1, TELA_INDEX_1)
		assert.NoError(t, err, "These contracts should be equal: %s", err)

		// Test ParseDocType
		assert.Equal(t, DOC_HTML, ParseDocType("index.html"), "Filename should parse as %s", DOC_HTML)
		assert.Equal(t, DOC_HTML, ParseDocType("subDir/index.html"), "Filename should parse as %s", DOC_HTML)
		assert.Equal(t, DOC_JSON, ParseDocType("manifest.json"), "Filename should parse as %s", DOC_JSON)
		assert.Equal(t, DOC_CSS, ParseDocType("subDir/style.css"), "Filename should parse as %s", DOC_CSS)
		assert.Equal(t, DOC_JS, ParseDocType("main.js"), "Filename should parse as %s", DOC_JS)
		assert.Equal(t, DOC_MD, ParseDocType("read.md"), "Filename should parse as %s", DOC_MD)
		assert.Equal(t, DOC_MD, ParseDocType("read.MD"), "Filename should parse as %s", DOC_MD)
		assert.Equal(t, DOC_GO, ParseDocType("main.go"), "Filename should parse as %s", DOC_GO)
		assert.Equal(t, DOC_STATIC, ParseDocType("read.txt"), "Filename should parse as %s", DOC_STATIC)
		assert.Equal(t, DOC_STATIC, ParseDocType("static.tsx"), "Filename should parse as a %s", DOC_STATIC)
		assert.Equal(t, DOC_STATIC, ParseDocType("LICENSE"), "LICENSE should parse as a %s", DOC_STATIC)
		assert.Equal(t, "", ParseDocType("read"), "Filename should not parse as a DOC")

		// Test ParseINDEXForDOCs
		scids, err := ParseINDEXForDOCs(TELA_INDEX_1)
		assert.NoError(t, err, "ParseINDEXForDOCs should not error with valid contract: %s", err)
		assert.Len(t, scids, 1, "There should be one scid")
		scids, err = ParseINDEXForDOCs(TELA_DOC_1)
		assert.NoError(t, err, "ParseINDEXForDOCs should not error with valid contract")
		assert.Empty(t, scids, "ParseINDEXForDOCs should not have found SCIDs on a DOC")
		_, err = ParseINDEXForDOCs("ThisIsNotASC")
		assert.Error(t, err, "ParseINDEXForDOCs should error with invalid contract")

		// Test ParseSignature
		signature := wallets[0].SignData([]byte("some data"))
		signatureHeader := "-----BEGIN DERO SIGNED MESSAGE-----"
		thisAddress := "deto1qy87ghfeeh6n6tdxtgh7yuvtp6wh2uwfzvz7vjq0krjy4smmx62j5qgqht7t3"
		signatureAddress := fmt.Sprintf("Address: %s", thisAddress)
		signatureCheckC := "C: 1c37f9e61f15a9526ba680dce0baa567e642ca2cd0ddea71649dab415dad8cb2"
		signatureFooter := "-----END DERO SIGNED MESSAGE-----"

		addr, _, _, err := ParseSignature(signature)
		assert.NoError(t, err, "Parse signature should not error: %s", err)
		assert.Equal(t, thisAddress, addr, "Signature address does not match %s: %s", thisAddress, addr)
		_, _, _, err = ParseSignature(nil)
		assert.Error(t, err, "Parse signature should error and did not")
		_, _, _, err = ParseSignature([]byte(fmt.Sprintf("%s\n%s\n\n%s", signatureHeader, "Address: deto1qy", signatureFooter)))
		assert.Error(t, err, "Parse signature should error with invalid address and did not")
		_, _, _, err = ParseSignature([]byte(fmt.Sprintf("%s\n%s\n%s\n\n%s", signatureHeader, signatureAddress, "C: string", signatureFooter)))
		assert.EqualError(t, fmt.Errorf("unknown C format"), err.Error(), "Parse signature should error with invalid C and did not")
		_, _, _, err = ParseSignature([]byte(fmt.Sprintf("%s\n%s\n%s\n%s\n\n%s", signatureHeader, signatureAddress, signatureCheckC, "S: string", signatureFooter)))
		assert.EqualError(t, fmt.Errorf("unknown S format"), err.Error(), "Parse signature should error with invalid S and did not")

		// Parse structures outside of standard contracts
		_, err = ParseHeaders(TELA_INDEX_1, map[Header]interface{}{HEADER_COVER_URL: "cover", HEADER_FILE_URL: "file", HEADER_ROYALTY: 1})
		assert.NoError(t, err, "map[Header]interface{} should be valid: %s", err)
		_, err = ParseHeaders(TELA_INDEX_1, map[string]interface{}{"1": 1, `"two"`: "two", "three": uint64(3)})
		assert.NoError(t, err, "map[string]interface{} should be valid: %s", err)

		// Test contract (passes dvm.ParseSmartContract) for further parsing outside of standard contracts
		file := "test1.bas"
		contractsPath := filepath.Join(testDir, "contracts")
		code, err := readFile(contractsPath, file)
		if err != nil {
			t.Fatalf("Could not read %s file: %s", file, err)
		}

		sc, _, err := dvm.ParseSmartContract(code)
		assert.NoError(t, err, "DVM parse should not error: %s", err)
		_, err = FormatSmartContract(sc, code)
		assert.NoError(t, err, "Format should not error: %s", err)
		_, err = FormatSmartContract(sc, TELA_INDEX_1) // code does not have entrypoint from sc
		assert.Error(t, err, "Format did not return error and should")

		_, err = EqualSmartContracts(TELA_INDEX_1, code)
		assert.Error(t, err, "Contracts should not be equal")
		_, err = EqualSmartContracts(code, code)
		assert.NoError(t, err, "Contracts should be equal: %s", err)

		// Parse valid and invalid header types
		_, err = ParseHeaders("", nil)
		assert.Error(t, err, "Invalid header interface should return error")
		_, err = ParseHeaders(TELA_INDEX_1, &Headers{NameHdr: "TELA Test"}) // Headers case in switch
		assert.NoError(t, err, "Valid header interface should not return error: %s", err)
		_, err = ParseHeaders(TELA_INDEX_1, &Headers{NameHdr: ""}) // Empty "nameHdr" error
		assert.Error(t, err, "Empty nameHdr key should return error")
		headers := &INDEX{Headers: Headers{NameHdr: "TELA TEST"}, DOCs: []string{"ONE", "TWO", "THREE", "FOUR", "FIVE", "SIX"}} // More DOCs parsed then on template
		_, err = ParseHeaders(TELA_INDEX_1, headers)
		assert.NoError(t, err, "Should be able to parse these docs: %s", err)
		_, err = ParseHeaders(code, &INDEX{DOCs: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}})
		assert.Error(t, err, "Contract should not have enough room to add this many headers")
		_, err = ParseHeaders("This is not valid DVM code", headers)
		assert.Error(t, err, "Contract should error with invalid DVM code")

		// Test contracts (passes dvm.ParseSmartContract) for further comparison against test1.bas
		file = "test3.bas"
		code3, err := readFile(contractsPath, file)
		if err != nil {
			t.Fatalf("Could not read %s file: %s", file, err)
		}

		_, err = EqualSmartContracts(code3, code) // Lines should be different
		assert.Error(t, err, "Contracts did not return error and should")

		file = "test4.bas"
		code4, err := readFile(contractsPath, file)
		if err != nil {
			t.Fatalf("Could not read %s file: %s", file, err)
		}

		_, err = EqualSmartContracts(code4, code) // Lines parts should be different
		assert.Error(t, err, "Contracts did not return error and should")

		file = "test5.bas"
		code5, err := readFile(contractsPath, file)
		if err != nil {
			t.Fatalf("Could not read %s file: %s", file, err)
		}

		_, err = EqualSmartContracts(code5, code) // Lines index should be missing
		assert.Error(t, err, "Contracts did not return error and should")

		file = "test6.bas"
		code6, err := readFile(contractsPath, file)
		if err != nil {
			t.Fatalf("Could not read %s file: %s", file, err)
		}

		_, err = EqualSmartContracts(code6, code) // Function name missing
		assert.Error(t, err, "Contracts did not return error and should")

		_, err = EqualSmartContracts(code6, "This is not valid DVM code") // Error on v parse
		assert.Error(t, err, "Contracts did not return error and should")

		// Test contract (fails dvm.ParseSmartContract) for further parsing outside of standard contracts
		file = "test2.bas"
		code, err = readFile(contractsPath, file)
		if err != nil {
			t.Fatalf("Could not read %s file: %s", file, err)
		}

		_, err = EqualSmartContracts(code, code) // Error on c parse
		assert.Error(t, err, "Contracts did not return error and should")

		sc, _, err = dvm.ParseSmartContract(code)
		assert.Error(t, err, "DVM parse did not return error and should")
		_, err = FormatSmartContract(sc, TELA_INDEX_1)
		assert.Error(t, err, "Format did not return error and should")

		// Exceed max docCode size
		codeSizeToBig := int(MAX_DOC_CODE_SIZE+1) * 1024
		dvmCode := "Function InitializePrivate() Uint64\n10 RETURN 0\nEnd Function\n\n"
		docCode := strings.Repeat(dvmCode, codeSizeToBig/len(dvmCode))

		doc := &DOC{
			DocType: "TELA-HTML-1",
			Code:    docCode,
			DURL:    "error.tela",
			Signature: Signature{
				CheckC: "1c37f9e61f15a9526ba680dce0baa567e642ca2cd0ddea71649dab415dad8cb2",
				CheckS: "1c37f9e61f15a9526ba680dce0baa567e642ca2cd0ddea71649dab415dad8cb2",
			},
			Headers: Headers{
				NameHdr: "nameHdr",
			},
		}

		_, err = Installer(wallets[0], 2, doc)
		assert.ErrorContains(t, err, fmt.Sprintf("docCode size is to large, max %.2fKB", MAX_DOC_CODE_SIZE), "Did not exceed max docCode size")

		// Stay within docCode size limit but exceed max total size
		codeSizeOk := int(MAX_DOC_CODE_SIZE-2) * 1024
		doc.Code = strings.Repeat(dvmCode, codeSizeOk/len(dvmCode))
		doc.Headers = Headers{
			NameHdr:  strings.Repeat("A long name", 50),
			DescrHdr: strings.Repeat("A long description", 50),
			IconHdr:  strings.Repeat("A long url", 50),
		}

		_, err = Installer(wallets[1], 2, doc)
		assert.ErrorContains(t, err, fmt.Sprintf("DOC SC size is to large, max %.2fKB", MAX_DOC_INSTALL_SIZE), "Did not exceed max total size")
	})

	// // Test ratings functions
	t.Run("Ratings", func(t *testing.T) {
		for u := uint64(0); u < 102; u++ {
			f := float64(u)
			result := Rating_Result{Average: f}
			rating, err := Ratings.ParseString(u)
			fP := u / 10
			sP := u % 10
			if u > 99 {
				assert.Error(t, err, "Invalid rating %d should error", u)
				assert.Empty(t, result.ParseAverage(), "Parsing invalid average %d should be empty string", u)
			} else {
				assert.NoError(t, err, "Parsing valid rating %d should not error: %s", u, err)
				category := Ratings.Category(fP)
				isPositive := fP >= 5
				detail := Ratings.Detail(sP, isPositive)
				expected := category
				if sP > 0 {
					expected = fmt.Sprintf("%s (%s)", category, detail)
				}

				assert.Equal(t, expected, rating, "Rating string %d should be the same", u)
				assert.Equal(t, category, result.ParseAverage(), "Parsing %d average to category string should be equal", u)
			}
		}

		shouldNotExists := uint64(100)

		assert.Empty(t, Ratings.Category(shouldNotExists), "Category should not exists")
		assert.Equal(t, Ratings.categories, Ratings.Categories(), "Categories should be the same")

		assert.Empty(t, Ratings.Detail(shouldNotExists, false), "Negative detail should not exists")

		assert.Equal(t, detail_Errors, Ratings.Detail(4, false), "Negative detail should be the same")

		assert.Empty(t, Ratings.Detail(shouldNotExists, true), "Positive detail should not exists")
		assert.Equal(t, detail_Errors, Ratings.Detail(4, true), "Positive detail should be the same")

		assert.Equal(t, Ratings.negativeDetails, Ratings.NegativeDetails(), "Negative details should be the same")
		assert.Equal(t, Ratings.positiveDetails, Ratings.PositiveDetails(), "Positive details should be the same")
	})

	// // Test errors
	t.Run("Errors", func(t *testing.T) {
		// Invalid scids
		for _, scid := range invalidSCIDS {
			_, err := ServeTELA(scid, endpoint)
			assert.Error(t, err, "SCID should error and did not")
		}

		// Test SetVar errors
		_, err = SetVar(wallets[0], validSCIDs[0], strings.Repeat("1", 257), "")
		assert.Error(t, err, "Setting invalid key should error")
		_, err = SetVar(nil, validSCIDs[0], "1", "")
		assert.Error(t, err, "Setting key with nil wallet should error")

		// Test DeleteVar errors
		_, err = DeleteVar(wallets[0], validSCIDs[0], strings.Repeat("1", 257))
		assert.Error(t, err, "Deleting invalid key should error")
		_, err = DeleteVar(nil, validSCIDs[0], "1")
		assert.Error(t, err, "Deleting key with nil wallet should error")

		// Test NewInstallArgs errors
		_, err = NewInstallArgs(&INDEX{Mods: "badTag"})
		assert.Error(t, err, "NewInstallArgs with bad modTags should error")
		// Test NewUpdateArgs errors
		_, err = NewUpdateArgs(&INDEX{Mods: "badTag"})
		assert.Error(t, err, "NewUpdateArgs with bad modTags should error")

		// Invalid daemon address
		_, err := ServeTELA(validSCIDs[0], "")
		assert.Error(t, err, "Daemon address on getContractVar should not have connected")
		_, err = getContractCode(validSCIDs[0], "")
		assert.Error(t, err, "Daemon address on getContractCode should not have connected")
		_, err = getContractCodeAtHeight(0, validSCIDs[0], "")
		assert.Error(t, err, "Daemon address on getContractCodeAtHeight should not have connected")
		_, err = getContractCodeAtHeight(0, scDoesNotExist, endpoint)
		assert.Error(t, err, "getContractCodeAtHeight should error when no code is found")
		_, err = getContractVars(validSCIDs[0], "")
		assert.Error(t, err, "Daemon address on getContractVars should not have connected")

		// No servers should be started on errors
		assert.Empty(t, GetServerInfo(), "Server info should be empty after error tests")
		assert.Nil(t, tela.servers, "Servers should be nil after error tests")

		invalidPath := filepath.Join("/path/", "/to/", "/file/", "/file.txt")

		err = CreateShardFiles(invalidPath)
		assert.Error(t, err, "CreateShardFiles should error when file is not present")

		err = ConstructFromShards(nil, "", invalidPath)
		assert.Error(t, err, "ConstructFromShards should error when path is invalid")
		err = cloneDocShards(dvm.SmartContract{}, invalidPath, "")
		assert.Error(t, err, "cloneDocShards should error when path is invalid")

		// tela.servers should be nil here
		ShutdownServer("")
		ShutdownTELA()

		// Reset testnet flag for nil/network/ringsize cases and daemon transfer error
		globals.Arguments["--testnet"] = false
		walletapi.Daemon_Endpoint_Active = ""
		transfer0(nil, 0, nil)
		transfer0(wallets[0], 0, nil)
		transfer0(wallets[0], 256, nil)
		globals.Arguments["--testnet"] = true
		globals.Arguments["--simulator"] = false
		transfer0(wallets[0], 0, nil)

		// Test getSCErrors
		assert.True(t, getSCErrors("NOT AVAILABLE err:"), "getSCErrors should be true on error")
	})

	// // Test MODs
	t.Run("MODs", func(t *testing.T) {
		// Test Functions
		functionCode, functionNames := Mods.Functions("notHere")
		assert.Empty(t, functionCode, "Function code should be empty")
		assert.Empty(t, functionNames, "Function names should be empty")
		// Test GetAllMods
		assert.Equal(t, Mods.mods, Mods.GetAllMods(), "GetAllMods should be equal")
		// Test GetAllMods
		assert.Equal(t, Mods.classes, Mods.GetAllClasses(), "GetAllClasses should be equal")
		// Test GetMod
		for i := range Mods.mods {
			getMod := Mods.GetMod(Mods.Tag(i))
			assert.Equal(t, Mods.mods[i].Name, getMod.Name, "GetMod.Name should be equal")
			assert.Equal(t, Mods.mods[i].Tag, getMod.Tag, "GetMod.Tag should be equal")
			assert.Equal(t, Mods.mods[i].Description, getMod.Description, "GetMod.Description should be equal")
			assert.Equal(t, Mods.mods[i].FunctionCode(), getMod.FunctionCode(), "GetMod.FunctionCode() should be equal")
			assert.Equal(t, Mods.mods[i].FunctionNames, getMod.FunctionNames, "GetMod.FunctionNames should be equal")

		}
		assert.Empty(t, Mods.GetMod("notHere"), "GetMod with invalid tag should be nil")
		// Test GetRules
		assert.Equal(t, Mods.rules, Mods.GetRules(), "GetRules should be equal")
		// Test Index
		assert.Equal(t, Mods.index, Mods.Index(), "Index should be equal")
		expectedModTag := "mod1,mod2,mod3"
		assert.Equal(t, expectedModTag, NewModTag([]string{"mod1", "mod2", "mod3"}), "Index should be equal")

		// Test TagsAreValid
		_, err := Mods.TagsAreValid("txto,txto")
		assert.Error(t, err, "MOD tags should be invalid with duplicate tags")
		_, err = Mods.TagsAreValid("vsoo,vspubsu,vspubow")
		assert.Error(t, err, "MOD tags should be invalid with Single MOD rule")
		// Test InjectMODs/injectMOD
		_, _, err = Mods.InjectMODs("vsoo", "Function InitializePrivate() bool\n10 IF GOTO ELSE THEN 10")
		assert.Error(t, err, "InjectMODs should error with invalid SC code")
		_, _, err = Mods.InjectMODs("notHere", "")
		assert.Error(t, err, "InjectMODs should error with invalid MOD tag")
		_, _, err = Mods.injectMOD("notHere", "")
		assert.Error(t, err, "injectMOD should error with invalid MOD tag")
		_, _, err = Mods.InjectMODs(Mods.Tag(3), TELA_INDEX_1)
		assert.NoError(t, err, "injectMOD should not error with valid code and %s tag", Mods.Tag(3))
		_, _, err = Mods.InjectMODs(Mods.Tag(4), TELA_INDEX_1)
		assert.NoError(t, err, "injectMOD should not error with valid code and %s tag", Mods.Tag(4))
		_, _, err = Mods.InjectMODs(Mods.Tag(5), TELA_INDEX_1)
		assert.NoError(t, err, "injectMOD should not error with valid code and %s tag", Mods.Tag(5))

		// Test Add
		err = Mods.Add(
			MODClass{
				Name:  "Transfers",
				Tag:   "tx", // MODClass tag exists already
				Rules: []MODClassRule{},
			},
			[]MOD{},
		)
		assert.Error(t, err, "Adding duplicate MODClass should error")

		err = Mods.Add(
			MODClass{
				Name:  "NewClass",
				Tag:   "nc",
				Rules: []MODClassRule{},
			},
			[]MOD{{Name: "MOD1"}, {Name: "MOD2"}}, // empty mod tags
		)
		assert.Error(t, err, "Adding MODs without tag should error")

		err = Mods.Add(
			MODClass{
				Name:  "NewClass",
				Tag:   "nc",
				Rules: []MODClassRule{},
			},
			[]MOD{{Name: "MOD1", Tag: "nc1"}, {Name: "MOD2", Tag: "nc2"}}, // No function code
		)
		assert.Error(t, err, "Adding MODs without function code should error")

		err = Mods.Add(
			MODClass{
				Name:  "NewClass",
				Tag:   "nc",
				Rules: []MODClassRule{},
			},
			[]MOD{{Name: "MOD1", Tag: "nc11", FunctionCode: func() string { return "ThisIsNotASC" }}}, // invalid function code
		)
		assert.Error(t, err, "Adding MODs with invalid function code should error")

		err = Mods.Add(
			MODClass{
				Name:  "NewClass",
				Tag:   "nc",
				Rules: []MODClassRule{},
			},
			[]MOD{{Name: "MOD1", Tag: "nc1", FunctionCode: func() string { return TELA_MOD_1_TXTO }}}, // empty function names
		)
		assert.Error(t, err, "Adding MODs without function names should error")

		err = Mods.Add(
			MODClass{
				Name:  "NewClass",
				Tag:   "nc",
				Rules: []MODClassRule{},
			},
			[]MOD{
				{Name: "MOD1", Tag: "nc1", FunctionCode: func() string { return TELA_MOD_1_TXTO }, FunctionNames: []string{"Name1", "Name2"}}, // duplicate mod tags
				{Name: "MOD2", Tag: "nc1", FunctionCode: func() string { return TELA_MOD_1_TXTO }, FunctionNames: []string{"Name1", "Name2"}},
			},
		)
		assert.Error(t, err, "Adding MODs with duplicate tags should error")

		err = Mods.Add(
			MODClass{
				Name:  "NewClass",
				Tag:   "nc",
				Rules: []MODClassRule{},
			},
			[]MOD{
				{Name: "MOD1", Tag: "nc1", FunctionCode: func() string { return TELA_MOD_1_TXTO }, FunctionNames: []string{"Name1", "Name2"}}, // duplicate mod names
				{Name: "MOD1", Tag: "nc2", FunctionCode: func() string { return TELA_MOD_1_TXTO }, FunctionNames: []string{"Name1", "Name2"}},
			},
		)
		assert.Error(t, err, "Adding MODs with duplicate names should error")

		err = Mods.Add(
			MODClass{
				Name:  "NewClass",
				Tag:   "n",
				Rules: []MODClassRule{Mods.rules[0], Mods.rules[1]}, // Conflicting MODClass rules
			},
			[]MOD{{Name: "MOD1", Tag: "nc1", FunctionCode: func() string { return TELA_MOD_1_TXTO }, FunctionNames: []string{"Name"}}},
		)
		assert.Error(t, err, "Adding MODs with conflicting rules should error")

		err = Mods.Add(
			MODClass{
				Name:  "NewClass",
				Tag:   "n",
				Rules: []MODClassRule{},
			},
			[]MOD{
				{Name: "MOD1", Tag: "nc1", FunctionCode: func() string { return TELA_MOD_1_TXTO }, FunctionNames: []string{"Name1", "Name2"}},
				{Name: "MOD2", Tag: "nc2", FunctionCode: func() string { return TELA_MOD_1_TXTO }, FunctionNames: []string{"Name1", "Name2"}},
			},
		)
		assert.NoError(t, err, "Adding a valid MODClass and MOD should not error: %s", err)

		assert.Len(t, Mods.classes, len(Mods.index), "MODClass and index should be the same len")

		// Force index error by completing a MODClass without data
		Mods.index = append(Mods.index, 0)
		err = Mods.Verify()
		assert.Error(t, err, "Invalid MODClass index should error")

		// If an invalid MOD was added
		Mods.mods = append(Mods.mods, MOD{
			Name:        "BadMod",
			Tag:         "bm1",
			Description: "",
			FunctionCode: func() string {
				return "NOTDVMCODE\\"
			},
			FunctionNames: []string{""},
		})

		_, _, err = Mods.InjectMODs("bm1", TELA_INDEX_1)
		assert.Error(t, err, "Invalid DVM code should not be injected")
	})
}

// TestMODs will test the DVM code in TELA-MOD-1
func TestMODs(t *testing.T) {
	if os.Getenv("RUN_MOD_TEST") != "true" {
		t.Skipf("Use %q to run MODs test", "RUN_MOD_TEST=true go test -run TestMODs -v")
	}

	endpoint, _, wallets := createTestEnvironment(t)

	destination := "deto1qyvyeyzrcm2fzf6kyq7egkes2ufgny5xn77y6typhfx9s7w3mvyd5qqynr5hx"

	assetContractCode := `Function InitializePrivate() Uint64
10 SEND_ASSET_TO_ADDRESS(SIGNER(), 100000000, SCID())
20 RETURN 0
End Function`

	assetArgs := rpc.Arguments{
		rpc.Argument{Name: rpc.SCACTION, DataType: rpc.DataUint64, Value: uint64(rpc.SC_INSTALL)},
		rpc.Argument{Name: rpc.SCCODE, DataType: rpc.DataString, Value: assetContractCode},
	}

	assetSCID, err := transfer0(wallets[2], 2, assetArgs)
	if err != nil {
		t.Fatalf("Failed to install asset scid: %s", err)
	}
	t.Logf("Simulator asset SC installed: %s", assetSCID)

	variableStoreMODs := []string{
		"vsoo",
		"vsooim",
		"vspubsu",
		"vspubow",
		"vspubim",
	}

	transferMODs := []string{
		"txdwd,txdwa,txto", // Can do these all on one SC
	}

	installAnon := []string{
		"vspubow,txdwd,txdwa,txto",
	}

	// Install smart contracts with MODs
	var installSCIDs []string
	installTags := append(variableStoreMODs, transferMODs...)
	installTags = append(installTags, installAnon...)
	for i, modTag := range installTags {
		index := &INDEX{
			DURL: fmt.Sprintf("mod.test.%s.tela", modTag),
			Mods: modTag,
			DOCs: []string{},
			Headers: Headers{
				NameHdr:  fmt.Sprintf("MOD %d Test", i),
				DescrHdr: modTag,
			},
		}

		_, err := Mods.TagsAreValid(index.Mods)
		if err != nil {
			t.Fatalf("MODs are not valid")
		}

		ringsize := uint64(2)
		if i == len(variableStoreMODs)+len(transferMODs) {
			ringsize = 16
		}

		tx, err := retry(t, fmt.Sprintf("INDEX %s install", modTag), func() (string, error) {
			return Installer(wallets[i], ringsize, index)
		})

		assert.NoError(t, err, "Install %s should not error: %s", index.Mods, err)

		_, err = retry(t, fmt.Sprintf("confirming install TX %s", tx), func() (string, error) {
			return getContractCode(tx, endpoint)
		})
		if err != nil {
			t.Fatalf("Could not confirm INDEX %s TX %s: %s", index.Mods, tx, err)
		}

		t.Logf("Simulator MOD %q SC installed: %s", index.Mods, tx)
		installSCIDs = append(installSCIDs, tx)
	}

	time.Sleep(sleepFor)

	// // Test Variable Store functions
	t.Run("Variable Store", func(t *testing.T) {
		for i := 0; i < len(variableStoreMODs); i++ {
			kv := fmt.Sprintf("%d", i)
			varK := fmt.Sprintf("var_%s", kv)

			// Call SetVar functions
			_, err := SetVar(wallets[i], installSCIDs[i], kv, kv)
			assert.NoError(t, err, "Owner calling SetVar %d should not error: %s", i, err)
			_, err = SetVar(wallets[i+1], installSCIDs[i], kv, kv)
			if i < 2 {
				assert.Error(t, err, "Calling SetVar %d when not owner should error", i)
				if i == 1 {
					varK = fmt.Sprintf("i%s", varK)
				}
			} else {
				assert.NoError(t, err, "Wallet calling SetVar %d should not error: %s", i, err)
				varK = fmt.Sprintf("var_%s_%s", wallets[i+1].GetAddress().String(), kv)
				if i == 4 {
					varK = fmt.Sprintf("i%s", varK)
				}
			}

			time.Sleep(sleepFor)
			v, err := retry(t, fmt.Sprintf("confirming set var %d", i), func() (string, error) {
				return getContractVar(installSCIDs[i], varK, endpoint)
			})
			assert.NoError(t, err, "SetVar %d variable should not error: %s", i, err)

			var exists bool
			// Check that variable was set
			_, exists, err = KeyExists(installSCIDs[i], varK, endpoint)
			assert.NoError(t, err, "KeyExists %d should not error: %s", i, err)
			assert.True(t, exists, "Key %s should exist", kv)

			switch i {
			case 1, 4:
				_, err = SetVar(wallets[i], installSCIDs[i], kv, kv)
				assert.Error(t, err, "Overwriting owner immutable SetVar %d should error", i)
				if i == 4 {
					_, err = SetVar(wallets[i+1], installSCIDs[i], kv, kv)
					assert.Error(t, err, "Overwriting wallet immutable SetVar %d should error", i)
				}
			default:
				// Call DeleteVar functions
				_, err = DeleteVar(wallets[i+1], installSCIDs[i], kv)
				assert.Error(t, err, "Calling DeleteVar %d when not owner should error", i)

				// If not owner address prefix is needed
				deleteKey := kv
				if i > 0 {
					deleteKey = strings.TrimPrefix(varK, "var_")
				}

				_, err = DeleteVar(wallets[i], installSCIDs[i], deleteKey)
				assert.NoError(t, err, "Calling DeleteVar %d should not error: %s", i, err)

				time.Sleep(sleepFor)
				// Check that variable was deleted
				_, exists, err = KeyExists(installSCIDs[i], varK, endpoint)
				assert.NoError(t, err, "KeyExists %d should not error: %s", i, err)
				assert.False(t, exists, "Key %s should not exist after deletion", kv)
				assert.NotEmpty(t, v, "SetVar %d Value should not be empty", i)
			}
		}
	})

	// // Test Transfers functions
	t.Run("Transfers", func(t *testing.T) {
		amount := uint64(100)
		tIndex := len(variableStoreMODs)
		scid := installSCIDs[tIndex]
		transferTags := strings.Split(transferMODs[0], ",")
		for _, tag := range transferTags {
			_, functionNames := Mods.Functions(tag)
			switch tag {
			case "txdwd":
				{
					// No DERO balance in SC
					args := rpc.Arguments{
						rpc.Argument{Name: "entrypoint", DataType: rpc.DataString, Value: "WithdrawDero"},
						rpc.Argument{Name: rpc.SCID, DataType: rpc.DataHash, Value: crypto.HashHexToHash(scid)},
						rpc.Argument{Name: rpc.SCACTION, DataType: rpc.DataUint64, Value: uint64(rpc.SC_CALL)},
						rpc.Argument{Name: "amt", DataType: rpc.DataUint64, Value: amount},
					}

					_, err = transfer0(wallets[tIndex], 2, args)
					assert.Error(t, err, "%s should error when no DERO balance in smart contract", "WithdrawDero")
				}

				for _, fn := range functionNames {
					baseArgs := rpc.Arguments{
						rpc.Argument{Name: "entrypoint", DataType: rpc.DataString, Value: fn},
						rpc.Argument{Name: rpc.SCID, DataType: rpc.DataHash, Value: crypto.HashHexToHash(scid)},
						rpc.Argument{Name: rpc.SCACTION, DataType: rpc.DataUint64, Value: uint64(rpc.SC_CALL)},
					}

					switch fn {
					case "DepositDero":
						var transfers []rpc.Transfer
						transfers0 := append(transfers, rpc.Transfer{Destination: destination, Burn: 0})
						transfersAmt := append(transfers, rpc.Transfer{Destination: destination, Burn: amount})
						for i := 0; i < 2; i++ {
							_, err := Transfer(wallets[i+2], 2, transfersAmt, baseArgs)
							assert.NoError(t, err, "%s %d should not error: %s", fn, i, err)
							time.Sleep(sleepFor)
						}
						_, err := Transfer(wallets[0], 2, transfers0, baseArgs)
						assert.Error(t, err, "%s zero should error", fn)
					case "WithdrawDero":
						amountArgs := append(baseArgs, rpc.Argument{Name: "amt", DataType: rpc.DataUint64, Value: amount * 2})
						overAmountArgs := append(baseArgs, rpc.Argument{Name: "amt", DataType: rpc.DataUint64, Value: amount * 3})
						_, err = transfer0(wallets[tIndex-1], 2, amountArgs)
						assert.Error(t, err, "%s should error when not owner", fn)
						_, err = transfer0(wallets[tIndex], 2, overAmountArgs)
						assert.Error(t, err, "%s should error when over balance amount", fn)
						_, err := transfer0(wallets[tIndex], 2, amountArgs)
						assert.NoError(t, err, "%s should not error: %s", fn, err)
					default:
						t.Fatalf("Invalid %s function: %s", tag, fn)
					}

					time.Sleep(sleepFor)
					v, err := retry(t, fmt.Sprintf("confirming transfer %s", fn), func() (string, error) {
						return getContractVar(scid, "balance_dero", endpoint)
					})
					assert.NoError(t, err, "Getting %s variable should not error: %s", fn, err)

					switch fn {
					case "DepositDero":
						assert.Equal(t, "200", v, "%s result should be equal", tag)
						time.Sleep(sleepFor)
					case "WithdrawDero":
						assert.Equal(t, "0", v, "%s result should be equal", tag)
					}
				}
			case "txdwa":
				{
					// No asset balance in SC
					args := rpc.Arguments{
						rpc.Argument{Name: "entrypoint", DataType: rpc.DataString, Value: "WithdrawAsset"},
						rpc.Argument{Name: rpc.SCID, DataType: rpc.DataHash, Value: crypto.HashHexToHash(scid)},
						rpc.Argument{Name: rpc.SCACTION, DataType: rpc.DataUint64, Value: uint64(rpc.SC_CALL)},
						rpc.Argument{Name: "amt", DataType: rpc.DataUint64, Value: amount},
						rpc.Argument{Name: "scid", DataType: rpc.DataString, Value: assetSCID},
					}

					_, err = Transfer(wallets[tIndex], 2, nil, args)
					assert.Error(t, err, "%s should error when no asset balance in smart contract", "WithdrawAsset")
				}

				for _, fn := range functionNames {
					baseArgs := rpc.Arguments{
						rpc.Argument{Name: "entrypoint", DataType: rpc.DataString, Value: fn},
						rpc.Argument{Name: rpc.SCID, DataType: rpc.DataHash, Value: crypto.HashHexToHash(scid)},
						rpc.Argument{Name: rpc.SCACTION, DataType: rpc.DataUint64, Value: uint64(rpc.SC_CALL)},
					}

					switch fn {
					case "DepositAsset":
						var transfers []rpc.Transfer
						amountArgs := append(baseArgs, rpc.Argument{Name: "scid", DataType: rpc.DataString, Value: assetSCID})
						transfers0 := append(transfers, rpc.Transfer{Destination: destination, SCID: crypto.HashHexToHash(assetSCID), Burn: 0})
						transfersAmt := append(transfers, rpc.Transfer{Destination: destination, SCID: crypto.HashHexToHash(assetSCID), Burn: amount})
						_, err := Transfer(wallets[2], 2, transfersAmt, amountArgs)
						assert.NoError(t, err, "%s should not error: %s", fn, err)
						_, err = Transfer(wallets[3], 2, transfersAmt, amountArgs)
						assert.Error(t, err, "%s should error when no wallet asset balance", fn)
						_, err = Transfer(wallets[2], 2, transfers0, amountArgs)
						assert.Error(t, err, "%s zero should error", fn)
					case "WithdrawAsset":
						amountArgs := append(baseArgs, rpc.Argument{Name: "scid", DataType: rpc.DataString, Value: assetSCID})
						amountArgs = append(amountArgs, rpc.Argument{Name: "amt", DataType: rpc.DataUint64, Value: amount})
						_, err = transfer0(wallets[tIndex-1], 2, amountArgs)
						assert.Error(t, err, "%s should error when not owner", fn)
						_, err := transfer0(wallets[tIndex], 2, amountArgs)
						assert.NoError(t, err, "%s should not error: %s", fn, err)
						time.Sleep(sleepFor)
						_, err = transfer0(wallets[tIndex], 2, amountArgs) // this errors on daemon if called before first withdraw
						assert.Error(t, err, "%s should error when over asset balance amount", fn)
					default:
						t.Fatalf("Invalid %s function: %s", tag, fn)
					}

					time.Sleep(sleepFor)
					v, err := retry(t, fmt.Sprintf("confirming transfer %s", fn), func() (string, error) {
						return getContractVar(scid, fmt.Sprintf("balance_%s", assetSCID), endpoint)
					})
					assert.NoError(t, err, "Getting %s variable should not error: %s", fn, err)

					switch fn {
					case "DepositAsset":
						assert.Equal(t, "100", v, "%s result should be equal", tag)
						time.Sleep(sleepFor)
					case "WithdrawAsset":
						assert.Equal(t, "0", v, "%s result should be equal", tag)
					}
				}
			case "txto":
				{
					// There is no tempowner
					args := rpc.Arguments{
						rpc.Argument{Name: "entrypoint", DataType: rpc.DataString, Value: "ClaimOwnership"},
						rpc.Argument{Name: rpc.SCID, DataType: rpc.DataHash, Value: crypto.HashHexToHash(scid)},
						rpc.Argument{Name: rpc.SCACTION, DataType: rpc.DataUint64, Value: uint64(rpc.SC_CALL)},
					}

					_, err = transfer0(wallets[4], 2, args)
					assert.Error(t, err, "%s should error when there is no tempowner", "ClaimOwnership")
				}

				for _, fn := range functionNames {
					baseArgs := rpc.Arguments{
						rpc.Argument{Name: "entrypoint", DataType: rpc.DataString, Value: fn},
						rpc.Argument{Name: rpc.SCID, DataType: rpc.DataHash, Value: crypto.HashHexToHash(scid)},
						rpc.Argument{Name: rpc.SCACTION, DataType: rpc.DataUint64, Value: uint64(rpc.SC_CALL)},
					}

					switch fn {
					case "TransferOwnership":
						transferArgsValid := append(baseArgs, rpc.Argument{Name: "addr", DataType: rpc.DataString, Value: wallets[2].GetAddress().String()})
						transferArgsInvalid := append(baseArgs, rpc.Argument{Name: "addr", DataType: rpc.DataString, Value: "address"})
						_, err = transfer0(wallets[4], 2, transferArgsValid)
						assert.Error(t, err, "%s should error when not owner", fn)
						_, err = transfer0(wallets[tIndex], 2, transferArgsInvalid)
						assert.Error(t, err, "%s to invalid address should error", fn)
						_, err = transfer0(wallets[tIndex], 2, transferArgsValid)
						assert.NoError(t, err, "%s should not error: %s", fn, err)
					case "ClaimOwnership":
						_, err = transfer0(wallets[4], 2, baseArgs)
						assert.Error(t, err, "%s should error when not tempowner", fn)
						_, err := transfer0(wallets[2], 2, baseArgs)
						assert.NoError(t, err, "%s should not error: %s", fn, err)
					default:
						t.Fatalf("Invalid %s function: %s", tag, fn)
					}

					time.Sleep(sleepFor)
					v, err := retry(t, fmt.Sprintf("confirming transfer %s", fn), func() (string, error) {
						return getContractVar(scid, "owner", endpoint)
					})
					assert.NoError(t, err, "Getting %s variable should not error: %s", fn, err)

					switch fn {
					case "TransferOwnership":
						assert.Equal(t, wallets[tIndex].GetAddress().String(), v, "%s result should be equal", tag)
						time.Sleep(sleepFor)
					case "ClaimOwnership":
						assert.Equal(t, wallets[2].GetAddress().String(), v, "%s result should be equal", tag)
					}
				}
			default:
				t.Fatalf("All transfer test tags should be valid")
			}
		}
	})

	// // Test functions on INDEX installed as anon
	t.Run("Anon INDEX", func(t *testing.T) {
		scid := installSCIDs[len(installSCIDs)-1]
		_, err := SetVar(wallets[0], scid, "kv", "kv")
		assert.Error(t, err, "SetVar on anon INDEX should error")
		_, err = DeleteVar(wallets[0], scid, "kv")
		assert.Error(t, err, "DeleteVar on anon INDEX should error")

		var transfers []rpc.Transfer
		baseArgs := rpc.Arguments{
			rpc.Argument{Name: rpc.SCID, DataType: rpc.DataHash, Value: crypto.HashHexToHash(scid)},
			rpc.Argument{Name: rpc.SCACTION, DataType: rpc.DataUint64, Value: uint64(rpc.SC_CALL)},
		}

		depositDeroArgs := append(baseArgs, rpc.Argument{Name: "entrypoint", DataType: rpc.DataString, Value: "DepositDero"})
		depositDeroTransfer := append(transfers, rpc.Transfer{Destination: destination, Burn: 100})
		_, err = Transfer(wallets[1], 2, depositDeroTransfer, depositDeroArgs)
		assert.Error(t, err, "DepositDero on anon INDEX should error")

		withdrawDeroArgs := append(baseArgs, rpc.Argument{Name: "entrypoint", DataType: rpc.DataString, Value: "WithdrawDero"})
		withdrawDeroArgs = append(withdrawDeroArgs, rpc.Argument{Name: "amt", DataType: rpc.DataUint64, Value: uint64(100)})
		_, err = transfer0(wallets[1], 2, withdrawDeroArgs)
		assert.Error(t, err, "WithdrawDero on anon INDEX should error")

		depositAssetArgs := append(baseArgs, rpc.Argument{Name: "entrypoint", DataType: rpc.DataString, Value: "DepositAsset"})
		depositAssetArgs = append(depositAssetArgs, rpc.Argument{Name: "scid", DataType: rpc.DataString, Value: assetSCID})
		depositAssetTransfer := append(transfers, rpc.Transfer{Destination: destination, SCID: crypto.HashHexToHash(assetSCID), Burn: 100})
		_, err = Transfer(wallets[1], 2, depositAssetTransfer, depositAssetArgs)
		assert.Error(t, err, "DepositAsset on anon INDEX should error")

		withdrawAssetArgs := append(baseArgs, rpc.Argument{Name: "entrypoint", DataType: rpc.DataString, Value: "WithdrawAsset"})
		withdrawAssetArgs = append(withdrawAssetArgs, rpc.Argument{Name: "amt", DataType: rpc.DataUint64, Value: uint64(100)})
		_, err = transfer0(wallets[1], 2, withdrawAssetArgs)
		assert.Error(t, err, "WithdrawAsset on anon INDEX should error")

		transferOwnershipArgs := append(baseArgs, rpc.Argument{Name: "entrypoint", DataType: rpc.DataString, Value: "TransferOwnership"})
		transferOwnershipArgs = append(transferOwnershipArgs, rpc.Argument{Name: "addr", DataType: rpc.DataString, Value: wallets[1].GetAddress().String()})
		_, err = transfer0(wallets[1], 2, transferOwnershipArgs)
		assert.Error(t, err, "TransferOwnership on anon INDEX should error")

		claimOwnershipArgs := append(baseArgs, rpc.Argument{Name: "entrypoint", DataType: rpc.DataString, Value: "ClaimOwnership"})
		_, err = transfer0(wallets[1], 2, claimOwnershipArgs)
		assert.Error(t, err, "ClaimOwnership on anon INDEX should error")
	})
}

// Set up test environment for DERO simulator
func createTestEnvironment(t *testing.T) (endpoint, datashards string, wallets []*walletapi.Wallet_Disk) {
	// Cleanup directories used for package test
	walletName := "tela_sim"
	testPath := filepath.Join(mainPath, testDir)
	walletPath := filepath.Join(testPath, "tela_sim_wallets")
	datashards = filepath.Join(testPath, "datashards")

	err := SetShardPath(testPath)
	if err != nil {
		t.Fatalf("Could not set test directory: %s", err)
	}

	t.Cleanup(func() {
		ShutdownTELA()
		os.RemoveAll(datashards)
		os.RemoveAll(walletPath)
	})

	os.RemoveAll(datashards)
	os.RemoveAll(walletPath)

	endpoint = "127.0.0.1:20000"
	globals.Arguments["--testnet"] = true
	globals.Arguments["--simulator"] = true
	globals.Arguments["--daemon-address"] = endpoint
	globals.InitNetwork()

	// Create simulator wallets to use for contract installs
	for i, seed := range walletSeeds {
		w, err := createTestWallet(fmt.Sprintf("%s%d", walletName, i), walletPath, seed)
		if err != nil {
			t.Fatalf("Failed to create wallet %d: %s", i, err)
		}

		wallets = append(wallets, w)
		assert.NotNil(t, wallets[i], "Wallet %s should not be nil", i)
		wallets[i].SetDaemonAddress(endpoint)
		wallets[i].SetOnlineMode()
	}

	if err := walletapi.Connect(endpoint); err != nil {
		t.Fatalf("Failed to connect wallets to simulator: %s", err)
	}

	return
}

// Create test wallet for simulator
func createTestWallet(name, dir, seed string) (wallet *walletapi.Wallet_Disk, err error) {
	seed_raw, err := hex.DecodeString(seed)
	if err != nil {
		return
	}

	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return
	}

	filename := filepath.Join(dir, name)

	os.Remove(filename)

	wallet, err = walletapi.Create_Encrypted_Wallet(filename, "", new(crypto.BNRed).SetBytes(seed_raw))
	if err != nil {
		return
	}

	wallet.SetNetwork(false)
	wallet.Save_Wallet()

	return
}

// Read local file for tests
func readFile(elem ...string) (string, error) {
	path := mainPath
	for _, e := range elem {
		path = filepath.Join(path, e)
	}

	file, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return string(file), nil
}

// Retry wrapper for various test functions
func retry(t *testing.T, print string, f func() (string, error)) (result string, err error) {
	for retry := 0; retry < 3; retry++ {
		if result, err = f(); err == nil {
			return
		}

		if !strings.HasPrefix(print, "confirming") {
			t.Logf("Retrying %s: %s", print, err)
		}

		time.Sleep(sleepFor)
	}

	err = fmt.Errorf("max retries exceeded for %s: %s", print, err)

	return
}

// Watch if existing var STORE v changes value to e
func varChanged(scid, v, e, endpoint string) (scv string, err error) {
	for retry := 0; retry < 3; retry++ {
		scv, err = getContractVar(scid, v, endpoint)
		if err != nil {
			return
		}

		if scv == e {
			err = nil
			return
		}

		time.Sleep(sleepFor)
	}

	err = fmt.Errorf("%q STORE did not change", v)

	return
}
