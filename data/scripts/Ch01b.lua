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
--
play_as_denizens = false
function Init()
   store.Ch01b = {}
   store.Ch01b.Spawnpoints_complete={}
   store.Ch01b.Spawnpoints = {
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

    if name == "Shade Trigger01" then
      shade_spawn = Script.GetSpawnPointsMatching("Shade Spawn01")
      for i = 1 do
        ent = SpawnEntitySomewhereInSpawnPoints("Shade", shade_spawn)
        store.temporary_shades[i] = ent
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
    Script.LoadScript("Chapter_01_c")
  end
end

--ANY WAY TO CHANGE THE TIMING ON THIS?
--if ent.HpCur > 1 then 
--  Script.SetHp(ent, 1)
--if ent.HpCur == 1 then
--  Script.SetHp(ent, 0)
--end

function RoundEnd(intruders, round)
  for _, ent in pairs(store.temporary_shades) do
    if ent.HpCur > 1 then 
      Script.SetHp(ent, 1)
    if ent.HpCur == 1 then
      Script.SetHp(ent, 0)
    end
  end
  store.temporary_shades = nil
  print("end", intruders, round)
end