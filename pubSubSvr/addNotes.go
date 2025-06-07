package main

import (
	"github.com/nickwells/param.mod/v6/param"
)

const (
	noteBaseName = "pubSubSvr - "

	noteNameSecurity = noteBaseName + "security"
)

// addNotes adds the notes for this program.
func addNotes(_ *prog) param.PSetOptFunc {
	return func(ps *param.PSet) error {
		ps.AddNote(noteNameSecurity,
			"security is provided by the use of mutual TLS")

		return nil
	}
}
