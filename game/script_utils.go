package game

import (
  lua "github.com/xenith-studios/golua"
  "fmt"
  "io"
  "encoding/binary"
  "errors"
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
func luaEncodeValue(w io.Writer, L *lua.State, index int) error {
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
    err2 = luaEncodeTable(w, L, index)
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
func luaDecodeValue(r io.Reader, L *lua.State) error {
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
    err = luaDecodeTable(r, L)
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

func luaEncodeTable(w io.Writer, L *lua.State, index int) error {
  L.PushNil()
  for L.Next(index-1) != 0 {
    binary.Write(w, binary.LittleEndian, byte(1))
    err := luaEncodeValue(w, L, -2)
    if err != nil {
      return err
    }
    err = luaEncodeValue(w, L, -1)
    if err != nil {
      return err
    }
    L.Pop(1)
  }
  return binary.Write(w, binary.LittleEndian, byte(0))
}

// decodes a lua table and pushes it onto the stack
func luaDecodeTable(r io.Reader, L *lua.State) error {
  L.NewTable()
  var cont byte
  err := binary.Read(r, binary.LittleEndian, &cont)
  for cont != 0 && err == nil {
    for i := 0; i < 2 && err == nil; i++ {
      err = luaDecodeValue(r, L)
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
