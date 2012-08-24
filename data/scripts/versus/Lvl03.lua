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

  --spawn the objective point
  waypoint_spawn = Script.GetSpawnPointsMatching("Relic_Spawn")
  Waypoint = Script.SpawnEntitySomewhereInSpawnPoints("Antidote", waypoint_spawn)

  --Spawn the patients
  patient_spawn = Script.GetSpawnPointsMatching("Patient_Start")
  for i = 1, 10, 1 do  
    --this will cause some "cannot find place to spawn" errors in the log.  
    --But covers us in case we decide to mess with the number of patients.  I'm ok with it.
    Script.SpawnEntitySomewhereInSpawnPoints("Patient", patient_spawn)
  end 

  --we will incorporate some randomness here.
  math.randomseed(os.time())

  --set these modular variables.
  nIntrudersFound = 0
  nMonstersFound = 0
  bWaypointDown = false
  bFloodStarted = false
end

function intrudersSetup()

  if IsStoryMode() then
    intruder_names = {"Teen"}
    intruder_spawn = Script.GetSpawnPointsMatching("Intruder_Start1")
  -- else
  --   --permit all choices for normal vs play
  end 

  for _, name in pairs(intruder_names) do
    ent = Script.SpawnEntitySomewhereInSpawnPoints(name, intruder_spawn)
    
  --Don't understand gear yet...halp!?
    Script.SetGear(ent, "Pre-loaded Playlist")
    -- PRETEND!
  end

  -- Choose entry point here.
  Script.SaveStore()
end

function denizensSetup()
  -- This creates a list of entities and associated point values.  The order
  -- the names are listed in here is the order they will appear to the user.
  if IsStoryMode() then
    ents = {
      {"Administrator", 1},
    }
  else
    --permit all choices for normal vs play.

  end
  
  Script.SetVisibility("denizens")
  setLosModeToRoomsWithSpawnsMatching("denizens", "Master_.*")

  -- Now we give the user a ui with which to place these entities.  The user
  -- will have 1 point to spend, and each of the options costs one point, so
  -- they will only place 1.  We will make sure they place exactly one.
  -- Also the "Master-.*" indicates that the entity can only be placed in
  -- spawn points that have a name that matches the regular expression
  -- "Master-.*", which means anything that begins with "Master-".  So
  -- "Master-BackRoom" and "Master-Center" both match, for example.
  placed = {}
  while table.getn(placed) == 0 do
    placed = Script.PlaceEntities("Master_.*", ents, 1, 1)
  end

  -- placed is an array containing all of the entities placed, in this case
  -- there will only be one, and we will use that one to determine what
  -- servitors to make available to the user to place.
  if placed[1].Name == "Chosen One" then
    MasterName = "Chosen One"
    ServitorEnts = {
      {"Disciple", 1},
      {"Devotee", 1},
      {"Eidolon", 3},
    }
  end
  if placed[1].Name == "Administrator" then
    MasterName = "Administrator"
    ServitorEnts = 
    {
      {"Attendant", 2},
      {"Technician", 3},
    }  
  end

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

  if bFloodStarted then
    --At the start of each denizen turn, we're going to randomly spawn attendants at 
    --the flood points.  Also need to prevent spawning within LoS of an intruder.

    if TotalDeniCount() < 12 then  --don't want to overdo it
      available_spawns = GetSpawnsFromListWhereNoLoS(Script.GetSpawnPointsMatching("Flood_Point"))
      for i = 1, 3, 1 do
        --Pick an entity
        if math.random(4) > 2 then
          floodEnt = ServitorEnts[1][1]
        else
          floodEnt = ServitorEnts[2][1]
        end     
        Script.SpawnEntitySomewhereInSpawnPoints(floodEnt, available_spawns)
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
--    if LastIntruderEnt.Side == side then
      Script.SelectEnt(LastIntruderEnt)
      bDone = true
    end
  end  
  if LastDenizenEnt and not bDone then
    if side.Denizen then
--    if LastDenizenEnt.Side == side then      
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

  if exec.Ent.Side.Intruder and GetDistanceBetweenEnts(exec.Ent, Waypoint) <= 3 and not bWaypointDown then
    --The intruders got to the first waypoint.
    bWaypointDown = true
    --!!!! need a way to despawn objects
    -- StoreDespawn(Waypoint)
    -- deSpawn(despawn_exec)

    Script.SetCondition(exec.Ent, "Carrying Antidote", true)
--    Script.SetGear(exec.Ent, "Antidote")
    --exec.Ent.Actions[table.getn(exec.Ent.Actions) + 1] = "Hand Antidote"

    if not bFloodStarted then    
      --The denizens have not yet activated the alarm.
      bFloodStarted = true
      Script.DialogBox("ui/dialog/Lvl03/Lvl_03_Waypoint_And_Alarm_Intruders.json")
      bToldIntrudersAboutAlarm = true
    else
      --Denizens already started the alarm.  Just tell the intruders about the waypoint.
      Script.DialogBox("ui/dialog/Lvl03/Lvl_03_Waypoint_Only_Intruders.json")
    end
  end 

  if exec.Action.Name == "Hand Antidote" then
--    Script.SetGear(exec.Target, "Antidote")
    --remove the carrying condition from the exec ent.  The target will get the
    --condition b/c of the ability.
    Script.SetCondition(exec.Ent, "Carrying Antidote", false)
  end

  --The big check: Have the intruders won.
  if bWaypointDown then
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
    Script.SetVisibility("denizens")
    setLosModeToRoomsWithSpawnsMatching("denizens", "Servitors_Start")
    placed = Script.PlaceEntities("Servitors_Start", ServitorEnts, 0, ValueForReinforce())
    Script.SetLosMode("intruders", "blind")
    Script.SetLosMode("denizens", "blind")   

    Script.EndPlayerInteraction()
  end  

  --if the deni master spotted the intruders, show dialogue
  if exec.Action.Name == "Identify Escapee" then
    --Can only be used on an intruder, so we don't need to check the target.
    --Once used, the master may escape the board in order to activate the alarm.
    bInstrudersSpotted = true
    Script.DialogBox("ui/dialog/Lvl03/Lvl_03_Identified_Escapees_Denizens.json")
  end 

  --if the deni mast reached the exit after seeing the intruders, sound the alarm
  if bInstrudersSpotted then
    if exec.Ent.Name == MasterName then
      if pointIsInSpawns(exec.Ent.Pos, "Escape") then
        bFloodStarted = true
        Script.DialogBox("ui/dialog/Lvl03/Lvl_03_Alarm_Started_Denizens.json")
        bToldDenizensAboutFloodStart = true
      end
    end
  end

  --after any action, if this ent's Ap is 0, we can select the next ent for them
  if exec.Ent.ApCur == 0 then
    nextEnt = GetEntityWithMostAP(exec.Ent.Side)
    if nextEnt.ApCur > 0 then
      Script.SelectEnt(nextEnt)
    end
  end
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

      if bFloodStarted and not bToldDenizensAboutFloodStart then
        --if we haven't told the denizens about flood start then the intruders must have
        --got to the waypoint.  Tell deni's about both the waypoint and the alarm.
        bToldDenizensAboutFloodStart = true
        Script.DialogBox("ui/dialog/Lvl03/Lvl_03_Intruders_Got_To_Waypoint_Denizens.json")
      end
    else
      Script.DialogBox("ui/dialog/Lvl03/pass_to_intruders.json")
      if bIntrudersSpotted and not bTalkedAboutIntruderSpot then
        --The master saw an intruder last turn. Need to let the truds know that the
        --master can now set off the alarm.
        bTalkedAboutIntruderSpot = true
        Script.DialogBox("ui/dialog/Lvl03/Lvl_03_Identified_Escapees_Intruders.json")
      end

      if bFloodStarted and not bToldIntrudersAboutAlarm then
        --If the intruders don't know about the alarm, then the master triggered it this turn.
        --Tell the intruders that the alarm has sounded.
        bToldIntrudersAboutAlarm =  true
        Script.DialogBox("ui/dialog/Lvl03/Lvl_03_Alarm_Started_Intruders.json")
      end
    end

    Script.SetLosMode("intruders", "entities")
    Script.SetLosMode("denizens", "entities")
    Script.LoadGameState(store.game)
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
end

function doSpawn(spawnExec)
  Script.SpawnEntityAtPosition(spawnExec.name, spawnExec.pos)
end

function StoreDespawn(ent)
  despawn_exec = {script_despawn=true, entity=ent}
  store.execs[table.getn(store.execs) + 1] = despawn_exec
end

function deSpawn(despawnExec)
  if despawnExec.entity.HpMax then  --can only kill things that have hp
    Script.PlayAnimations(despawnExec.entity, {"defend", "killed"})
    Script.SetHp(despawnExec.entity, 0)
  end
  DeadBodyDump = Script.GetSpawnPointsMatching("Dead_People")
  Script.SetPosition(despawnExec.entity, DeadBodyDump[1].Pos)
end

function GetDistanceBetweenEnts(ent1, ent2)
  return (math.abs(ent1.Pos.X - ent2.Pos.X) + math.abs(ent1.Pos.Y - ent2.Pos.Y))
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

function GetSpawnsFromListWhereNoLoS(spawns)
  -- for _, possibleSpawn in pairs(spawns) do
  --   --nasty set of loops here.
  --   for _, ent in pairs(Script.GetAllEnts()) do
  --     if ent.Side.Intruder then
  --       for _, possiblePositions in pairs(Script.)

  --     end
  --   end
  -- end

  --!!!!ignore this for now.
  return spawns
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
  deSpawn(store.execs[table.getn(store.execs)]) 

  thingToSpawn = ""

  if nIntrudersFound == 0 then
    nIntrudersFound = nIntrudersFound + 1
    if nIntrudersFound == 1 then
      thingToSpawn = "Ghost Hunter"
      Script.DialogBox("ui/dialog/Lvl03/Lvl_03_Rescued_Intruder1.json")
    end
  else
    if (math.random(1, 5) > 2 and nIntrudersFound <= 3) then
      --Spawn intruder
      nIntrudersFound = nIntrudersFound + 1
      if nIntrudersFound == 2 then
        thingToSpawn = "Researcher"
        Script.DialogBox("ui/dialog/Lvl03/Lvl_03_Rescued_Intruder2.json")
      end
      if nIntrudersFound == 3 then
        thingToSpawn = "Cordelia"
        Script.DialogBox("ui/dialog/Lvl03/Lvl_03_Rescued_Intruder3.json")    
      end
    else
      --Spawn monster
      nMonstersFound = nMonstersFound + 1
      if nMonstersFound == 1 then
        Script.DialogBox("ui/dialog/Lvl03/Lvl_03_Spawned_Infected1.json") 
      end
      if nMonstersFound > 1 then
        Script.DialogBox("ui/dialog/Lvl03/Lvl_03_Spawned_Addl_Infected.json") 
      end    
      thingToSpawn = "Infected"
    end
  end

  StoreSpawn(thingToSpawn, SpawnPos)
  doSpawn(store.execs[table.getn(store.execs)])

end
