--Chapter 2
function setLosModeToRoomsWithSpawnsMatching(side, pattern)
  sp = Script.GetSpawnPointsMatching(pattern)
  rooms = {}
  for i, spawn in pairs(sp) do
    rooms[i] = Script.RoomAtPos(spawn.Pos)
  end
  Script.SetLosMode(side, rooms)
end



play_as_denizens = true
function Init()
  if not store.Ch02 then
  store.Ch02 = {}
  end
   Ch01.Spawnpoints_complete{}
   store.Ch01.Spawnpoints = {
      "Ch01_Dialog01",
      "Ch01_Dialog02",
      "Ch01_Dialog03",
      "Ch01_Dialog04",
      "Ch01_Dialog05",
      "Ch01_Dialog06",
      "Ch01_Dialog07",
      "Ch01_Dialog08",
      "Ch01_Dialog09",
      "Ch01_Dialog10",
   } 

  Script.LoadHouse("Chapter_02")
  Script.DialogBox("ui/dialog/Ch01/Ch02_Dialog01.json") 

  Script.BindAi("denizen", "human")
  Script.BindAi("minions", "minions.lua")
  Script.BindAi("intruder", "intruders.lua")
    --always bind one to human!
end


function doDenizenSetup()
  ents = {
    {"Bosch", 1},
  }

  setLosModeToRoomsWithSpawnsMatching("denizens", "Master-.*")

  placed = {}
  while table.getn(placed) == 0 do
    placed = Script.PlaceEntities("Master-.*", 1, ents)
  end
  ents = {
      {"Lost Soul", 1},
      {"Poltergeist", 1},
      {"Angry Shade", 1},
    }
  setLosModeToRoomsWithSpawnsMatching("denizens", "Servitor-.*")
  Script.PlaceEntities("Servitor-.*", 10, ents)

  if store.Ch01c.choice_a == "Greedy" then
 	ents = {
 		{"Tyree's Ghost", 1},
 	}
	Script.PlaceEntities("Servitor-.*", 1, ents)
	Script.SetLosMode("denizens", "none")
  end
end

Script.DialogBox("ui/dialog/Ch02/Ch02_Dialog02.json")

-- Spawn Intruders
-- "Elias Spawn"
-- "Cordelia Spawn"
-- "Sabina Spawn"



function RoundStart(denizens, round)
    Script.SetVisibility("denizens")
    Script.SetLosMode("intruders", "entities")
    Script.SetLosMode("denizens", "entities")
    Script.ShowMainBar(intruders ~= play_as_denizens)
end


function pointIsInSpawn(pos, sp)
  return pos.X >= sp.Pos.X and pos.X < sp.Pos.X + sp.Dims.Dx and pos.Y >= sp.Pos.Y and pos.Y < sp.Pos.Y + sp.Dims.Dy
end

function pointIsInSpawns(pos, regexp)
  sps = Script.GetSpawnPointsMatching(regexp)
  for _, sp in pairs(sps) do
    if pointIsInSpawn(pos, sp) then
      return true
    end
  end
  return false
end

function IsPosInUnusedSpawnpoint(pos, name, list)
  --name identifies spawnpoint
  for _, spawn in pairs(list) do
    if not used[name] and pointIsInSpawns(pos, name) then
      return name
    end
  end
  return nil
end

function OnMove(ent, path)
  if not ent.Side.Intruder then
    return table.getn(path)
  end
  for i, pos in pairs(path) do
    name = IsPosInUnusedSpawnpoint(pos, store.Spawnpoints, store.Spawnpoints_complete)
    if name then
      return i
    end
  end
  return table.getn(path)
end



-- TRIGGER ON DEATH of
--CORDELIA
Script.DialogBox()


-- SABINA
-- TYREE
-- Elia - CHEAT DEATH sequence
	-- SPAWN Replacement?

-- Record Deaths of Units
-- Decisions
















