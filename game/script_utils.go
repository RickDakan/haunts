package game

import (
  "fmt"
  "io"
  "encoding/binary"
  "errors"
  "sort"
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

type FunctionTable map[string]func()

func LuaPushSmartFunctionTable(L *lua.State, ft FunctionTable) {
  // Copy it just in case - I can't imagine someone changing it after passing
  // it to this function, but I don't want to take any chances.
  myft := make(FunctionTable)
  for n, f := range ft {
    myft[n] = f
  }
  names := make([]string, len(myft))[0:0]
  for name := range myft {
    names = append(names, name)
  }
  sort.Strings(names)
  valid_selectors := "["
  for i, name := range names {
    if i > 0 {
      valid_selectors += ", "
    }
    valid_selectors += fmt.Sprintf("'%s'", name)
  }
  valid_selectors += "]."

  L.NewTable()
  L.PushString("__index")
  L.PushGoFunctionAsCFunction(func(L *lua.State) int {
    name := L.ToString(-1)
    if f, ok := myft[name]; ok {
      f()
    } else {
      base.Error().Printf("'%s' is not a valid selector, valid seletors are %s", name, valid_selectors)
      L.PushNil()
    }
    return 1
  })
  L.SetTable(-3)
}

// Pushes an entity onto the stack, it is a table containing the following:
// e.id -> EntityId of this entity
// e.name -> Name as displayed to the user
// e.gear_options -> Table mapping gear to icon for all available gear
// e.gear -> Name of the selected gear, nil if none is selected
// e.actions -> Array of actions this entity has available
func LuaPushEntity(L *lua.State, ent *Entity) {
  if ent == nil {
    L.PushNil()
    return
  }
  // id and Name can be added to the ent table as static data since they 
  // never change.
  L.NewTable()
  L.PushString("id")
  L.PushInteger(int(ent.Id))
  L.SetTable(-3)
  L.PushString("Name")
  L.PushString(ent.Name)
  L.SetTable(-3)

  // Meta table for the Entity so that any dynamic data is generated
  // on-the-fly
  LuaPushSmartFunctionTable(L, FunctionTable{
    "Conditions": func() {
      L.NewTable()
      for _, condition := range ent.Stats.ConditionNames() {
        L.PushString(condition)
        L.PushBoolean(true)
        L.SetTable(-3)
      }
    },
    "GearOptions": func() {
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
    },
    "Gear": func() {
      if ent.ExplorerEnt != nil && ent.ExplorerEnt.Gear != nil {
        L.PushString(ent.ExplorerEnt.Gear.Name)
      } else {
        L.PushNil()
      }
    },
    "Actions": func() {
      L.NewTable()
      for _, action := range ent.Actions {
        L.PushString(action.String())
        action.Push(L)
        L.SetTable(-3)
      }
    },
    "Pos": func() {
      x, y := ent.Pos()
      pushPoint(L, x, y)
    },
    "Corpus": func() {
      L.PushInteger(ent.Stats.Corpus())
    },
    "Ego": func() {
      L.PushInteger(ent.Stats.Ego())
    },
    "HpCur": func() {
      L.PushInteger(ent.Stats.HpCur())
    },
    "HpMax": func() {
      L.PushInteger(ent.Stats.HpMax())
    },
    "ApCur": func() {
      L.PushInteger(ent.Stats.ApCur())
    },
    "ApMax": func() {
      L.PushInteger(ent.Stats.ApMax())
    },
    "Info": func() {
      L.NewTable()
      L.PushString("LastEntityThatIAttacked")
      LuaPushEntity(L, ent.Game().EntityById(ent.Info.LastEntThatIAttacked))
      L.SetTable(-3)
      L.PushString("LastEntThatAttackedMe")
      LuaPushEntity(L, ent.Game().EntityById(ent.Info.LastEntThatAttackedMe))
      L.SetTable(-3)
    },
  })
  L.SetMetaTable(-2)
}
