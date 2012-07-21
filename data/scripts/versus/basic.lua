function setLosModeToRoomsWithSpawnsMatching(side, pattern)
  sp = Script.GetSpawnPointsMatching(pattern)
  rooms = {}
  for i, spawn in pairs(sp) do
    rooms[i] = Script.RoomAtPos(spawn.Pos)
  end
  Script.SetLosMode(side, rooms)
end

play_as_denizens = false
function Init(data)
  for k, v in pairs(data) do
    print("data:", k, v)
  end
  Script.LoadHouse("Chapter_01_a")
  -- Script.DialogBox("ui/dialog/Ch01/Ch01_Dialog01.json") 

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
  -- Script.SetVisibility("intruders")
  -- Script.SetLosMode("intruders", "entities")
  -- Script.SetLosMode("denizens", "entities")
  -- Script.ShowMainBar(intruders ~= play_as_denizens)
end


function OnMove(ent, path)
  return table.getn(path)
end

function OnAction(intruders, round, exec)
end
 

function RoundEnd(intruders, round)
end

