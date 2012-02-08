package status

import (
  "fmt"
  "haunts/base"
  "path/filepath"
  "encoding/gob"
)

// Conditions represent instantaneous or ongoing Conditions on an entity.
// Every round the Condition can 
type Condition interface {
  // Called any time a Base stat is queried
  ModifyBase(Base, Kind) Base

  // Called at the beginning of each round.  May return a damage object to
  // deal damage, and must return a bool indicating whether this effect has
  // completed or not.
  Think() (dmg *Damage, complete bool)
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

func registerConditionMaker(name string, maker func() Condition) {
  if _,ok := condition_makers[name]; ok {
    panic(fmt.Sprintf("Cannot register the condition maker '%s' more than once.", name))
  }
  condition_makers[name] = maker
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

  // On Think() this Condition will create a Damage object with this Dynamic
  // object.
  Dynamic

  // This Condition will modify its target unit by adding every value in Base
  // to the unit's Base stats
  Base

  Kind Kind

  // This Condition will Think() exactly Time + 1 times.  If Time < 0 then
  // it will Think() forever.
  Time int
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

func (bc *BasicCondition) Think() (dmg *Damage, complete bool) {
  var d Dynamic
  if bc.Dynamic != d {
    dmg = &Damage{ Dynamic: bc.Dynamic, Kind: bc.Kind }
  }
  complete = bc.time == 0
  bc.time--
  return
}
