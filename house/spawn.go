package house

import (
  "github.com/runningwild/haunts/base"
  "github.com/runningwild/haunts/texture"
)

func MakeRelic(name string) *Relic {
  r := Relic{ Defname: name }
  base.GetObject("relic", &r)
  return &r
}

func GetAllRelicNames() []string {
  return base.GetAllNamesInRegistry("relic")
}

func LoadAllRelicsInDir(dir string) {
  base.RemoveRegistry("relic")
  base.RegisterRegistry("relic", make(map[string]*relicDef))
  base.RegisterAllObjectsInDir("relic", dir, ".json", "json")
}

type relicDef struct {
  Name  string
  Text  string
  Image texture.Object
}

type Relic struct {
  Defname string
  *relicDef

  // The pointer is used in the editor, but also stores the position of the
  // spawn point for use when the game is actually running.
  Pointer *Furniture  `registry:"loadfrom-furniture"`
}
func (s *Relic) Furniture() *Furniture {
  if s.Pointer == nil {
    s.Pointer = MakeFurniture("SpawnRelic")
  }
  return s.Pointer
}

type spawnError struct {
  msg string
}
func (se *spawnError) Error() string {
  return se.msg
}

func verifyRelicSpawns(h *HouseDef) error {
  total := 0
  for i := range h.Floors {
    total += len(h.Floors[i].Relics)
  }
  if total < 5 {
    return &spawnError{ "House needs at least five relic spawn points." }
  }
  return nil
}

// func verifyPlayerSpawns(h *HouseDef) error {
//   total := 0
//   for i := range h.Floors {
//     total += len(h.Floors[i].Players)
//   }
//   if total < 1 {
//     return &spawnError{ "House needs at least one player spawn point." }
//   }
//   return nil
// }

// func verifyCleanseSpawns(h *HouseDef) error {
//   total := 0
//   for i := range h.Floors {
//     total += len(h.Floors[i].Cleanse)
//   }
//   if total < 3 {
//     return &spawnError{ "House needs at least cleanse spawn points." }
//   }
//   return nil
// }

// func verifyClueSpawns(h *HouseDef) error {
//   total := 0
//   for i := range h.Floors {
//     total += len(h.Floors[i].Clues)
//   }
//   if total < 10 {
//     return &spawnError{ "House needs at least ten clue spawn points." }
//   }
//   return nil
// }

// func verifyExitSpawns(h *HouseDef) error {
//   total := 0
//   for i := range h.Floors {
//     total += len(h.Floors[i].Exits)
//   }
//   if total < 1 {
//     return &spawnError{ "House needs at least one exit spawn point." }
//   }
//   return nil
// }
