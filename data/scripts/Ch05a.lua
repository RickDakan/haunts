--Ch05a
function setLosModeToRoomsWithSpawnsMatching(side, pattern)
  sp = Script.GetSpawnPointsMatching(pattern)
  rooms = {}
  for i, spawn in pairs(sp) do
    rooms[i] = Script.RoomAtPos(spawn.Pos)
  end
  Script.SetLosMode(side, rooms)
end

if not store.Ch05a then
  store.Ch05a = {}
  end
--
play_as_denizens = false
function Init()
   store.Ch05a = {}
   store.Ch05a.Spawnpoints_complete={}
   store.Ch05a.Spawnpoints = {
      "Foyer",
      "Private Entrance",
      "Study Desk",
      "Lab",
      "Vault Door",
      "Choose Character"
      "Creepy1",
      "Creepy2",
      "Creepy3",

   } 

  Script.LoadHouse("Chapter_05_a")
  Script.DialogBox("ui/dialog/Ch05/Ch05_Intro.json") 

  Script.BindAi("denizen", "denizens.lua")
  Script.BindAi("minions", "minions.lua")
  Script.BindAi("intruder", "human")
    --always bind one to human!
  intruder_spawn = Script.GetSpawnPointsMatching("Intruders-Start")
  Script.SpawnEntitySomewhereInSpawnPoints("Timothy K.", intruder_spawn)
  Script.SpawnEntitySomewhereInSpawnPoints("Constance M.", intruder_spawn)
  Script.SpawnEntitySomewhereInSpawnPoints("Samuel T.", intruder_spawn)



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
    name = IsPosInUnusedSpawnpoint(pos, store.Ch05a.Spawnpoints, store.Ch05a.Spawnpoints_complete)
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
  name = IsPosInUnusedSpawnpoint(exec.Ent.Pos, store.Ch05a.Spawnpoints, store.Ch05a.Spawnpoints_complete)

  if name == "Creep1" then
  	--SOUND CUE!


 end
end


function RoundEnd(intruders, round)
  print("end", intruders, round)
end




