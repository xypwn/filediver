package state_machine

import (
	"encoding/json"
	"errors"

	"github.com/xypwn/filediver/extractor"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/state_machine"
)

func ExtractStateMachineJson(ctx extractor.Context) error {
	if !ctx.File().Exists(stingray.DataMain) {
		return errors.New("no main data")
	}
	r, err := ctx.File().Open(ctx.Ctx(), stingray.DataMain)
	if err != nil {
		return err
	}
	defer r.Close()

	stateMachine, err := state_machine.LoadStateMachine(r)
	if err != nil {
		return err
	}

	text, err := json.Marshal(stateMachine)
	if err != nil {
		return err
	}

	out, err := ctx.CreateFile(".state_machine.json")
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = out.Write(text)
	return err
}
