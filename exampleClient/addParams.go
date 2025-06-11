package main

import (
	"time"

	"github.com/nickwells/check.mod/v2/check"
	"github.com/nickwells/param.mod/v6/param"
	"github.com/nickwells/param.mod/v6/psetter"
	"github.com/nickwells/pusu.mod/pusu"
	"github.com/nickwells/pusuparams.mod/pusuparams"
	"github.com/nickwells/slogsetter.mod/slogsetter"
)

const (
	paramNameLogLevel  = "log-level"
	paramNameNamespace = "namespace"
	paramNameTopic     = "topic"
	paramNamePayload   = "payload"
	paramNameTimeout   = "timeout"
	paramNameCount     = "count"
)

// addParams adds the parameters for this program
func addParams(prog *prog) param.PSetOptFunc {
	return func(ps *param.PSet) error {
		ps.Add(paramNameLogLevel,
			slogsetter.Level{
				Value: &prog.logLevel,
			},
			"the level of logging")

		ps.Add(paramNameNamespace,
			psetter.String[pusu.Namespace]{
				Value: &prog.namespace,
				Checks: []check.ValCk[pusu.Namespace]{
					check.StringLength[pusu.Namespace](check.ValGT(0)),
				},
			},
			"the namespace for the topics",
			param.Attrs(param.MustBeSet))

		ps.Add(paramNameTopic,
			pusuparams.TopicSetter{
				Value: &prog.topic,
			},
			"the topic to use",
			param.Attrs(param.MustBeSet))

		prog.payloadParam = ps.Add(paramNamePayload,
			psetter.String[string]{
				Value: &prog.payload,
			},
			"the payload to send. Note that if this parameter is set"+
				" then the program will send this payload on the given"+
				" topic and then exit. Otherwise it will listen on the"+
				" topic for a period of time and report each message"+
				" received",
			param.AltNames("message"),
		)

		prog.timeoutParam = ps.Add(paramNameTimeout,
			psetter.Duration{
				Value: &prog.timeout,
				Checks: []check.Duration{
					check.ValGT(time.Duration(0)),
				},
			},
			"how long to wait before exiting")

		prog.countParam = ps.Add(paramNameCount,
			psetter.Int[int]{
				Value: &prog.count,
				Checks: []check.ValCk[int]{
					check.ValGT(1),
				},
			},
			"how many messages should be Published or received before"+
				" exiting")

		ps.AddFinalCheck(func() error {
			prog.progName = ps.ProgName()
			return nil
		})

		return nil
	}
}
