package house

import (
  "glop/gui"
  "github.com/arbaal/mathgl"
  "math"
  "haunts/base"
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
  gui.EmbeddedWidget
  gui.BasicZone
  gui.NonFocuser
  gui.NonResponder
  gui.NonThinker

  house *houseDef

  zoom,angle,fx,fy float32
  floor,ifloor mathgl.Mat4

  Temp struct {
    Room *Room

    // If we have a temporary door then this is the room it is attached to
    Door_room *Room
    Door_info doorInfo
  }
}

func MakeHouseViewer(house *houseDef, angle float32) *HouseViewer {
  var hv HouseViewer
  hv.EmbeddedWidget = &gui.BasicWidget{ CoreWidget: &hv }
  hv.Request_dims.Dx = 100
  hv.Request_dims.Dy = 100
  hv.Ex = true
  hv.Ey = true
  hv.house = house
  hv.angle = angle
  hv.Zoom(1)
  return &hv
}

func (hv *HouseViewer) modelviewToBoard(mx, my float32) (x,y,dist float32) {
  mz := d2p(hv.floor, mathgl.Vec3{mx, my, 0}, mathgl.Vec3{0,0,1})
  v := mathgl.Vec4{X: mx, Y: my, Z: mz, W: 1}
  v.Transform(&hv.ifloor)
  return v.X, v.Y, mz
}

func (hv *HouseViewer) WindowToBoard(wx, wy int) (float32, float32) {
  hv.floor, hv.ifloor, _, _, _, _ = makeRoomMats(&roomDef{}, hv.Render_region, hv.fx, hv.fy, hv.angle, hv.zoom)

  fx,fy,_ := hv.modelviewToBoard(float32(wx), float32(wy))
  return fx, fy
}

// Changes the current zoom from e^(zoom) to e^(zoom+dz)
func (hv *HouseViewer) Zoom(dz float64) {
  if dz == 0 {
    return
  }
  exp := math.Log(float64(hv.zoom)) + dz
  exp = float64(clamp(float32(exp), 2.5, 5.0))
  hv.zoom = float32(math.Exp(exp))
}

func (hv *HouseViewer) Drag(dx, dy float64) {
  hv.floor, hv.ifloor, _, _, _, _ = makeRoomMats(&roomDef{}, hv.Render_region, hv.fx, hv.fy, hv.angle, hv.zoom)

  v := mathgl.Vec4{X: hv.fx, Y: hv.fy, W: 1}
  v.Transform(&hv.floor)
  v.X += float32(dx)
  v.Y += float32(dy)

  v.Z = d2p(hv.floor, mathgl.Vec3{v.X, v.Y, 0}, mathgl.Vec3{0,0,1})
  v.Transform(&hv.ifloor)
  hv.fx, hv.fy = v.X, v.Y
}

func (hv *HouseViewer) String() string {
  return "house viewer"
}

func roomOverlapOnce(a,b *Room) bool {
  x1in := a.X + a.Size.Dx > b.X && a.X + a.Size.Dx <= b.X + b.Size.Dx
  x2in := b.X + b.Size.Dx > a.X && b.X + b.Size.Dx <= a.X + a.Size.Dx
  y1in := a.Y + a.Size.Dy > b.Y && a.Y + a.Size.Dy <= b.Y + b.Size.Dy
  y2in := b.Y + b.Size.Dy > a.Y && b.Y + b.Size.Dy <= a.Y + a.Size.Dy
  return (x1in || x2in) && (y1in || y2in)
}

func roomOverlap(a,b *Room) bool {
  return roomOverlapOnce(a, b) || roomOverlapOnce(b, a)
}

func (hv *HouseViewer) FindClosestDoorPos(ddef *doorDef, bx,by float32) (*Room, DoorInst) {
  current_floor := 0
  best := 1.0e9  // If this is unsafe then the house is larger than earth
  var best_room *Room
  var best_door DoorInst
  clamp_int := func(n, min, max int) int {
    if n < min { return min }
    if n > max { return max }
    return n
  }
  for _,room := range hv.house.Floors[current_floor].Rooms {
    fl := math.Abs(float64(by) - float64(room.Y + room.Size.Dy))
    fr := math.Abs(float64(bx) - float64(room.X + room.Size.Dx))
    if bx < float32(room.X) {
      fl += float64(room.X) - float64(bx)
    }
    if bx > float32(room.X + room.Size.Dx) {
      fl += float64(bx) - float64(room.X + room.Size.Dx)
    }
    if by < float32(room.Y) {
      fr += float64(room.Y) - float64(by)
    }
    if by > float32(room.Y + room.Size.Dy) {
      fr += float64(by) - float64(room.Y + room.Size.Dy)
    }
    if best <= fl && best <= fr { continue }
    best_room = room
    switch {
      case fl < fr:
        best = fl
        best_door.Facing = FarLeft
        best_door.Pos = clamp_int(int(bx - float32(room.X) - float32(ddef.Width) / 2), 0, room.Size.Dx - ddef.Width)

//      case fr < fl:  this case must be true, so we just call it default here
      default:
        best = fr
        best_door.Facing = FarRight
        best_door.Pos = clamp_int(int(by - float32(room.Y) - float32(ddef.Width) / 2), 0, room.Size.Dy - ddef.Width)
    }
  }
  return best_room, best_door
}

func (hv *HouseViewer) FindClosestExistingDoor(bx,by float32) (*Room, *Door) {
  current_floor := 0
  for _,room := range hv.house.Floors[current_floor].Rooms {
    for _,door := range room.Doors {
      if door.Facing != FarLeft && door.Facing != FarRight { continue }
      var vx,vy float32
      if door.Facing == FarLeft {
        vx = float32(room.X + door.Pos) + float32(door.Width) / 2
        vy = float32(room.Y + room.Size.Dy)
      } else {
        // door.Facing == FarRight
        vx = float32(room.X + room.Size.Dx)
        vy = float32(room.Y + door.Pos) + float32(door.Width) / 2
      }
      dsq := (vx - bx) * (vx - bx) + (vy - by) * (vy - by)
      if dsq <= float32(door.Width * door.Width) {
        return room, door
      }
    }
  }
  return nil, nil
}

func (hv *HouseViewer) Draw(region gui.Region) {
  region.PushClipPlanes()
  defer region.PopClipPlanes()
  hv.Render_region = region

  current_floor := 0

  var rooms rectObjectArray
  for _,room := range hv.house.Floors[current_floor].Rooms {
    rooms = append(rooms, room)
  }
  if hv.Temp.Room != nil {
    rooms = append(rooms, hv.Temp.Room)
  }
  rooms = rooms.Order()

  drawPrep()
  for i := len(rooms) - 1; i >= 0; i-- {
    room := rooms[i].(*Room)
    // TODO: Would be better to be able to just get the floor mats alone
    m_floor,_,m_left,_,m_right,_ := makeRoomMats(room.roomDef, region, hv.fx - float32(room.X), hv.fy - float32(room.Y), hv.angle, hv.zoom)

    var cstack base.ColorStack
    if room == hv.Temp.Room {
      valid := true
      for _,room := range hv.house.Floors[current_floor].Rooms {
        if roomOverlap(room, hv.Temp.Room) {
          valid = false
          break
        }
      }
      if valid {
        cstack.Push(0.5, 0.5, 1, 0.75)
      } else {
        cstack.Push(1.0, 0.2, 0.2, 0.9)
      }
    } else {
      cstack.Push(1, 1, 1, 1)
    }
    if room == hv.Temp.Door_room && hv.Temp.Door_info.Door != nil {
      hv.Temp.Door_info.Valid = hv.house.Floors[current_floor].canAddDoor(room, hv.Temp.Door_info.Door)
      drawWall(room, m_floor, m_left, m_right, nil, hv.Temp.Door_info, cstack)
    } else {
      drawWall(room, m_floor, m_left, m_right, nil, doorInfo{}, cstack)
    }
    drawFloor(room.roomDef, m_floor, nil, cstack)
    drawFurniture(m_floor, room.roomDef.Furniture, nil, cstack)
    // drawWall(room *roomDef, wall *texture.Data, left, right mathgl.Mat4, temp *WallTexture)
  }
}














