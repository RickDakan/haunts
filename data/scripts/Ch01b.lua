function setLosModeToRoomsWithSpawnsMatching(side, pattern)
  sp = Script.GetSpawnPointsMatching(pattern)
  rooms = {}
  for i, spawn in pairs(sp) do
    rooms[i] = Script.RoomAtPos(spawn.Pos)
  end
  Script.SetLosMode(side, rooms)
end

if not store.Ch01b then
  store.Ch01b = {}
  end

play_as_denizens = false
function Init()
  if not store.Ch01 then
  store.Ch01 = {}
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
      "Shade Trigger01",
      "Dining Trigger01",
      "Harry Trigger01",
   } 


  Script.LoadHouse("Chapter_01_b")
 
  Script.BindAi("denizen", "denizens.lua")
  Script.BindAi("minions", "minions.lua")
  Script.BindAi("intruder", "human")
 
  intruder_spawn = Script.GetSpawnPointsMatching("Intruders-Start")
  Script.SpawnEntitySomewhereInSpawnPoints("Caitlin", intruder_spawn)
  Script.SpawnEntitySomewhereInSpawnPoints("Percy", intruder_spawn)
  ents = Script.GetAllEnts()
  end
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
  print("Regexp: '",regexp,"'")
  sps = Script.GetSpawnPointsMatching(regexp)
  print("Got ", "spawns for ", regexp)
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

function OnAction(intruders, round, exec)
  if not exec.ent.Side.Intruder then
    return
  end
  name = IsPosInUnusedSpawnpoint(pos, store.Spawnpoints, store.Spawnpoints_complete)
  if name then
    dialog_path = "ui/dialog/Ch01/" .. name .. ".json"
    Script.DialogBox(dialog_path)
    store.Ch01.Spawnpoints_complete[name] = true 
    if name == "Ch01_Dialog02" then
      Script.LoadScript("Chapter_01_b")
    end

    --INSERT other names and functions here

    if name == "Shade Trigger01" then
      shade_spawn = Script.GetSpawnPointsMatching("Shade Spawn01")
      SpawnEntitySomewhereInSpawnPoints("Shade", shade_spawn)
    end

    --NEED TO UNSPAWN IT???

    if name == "Dining Trigger01" then
      shade_spawn = Script.GetSpawnPointsMatching("Shade Spawn02")
      SpawnEntitySomewhereInSpawnPoints("Shade", shade_spawn)
      SpawnEntitySomewhereInSpawnPoints("Shade", shade_spawn)
      SpawnEntitySomewhereInSpawnPoints("Shade", shade_spawn)
      SpawnEntitySomewhereInSpawnPoints("Shade", shade_spawn)
    end

    -- NEED TO UNSPAWN THEM???


    if name == "Harry Trigger01" then
      Script.DialogBox(ui/dialog/Ch01/"Ch01_Dialog04")
      angry_shade_spawn = Script.GetSpawnPointsMatching("Angry Shade Spawn01")
      harry_spawn = Script.GetSpawnPointsMatching("Harry Spawn")
      SpawnEntitySomewhereInSpawnPoints("Angry Shade", angry_shade_spawn)
      SpawnEntitySomewhereInSpawnPoints("Angry Shade", angry_shade_spawn)
      SpawnEntitySomewhereInSpawnPoints("Angry Shade", angry_shade_spawn)
      SpawnEntitySomewhereInSpawnPoints("Angry Shade", angry_shade_spawn)
      SpawnEntitySomewhereInSpawnPoints("Angry Shade", angry_shade_spawn)
      SpawnEntitySomewhereInSpawnPoints("Scared Man", harry_spawn)
     end  
  end
-- Kill Entity???? Make Shade Disappear


--ENTER DINING ROOM
--SPAWN 4 shades
--Kill 4 shades later

--- SHADES ALL DEAD???

Script.DialogBox("ui/dialog/Ch01/Ch01_Dialog06.json")
Script.LoadScript("Chapter_01_c")

end
 




function RoundEnd(intruders, round)
  print("end", intruders, round)

-- Kill Dudes on Round end

-- 


end