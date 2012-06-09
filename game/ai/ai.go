package ai

import (
  "math/rand"
  "fmt"
  "encoding/gob"
  "io/ioutil"
  "github.com/runningwild/haunts/base"
  "github.com/runningwild/haunts/game"
  lua "github.com/xenith-studios/golua"
)

// The Ai struct contains a glop.AiGraph object as well as a few channels for
// communicating with the game itself.
type Ai struct {
  L *lua.State

  game *game.Game

  ent *game.Entity

  // The actual lua program to run when executing this ai
  Prog string

  // new stuff

  // Set to true at the beginning of the turn, turned off as soon as the ai is
  // done for the turn.
  active bool
  evaluating bool
  active_set chan bool
  active_query chan bool
  exec_query chan struct{}
  terminate chan struct{}

  // Once we send an Action for execution we have to wait until it is done
  // before we make the next one.  This channel is used to handle that.
  pause chan struct{}
  execs chan game.ActionExec

  // This exists so that we can gob this without error.  Gob doesn't like
  // gobbing things that don't have any exported fields, and since we might
  // want exported fields later we'll just have this here for now so we can
  // leave the Ai as an exported field in the entities.
  Dummy int
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
  prog, err := ioutil.ReadFile(path)
  if err != nil {
    base.Error().Printf("Unable to load ai file %s: %v", path, err)
    return
  }
  ai_struct.Prog = string(prog)
  ai_struct.ent = ent
  ai_struct.game = g
  ai_struct.L = lua.NewState();
  ai_struct.L.OpenLibs();

  ai_struct.active_set = make(chan bool)
  ai_struct.active_query = make(chan bool)
  ai_struct.exec_query = make(chan struct{})
  ai_struct.pause = make(chan struct{})
  ai_struct.terminate = make(chan struct{})
  ai_struct.execs = make(chan game.ActionExec)

  switch kind {
  case game.EntityAi:
    base.Log().Printf("Adding entity context for %s", ent.Name)
    ai_struct.addEntityContext()

  case game.MinionsAi:
    ai_struct.addMinionsContext()

  case game.DenizensAi:
    base.Log().Printf("Adding denizens context")
    ai_struct.addDenizensContext()

  case game.IntrudersAi:
    // ai_struct.addIntrudersContext(g)

  default:
    panic("Unknown ai kind")
  }
  // Add this to all contexts
  ai_struct.L.Register("print", func(L *lua.State) int {
    var res string
    n := L.GetTop()
    for i := -n; i < 0; i++ {
      res += luaStringifyParam(L, i) + " "
    }
    base.Log().Printf("Ai(%p): %s", ai_struct, res)
    return 0
  })
  ai_struct.L.Register("randN", func(L *lua.State) int {
    n := L.GetTop()
    if n == 0 || !L.IsNumber(-1) {
      L.PushInteger(0)
      return 1
    }
    val := L.ToInteger(-1)
    L.PushInteger(rand.Intn(val) + 1)
    return 1
  })

  go ai_struct.masterRoutine()
  *dst_iface = ai_struct
}

func luaStringifyParam(L *lua.State, index int) string {
  if L.IsTable(index) {
    return "table"
  }
  if L.IsBoolean(index) {
    if L.ToBoolean(index) {
      return "true"
    }
    return "false"
  }
  return L.ToString(index)
}

// Need a goroutine for each ai - all things will go through is so that things
// stay synchronized
func (a *Ai) masterRoutine() {
  for {
    select {
    case <-a.terminate:
      base.Log().Printf("Terminated Ai(p=%p)", a)
      return

    case a.active = <-a.active_set:
      if a.active {
        <-a.active_set
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
          base.Log().Printf("Evaluating lua script: %s", a.Prog)
          // Reset the execution limit in case it was set to 0 due to a
          // previous error
          a.L.SetExecutionLimit(25000)
          res := a.L.DoString(a.Prog)
          base.Log().Printf("Res: %t", res)
          if a.ent == nil {
            base.Log().Printf("Completed master")
          } else {
            base.Log().Printf("Completed ent: %p", a.ent)
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

func (a *Ai) Terminate() {
  a.terminate <- struct{}{}
}

func (a *Ai) Activate() {
  a.active_set <- true
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

type luaType int
const(
  luaInteger luaType = iota
  luaString
  luaTable
)

func makeLuaSigniature(name string, params []luaType) string {
  sig := name + "("
  for i := range params {
    switch params[i] {
    case luaInteger:
      sig += "integer"
    case luaString:
      sig += "string"
    case luaTable:
      sig += "table"
    default:
      sig += "<unknown type>"
    }
    if i != len(params) - 1 {
      sig += ", "
    }
  }
  sig += ")"
  return sig
}

func luaCheckParamsOk(L *lua.State, name string, params ...luaType) bool {
  fmt.Sprintf("%s(")
  n := L.GetTop()
  if n != len(params) {
    luaDoError(L, fmt.Sprintf("Got %d parameters to %s.", n, makeLuaSigniature(name, params)))
    return false
  }
  for i := -n; i < 0; i++ {
    ok := false
    switch params[i + n] {
    case luaInteger:
      ok = L.IsNumber(i)
    case luaString:
      ok = L.IsString(i)
    case luaTable:
      ok = L.IsTable(i)
    }
    if !ok {
      luaDoError(L, fmt.Sprintf("Unexpected parameters to %s.", makeLuaSigniature(name, params)))
      return false
    }
  }
  return true
}

func luaDoError(L *lua.State, err_str string) {
  base.Error().Printf(err_str)
  L.PushString(err_str)
  L.SetExecutionLimit(1)
}

func luaNumParamsOk(L *lua.State, num_params int, name string) bool {
  n := L.GetTop()
  if n != num_params {
    err_str := fmt.Sprintf("%s expects exactly %d parameters, got %d.", name, num_params, n)
    luaDoError(L, err_str)
    return false
  }
  return true
}

func rangedDistBetween(e1,e2 *game.Entity) int {
  e1x,e1y := e1.Pos()
  e2x,e2y := e2.Pos()
  dx := e1x - e2x
  dy := e1y - e2y
  if dx < 0 { dx = -dx }
  if dy < 0 { dy = -dy }
  if dx > dy {
    return dx
  }
  return dy
}

