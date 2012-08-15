function Init()
   store.Ch03d = {}
   store.Ch03d.Spawnpoints_complete={}
   store.Ch03d.Spawnpoints = {
      "Patient_One_Finale",
      "Patient_Two_Finale",
      "Patient_Three_Finale",
      "Patient_Four_Finale",
      "Dark_Offer_Trigger",
      "Dark_Offer_Finale"
   } 

  Script.LoadHouse("Chapter_03_d")

  Script.BindAi("denizen", "denizens.lua")
  Script.BindAi("minions", "minions.lua")
  Script.BindAi("intruder", "human")
    --always bind one to human!
  intruder_spawn = Script.GetSpawnPointsMatching("Intruders-Start")
  Script.SpawnEntitySomewhereInSpawnPoints("Lt. Wilson Sax", intruder_spawn)

  tryee_spawn = Script.GetSpawnPointsMatching("Tyree-Start")
  Script.SpawnEntitySomewhereInSpawnPoints("The Director")
  
  patients_spawn = Script.GetSpawnPointsMatching("Patients-Start")
  Script.SpawnEntitySomewhereInSpawnPoints("Patient", patients_spawn)
  Script.SpawnEntitySomewhereInSpawnPoints("Patient", patients_spawn)
  Script.SpawnEntitySomewhereInSpawnPoints("Patient", patients_spawn)
  Script.SpawnEntitySomewhereInSpawnPoints("Patient", patients_spawn)
  
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
    name = IsPosInUnusedSpawnpoint(pos, store.Ch03d.Spawnpoints, store.Ch03d.Spawnpoints_complete)
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
  name = IsPosInUnusedSpawnpoint(exec.Ent.Pos, store.Ch03d.Spawnpoints, store.Ch03d.Spawnpoints_complete)

  -- store.Ch03c.choice_a = choices["Cultists"]

  if name == "Patient_One_Finale" then
    Script.DialogBox("ui/dialog/Ch03/Ch03_Patient01_finale.json")
    store.Ch03d.Spawnpoints_complete["Ch03_Patient_One_Finale"] = true
  end

  if name == "Patient_Two_Finale" then
    Script.DialogBox("ui/dialog/Ch03/Ch03_Patient02_finale.json")
    store.Ch03d.Spawnpoints_complete["Ch03_Patient_Two_Finale"] = true
  end

  if name == "Patient_Three_Finale" then
    Script.DialogBox("ui/dialog/Ch03/Ch03_Patient03_finale.json")
    store.Ch03d.Spawnpoints_complete["Ch03_Patient_Three_Finale"] = true
  end

  if name == "Patient_Four_Finale" then
    Script.DialogBox("ui/dialog/Ch03/Ch03_Patient04_finale.json")
    store.Ch03d.Spawnpoints_complete["Ch03_Patient_Four_Finale"] = true
  end

  -- if name == "Dark_Offer_Trigger" then
  --   dark_offer_spawn = Script.GetSpawnPointsMatching("Dark_Offer_Start")
  --   if store.Ch03c.choice_a == "Cultists" then 
  --     Script.SpawnEntitySomewhereInSpawnPoints("Dark Emmisary")
  --   else
  --     Script.SpawnEntitySomewhereInSpawnPoints("Chosen One")
  --   end
  --   store.Ch03d.Spawnpoints_complete["Ch03_Dark_Offer_Trigger"]
  -- end

  -- if name == "Dark_Offer_Finale" then
  --   if store.Ch03c.choice_a == "Cultists" then
  --     Script.DialogBox("ui/dialog/Ch03/Ch03_Dark_Offer_Finale_Cultists.json")
  --     store.Ch03d.choice_a = choices[1]
  --   end
  --   if store.Ch03c.choice_a == "Nightmares" then   
  --     Script.DialogBox("ui/dialog/Ch03/Ch03_Dark_Offer_Finale_Nightmares.json")
  --     store.Ch03d.choice_a = choices[1]
  --   end
  -- end
end


function RoundEnd(intruders, round)
  print("end", intruders, round)
end



