package game

import (
  "fmt"
  "io"
  "encoding/binary"
  "errors"
  "github.com/runningwild/haunts/base"
  lua "github.com/xenith-studios/golua"
)

type luaEncodable int32

const (
  luaBool luaEncodable = iota
  luaNumber
  luaNil
  luaTable
  luaString
)

// Encodes a lua value, only bool, number, nil, table, and string can be
// encoded.  If an unencodable value is encountered an error will be
// returned.
func LuaEncodeValue(w io.Writer, L *lua.State, index int) error {
  var err1, err2, err3 error
  switch {
  case L.IsBoolean(index):
    err1 = binary.Write(w, binary.LittleEndian, luaBool)
    var v byte = 0
    if L.ToBoolean(index) {
      v = 1
    }
    err2 = binary.Write(w, binary.LittleEndian, v)
  case L.IsNumber(index):
    err1 = binary.Write(w, binary.LittleEndian, luaNumber)
    err2 = binary.Write(w, binary.LittleEndian, L.ToNumber(index))
  case L.IsNil(index):
    err1 = binary.Write(w, binary.LittleEndian, luaNil)
  case L.IsTable(index):
    err1 = binary.Write(w, binary.LittleEndian, luaTable)
    err2 = LuaEncodeTable(w, L, index)
  case L.IsString(index):
    err1 = binary.Write(w, binary.LittleEndian, luaString)
    str := L.ToString(index)
    err2 = binary.Write(w, binary.LittleEndian, uint32(len(str)))
    err3 = binary.Write(w, binary.LittleEndian, []byte(str))
  default:
    return errors.New(fmt.Sprintf("Cannot encode lua type id == %d.", L.Type(index)))
  }
  switch {
  case err1 != nil:
    return err1
  case err2 != nil:
    return err2
  case err3 != nil:
    return err3
  }
  return nil
}

// Decodes a value from the reader and pushes it onto the stack
func LuaDecodeValue(r io.Reader, L *lua.State) error {
  var le luaEncodable
  err := binary.Read(r, binary.LittleEndian, &le)
  if err != nil {
    return err
  }
  switch le {
  case luaBool:
    var v byte
    err = binary.Read(r, binary.LittleEndian, &v)
    L.PushBoolean(v == 1)
  case luaNumber:
    var f float64
    err = binary.Read(r, binary.LittleEndian, &f)
    L.PushNumber(f)
  case luaNil:
    L.PushNil()
  case luaTable:
    err = LuaDecodeTable(r, L)
  case luaString:
    var length uint32
    err = binary.Read(r, binary.LittleEndian, &length)
    if err != nil {
      return err
    }
    sb := make([]byte, length)
    err = binary.Read(r, binary.LittleEndian, &sb)
    L.PushString(string(sb))
  default:
    return errors.New(fmt.Sprintf("Unknown lua value id == %d.", le))
  }
  if err != nil {
    return err
  }
  return nil
}

func LuaEncodeTable(w io.Writer, L *lua.State, index int) error {
  L.PushNil()
  for L.Next(index-1) != 0 {
    binary.Write(w, binary.LittleEndian, byte(1))
    err := LuaEncodeValue(w, L, -2)
    if err != nil {
      return err
    }
    err = LuaEncodeValue(w, L, -1)
    if err != nil {
      return err
    }
    L.Pop(1)
  }
  return binary.Write(w, binary.LittleEndian, byte(0))
}

// decodes a lua table and pushes it onto the stack
func LuaDecodeTable(r io.Reader, L *lua.State) error {
  L.NewTable()
  var cont byte
  err := binary.Read(r, binary.LittleEndian, &cont)
  for cont != 0 && err == nil {
    for i := 0; i < 2 && err == nil; i++ {
      err = LuaDecodeValue(r, L)
    }
    if err == nil {
      err = binary.Read(r, binary.LittleEndian, &cont)
    }
    L.SetTable(-3)
  }
  if err != nil {
    return err
  }
  return nil
}

// Gets the id out of the table at the specified index and returns the
// associated Entity, or nil if there is none.
func LuaToEntity(L *lua.State, game *Game, index int) *Entity {
  L.PushString("id")
  L.GetTable(index - 1)
  id := EntityId(L.ToInteger(-1))
  L.Pop(1)
  return game.EntityById(id)
}

// Pushes an entity onto the stack, it is a table containing the following:
// e.id -> EntityId of this entity
// e.name -> Name as displayed to the user
// e.gear_options -> Table mapping gear to icon for all available gear
// e.gear -> Name of the selected gear, nil if none is selected
// e.actions -> Array of actions this entity has available
func LuaPushEntity(L *lua.State, ent *Entity) {
  L.NewTable()
  L.PushString("id")
  L.PushInteger(int(ent.Id))
  L.SetTable(-3)
  L.PushString("Name")
  L.PushString(ent.Name)
  L.SetTable(-3)

  L.PushString("GearOptions")
  L.NewTable()
  if ent.ExplorerEnt != nil {
    for _, gear_name := range ent.ExplorerEnt.Gear_names {
      var g Gear
      g.Defname = gear_name
      base.GetObject("gear", &g)
      L.PushString(gear_name)
      L.PushString(g.Large_icon.Path.String())
      L.SetTable(-3)
    }
  }
  L.SetTable(-3)

  L.PushString("Gear")
  if ent.ExplorerEnt != nil && ent.ExplorerEnt.Gear != nil {
    L.PushString(ent.ExplorerEnt.Gear.Name)
  } else {
    L.PushNil()
  }
  L.SetTable(-3)

  L.PushString("Actions")
  L.NewTable()
  for _, action := range ent.Actions {
    L.PushString(action.String())
    action.Push(L)
    L.SetTable(-3)
  }
  L.SetTable(-3)

  L.PushString("Conditions")
  L.NewTable()
  for _, condition := range ent.Stats.ConditionNames() {
    L.PushString(condition)
    L.PushBoolean(true)
    L.SetTable(-3)
  }
  L.SetTable(-3)

  L.PushString("Info")
  L.PushGoFunction(func(L *lua.State) int {
    L.NewTable()
    L.PushString("LastEntityThatIAttacked")
    e := ent.Game().EntityById(ent.Info.LastEntThatIAttacked)
    if e != nil {
      LuaPushEntity(L, e)
    } else {
      L.PushNil()
    }
    L.SetTable(-3)
    L.PushString("LastEntityThatAttackedMe")
    e = ent.Game().EntityById(ent.Info.LastEntThatAttackedMe)
    if e != nil {
      LuaPushEntity(L, e)
    } else {
      L.PushNil()
    }
    L.SetTable(-3)
    return 1
  })
  L.SetTable(-3)

  L.PushString("Pos")
  x, y := ent.Pos()
  pushPoint(L, x, y)
  L.SetTable(-3)

  L.PushString("Corpus")
  L.PushInteger(ent.Stats.Corpus())
  L.SetTable(-3)
  L.PushString("Ego")
  L.PushInteger(ent.Stats.Ego())
  L.SetTable(-3)
  L.PushString("HpCur")
  L.PushInteger(ent.Stats.HpCur())
  L.SetTable(-3)
  L.PushString("HpMax")
  L.PushInteger(ent.Stats.HpMax())
  L.SetTable(-3)
  L.PushString("ApCur")
  L.PushInteger(ent.Stats.ApCur())
  L.SetTable(-3)
  L.PushString("ApMax")
  L.PushInteger(ent.Stats.ApMax())
  L.SetTable(-3)
}
