-- function setLosModeToRoomsWithSpawnsMatching(side, pattern)
--   sp = Script.GetSpawnPointsMatching(pattern)
--   rooms = {}
--   for i, spawn in pairs(sp) do
--     rooms[i] = Script.RoomAtPos(spawn.Pos)
--   end
--   Script.SetLosMode(side, rooms)
-- end



function Init()
  if not store.Ch01a then
    store.Ch01b = {}
  end
  store.Ch01b.Spawnpoints_complete={}
  store.Ch01b.Spawnpoints = {
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
  print("Found", table.getn(intruder_spawn), "points")
  Script.SpawnEntitySomewhereInSpawnPoints("Caitlin", intruder_spawn)
  Script.SpawnEntitySomewhereInSpawnPoints("Percy", intruder_spawn)
end
 

function RoundStart(intruders, round)
  Script.SetVisibility("intruders")
  Script.SetLosMode("intruders", "entities")
  Script.SetLosMode("denizens", "entities")
  Script.ShowMainBar(intruders)
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

function OnMove(ent, path)
  if not ent.Side.Intruder then
    return table.getn(path)
  end
  for i, pos in pairs(path) do
    name = IsPosInUnusedSpawnpoint(pos, store.Ch01b.Spawnpoints, store.Ch01b.Spawnpoints_complete)
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
  name = IsPosInUnusedSpawnpoint(exec.Ent.Pos, store.Ch01b.Spawnpoints, store.Ch01b.Spawnpoints_complete)
  if name == "Ch01_Dialog03" then
    Script.DialogBox("ui/dialog/Ch01/Ch01_Dialog03.json")
    store.Ch01b.Spawnpoints_complete["Ch01_Dialog03"] = true
  end
  if name == "Ch01_Dialog04" then
    Script.DialogBox("ui/dialog/Ch01/Ch01_Dialog04.json")
    store.Ch01b.Spawnpoints_complete["Ch01_Dialog04"] = true
  end
  if name == "Ch01_Dialog05" then
    Script.DialogBox("ui/dialog/Ch01/Ch01_Dialog05.json")
    store.Ch01b.Spawnpoints_complete["Ch01_Dialog05"] = true
  end
  if name == "Ch01_Dialog06" then
    Script.DialogBox("ui/dialog/Ch01/Ch01_Dialog06.json")
    store.Ch01b.Spawnpoints_complete["Ch01_Dialog06"] = true
  end
  if name == "Ch01_Dialog07" then
    Script.DialogBox("ui/dialog/Ch01/Ch01_Dialog07.json")
    store.Ch01b.Spawnpoints_complete["Ch01_Dialog07"] = true
  end
  if name == "Ch01_Dialog08" then
    Script.DialogBox("ui/dialog/Ch01/Ch01_Dialog08.json")
    store.Ch01b.Spawnpoints_complete["Ch01_Dialog08"] = true
  end
  if name == "Ch01_Dialog09" then
    Script.DialogBox("ui/dialog/Ch01/Ch01_Dialog09.json")
    store.Ch01b.Spawnpoints_complete["Ch01_Dialog09"] = true
  end
  if name == "Ch01_Dialog10" then
    Script.DialogBox("ui/dialog/Ch01/Ch01_Dialog10.json")
    store.Ch01b.Spawnpoints_complete["Ch01_Dialog10"] = true
  end

  if name == "Shade Trigger01" then
    print("SPAWN SHADE")
    shade_spawn = Script.GetSpawnPointsMatching("^Shade Spawn01$")
    print("found", shade_spawn)
    ent = Script.SpawnEntitySomewhereInSpawnPoints("Shade", shade_spawn)
    print("ENT:", ent.Name)
    store.Ch01b.Spawnpoints_complete["Shade Trigger01"] = true
  end

  if name == "Dining Trigger01" then
    shade_spawn = Script.GetSpawnPointsMatching("Shade Spawn02")
    store.temporary_shades = {}
    for i = 1,4 do
      ent = SpawnEntitySomewhereInSpawnPoints("Shade", shade_spawn)
      store.temporary_shades[i] = ent
    end
  end

  if name == "Harry Trigger01" then
    Script.DialogBox(ui/dialog/Ch01/"Ch01_Dialog04")
    harry_spawn = Script.GetSpawnPointsMatching("Harry Spawn")

    store.angry_shades = {}
    for i = 1,5 do
      ent = SpawnEntitySomewhereInSpawnPoints("Angry Shade", angry_shade_spawn)
      SpawnEntitySomewhereInSpawnPoints("Scared Man", harry_spawn)
   end  
end

  all_dead = true
  for _, ent in pairs(store.angry_shades) do
    if ent.Stats.HpCur > 0 then
      all_dead = false
    end
  end
  if all_dead then
    Script.DialogBox("ui/dialog/Ch01/Ch01_Dialog06.json")
    store.Ch01b.Spawnpoints_complete["Ch01_Dialog06"] = true
    Script.LoadScript("Ch01c.lua")
  end
end

end

function RoundEnd(intruders, round)
  if not intruders then
    for _, ent in pairs(Script.GetAllEnts()) do
      if ent.Name == "Shade" then
        if ent.HpCur == 1 then
          Script.PlayAnimations(ent, {"defend", "killed"})
          Script.SetHp(ent, 0)
        end
        if ent.HpCur > 1 then 
          Script.SetHp(ent, 1)
        end
      end
    end
  end
end