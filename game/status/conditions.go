package status

import (
  "encoding/gob"
  "path/filepath"
  "github.com/runningwild/haunts/base"
)

// Conditions represent instantaneous or ongoing Conditions on an entity.
// Every round the Condition can 
type Condition interface {
  // Returns whether or not this condition is a buff
  Buff() bool

  // Returns this condition's Kind
  Kind() Kind

  // Returns the strength of this condition relative to other conditions of
  // the same Kind.  This is used to determine which conditions will displace
  // others.
  Strength() int

  ModifyDamage(Damage) Damage

  // Called any time a Base stat is queried
  ModifyBase(Base, Kind) Base

  // Called at the beginning of each round.  May return a damage object to
  // deal damage, and must return a bool indicating whether this effect has
  // completed or not.
  OnRound() (dmg *Damage, complete bool)
}

var condition_registerers []func()
var condition_makers map[string]func() Condition
func RegisterAllConditions() {
  condition_makers = make(map[string]func() Condition)
  for _,f := range condition_registerers {
    f()
  }
}

func MakeCondition(name string) Condition {
  return condition_makers[name]()
}

func registerBasicConditions() {
  registry_name := "conditions-basic_conditions"
  base.RemoveRegistry(registry_name)
  base.RegisterRegistry(registry_name, make(map[string]*basicConditionDef))
  base.RegisterAllObjectsInDir(registry_name, filepath.Join(base.GetDataDir(), "conditions", "basic_conditions"), ".json", "json")
  names := base.GetAllNamesInRegistry(registry_name)
  for _,name := range names {
    cname := name
    f := func() Condition {
      c := BasicCondition{ Defname: cname }
      base.GetObject(registry_name, &c)
      return &c
    }
    condition_makers[name] = f
  }
}

func init() {
  condition_registerers = append(condition_registerers, registerBasicConditions)
  gob.Register(&BasicCondition{})
}

type BasicCondition struct {
  Defname string
  *basicConditionDef
  time int
}

type basicConditionDef struct {
  Name string

  // On OnRound() this Condition will create a Damage object with this Dynamic
  // object.
  Dynamic

  // This Condition will modify its target unit by adding every value in Base
  // to the unit's Base stats
  Base

  // Use Type here instead of Kind so it doesn't overlap with the required
  // method name Kind.  Also Type will be used in the json files so it should
  // be no less obvious what it is.
  Kind Kind

  Buff bool

  // The strength of this condition
  Strength int

  // This Condition will OnRound() exactly Time + 1 times.  If Time < 0 then
  // it will OnRound() forever.
  Time int
}

func (bc *BasicCondition) Buff() bool {
  return bc.basicConditionDef.Buff
}

func (bc *BasicCondition) Strength() int {
  return bc.basicConditionDef.Strength
}

func (bc *BasicCondition) Kind() Kind {
  return bc.basicConditionDef.Kind
}

func (bc *basicConditionDef) ModifyDamage(dmg Damage) Damage {
  return dmg
}

func (bc *basicConditionDef) ModifyBase(base Base, kind Kind) Base {
  base.Ap_max += bc.Ap_max
  base.Hp_max += bc.Hp_max
  base.Sight += bc.Sight
  base.Attack += bc.Attack
  base.Corpus += bc.Corpus
  base.Ego += bc.Ego
  return base
}

func (bc *BasicCondition) OnRound() (dmg *Damage, complete bool) {
  var d Dynamic
  if bc.Dynamic != d {
    dmg = &Damage{ Dynamic: bc.Dynamic, Kind: bc.Kind() }
  }
  bc.time++
  complete = (bc.time == bc.Time)
  return
}
