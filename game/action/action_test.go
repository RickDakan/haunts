package action_test

import (
  . "gospec"
  "gospec"
  "path/filepath"
  "encoding/gob"
  "bytes"
  "haunts/base"
  "haunts/game/action"
)

var datadir string

func init() {
  datadir,_ = filepath.Abs("../../data_test")
  base.SetDatadir(datadir)
  action.LoadAllActionsInDir(filepath.Join(datadir, "actions"))
}

func ActionSpec(c gospec.Context) {
  c.Specify("Actions are loaded properly.", func() {
    basic := action.MakeAction("Basic Test")
    c.Expect(basic.Cost(), Equals, 3)
    charge := action.MakeAction("Charge Test")
    c.Expect(charge.Cost(), Equals, 4)
  })

  c.Specify("Actions can be gobbed without loss of type.", func() {
    buf := bytes.NewBuffer(nil)
    enc := gob.NewEncoder(buf)

    var basic action.Action
    basic = action.MakeAction("Basic Test")
    c.Expect(basic.Cost(), Equals, 3)

    err := enc.Encode(basic)
    c.Expect(err, Equals, nil)

    dec := gob.NewDecoder(buf)
    var basic2 action.ActionBasicAttack
    err = dec.Decode(&basic2)
    c.Expect(err, Equals, nil)
    c.Expect(basic2.Cost(), Equals, 3)
  })
}
