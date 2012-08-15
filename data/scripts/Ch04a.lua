function Init()
  if not store.Ch04a then
  store.Ch04a = {}
  end
   store.Ch04a.Spawnpoints_complete{}
   store.Ch04a.Spawnpoints = {
    "Penelope-Talk",
    "Simon-Talk",
    "Antonia-Talk",
    "Gottlieb-Talk",
    "Elizabeth-Talk",
    "Hector-Talk"      
   } 

  Script.LoadHouse("Chapter_04_a")
  Script.DialogBox("ui/dialog/Ch04a/Ch04a_Intro.json") 

  penelope_spawn = Script.GetSpawnPointsMatching("Penelope-Start")
  Script.SpawnEntitySomewhereInSpawnPoints ("Penelope", penelope_spawn)
  Script.BindAi("Penelope", "idle.lua")

  simon_spawn = Script.GetSpawnPointsMatching("Simon-Start")
  Script.SpawnEntitySomewhereInSpawnPoints ("Simon", simon_spawn)
  Script.BindAi("Simon", "idle.lua")


  antonia_spawn = Script.GetSpawnPointsMatching("Antonia-Start")
  Script.SpawnEntitySomewhereInSpawnPoints ("Antonia", antonia_spawn)
  Script.BindAi("Antonia", "idle.lua")

  gottlieb_spawn = Script.GetSpawnPointsMatching("Gottlieb-Start")
  Script.SpawnEntitySomewhereInSpawnPoints ("Gottlieb", gottlieb_spawn)
  Script.BindAi("Gottlieb", "idle.lua")

  elizabeth_spawn = Script.GetSpawnPointsMatching("Elizabeth-Start")
  Script.SpawnEntitySomewhereInSpawnPoints ("Elizabeth", penelope_spawn)
  Script.BindAi("Elizabeth", "idle.lua")

  hector_spawn = Script.GetSpawnPointsMatching("Hector-Start")
  Script.SpawnEntitySomewhereInSpawnPoints ("Hector", penelope_spawn)
  Script.BindAi("Elizabeth", "idle.lua")

  tyree_spawn = Script.GetSpawnPointsMatching("Tyree-Start")
  Script.SpawnEntitySomewhereInSpawnPoints("Dr. Elias Tyree", tyree_spawn)

  Script.BindAi("denizen", "human")
  Script.BindAi("minions", "minions.lua")
  Script.BindAi("intruder", "intruders.lua")
    --always bind one to human!

end


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


function IsPosInUnusedSpawnpoint(pos, list, used)
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
    name = IsPosInUnusedSpawnpoint(pos, store.Ch04a.Spawnpoints, store.Ch04a.Spawnpoints_complete)
    if name then
      return i
    end
  end
  return table.getn(path)
end

--MAKE NO INTRUDERS TURNS!!!

function OnAction(intruders, round, exec)
  if not exec.Ent.Side.Intruder then
    return
  end

function OnDenizensAction(intruders, round, exec)
 name = IsPosInUnusedSpawnpoint(exec.Ent.Pos, store.Ch04a.Spawnpoints, store.Ch04a.Spawnpoints_complete)
   
   if name == "Penelope-Talk" then
    Script.DialogBox("ui/dialog/Ch04/Ch04_Penelope_Talk.json")
    store.Ch04.Spawnpoints_complete["Ch04_Penelope_Talk"]
  end

  if name == "Simon-Talk" then
    Script.DialogBox("ui/dialog/Ch04/Ch04_Simon_Talk.json")
    store.Ch04.Spawnpoints_complete["Ch04_Simon_Talk"]
  end

  if name == "Hector-Talk" then
    Script.DialogBox("ui/dialog/Ch04/Ch04_Hector_Talk.json")
    store.Ch04.Spawnpoints_complete["Ch04_Hector_Talk"]
  end

  if name == "Antonia-Talk" then
    Script.DialogBox("ui/dialog/Ch04/Ch04_Antonia_Talk.json")
    store.Ch04.Spawnpoints_complete["Ch04_Antonia_Talk"]
  end

  if name == "Elizabeth-Talk" then
    Script.DialogBox("ui/dialog/Ch04/Ch04_Elizabeth_Talk.json")
    store.Ch04.Spawnpoints_complete["Ch04_Elizabeth_Talk"]
  end

  if name == "Gottlieb-Talk" then
    Script.DialogBox("ui/dialog/Ch04/Ch04_Gottlieb_Talk.json")
    store.Ch04.Spawnpoints_complete["Ch04_Gottlieb_Talk"]
  end
end




function OnIntrudersAction(intruders, round, exec)
 name = IsPosInUnusedSpawnpoint(exec.Ent.Pos, store.Ch04a.Spawnpoints, store.Ch04a.Spawnpoints_complete)
   if store.Ch04.Spawnpoints_complete["Penelope-Talk"] 
    and store.Ch04.Spawnpoints_complete["Simon-Talk"] 
    and store.Ch04.Spawnpoints_complete["Antonia-Talk"]
    and store.Ch04.Spawnpoints_complete["Gottlieb-Talk"]
    and store.Ch04.Spawnpoints_complete["Elizabeth-Talk"]
    and store.Ch04.Spawnpoints_complete["Hector-Talk"] then
    Script.DialogBox("ui/dialog/Ch04/Ch04_choices.json")
  --IS THIS RIGHT? For storing all three choices in the dialog?
    store.Ch04a.choice_a = choices
    Script.StartScript("Ch04b.lua")
end
function RoundEnd(intruders, round)
  print("end", intruders, round)
end

