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
  c.Specify("Conditions are loaded properly.", func() {
    basic := status.MakeCondition("Basic Test")
    _,ok := basic.(*status.BasicCondition)
    c.Expect(ok, Equals, true)
    c.Expect(basic.Strength(), Equals, 5)
    c.Expect(basic.Kind(), Equals, status.Fire)
    var b status.Base
    b = basic.ModifyBase(b, status.Unspecified)
    c.Expect(b.Attack, Equals, 3)
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

  c.Specify("Conditions stack properly", func() {
    var s status.Inst
    fd := status.MakeCondition("Fire Debuff Attack")
    pd := status.MakeCondition("Poison Debuff Attack")
    pd2 := status.MakeCondition("Poison Debuff Attack 2")
    c.Expect(s.AttackBonusWith(status.Unspecified), Equals, 0)
    s.ApplyCondition(pd)
    c.Expect(s.AttackBonusWith(status.Unspecified), Equals, -1)
    s.ApplyCondition(fd)
    c.Expect(s.AttackBonusWith(status.Unspecified), Equals, -2)
    s.ApplyCondition(fd)
    c.Expect(s.AttackBonusWith(status.Unspecified), Equals, -2)
    s.ApplyCondition(pd)
    c.Expect(s.AttackBonusWith(status.Unspecified), Equals, -2)
    s.ApplyCondition(pd2)
    c.Expect(s.AttackBonusWith(status.Unspecified), Equals, -3)
  })
}
