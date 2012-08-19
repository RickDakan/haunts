function setLosModeToRoomsWithSpawnsMatching(side, pattern)
  sp = Script.GetSpawnPointsMatching(pattern)
  rooms = {}
  for i, spawn in pairs(sp) do
    rooms[i] = Script.RoomAtPos(spawn.Pos)
  end
  Script.SetLosMode(side, rooms)
end

function Init()
  Script.LoadHouse("versus-1")

  Script.BindAi("denizen", "human")
  Script.BindAi("minions", "minions.lua")
  Script.BindAi("intruder", "smrt/intruders.lua")
    --always bind one to human!

  intruders_spawn = Script.GetSpawnPointsMatching("Intruders-FrontDoor")
  denizens_spawn = Script.GetSpawnPointsMatching("Master")
  relics_spawn = Script.GetSpawnPointsMatching("Relics")
  for i=1,5 do
    Script.SpawnEntitySomewhereInSpawnPoints("Altar01", relics_spawn)
  end
  Script.SpawnEntitySomewhereInSpawnPoints("Chosen One", denizens_spawn)
  occ = Script.SpawnEntitySomewhereInSpawnPoints("Occultist", intruders_spawn)
  Script.BindAi(occ, "smrt/occultist.lua")
  teen = Script.SpawnEntitySomewhereInSpawnPoints("Teen", intruders_spawn)
  Script.BindAi(teen, "smrt/teen.lua")
end
 

function RoundStart(intruders, round)
  if intruders then
    Script.SetVisibility("intruders")
  else
    Script.SetVisibility("denizens")
  end
  Script.SetLosMode("intruders", "entities")
  Script.SetLosMode("denizens", "entities")
  Script.ShowMainBar(true)
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
  return table.getn(path)
end

function OnAction(intruders, round, exec)
end
 

function RoundEnd(intruders, round)
  print("end", intruders, round)
end

