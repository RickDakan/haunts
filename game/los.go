package game

import (
  "math/rand"
  "path/filepath"
  "github.com/runningwild/glop/gui"
  "github.com/runningwild/glop/util/algorithm"
  "github.com/runningwild/haunts/house"
  "github.com/runningwild/haunts/base"
  "github.com/runningwild/haunts/game/status"
)

type Purpose int
const(
  PurposeNone Purpose = iota
  PurposeRelic
  PurposeMystery
  PurposeCleanse
)

type Game struct {
  // TODO: No idea if this thing can be loaded from the registry - should
  // probably figure that out at some point
  House *house.HouseDef
  Ents []*Entity  `registry:"loadfrom-entities"`

  // Next unique EntityId to be assigned
  Entity_id EntityId

  // The side of the human player in the case that this is a one player game,
  // this is so that if we load a saved game we know which side to put ais on.
  // TODO: IN a game with two humans we'll need something more complicated
  // here.
  Human Side

  // Current player
  Side Side

  // Current turn number - incremented on each OnRound() so every two
  // indicates that a complete round has happened.
  Turn int

  // The purpose that the explorers have for entering the house, chosen at the
  // beginning of the game.
  Purpose Purpose

  // The active cleanse points - when interacted with they will be removed
  // from this list, so in a Cleanse scenario the mission is accomplished
  // when this list is empty.  In other scenarios this list is always empty.
  Active_cleanses []*Entity

  // The active relics - when interacted it will be nilified
  // If the scenario is not a Relic mission this is always nil
  Active_relic *Entity


  // Transient data - none of the following are exported

  viewer *house.HouseViewer

  selection_tab      *gui.TabFrame
  haunts_selection   *gui.VerticalTable
  explorer_selection *gui.VerticalTable

  // If the user is dragging around a new Entity to place, this is it
  new_ent *Entity

  // Might want to keep several of them for different POVs, but one is
  // fine for now
  los_tex *house.LosTexture

  // When merging the los from different entities we'll do it here, and we
  // keep it around to avoid reallocating it every time we need it.
  los_full_merger []bool
  los_merger [][]bool

  selected_ent *Entity
  hovered_ent *Entity

  // Stores the current acting entity - if it is an Ai controlled entity
  ai_ent *Entity

  // This controls the minions on the haunting side, regardless of whether a
  // human or ai is controlling the rest of that side.
  minion_ai Ai

  haunts_ai Ai
  explorers_ai Ai

  // If an Ai is executing currently it is referenced here
  active_ai Ai

  action_state actionState
  current_exec   ActionExec
  current_action Action

  // Goals ******************************************************

  // Defaults to the zero value which is NoSide
  winner Side
}

func (g *Game) EntityById(id EntityId) *Entity {
  for i := range g.Ents {
    if g.Ents[i].Id == id {
      return g.Ents[i]
    }
  }
  return nil
}

func (g *Game) HoveredEnt() *Entity {
  return g.hovered_ent
}

func (g *Game) SelectEnt(ent *Entity) bool {
  if g.action_state != noAction {
    return false
  }
  found := false
  for i := range g.Ents {
    if g.Ents[i] == ent {
      found = true
      break
    }
  }
  if !found {
    return false
  }
  if g.selected_ent != nil {
    g.selected_ent.selected = false
    g.selected_ent.hovered = false
  }
  g.selected_ent = ent
  if g.selected_ent != nil {
    g.selected_ent.selected = true
  }
  g.viewer.Focus(ent.FPos())
  return true
}

func (g *Game) OnBegin() {
  for i := range g.Ents {
    if g.Ents[i].ExplorerEnt != nil && g.Ents[i].ExplorerEnt.Gear != nil {
      gear := g.Ents[i].ExplorerEnt.Gear
      if gear.Action != "" {
        g.Ents[i].Actions = append(g.Ents[i].Actions, MakeAction(gear.Action))
      }
      if gear.Condition != "" {
        g.Ents[i].Stats.ApplyCondition(status.MakeCondition(gear.Condition))
      }
    }
    if g.Ents[i].Stats != nil {
      g.Ents[i].Stats.OnBegin()
    }
  }
}

func (g *Game) PlaceInitialExplorers(ents []*Entity) {
  if len(ents) > 9 {
    base.Error().Printf("Cannot place more than 9 ents at a starting position.")
    return
  }

  var sp *house.SpawnPoint
  for _, s := range g.House.Floors[0].Spawns {
    if s.Type() == house.SpawnExplorers {
      sp = s
      break
    }
  }
  if sp == nil {
    base.Error().Printf("No initial explorer positions listed.")
    return
  }
  if sp.Dx * sp.Dy < len(ents) {
    base.Error().Printf("Not enough space to place the explorers.")
    return
  }
  x, y := sp.Pos()
  base.Log().Printf("Starting %d explorers at (%d, %d)", len(ents), x, y)
  var poss [][2]int
  for dx := 0; dx < sp.Dx; dx++ {
    for dy := 0; dy < sp.Dy; dy++ {
      poss = append(poss, [2]int{x+dx, y+dy})
    }
  }
  for i := range ents {
    g.Ents = append(g.Ents, ents[i])
    g.viewer.AddDrawable(ents[i])
    index := rand.Intn(len(poss))
    pos := poss[index]
    poss[index] = poss[len(poss)-1]
    poss = poss[0:len(poss)-1]
    ents[i].X = float64(pos[0])
    ents[i].Y = float64(pos[1])
  }
}

func (g *Game) checkWinConditions() {
  // Check for explorer win conditions
  explorer_win := false
  switch g.Purpose {
  case PurposeRelic:
    if g.Active_relic == nil {
      // If the relic has been taken and at least one explorer is standing in
      // an exit then the explorers win.
      for _, ent := range g.Ents {
        if ent.Side() != SideExplorers { continue }
        for _, spawn := range g.House.Floors[0].Spawns {
          if spawn.Type() != house.SpawnExit { continue }
          x, y := ent.Pos()
          sx, sy := spawn.Pos()
          sdx, sdy := spawn.Dims()
          if x >= sx && x < sx + sdx && y >= sy && y < sy + sdy {
            explorer_win = true
          }
        }
      }
    }

  case PurposeMystery:
  case PurposeCleanse:
    if len(g.Active_cleanses) == 0 {
      explorer_win = true
    }

  default:
    base.Log().Printf("No purpose set - unable to check for explorer win condition")
  }
  if explorer_win {
    base.Log().Printf("Explorers won - kaboom")
  }

  // Check for haunt win condition - all intruders dead
  haunts_win := true
  for i := range g.Ents {
    if g.Ents[i].Side() == SideExplorers {
      haunts_win = false
    }
  }
  if haunts_win {
    base.Log().Printf("Haunts won - kaboom")
  }

  if g.winner == SideNone {
    if explorer_win && !haunts_win {
      g.winner = SideExplorers
    }
    if haunts_win && !explorer_win {
      g.winner = SideHaunt
    }
    if explorer_win && haunts_win {
      // Let's just hope this is far beyond impossible
      base.Error().Printf("Both sides won at the same time omg!")
    }
  }
}

func (g *Game) OnRound() {
  // Don't end the round if any of the following are true
  // An action is currently executing
  if g.action_state != noAction { return }
  // Any master ai is still active
  if g.Side == SideHaunt && (g.minion_ai.Active() || g.haunts_ai.Active()) { return }

  g.Turn++
  if g.Side == SideExplorers {
    g.Side = SideHaunt
  } else {
    g.Side = SideExplorers
  }

  if g.Turn == 1 {
    var action_name string
    switch g.Purpose {
    case PurposeNone:
      base.Error().Printf("Explorers have not set a purpose")

    case PurposeCleanse:
      action_name = "Cleanse"
      // If this is a cleanse scenario we need to choose the active cleanse points
      cleanses := algorithm.Choose(g.Ents, func(a interface{}) bool {
        ent := a.(*Entity)
        return ent.ObjectEnt != nil && ent.ObjectEnt.Goal == GoalCleanse
      }).([]*Entity)
      count := 3
      if len(cleanses) < 3 {
        count = len(cleanses)
      }
      for i := 0; i < count; i++ {
        n := rand.Intn(len(cleanses))
        active := cleanses[n]
        cleanses[n] = cleanses[len(cleanses)-1]
        cleanses = cleanses[0:len(cleanses)-1]
        g.Active_cleanses = append(g.Active_cleanses, active)
      }
      for _, active := range g.Active_cleanses {
        base.Log().Printf("Active cleanse point: %s", active.Name)
      }

    case PurposeMystery:
      action_name = "Mystery"

    case PurposeRelic:
      action_name = "Relic"
      relics := algorithm.Choose(g.Ents, func(a interface{}) bool {
        ent := a.(*Entity)
        return ent.ObjectEnt != nil && ent.ObjectEnt.Goal == GoalRelic
      }).([]*Entity)
      if len(relics) == 0 {
        base.Warn().Printf("Unable to find any relics for the relic mission")
      } else {
        g.Active_relic = relics[rand.Intn(len(relics))]
        x,y := g.Active_relic.Pos()
        base.Log().Printf("Chose active relic at position %d %d", x, y)
      }
    }
    if action_name != "" {
      for i := range g.Ents {
        ent := g.Ents[i]
        if ent.Side() != SideExplorers { continue }
        action := MakeAction(action_name)
        ent.Actions = append(ent.Actions, action)
        for j := len(ent.Actions) - 1; j > 1; j-- {
          ent.Actions[j], ent.Actions[j-1] = ent.Actions[j-1], ent.Actions[j]
        }
      }
    }
  }

  if g.Turn <= 2 {
    g.viewer.RemoveDrawable(g.new_ent)
    g.new_ent = nil
  }
  if g.Turn < 2 {
    return
  }
  if g.Turn == 2 {
    g.viewer.Los_tex = g.los_tex
    g.OnBegin()
  }
  for i := range g.Ents {
    if g.Ents[i].Stats != nil && g.Ents[i].Stats.HpCur() <= 0 {
      g.viewer.RemoveDrawable(g.Ents[i])
    }
  }
  g.Ents = algorithm.Choose(g.Ents, func(a interface{}) bool {
    return a.(*Entity).Stats == nil || a.(*Entity).Stats.HpCur() > 0
  }).([]*Entity)
  g.minion_ai.Activate()
  g.haunts_ai.Activate()
  g.explorers_ai.Activate()

  for i := range g.Ents {
    if g.Ents[i].Side() == g.Side {
      g.Ents[i].OnRound()
    }
  }
  if g.selected_ent != nil {
    g.selected_ent.hovered = false
    g.selected_ent.selected = false
  }
  if g.hovered_ent != nil {
    g.hovered_ent.hovered = false
    g.hovered_ent.selected = false
  }
  g.selected_ent = nil
  g.hovered_ent = nil
}

type actionState int
const (
  noAction       actionState = iota

  // The Ai is running and determining the next action to run
  waitingAction

  // The player has selected an action and is determining whether or not to
  // use it, and how.
  preppingAction

  // An action is currently running, everything should pause while this runs.
  doingAction
)

func (g *Game) GetViewer() *house.HouseViewer {
  return g.viewer
}

func (g *Game) NumVertex() int {
  total := 0
  for _,room := range g.House.Floors[0].Rooms {
    total += room.Size.Dx * room.Size.Dy
  }
  return total
}
func (g *Game) FromVertex(v int) (room *house.Room, x,y int) {
  for _,room := range g.House.Floors[0].Rooms {
    size := room.Size.Dx * room.Size.Dy
    if v >= size {
      v -= size
      continue
    }
    return room, room.X + (v % room.Size.Dx), room.Y + (v / room.Size.Dx)
  }
  return nil, 0, 0
}
func (g *Game) ToVertex(x, y int) int {
  v := 0
  for _,room := range g.House.Floors[0].Rooms {
    if x >= room.X && y >= room.Y && x < room.X + room.Size.Dx && y < room.Y + room.Size.Dy {
      x -= room.X
      y -= room.Y
      v += x + y * room.Size.Dx
      break
    }
    v += room.Size.Dx * room.Size.Dy
  }
  return v
}

// x and y are given in room coordinates
func furnitureAt(room *house.Room, x,y int) *house.Furniture {
  for _,f := range room.Furniture {
    fx,fy := f.Pos()
    fdx,fdy := f.Dims()
    if x >= fx && x < fx + fdx && y >= fy && y < fy + fdy {
      return f
    }
  }
  return nil
}

// x and y are given in floor coordinates
func roomAt(floor *house.Floor, x,y int) *house.Room {
  for _,room := range floor.Rooms {
    rx,ry := room.Pos()
    rdx,rdy := room.Dims()
    if x >= rx && x < rx + rdx && y >= ry && y < ry + rdy {
      return room
    }
  }
  return nil
}

func connected(r,r2 *house.Room, x,y,x2,y2 int) bool {
  if r == r2 { return true }
  x -= r.X
  y -= r.Y
  x2 -= r2.X
  y2 -= r2.Y
  var facing house.WallFacing
  if x == 0 && x2 != 0 {
    facing = house.NearLeft
  } else if y == 0 && y2 != 0 {
    facing = house.NearRight
  } else if x != 0 && x2 == 0 {
    facing = house.FarRight
  } else if y != 0 && y2 == 0 {
    facing = house.FarLeft
  } else {
    // This shouldn't happen, but in case it does we certainly shouldn't treat
    // it as an open door
    return false
  }
  for _,door := range r.Doors {
    if door.Facing != facing { continue }
    var pos int
    switch facing {
      case house.NearLeft:
      fallthrough
      case house.FarRight:
        pos = y

      case house.NearRight:
        fallthrough
      case house.FarLeft:
        pos = x
    }
    if pos >= door.Pos && pos < door.Pos + door.Width {
      if !door.Opened {
        base.Log().Printf("Door is closed!")
      }
      return door.Opened
    }
  }
  return false
}

func (g *Game) IsCellOccupied(x,y int) bool {
  r := roomAt(g.House.Floors[0], x, y)
  if r == nil { return true }
  f := furnitureAt(r, x - r.X, y - r.Y)
  if f != nil { return true }
  for _,ent := range g.Ents {
    ex,ey := ent.Pos()
    if x == ex && y == ey { return true }
  }
  return false
}

type exclusionGraph struct {
  ex map[*Entity]bool
  g *Game
}
func (eg *exclusionGraph) Adjacent(v int) ([]int, []float64) {
  return eg.g.Adjacent(v, eg.ex)
}
func (eg *exclusionGraph) NumVertex() int {
  return eg.g.NumVertex()
}

func (g *Game) RecalcLos() {
  for i := range g.Ents {
    if g.Ents[i].los != nil {
      g.Ents[i].los.x = -1
    }
  }
}

func (g *Game) Graph(exclude []*Entity) algorithm.Graph {
  ex := make(map[*Entity]bool, len(exclude))
  for i := range exclude {
    ex[exclude[i]] = true
  }
  return &exclusionGraph{ex, g}
}

func (g *Game) Adjacent(v int, ex map[*Entity]bool) ([]int, []float64) {
  room,x,y := g.FromVertex(v)
  var adj []int
  var weight []float64
  var moves [3][3]float64
  ent_occupied := make(map[[2]int]bool)
  for _,ent := range g.Ents {
    if ex[ent] { continue }
    x,y := ent.Pos()
    dx,dy := ent.Dims()
    for i := x; i < x+dx; i++ {
      for j := y; j < y+dy; j++ {
        ent_occupied[[2]int{ i, j }] = true
      }
    }
  }
  for dx := -1; dx <= 1; dx++ {
    for dy := -1; dy <= 1; dy++ {
      // Only run this loop if exactly one of dx and dy is non-zero
      if (dx == 0) == (dy == 0) { continue }
      tx := x + dx
      ty := y + dy
      if ent_occupied[[2]int{tx,ty}] { continue }
      if g.los_tex.Pix()[tx][ty] < house.LosVisibilityThreshold { continue }
      // TODO: This is obviously inefficient
      troom,_,_ := g.FromVertex(g.ToVertex(tx, ty))
      if troom == nil { continue }
      if furnitureAt(troom, tx - troom.X, ty - troom.Y) != nil { continue }
      if !connected(room, troom, x, y, tx, ty) { continue }
      adj = append(adj, g.ToVertex(tx, ty))
      moves[dx+1][dy+1] = 1
      weight = append(weight, 1)
    }
  }
  for dx := -1; dx <= 1; dx++ {
    for dy := -1; dy <= 1; dy++ {
      // Only run this loop if both dx and dy are non-zero (moving diagonal)
      if (dx == 0) != (dy == 0) { continue }
      tx := x + dx
      ty := y + dy
      if ent_occupied[[2]int{tx,ty}] { continue }
      if g.los_tex.Pix()[tx][ty] < house.LosVisibilityThreshold { continue }
      // TODO: This is obviously inefficient
      troom,_,_ := g.FromVertex(g.ToVertex(tx, ty))
      if troom == nil { continue }
      if furnitureAt(troom, tx - troom.X, ty - troom.Y) != nil { continue }
      if !connected(room, troom, x, y, tx, ty) { continue }
      if !connected(troom, room, tx, ty, x, y) { continue }
      if moves[dx+1][1] == 0 || moves[1][dy+1] == 0 { continue }
      adj = append(adj, g.ToVertex(tx, ty))
      w := (moves[dx+1][1] + moves[1][dy+1]) / 2
      moves[dx+1][dy+1] = w
      weight = append(weight, w)
    }
  }
  return adj, weight
}

func (g *Game) setup() {
  g.los_tex = house.MakeLosTexture()
  g.los_full_merger = make([]bool, house.LosTextureSizeSquared)
  g.los_merger = make([][]bool, house.LosTextureSize)
  for i := range g.los_merger {
    g.los_merger[i] = g.los_full_merger[i * house.LosTextureSize : (i + 1) * house.LosTextureSize]
  }
  for i := range g.Ents {
    if g.Ents[i].Side() == g.Side {
      g.DetermineLos(g.Ents[i], true)
    }
  }

  g.MergeLos(g.Ents)

  ai_maker(filepath.Join(base.GetDataDir(), "ais", "random_minion.xgml"), g, nil, &g.minion_ai, MinionsAi)
  if g.Human == SideExplorers {
    ai_maker(filepath.Join(base.GetDataDir(), "ais", "denizens.xgml"), g, nil, &g.haunts_ai, DenizensAi)
    g.explorers_ai = inactiveAi{}
  } else {
    ai_maker(filepath.Join(base.GetDataDir(), "ais", "intruders.xgml"), g, nil, &g.explorers_ai, IntrudersAi)
    g.haunts_ai = inactiveAi{}
  }
}

func makeGame(h *house.HouseDef, viewer *house.HouseViewer, side Side) *Game {
  var g Game
  g.Human = side
  g.Side = SideExplorers
  g.House = h
  g.House.Normalize()
  g.viewer = viewer

  // This way an unset id will be invalid
  g.Entity_id = 1

  g.setup()

  g.explorer_selection = gui.MakeVerticalTable()
  g.explorer_selection.AddChild(gui.MakeTextLine("standard", "foo", 300, 1, 1, 1, 1))
  g.haunts_selection = gui.MakeVerticalTable()
  g.haunts_selection.AddChild(gui.MakeTextLine("standard", "bar", 300, 1, 1, 1, 1))
  g.selection_tab = gui.MakeTabFrame([]gui.Widget{g.explorer_selection, g.haunts_selection})
  return &g
}

func (g *Game) Think(dt int64) {
  if g.current_exec != nil {
    ent := g.EntityById(g.current_exec.EntityId())
    g.current_action = ent.Actions[g.current_exec.ActionIndex()]
    g.action_state = doingAction
  }

  // If there is an action that is currently executing we need to advance that
  // action.
  if g.action_state == doingAction {
    res := g.current_action.Maintain(dt, g, g.current_exec)
    g.current_exec = nil
    switch res {
      case Complete:
        g.current_action.Cancel()
        g.current_action = nil
        g.action_state = noAction
        g.checkWinConditions()

      case InProgress:
      case CheckForInterrupts:
    }
  }

  g.viewer.Floor_drawer = g.current_action
  for i := range g.Ents {
    g.Ents[i].Think(dt)
  }
  var side_ents []*Entity
  for i := range g.Ents {
    if g.Ents[i].Side() == g.Side {
      g.DetermineLos(g.Ents[i], false)
      side_ents = append(side_ents, g.Ents[i])
    }
  }
  g.MergeLos(side_ents)
  pix := g.los_tex.Pix()
  amt := dt / 5
  mod := false
  for i := range pix {
    for j := range pix[i] {
      v := int64(pix[i][j])
      if v < house.LosVisibilityThreshold {
        v -= amt
      } else {
        v += amt
      }
      if v < house.LosMinVisibility {
        v = house.LosMinVisibility
      }
      if v < 0 { v = 0 }
      if v > 255 { v = 255 }
      mod = mod || (byte(v) != pix[i][j])
      pix[i][j] = byte(v)
    }
  }
  if mod {
    g.los_tex.Remap()
  }

  // Don't do any ai stuff if there is a pending action
  if g.current_action != nil {
    return
  }

  // If any entities are not either ready or dead let's wait until they are
  // before we do any of the ai stuff
  for _,ent := range g.Ents {
    state := ent.sprite.Sprite().AnimState()
    if state != "ready" && state != "killed" {
      return
    }
    if ent.sprite.Sprite().NumPendingCmds() > 0 {
      return
    }
  }

  // Do Ai - if there is any to do
  if g.Side == SideHaunt {
    if g.minion_ai.Active() {
      g.active_ai = g.minion_ai
      g.action_state = waitingAction
    } else {
      if g.haunts_ai.Active() {
        g.active_ai = g.haunts_ai
        g.action_state = waitingAction
      }
    }
  } else {
    if g.explorers_ai.Active() {
      g.active_ai = g.explorers_ai
      g.action_state = waitingAction
    }
  }
  if g.action_state == waitingAction {
    select {
    case exec := <-g.active_ai.ActionExecs():
      base.Log().Printf("Got %v from master", exec)
      if exec != nil {
        g.current_exec = exec
      } else {
        g.action_state = noAction
        // TODO: indicate that the master ai can go now
      }

    default:
    }
  }
}

func (g *Game) doLos(dist int, line [][2]int, los [][]bool) {
  var x0,y0,x,y int
  var room0,room *house.Room
  x, y = line[0][0], line[0][1]
  los[x][y] = true
  room = roomAt(g.House.Floors[0], x, y)
  for _,p := range line[1:] {
    x0,y0 = x,y
    x,y = p[0], p[1]
    room0 = room
    room = roomAt(g.House.Floors[0], x, y)
    if room == nil { return }
    if x == x0 || y == y0 {
      if room0 != nil && room0 != room && !connected(room, room0, x, y, x0, y0) { return }
    } else {
      roomA := roomAt(g.House.Floors[0], x0, y0)
      roomB := roomAt(g.House.Floors[0], x, y0)
      roomC := roomAt(g.House.Floors[0], x0, y)
      if roomA != nil && roomB != nil && roomA != roomB && !connected(roomA, roomB, x0, y0, x, y0) { return }
      if roomA != nil && roomC != nil && roomA != roomC && !connected(roomA, roomC, x0, y0, x0, y) { return }
      if roomB != nil && room != roomB && !connected(room, roomB, x, y, x, y0) { return }
      if roomC != nil && room != roomC && !connected(room, roomC, x, y, x0, y) { return }
    }
    furn := furnitureAt(room, x - room.X, y - room.Y)
    if furn != nil && furn.Blocks_los { return }
    dist -= 1  // or whatever
    if dist <= 0 { return }
    los[x][y] = true
  }
}

func (g *Game) MergeLos(ents []*Entity) {
  for i := range g.los_full_merger {
    g.los_full_merger[i] = false
  }
  for _,ent := range ents {
    if ent.los == nil { continue }
    for i := ent.los.minx; i <= ent.los.maxx; i++ {
      for j := ent.los.miny; j <= ent.los.maxy; j++ {
        if ent.los.grid[i][j] {
          g.los_merger[i][j] = true
        }
      }
    }
  }
  pix := g.los_tex.Pix()
  for i := 0; i < len(pix); i++ {
    for j := 0; j < len(pix); j++ {
      if g.los_merger[i][j] { continue }
      if pix[i][j] >= house.LosVisibilityThreshold {
        pix[i][j] = house.LosVisibilityThreshold - 1
      }
    }
  }
  for i := range g.los_merger {
    for j := range g.los_merger[i] {
      if g.los_merger[i][j] {
        if pix[i][j] < house.LosVisibilityThreshold {
          pix[i][j] = house.LosVisibilityThreshold
        }
      }
    }
  }
}

func (g *Game) DetermineLos(ent *Entity, force bool) {
  if ent.los == nil || ent.Stats == nil { return }
  ex,ey := ent.Pos()
  if !force && ex == ent.los.x && ey == ent.los.y { return }
  for i := range ent.los.grid {
    for j := range ent.los.grid[i] {
      ent.los.grid[i][j] = false
    }
  }
  ent.los.x = ex
  ent.los.y = ey

  minx := ex - ent.Stats.Sight()
  miny := ey - ent.Stats.Sight()
  maxx := ex + ent.Stats.Sight()
  maxy := ey + ent.Stats.Sight()
  line := make([][2]int, ent.Stats.Sight())
  for x := minx; x <= maxx; x++ {
    line = line[0:0]
    bresenham(ex, ey, x, miny, &line)
    g.doLos(ent.Stats.Sight(), line, ent.los.grid)
    line = line[0:0]
    bresenham(ex, ey, x, maxy, &line)
    g.doLos(ent.Stats.Sight(), line, ent.los.grid)
  }
  for y := miny; y <= maxy; y++ {
    line = line[0:0]
    bresenham(ex, ey, minx, y, &line)
    g.doLos(ent.Stats.Sight(), line, ent.los.grid)
    line = line[0:0]
    bresenham(ex, ey, maxx, y, &line)
    g.doLos(ent.Stats.Sight(), line, ent.los.grid)
  }

  ent.los.minx = len(ent.los.grid)
  ent.los.miny = len(ent.los.grid)
  ent.los.maxx = 0
  ent.los.maxy = 0
  for i := range ent.los.grid {
    for j := range ent.los.grid[i] {
      if ent.los.grid[i][j] {
        if i < ent.los.minx {
          ent.los.minx = i
        }
        if j < ent.los.miny {
          ent.los.miny = j
        }
        if i > ent.los.maxx {
          ent.los.maxx = i
        }
        if j > ent.los.maxy {
          ent.los.maxy = j
        }
      }
    }
  }
}

// Uses Bresenham's alogirthm to determine the points to rasterize a line from
// x,y to x2,y2.
func bresenham(x, y, x2, y2 int, res *[][2]int) {
  dx := x2 - x
  if dx < 0 {
    dx = -dx
  }
  dy := y2 - y
  if dy < 0 {
    dy = -dy
  }

  steep := dy > dx
  if steep {
    x, y = y, x
    x2, y2 = y2, x2
    dx, dy = dy, dx
  }

  err := dx >> 1
  cy := y

  xstep := 1
  if x2 < x {
    xstep = -1
  }
  ystep := 1
  if y2 < y {
    ystep = -1
  }
  for cx := x; cx != x2; cx += xstep {
    if !steep {
      *res = append(*res, [2]int{cx, cy})
    } else {
      *res = append(*res, [2]int{cy, cx})
    }
    err -= dy
    if err < 0 {
      cy += ystep
      err += dx
    }
  }
  if !steep {
    *res = append(*res, [2]int{x2, cy})
  } else {
    *res = append(*res, [2]int{cy, x2})
  }
}
