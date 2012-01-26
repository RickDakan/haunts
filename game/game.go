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
  minx,miny int
  v [][]floorVertex

  // Need to maintain a pointer to this because the state of doors can change.
  floor *house.Floor
}
func makeFloorGraph(floor *house.Floor) *floorGraph {
  var graph floorGraph
  graph.floor = floor

  var maxx,maxy int
  graph.minx = floor.Rooms[0].X
  maxx = floor.Rooms[0].X
  graph.miny = floor.Rooms[0].Y
  maxy = floor.Rooms[0].Y
  for _,room := range floor.Rooms {
    x2 := room.X + room.Size.Dx
    y2 := room.Y + room.Size.Dy
    if room.X < graph.minx { graph.minx = room.X }
    if room.Y < graph.miny { graph.miny = room.Y }
    if x2 > maxx { maxx = x2 }
    if y2 > maxy { maxy = y2 }
  }

  // Allocate a grid that can contain the entire floor, then mark
  // everything as impassable
  graph.v = make([][]floorVertex, maxx - graph.minx)
  for i := range graph.v {
    graph.v[i] = make([]floorVertex, maxy - graph.miny)
    for j := range graph.v[i] {
      graph.v[i][j].impassable = true
    }
  }

  // Now cut out regions of passable terrain where there are rooms,
  // then again fill in impassable regions where there is furniture.
  for _,room := range floor.Rooms {
    for x := room.X; x < room.X + room.Size.Dx; x++ {
      for y := room.Y; y < room.Y + room.Size.Dy; y++ {
        graph.v[x - graph.minx][y - graph.miny].impassable = false
      }
    }
    for _,furn := range room.Furniture {
      for x := furn.X; x < furn.X + furn.Orientations[furn.Rotation].Dx; x++ {
        for y := furn.Y; y < furn.Y + furn.Orientations[furn.Rotation].Dy; y++ {
          graph.v[room.X + x - graph.minx][room.Y + y - graph.miny].impassable = true
        }
      }
    }
  }

  return &graph
}

func (g *floorGraph) NumVertex() int {
  return len(g.v) * len(g.v[0])
}
func (g *floorGraph) fromVertex(v int) (int, int) {
  return v % len(g.v), v / len(g.v)
}
func (g *floorGraph) toVertex(x, y int) int {
  return x + y*len(g.v)
}
func (g *floorGraph) moveCost(x, y, x2, y2 int) float64 {
  cost_c := float64(g.v[x2][y2].movementCost)
  if cost_c < 0 || g.v[x2][y2].impassable {
    return -1
  }
  if x == x2 || y == y2 {
    return float64(cost_c + 1)
  }

  cost_a := g.v[x][y2].movementCost
  if cost_a < 0 || g.v[x][y2].impassable {
    return -1
  }
  cost_b := g.v[x2][y].movementCost
  if cost_b < 0 || g.v[x2][y].impassable {
    return -1
  }

  cost_ab := float64(cost_a+cost_b+2) / 2
  if cost_ab > cost_c {
    return cost_ab
  }
  return cost_c
}
func (g *floorGraph) Adjacent(v int) ([]int, []float64) {
  x, y := g.fromVertex(v)
  var adj []int
  var weight []float64

  // separate arrays for the adjacent diagonal cells, this way we make sure they are listed
  // at the end so that searches will prefer orthogonal adjacent cells
  var adj_diag []int
  var weight_diag []float64

  for dx := -1; dx <= 1; dx++ {
    if x+dx < 0 || x+dx >= len(g.v) {
      continue
    }
    for dy := -1; dy <= 1; dy++ {
      if dx == 0 && dy == 0 {
        continue
      }
      if y+dy < 0 || y+dy >= len(g.v[0]) {
        continue
      }

      // // Don't want to be able to walk through other units
      // occupied := false
      // for i := range g.Entities {
      //   if int(g.Entities[i].Pos.X) == x+dx && int(g.Entities[i].Pos.Y) == y+dy {
      //     occupied = true
      //     break
      //   }
      // }
      // if occupied {
      //   continue
      // }

      // Prevent moving along a diagonal if we couldn't get to that space normally via
      // either of the non-diagonal paths
      cost := g.moveCost(x, y, x+dx, y+dy)
      if cost < 0 {
        continue
      }
      if dx != 0 && dy != 0 {
        adj_diag = append(adj_diag, g.toVertex(x+dx, y+dy))
        weight_diag = append(weight_diag, cost)
      } else {
        adj = append(adj, g.toVertex(x+dx, y+dy))
        weight = append(weight, cost)
      }
    }
  }
  for i := range adj_diag {
    adj = append(adj, adj_diag[i])
    weight = append(weight, weight_diag[i])
  }
  return adj, weight
}

func MakeGamePanel() *GamePanel {
  var gp GamePanel
  gp.house = house.MakeHouseDef()
  gp.viewer = house.MakeHouseViewer(gp.house, 62)
  gp.HorizontalTable = gui.MakeHorizontalTable()
  gp.HorizontalTable.AddChild(gp.viewer)
  gp.ent = MakeEntity("Master of the Manse")
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
    _,res := algorithm.Dijkstra(gp.graph, []int{ gp.graph.toVertex(gp.ent.Pos()) }, []int{ gp.graph.toVertex(int(x), int(y)) })
    gp.ent.Path = algorithm.Map(res, [][2]int{}, func(a interface{}) interface{} {
      x,y := gp.graph.fromVertex(a.(int))
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
