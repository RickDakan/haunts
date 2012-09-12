function setLosModeToRoomsWithSpawnsMatching(side, pattern)
  sp = Script.GetSpawnPointsMatching(pattern)
  rooms = {}
  for i, spawn in pairs(sp) do
    rooms[i] = Script.RoomAtPos(spawn.Pos)
  end
  Script.SetLosMode(side, rooms)
end

function IsStoryMode()
  return true
end

function DoTutorials()
  --We should totally do some tutorials here.
  --It would be super cool.
end

function Init(data)
  side_choices = Script.ChooserFromFile("ui/start/versus/side.json")

  -- check data.map == "random" or something else
  Script.LoadHouse("Lvl_03_Sanitorium")
  Script.PlayMusic("Haunts/Music/Adaptive/Bed 1")   

  store.side = side_choices[1]
  if store.side == "Humans" then
    Script.BindAi("denizen", "human")
    Script.BindAi("minions", "minions.lua")
    Script.BindAi("intruder", "human")
  end
  if store.side == "Denizens" then
    Script.BindAi("denizen", "human")
    Script.BindAi("minions", "minions.lua")
    Script.BindAi("intruder", "intruders.lua")
  end
  if store.side == "Intruders" then
    Script.BindAi("denizen", "denizens.lua")
    Script.BindAi("minions", "minions.lua")
    Script.BindAi("intruder", "human")
  end

  --spawn the objective point and the fake op points
  --!!!!
  waypoint_spawn = Script.GetSpawnPointsMatching("CrashTest")
  --waypoint_spawn = Script.GetSpawnPointsMatching("Relic_Spawn")
  objSpawn = Script.SpawnEntitySomewhereInSpawnPoints("Antidote", waypoint_spawn, false)

  store.Waypoint = objSpawn
  Script.SetWaypoint("Waypoint1" , "intruders", store.Waypoint.Pos, 3)

  --decoys
  objSpawn = Script.SpawnEntitySomewhereInSpawnPoints("Antidote", waypoint_spawn, false)    
  objSpawn = Script.SpawnEntitySomewhereInSpawnPoints("Antidote", waypoint_spawn, false)    
  objSpawn = Script.SpawnEntitySomewhereInSpawnPoints("Antidote", waypoint_spawn, false)    

  --Spawn the patients
  patient_spawn = Script.GetSpawnPointsMatching("Patient_Start")
  for i = 1, 10, 1 do  
    --this will cause some "cannot find place to spawn" errors in the log.  
    --But covers us in case we decide to mess with the number of patients.  I'm ok with it.
    Script.SpawnEntitySomewhereInSpawnPoints("Patient", patient_spawn, false)
  end 

  --we will incorporate some randomness here.
  math.randomseed(os.time())

  --set these modular variables.
  store.nIntrudersFound = 0
  store.nMonstersFound = 0
  store.bWaypointDown = false
  store.bFloodStarted = false
  store.bShiftChange = false

  --We can't kill objects, so we just have to move the patients to a hidden room.
  --We also can't move them all to the same place, so we'll need to gimmick the "Despawn"
  --with a counter
  store.DeathCounter = 0
end

function intrudersSetup()

  if IsStoryMode() then
    intruder_names = {"Ghost Hunter"}
    intruder_spawn = Script.GetSpawnPointsMatching("Intruder_Start1")
  -- else
  --   --permit all choices for normal vs play
  end 

  for _, name in pairs(intruder_names) do
    ent = Script.SpawnEntitySomewhereInSpawnPoints(name, intruder_spawn, false)
  end

  -- Choose entry point here.
  Script.SaveStore()
end

function denizensSetup()
  Script.SetVisibility("denizens")

  master_spawn = Script.GetSpawnPointsMatching("Master_.*")
  Script.SpawnEntitySomewhereInSpawnPoints("Administrator", master_spawn, false)

  -- placed is an array containing all of the entities placed, in this case
  -- there will only be one, and we will use that one to determine what
  -- servitors to make available to the user to place.
  MasterName = "Administrator"
  ServitorEnts = 
  {
    {"Attendant", 2},
    {"Technician", 3},
  }  

  -- Just like before the user gets a ui to place these entities, but this
  -- time they can place more, and this time they go into spawn points that
  -- match anything with the prefix "Servitor_".
  setLosModeToRoomsWithSpawnsMatching("denizens", "Servitors_Start")
  placed = Script.PlaceEntities("Servitors_Start", ServitorEnts, 0,10)
end

function RoundStart(intruders, round)
  side = {Intruder = intruders, Denizen = not intruders, Npc = false, Object = false}

  if round == 1 then
    if intruders then
      intrudersSetup()     
    else
      Script.FocusPos(Script.GetSpawnPointsMatching("Master_Start")[1].Pos)
      Script.DialogBox("ui/dialog/Lvl03/Lvl_03_Opening_Denizens.json")
      denizensSetup()
    end
    Script.SetLosMode("intruders", "blind")
    Script.SetLosMode("denizens", "blind")

    if IsStoryMode() then
      DoTutorials()
    end

    Script.EndPlayerInteraction()

    return
  end

  if store.bFloodStarted then
    --At the start of each denizen turn, we're going to randomly spawn attendants at 
    --the flood points.  Also need to prevent spawning within LoS of an intruder.

    available_spawns = Script.GetSpawnPointsMatching("Flood_Point")
    if TotalDeniCount() < 20 and intruders then
      for i = 1, 3, 1 do
        --Pick an entity    

        --JONATHAN - The orderlies spawn at the start of the intruders turn and then move
        --on the next deni turn.  That's when the crash happens. 
        ent = Script.SpawnEntitySomewhereInSpawnPoints("Orderly", available_spawns, true)
        Script.SetAp(ent, 0)
      end
    end
  end

  if store.bShiftChange and not intruders then
    store.bShiftChange = false  

    if ValueForReinforce() > 1 then
      for i = 1,  math.floor((ValueForReinforce()/2) + .5) , 1 do
        --Pick an entity
        if math.random(4) > 2 then
          floodEnt = ServitorEnts[1][1]
        else
          floodEnt = ServitorEnts[2][1]
        end     
        Script.SpawnEntitySomewhereInSpawnPoints(floodEnt, Script.GetSpawnPointsMatching("Servitors_Start"), true)
      end
    end 
  end

  store.game = Script.SaveGameState()
  SelectCharAtTurnStart(side)
  if store.side == "Humans" then
    Script.SetLosMode("intruders", "entities")
    Script.SetLosMode("denizens", "entities")
    if intruders then
      Script.SetVisibility("intruders")
    else
      Script.SetVisibility("denizens")
    end
    Script.ShowMainBar(true)
  else
    Script.ShowMainBar(intruders == (store.side == "Intruders"))
  end
end

function SelectCharAtTurnStart(side)
  bDone = false
  if LastIntruderEnt then
    if side.Intruder then
      Script.SelectEnt(LastIntruderEnt)
      bDone = true
    end
  end  
  if LastDenizenEnt and not bDone then
    if side.Denizen then     
      Script.SelectEnt(LastDenizenEnt)
      bDone = true
    end  
  end   

  if not bDone then
    --select the dood with the most AP
    Script.SelectEnt(GetEntityWithMostAP(side))
  end  
end

function GetEntityWithMostAP(side)
  entToSelect = nil
  for _, ent in pairs(Script.GetAllEnts()) do
    if (ent.Side.Intruder and side.Intruder) or (ent.Side.Denizen and side.Denizen) then   
      if entToSelect then    
        if entToSelect.ApCur < ent.ApCur then      
          entToSelect = ent
        end 
      else
        --first pass.  select this one.
        entToSelect = ent
      end
    end
  end
  return entToSelect
end

function OnMove(ent, path)
  return table.getn(path)
end

function OnAction(intruders, round, exec)
  -- Check for players being dead here
  if store.execs == nil then
    store.execs = {}
  end
  store.execs[table.getn(store.execs) + 1] = exec

  if exec.Ent.Side.Intruder and GetDistanceBetweenEnts(exec.Ent, store.Waypoint) <= 3 and not store.bWaypointDown then
    --The intruders got to the first waypoint.
    store.bWaypointDown = true

    StoreWaypoint("Waypoint", "", "", "", true)
    StoreWaypoint("Waypoint", "intruders", Script.GetSpawnPointsMatching("Waypoint2")[1].Pos, 3, false)  

    StoreCondition("Carrying Antidote", exec.Ent, true)
    doCondition(condition_exec)

    StoreGear("Antidote", exec.Ent)

    if not store.bFloodStarted then    
      --The denizens have not yet activated the alarm.
      store.bFloodStarted = true
      Script.DialogBox("ui/dialog/Lvl03/Lvl_03_Waypoint_And_Alarm_Intruders.json")
      Script.SetMusicParam("tension_level", 0.9) 
      store.bToldIntrudersAboutAlarm = true
      for i = 1, 3, 1 do
        --Pick an entity

        --JONATHAN - Uncomment these two lines to see a crash when the waypoint is first activated.    
        -- ent = Script.SpawnEntitySomewhereInSpawnPoints("Orderly", available_spawns, true)
        -- Script.SetAp(ent, 0)
      end
    else
      --Denizens already started the alarm.  Just tell the intruders about the waypoint.
      Script.DialogBox("ui/dialog/Lvl03/Lvl_03_Waypoint_Only_Intruders.json")
    end
  end 

  if exec.Action.Name == "Hand Antidote" then
    StoreGear("", exec.Ent) 
    StoreGear("Antidote", exec.Target)     
    --remove the carrying condition from the exec ent.  The target will get the
    --condition b/c of the ability.
    StoreCondition("Carrying Antidote", exec.Ent, false)
    doCondition(condition_exec)
  end

  --The big check: Have the intruders won.
  if store.bWaypointDown then
    if  exec.Ent.Side.Intruder then
      if pointIsInSpawns(exec.Ent.Pos, "Escape") then
        if exec.Ent.Conditions["Carrying Antidote"] then
          --The intruders got to the exit with the Antidote.  Game over.
          Script.DialogBox("ui/dialog/Lvl03/Lvl_03_Victory_Intruders.json")    
        end
      end
    end
  end  

  --have the intruders attempted to rescue a patient?
  if  exec.Ent.Side.Intruder then
    patientToActivate = nil
    patientToActivate = EntIsNextToPatient(exec.Ent)
    if patientToActivate then
      SpawnedEnt = SpawnIntruderOrMonster(patientToActivate)
    end
  end 

  --deni's win when all intruders dead.
  if not AnyIntrudersAlive() then
    Script.DialogBox("ui/dialog/Lvl03/Lvl_03_Victory_Denizens.json")
  end 

  --if the deni master used Signal Shift Change, permit spawning and end the turn.
  if exec.Action.Name == "Signal Shift Change" then
    store.bShiftChange = true
  end  

  --if the deni master spotted the intruders, show dialogue
  if exec.Action.Name == "Identify Escapee" then

    -- Can only be used on an intruder, so we don't need to check the target.
    -- Once used, the flood starts.
    -- bInstrudersSpotted = true

    --Forcing the master to escape took too long.  Now if he spots the intruders,
    --he can sound the alarm immediately
    store.bFloodStarted = true
    Script.DialogBox("ui/dialog/Lvl03/Lvl_03_Alarm_Started_Denizens.json")
    store.bToldDenizensAboutFloodStart = true
  end 

  --after any action, if this ent's Ap is 0, we can select the next ent for them
  -- if exec.Ent.ApCur == 0 then
  --   nextEnt = GetEntityWithMostAP(exec.Ent.Side)
  --   if nextEnt.ApCur > 0 then
  --     Script.SelectEnt(nextEnt)
  --   end
  -- end
end
 

function RoundEnd(intruders, round)
  if round == 1 then
    return
  end

  bSkipOtherChecks = false  --Resets this every round

  if store.side == "Humans" then
    Script.ShowMainBar(false)
    Script.SetLosMode("intruders", "blind")
    Script.SetLosMode("denizens", "blind")
    if intruders then
      Script.SetVisibility("denizens")
    else
      Script.SetVisibility("intruders")
    end

    if intruders then
      Script.DialogBox("ui/dialog/Lvl03/pass_to_denizens.json")

      if store.bFloodStarted and not store.bToldDenizensAboutFloodStart then
        --if we haven't told the denizens about flood start then the intruders must have
        --got to the waypoint.  Tell deni's about both the waypoint and the alarm.
        store.bToldDenizensAboutFloodStart = true
        Script.DialogBox("ui/dialog/Lvl03/Lvl_03_Intruders_Got_To_Waypoint_Denizens.json")
        Script.SetMusicParam("tension_level", 0.7)
      end 
    else
      Script.DialogBox("ui/dialog/Lvl03/pass_to_intruders.json")
      if not store.bDoneIntruderIntro then
        store.bDoneIntruderIntro = true
        Script.DialogBox("ui/dialog/Lvl03/Lvl_03_Opening_Intruders.json")
      end

      if bIntrudersSpotted and not store.bTalkedAboutIntruderSpot then
        --The master saw an intruder last turn. Need to let the truds know that the
        --master can now set off the alarm.
        store.bTalkedAboutIntruderSpot = true
        Script.DialogBox("ui/dialog/Lvl03/Lvl_03_Identified_Escapees_Intruders.json")
      end

      if store.bFloodStarted and not store.bToldIntrudersAboutAlarm then
        --If the intruders don't know about the alarm, then the master triggered it this turn.
        --Tell the intruders that the alarm has sounded.
        store.bToldIntrudersAboutAlarm =  true
        Script.DialogBox("ui/dialog/Lvl03/Lvl_03_Alarm_Started_Intruders.json")
      end
    end

    Script.SetLosMode("intruders", "entities")
    Script.SetLosMode("denizens", "entities")
    Script.LoadGameState(store.game)

    --focus the camera on somebody on each team.
    side2 = {Intruder = not intruders, Denizen = intruders, Npc = false, Object = false}  --reversed because it's still one side's turn when we're replaying their actions for the other side.
    Script.FocusPos(GetEntityWithMostAP(side2).Pos)

    for _, exec in pairs(store.execs) do
      bDone = false
      if exec.script_spawn then
        doSpawn(exec)
        bDone = true
      end
      if exec.script_despawn then
        deSpawn(exec)
        bDone = true
      end
      if exec.script_gear then
        doGear(exec)
        bDone = true
      end
      if exec.script_condition then
        doCondition(exec)
        bDone = true
      end     
      if exec.script_waypoint then
        doWaypoint(exec)
        bDone = true
      end      
      if not bDone then
        Script.DoExec(exec)

        --will be used at turn start to try to reselect the last thing they acted with.
        if exec.Ent.Side == "intruders" then
          LastIntruderEnt = exec.Ent
        end 
        if exec.Ent.Side == "denizens" then
          LastDenizenEnt = exec.Ent
        end 
      end
    end
    store.execs = {}
  end
end

function StoreSpawn(name, spawnPos)
  spawn_exec = {script_spawn=true, name=name, pos=spawnPos}
  store.execs[table.getn(store.execs) + 1] = spawn_exec
  return doSpawn(spawn_exec)
end

function doSpawn(spawnExec)
  return Script.SpawnEntityAtPosition(spawnExec.name, spawnExec.pos)
end

function StoreGear(name, ent)
  gear_exec = {script_gear=true, name=name, entity=ent, add=addGear}
  store.execs[table.getn(store.execs) + 1] = gear_exec
  doGear(gear_exec)
end

function doGear(gearExec)
  Script.SetGear(gearExec.entity, gearExec.name)
end

function StoreCondition(name, ent, addCondition)
  condition_exec = {script_condition=true, name=name, entity=ent, add=addCondition}
  store.execs[table.getn(store.execs) + 1] = condition_exec
end

function doCondition(conditionExec)
  Script.SetCondition(conditionExec.entity, conditionExec.name, conditionExec.add)
end

function StoreDespawn(ent)
  despawn_exec = {script_despawn=true, entity=ent}
  store.execs[table.getn(store.execs) + 1] = despawn_exec
  deSpawn(despawn_exec)
end

function deSpawn(despawnExec)
  DeadBodyDump = Script.GetSpawnPointsMatching("Dead_People")
  DeadBodyDump[1].Pos.X = DeadBodyDump[1].Pos.X + store.DeathCounter
  store.DeathCounter = store.DeathCounter + 1
  Script.SetPosition(despawnExec.entity, DeadBodyDump[1].Pos)
end

function GetDistanceBetweenEnts(ent1, ent2)
  v1 = ent1.Pos.X - ent2.Pos.X
  if v1 < 0 then
    v1 = 0-v1
  end
  v2 = ent1.Pos.Y - ent2.Pos.Y
  if v2 < 0 then
    v2 = 0-v2
  end
  return v1 + v2
end

function MasterIsAlive()
  for _, ent in pairs(Script.GetAllEnts()) do
    if ent.Name == MasterName and ent.HpCur > 0 then
      return true
    end
  end
  return false  
end

function AnyIntrudersAlive()
  for _, ent in pairs(Script.GetAllEnts()) do
    if ent.Side.Intruder and ent.HpCur > 0 then
      return true
    end
  end
  return false  
end

function EntIsNextToPatient(ExecEnt)
  for _, ent in pairs(Script.GetAllEnts()) do
    if ent.Name == "Patient" then
      if GetDistanceBetweenEnts(ExecEnt, ent) <= 3 then
        return ent
      end
    end
  end
  return false  
end

function TotalDeniCount()
  count = 0
  for _, ent in pairs(Script.GetAllEnts()) do
    if ent.Side.Denizen then
      count = count + 1
    end
  end
  return count
end

function ValueForReinforce()
  --The denizens get to reinforce after each waypoint goes down.
  --They get 10 - (value of units on the board) + (2 for each dead waypoint)

  nTotalValueOnBoard = 0
  for _, ent in pairs(Script.GetAllEnts()) do
    for _, entValue in pairs(ents) do
      if ent.Name == entValue[1] then
        nTotalValueOnBoard = nTotalValueOnBoard + entValue[2]
      end 
    end
  end
  return (10 - nTotalValueOnBoard)
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

function pointIsInSpawn(pos, sp)
  return pos.X >= sp.Pos.X and pos.X < sp.Pos.X + sp.Dims.Dx and pos.Y >= sp.Pos.Y and pos.Y < sp.Pos.Y + sp.Dims.Dy
end

function SpawnIntruderOrMonster(entToKillAndReplace)
  SpawnPos = entToKillAndReplace.Pos 
  StoreDespawn(entToKillAndReplace)
  thingToSpawn = ""

  if store.nIntrudersFound == 0 then
    store.nIntrudersFound = store.nIntrudersFound + 1
    if store.nIntrudersFound == 1 then
      thingToSpawn = "Collector"
      Script.DialogBox("ui/dialog/Lvl03/Lvl_03_Rescued_Intruder1.json")
    end
  else
    if (math.random(1, 5) > 2 and store.nIntrudersFound <= 3) then
      --Spawn intruder
      store.nIntrudersFound = store.nIntrudersFound + 1
      if store.nIntrudersFound == 2 then
        thingToSpawn = "Reporter"
        Script.DialogBox("ui/dialog/Lvl03/Lvl_03_Rescued_Intruder2.json")
      end
      if store.nIntrudersFound == 3 then
        thingToSpawn = "Detective"
        Script.DialogBox("ui/dialog/Lvl03/Lvl_03_Rescued_Intruder3.json")    
      end
    else
      --Spawn monster
      store.nMonstersFound = store.nMonstersFound + 1
      if store.nMonstersFound == 1 then
        Script.DialogBox("ui/dialog/Lvl03/Lvl_03_Spawned_Infected1.json") 
      end
      if store.nMonstersFound > 1 then
        Script.DialogBox("ui/dialog/Lvl03/Lvl_03_Spawned_Addl_Infected.json") 
      end    
      thingToSpawn = "Infected"
    end
  end

  StoreSpawn(thingToSpawn, SpawnPos)

end

function StoreWaypoint(wpname, wpside, wppos, wpradius, wpremove)
  waypoint_exec = {script_waypoint=true, name=wpname, side=wpside, pos=wppos, radius=wpradius, remove=wpremove}
  store.execs[table.getn(store.execs) + 1] = waypoint_exec
  doWaypoint(waypoint_exec)
end

function doWaypoint(waypointExec)
  if waypointExec.remove then
    return Script.RemoveWaypoint(waypointExec.name)
  else
    return Script.SetWaypoint(waypointExec.name, waypointExec.side, waypointExec.pos, waypointExec.radius)
  end
end