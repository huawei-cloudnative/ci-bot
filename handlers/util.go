package handlers

import (
	"regexp"
)

var (
	// label
	LabelReg       = regexp.MustCompile("^/[Ll][Aa][Bb][Ee][Ll]")
	LabelCancelReg = regexp.MustCompile("^/[Rr][Ee][Mm][Oo][Vv][Ee]-[Ll][Aa][Bb][Ee][Ll]")

	// test
	OkToTestReg = regexp.MustCompile("^/[Oo][Kk]-[Tt][Oo]-[Tt][Ee][Ss][Tt]")
	RetestReg   = regexp.MustCompile("^/[Rr][Ee][Tt][Ee][Ss][Tt]")
	TestReg     = regexp.MustCompile("^/[Tt][Ee][Ss][Tt]")

	// review and approve
	LgtmReg          = regexp.MustCompile("^/[Ll][Gg][Tt][Mm]")
	LgtmCancelReg    = regexp.MustCompile("^/[Ll][Gg][Tt][Mm] [Cc][Aa][Nn][Cc][Ee][Ll]")
	ApproveReg       = regexp.MustCompile("^/[Aa][Pp][Pp][Rr][Oo][Vv][Ee]")
	ApproveCancelReg = regexp.MustCompile("^/[Aa][Pp][Pp][Rr][Oo][Vv][Ee] [Cc][Aa][Nn][Cc][Ee][Ll]")

	//assign/unassign
	AssignOrUnassing			 = regexp.MustCompile("(?mi)^/(un)?assign(( @?[-\\w]+?)*)\\s*$")
	// CCRegexp parses and validates /cc commands, also used by blunderbuss
	CCRegexp 	= regexp.MustCompile(`(?mi)^/(un)?cc(( +@?[-/\w]+?)*)\s*$`)
)

const (
	needsOKtoTest = "needs-ok-to-test"
)
