package actions_test

import (
  . "gospec"
  "gospec"
  "path/filepath"
  "encoding/gob"
  "bytes"
  "haunts/base"
  "haunts/game"
  "haunts/game/actions"
)

var datadir string

func init() {
  datadir,_ = filepath.Abs("../../data_test")
  base.SetDatadir(datadir)
}

func ActionSpec(c gospec.Context) {
  game.RegisterActions()
  // c.Specify("Actions are loaded properly.", func() {
  //   basic := actions.MakeAction("Basic Test")
  //   c.Expect(basic.Cost(), Equals, 3)
  //   charge := actions.MakeAction("Charge Test")
  //   c.Expect(charge.Cost(), Equals, 4)
  // })

  c.Specify("Actions can be gobbed without loss of type.", func() {
    buf := bytes.NewBuffer(nil)
    enc := gob.NewEncoder(buf)

    var basic game.Action
    basic = game.MakeAction("Move")
    c.Expect(basic.Cost(), Equals, 3)

    err := enc.Encode(basic)
    c.Expect(err, Equals, nil)

    dec := gob.NewDecoder(buf)
    var basic2 actions.Move
    err = dec.Decode(&basic2)
    c.Expect(err, Equals, nil)
    c.Expect(basic2.Cost(), Equals, 3)
  })
}
