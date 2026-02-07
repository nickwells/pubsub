package main

import (
	"github.com/nickwells/param.mod/v7/param"
	"github.com/nickwells/param.mod/v7/paramset"
	"github.com/nickwells/pusuparams.mod/pusuparams"
	"github.com/nickwells/verbose.mod/verbose"
	"github.com/nickwells/versionparams.mod/versionparams"
)

// makeParamSet generates the param set ready for parsing
func makeParamSet(prog *prog) *param.PSet {
	return paramset.New(
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
			"This is an example of a publish/subscribe client program"),
	)
}
