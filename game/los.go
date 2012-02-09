package game

import (
  "haunts/house"
)

type Game struct {
  Defname string

  // TODO: No idea if this thing can be loaded from the registry - should
  // probably figure that out at some point
  house *house.HouseDef

  viewer *house.HouseViewer

  // Might want to keep several of them for different POVs, but one is
  // fine for now
  los_tex *house.LosTexture

  Ents []*Entity  `registry:"loadfrom-entities"`

  action_state actionState
  current_action Action
}

type actionState int
const (
  noAction       actionState = iota
  preppingAction
  doingAction
)

func (g *Game) GetViewer() *house.HouseViewer {
  return g.viewer
}

func (g *Game) NumVertex() int {
  total := 0
  for _,room := range g.house.Floors[0].Rooms {
    total += room.Size.Dx * room.Size.Dy
  }
  return total
}
func (g *Game) FromVertex(v int) (room *house.Room, x,y int) {
  for _,room := range g.house.Floors[0].Rooms {
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
  for _,room := range g.house.Floors[0].Rooms {
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
      return door.Opened
    }
  }
  return false
}

func (g *Game) Adjacent(v int) ([]int, []float64) {
  room,x,y := g.FromVertex(v)
  var adj []int
  var weight []float64
  var moves [3][3]float64
  ent_occupied := make(map[[2]int]bool)
  for _,ent := range g.Ents {
    x,y := ent.Pos()
    ent_occupied[[2]int{ x, y }] = true
  }
  for dx := -1; dx <= 1; dx++ {
    for dy := -1; dy <= 1; dy++ {
      // Only run this loop if exactly one of dx and dy is non-zero
      if (dx == 0) == (dy == 0) { continue }
      tx := x + dx
      ty := y + dy
      if ent_occupied[[2]int{tx,ty}] { continue }
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

func makeGame(h *house.HouseDef, viewer *house.HouseViewer) *Game {
  var g Game
  g.house = h
  g.viewer = viewer

  g.Ents = append(g.Ents, MakeEntity("Angry Shade"))
  g.Ents[0].X = 3
  g.Ents[0].Y = 3
  g.viewer.AddDrawable(g.Ents[0])

  g.Ents = append(g.Ents, MakeEntity("Ghost Hunter"))
  g.Ents[1].X = 2
  g.Ents[1].Y = 3
  g.viewer.AddDrawable(g.Ents[1])

  g.Ents = append(g.Ents, MakeEntity("Master of the Manse"))
  g.Ents[2].X = 1
  g.Ents[2].Y = 1
  g.viewer.AddDrawable(g.Ents[2])

  g.los_tex = house.MakeLosTexture(256)
  g.los_tex.Remap(-20, -20)
  for i := range g.Ents[:1] {
    g.DetermineLos(g.Ents[i], true)
  }
  return &g
}

func (g *Game) Think(dt int64) {
  g.viewer.Floor_drawer = g.current_action
  for i := range g.Ents {
    g.Ents[i].Think(dt)
  }
  for i := range g.Ents[:1] {
    g.DetermineLos(g.Ents[i], false)
  }
  pix,_,_ := g.los_tex.Pix()
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
    g.los_tex.Remap(-20, -20)
  }
}

func (g *Game) doLos(dist int, line [][2]int, los map[[2]int]bool) {
  los[line[0]] = true
  var x0,y0,x,y int
  var room0,room *house.Room
  x, y = line[0][0], line[0][1]
  room = roomAt(g.house.Floors[0], x, y)
  for _,p := range line[1:] {
    x0,y0 = x,y
    x,y = p[0], p[1]
    room0 = room
    room = roomAt(g.house.Floors[0], x, y)
    if room == nil { return }
    if x == x0 || y == y0 {
      if room0 != nil && room0 != room && !connected(room, room0, x, y, x0, y0) { return }
    } else {
      roomA := roomAt(g.house.Floors[0], x0, y0)
      roomB := roomAt(g.house.Floors[0], x, y0)
      roomC := roomAt(g.house.Floors[0], x0, y)
      if roomA != nil && roomB != nil && roomA != roomB && !connected(roomA, roomB, x0, y0, x, y0) { return }
      if roomA != nil && roomC != nil && roomA != roomC && !connected(roomA, roomC, x0, y0, x0, y) { return }
      if roomB != nil && room != roomB && !connected(room, roomB, x, y, x, y0) { return }
      if roomC != nil && room != roomC && !connected(room, roomC, x, y, x0, y) { return }
    }
    furn := furnitureAt(room, x - room.X, y - room.Y)
    if furn != nil && furn.Blocks_los { return }
    dist -= 1  // or whatever
    if dist <= 0 { return }
    los[p] = true
  }
}

func (g *Game) DetermineLos(ent *Entity, force bool) {
  ex,ey := ent.Pos()
  if !force && ex == ent.losx && ey == ent.losy { return }
  ent.los = make(map[[2]int]bool)
  ent.losx = ex
  ent.losy = ey

  minx := ex - ent.Stats.Sight()
  miny := ey - ent.Stats.Sight()
  maxx := ex + ent.Stats.Sight()
  maxy := ey + ent.Stats.Sight()
  for x := minx; x <= maxx; x++ {
    g.doLos(ent.Stats.Sight(), bresenham(ex, ey, x, miny), ent.los)
    g.doLos(ent.Stats.Sight(), bresenham(ex, ey, x, maxy), ent.los)
  }
  for y := miny; y <= maxy; y++ {
    g.doLos(ent.Stats.Sight(), bresenham(ex, ey, minx, y), ent.los)
    g.doLos(ent.Stats.Sight(), bresenham(ex, ey, maxx, y), ent.los)
  }
  ltx,lty,ltx2,lty2 := g.los_tex.Region()
  for i := ltx; i <= ltx2; i++ {
    for j := lty; j <= lty2; j++ {
      if ent.los[[2]int{i,j}] { continue }
      if g.los_tex.Get(i, j) >= house.LosVisibilityThreshold {
        g.los_tex.Set(i, j, house.LosVisibilityThreshold - 1)
      }
    }
  }
  for p := range ent.los {
    if g.los_tex.Get(p[0], p[1]) < house.LosVisibilityThreshold {
      g.los_tex.Set(p[0], p[1], house.LosVisibilityThreshold)
    }
  }
}

// Uses Bresenham's alogirthm to determine the points to rasterize a line from
// x,y to x2,y2.
func bresenham(x, y, x2, y2 int) [][2]int {
  dx := x2 - x
  if dx < 0 {
    dx = -dx
  }
  dy := y2 - y
  if dy < 0 {
    dy = -dy
  }

  var ret [][2]int
  steep := dy > dx
  if steep {
    x, y = y, x
    x2, y2 = y2, x2
    dx, dy = dy, dx
    ret = make([][2]int, dy)[0:0]
  } else {
    ret = make([][2]int, dx)[0:0]
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
      ret = append(ret, [2]int{cx, cy})
    } else {
      ret = append(ret, [2]int{cy, cx})
    }
    err -= dy
    if err < 0 {
      cy += ystep
      err += dx
    }
  }
  if !steep {
    ret = append(ret, [2]int{x2, cy})
  } else {
    ret = append(ret, [2]int{cy, x2})
  }
  return ret
}
