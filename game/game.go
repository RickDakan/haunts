package game

import (
  "glop/gui"
  "glop/gin"
  "glop/util/algorithm"
  "haunts/house"
)

type GamePanel struct {
  *gui.HorizontalTable

  house  *house.HouseDef
  viewer *house.HouseViewer

  ent   *Entity
  graph *floorGraph

  // Keep track of this so we know how much time has passed between
  // calls to Think()
  last_think int64
}

type floorVertex struct {
  impassable   bool
  movementCost int
}
type floorGraph struct {
  // Need to maintain a pointer to this because the state of doors can change.
  floor *house.Floor
}
func makeFloorGraph(floor *house.Floor) *floorGraph {
  var graph floorGraph
  graph.floor = floor
  return &graph
}

func (g *floorGraph) NumVertex() int {
  total := 0
  for _,room := range g.floor.Rooms {
    total += room.Size.Dx * room.Size.Dy
  }
  return total
}
func (g *floorGraph) fromVertex(v int) (room *house.Room, x,y int) {
  for _,room := range g.floor.Rooms {
    size := room.Size.Dx * room.Size.Dy
    if v >= size {
      v -= size
      continue
    }
    return room, room.X + (v % room.Size.Dx), room.Y + (v / room.Size.Dx)
  }
  return nil, 0, 0
}
func (g *floorGraph) toVertex(x, y int) int {
  v := 0
  for _,room := range g.floor.Rooms {
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
func open(room *house.Room, x,y int) bool {
  for _,f := range room.Furniture {
    fx,fy := f.Pos()
    fdx,fdy := f.Dims()
    if x >= fx && x < fx + fdx && y >= fy && y < fy + fdy {
      return false
    }
  }
  return true
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
      return door.Opened
    }
  }
  return false
}

func (g *floorGraph) Adjacent(v int) ([]int, []float64) {
  room,x,y := g.fromVertex(v)
  var adj []int
  var weight []float64
  var moves [3][3]float64
  for dx := -1; dx <= 1; dx++ {
    for dy := -1; dy <= 1; dy++ {
      // Only run this loop if exactly one of dx and dy is non-zero
      if (dx == 0) == (dy == 0) { continue }
      tx := x + dx
      ty := y + dy
      // TODO: This is obviously inefficient
      troom,_,_ := g.fromVertex(g.toVertex(tx, ty))
      if troom == nil { continue }
      if !open(troom, tx - troom.X, ty - troom.Y) { continue }
      if !connected(room, troom, x, y, tx, ty) { continue }
      adj = append(adj, g.toVertex(tx, ty))
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
      // TODO: This is obviously inefficient
      troom,_,_ := g.fromVertex(g.toVertex(tx, ty))
      if troom == nil { continue }
      if !open(troom, tx - troom.X, ty - troom.Y) { continue }
      if !connected(room, troom, x, y, tx, ty) { continue }
      if !connected(troom, room, tx, ty, x, y) { continue }
      if moves[dx+1][1] == 0 || moves[1][dy+1] == 0 { continue }
      adj = append(adj, g.toVertex(tx, ty))
      w := (moves[dx+1][1] + moves[1][dy+1]) / 2
      moves[dx+1][dy+1] = w
      weight = append(weight, w)
    }
  }
  return adj, weight
}

func MakeGamePanel() *GamePanel {
  var gp GamePanel
  gp.house = house.MakeHouseDef()
  gp.viewer = house.MakeHouseViewer(gp.house, 62)
  gp.HorizontalTable = gui.MakeHorizontalTable()
  gp.HorizontalTable.AddChild(gp.viewer)
  gp.ent = MakeEntity("Angry Shade")
  gp.ent.X = 1
  gp.ent.Y = 2
  gp.viewer.AddDrawable(gp.ent)
  return &gp
}
func (gp *GamePanel) Think(ui *gui.Gui, t int64) {
  if gp.last_think == 0 {
    gp.last_think = t
  }
  dt := t - gp.last_think
  gp.last_think = t
  gp.ent.Think(dt)
}
func (gp *GamePanel) Respond(ui *gui.Gui, group gui.EventGroup) bool {
  if gp.HorizontalTable.Respond(ui, group) {
    return true
  }
  if found,event := group.FindEvent(gin.MouseLButton); found && event.Type == gin.Press {
    x,y := gp.viewer.WindowToBoard(event.Key.Cursor().Point())
    if x < 0 { x-- }
    if y < 0 { y-- }
    _,res := algorithm.Dijkstra(gp.graph, []int{ gp.graph.toVertex(gp.ent.Pos()) }, []int{ gp.graph.toVertex(int(x), int(y)) })
    gp.ent.Path = algorithm.Map(res, [][2]int{}, func(a interface{}) interface{} {
      _,x,y := gp.graph.fromVertex(a.(int))
      return [2]int{ int(x), int(y) }
    }).([][2]int)
    if len(gp.ent.Path) > 0 {
      gp.ent.Path = gp.ent.Path[1:]
    }
  }
  return false
}

func (gp *GamePanel) LoadHouse(name string) {
  gp.HorizontalTable.RemoveChild(gp.viewer)
  gp.house = house.MakeHouse(name)
  if len(gp.house.Floors) == 0 {
    gp.house = house.MakeHouseDef()
  }
  gp.graph = makeFloorGraph(gp.house.Floors[0])
  gp.viewer = house.MakeHouseViewer(gp.house, 62)
  gp.viewer.AddDrawable(gp.ent)
  gp.HorizontalTable.AddChild(gp.viewer)
}

func (gp *GamePanel) GetViewer() house.Viewer {
  return gp.viewer
}
