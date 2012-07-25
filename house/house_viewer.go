package house

import (
  "math"
  "github.com/runningwild/glop/gui"
  "github.com/runningwild/glop/util/algorithm"
  "github.com/runningwild/mathgl"
)

// This structure is used for temporary doors (that are being dragged around in
// the editor) since the normal Door struct only handles one door out of a pair
// and doesn't know what rooms it connects.
type doorInfo struct {
  Door  *Door
  Valid bool
}

type HouseViewer struct {
  gui.Childless
  gui.BasicZone
  gui.NonFocuser

  house *HouseDef

  zoom, angle, fx, fy float32
  floor, ifloor       mathgl.Mat4

  // target[xy] are the values that f[xy] approach, this gives us a nice way
  // to change what the camera is looking at
  targetx, targety float32
  target_on        bool

  // Need to keep track of time so we can measure time between thinks
  last_timestamp int64

  drawables     []Drawable
  Los_tex       *LosTexture
  Floor_drawer  FloorDrawer
  Floor_drawers []FloorDrawer
  Edit_mode     bool

  bounds struct {
    on  bool
    min struct{ x, y float32 }
    max struct{ x, y float32 }
  }

  Temp struct {
    Room *Room

    // If we have a temporary door then this is the room it is attached to
    Door_room *Room
    Door_info doorInfo

    Spawn *SpawnPoint
  }

  // Keeping a variety of slices here so that we don't keep allocating new
  // ones every time we render everything
  rooms          []RectObject
  temp_drawables []Drawable
  all_furn       []*Furniture
  spawns         []*SpawnPoint
  floor_drawers  []FloorDrawer
}

func MakeHouseViewer(house *HouseDef, angle float32) *HouseViewer {
  var hv HouseViewer
  hv.Request_dims.Dx = 100
  hv.Request_dims.Dy = 100
  hv.Ex = true
  hv.Ey = true
  hv.house = house
  hv.angle = angle
  hv.Zoom(1)

  if hv.house != nil && len(hv.house.Floors) > 0 && len(hv.house.Floors[0].Rooms) > 0 {
    minx := hv.house.Floors[0].Rooms[0].X
    maxx := minx
    miny := hv.house.Floors[0].Rooms[0].Y
    maxy := miny
    for _, floor := range hv.house.Floors {
      for _, room := range floor.Rooms {
        if room.X < minx {
          minx = room.X
        }
        if room.Y < miny {
          miny = room.Y
        }
        if room.X+room.Size.Dx > maxx {
          maxx = room.X + room.Size.Dx
        }
        if room.Y+room.Size.Dy > maxy {
          maxy = room.Y + room.Size.Dy
        }
      }
    }
    hv.SetBounds(minx, miny, maxx, maxy)
  }

  return &hv
}

func (hv *HouseViewer) Respond(g *gui.Gui, group gui.EventGroup) bool {
  return false
}

func (hv *HouseViewer) Think(g *gui.Gui, t int64) {
  dt := t - hv.last_timestamp
  if hv.last_timestamp == 0 {
    dt = 0
  }
  hv.last_timestamp = t
  if hv.target_on {
    f := mathgl.Vec2{hv.fx, hv.fy}
    v := mathgl.Vec2{hv.targetx, hv.targety}
    v.Subtract(&f)
    scale := 1 - float32(math.Pow(0.005, float64(dt)/1000))
    v.Scale(scale)
    f.Add(&v)
    hv.fx = f.X
    hv.fy = f.Y
  }
}

func (hv *HouseViewer) AddDrawable(d Drawable) {
  hv.drawables = append(hv.drawables, d)
}
func (hv *HouseViewer) RemoveDrawable(d Drawable) {
  algorithm.Choose2(&hv.drawables, func(t Drawable) bool {
    return t != d
  })
}

func (hv *HouseViewer) modelviewToBoard(mx, my float32) (x, y, dist float32) {
  mz := d2p(hv.floor, mathgl.Vec3{mx, my, 0}, mathgl.Vec3{0, 0, 1})
  v := mathgl.Vec4{X: mx, Y: my, Z: mz, W: 1}
  v.Transform(&hv.ifloor)
  return v.X, v.Y, mz
}

func (hv *HouseViewer) boardToModelview(mx, my float32) (x, y, z float32) {
  v := mathgl.Vec4{X: mx, Y: my, W: 1}
  v.Transform(&hv.floor)
  x, y, z = v.X, v.Y, v.Z
  return
}

func (hv *HouseViewer) WindowToBoard(wx, wy int) (float32, float32) {
  hv.floor, hv.ifloor, _, _, _, _ = makeRoomMats(&roomDef{}, hv.Render_region, hv.fx, hv.fy, hv.angle, hv.zoom)

  fx, fy, _ := hv.modelviewToBoard(float32(wx), float32(wy))
  return fx, fy
}

func (hv *HouseViewer) BoardToWindow(bx, by float32) (int, int) {
  hv.floor, hv.ifloor, _, _, _, _ = makeRoomMats(&roomDef{}, hv.Render_region, hv.fx, hv.fy, hv.angle, hv.zoom)

  fx, fy, _ := hv.boardToModelview(bx, by)
  return int(fx), int(fy)
}

// Changes the current zoom from e^(zoom) to e^(zoom+dz)
func (hv *HouseViewer) Zoom(dz float64) {
  if dz == 0 {
    return
  }
  exp := math.Log(float64(hv.zoom)) + dz
  // This effectively clamps the 100x150 sprites to a width in the range
  // [25,100]
  exp = float64(clamp(float32(exp), 2.87130468509059, 4.25759904621048))
  hv.zoom = float32(math.Exp(exp))
}

func (hv *HouseViewer) SetBounds(x, y, x2, y2 int) {
  hv.bounds.on = true
  hv.bounds.min.x = float32(x)
  hv.bounds.min.y = float32(y)
  hv.bounds.max.x = float32(x2)
  hv.bounds.max.y = float32(y2)
}

func (hv *HouseViewer) Drag(dx, dy float64) {
  v := mathgl.Vec3{X: hv.fx, Y: hv.fy}
  vx := mathgl.Vec3{1, -1, 0}
  vx.Normalize()
  vy := mathgl.Vec3{1, 1, 0}
  vy.Normalize()
  vx.Scale(float32(dx) / hv.zoom * 2)
  vy.Scale(float32(dy) / hv.zoom * 2)
  v.Add(&vx)
  v.Add(&vy)
  if hv.bounds.on {
    hv.fx = clamp(v.X, hv.bounds.min.x, hv.bounds.max.x)
    hv.fy = clamp(v.Y, hv.bounds.min.y, hv.bounds.max.y)
  } else {
    hv.fx, hv.fy = v.X, v.Y
  }
  hv.target_on = false
}

func (hv *HouseViewer) Focus(bx, by float64) {
  hv.targetx = float32(bx)
  hv.targety = float32(by)
  hv.target_on = true
}

func (hv *HouseViewer) String() string {
  return "house viewer"
}

func roomOverlapOnce(a, b *Room) bool {
  x1in := a.X+a.Size.Dx > b.X && a.X+a.Size.Dx <= b.X+b.Size.Dx
  x2in := b.X+b.Size.Dx > a.X && b.X+b.Size.Dx <= a.X+a.Size.Dx
  y1in := a.Y+a.Size.Dy > b.Y && a.Y+a.Size.Dy <= b.Y+b.Size.Dy
  y2in := b.Y+b.Size.Dy > a.Y && b.Y+b.Size.Dy <= a.Y+a.Size.Dy
  return (x1in || x2in) && (y1in || y2in)
}

func roomOverlap(a, b *Room) bool {
  return roomOverlapOnce(a, b) || roomOverlapOnce(b, a)
}

func (hv *HouseViewer) FindClosestDoorPos(door *Door, bx, by float32) *Room {
  current_floor := 0
  best := 1.0e9 // If this is unsafe then the house is larger than earth
  var best_room *Room

  clamp_int := func(n, min, max int) int {
    if n < min {
      return min
    }
    if n > max {
      return max
    }
    return n
  }
  for _, room := range hv.house.Floors[current_floor].Rooms {
    fl := math.Abs(float64(by) - float64(room.Y+room.Size.Dy))
    fr := math.Abs(float64(bx) - float64(room.X+room.Size.Dx))
    if bx < float32(room.X) {
      fl += float64(room.X) - float64(bx)
    }
    if bx > float32(room.X+room.Size.Dx) {
      fl += float64(bx) - float64(room.X+room.Size.Dx)
    }
    if by < float32(room.Y) {
      fr += float64(room.Y) - float64(by)
    }
    if by > float32(room.Y+room.Size.Dy) {
      fr += float64(by) - float64(room.Y+room.Size.Dy)
    }
    if best <= fl && best <= fr {
      continue
    }
    best_room = room
    switch {
    case fl < fr:
      best = fl
      door.Facing = FarLeft
      door.Pos = clamp_int(int(bx-float32(room.X)-float32(door.Width)/2), 0, room.Size.Dx-door.Width)

      //      case fr < fl:  this case must be true, so we just call it default here
    default:
      best = fr
      door.Facing = FarRight
      door.Pos = clamp_int(int(by-float32(room.Y)-float32(door.Width)/2), 0, room.Size.Dy-door.Width)
    }
  }
  return best_room
}

func (hv *HouseViewer) FindClosestExistingDoor(bx, by float32) (*Room, *Door) {
  current_floor := 0
  for _, room := range hv.house.Floors[current_floor].Rooms {
    for _, door := range room.Doors {
      if door.Facing != FarLeft && door.Facing != FarRight {
        continue
      }
      var vx, vy float32
      if door.Facing == FarLeft {
        vx = float32(room.X+door.Pos) + float32(door.Width)/2
        vy = float32(room.Y + room.Size.Dy)
      } else {
        // door.Facing == FarRight
        vx = float32(room.X + room.Size.Dx)
        vy = float32(room.Y+door.Pos) + float32(door.Width)/2
      }
      dsq := (vx-bx)*(vx-bx) + (vy-by)*(vy-by)
      if dsq <= float32(door.Width*door.Width) {
        return room, door
      }
    }
  }
  return nil, nil
}

type offsetDrawable struct {
  Drawable
  dx, dy int
}

func (o offsetDrawable) FPos() (float64, float64) {
  x, y := o.Drawable.FPos()
  return x + float64(o.dx), y + float64(o.dy)
}
func (o offsetDrawable) Pos() (int, int) {
  x, y := o.Drawable.Pos()
  return x + o.dx, y + o.dy
}

func (hv *HouseViewer) Draw(region gui.Region) {
  region.PushClipPlanes()
  defer region.PopClipPlanes()
  hv.Render_region = region

  hv.Floor_drawers = hv.Floor_drawers[0:0]
  if hv.Edit_mode {
    for _, spawn := range hv.house.Floors[0].Spawns {
      hv.Floor_drawers = append(hv.Floor_drawers, spawn)
    }
  }
  if hv.Floor_drawer != nil {
    hv.Floor_drawers = append(hv.Floor_drawers, hv.Floor_drawer)
  }
  hv.house.Floors[0].render(region, hv.fx, hv.fy, hv.angle, hv.zoom, hv.drawables, hv.Los_tex, hv.Floor_drawers)
}
