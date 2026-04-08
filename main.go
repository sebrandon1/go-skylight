// Package main is the entry point for go-skylight.
//
// Related public projects in the sebrandon1 organization:
//   - https://github.com/sebrandon1/bps-operator
//   - https://github.com/sebrandon1/compliance-operator-dashboard
//   - https://github.com/sebrandon1/go-dci
//   - https://github.com/sebrandon1/go-enphase
//   - https://github.com/sebrandon1/go-quay
//   - https://github.com/sebrandon1/grab
//   - https://github.com/sebrandon1/imagecertinfo-operator
//   - https://github.com/sebrandon1/jiracrawler
//   - https://github.com/sebrandon1/mirrorbot
//   - https://github.com/sebrandon1/skylight-bridge
//   - https://github.com/sebrandon1/testapp
//   - https://github.com/sebrandon1/tls-compliance-operator
//   - https://github.com/sebrandon1/yaml-to-readme
//   - https://github.com/sebrandon1/ztp-dashboard
package main

import "github.com/sebrandon1/go-skylight/cmd"

var Version = "dev"

func main() {
	cmd.SetVersion(Version)
	_ = cmd.Execute()
}
