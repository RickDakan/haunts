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

  Script.LoadHouse("Lvl_05_museum")  
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

  store.ActivatedObjectives = {}
  --ok lissen.  We're gonna spawn a lotta buncha op points.
  store.Objectives = Script.GetSpawnPointsMatching("Artifact")
  for i = 1, 10, 1 do
    Script.SetWaypoint("Objective" .. i, "denizens", store.Objectives[i].Pos, 1)
    print("Objective" .. i)
  end   

  store.ObjectivesAcquired = 0
  store.bSummoning = false
end

function intrudersSetup()

  if IsStoryMode() then
    intruder_names = {"Sonico Mono", "Christopher Matthias", "Cora Phinneas"}
    intruder_spawn = Script.GetSpawnPointsMatching("Intruder_Start")
  end 

  for _, name in pairs(intruder_names) do
    ent = Script.SpawnEntitySomewhereInSpawnPoints(name, intruder_spawn, false)
  end

  -- Choose entry point here.
  Script.SaveStore()
end

function denizensSetup()
  if IsStoryMode() then
    denizen_names = {"Alexander Tostowaryk", "James MacLeod", "Danielle Marricotte", "Nelle Anders"}
    denizen_spawn = Script.GetSpawnPointsMatching("Denizen_Start")
  end 

  choice = Script.DialogBox("ui/dialog/Lvl05/Lvl_05_Master_Choice_Denizens.json")

  i = 1
  for _, name in pairs(denizen_names) do
    ent = Script.SpawnEntitySomewhereInSpawnPoints(name, denizen_spawn, false)
    if tonumber(choice[1]) == i then
      store.MasterName = ent.Name
      store.MasterEnt = ent
    end
    i = i + 1
  end

  -- Choose entry point here.
  Script.SaveStore()
end

function RoundStart(intruders, round)
  if round == 1 then
    if intruders then
      intrudersSetup() 
    else
      Script.DialogBox("ui/dialog/Lvl05/Lvl_05_Opening_Denizens.json")
      Script.FocusPos(Script.GetSpawnPointsMatching("Denizen_Start")[1].Pos)
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

  if intruders and store.bDenizenMasterFoundObLastTurn then
    store.bDenizenMasterFoundObLastTurn = false
    Script.DialogBox("ui/dialog/Lvl05/Lvl_05_Artifact_Found_Intruders.json")
  end

  if not intruders and not store.bSummoning then
    Script.SetAp(store.MasterEnt, 300)
  end

  if store.bSummoning and not intruders then
    --the deni master is locked in place (MasterEnt will have been replaced with the real master)
    Script.SetAp(store.MasterEnt, 0)
    Script.SetWaypoint("Master", "intruders", store.MasterEnt.Pos, 1)
  end

  store.game = Script.SaveGameState()

  side = {Intruder = intruders, Denizen = not intruders, Npc = false, Object = false}
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

function GetDistanceBetweenPoints(pos1, pos2)
  v1 = pos1.X - pos2.X
  if v1 < 0 then
    v1 = 0-v1
  end
  v2 = pos1.Y - pos2.Y
  if v2 < 0 then
    v2 = 0-v2
  end
  return v1 + v2
end

function StartSummon()
  --Try to spawn a zombie around the master for each artifact he found
  store.bSummoning = true

  --replace the master ent with the real master.
  posToUse = store.MasterEnt.Pos

  --Get rid of the other 
  for _, ent in pairs(Script.GetAllEnts()) do
    if ent.Side.Denizen then
      StoreDespawn(ent)
    end 
  end 

  StoreSpawn("Bosch", posToUse)

  for i = 1, store.ObjectivesAcquired, 1  do
    --Summon a new minion
    omgCounter = 1
    for _, PossibleSpawn in pairs(Script.GetLos(store.MasterEnt)) do
      if Script.SpawnEntityAtPosition("Shade", PossibleSpawn.Pos) then
        i = i + 1
      end
    end
    omgCounter = omgCounter + 1
    if omgCounter >= 100 then
      break
    end
  end

  store.nTurnsRemaining = 11 - store.ObjectivesAcquired 

  --deactivate all the remaining waypoints
  for i = 1, 10, 1 do
    StoreWaypoint("Objective" .. i, "", "", "", true)
  end
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

  if  exec.Ent.Name == store.MasterEnt.Name then
    IndexTriggered = IsNextToActiveWaypoint(exec.Ent)
    if IndexTriggered then
      print("happy1")
      store.ActivatedObjectives[IndexTriggered] = true
      print("happy2:", IndexTriggered)
      StoreWaypoint("Objective" .. IndexTriggered, "", "", "", true)
      print("happy3")
      store.ObjectivesAcquired = store.ObjectivesAcquired + 1
      print("happy4")
      store.bDenizenMasterFoundObLastTurn = true
      Script.DialogBox("ui/dialog/Lvl05/Lvl_05_Artifact_Found_Denizens.json")
    end
  end 

  if exec.Action.Name == "Start Summon" then
    StartSummon()
  end

  if exec.Ent.Side.Intruder and exec.Ent.Name == store.MasterEnt.Name then
    if exec.Target.Side.Denizen and not exec.Target.Name == store.MasterEnt.Name then
      --they intruders attacked a decoy.  The attacking intruder is removed from the game
      Script.DialogBox("ui/dialog/Lvl05/Lvl_05_Attacked_Wrong_Entity_Intruders.json")
      StoreDespawn(exec.Ent)
    end

    --if the intruders attack the denizens, the summon starts immediately
    if exec.Target.Side.Denizen and exec.Target.Name == store.MasterEnt.Name and not store.bSummoning then
      Script.DialogBox("ui/dialog/Lvl05/Lvl_05_Master_Attacked_Intruders.json")
      store.bMasterAttacked = true
      StartSummon()
    end    

    if exec.Target.Name == store.MasterEnt.Name and store.bSummoning then
      if exec.Target.HpCur <= 0 then
        Script.DialogBox("ui/dialog/Lvl05/Lvl_05_Victory_Intruders.json")
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
 
function IsNextToActiveWaypoint(ent)
  i = 1
  for _, obj in pairs(store.Objectives) do
    if i <= 10 then
      --is this objective already activated
      print("Happy:")
      if not store.ActivatedObjectives[i] then
        print("Ent:", ent.Pos.X)
        print("Obj:", obj.Pos.X)
        if GetDistanceBetweenPoints(ent.Pos, obj.Pos) <= 2 then
          --They finded an artifact!
          return i  --Need the index to know which one to activate
        end
      end
    end
    i = i + 1
  end

  return false
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
      Script.DialogBox("ui/dialog/Lvl05/pass_to_denizens.json")
      if store.nTurnsRemaining == 0 and store.bSummoning then
        Script.DialogBox("ui/dialog/Lvl05/Lvl_05_Victory_Denizens.json")
      end
      if store.bMasterAttacked then
        Script.DialogBox("ui/dialog/Lvl05/Lvl_05_Master_Attacked_Denizens.json")
      end
      if store.bSummoning then
        Script.DialogBox("ui/dialog/Lvl05/Lvl_05_Turns_Remaining_Denizens.json", {turns=store.nTurnsRemaining})
      end
    else
      ZeroDenizenAp()  --Always set all deni AP = 0 to protect the master.
      if not bIntruderIntroDone then
        bIntruderIntroDone = true
        Script.DialogBox("ui/dialog/Lvl05/pass_to_intruders.json")
        Script.DialogBox("ui/dialog/Lvl05/Lvl_05_Opening_Intruders.json")
        bSkipOtherChecks = true
      end

      if not bSkipOtherChecks then  --if we haven't showed any of the other start messages, use the generic pass.
        Script.DialogBox("ui/dialog/Lvl04/pass_to_intruders.json")
        
        if store.bSummoning then
          --reduce the turns remaining and tell the intruders about it.
          store.nTurnsRemaining = store.nTurnsRemaining - 1
          if store.nTurnsRemaining == 0 then
            Script.DialogBox("ui/dialog/Lvl05/Lvl_05_Last_Turn_Intruders.json")
          else
            Script.DialogBox("ui/dialog/Lvl05/Lvl_05_Turns_Remaining_Intruders.json", {turns=store.nTurnsRemaining})
          end
        end
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
          LastDenizenEnt = store.MasterEnt --always the master on this board
        end 
      end
    end
    store.execs = {}
  end
end

function ZeroDenizenAp()
  for _, ent in pairs(Script.GetAllEnts()) do
    if ent.Side.Denizen then
      Script.SetAp(ent, 0)
    end
  end 
end

function AnyIntrudersAlive()
  for _, ent in pairs(Script.GetAllEnts()) do
    if ent.Side.Intruder and ent.HpCur > 0 then
      return true
    end
  end
  return false  
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

function StoreSpawn(entName, spawnPos)
  spawn_exec = {script_spawn=true, name=entName, pos=spawnPos}
  store.execs[table.getn(store.execs) + 1] = spawn_exec
  return doSpawn(spawn_exec)
end

function doSpawn(spawnExec)
  return Script.SpawnEntityAtPosition(spawnExec.name, spawnExec.pos)
end

function StoreDespawn(ent)
  despawn_exec = {script_despawn=true, entity=ent}
  store.execs[table.getn(store.execs) + 1] = despawn_exec
  deSpawn(despawn_exec)
end

function deSpawn(despawnExec)
  if despawnExec.entity.HpMax then  --can only kill things that have hp
    Script.PlayAnimations(despawnExec.entity, {"defend", "killed"})
    Script.SetHp(despawnExec.entity, 0)
  end
  DeadBodyDump = Script.GetSpawnPointsMatching("Dead_People")
  Script.SetPosition(despawnExec.entity, DeadBodyDump[1].Pos)
end

function StoreWaypoint(wpname, wpside, wppos, wpradius, wpremove)
  print("Name:", wpname)
  waypoint_exec = {script_waypoint=true, name=wpname, side=wpside, pos=wppos, radius=wpradius, remove=wpremove}
  print("remove:", wpremove)
  store.execs[table.getn(store.execs) + 1] = waypoint_exec
  print("DO IT") 
  doWaypoint(waypoint_exec)
  print("hippo")
end

function doWaypoint(waypointExec)
  if waypointExec.remove then
    print("Why won't you work?", waypointExec.name)
    return Script.RemoveWaypoint(waypointExec.name)
  else
    return Script.SetWaypoint(waypointExec.name, waypointExec.side, waypointExec.pos, waypointExec.radius)
  end
end