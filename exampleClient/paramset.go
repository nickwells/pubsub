package main

import (
	"github.com/nickwells/param.mod/v6/param"
	"github.com/nickwells/param.mod/v6/paramset"
	"github.com/nickwells/pusuparams.mod/pusuparams"
	"github.com/nickwells/verbose.mod/verbose"
	"github.com/nickwells/versionparams.mod/versionparams"
)

// makeParamSet generates the param set ready for parsing
func makeParamSet(prog *prog) *param.PSet {
	return paramset.NewOrPanic(
		verbose.AddParams,
		verbose.AddTimingParams(prog.stack),
		versionparams.AddParams,

		pusuparams.AddPusuParams(prog.cci, ""),
		pusuparams.AddCertInfoParams(&prog.cci.CertInfo, ""),

		addParams(prog),
		addNotes(prog),

		pusuparams.AddNoteNamespaces(),
		pusuparams.AddNoteTopics(),

		param.SetProgramDescription(
			"This is an example of a publist/subscribe client program"),
	)
}
