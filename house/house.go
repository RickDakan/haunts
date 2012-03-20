package house

import (
  "fmt"
  "path/filepath"
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
}

func (d *Door) TextureData() *texture.Data {
  if d.Opened {
    return d.Opened_texture.Data()
  }
  return d.Closed_texture.Data()
}

func (r *Room) Pos() (x,y int) {
  return r.X, r.Y
}

type SpawnPoint interface {
  Furniture() *Furniture
}
func getSpawnPointDefName(sp SpawnPoint) string {
  switch s := sp.(type) {
  case *Relic:
    return s.Defname
  case *Clue:
    return s.Defname
  case *Exit:
    return s.Defname
  case *Explorer:
    return s.Defname
  case *Haunt:
    return s.Defname
  }
  return "Unknown Spawn Type"
}

type Floor struct {
  Rooms []*Room  `registry:"loadfrom-rooms"`

  Relics    []*Relic     `registry:"loadfrom-relic"`
  Clues     []*Clue      `registry:"loadfrom-clue"`
  Exits     []*Exit      `registry:"loadfrom-exit"`
  Explorers []*Explorer  `registry:"loadfrom-explorer"`
  Haunts    []*Haunt     `registry:"loadfrom-haunt"`
}

func (f *Floor) allSpawns() []SpawnPoint {
  var sps []SpawnPoint
  for _, sp := range f.Relics {
    sps = append(sps, sp)
  }
  for _, sp := range f.Clues {
    sps = append(sps, sp)
  }
  for _, sp := range f.Exits {
    sps = append(sps, sp)
  }
  for _, sp := range f.Explorers {
    sps = append(sps, sp)
  }
  for _, sp := range f.Haunts {
    sps = append(sps, sp)
  }
  return sps
}

func (f *Floor) removeSpawn(sp SpawnPoint) {
  chooser := func(a interface{}) bool {
    return a.(SpawnPoint) != sp
  }
  switch sp.(type) {
  case *Relic:
    f.Relics = algorithm.Choose(f.Relics, chooser).([]*Relic)
  case *Clue:
    f.Clues = algorithm.Choose(f.Clues, chooser).([]*Clue)
  case *Exit:
    f.Exits = algorithm.Choose(f.Exits, chooser).([]*Exit)
  case *Explorer:
    f.Explorers = algorithm.Choose(f.Explorers, chooser).([]*Explorer)
  case *Haunt:
    f.Haunts = algorithm.Choose(f.Haunts, chooser).([]*Haunt)
  }
}

func (f *Floor) canAddRoom(add *Room) bool {
  for _,room := range f.Rooms {
    if roomOverlap(room, add) { return false }
  }
  return true
}

func (room *Room) canAddDoor(door *Door) bool {
  if door.Pos < 0 {
    return false
  }

  // Make sure that the door only occupies valid cells and only cells that have
  // CanHaveDoor set to true
  if door.Facing == FarLeft || door.Facing == NearRight {
    if door.Pos + door.Width >= room.Size.Dx { return false }
    y := 0
    if door.Facing == FarLeft {
      y = room.Size.Dy - 1
    }
    for pos := door.Pos; pos <= door.Pos + door.Width; pos++ {
      if !room.Cell_data[pos][y].CanHaveDoor { return false }
    }
  }
  if door.Facing == FarRight || door.Facing == NearLeft {
    if door.Pos + door.Width >= room.Size.Dy { return false }
    x := 0
    if door.Facing == FarRight {
      x = room.Size.Dx - 1
    }
    for pos := door.Pos; pos <= door.Pos + door.Width; pos++ {
      if !room.Cell_data[x][pos].CanHaveDoor { return false }
    }
  }

  // Now make sure that the door doesn't overlap any other doors
  for _,other := range room.Doors {
    if other.Facing != door.Facing { continue }
    if other.Pos >= door.Pos && other.Pos - door.Pos < door.Width { return false }
    if door.Pos >= other.Pos && door.Pos - other.Pos < other.Width { return false }
  }

  return true
}

func (f *Floor) findMatchingDoor(room *Room, door *Door) *Door {
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
          return other_door
        }
      }
      if door.Facing == FarRight || door.Facing == NearLeft {
        if door.Pos == other_door.Pos - (room.Y - other_room.Y) {
          return other_door
        }
      }
    }
  }
  return nil
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
      return f.findMatchingDoor(room, a.(*Door)) != nil
    }).([]*Door)
  }
}

func (f *Floor) RoomAndFurnAtPos(x, y int) (*roomDef, *furnitureDef) {
  for _, room := range f.Rooms {
    rx,ry := room.Pos()
    rdx,rdy := room.Dims()
    if x < rx || y < ry || x >= rx + rdx || y >= ry + rdy { continue }
    for _, furn := range room.Furniture {
      tx := x - rx
      ty := y - ry
      fx,fy := furn.Pos()
      fdx,fdy := furn.Dims()
      if tx < fx || ty < fy || tx >= fx + fdx || ty >= fy + fdy { continue }
      return room.roomDef, furn.furnitureDef
    }
    for _, spawn := range f.allSpawns() {
      sx,sy := spawn.Furniture().Pos()
      if sx == x && sy == y {
        return room.roomDef, spawn.Furniture().furnitureDef
      }
    }
    return room.roomDef, nil
  }
  return nil, nil
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
    for _, sp := range h.Floors[0].allSpawns() {
      sp.Furniture().X -= minx - 1
      sp.Furniture().Y -= miny - 1
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
func (hdt *houseDataTab) Reload() {}

type houseDoorTab struct {
  *gui.VerticalTable

  num_floors *gui.ComboBox

  house  *HouseDef
  viewer *HouseViewer

  // Distance from the mouse to the center of the object, in board coordinates
  drag_anchor struct{ x,y float32 }

  // Which floor we are viewing and editing
  current_floor int
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
      hdt.viewer.Temp.Door_info.Door = MakeDoor(n)
    }))
  }

  return &hdt
}
func (hdt *houseDoorTab) Think(ui *gui.Gui, t int64) {
}
func (hdt *houseDoorTab) Respond(ui *gui.Gui, group gui.EventGroup) bool {
  if hdt.VerticalTable.Respond(ui, group) {
    return true
  }

  if found,event := group.FindEvent(gin.Escape); found && event.Type == gin.Press {
    hdt.viewer.Temp.Door_info.Door = nil
    return true
  }

  cursor := group.Events[0].Key.Cursor()
  if cursor != nil && hdt.viewer.Temp.Door_info.Door != nil {
    bx,by := hdt.viewer.WindowToBoard(cursor.Point())
    room, door_inst := hdt.viewer.FindClosestDoorPos(hdt.viewer.Temp.Door_info.Door.doorDef, bx, by)
    hdt.viewer.Temp.Door_room = room
    hdt.viewer.Temp.Door_info.Door = &door_inst
  }

  floor := hdt.house.Floors[hdt.current_floor]
  if found,event := group.FindEvent(gin.MouseLButton); found && event.Type == gin.Press {
    if hdt.viewer.Temp.Door_info.Door != nil {
      other_room, other_door := floor.findRoomForDoor(hdt.viewer.Temp.Door_room, hdt.viewer.Temp.Door_info.Door)
      if other_room != nil {
        other_room.Doors = append(other_room.Doors, other_door)
        hdt.viewer.Temp.Door_room.Doors = append(hdt.viewer.Temp.Door_room.Doors, hdt.viewer.Temp.Door_info.Door)
        hdt.viewer.Temp.Door_room = nil
        hdt.viewer.Temp.Door_info.Door = nil
      }
    } else {
      bx,by := hdt.viewer.WindowToBoard(cursor.Point())
      r,d := hdt.viewer.FindClosestExistingDoor(bx, by)
      if r != nil {
        r.Doors = algorithm.Choose(r.Doors, func(a interface{}) bool {
          return a.(*Door) != d
        }).([]*Door)
        hdt.viewer.Temp.Door_room = r
        hdt.viewer.Temp.Door_info.Door = d
        floor.removeInvalidDoors()
      }
    }
    return true
  }

  return false
}
func (hdt *houseDoorTab) Collapse() {}
func (hdt *houseDoorTab) Expand() {}
func (hdt *houseDoorTab) Reload() {}

type houseRelicsTab struct {
  *gui.VerticalTable

  num_floors *gui.ComboBox
  spawn_name *gui.TextLine

  house  *HouseDef
  viewer *HouseViewer

  // Which floor we are viewing and editing
  current_floor int
}
func makeHouseRelicsTab(house *HouseDef, viewer *HouseViewer) *houseRelicsTab {
  var hdt houseRelicsTab
  hdt.VerticalTable = gui.MakeVerticalTable()
  hdt.house = house
  hdt.viewer = viewer

  hdt.VerticalTable.AddChild(gui.MakeTextLine("standard", "Spawns", 300, 1, 1, 1, 1))
  hdt.spawn_name = gui.MakeTextLine("standard", "", 300, 1, 1, 1, 1)
  hdt.VerticalTable.AddChild(hdt.spawn_name)

  for _, spawn_name := range GetAllRelicNames() {
    name := spawn_name
    hdt.VerticalTable.AddChild(gui.MakeButton("standard", name, 300, 1, 1, 1, 1, func(int64) {
        hdt.viewer.Temp.Spawn = MakeRelic(name)
        // Pushes it way off screen - as soon as the mouse is moved it will
        // move to the mouse, and since the mouse will be off the map at the
        // time it will look fine.
        hdt.viewer.Temp.Spawn.Furniture().X = 100000
      }))
  }

  for _, spawn_name := range GetAllClueNames() {
    name := spawn_name
    hdt.VerticalTable.AddChild(gui.MakeButton("standard", name, 300, 1, 1, 1, 1, func(int64) {
        hdt.viewer.Temp.Spawn = MakeClue(name)
        hdt.viewer.Temp.Spawn.Furniture().X = 100000
      }))
  }

  for _, spawn_name := range GetAllExitNames() {
    name := spawn_name
    hdt.VerticalTable.AddChild(gui.MakeButton("standard", name, 300, 1, 1, 1, 1, func(int64) {
        hdt.viewer.Temp.Spawn = MakeExit(name)
        hdt.viewer.Temp.Spawn.Furniture().X = 100000
      }))
  }

  for _, spawn_name := range GetAllExplorerNames() {
    name := spawn_name
    hdt.VerticalTable.AddChild(gui.MakeButton("standard", name, 300, 1, 1, 1, 1, func(int64) {
        hdt.viewer.Temp.Spawn = MakeExplorer(name)
        hdt.viewer.Temp.Spawn.Furniture().X = 100000
      }))
  }

  for _, spawn_name := range GetAllHauntNames() {
    name := spawn_name
    hdt.VerticalTable.AddChild(gui.MakeButton("standard", name, 300, 1, 1, 1, 1, func(int64) {
        hdt.viewer.Temp.Spawn = MakeHaunt(name)
        hdt.viewer.Temp.Spawn.Furniture().X = 100000
      }))
  }

  return &hdt
}

func (hdt *houseRelicsTab) Think(ui *gui.Gui, t int64) {
  defer hdt.VerticalTable.Think(ui, t)
  rbx,rby := hdt.viewer.WindowToBoard(gin.In().GetCursor("Mouse").Point())
  bx := roundDown(rbx)
  by := roundDown(rby)
  if hdt.viewer.Temp.Spawn != nil {
    hdt.spawn_name.SetText(getSpawnPointDefName(hdt.viewer.Temp.Spawn))
  } else {
    set := false
    for _, sp := range hdt.house.Floors[0].allSpawns() {
      x,y := sp.Furniture().Pos()
      if x == bx && y == by {
        hdt.spawn_name.SetText(getSpawnPointDefName(sp))
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
    bx = roundDown(rbx)
    by = roundDown(rby)
  } 
  if cursor != nil && hdt.viewer.Temp.Spawn != nil {
    hdt.viewer.Temp.Spawn.Furniture().X = bx
    hdt.viewer.Temp.Spawn.Furniture().Y = by
  }
  floor := hdt.house.Floors[hdt.current_floor]
  if found,event := group.FindEvent(gin.MouseLButton); found && event.Type == gin.Press {
    if hdt.viewer.Temp.Spawn != nil {
      x := hdt.viewer.Temp.Spawn.Furniture().X
      y := hdt.viewer.Temp.Spawn.Furniture().Y
      room_at, furn_at := floor.RoomAndFurnAtPos(x, y)
      if room_at != nil && furn_at == nil {
        switch s := hdt.viewer.Temp.Spawn.(type) {
        case *Relic:
          floor.Relics = append(floor.Relics, s)
        case *Clue:
          floor.Clues = append(floor.Clues, s)
        case *Exit:
          floor.Exits = append(floor.Exits, s)
        case *Explorer:
          floor.Explorers = append(floor.Explorers, s)
        case *Haunt:
          floor.Haunts = append(floor.Haunts, s)
        default:
          base.Error().Printf("Tried to place unknown spawn type '%t'", s)
        }
        hdt.viewer.Temp.Spawn = nil
      }
    } else {
      for _, sp := range floor.allSpawns() {
        x,y := sp.Furniture().Pos()
        if bx == x && by == y {
          hdt.viewer.Temp.Spawn = sp
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
  // for i,floor := range house.Floors {
  //   for _,room := range floor.Rooms {
  //     for _,door := range room.Doors {
  //       door.Opened = true
  //     }
  //   }
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
}
