package conditions_test

import (
  . "gospec"
  "gospec"
  "path/filepath"
  "encoding/gob"
  "bytes"
  "haunts/base"
  "haunts/game/status"
)

var datadir string

func init() {
  datadir,_ = filepath.Abs("../../data_test")
  base.SetDatadir(datadir)
}

func ConditionsSpec(c gospec.Context) {
  status.RegisterAllConditions()
  c.Specify("Conditions are loaded properly.", func() {
    basic := status.MakeCondition("Basic Test")
    _,ok := basic.(*status.BasicCondition)
    c.Expect(ok, Equals, true)
  })

  c.Specify("Conditions can be gobbed without loss of type.", func() {
    buf := bytes.NewBuffer(nil)
    enc := gob.NewEncoder(buf)

    var cs []status.Condition
    cs = append(cs, status.MakeCondition("Basic Test"))

    err := enc.Encode(cs)
    c.Assume(err, Equals, nil)

    dec := gob.NewDecoder(buf)
    var cs2 []status.Condition
    err = dec.Decode(&cs2)
    c.Assume(err, Equals, nil)

    _,ok := cs2[0].(*status.BasicCondition)
    c.Expect(ok, Equals, true)
  })
}
