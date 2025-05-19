package strings

import (
	"errors"
	"fmt"

	"github.com/xypwn/filediver/extractor"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/physics"
)

func ExtractPhysicsNames(ctx extractor.Context) error {
	if !ctx.File().Exists(stingray.DataMain) {
		return errors.New("no main data")
	}
	r, err := ctx.File().Open(ctx.Ctx(), stingray.DataMain)
	if err != nil {
		return err
	}
	defer r.Close()
	physics, err := physics.LoadPhysics(r)
	if err != nil {
		return err
	}
	physicsSuffix := string(physics.NameEnd[:23])
	knownName, ok := ctx.Hashes()[ctx.File().ID().Name]
	if ok {
		fmt.Println(knownName)
	} else {
		fmt.Println(physicsSuffix)
	}
	return err
}
