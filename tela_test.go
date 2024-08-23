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
	// Cleanup directories used for package test
	walletName := "tela_sim"
	testPath := filepath.Join(mainPath, testDir)
	walletPath := filepath.Join(testPath, "tela_sim_wallets")
	datashards := filepath.Join(testPath, "datashards")

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

	endpoint := "127.0.0.1:20000"
	globals.Arguments["--testnet"] = true
	globals.Arguments["--simulator"] = true
	globals.Arguments["--daemon-address"] = endpoint
	globals.InitNetwork()

	// Create simulator wallets to use for contract installs
	var wallets []*walletapi.Wallet_Disk
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

	var docs [3][]string     // Organize DOC scids
	var commitTXIDs []string // TXIDs of updates
	var noCodeTXIDs []string // TXIDs without any SC code

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

				install := &INDEX{
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
			// GetCodeSizeInKB() returns SC code size of 9.0517578125 KB
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
			// GetCodeSizeInKB() returns SC code size of 11.560546875 KB
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
			// GetCodeSizeInKB() returns SC code size of 11.7373046875 KB
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
		update := &INDEX{
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

	t.Run("INDEX lib embed", func(t *testing.T) {
		// Tag INDEX as library
		libraryEmbeds := []*INDEX{
			{
				DURL: "zero.lib",
				DOCs: docs[0],
				Headers: Headers{
					NameHdr:  "library0",
					DescrHdr: "Zero lib",
					IconHdr:  "icon.url",
				},
			},
			{
				DURL: "one.lib",
				DOCs: docs[1],
				Headers: Headers{
					NameHdr:  "library1",
					DescrHdr: "One lib",
					IconHdr:  "icon.url",
				},
			},
		}

		var librarySCIDs []string
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

			t.Logf("Simulator INDEX %d library SC installed %s: %s", i, il.NameHdr, tx)
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

			t.Logf("Simulator INDEX %d embed SC installed %s: %s", i, ie.NameHdr, tx)
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

	// // Test internal functions
	t.Run("Internal", func(t *testing.T) {
		// Invalid docType language
		assert.False(t, IsAcceptedLanguage("TELA-NotAcceptedLanguage-1"), "Should not be a accepted language")

		// Languages are case sensitive
		assert.False(t, IsAcceptedLanguage("TELA-html-1"), "Language should be case sensitive")

		// Parse empty document with no multiline comment
		assert.Error(t, parseAndSaveTELADoc("", "", ""), "Should not be able to parse this document")
		// Parse invalid docType
		assert.Error(t, parseAndSaveTELADoc("", "/*\n*/", "invalid"), "DocType should be invalid")
		// Save to invalid path
		assert.Error(t, parseAndSaveTELADoc(filepath.Join(testPath, "app2", "index.html", "filename.html"), TELA_DOC_1, DOC_HTML), "Path should be invalid")
		// Parse types not installed in tests
		assert.NoError(t, parseAndSaveTELADoc(filepath.Join(datashards, "json.json"), TELA_DOC_1, DOC_JSON), "DOC_JSON should be valid")
		assert.NoError(t, parseAndSaveTELADoc(filepath.Join(datashards, "markdown.md"), TELA_DOC_1, DOC_MD), "DOC_MD should be valid")
		assert.NoError(t, parseAndSaveTELADoc(filepath.Join(datashards, "other.jsx"), TELA_DOC_1, DOC_STATIC), "DOC_STATIC should be valid")

		// decodeHexString return non hex
		expectedAddress := "deto1qy87ghfeeh6n6tdxtgh7yuvtp6wh2uwfzvz7vjq0krjy4smmx62j5qgqht7t3"
		assert.Equal(t, expectedAddress, decodeHexString(expectedAddress), "decodeHexString should return a DERO address when passed")

		// Test cloneDOC
		_, err := cloneDOC(scDoesNotExist, "", "", endpoint)
		assert.Error(t, err, "cloneDOC should error with invalid scid")
		_, err = cloneDOC(nameservice, "", "", endpoint)
		assert.Error(t, err, "cloneDOC with NON TELA should error")

		// Test getTXID
		_, err = getTXID(scDoesNotExist, endpoint)
		assert.Error(t, err, "Getting invalid TX should error")
		_, err = getTXID(scDoesNotExist, "")
		assert.Error(t, err, "Getting invalid TX with invalid endpoint should error")

		// Test cloneINDEXAtCommit
		_, err = cloneINDEXAtCommit("", "", tela.path.clone(), endpoint) // invalid scid
		assert.Error(t, err, "cloneINDEXAtCommit with invalid SCID should error")
		_, err = cloneINDEXAtCommit(nameservice, commitTXIDs[0], tela.path.clone(), endpoint) // NON TELA
		assert.Error(t, err, "cloneINDEXAtCommit with NON TELA should error")

		// Test extractCodeFromTXID
		_, err = extractCodeFromTXID(hex.EncodeToString([]byte("someHex"))) // random hex
		assert.Error(t, err, "Extracting hex without SC code should error")
		_, err = extractCodeFromTXID(hex.EncodeToString([]byte("Function "))) // malformed function in hex
		assert.Error(t, err, "Extracting hex malformed SC code should error")
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

	// Test parse functions
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
		assert.Equal(t, DOC_STATIC, ParseDocType("read.txt"), "Filename should parse as %s", DOC_STATIC)
		assert.Equal(t, DOC_STATIC, ParseDocType("static.tsx"), "Filename should parse as a %s", DOC_STATIC)
		assert.Equal(t, DOC_STATIC, ParseDocType("LICENSE"), "LICENSE should parse as a %s", DOC_STATIC)
		assert.Equal(t, "", ParseDocType("read"), "Filename should not parse as a DOC")

		// Test ParseINDEXForDOCs
		scids, err := ParseINDEXForDOCs(TELA_INDEX_1)
		assert.NoError(t, err, "ParseINDEXForDOCs should not error with valid contract: %s", err)
		assert.Len(t, scids, 1, "There should be one scid")
		_, err = ParseINDEXForDOCs(TELA_DOC_1)
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

		_, err = EqualSmartContracts(code, code)
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

	// Test ratings functions
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

		// Invalid daemon address
		_, err := ServeTELA(validSCIDs[0], "")
		assert.Error(t, err, "Daemon address on getContractVar should not have connected")
		_, err = getContractCode(validSCIDs[0], "")
		assert.Error(t, err, "Daemon address on getContractCode should not have connected")
		_, err = getContractVars(validSCIDs[0], "")
		assert.Error(t, err, "Daemon address on getContractVars should not have connected")

		// No servers should be started on errors
		assert.Empty(t, GetServerInfo(), "Server info should be empty after error tests")
		assert.Nil(t, tela.servers, "Servers should be nil after error tests")

		// tela.servers should be nil here
		ShutdownServer("")
		ShutdownTELA()

		// Reset testnet flag for nil/network/ringsize cases and daemon transfer error
		globals.Arguments["--testnet"] = false
		walletapi.Daemon_Endpoint_Active = ""
		transfer(nil, 0, nil)
		transfer(wallets[0], 0, nil)
		transfer(wallets[0], 256, nil)
	})
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
