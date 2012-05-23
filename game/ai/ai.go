package ai

import (
  "strings"
  "math/rand"
  "reflect"
  "encoding/gob"
  "github.com/runningwild/glop/ai"
  "github.com/runningwild/haunts/base"
  "github.com/runningwild/haunts/game"
  "github.com/runningwild/haunts/game/actions"
  "github.com/runningwild/polish"
  "github.com/runningwild/yedparse"
)

// The Ai struct contains a glop.AiGraph object as well as a few channels for
// communicating with the game itself.
type Ai struct {
  graph   *ai.AiGraph

  game *game.Game

  ent *game.Entity

  State AiState


  // new stuff

  // Set to true at the beginning of the turn, turned off as soon as the ai is
  // done for the turn.
  active bool
  evaluating bool
  active_set chan bool
  active_query chan bool
  exec_query chan struct{}

  // Once we send an Action for execution we have to wait until it is done
  // before we make the next one.  This channel is used to handle that.
  pause chan struct{}
  execs chan game.ActionExec
}

type AiState struct {
  Last_offensive_target game.EntityId
}

func init() {
  gob.Register(&Ai{})
  game.SetAiMaker(makeAi)
}

func makeAi(path string, g *game.Game, ent *game.Entity, dst_iface *game.Ai, kind game.AiKind) {
  var ai_struct *Ai
  if *dst_iface == nil {
    ai_struct = new(Ai)
  } else {
    ai_struct = (*dst_iface).(*Ai)
  }
  ai_graph := ai.NewGraph()
  graph,err := yed.ParseFromFile(path)
  if err != nil {
    base.Error().Printf("%v", err)
    panic(err.Error())
  }
  ai_graph.Graph = &graph.Graph
  ai_graph.Context = polish.MakeContext()
  ai_struct.ent = ent
  ai_struct.graph = ai_graph
  ai_struct.game = g

  ai_struct.active_set = make(chan bool)
  ai_struct.active_query = make(chan bool)
  ai_struct.exec_query = make(chan struct{})
  ai_struct.pause = make(chan struct{})
  ai_struct.execs = make(chan game.ActionExec)

  switch kind {
  case game.EntityAi:
    ai_struct.addEntityContext(ai_struct.ent, ai_struct.graph.Context)

  case game.MinionsAi:
    ai_struct.addMinionsContext()

  case game.DenizensAi:
    ai_struct.addDenizensContext()

  case game.IntrudersAi:
  default:
    panic("Unknown ai kind")
  }
  go ai_struct.masterRoutine()
  *dst_iface = ai_struct
}

// Need a goroutine for each ai - all things will go through is so that things
// stay synchronized
func (a *Ai) masterRoutine() {
  for {
    select {
    case a.active = <-a.active_set:
      if a.active && a.ent == nil {
        // The master is responsible for activating all entities
        for i := range a.game.Ents {
          if a.game.Ents[i].Ai != nil {
            a.game.Ents[i].Ai.Activate()
          }
        }
      }
      if a.active == false {
        if a.ent == nil {
          base.Log().Printf("Evaluating = false")
        } else {
          base.Log().Printf("Ent %p inactivated", a.ent)
        }
        a.evaluating = false
      }

    case a.active_query <- a.active:

    case <-a.exec_query:
      if a.active {
        select {
          case a.pause <- struct{}{}:
          default:
        }
      }
      if a.active && !a.evaluating {
        a.evaluating = true
        go func() {
          if a.ent == nil {
            base.Log().Printf("Eval master")
          } else {
            base.Log().Printf("Eval ent: %p", a.ent)
          }
          labels, err := a.graph.Eval()
          if a.ent == nil {
            base.Log().Printf("Completed master")
          } else {
            base.Log().Printf("Completed ent: %p", a.ent)
          }
          for i := range labels {
            base.Log().Printf("Execed: %s", labels[i])
          }
          if err != nil {
            base.Warn().Printf("%v", err)
            if e, ok := err.(*polish.Error); ok {
              base.Warn().Printf("%s", e.Stack)
            }
          }
          a.active_set <- false
          a.active_set <- false
          a.execs <- nil
          base.Log().Printf("Sent nil value")
        } ()
      }
    }
  }
}

func (a *Ai) Activate() {
  a.active_set <- true
}

func (a *Ai) Active() bool {
  return <-a.active_query
}

func (a *Ai) ActionExecs() <-chan game.ActionExec {
  select {
    case a.pause <- struct{}{}:
    default:
  }
  a.exec_query <- struct{}{}
  return a.execs
}

// Does the roll dice-d-sides, like 3d6, and returns the result
func roll(dice, sides float64) float64 {
  result := 0
  for i := 0; i < int(dice); i++ {
    result += rand.Intn(int(sides)) + 1
  }
  return float64(result)
}

func (a *Ai) addMinionsContext() {
  polish.AddFloat64MathContext(a.graph.Context)
  polish.AddBooleanContext(a.graph.Context)
  a.graph.Context.SetParseOrder(polish.Float, polish.String)
  a.graph.Context.AddFunc("numActiveMinions", func() float64 {
    count := 0.0
    for _, e := range a.game.Ents {
      if e.Side() != game.SideHaunt { continue }
      if e.HauntEnt.Level != game.LevelMinion { continue }
      if !e.Ai.Active() { continue }
      count++
    }
    base.Log().Printf("Num active minions: %f", count)
    return count
  })
  a.graph.Context.AddFunc("randomActiveMinion", func() *game.Entity {
    base.Log().Printf("randomActiveMinion")
    var ent *game.Entity
    count := 1.0
    for _, e := range a.game.Ents {
      if e.Side() != game.SideHaunt { continue }
      if e.HauntEnt.Level != game.LevelMinion { continue }
      if !e.Ai.Active() { continue }
      if rand.Float64() < 1.0 / count {
        ent = e
      }
      count++
    }
    base.Log().Printf("Selected %s (%p)", ent.Name, ent)
    return ent
  })
  a.graph.Context.AddFunc("exec", func(ent *game.Entity) {
    base.Log().Printf("Execute %p", ent)
    exec := <-ent.Ai.ActionExecs()
    base.Log().Printf("Got an action: %v", exec)
    if exec != nil {
      base.Log().Printf("Sending that action")
      a.execs <- exec
      base.Log().Printf("Sent.")
    }
    <-a.pause
  })
  a.graph.Context.AddFunc("done", func() {
    base.Log().Printf("master done")
    <-a.pause
  })
}

func (a *Ai) addDenizensContext() {
  polish.AddFloat64MathContext(a.graph.Context)
  polish.AddBooleanContext(a.graph.Context)
  a.graph.Context.SetParseOrder(polish.Float, polish.String)
  a.graph.Context.AddFunc("numActiveServitors", func() float64 {
    count := 0.0
    for _, e := range a.game.Ents {
      if e.Side() != game.SideHaunt { continue }
      if e.HauntEnt.Level != game.LevelServitor { continue }
      if !e.Ai.Active() { continue }
      count++
    }
    return count
  })
  a.graph.Context.AddFunc("randomActiveServitor", func() *game.Entity {
    var ent *game.Entity
    count := 1.0
    for _, e := range a.game.Ents {
      if e.Side() != game.SideHaunt { continue }
      if e.HauntEnt.Level != game.LevelServitor { continue }
      if !e.Ai.Active() { continue }
      if rand.Float64() < 1.0 / count {
        ent = e
      }
      count++
    }
    base.Log().Printf("Selected %s (%p)", ent.Name, ent)
    return ent
  })
  a.graph.Context.AddFunc("exec", func(ent *game.Entity) {
    base.Log().Printf("Execute %p", ent)
    exec := <-ent.Ai.ActionExecs()
    base.Log().Printf("Got an action: %v", exec)
    if exec != nil {
      base.Log().Printf("Sending that action")
      a.execs <- exec
      base.Log().Printf("Sent.")
    }
    <-a.pause
  })
  a.graph.Context.AddFunc("done", func() {
    base.Log().Printf("master done")
    // a.graph.Term() <- nil
    <-a.pause
  })
}

func (a *Ai) addEntityContext(ent *game.Entity, context *polish.Context) {
  polish.AddFloat64MathContext(context)
  polish.AddBooleanContext(context)
  context.SetParseOrder(polish.Float, polish.String)

  // This entity, the one currently taking its turn
  context.SetValue("me", ent)

  // All actions that the entity has are available using their names,
  // converted to lower case, and replacing spaces with underscores.
  // For example, "Kiss of Death" -> "kiss_of_death"

  // rolls dice, for example "roll 3 6" is a roll of 3d6
  context.AddFunc("roll", roll)

  // These functions are self-explanitory, they are all relative to the
  // current entity
  context.AddFunc("numVisibleEnemies",
      func() float64 {
        base.Log().Printf("numVisibleEnemies")
        return float64(numVisibleEntities(ent, false))
      })
  context.AddFunc("nearestEnemy",
      func() *game.Entity {
        return nearestEntity(ent, false)
      })
  context.AddFunc("distBetween", distBetween)

  // Ends an entity's turn
  context.AddFunc("done",
      func() {
        base.Log().Printf("done")
        a.active_set <- false
        base.Log().Printf("done")
        // a.graph.Term() <- nil
      })

  // Checks whether an entity is nil, this is important to check when using
  // function that returns an entity (like lastOffensiveTarget)
  context.AddFunc("stillExists", func(target *game.Entity) bool {
    base.Log().Printf("stillExists")
    return target != nil
  })

  // Returns the last entity that this ai attacked.  If the entity has died
  // this can return nil, so be sure to check that before using it.
  context.AddFunc("lastOffensiveTarget", func() *game.Entity {
    base.Log().Printf("lastOffensiveTarget")
    return ent.Game().EntityById(a.State.Last_offensive_target)
  })

  // Advances as far as possible towards the target entity.
  context.AddFunc("advanceTowards", func(target *game.Entity) {
    base.Log().Printf("advanceTowards")
    name := getActionName(ent, reflect.TypeOf(&actions.Move{}))
    move := getActionByName(ent, name).(*actions.Move)
    x,y := target.Pos()
    exec := move.AiMoveToWithin(ent, x, y, 1)
    if exec != nil {
      base.Log().Printf("Sending exec: %v", exec)
      a.execs <- exec
      base.Log().Printf("Sent")
      <-a.pause
    } else {
      base.Log().Printf("Terminating")
      a.graph.Term() <- ai.TermError
      base.Log().Printf("Terminated")
    }
  })

  context.AddFunc("getBasicAttack", func() string {
    base.Log().Printf("getBasicAttack")
    return getActionName(ent, reflect.TypeOf(&actions.BasicAttack{}))
  })

  context.AddFunc("doBasicAttack", func(target *game.Entity, attack_name string) {
    base.Log().Printf("doBasicAttack")
    _attack := getActionByName(ent, attack_name)
    attack := _attack.(*actions.BasicAttack)
    exec := attack.AiAttackTarget(ent, target)
    if exec != nil {
      base.Log().Printf("Sending exec: %v", exec)
      a.execs <- exec
      a.State.Last_offensive_target = target.Id
      base.Log().Printf("Sent")
      <-a.pause
    } else {
      base.Log().Printf("Terminating")
      a.graph.Term() <- ai.TermError
      base.Log().Printf("Terminated")
    }
  })

  context.AddFunc("corpus", func(target *game.Entity) float64 {
    return float64(target.Stats.Corpus())
  })
  context.AddFunc("ego", func(target *game.Entity) float64 {
    return float64(target.Stats.Ego())
  })
  context.AddFunc("hpMax", func(target *game.Entity) float64 {
    return float64(target.Stats.HpMax())
  })
  context.AddFunc("apMax", func(target *game.Entity) float64 {
    return float64(target.Stats.ApMax())
  })
  context.AddFunc("hpCur", func(target *game.Entity) float64 {
    return float64(target.Stats.HpCur())
  })
  context.AddFunc("apCur", func(target *game.Entity) float64 {
    return float64(target.Stats.ApCur())
  })
  context.AddFunc("hasCondition", func(target *game.Entity, name string) bool {
    for _, con := range target.Stats.ConditionNames() {
      if lowerAndUnderscore(con) == name {
        return true
      }
    }
    return false
  })
}

func numVisibleEntities(e *game.Entity, ally bool) float64 {
  count := 0
  for _,ent := range e.Game().Ents {
    if ent == e { continue }
    if ent.Stats == nil || ent.Stats.HpCur() <= 0 { continue }
    if ally != (e.Side() == ent.Side()) { continue }
    x,y := ent.Pos()
    if e.HasLos(x, y, 1, 1) {
      count++
    }
  }
  return float64(count)
}

func distBetween(e1,e2 *game.Entity) float64 {
  e1x,e1y := e1.Pos()
  e2x,e2y := e2.Pos()
  dx := e1x - e2x
  dy := e1y - e2y
  if dx < 0 { dx = -dx }
  if dy < 0 { dy = -dy }
  if dx > dy {
    return float64(dx)
  }
  return float64(dy)
}

func nearestEntity(e *game.Entity, ally bool) *game.Entity {
  var nearest *game.Entity
  cur_dist := 1.0e9
  for _,ent := range e.Game().Ents {
    if ent == e { continue }
    if ent.Stats == nil || ent.Stats.HpCur() <= 0 { continue }
    if ally != (e.Side() == ent.Side()) { continue }
    dist := distBetween(e, ent)
    if cur_dist > dist {
      cur_dist = dist
      nearest = ent
    }
  }
  return nearest
}

func lowerAndUnderscore(s string) string {
  b := []byte(strings.ToLower(s))
  for i := range b {
    if b[i] == ' ' {
      b[i] = '_'
    }
  }
  return string(b)
}

func getActionByName(e *game.Entity, name string) game.Action {
  for _,action := range e.Actions {
    if lowerAndUnderscore(action.String()) == name {
      return action
    }
  }
  return nil
}

func getActionName(e *game.Entity, typ reflect.Type) string {
  for _,action := range e.Actions {
    if reflect.TypeOf(action) == typ {
      return lowerAndUnderscore(action.String())
    }
  }
  return ""
}
