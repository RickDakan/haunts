package base

import (
  "encoding/gob"
  "encoding/json"
  "image/color"
  "io/ioutil"
  "os"
  "path/filepath"
  "github.com/runningwild/opengl/gl"
  "github.com/runningwild/glop/gui"
  "github.com/runningwild/glop/render"
  "code.google.com/p/freetype-go/freetype/truetype"
  "log"
  "fmt"
  "time"
  "sync"
)

var datadir string
var logger *log.Logger
func SetDatadir(_datadir string) {
  datadir = _datadir
  setupLogger()
}
func GetDataDir() string {
  return datadir
}

func setupLogger() {
  // If an error happens when making this directory it might already exist,
  // all that really matters is making the log file in the directory.
  os.Mkdir(filepath.Join(datadir, "logs"), 0777)
  logger = nil
  var out *os.File
  var err error
  name := time.Now().Format("2006-01-02-15-04-05") + ".log"
  out, err = os.Create(filepath.Join(datadir, "logs", name))
  if err != nil {
    fmt.Printf("Unable to open log file: %v\nLogging to stdout...\n", err.Error())
    out = os.Stdout
  }
  logger = log.New(out, "> ", log.Ltime | log.Lshortfile)
}

// TODO: This probably isn't the best way to do things - different go-routines
// can call these and screw up prefixes for each other.
func Log() *log.Logger {
  logger.SetPrefix("LOG  > ")
  return logger
}

func Warn() *log.Logger {
  logger.SetPrefix("WARN > ")
  return logger
}

func Error() *log.Logger {
  logger.SetPrefix("ERROR> ")
  return logger
}

var font_dict map[int]*gui.Dictionary
var dictionary_once sync.Once
func setupFontDictionaries() {
  font_dict = make(map[int]*gui.Dictionary)
  f, err := os.Open(filepath.Join(datadir, "fonts", "tomnr.ttf"))
  if err != nil {
    Error().Fatalf("Unable to load font: %v", err)
  }
  data, err := ioutil.ReadAll(f)
  f.Close()
  if err != nil {
    Error().Fatalf("Unable to read font: %v", err)
  }
  font, err := truetype.Parse(data)
  if err != nil {
    Error().Fatalf("Unable to parse font: %v", err)
  }
  render.Init()
  font_dict[10] = gui.MakeDictionary(font, 10)
  font_dict[12] = gui.MakeDictionary(font, 12)
  font_dict[15] = gui.MakeDictionary(font, 15)
  font_dict[25] = gui.MakeDictionary(font, 25)
}

func GetDictionary(point_size int) *gui.Dictionary {
  dictionary_once.Do(setupFontDictionaries)
  return font_dict[point_size]
}

// A Path is a string that is intended to store a path.  When it is encoded
// with gob or json it will convert itself to a relative path relative to
// datadir.  When it is decoded from gob or json it will convert itself to an
// absolute path based on datadir.
type Path string
func (p Path) String() string {
  return string(p)
}
func (p Path) GobEncode() ([]byte, error) {
  return []byte(TryRelative(datadir, string(p))), nil
}
func (p *Path) GobDecode(data []byte) error {
  *p = Path(filepath.Join(datadir, string(data)))
  return nil
}
func (p Path) MarshalJSON() ([]byte, error) {
  val := filepath.ToSlash(TryRelative(datadir, string(p)))
  return []byte("\"" + val + "\""), nil
}
func (p *Path) UnmarshalJSON(data []byte) error {
  rel := filepath.FromSlash(string(data[1 : len(data) - 1]))
  *p = Path(filepath.Join(datadir, rel))
  return nil
}

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
  return err
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
  return err
}

// Opens the file named by path, reads it all, decodes it as gob into target,
// then closes the file.  Returns the first error found while doing this or nil.
func LoadGob(path string, target interface{}) error {
  f, err := os.Open(path)
  if err != nil {
    return err
  }
  defer f.Close()
  dec := gob.NewDecoder(f)
  err = dec.Decode(target)
  return err
}

func SaveGob(path string, source interface{}) error {
  f, err := os.Create(path)
  if err != nil {
    return err
  }
  defer f.Close()
  enc := gob.NewEncoder(f)
  err = enc.Encode(source)
  return err
}

// Returns a path rel such that filepath.Join(a, rel) and b refer to the same
// file.  a and b must both be relative paths or both be absolute paths.  If
// they are not then b will be returned in either case.
func TryRelative(base,target string) string {
  rel,err := filepath.Rel(base, target)
  if err == nil {
    return rel
  }
  return target
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

type ColorStack struct {
  colors []color.NRGBA
}
func (cs *ColorStack) Push(r,g,b,a float64) {
  c := color.NRGBA{ byte(255 * r), byte(255 * g), byte(255 * b), byte(255 * a) }
  cs.colors = append(cs.colors, c)
}
func (cs *ColorStack) Pop() {
  cs.colors = cs.colors[0 : len(cs.colors) - 1]
}
func (cs *ColorStack) subApply(n int) (r,g,b,a float64) {
  if n < 0 {
    return 1, 1, 1, 0
  }
  dr, dg, db, da := cs.subApply(n - 1)
  a = float64(cs.colors[n].A) / 255
  r = float64(cs.colors[n].R) / 255 * a + dr * (1 - a)
  g = float64(cs.colors[n].G) / 255 * a + dg * (1 - a)
  b = float64(cs.colors[n].B) / 255 * a + db * (1 - a)
  a = a + (1 - a) * da
  return
}
func (cs *ColorStack) Apply() {
  gl.Color4d(cs.subApply(len(cs.colors) - 1))
}
func (cs *ColorStack) ApplyWithAlpha(alpha float64) {
  r,g,b,a := cs.subApply(len(cs.colors) - 1)
  gl.Color4d(r, g, b, a * alpha)
}

