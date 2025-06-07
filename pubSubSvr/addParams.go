package main

import (
	"fmt"

	"github.com/nickwells/check.mod/v2/check"
	"github.com/nickwells/param.mod/v6/param"
	"github.com/nickwells/param.mod/v6/psetter"
	"github.com/nickwells/pusu.mod/pusu"
	"github.com/nickwells/slogsetter.mod/slogsetter"
)

const (
	paramNamePort           = "port"
	paramNameLogLevel       = "log-level"
	paramNameStatusInterval = "status-interval"

	paramNameAllowedNamespaces = "namespaces-allowed"
	paramNameNamespacePrefixes = "namespace-prefixes"
)

// addParams adds the parameters for this program
func addParams(prog *prog) param.PSetOptFunc {
	return func(ps *param.PSet) error {
		ps.Add(paramNamePort,
			psetter.Int[int]{
				Value: &prog.port,
				Checks: []check.ValCk[int]{
					check.ValGE(0),
				},
			},
			"the port number for the server to listen on",
			param.Attrs(param.MustBeSet))

		ps.Add(paramNameLogLevel,
			slogsetter.Level{
				Value: &prog.logLevel,
			},
			"the level of logging")

		ps.Add(paramNameStatusInterval,
			psetter.Duration{
				Value: &prog.statusReportingInterval,
			},
			"the time to wait between status reports")

		allowedNSParam := ps.Add(paramNameAllowedNamespaces,
			psetter.Map[pusu.Namespace]{
				Value: (*map[pusu.Namespace]bool)(&prog.nsRules.allowed),
			},
			"the namespaces to allow clients to connect with")

		nsPfxParam := ps.Add(paramNameNamespacePrefixes,
			psetter.StrList[string]{
				Value: &prog.nsRules.prefixes,
			},
			"the prefixes which a namespace must have to allow"+
				" clients to connect with it")

		ps.AddFinalCheck(func() error {
			prog.progName = ps.ProgName()

			return nil
		})

		ps.AddFinalCheck(func() error {
			if allowedNSParam.HasBeenSet() &&
				nsPfxParam.HasBeenSet() {
				return fmt.Errorf(
					"you may set the %q parameter or the %q parameter"+
						" or neither but not both",
					paramNameAllowedNamespaces,
					paramNameNamespacePrefixes)
			}

			return nil
		})

		ps.AddFinalCheck(func() error {
			if allowedNSParam.HasBeenSet() &&
				nsPfxParam.HasBeenSet() {
				return fmt.Errorf(
					"you may set the %q parameter or the %q parameter"+
						" or neither but not both",
					paramNameAllowedNamespaces,
					paramNameNamespacePrefixes)
			}

			return nil
		})

		ps.AddFinalCheck(func() error {
			return prog.nsRules.checkPrefixes()
		})

		return nil
	}
}
