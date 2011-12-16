package base

import (
  "os"
  "encoding/json"
  "io/ioutil"
  "strings"
  "path/filepath"
)

// Opens the file named by path, reads it all, decodes it as json into target,
// then closes the file.  Returns the first error found while doing this or nil.
func LoadJson(path string, target interface{}) error {
  f, err := os.Open(path)
  if err != nil {
    return err
  }
  defer f.Close()
  data, err := ioutil.ReadAll(f)
  if err != nil {
    return err
  }
  err = json.Unmarshal(data, target)
  if err != nil {
    return err
  }
  return nil
}

func SaveJson(path string, source interface{}) error {
  data, err := json.Marshal(source)
  if err != nil {
    return err
  }
  f, err := os.Create(path)
  if err != nil {
    return err
  }
  defer f.Close()
  _,err = f.Write(data)
  if err != nil {
    return err
  }
  return nil
}

// Returns a path rel such that filepath.Join(a, rel) and b refer to the same
// file.  a and b must both be relative paths or both be absolute paths.  If
// they are not then b will be returned in either case.
func RelativePath(a,b string) string {
  if filepath.IsAbs(a) != filepath.IsAbs(b) {
    return b
  }
  aparts := strings.Split(filepath.ToSlash(filepath.Clean(a)), "/")
  bparts := strings.Split(filepath.ToSlash(filepath.Clean(b)), "/")
  for len(aparts) > 0 && len(bparts) > 0 && aparts[0] == bparts[0] {
    aparts = aparts[1:]
    bparts = bparts[1:]
  }
  for i := range aparts {
    aparts[i] = ".."
  }
  ret := filepath.Join(filepath.Join(aparts...), filepath.Join(bparts...))
  return filepath.Clean(ret)
}

var datadir string
func SetDatadir(_datadir string) {
  datadir = _datadir
}
func GetStoreVal(key string) string {
  var store map[string]string
  LoadJson(filepath.Join(datadir, "store"), &store)
  if store == nil {
    store = make(map[string]string)
  }
  val := store[key]
  return val
}

func SetStoreVal(key,val string) {
  var store map[string]string
  path := filepath.Join(datadir, "store")
  LoadJson(path, &store)
  if store == nil {
    store = make(map[string]string)
  }
  store[key] = val
  SaveJson(path, store)
}
