function Init()
  if not store.Ch04b then
    store.Ch04b = {}
  end
  store.Ch04b.Spawnpoints_complete{}
  store.Ch04b.Spawnpoints = {
    "Timer_Start",
    "Helpful_Devotee01",
    "Helpful_Devotee02",
    "Helpful_Devotee03",
    "Penelope_Escapes",
    "Penelope_Caught",
    "Hector_Escapes",
    "Hector_Caught",
    "Elizabeth_Escapes",
    "Elizabeth_Caught",
    "Gottlieb_Escapes",
    "Gottlieb_Caught",
    "Simon_Escapes",
    "Simon_Caught",
    "Antonia_Escapes",
    "Antonia_Caught",
    "First_Escape",
    "Second_Escape",
    "Third_Escape",
    "One_Done",
    "Two_Done",
    "Three_Done"
   } 

  Script.LoadHouse("Chapter_04_b")
  Script.DialogBox("ui/dialog/Ch04/Ch04_Intro_Part2.json") 

  if store.Ch04a.choice_a[2] == "Penelope" then
    hector_spawn = Script.GetSpawnPointsMatching("Hector-Start")
    Script.SpawnEntitySomewhereInSpawnPoints ("Hector", hector_spawn)
    evil_penelope_spawn = Script.GetSpawnPointsMatching("Evil-Penelope-Start")
    Script.SpawnEntitySomewhereInSpawnPoints("Penelope Transformed", evil_penelope_spawn)
  end

  if store.Ch04a.choice_a[2] == "Hector" then
    penelope_spawn = Script.GetSpawnPointsMatching("Penelope-Start")
    Script.SpawnEntitySomewhereInSpawnPoints ("Penelope", penelope_spawn)
    evil_hector_spawn = Script.GetSpawnPointsMatching("Evil-Hector-Start")
    Script.SpawnEntitySomewhereInSpawnPoints("Hector Transformed", evil_hector_spawn)
  end

  if store.Ch04a.choice_a[3] == "Elizabeth" then
    simon_spawn = Script.GetSpawnPointsMatching("Simon-Start")
    Script.SpawnEntitySomewhereInSpawnPoints ("Simon", simon_spawn)
    evil_Elizabeth_spawn = Script.GetSpawnPointsMatching("Evil-Elizabeth-Start")
    Script.SpawnEntitySomewhereInSpawnPoints("Elizabeth Transformed", evil_Elizabeth_spawn)
  end

  if store.Ch04a.choice_a[3] == "Simon" then
    Elizabeth_spawn = Script.GetSpawnPointsMatching("Elizabeth-Start")
    Script.SpawnEntitySomewhereInSpawnPoints ("Elizabeth", Elizabeth_spawn)
    evil_Simon_spawn = Script.GetSpawnPointsMatching("Evil-Simon-Start")
    Script.SpawnEntitySomewhereInSpawnPoints("Simon Transformed", evil_hector_spawn)
  end

  if store.Ch04a.choice_a[1] == "Antonia" then
    Gottlieb_spawn = Script.GetSpawnPointsMatching("Gottlieb-Start")
    Script.SpawnEntitySomewhereInSpawnPoints ("Gottlieb", Gottlieb_spawn)
    evil_Antonia_spawn = Script.GetSpawnPointsMatching("Evil-Antonia-Start")
    Script.SpawnEntitySomewhereInSpawnPoints("Antonia Transformed", evil_Antonia_spawn)
  end

  if store.Ch04a.choice_a[1] == "Gottlieb" then
    Antonia_spawn = Script.GetSpawnPointsMatching("Antonia-Start")
    Script.SpawnEntitySomewhereInSpawnPoints ("Antonia", Antonia_spawn)
    evil_Gottlieb_spawn = Script.GetSpawnPointsMatching("Evil-Gottlieb-Start")
    Script.SpawnEntitySomewhereInSpawnPoints("Gottlieb Transformed", evil_hector_spawn)
  end

  tyree_spawn = Script.GetSpawnPointsMatching("Tyree-Start")
  Script.SpawnEntitySomewhereInSpawnPoints("Dr. Elias Tyree", tyree_spawn)

  tyree_servitors_spawn = Script.GetSpawnPointsMatching("Servitors-Start")
  Script.SpawnEntitySomewhereInSpawnPoints("Devoted Servant")
  Script.SpawnEntitySomewhereInSpawnPoints("Devoted Servant")
  Script.SpawnEntitySomewhereInSpawnPoints("Devoted Servant")

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
    name = IsPosInUnusedSpawnpoint(pos, store.Ch04b.Spawnpoints, store.Ch04b.Spawnpoints_complete)
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

function OnDenizensAction(intruders, round, exec)
 name = IsPosInUnusedSpawnpoint(exec.Ent.Pos, store.Ch04b.Spawnpoints, store.Ch04b.Spawnpoints_complete)
 
if name == "Timer_Start" then
  Script.DialogBox("ui/dialog/Ch04/Ch04_Timer_Starts.json")
  store.Ch04.Spawnpoints_complete["Timer_Start"] = true
  --Start Timer
end

if name == "Helpful_Devotee01" then
  Script.DialogBox("ui/dialog/Ch04/Ch04_Helpful_Devotee01.json")
  store.Ch04.Spawnpoints_complete["Helpful_Devotee01"] = true
  tyree_servitors_spawn = Script.GetSpawnPointsMatching("Servitors-Start2")
  Script.SpawnEntitySomewhereInSpawnPoints("Devoted Servant")
end

if name == "Helpful_Devotee02" then
  Script.DialogBox("ui/dialog/Ch04/Ch04_Helpful_Devotee02.json")
  store.Ch04.Spawnpoints_complete["Helpful_Devotee02"] = true
  tyree_servitors_spawn = Script.GetSpawnPointsMatching("Servitors-Start3")
  Script.SpawnEntitySomewhereInSpawnPoints("Devoted Servant")
end

if name == "Helpful_Devotee03" then
  Script.DialogBox("ui/dialog/Ch04/Ch04_Helpful_Devotee03.json")
  store.Ch04.Spawnpoints_complete["Helpful_Devotee03"] = true
  tyree_servitors_spawn = Script.GetSpawnPointsMatching("Servitors-Start4")
  Script.SpawnEntitySomewhereInSpawnPoints("Devoted Servant")
end

if exec.Ent.Name("Penelope") and ent.HpCur == 0 then
  store.Ch04b.Spawnpoints_complete["Penelope_Caught"] = true
  store.Ch04b.Spawnpoints_complete["One_Done"] = true
  Script.DialogBox("ui/dialog/Ch04/Ch04_Penelope_Caught.json")
end 

if exec.Ent.Name("Hector") and ent.HpCur == 0 then
  store.Ch04b.Spawnpoints_complete["Hector_Caught"] = true
  store.Ch04b.Spawnpoints_complete["One_Done"] = true
  Script.DialogBox("ui/dialog/Ch04/Ch04_Hector_Caught.json")
end 

if exec.Ent.Name("Elizabeth") and ent.HpCur == 0 then
  store.Ch04b.Spawnpoints_complete["Elizabeth_Caught"] = true
  store.Ch04b.Spawnpoints_complete["Two_Done"] = true
  Script.DialogBox("ui/dialog/Ch04/Ch04_Elizabeth_Caught.json")
end 

if exec.Ent.Name("Simon") and ent.HpCur == 0 then
  store.Ch04b.Spawnpoints_complete["Simon_Caught"] = true
  store.Ch04b.Spawnpoints_complete["Two_Done"] = true
  Script.DialogBox("ui/dialog/Ch04/Ch04_Simon_Caught.json")
end 

if exec.Ent.Name("Antonia") and ent.HpCur == 0 then
  store.Ch04b.Spawnpoints_complete["Antonia_Caught"] = true
  store.Ch04b.Spawnpoints_complete["Three_Done"] = true
  Script.DialogBox("ui/dialog/Ch04/Ch04_Antonia_Caught.json")
end 

if exec.Ent.Name("Gottlieb") and ent.HpCur == 0 then
  store.Ch04b.Spawnpoints_complete["Gottlieb_Caught"] = true
  store.Ch04b.Spawnpoints_complete["Three_Done"] = true
  Script.DialogBox("ui/dialog/Ch04/Ch04_Gottlieb_Caught.json")
end 


function OnIntrudersAction(intruders, round, exec)
 name = IsPosInUnusedSpawnpoint(exec.Ent.Pos, store.Ch04b.Spawnpoints, store.Ch04b.Spawnpoints_complete)
 
end



function RoundEnd(intruders, round)

  if round == 5 then
    if store.Ch04b.Spanwpoints_Complete["Penelope_Caught"] or store.Ch04b.Spanwpoints_Complete["Hector_Caught"] then
      return
    end
  
    else 
      store.Ch04b.Spawnpoints_complete["First_Escape"] = true
      if store.Ch04a.choice_a[2] == "Penelope" then
        Script.DialogBox("ui/dialog/Ch04/Ch04_Penelope_Escapes.json")
        store.Ch04b.Spawnpoints_complete["Penelope_Escapes"] = true
        store.Ch04b.Spawnpoints_complete["One_Done"] = true
        for _, ent in pairs(Script.GetAllEnts()) do
          if ent.Name == "Penelope" then
              Script.SetHp(ent, 0)
            end
          end
        end
      end

     if store.Ch04a.choice_a[2] == "Hector" then
       Script.DialogBox("ui/dialog/Ch04/Ch04_Hector_Escapes.json")
       store.Ch04b.Spawnpoints_complete["Hector_Escapes"] = true
       store.Ch04b.Spawnpoints_complete["One_Done"] = true
       for _, ent in pairs(Script.GetAllEnts()) do
          if ent.Name == "Hector" then
              Script.SetHp(ent, 0)
            end
          end
        end
      end
    end
  end

  if round == 10 then
    if store.Ch04b.Spanwpoints_Complete["Elizabeth_Caught"] or store.Ch04b.Spanwpoints_Complete["Simon_Caught"] then
      return
    end
  
    else
      store.Ch04b.Spawnpoints_complete["Second_Escape"] = true
      if store.Ch04a.choice_a[3] == "Elizabeth" then
       Script.DialogBox("ui/dialog/Ch04/Ch04_Elizabeth_Escapes.json")
       store.Ch04b.Spawnpoints_complete["Elizabeth_Escapes"] = true
       store.Ch04b.Spawnpoints_complete["Two_Done"] = true
       for _, ent in pairs(Script.GetAllEnts()) do
          if ent.Name == "Elizabeth" then
              Script.SetHp(ent, 0)
            end
          end
        end
      end

      if store.Ch04a.choice_a[3] == "Simon" then
       Script.DialogBox("ui/dialog/Ch04/Ch04_Simon_Escapes.json")
       store.Ch04b.Spawnpoints_complete["Simon_Escapes"] = true
       store.Ch04b.Spawnpoints_complete["Two_Done"] = true
       for _, ent in pairs(Script.GetAllEnts()) do
          if ent.Name == "Simon" then
              Script.SetHp(ent, 0)
            end
          end
        end
      end
    end
  end

  if timer == 15 then
    if store.Ch04b.Spanwpoints_Complete["Antonia_Caught"] or store.Ch04b.Spanwpoints_Complete["Gottlieb_Caught"] then
      return
    end
  
    else 
      store.Ch04b.Spawnpoints_complete["Third_Escape"] = true

      if store.Ch04a.choice_a[1] == "Antonia" then
       Script.DialogBox("ui/dialog/Ch04/Ch04_Antonia_Escapes.json")
       store.Ch04b.Spawnpoints_complete["Antonia_Escapes"] = true
       store.Ch04b.Spawnpoints_complete["Three_Done"] = true
       for _, ent in pairs(Script.GetAllEnts()) do
          if ent.Name == "Antonia" then
              Script.SetHp(ent, 0)
            end
          end
        end
      end

      if store.Ch04a.choice_a[1] == "Gottlieb" then
       Script.DialogBox("ui/dialog/Ch04/Ch04_Gottlieb_Escapes.json")
       store.Ch04b.Spawnpoints_complete["Gottlieb_Escapes"] = true
       store.Ch04b.Spawnpoints_complete["Three_Done"] = true
       for _, ent in pairs(Script.GetAllEnts()) do
          if ent.Name == "Gottlieb" then
              Script.SetHp(ent, 0)
            end
          end
        end
      end
    Script.DialogBox("ui/dialog/Ch04/Ch04_All_Escape.json")
    Script.StartScript("Ch04c.lua")
  end

  if store.Ch04b.Spawnpoints_complete["One_Done"] 
    and store.Ch04b.Spawnpoints_complete["Two_Done"]
    and store.Ch04b.Spawnpoints_complete["Three_Done"] then
    Script.DialogBox("ui/dialog/Ch04/Ch04_Part2_Finale.json")
    Script.StartScript("Ch04c.lua")
  end
end