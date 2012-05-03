package house

import (
  "fmt"
  "math"
  "path/filepath"
  "image"
  "github.com/runningwild/glop/gin"
  "github.com/runningwild/glop/gui"
  "github.com/runningwild/glop/util/algorithm"
  "github.com/runningwild/haunts/base"
  "github.com/runningwild/haunts/texture"
)

type Room struct {
  Defname string
  *roomDef

  // The placement of doors in this room
  Doors []*Door  `registry:"loadfrom-doors"`

  // The offset of this room on this floor
  X,Y int


  // whether or not to draw the walls transparent
  far_left struct {
    wall_alpha, door_alpha byte
  }
  far_right struct {
    wall_alpha, door_alpha byte
  }

  // opengl stuff
  // Vertex buffer storing the vertices of the room as well as the texture
  // coordinates for the los texture.
  vbuffer uint32

  // index buffers
  left_buffer  uint32
  right_buffer uint32
  floor_buffer uint32

  // we don't want to redo all of the vertex and index buffers unless we
  // need to, so we keep track of the position and size of the room when they
  // were made so we don't have to.
  gl struct {
    x, y, dx, dy int
  }

  wall_texture_gl_map map[*WallTexture]wallTextureGlIds
  wall_texture_state_map map[*WallTexture]wallTextureState
}

type WallFacing int
const (
  NearLeft WallFacing = iota
  NearRight
  FarLeft
  FarRight
)

func MakeDoor(name string) *Door {
  d := Door{ Defname: name }
  base.GetObject("doors", &d)
  return &d
}

func GetAllDoorNames() []string {
  return base.GetAllNamesInRegistry("doors")
}

func LoadAllDoorsInDir(dir string) {
  base.RemoveRegistry("doors")
  base.RegisterRegistry("doors", make(map[string]*doorDef))
  base.RegisterAllObjectsInDir("doors", dir, ".json", "json")
}

func (d *Door) Load() {
  base.GetObject("doors", d)
}

type doorDef struct {
  // Name of this texture as it appears in the editor, should be unique among
  // all Doors
  Name string

  // Number of cells wide the door is
  Width int

  Opened_texture texture.Object
  Closed_texture texture.Object
}

type Door struct {
  Defname string
  *doorDef

  // Which wall the door is on
  Facing WallFacing

  // How far along this wall the door is located
  Pos int

  // Whether or not the door is opened - determines what texture to use
  Opened bool

  temporary, invalid bool
}

func (d *Door) TextureData() *texture.Data {
  if d.Opened {
    return d.Opened_texture.Data()
  }
  return d.Closed_texture.Data()
}

func (d *Door) Color() (r,g,b,a byte) {
  if d.temporary {
    if d.invalid {
      return 255, 127, 127, 200
    } else {
      return 127, 127, 255, 200
    }
  }
  return 255, 255, 255, 255
}

func (r *Room) Pos() (x,y int) {
  return r.X, r.Y
}

func getSpawnPointDefName(sp SpawnPoint) string {
  return "Unknown Spawn Type"
}

type Floor struct {
  Rooms  []*Room        `registry:"loadfrom-rooms"`
  Spawns []*SpawnPoint  `registry:"loadfrom-spawns"`
}

func (f *Floor) removeSpawn(sp *SpawnPoint) {
  f.Spawns = algorithm.Choose(f.Spawns, func(a interface{}) bool {
    return a.(*SpawnPoint) != sp
  }).([]*SpawnPoint)
}

func (f *Floor) canAddRoom(add *Room) bool {
  for _,room := range f.Rooms {
    if roomOverlap(room, add) { return false }
  }
  return true
}

func (room *Room) canAddDoor(door *Door) bool {
  if door.Pos < 0 { return false }

  // Make sure that the door only occupies valid cells
  if door.Facing == FarLeft || door.Facing == NearRight {
    if door.Pos + door.Width >= room.Size.Dx { return false }
  }
  if door.Facing == FarRight || door.Facing == NearLeft {
    if door.Pos + door.Width >= room.Size.Dy { return false }
  }

  // Now make sure that the door doesn't overlap any other doors
  for _,other := range room.Doors {
    if other.Facing != door.Facing { continue }
    if other.temporary { continue }
    if other.Pos >= door.Pos && other.Pos - door.Pos < door.Width { return false }
    if door.Pos >= other.Pos && door.Pos - other.Pos < other.Width { return false }
  }

  return true
}

func (f *Floor) findMatchingDoor(room *Room, door *Door) (*Room, *Door) {
  for _,other_room := range f.Rooms {
    if other_room == room { continue }
    for _,other_door := range other_room.Doors {
      if door.Facing == FarLeft && other_door.Facing != NearRight { continue }
      if door.Facing == FarRight && other_door.Facing != NearLeft { continue }
      if door.Facing == NearLeft && other_door.Facing != FarRight { continue }
      if door.Facing == NearRight && other_door.Facing != FarLeft { continue }
      if door.Facing == FarLeft && other_room.Y != room.Y + room.Size.Dy { continue }
      if door.Facing == NearRight && room.Y != other_room.Y + other_room.Size.Dy { continue }
      if door.Facing == FarRight && other_room.X != room.X + room.Size.Dx { continue }
      if door.Facing == NearLeft && room.X != other_room.X + other_room.Size.Dx { continue }
      if door.Facing == FarLeft || door.Facing == NearRight {
        if door.Pos == other_door.Pos - (room.X - other_room.X) {
          return other_room, other_door
        }
      }
      if door.Facing == FarRight || door.Facing == NearLeft {
        if door.Pos == other_door.Pos - (room.Y - other_room.Y) {
          return other_room, other_door
        }
      }
    }
  }
  return nil, nil
}

func (f *Floor) findRoomForDoor(target *Room, door *Door) (*Room, *Door) {
  if !target.canAddDoor(door) { return nil, nil }

  if door.Facing == FarLeft {
    for _,room := range f.Rooms {
      if room.Y == target.Y + target.Size.Dy {
        temp := MakeDoor(door.Defname)
        temp.Pos = door.Pos - (room.X - target.X)
        temp.Facing = NearRight
        if room.canAddDoor(temp) {
          return room, temp
        }
      }
    }
  } else if door.Facing == FarRight {
    for _,room := range f.Rooms {
      if room.X == target.X + target.Size.Dx {
        temp := MakeDoor(door.Defname)
        temp.Pos = door.Pos - (room.Y - target.Y)
        temp.Facing = NearLeft
        if room.canAddDoor(temp) {
          return room, temp
        }
      }
    }
  }
  return nil, nil
}

func (f *Floor) canAddDoor(target *Room, door *Door) bool {
  r,_ := f.findRoomForDoor(target, door)
  return r != nil
}

func (f *Floor) removeInvalidDoors() {
  for _,room := range f.Rooms {
    room.Doors = algorithm.Choose(room.Doors, func(a interface{}) bool {
      _, other_door := f.findMatchingDoor(room, a.(*Door))
      return other_door != nil && !other_door.temporary
    }).([]*Door)
  }
}

func (f *Floor) RoomFurnSpawnAtPos(x, y int) (room_def *roomDef, furn, spawn bool) {
  for _, room := range f.Rooms {
    rx,ry := room.Pos()
    rdx,rdy := room.Dims()
    if x < rx || y < ry || x >= rx + rdx || y >= ry + rdy { continue }
    room_def = room.roomDef
    for _, furniture := range room.Furniture {
      tx := x - rx
      ty := y - ry
      fx,fy := furniture.Pos()
      fdx,fdy := furniture.Dims()
      if tx < fx || ty < fy || tx >= fx + fdx || ty >= fy + fdy { continue }
      furn = true
      break
    }
    for _, sp := range f.Spawns {
      sx, sy := sp.Pos()
      sdx, sdy := sp.Dims()
      if x >= sx && x < sx + sdx && y >= sy && y < sy + sdy {
        spawn = true
        break
      }
    }
    return
  }
  return
}

func (f *Floor) render(region gui.Region, focusx,focusy,angle,zoom float32, drawables []Drawable, los_tex *LosTexture, floor_drawer FloorDrawer) {
  var ros []RectObject
  algorithm.Map2(f.Rooms, &ros, func(r *Room) RectObject { return r })
  ros = OrderRectObjects(ros)
  alpha_map := make(map[*Room]byte)

  // First pass over the rooms - this will determine at what alpha the rooms
  // should be draw.  We will use this data later to determine the alpha for
  // the doors of adjacent rooms.
  for i := len(ros) - 1; i >= 0; i-- {
    room := ros[i].(*Room)
    los_alpha := room.getMaxLosAlpha(los_tex)
    room.setupGlStuff()
    tx := (focusx + 3) - float32(room.X + room.Size.Dx)
    if tx < 0 { tx = 0 }
    ty := (focusy + 3) - float32(room.Y + room.Size.Dy)
    if ty < 0 { ty = 0 }
    if tx < ty {
      tx = ty
    }
    // z := math.Log10(float64(zoom))
    z := float64(zoom) / 10
    v := math.Pow(z, float64(2 * tx) / 3)
    if v > 255 {
      v = 255
    }
    bv := 255 - byte(v)
    alpha_map[room] = byte((int(bv) * int(los_alpha)) >> 8)
    // room.render(floor, left, right, , 255)
  }

  // Second pass - this time we fill in the alpha that we should use for the
  // doors, using the values we've already calculated in the first pass.
  for _, r1 := range f.Rooms {
    // It is possible that we get two doors to different rooms on one wall,
    // and we might display them with the same alpha even though the rooms
    // they are attached to have different alpha.  This is probably not a
    // big deal so we'll just ignore it.
    for _, door := range r1.Doors {
      r2, _ := f.findMatchingDoor(r1, door)
      if r2 == nil { continue }
      // alpha := alpha_map[r2]
      // base.Log().Printf("%p %p %d %d", r1, r2, alpha, door.Facing)
      left, right := r2.getNearWallAlpha(los_tex)
      switch door.Facing {
      case FarLeft:
        // if left > alpha {
          r1.far_left.door_alpha = left
        // } else {
        //   r1.far_left.door_alpha = alpha
        // }
      case FarRight:
        // if right > alpha {
          r1.far_right.door_alpha = right
        // } else {
        //   r1.far_right.door_alpha = alpha
        // }
      }
    }

    r1.far_right.wall_alpha = 255
    r1.far_left.wall_alpha = 255
    for _, r2 := range f.Rooms {
      if r1 == r2 { continue }
      left, right := r2.getNearWallAlpha(los_tex)
      r1_rect := image.Rect(r1.X, r1.Y + r1.Size.Dy, r1.X + r1.Size.Dx, r1.Y + r1.Size.Dy + 1)
      r2_rect := image.Rect(r2.X, r2.Y, r2.X + r2.Size.Dx, r2.Y + r2.Size.Dy)
      if r1_rect.Overlaps(r2_rect) {
        r1.far_left.wall_alpha = byte((int(left) * 200) >> 8)
      }
      r1_rect = image.Rect(r1.X + r1.Size.Dx, r1.Y, r1.X + r1.Size.Dx + 1, r1.Y + r1.Size.Dy)
      if r1_rect.Overlaps(r2_rect) {
        r1.far_right.wall_alpha = byte((int(right) * 200) >> 8)
      }
    }
  }

  // Third pass - now that we know what alpha to use on the rooms, walls, and
  // doors we can actually render everything.  We still need to go back to
  // front though.
  for i := len(ros) - 1; i >= 0; i-- {
    room := ros[i].(*Room)
    fx := focusx - float32(room.X)
    fy := focusy - float32(room.Y)
    floor, _, left, _, right, _ := makeRoomMats(room.roomDef, region, fx, fy, angle, zoom)
    v := alpha_map[room]
    room.render(floor, left, right, zoom, v, drawables, los_tex, floor_drawer)
  }
}

type HouseDef struct {
  Name string

  Floors []*Floor
}

func MakeHouseDef() *HouseDef {
  var h HouseDef
  h.Name = "name"
  h.Floors = append(h.Floors, &Floor{})
  return &h
}

// Shifts the rooms in all floors such that the coordinates of all rooms are
// as low on each axis as possible without being zero or negative.
func (h *HouseDef) Normalize() {
  for i := range h.Floors {
    if len(h.Floors[i].Rooms) == 0 {
      continue
    }
    var minx,miny int
    minx,miny = h.Floors[i].Rooms[0].Pos()
    for j := range h.Floors[i].Rooms {
      x,y := h.Floors[i].Rooms[j].Pos()
      if x < minx {
        minx = x
      }
      if y < miny {
        miny = y
      }
    }
    for j := range h.Floors[i].Rooms {
      h.Floors[i].Rooms[j].X -= minx - 1
      h.Floors[i].Rooms[j].Y -= miny - 1
    }
    for _, sp := range h.Floors[0].Spawns {
      sp.X -= minx - 1
      sp.Y -= miny - 1
    }
  }
}

type HouseEditor struct {
  *gui.HorizontalTable
  tab *gui.TabFrame
  widgets []tabWidget

  house  HouseDef
  viewer *HouseViewer
}

func (he *HouseEditor) GetViewer() Viewer {
  return he.viewer
}

func (w *HouseEditor) SelectTab(n int) {
  if n < 0 || n >= len(w.widgets) { return }
  if n != w.tab.SelectedTab() {
    w.widgets[w.tab.SelectedTab()].Collapse()
    w.tab.SelectTab(n)
    // w.viewer.SetEditMode(editNothing)
    w.widgets[n].Expand()
  }
}

type houseDataTab struct {
  *gui.VerticalTable

  name       *gui.TextEditLine
  num_floors *gui.ComboBox

  house  *HouseDef
  viewer *HouseViewer

  // Distance from the mouse to the center of the object, in board coordinates
  drag_anchor struct{ x,y float32 }

  // Which floor we are viewing and editing
  current_floor int
}
func makeHouseDataTab(house *HouseDef, viewer *HouseViewer) *houseDataTab {
  var hdt houseDataTab
  hdt.VerticalTable = gui.MakeVerticalTable()
  hdt.house = house
  hdt.viewer = viewer

  hdt.name = gui.MakeTextEditLine("standard", "name", 300, 1, 1, 1, 1)
  num_floors_options := []string{ "1 Floor", "2 Floors", "3 Floors", "4 Floors" }
  hdt.num_floors = gui.MakeComboTextBox(num_floors_options, 300)

  hdt.VerticalTable.AddChild(hdt.name)
  hdt.VerticalTable.AddChild(hdt.num_floors)
  
  names := GetAllRoomNames()
  for _,name := range names {
    n := name
    hdt.VerticalTable.AddChild(gui.MakeButton("standard", name, 300, 1, 1, 1, 1, func(int64) {
      hdt.viewer.Temp.Room = &Room{ Defname: n }
      fmt.Printf("room: %v\n", hdt.viewer.Temp.Room)
      base.GetObject("rooms", hdt.viewer.Temp.Room)
      hdt.drag_anchor.x = float32(hdt.viewer.Temp.Room.Size.Dx / 2)
      hdt.drag_anchor.y = float32(hdt.viewer.Temp.Room.Size.Dy / 2)
    }))
  }
  return &hdt
}
func (hdt *houseDataTab) Think(ui *gui.Gui, t int64) {
  if hdt.viewer.Temp.Room != nil {
    mx,my := gin.In().GetCursor("Mouse").Point()
    bx,by := hdt.viewer.WindowToBoard(mx, my)
    hdt.viewer.Temp.Room.X = int(bx - hdt.drag_anchor.x)
    hdt.viewer.Temp.Room.Y = int(by - hdt.drag_anchor.y)
  }
  hdt.VerticalTable.Think(ui, t)
  num_floors := hdt.num_floors.GetComboedIndex() + 1
  if len(hdt.house.Floors) != num_floors {
    for len(hdt.house.Floors) < num_floors {
      hdt.house.Floors = append(hdt.house.Floors, &Floor{})
    }
    if len(hdt.house.Floors) > num_floors {
      hdt.house.Floors = hdt.house.Floors[0 : num_floors]
    }
  }
  hdt.house.Name = hdt.name.GetText()
}
func (hdt *houseDataTab) Respond(ui *gui.Gui, group gui.EventGroup) bool {
  if hdt.VerticalTable.Respond(ui, group) {
    return true
  }

  if found,event := group.FindEvent(gin.Escape); found && event.Type == gin.Press {
    hdt.viewer.Temp.Room = nil
    return true
  }

  floor := hdt.house.Floors[hdt.current_floor]
  if found,event := group.FindEvent(gin.MouseLButton); found && event.Type == gin.Press {
    if hdt.viewer.Temp.Room != nil {
      if floor.canAddRoom(hdt.viewer.Temp.Room) {
        floor.Rooms = append(floor.Rooms, hdt.viewer.Temp.Room)
        hdt.viewer.Temp.Room = nil
        floor.removeInvalidDoors()
      }
    } else {
      bx,by := hdt.viewer.WindowToBoard(event.Key.Cursor().Point())
      for i := range floor.Rooms {
        x,y := floor.Rooms[i].Pos()
        dx,dy := floor.Rooms[i].Dims()
        if int(bx) >= x && int(bx) < x + dx && int(by) >= y && int(by) < y + dy {
          hdt.viewer.Temp.Room = floor.Rooms[i]
          floor.Rooms[i] = floor.Rooms[len(floor.Rooms) - 1]
          floor.Rooms = floor.Rooms[0 : len(floor.Rooms) - 1]
          break
        }
      }
      if hdt.viewer.Temp.Room != nil {
        px,py := hdt.viewer.Temp.Room.Pos()
        hdt.drag_anchor.x = bx - float32(px)
        hdt.drag_anchor.y = by - float32(py)
      }
    }
    return true
  }

  return false
}
func (hdt *houseDataTab) Collapse() {}
func (hdt *houseDataTab) Expand() {}
func (hdt *houseDataTab) Reload() {
  hdt.name.SetText(hdt.house.Name)
}

type houseDoorTab struct {
  *gui.VerticalTable

  num_floors *gui.ComboBox

  house  *HouseDef
  viewer *HouseViewer

  // Distance from the mouse to the center of the object, in board coordinates
  drag_anchor struct{ x,y float32 }

  // Which floor we are viewing and editing
  current_floor int

  // board pos of the cursor
  bx, by float32

  room, prev_room *Room
  door, prev_door *Door
}
func makeHouseDoorTab(house *HouseDef, viewer *HouseViewer) *houseDoorTab {
  var hdt houseDoorTab
  hdt.VerticalTable = gui.MakeVerticalTable()
  hdt.house = house
  hdt.viewer = viewer

  names := GetAllDoorNames()
  for _,name := range names {
    n := name
    hdt.VerticalTable.AddChild(gui.MakeButton("standard", name, 300, 1, 1, 1, 1, func(int64) {
      if len(hdt.house.Floors[0].Rooms) < 2 {
        return
      }
      hdt.door = MakeDoor(n)
      hdt.door.temporary = true
      hdt.door.invalid = true
      hdt.room = hdt.house.Floors[0].Rooms[0]
    }))
  }

  return &hdt
}
func (hdt *houseDoorTab) Think(ui *gui.Gui, t int64) {
}
func (hdt *houseDoorTab) onEscape() {
  if hdt.door != nil {
    if hdt.room != nil {
      algorithm.Choose2(&hdt.room.Doors, func(d *Door) bool {
        return d != hdt.door
      })
    }
    if hdt.prev_door != nil {
      hdt.prev_room.Doors = append(hdt.prev_room.Doors, hdt.prev_door)
      hdt.prev_door = nil
      hdt.prev_room = nil
    }
    hdt.door = nil
    hdt.room = nil
  }
}
func (hdt *houseDoorTab) Respond(ui *gui.Gui, group gui.EventGroup) bool {
  if hdt.VerticalTable.Respond(ui, group) {
    return true
  }

  if found,event := group.FindEvent(gin.Escape); found && event.Type == gin.Press {
    hdt.onEscape()
    return true
  }

  cursor := group.Events[0].Key.Cursor()
  if cursor != nil {
    hdt.bx, hdt.by = hdt.viewer.WindowToBoard(cursor.Point())
  }
  if cursor != nil && hdt.door != nil {
    room := hdt.viewer.FindClosestDoorPos(hdt.door, hdt.bx, hdt.by)
    if room != hdt.room {
      algorithm.Choose2(&hdt.room.Doors, func(d *Door) bool {
        return d != hdt.door
      })
      hdt.room = room
      hdt.door.invalid = (hdt.room == nil)
      hdt.room.Doors = append(hdt.room.Doors, hdt.door)
    }
    if hdt.room == nil {
      hdt.door.invalid = true
    } else {
      other_room, _ := hdt.house.Floors[0].findRoomForDoor(hdt.room, hdt.door)
      hdt.door.invalid = (other_room == nil)
    }
  }

  floor := hdt.house.Floors[hdt.current_floor]
  if found,event := group.FindEvent(gin.MouseLButton); found && event.Type == gin.Press {
    if hdt.door != nil {
      other_room, other_door := floor.findRoomForDoor(hdt.room, hdt.door)
      if other_room != nil {
        other_room.Doors = append(other_room.Doors, other_door)
        hdt.door.temporary = false
        hdt.door = nil
        hdt.prev_door = nil
      }
    } else {
      hdt.room, hdt.door = hdt.viewer.FindClosestExistingDoor(hdt.bx, hdt.by)
      if hdt.door != nil {
        hdt.prev_door = new(Door)
        *hdt.prev_door = *hdt.door
        hdt.prev_room = hdt.room
        hdt.door.temporary = true
        room, door := hdt.house.Floors[0].findMatchingDoor(hdt.room, hdt.door)
        if room != nil {
          algorithm.Choose2(&room.Doors, func(d *Door) bool {
            return d != door
          })
        }
      }
    }
    return true
  }

  return false
}
func (hdt *houseDoorTab) Collapse() {
  hdt.onEscape()
}
func (hdt *houseDoorTab) Expand() {}
func (hdt *houseDoorTab) Reload() {
  hdt.onEscape()
}

type houseRelicsTab struct {
  *gui.VerticalTable

  num_floors *gui.ComboBox
  spawn_name *gui.TextLine
  spawn_type *gui.ComboBox
  spawn_dims *gui.ComboBox

  house  *HouseDef
  viewer *HouseViewer

  // Which floor we are viewing and editing
  current_floor int

  drag_anchor struct{x, y float32}
}
func makeHouseRelicsTab(house *HouseDef, viewer *HouseViewer) *houseRelicsTab {
  var hdt houseRelicsTab
  hdt.VerticalTable = gui.MakeVerticalTable()
  hdt.house = house
  hdt.viewer = viewer

  hdt.VerticalTable.AddChild(gui.MakeTextLine("standard", "Spawns", 300, 1, 1, 1, 1))
  hdt.spawn_name = gui.MakeTextLine("standard", "", 300, 1, 1, 1, 1)
  hdt.VerticalTable.AddChild(hdt.spawn_name)


  var dims_options []string
  for i := 1; i <= 4; i++ {
    for j := 1; j <= 4; j++ {
      dims_options = append(dims_options, fmt.Sprintf("%dx%d", i, j))
    }
  }
  hdt.spawn_dims = gui.MakeComboTextBox(dims_options, 300)

  hdt.VerticalTable.AddChild(hdt.spawn_dims)
  spawn_names := GetAllSpawnPointNames()
  for _, spawn_name := range spawn_names {
    name := spawn_name
    hdt.VerticalTable.AddChild(gui.MakeButton("standard", name, 300, 1, 1, 1, 1, func(int64) {
      var dx, dy int
      dims_str := dims_options[hdt.spawn_dims.GetComboedIndex()]
      fmt.Sscanf(dims_str, "%dx%d", &dx, &dy)
      hdt.viewer.Temp.Spawn = MakeSpawnPoint(name)
      hdt.viewer.Temp.Spawn.X = 100000
      hdt.viewer.Temp.Spawn.Dx = dx
      hdt.viewer.Temp.Spawn.Dy = dy
      // hdt.viewer.Temp.Spawn.Type = spawn_options[hdt.spawn_type.GetComboedOption().(string)]
    }))
  }

  return &hdt
}

func (hdt *houseRelicsTab) Think(ui *gui.Gui, t int64) {
  defer hdt.VerticalTable.Think(ui, t)
  rbx,rby := hdt.viewer.WindowToBoard(gin.In().GetCursor("Mouse").Point())
  bx := roundDown(rbx - hdt.drag_anchor.x)
  by := roundDown(rby - hdt.drag_anchor.y)
  if hdt.viewer.Temp.Spawn != nil {
    hdt.spawn_name.SetText("Monkey cake")
    hdt.viewer.Temp.Spawn.X = bx
    hdt.viewer.Temp.Spawn.Y = by
  } else {
    set := false
    for _, sp := range hdt.house.Floors[0].Spawns {
      x,y := sp.Pos()
      dx,dy := sp.Dims()
      if bx >= x && bx < x + dx && by >= y && by < y + dy {
        hdt.spawn_name.SetText("Hoogabooo")
        set = true
        break
      }
    }
    if !set {
      hdt.spawn_name.SetText("")
    }
  }
}

// Rounds a float32 down, instead of towards zero
func roundDown(f float32) int {
  if f >= 0 {
    return int(f)
  }
  return int(f-1)
}

func (hdt *houseRelicsTab) Respond(ui *gui.Gui, group gui.EventGroup) bool {
  if hdt.VerticalTable.Respond(ui, group) {
    return true
  }

  if found,event := group.FindEvent(gin.Escape); found && event.Type == gin.Press {
    hdt.viewer.Temp.Spawn = nil
    return true
  }

  cursor := group.Events[0].Key.Cursor()
  var rbx, rby float32
  var bx, by int
  if cursor != nil {
    rbx, rby = hdt.viewer.WindowToBoard(cursor.Point())
    bx = roundDown(rbx - hdt.drag_anchor.x)
    by = roundDown(rby - hdt.drag_anchor.y)
  } 
  floor := hdt.house.Floors[hdt.current_floor]
  if found,event := group.FindEvent(gin.MouseLButton); found && event.Type == gin.Press {
    if hdt.viewer.Temp.Spawn != nil {
      x := hdt.viewer.Temp.Spawn.X
      y := hdt.viewer.Temp.Spawn.Y
      ok := true
      var room_def *roomDef
      for ix := 0; ix < hdt.viewer.Temp.Spawn.Dx; ix++ {
        for iy := 0; iy < hdt.viewer.Temp.Spawn.Dy; iy++ {
          room_at, furn_at, spawn_at := floor.RoomFurnSpawnAtPos(x + ix, y + iy)
          if room_at == nil || (room_def != nil && room_at != room_def) || furn_at || spawn_at {
            ok = false
          } else {
            room_def = room_at
          }
        }
      }
      if ok {
        floor.Spawns = append(floor.Spawns, hdt.viewer.Temp.Spawn)
        hdt.viewer.Temp.Spawn = nil
        hdt.drag_anchor.x = 0
        hdt.drag_anchor.y = 0
      }
    } else {
      for _, sp := range floor.Spawns {
        x, y := sp.Pos()
        dx, dy := sp.Dims()
        if bx >= x && bx < x + dx && by >= y && by < y + dy {
          hdt.viewer.Temp.Spawn = sp
          hdt.drag_anchor.x = float32(bx) - float32(hdt.viewer.Temp.Spawn.X)
          hdt.drag_anchor.y = float32(by) - float32(hdt.viewer.Temp.Spawn.Y)
          floor.removeSpawn(sp)
          break
        }
      }
    }
  }
  return false
}
func (hdt *houseRelicsTab) Collapse() {}
func (hdt *houseRelicsTab) Expand() {}
func (hdt *houseRelicsTab) Reload() {}

func (h *HouseDef) Save(path string) {
  base.SaveJson(path, h)
}

func LoadAllHousesInDir(dir string) {
  base.RemoveRegistry("houses")
  base.RegisterRegistry("houses", make(map[string]*HouseDef))
  base.RegisterAllObjectsInDir("houses", dir, ".house", "json")
}

func MakeHouseFromPath(path string) (*HouseDef, error) {
  var house HouseDef
  err := base.LoadAndProcessObject(path, "json", &house)
  if err != nil {
    return nil, err
  }
  for _,floor := range house.Floors {
    for _,room := range floor.Rooms {
      for _,door := range room.Doors {
        door.Opened = true
      }
    }
  }
  //   if house.Floors[i].Spawns == nil {
  //     house.Floors[i].Spawns = make(map[string][]*Furniture)
  //   } else {
  //     for _, spawns := range house.Floors[i].Spawns {
  //       for _, spawn := range spawns {
  //         spawn.Load()
  //       }
  //     }
  //   }
  // }
  return &house, nil
}

func MakeHouseEditorPanel() Editor {
  var he HouseEditor
  he.house = *MakeHouseDef()
  he.HorizontalTable = gui.MakeHorizontalTable()
  he.viewer = MakeHouseViewer(&he.house, 62)
  he.HorizontalTable.AddChild(he.viewer)

  he.widgets = append(he.widgets, makeHouseDataTab(&he.house, he.viewer))
  he.widgets = append(he.widgets, makeHouseDoorTab(&he.house, he.viewer))
  he.widgets = append(he.widgets, makeHouseRelicsTab(&he.house, he.viewer))
  var tabs []gui.Widget
  for _,w := range he.widgets {
    tabs = append(tabs, w.(gui.Widget))
  }
  he.tab = gui.MakeTabFrame(tabs)
  he.HorizontalTable.AddChild(he.tab)

  return &he
}

// Manually pass all events to the tabs, regardless of location, since the tabs
// need to know where the user clicks.
func (he *HouseEditor) Respond(ui *gui.Gui, group gui.EventGroup) bool {
  he.viewer.Respond(ui, group)
  return he.widgets[he.tab.SelectedTab()].Respond(ui, group)
}

func (he *HouseEditor) Load(path string) error {
  house, err := MakeHouseFromPath(path)
  if err != nil {
    return err
  }
  he.house = *house
  for _,tab := range he.widgets {
    tab.Reload()
  }
  return err
}

func (he *HouseEditor) Save() (string, error) {
  path := filepath.Join(datadir, "houses", he.house.Name + ".house")
  err := base.SaveJson(path, he.house)
  return path, err
}

func (he *HouseEditor) Reload() {
  for _,floor := range he.house.Floors {
    for i := range floor.Rooms {
      base.GetObject("rooms", floor.Rooms[i])
    }
  }
  for _,tab := range he.widgets {
    tab.Reload()
  }
}
