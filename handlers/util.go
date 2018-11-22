package handlers

import (
	"regexp"
)

var (
	// label
	labelReg       = regexp.MustCompile("^/[Ll][Aa][Bb][Ee][Ll]")
	labelCancelReg = regexp.MustCompile("^/[Rr][Ee][Mm][Oo][Vv][Ee]-[Ll][Aa][Bb][Ee][Ll]")

	// test
	okToTestReg = regexp.MustCompile("^/[Oo][Kk]-[Tt][Oo]-[Tt][Ee][Ss][Tt]")
	retestReg   = regexp.MustCompile("^/[Rr][Ee][Tt][Ee][Ss][Tt]")
	testReg     = regexp.MustCompile("^/[Tt][Ee][Ss][Tt]")

	// review and approve
	lgtmReg          = regexp.MustCompile("^/[Ll][Gg][Tt][Mm]")
	lgtmCancelReg    = regexp.MustCompile("^/[Ll][Gg][Tt][Mm] [Cc][Aa][Nn][Cc][Ee][Ll]")
	approveReg       = regexp.MustCompile("^/[Aa][Pp][Pp][Rr][Oo][Vv][Ee]")
	approveCancelReg = regexp.MustCompile("^/[Aa][Pp][Pp][Rr][Oo][Vv][Ee] [Cc][Aa][Nn][Cc][Ee][Ll]")
)

const (
	needsOKtoTest = "needs-ok-to-test"
)
