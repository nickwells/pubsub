package main

import (
	"github.com/nickwells/param.mod/v6/param"
)

const (
	noteBaseName = "exampleClient - "

	noteNamePurpose = noteBaseName + "purpose"
)

// addNotes adds the notes for this program.
func addNotes(_ *prog) param.PSetOptFunc {
	return func(ps *param.PSet) error {
		ps.AddNote(noteNamePurpose,
			"this serves as an example of how to write"+
				" a client program that uses the publish/subscribe server."+
				" It demonstrates both Publish and Subscribe functionality.")

		return nil
	}
}
