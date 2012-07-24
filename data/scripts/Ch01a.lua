function setLosModeToRoomsWithSpawnsMatching(side, pattern)
  sp = Script.GetSpawnPointsMatching(pattern)
  rooms = {}
  for i, spawn in pairs(sp) do
    rooms[i] = Script.RoomAtPos(spawn.Pos)
  end
  Script.SetLosMode(side, rooms)
end

if not store.Ch01a then
  store.Ch01a = {}
  end
--
play_as_denizens = false
function Init()
   store.Ch01a = {}
   store.Ch01a.Spawnpoints_complete={}
   store.Ch01a.Spawnpoints = {
      "Ch01_Dialog01",
      "Ch01_Dialog02",
      "Ch01_Dialog03",
   } 

  Script.LoadHouse("Chapter_01_a")
  Script.DialogBox("ui/dialog/Ch01/Ch01_Dialog01.json") 

  Script.BindAi("denizen", "denizens.lua")
  Script.BindAi("minions", "minions.lua")
  Script.BindAi("intruder", "human")
    --always bind one to human!
  intruder_spawn = Script.GetSpawnPointsMatching("Intruders-FrontDoor")
  Script.SpawnEntitySomewhereInSpawnPoints("Caitlin", intruder_spawn)
  Script.SpawnEntitySomewhereInSpawnPoints("Percy", intruder_spawn)
  ents = Script.GetAllEnts()
end
 

function RoundStart(intruders, round)
    Script.SetVisibility("intruders")
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

function IsPosInUnusedSpawnpoint(pos, list, used)
  --name identifies spawnpoint
  for _, spawn in pairs(list) do
    if not used[spawn] and pointIsInSpawns(pos, spawn) then
      return spawn
    end
  end
  return nil
end


--THIS STOPS a unit in a spawn point not yet activated.
function OnMove(ent, path)
  if not ent.Side.Intruder then
    return table.getn(path)
  end
  for i, pos in pairs(path) do
    name = IsPosInUnusedSpawnpoint(pos, store.Ch01a.Spawnpoints, store.Ch01a.Spawnpoints_complete)
    if name then
      return i
    end
  end
  return table.getn(path)
end

function OnAction(intruders, round, exec)
  if not exec.Ent.Side.Intruder then
    return
  end
  name = IsPosInUnusedSpawnpoint(exec.Ent.Pos, store.Ch01a.Spawnpoints, store.Ch01a.Spawnpoints_complete)
  if name then
    dialog_path = "ui/dialog/Ch01/" .. name .. ".json"
    Script.DialogBox(dialog_path)
    store.Ch01a.Spawnpoints_complete[name] = true 
    if name == "Ch01_Dialog02" then
      Script.StartScript("Ch01b.lua")
    --INSERT other names and functions here
    end  
  end
end
 

function RoundEnd(intruders, round)
  print("end", intruders, round)
end

