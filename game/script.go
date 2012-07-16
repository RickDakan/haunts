package game

import (
  "fmt"
  "bytes"
  "math/rand"
  "path/filepath"
  "io/ioutil"
  "regexp"
  "github.com/runningwild/glop/gui"
  "github.com/runningwild/haunts/base"
  "github.com/runningwild/haunts/texture"
  "github.com/runningwild/haunts/house"
  "github.com/runningwild/haunts/game/status"
  "github.com/runningwild/haunts/game/hui"
  lua "github.com/xenith-studios/golua"
)

type gameScript struct {
  L *lua.State

  // Since the scripts can do anything they want sometimes we want make sure
  // certain things only run when the game is ready for them.
  sync chan struct{}
}

func (gs *gameScript) syncStart() {
  <-gs.sync
}
func (gs *gameScript) syncEnd() {
  gs.sync <- struct{}{}
}

func startGameScript(gp *GamePanel, path string, player *Player) {
  // Clear out the panel, now the script can do whatever it wants
  gp.AnchorBox = gui.MakeAnchorBox(gui.Dims{1024, 700})
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
  gp.script.L.NewTable()
  LuaPushSmartFunctionTable(gp.script.L, FunctionTable{
    "StartScript":                       func() { gp.script.L.PushGoFunctionAsCFunction(startScript(gp, player)) },
    "SelectHouse":                       func() { gp.script.L.PushGoFunctionAsCFunction(selectHouse(gp)) },
    "LoadHouse":                         func() { gp.script.L.PushGoFunctionAsCFunction(loadHouse(gp)) },
    "ShowMainBar":                       func() { gp.script.L.PushGoFunctionAsCFunction(showMainBar(gp)) },
    "SpawnEntityAtPosition":             func() { gp.script.L.PushGoFunctionAsCFunction(spawnEntityAtPosition(gp)) },
    "GetSpawnPointsMatching":            func() { gp.script.L.PushGoFunctionAsCFunction(getSpawnPointsMatching(gp)) },
    "SpawnEntitySomewhereInSpawnPoints": func() { gp.script.L.PushGoFunctionAsCFunction(spawnEntitySomewhereInSpawnPoints(gp)) },
    "PlaceEntities":                     func() { gp.script.L.PushGoFunctionAsCFunction(placeEntities(gp)) },
    "RoomAtPos":                         func() { gp.script.L.PushGoFunctionAsCFunction(roomAtPos(gp)) },
    "SetLosMode":                        func() { gp.script.L.PushGoFunctionAsCFunction(setLosMode(gp)) },
    "GetAllEnts":                        func() { gp.script.L.PushGoFunctionAsCFunction(getAllEnts(gp)) },
    "DialogBox":                         func() { gp.script.L.PushGoFunctionAsCFunction(dialogBox(gp)) },
    "PickFromN":                         func() { gp.script.L.PushGoFunctionAsCFunction(pickFromN(gp)) },
    "SetGear":                           func() { gp.script.L.PushGoFunctionAsCFunction(setGear(gp)) },
    "BindAi":                            func() { gp.script.L.PushGoFunctionAsCFunction(bindAi(gp)) },
    "SetVisibility":                     func() { gp.script.L.PushGoFunctionAsCFunction(setVisibility(gp)) },
    "EndPlayerInteraction":              func() { gp.script.L.PushGoFunctionAsCFunction(endPlayerInteraction(gp)) },
    "SaveStore":                         func() { gp.script.L.PushGoFunctionAsCFunction(saveStore(gp, player)) },
    "SetCondition":                      func() { gp.script.L.PushGoFunctionAsCFunction(setCondition(gp)) },
    "SetPosition":                       func() { gp.script.L.PushGoFunctionAsCFunction(setPosition(gp)) },
  })
  gp.script.L.SetMetaTable(-2)
  gp.script.L.SetGlobal("Script")

  registerUtilityFunctions(gp.script.L)
  if player.Lua_store != nil {
    LuaDecodeTable(bytes.NewBuffer(player.Lua_store), gp.script.L)
    gp.script.L.SetGlobal("store")
  } else {
    gp.script.L.NewTable()
    gp.script.L.SetGlobal("store")
  }
  gp.script.sync = make(chan struct{})
  res := gp.script.L.DoString(string(prog))
  if !res {
    base.Error().Printf("There was an error running script %s:\n%s", path, prog)
  } else {
    go func() {
      gp.script.L.SetExecutionLimit(250000)
      gp.script.L.DoString("Init()")
      if gp.game == nil {
        base.Error().Printf("Script failed to load a house during Init().")
      } else {
        gp.game.comm.script_to_game <- nil
      }
    }()
  }
}

// Runs RoundStart
// Lets the game know that the round middle can begin
// Runs RoundEnd
func (gs *gameScript) OnRound(g *Game) {
  base.Log().Printf("Launching script.RoundStart")
  go func() {
    // // round begins automatically
    // <-round_middle
    // for
    //   <-action stuff
    // <- round end
    // <- round end done
    gs.L.SetExecutionLimit(250000)
    cmd := fmt.Sprintf("RoundStart(%t, %d)", g.Side == SideExplorers, (g.Turn+1)/2)
    base.Log().Printf("cmd: '%s'", cmd)
    gs.L.DoString(cmd)

    // signals to the game that we're done with the startup stuff
    g.comm.script_to_game <- nil
    base.Log().Printf("ScriptComm: Done with RoundStart")

    for {
      base.Log().Printf("ScriptComm: Waiting to verify action")
      _exec := <-g.comm.game_to_script
      base.Log().Printf("ScriptComm: Got exec: %v", _exec)
      if _exec == nil {
        base.Log().Printf("ScriptComm: No more exec: bailing")
        break
      }
      base.Log().Printf("ScriptComm: Verifying action")

      exec := _exec.(ActionExec)
      if vpath := exec.GetPath(); vpath != nil {
        gs.L.SetExecutionLimit(250000)
        exec.Push(gs.L, g)
        gs.L.NewTable()
        for i := range vpath {
          gs.L.PushInteger(i + 1)
          _, x, y := g.FromVertex(vpath[i])
          LuaPushPoint(gs.L, x, y)
          gs.L.SetTable(-3)
        }
        base.Log().Printf("Pathlength: %d", len(vpath))
        gs.L.SetGlobal("__path")
        LuaPushEntity(gs.L, g.EntityById(exec.EntityId()))
        gs.L.SetGlobal("__ent")
        cmd = fmt.Sprintf("__truncate = OnMove(__ent, __path)")
        base.Log().Printf("cmd: '%s'", cmd)
        func() {
          defer func() {
            if r := recover(); r != nil {
              base.Log().Printf("Error in OnMove(): ", r)
            }
          }()
          gs.L.DoString(cmd)
          gs.L.GetGlobal("__truncate")
          truncate := gs.L.ToInteger(-1)
          gs.L.Pop(1)
          base.Log().Printf("Truncating to length %d", truncate)
          exec.TruncatePath(truncate)
        }()
      }

      g.comm.script_to_game <- nil

      // The action is sent when it happens, and a nil is sent when it is done
      // being executed, we want to wait until then so that the game is in a
      // stable state before we do anything.
      <-g.comm.game_to_script
      base.Log().Printf("ScriptComm: Got action secondary")
      // Run OnAction here
      gs.L.SetExecutionLimit(250000)
      exec.Push(gs.L, g)
      //      base.Log().Printf("exec: ", LuaStringifyParam(gs.L, -1))
      gs.L.SetGlobal("__exec")
      cmd = fmt.Sprintf("OnAction(%t, %d, %s)", g.Side == SideExplorers, (g.Turn+1)/2, "__exec")
      base.Log().Printf("cmd: '%s'", cmd)
      gs.L.DoString(cmd)
      g.comm.script_to_game <- nil
      base.Log().Printf("ScriptComm: Done with OnAction")
    }

    gs.L.SetExecutionLimit(250000)
    gs.L.DoString(fmt.Sprintf("RoundEnd(%t, %d)", g.Side == SideExplorers, (g.Turn+1)/2))

    // Signal that we're done with the round end
    g.comm.script_to_game <- nil
    base.Log().Printf("ScriptComm: Done with RoundEnd")
  }()
}

// Can be called occassionally and will allow a script to progress whenever
// it is ready
func (gp *GamePanel) scriptThinkOnce() {
  if gp.script.L == nil {
    return
  }
  done := false
  for !done {
    select {
    // If a script has tried to run a function that requires running during
    // Think then it can run now and we'll wait for it to finish before
    // continuing.
    case gp.script.sync <- struct{}{}:
      <-gp.script.sync
    default:
      done = true
    }
  }
}

// Thinks continually until a value is passed along done
func (gp *GamePanel) scriptSitAndThink() (done chan<- struct{}) {
  done_chan := make(chan struct{})

  go func() {
    for {
      select {
      case <-gp.script.sync:
        <-gp.script.sync
      case <-done_chan:
        return
      }
    }
  }()

  return done_chan
}

func startScript(gp *GamePanel, player *Player) lua.GoFunction {
  return func(L *lua.State) int {
    if !LuaCheckParamsOk(L, "StartScript", LuaString) {
      return 0
    }
    gp.script.syncStart()
    defer gp.script.syncEnd()
    script := L.ToString(-1)
    startGameScript(gp, script, player)
    return 0
  }
}

func selectHouse(gp *GamePanel) lua.GoFunction {
  return func(L *lua.State) int {
    if !LuaCheckParamsOk(L, "SelectHouse") {
      return 0
    }
    gp.script.syncStart()
    defer gp.script.syncEnd()
    selector, output, err := MakeUiSelectMap(gp)
    if err != nil {
      base.Error().Printf("Error selecting map: %v", err)
      return 0
    }
    gp.AnchorBox.AddChild(selector, gui.Anchor{0.5, 0.5, 0.5, 0.5})
    gp.script.syncEnd()

    name := <-output
    base.Log().Printf("Received '%s'", name)
    gp.script.syncStart()
    gp.AnchorBox.RemoveChild(selector)
    base.Log().Printf("Removed seletor")
    L.PushString(name)
    return 1
  }
}

func loadHouse(gp *GamePanel) lua.GoFunction {
  return func(L *lua.State) int {
    if !LuaCheckParamsOk(L, "LoadHouse", LuaString) {
      return 0
    }
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

    gp.AnchorBox = gui.MakeAnchorBox(gui.Dims{1024, 700})

    gp.AnchorBox.AddChild(gp.viewer, gui.Anchor{0.5, 0.5, 0.5, 0.5})
    base.Log().Printf("Done making stuff")
    return 0
  }
}

func showMainBar(gp *GamePanel) lua.GoFunction {
  return func(L *lua.State) int {
    if !LuaCheckParamsOk(L, "ShowMainBar", LuaBoolean) {
      return 0
    }
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
      gp.main_bar, err = MakeMainBar(gp.game)
      if err != nil {
        LuaDoError(L, err.Error())
        return 0
      }
      gp.AnchorBox.AddChild(gp.main_bar, gui.Anchor{0.5, 0, 0.5, 0})
    }
    base.Log().Printf("Num kids: %d", len(gp.AnchorBox.GetChildren()))
    return 0
  }
}

func spawnEntityAtPosition(gp *GamePanel) lua.GoFunction {
  return func(L *lua.State) int {
    if !LuaCheckParamsOk(L, "SpawnEntityAtPosition", LuaString, LuaPoint) {
      return 0
    }
    gp.script.syncStart()
    defer gp.script.syncEnd()
    name := L.ToString(-2)
    x, y := LuaToPoint(L, -1)
    ent := MakeEntity(name, gp.game)
    if gp.game.SpawnEntity(ent, x, y) {
      LuaPushEntity(L, ent)
    } else {
      L.PushNil()
    }
    return 1
  }
}

func getSpawnPointsMatching(gp *GamePanel) lua.GoFunction {
  return func(L *lua.State) int {
    if !LuaCheckParamsOk(L, "GetSpawnPointsMatching", LuaString) {
      return 0
    }
    gp.script.syncStart()
    defer gp.script.syncEnd()
    spawn_pattern := L.ToString(-1)
    re, err := regexp.Compile(spawn_pattern)
    if err != nil {
      LuaDoError(L, fmt.Sprintf("Failed to compile regexp '%s': %v", spawn_pattern, err))
      return 0
    }
    L.NewTable()
    count := 0
    for _, sp := range gp.game.House.Floors[0].Spawns {
      if !re.MatchString(sp.Name) {
        continue
      }
      count++
      L.PushInteger(count)
      LuaPushSpawnPoint(L, gp.game, sp)
      L.SetTable(-3)
    }
    return 1
  }
}

func spawnEntitySomewhereInSpawnPoints(gp *GamePanel) lua.GoFunction {
  return func(L *lua.State) int {
    if !LuaCheckParamsOk(L, "SpawnEntitySomewhereInSpawnPoints", LuaString, LuaArray) {
      return 0
    }
    gp.script.syncStart()
    defer gp.script.syncEnd()
    name := L.ToString(-2)

    var tx, ty int
    count := 0
    L.PushNil()
    for L.Next(-2) != 0 {
      sp := LuaToSpawnPoint(L, gp.game, -1)
      L.Pop(1)
      if sp == nil {
        continue
      }
      sx, sy := sp.Pos()
      sdx, sdy := sp.Dims()
      for x := sx; x < sx+sdx; x++ {
        for y := sy; y < sy+sdy; y++ {
          if gp.game.IsCellOccupied(x, y) {
            continue
          }
          // This will choose a random position from all positions and giving
          // all positions an equal chance of being chosen.
          count++
          if rand.Intn(count) == 0 {
            tx = x
            ty = y
          }
        }
      }
    }
    if count == 0 {
      base.Error().Printf("Unable to find an available position to spawn")
      return 0
    }
    ent := MakeEntity(name, gp.game)
    if ent == nil {
      base.Error().Printf("Cannot make an entity named '%s', no such thing.", name)
      return 0
    }
    if gp.game.SpawnEntity(ent, tx, ty) {
      LuaPushEntity(L, ent)
    } else {
      L.PushNil()
    }
    return 1
  }
}

func placeEntities(gp *GamePanel) lua.GoFunction {
  return func(L *lua.State) int {
    if !LuaCheckParamsOk(L, "PlaceEntities", LuaString, LuaInteger, LuaTable) {
      return 0
    }
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
    gp.AnchorBox.AddChild(ep, gui.Anchor{0.5, 0.5, 0.5, 0.5})
    gp.script.syncEnd()

    placed := <-placed_chan
    L.NewTable()
    for i := range placed {
      L.PushInteger(i + 1)
      LuaPushEntity(L, placed[i])
      L.SetTable(-3)
    }

    gp.script.syncStart()
    gp.AnchorBox.RemoveChild(ep)
    gp.script.syncEnd()
    return 1
  }
}

func roomAtPos(gp *GamePanel) lua.GoFunction {
  return func(L *lua.State) int {
    if !LuaCheckParamsOk(L, "RoomAtPos", LuaPoint) {
      return 0
    }
    gp.script.syncStart()
    defer gp.script.syncEnd()
    x, y := LuaToPoint(L, -1)
    room, _, _ := gp.game.House.Floors[0].RoomFurnSpawnAtPos(x, y)
    for i, r := range gp.game.House.Floors[0].Rooms {
      if r == room {
        L.PushInteger(i)
        return 1
      }
    }
    LuaDoError(L, fmt.Sprintf("Tried to get the room at position (%d,%d), but there is no room there.", x, y))
    return 0
  }
}

func getAllEnts(gp *GamePanel) lua.GoFunction {
  return func(L *lua.State) int {
    if !LuaCheckParamsOk(L, "GetAllEnts") {
      return 0
    }
    gp.script.syncStart()
    defer gp.script.syncEnd()
    L.NewTable()
    for i := range gp.game.Ents {
      L.PushInteger(i + 1)
      LuaPushEntity(L, gp.game.Ents[i])
      L.SetTable(-3)
    }
    return 1
  }
}

func dialogBox(gp *GamePanel) lua.GoFunction {
  return func(L *lua.State) int {
    if !LuaCheckParamsOk(L, "DialogBox", LuaString) {
      return 0
    }
    gp.script.syncStart()
    defer gp.script.syncEnd()
    path := L.ToString(-1)
    box, output, err := MakeDialogBox(filepath.ToSlash(path))
    if err != nil {
      base.Error().Printf("Error making dialog: %v", err)
      return 0
    }
    gp.AnchorBox.AddChild(box, gui.Anchor{0.5, 0.5, 0.5, 0.5})
    gp.script.syncEnd()

    var choices []string
    for choice := range output {
      choices = append(choices, choice)
    }
    base.Log().Printf("Dialog box press: %v", choices)

    gp.script.syncStart()
    gp.AnchorBox.RemoveChild(box)
    L.NewTable()
    for i, choice := range choices {
      L.PushInteger(i + 1)
      L.PushString(choice)
      L.SetTable(-3)
    }
    return 1
  }
}

func pickFromN(gp *GamePanel) lua.GoFunction {
  return func(L *lua.State) int {
    if !LuaCheckParamsOk(L, "PickFromN", LuaInteger, LuaInteger, LuaTable) {
      return 0
    }
    min := L.ToInteger(-3)
    max := L.ToInteger(-2)
    var options []hui.Option
    var option_names []string
    L.PushNil()
    for L.Next(-2) != 0 {
      name := L.ToString(-2)
      option_names = append(option_names, name)
      path := L.ToString(-1)
      if !filepath.IsAbs(path) {
        path = filepath.Join(base.GetDataDir(), path)
      }
      option := iconWithText{
        Name: name,
        Icon: texture.Object{Path: base.Path(path)},
      }
      options = append(options, &option)
      L.Pop(1)
    }
    var selector hui.Selector
    if min == 1 && max == 1 {
      selector = hui.SelectExactlyOne
    } else {
      selector = hui.SelectInRange(min, max)
    }
    var chooser *hui.RosterChooser
    done := make(chan struct{})
    on_complete := func(m map[int]bool) {
      gp.RemoveChild(chooser)
      L.NewTable()
      count := 0
      for i := range options {
        if m[i] {
          count++
          L.PushInteger(count)
          L.PushString(option_names[i])
          L.SetTable(-3)
        }
      }
      done <- struct{}{}
    }
    chooser = hui.MakeRosterChooser(options, selector, on_complete, nil)
    gp.script.syncStart()
    gp.AddChild(chooser, gui.Anchor{0.5, 0.5, 0.5, 0.5})
    gp.script.syncEnd()
    <-done
    return 1
  }
}

func setGear(gp *GamePanel) lua.GoFunction {
  return func(L *lua.State) int {
    if !LuaCheckParamsOk(L, "SetGear", LuaEntity, LuaString) {
      return 0
    }
    gp.script.syncStart()
    defer gp.script.syncEnd()
    gear_name := L.ToString(-1)
    L.PushString("id")
    L.GetTable(-3)
    id := EntityId(L.ToInteger(-1))
    ent := gp.game.EntityById(id)
    if ent == nil {
      base.Error().Printf("Referenced an entity with id == %d which doesn't exist.", id)
      return 0
    }
    L.PushBoolean(ent.SetGear(gear_name))
    return 1
  }
}

// bindAi(target, source)
// bindAi("denizen", "denizen.lua")
// bindAi("intruder", "intruder.lua")
// bindAi("minions", "minions.lua")
// bindAi(ent, "fudgecake.lua")
// special sources: "human", "inactive", and in the future: "net"
// special targets: "denizen", "intruder", "minions", or an entity table
func bindAi(gp *GamePanel) lua.GoFunction {
  return func(L *lua.State) int {
    if !LuaCheckParamsOk(L, "BindAi", LuaAnything, LuaString) {
      return 0
    }
    gp.script.syncStart()
    defer gp.script.syncEnd()
    source := L.ToString(-1)
    if L.IsTable(-2) {
      L.PushString("id")
      L.GetTable(-3)
      target := EntityId(L.ToInteger(-1))
      L.Pop(1)
      ent := gp.game.EntityById(target)
      if ent == nil {
        base.Error().Printf("Referenced an entity with id == %d which doesn't exist.", target)
        return 0
      }
      ent.Ai_file_override = base.Path(filepath.Join(base.GetDataDir(), "ais", filepath.FromSlash(L.ToString(-1))))
      ent.LoadAi()
      return 0
    }
    target := L.ToString(-2)
    switch target {
    case "denizen":
      switch source {
      case "human":
        gp.game.haunts_ai = inactiveAi{}
      case "net":
        base.Error().Printf("bindAi('denizen', 'net') is not implemented.")
        return 0
      default:
        gp.game.haunts_ai = nil
        ai_maker(filepath.Join(base.GetDataDir(), "ais", source), gp.game, nil, &gp.game.haunts_ai, DenizensAi)
        if gp.game.haunts_ai == nil {
          gp.game.haunts_ai = inactiveAi{}
        }
      }
    case "intruder":
      switch source {
      case "human":
        gp.game.explorers_ai = inactiveAi{}
      case "net":
        base.Error().Printf("bindAi('intruder', 'net') is not implemented.")
        return 0
      default:
        gp.game.explorers_ai = nil
        ai_maker(filepath.Join(base.GetDataDir(), "ais", source), gp.game, nil, &gp.game.explorers_ai, IntrudersAi)
        if gp.game.explorers_ai == nil {
          gp.game.explorers_ai = inactiveAi{}
        }
      }
    case "minions":
      gp.game.minion_ai = nil
      ai_maker(filepath.Join(base.GetDataDir(), "ais", source), gp.game, nil, &gp.game.minion_ai, MinionsAi)
      if gp.game.minion_ai == nil {
        gp.game.minion_ai = inactiveAi{}
      }
    default:
      base.Error().Printf("Specified unknown Ai target '%s'", target)
      return 0
    }

    return 0
  }
}

func setVisibility(gp *GamePanel) lua.GoFunction {
  return func(L *lua.State) int {
    if !LuaCheckParamsOk(L, "SetVisibility", LuaString) {
      return 0
    }
    gp.script.syncStart()
    defer gp.script.syncEnd()
    side_str := L.ToString(-1)
    var side Side
    switch side_str {
    case "denizens":
      side = SideHaunt
    case "intruders":
      side = SideExplorers
    default:
      base.Error().Printf("Cannot pass '%s' as first parameter of setVisibility()", side_str)
      return 0
    }
    gp.game.SetVisibility(side)
    return 0
  }
}

func endPlayerInteraction(gp *GamePanel) lua.GoFunction {
  return func(L *lua.State) int {
    if !LuaCheckParamsOk(L, "EndPlayerInteraction") {
      return 0
    }
    gp.script.syncStart()
    defer gp.script.syncEnd()
    gp.game.player_active = false
    return 0
  }
}

func saveStore(gp *GamePanel, player *Player) lua.GoFunction {
  return func(L *lua.State) int {
    if !LuaCheckParamsOk(L, "SaveStore") {
      return 0
    }
    gp.script.syncStart()
    defer gp.script.syncEnd()
    UpdatePlayer(player, gp.script.L)
    err := SavePlayer(player)
    if err != nil {
      base.Warn().Printf("Unable to save player: %v", err)
    }
    return 0
  }
}

func setCondition(gp *GamePanel) lua.GoFunction {
  return func(L *lua.State) int {
    if !LuaCheckParamsOk(L, "SetCondition", LuaEntity, LuaString, LuaBoolean) {
      return 0
    }
    gp.script.syncStart()
    defer gp.script.syncEnd()
    ent := LuaToEntity(L, gp.game, -3)
    if ent == nil {
      base.Warn().Printf("Tried to SetCondition on an entity that doesn't exist.")
      return 0
    }
    name := L.ToString(-2)
    if L.ToBoolean(-1) {
      ent.Stats.ApplyCondition(status.MakeCondition(name))
    } else {
      ent.Stats.RemoveCondition(name)
    }
    return 0
  }
}

func setPosition(gp *GamePanel) lua.GoFunction {
  return func(L *lua.State) int {
    if !LuaCheckParamsOk(L, "SetPosition", LuaEntity, LuaPoint) {
      return 0
    }
    gp.script.syncStart()
    defer gp.script.syncEnd()
    ent := LuaToEntity(L, gp.game, -2)
    if ent == nil {
      base.Warn().Printf("Tried to SetPosition on an entity that doesn't exist.")
      return 0
    }
    x, y := LuaToPoint(L, -1)
    ent.X = float64(x)
    ent.Y = float64(y)
    return 0
  }
}

func setLosMode(gp *GamePanel) lua.GoFunction {
  return func(L *lua.State) int {
    if !LuaCheckParamsOk(L, "SetLosMode", LuaString, LuaAnything) {
      return 0
    }
    gp.script.syncStart()
    defer gp.script.syncEnd()
    side_str := L.ToString(-2)
    var mode_str string
    if L.IsString(-1) {
      mode_str = L.ToString(-1)
    } else {
      mode_str = "rooms"
    }
    var side Side
    switch side_str {
    case "denizens":
      side = SideHaunt
    case "intruders":
      side = SideExplorers
    default:
      base.Error().Printf("Cannot pass '%s' as first parameters of setLosMode()", side_str)
      return 0
    }
    switch mode_str {
    case "none":
      gp.game.SetLosMode(side, LosModeNone, nil)
    case "all":
      gp.game.SetLosMode(side, LosModeAll, nil)
    case "entities":
      gp.game.SetLosMode(side, LosModeEntities, nil)
    case "rooms":
      if !L.IsTable(-1) {
        base.Error().Printf("The last parameter to setLosMode should be an array of rooms if mode == 'rooms'")
        return 0
      }
      L.PushNil()
      all_rooms := gp.game.House.Floors[0].Rooms
      var rooms []*house.Room
      for L.Next(-2) != 0 {
        index := L.ToInteger(-1)
        if index < 0 || index > len(all_rooms) {
          base.Error().Printf("Tried to reference room #%d which doesn't exist.", index)
          continue
        }
        rooms = append(rooms, all_rooms[index])
        L.Pop(1)
      }
      gp.game.SetLosMode(side, LosModeRooms, rooms)

    default:
      base.Error().Printf("Unknown los mode '%s'", mode_str)
      return 0
    }
    return 0
  }
}

// Ripped from game/ai/ai.go - should probably sync up with it
func registerUtilityFunctions(L *lua.State) {
  L.Register("print", func(L *lua.State) int {
    var res string
    n := L.GetTop()
    for i := -n; i < 0; i++ {
      res += LuaStringifyParam(L, i) + " "
    }
    base.Log().Printf("GameScript(%p): %s", L, res)
    return 0
  })
}
