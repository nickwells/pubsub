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

		addParams(prog),
		pusuparams.AddCertInfoParams(&prog.certInfo, ""),

		addNotes(prog),

		param.SetProgramDescription(
			"This is a publish/subscribe server."+
				" Clients connect to it and subscribe to topics and"+
				" publish messages on those topics which are then"+
				" forwarded by the server to all the clients who"+
				" have subscribed"),
	)
}
