package game

import (
  "path/filepath"
  "io/ioutil"
  "github.com/runningwild/glop/gui"
  "github.com/runningwild/haunts/base"
  "github.com/runningwild/haunts/house"
  lua "github.com/xenith-studios/golua"
)

type gameScript struct {
  L *lua.State

  // Since the scripts can do anything they want sometimes we want make sure
  // certain things only run when the game is ready for them.
  sync chan struct{}
}

// syncStart and syncEnd are identical, just wanted different function names
// so things made more sense
func (gs *gameScript) syncStart() {
  gs.sync <- struct{}{}
}
func (gs *gameScript) syncEnd() {
  gs.sync <- struct{}{}
}

func startGameScript(gp *GamePanel, path string) {
  // Clear out the panel, now the script can do whatever it wants
  gp.AnchorBox = gui.MakeAnchorBox(gui.Dims{1024,700})
  base.Log().Printf("startGameScript")
  if !filepath.IsAbs(path) {
    path = filepath.Join(base.GetDataDir(), "scripts", path)
  }

  // The game script runs in a separate go routine and functions that need to
  // communicate with the game will do so via channels - DUH why did i even
  // write this comment?
  prog, err := ioutil.ReadFile(path)
  if err != nil {
    base.Error().Printf("Unable to load game script file %s: %v", path, err)
    return
  }
  gp.script = &gameScript{}
  gp.script.L = lua.NewState()
  gp.script.L.OpenLibs()
  gp.script.L.SetExecutionLimit(25000)
  registerUtilityFunctions(gp.script.L)
  gp.script.L.Register("loadHouse", loadHouse(gp))
  gp.script.L.Register("showMainBar", showMainBar(gp))
  gp.script.L.Register("spawnDude", spawnDude(gp))
  gp.script.L.Register("placeDude", placeDude(gp))
  gp.script.L.Register("getAllEnts", getAllEnts(gp))
  gp.script.L.Register("selectMap", selectMap(gp))

  gp.script.sync = make(chan struct{})
  res := gp.script.L.DoString(string(prog))
  if !res {
    base.Error().Printf("There was an error running script %s:\n%s", path, prog)
  } else {
    go func() {
      gp.script.L.SetExecutionLimit(250000)
      gp.script.L.GetField(lua.LUA_GLOBALSINDEX, "Init")
      gp.script.L.Call(0, 0)
    } ()
  }
}

func (gs *gameScript) OnRound(g *Game) {
  gs.L.SetExecutionLimit(250000)
  base.Log().Printf("Calling on round")
  gs.L.GetField(lua.LUA_GLOBALSINDEX, "OnRound")
  gs.L.PushBoolean(g.Side == SideExplorers)
  gs.L.Call(1, 0)
}

func (gp *GamePanel) scriptThink() {
  if gp.script.L == nil {
    return
  }
  done := false
  for !done {
    select {
    // If a script has tried to run a function that requires running during
    // Think then it can run now and we'll wait for it to finish before
    // continuing.
    case <-gp.script.sync:
      <-gp.script.sync
    default:
      done = true
    }
  }
}

func loadHouse(gp *GamePanel) lua.GoFunction {
  return func(L* lua.State) int {
    gp.script.syncStart()
    defer gp.script.syncEnd()

    name := L.ToString(-1)
    def := house.MakeHouseFromName(name)
    if def == nil || len(def.Floors) == 0 {
      base.Error().Printf("No house exists with the name '%s'.", name)
      return 0
    }
    gp.house = def
    gp.viewer = house.MakeHouseViewer(gp.house, 62)
    gp.viewer.Edit_mode = true
    gp.game = makeGame(gp.house, gp.viewer, SideExplorers)
    gp.game.script = gp.script

    gp.AnchorBox = gui.MakeAnchorBox(gui.Dims{1024,700})

    gp.AnchorBox.AddChild(gp.viewer, gui.Anchor{0.5,0.5,0.5,0.5})
    base.Log().Printf("Done making stuff")
    return 0
  }
}

func showMainBar(gp *GamePanel) lua.GoFunction {
  return func(L *lua.State) int {
    gp.script.syncStart()
    defer gp.script.syncEnd()
    show := L.ToBoolean(-1)

    // Remove it regardless of whether or not we want to hide it
    for _, child := range gp.AnchorBox.GetChildren() {
      if child == gp.main_bar {
        gp.AnchorBox.RemoveChild(child)
        break
      }
    }

    if show {
      var err error
      gp.main_bar,err = MakeMainBar(gp.game)
      if err != nil {
        base.Error().Printf("%v", err)
        return 0
      }
      gp.AnchorBox.AddChild(gp.main_bar, gui.Anchor{0.5,0,0.5,0})
    }
    base.Log().Printf("Num kids: %d", len(gp.AnchorBox.GetChildren()))
    return 0
  }
}

func spawnDude(gp *GamePanel) lua.GoFunction {
  return func(L *lua.State) int {
    gp.script.syncStart()
    defer gp.script.syncEnd()
    name := L.ToString(-3)
    x := L.ToInteger(-2)
    y := L.ToInteger(-1)
    gp.game.SpawnEntity(MakeEntity(name, gp.game), x, y)
    return 0
  }
}

func placeDude(gp *GamePanel) lua.GoFunction {
  return func(L *lua.State) int {
    gp.script.syncStart()
    pattern := L.ToString(-3)
    points := L.ToInteger(-2)

    var names []string
    var costs []int
    L.PushNil()
    for L.Next(-2) != 0 {
      L.PushInteger(1)
      L.GetTable(-2)
      names = append(names, L.ToString(-1))
      L.Pop(1)
      L.PushInteger(2)
      L.GetTable(-2)
      costs = append(costs, L.ToInteger(-1))
      L.Pop(1)
      L.Pop(1)
    }

    ep, placed_chan := MakeEntityPlacer(gp.game, pattern, points, names, costs)
    gp.AnchorBox.AddChild(ep, gui.Anchor{0.5,0.5,0.5,0.5})
    gp.script.syncEnd()

    placed := <-placed_chan
    L.NewTable()
    for i := range placed {
      L.PushInteger(i + 1)
      L.PushString(placed[i])
      L.SetTable(-3)
    }

    gp.script.syncStart()
    gp.AnchorBox.RemoveChild(ep)
    gp.script.syncEnd()
    return 1
  }
}

func getAllEnts(gp *GamePanel) lua.GoFunction {
  return func(L *lua.State) int {
    gp.script.syncStart()
    defer gp.script.syncEnd()
    L.NewTable()
    for i := range gp.game.Ents {
      L.PushInteger(i+1)
      L.PushInteger(int(gp.game.Ents[i].Id))
      L.SetTable(-3)
    }
    return 1
  }
}

func selectMap(gp *GamePanel) lua.GoFunction {
  return func(L *lua.State) int {
    gp.script.syncStart()
    selector, output, err := MakeUiSelectMap(gp)
    if err != nil {
      base.Error().Printf("Error selecting map: %v", err)
      return 0
    }
    gp.AnchorBox.AddChild(selector, gui.Anchor{0.5,0.5,0.5,0.5})
    gp.script.syncEnd()

    name := <-output

    gp.script.syncStart()
    gp.AnchorBox.RemoveChild(selector)
    L.PushString(name)
    gp.script.syncEnd()
    return 1
  }
}

// Ripped from game/ai/ai.go - should probably sync up with it
func registerUtilityFunctions(L *lua.State) {
  L.Register("print", func(L *lua.State) int {
    var res string
    n := L.GetTop()
    for i := -n; i < 0; i++ {
      res += luaStringifyParam(L, i) + " "
    }
    base.Log().Printf("GameScript(%p): %s", L, res)
    return 0
  })
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
