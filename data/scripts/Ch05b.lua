
function Init()
  if not store.Ch05b then
    store.Ch05b = {}
  end
  store.Ch05b.Spawnpoints_complete={}
  store.Ch05b.Spawnpoints = {
    "Intruder_Start",
    "Zombie2_Start",
    "Zombie3_Start",  
    "Blood_Comment",
    "Oh_Look_A_Shrieker",  
    "Trigger_Shrieker",
    "Shrieker_Start",   
    "Trigger_Horde1",
    "Oh_Look_Zombies",
    "Zombie_Start",
    "Zombie_Hole", 
    "Boss_Conversation",    
  } 


  Script.LoadHouse("Chapter_05_b")
    
  Script.BindAi("denizen", "denizens.lua")
  Script.BindAi("minions", "minions.lua")
  Script.BindAi("intruder", "human")
 
  intruder_spawn = Script.GetSpawnPointsMatching("Intruder_Start")

  shrieker_spawn = Script.GetSpawnPointsMatching("Shrieker_Start")
  Script.SpawnEntitySomewhereInSpawnPoints("Shrieker", shrieker_spawn)  

  SetIntruder()
  ent = Script.SpawnEntitySomewhereInSpawnPoints (IntruderName, intruder_spawn)
  Script.SelectEnt(ent)

  Script.DialogBox("ui/dialog/Ch05/ch05_" .. ShortName .. "_Start_Conversation.json")
  store.Ch05b.Spawnpoints_complete["Intruder_Start"] = true  

  TotalKillCount = 0
  bTriggerHorde = false
end

function setLosModeToRoomsWithSpawnsMatching(side, pattern)
  sp = Script.GetSpawnPointsMatching(pattern)
  rooms = {}
  for i, spawn in pairs(sp) do
    rooms[i] = Script.RoomAtPos(spawn.Pos)
  end
  Script.SetLosMode(side, rooms)
end

-- play_as_denizens = false

function SetIntruder()
  --IntruderName = store.Ch05a.choice_a[1]
  IntruderName = "Timothy K."
  ShortName = string.sub(IntruderName, 1, -4)
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
  for _, spawn in pairs(list) do
    if not used[spawn] and pointIsInSpawns(pos, spawn) then
      return spawn
    end
  end
  return nil
end

function IsNamedSpawnpointUsed(list, used)
  for _, spawn in pairs(list) do
    if not used[spawn] and pointIsInSpawns(pos, spawn) then
      return spawn
    end
  end
  return nil
end

function IsPosInNamedSpawnpoint(pos, list)
  for _, spawn in pairs(list) do

    --these spawn points will never get marked used, so we don't care whether it's complete.   
    if pointIsInSpawns(pos, spawn.Name) then
      return spawn
    end
  end
  return nil
end

function OnMove(ent, path)
  if not ent.Side.Intruder then

    if ent.Conditions["Off-Balance"] then
      Script.SetHp(ent, 11)
    end

    name = IsPosInNamedSpawnpoint(ent.Pos, Script.GetSpawnPointsMatching("Zombie_Hole"))
    if name then
      --this is a zombie and he is near a hole or trap.  If he is off-balance, then he falls in.
      if ent.Conditions["Off-Balance"] then
        Script.PlayAnimations(ent, {"defend", "killed"})
        Script.SetHp(ent, 0)
      end
    end
  else
    --intruders.  want to stop them on trigger spawnpoints, but not on holes.
    holes =  Script.GetSpawnPointsMatching("Zombie_Hole")    
    for i, pos in pairs(path) do
      in_a_spawn = IsPosInUnusedSpawnpoint(pos, store.Ch05b.Spawnpoints, store.Ch05b.Spawnpoints_complete)
      in_a_hole = IsPosInNamedSpawnpoint(pos, holes)
      if in_a_spawn and not in_a_hole then  --in a spawnpoint
        return i 
          --this stops them, if we don't stop them, then we need to store that it's true.
          --     store.Ch05b.Spawnpoints_complete["Ch05_Dialog04"] = true
        -- end
      end
    end
  end
  return table.getn(path)
end


function OnAction(intruders, round, exec)
  if not exec.Ent.Side.Intruder then
    return
  end
  name = IsPosInUnusedSpawnpoint(exec.Ent.Pos, store.Ch05b.Spawnpoints, store.Ch05b.Spawnpoints_complete)
 
  --ALL of our dialogues here will be character specific based on a choise in the previous section.
  if name == "Oh_Look_Zombies" then
    Script.DialogBox("ui/dialog/Ch05/ch05_" .. ShortName .. "_Oh_Look_Zombies.json")
    store.Ch05b.Spawnpoints_complete["Oh_Look_Zombies"] = true

    SpawnZombies("Zombie_Start")
  end

  if name == "Trigger_Horde1" then
    Script.DialogBox("ui/dialog/Ch05/ch05_" .. ShortName .. "_Trigger_Horde1.json")
    store.Ch05b.Spawnpoints_complete["Trigger_Horde1"] = true
    SpawnZombies("Zombie2_Start")
    bTriggerHorde = true
  end

  if name == "Blood_Comment" then
    Script.DialogBox("ui/dialog/Ch05/ch05_" .. ShortName .. "_Blood_Comment.json")
    store.Ch05b.Spawnpoints_complete["Blood_Comment"] = true
  end  

  if name == "Oh_Look_A_Shrieker" then
    Script.DialogBox("ui/dialog/Ch05/ch05_" .. ShortName .. "_Oh_Look_A_Shrieker.json")
    store.Ch05b.Spawnpoints_complete["Oh_Look_A_Shrieker"] = true
  end

  if name == "Trigger_Shrieker" then
    TriggerShrieker()
  end

  if name == "Boss_Conversation" then
    Script.DialogBox("ui/dialog/Ch05/ch05_" .. ShortName .. "_Boss_Conversation.json")
    store.Ch05b.Spawnpoints_complete["Boss_Conversation"] = true
    shrieker_spawn = Script.GetSpawnPointsMatching("Boss_Start")
    Script.SpawnEntitySomewhereInSpawnPoints("Shrieker", shrieker_spawn)
  end
  
  --If the action was an attack, let's see if any zombies were hit and apply "Off-Balance" to them
  if exec.Action.Type == "Basic Attack" then
    -- for _, ent in pairs(Script.GeStAllEnts()) do
    --   --We'll check whether a zombie falls in a hole after both a player action and after a zombie move.
    --   --If this changes, please also verify the other check.
    if exec.Target.Name == "Zombie" then
      --this is a zombie entity, and we just attacked it.  Apply a off-balance debuff
      Script.SetCondition(exec.Target, "Off-Balance", true)

      name = IsPosInNamedSpawnpoint(exec.Target.Pos, Script.GetSpawnPointsMatching("Zombie_Hole"))
      if name then
        --this is a zombie and he is near a hole or trap.  If he is off-balance, then he falls in.
        if exec.Target.Conditions["Off-Balance"] then
          Script.PlayAnimations(exec.Target, {"defend", "killed"})
          Script.SetHp(exec.Target, 0)

           TotalKillCount = TotalKillCount + 1
          --dialogue that happens on zombie killing goes here.
          if TotalKillCount == 1 then   
            Script.DialogBox("ui/dialog/Ch05/ch05_" .. ShortName .. "_Trap_Sprung.json")
          end 

        end
      end
    end
  
    if exec.Target.Name == "Shrieker" then
      shrieker_trigger = Script.GetSpawnPointsMatching("Trigger_Shrieker")
      boss_trigger = Script.GetSpawnPointsMatching("Boss_Start")
      if not store.Ch05b.Spawnpoints_complete[shrieker_trigger] and not store.Ch05b.Spawnpoints_complete[boss_trigger] then
        --they attacked the shrieker before stepping onto the trigger.
        TriggerShrieker()   
      end
    end
  end  
end

function TriggerShrieker()
  Script.DialogBox("ui/dialog/Ch05/ch05_" .. ShortName .. "_Trigger_Shrieker.json")
  store.Ch05b.Spawnpoints_complete["Trigger_Shrieker"] = true
  SpawnZombies("Zombie3_Start")
  
  for _, ent in pairs(Script.GetAllEnts()) do
    if ent.Name == "Shrieker" then
      Script.PlayAnimations(ent, {"defend", "killed"})
      Script.SetHp(ent, 0)
    end
  end
end

function SpawnZombies(spawnName)
  spawn = Script.GetSpawnPointsMatching(spawnName)
  for i = 1, 10 do  --Never more than 10 spawns of the same name      
    Script.SpawnEntitySomewhereInSpawnPoints("Zombie", spawn)
  end 
end



function RoundEnd(intruders, round)
  if not intruders then
    -- once the horde gets triggered
    --if store.Ch05b.Spawnpoints_complete[horde_trigger] then
    if bTriggerHorde then  --basing this off of the spawn being completed wasn't working... 
      SpawnZombies("Zombie_Start")
    end   
  end
--SOMETHING FOR GAME END
-- WIN and LOSE
end