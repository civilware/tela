package logger

import (
	"fmt"
	"strings"
)

// Small TELA ASCII logo
var ASCIISmall = []string{
	Color.red + `           ;;;` + Color.end,
	Color.red + `           ;;;` + Color.end,

	`          .WWW`,
	`::::::::::kMMM`,
	`WMMMMMMMMMMMMM`,
	`WMM,      'MMM`,
	`WMMx......dMMM`,
	`WMMMMMMMMMMMMM`,
	`WMM;`,
	`NWW'`,

	Color.green + `ooo.           ` + Color.end,
	Color.green + `'''            ` + Color.end,
}

// Main TELA ASCII logo
var ASCIIMain = []string{
	Color.red + `               ;;;;;` + Color.end,
	Color.red + `               ;;;;;` + Color.end,
	Color.red + `               ooooo` + Color.end,
	`               MMMMM`,
	`               MMMMM`,
	`NWWWWWWWWWWWWWWMMMMM`,
	`WMMMMMMMMMMMMMMMMMMM`,
	`WMMMM          MMMMM`,
	`WMMMM          MMMMM`,
	`WMMMM;;;;;;;;;;MMMMM`,
	`WMMMMMMMMMMMMMMMMMMM`,
	`WMMMM`,
	`WMMMM`,
	`WMMMM`,

	Color.green + `dxxxx               ` + Color.end,
	Color.green + `loooo               ` + Color.end,
	Color.green + `'''''               ` + Color.end,
}

// Blend ASCII into existing string for display purposes
func ASCIIBlend(ascii, info []string) {
	if info == nil {
		return
	}

	// find longest line and use it as margin
	var marginLen int
	for _, line := range info {
		l := len(line)
		if l > marginLen {
			marginLen = l + 4
		}
	}

	var printed int
	asciiLen := len(ascii)
	if marginLen < len(ascii[0]) {
		marginLen = len(ascii[0]) + 11
	}

	for i, line := range info {
		var margin string
		if i < asciiLen {
			linePad := marginLen - len(line)
			margin = strings.Repeat(" ", linePad)
			fmt.Println(line + margin + ascii[i])
		} else {
			fmt.Println(line)
		}
		printed++
	}

	for printed < asciiLen {
		fmt.Println(strings.Repeat(" ", marginLen-1), ascii[printed])
		printed++
	}
}

// Print TELA ASCII main or small logo
func ASCIIPrint(small bool) {
	ascii := ASCIIMain
	if small {
		ascii = ASCIISmall
	}

	for _, line := range ascii {
		fmt.Println(line)
	}
}
