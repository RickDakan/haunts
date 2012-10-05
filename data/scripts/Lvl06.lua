
--****TUTORIAL SECTION****--
function IsStoryMode()
  return true
end

function DoTutorials()
  --We should totally do some tutorials here.
  --It would be super cool.
end
--****END TUTORIAL****--

--****EVENTS****--
function Init(data)
  side_choices = Script.ChooserFromFile("ui/start/versus/side.json")

  Script.LoadHouse("Lvl_06_creature_feature")
  Script.PlayMusic("Haunts/Music/Adaptive/Bed 2")
  Script.SetMusicParam("tension_level", 0.1)     

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

  --Spawn intro versions of each of the 4 master.
  ent = Script.SpawnEntitySomewhereInSpawnPoints("Subject Orlac", Script.GetSpawnPointsMatching("Vampire_Start"), false)
  Script.SetWaypoint("Vampire", "intruders", ent.Pos, 1)
  Script.SetWaypoint("Vampire2", "denizens", ent.Pos, 1)
  ent = Script.SpawnEntitySomewhereInSpawnPoints("Subject Chosen", Script.GetSpawnPointsMatching("Cult_Leader_Start"), false)
  Script.SetWaypoint("Cult Leader", "intruders", ent.Pos, 1)
  Script.SetWaypoint("Cult Leader2", "denizens", ent.Pos, 1)
  ent = Script.SpawnEntitySomewhereInSpawnPoints("Subject Bosch", Script.GetSpawnPointsMatching("Bosch_Start"), false)
  Script.SetWaypoint("Bosch", "intruders", ent.Pos, 1)
  Script.SetWaypoint("Bosch2", "denizens", ent.Pos, 1)
  ent = Script.SpawnEntitySomewhereInSpawnPoints("Subject Ancient", Script.GetSpawnPointsMatching("Ancient_Start"), false)
  Script.SetWaypoint("Ancient", "intruders", ent.Pos, 1)
  Script.SetWaypoint("Ancient2", "denizens", ent.Pos, 1)
  --Spawn the golem.
  Script.SpawnEntitySomewhereInSpawnPoints("The Golem", Script.GetSpawnPointsMatching("Golem_Start"), false)

  store.bChoiceMode = true
  store.IntruderNames = {"Professor Keith Evans"}
  store.IntruderTypes = {"Intruder"}
  store.IntruderMinions = {}
  store.DenizenNames = {"Sir Wilhem Bohn"}
  store.DenizenTypes = {"Master"}
  store.DenizenMinions = {}

  store.ScoreCounter = 10
end

function intrudersSetup()
  if IsStoryMode() then
    intruder_names = {"Prof. Keith Evans"}
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
    denizen_names = {"Sir Wilhem"}
    denizen_spawn = Script.GetSpawnPointsMatching("Master_Start")
  end 

  for _, name in pairs(denizen_names) do
    ent = Script.SpawnEntitySomewhereInSpawnPoints(name, denizen_spawn, false)
  end

  -- Choose entry point here.
  Script.SaveStore()
end

function RoundStart(intruders, round)
  if not store.execs then
    store.execs = {}
  end

  if round == 1 then
    if intruders then
      intrudersSetup() 
    else
      Script.DialogBox("ui/dialog/Lvl06/Lvl_06_Opening_Denizens.json")
      Script.FocusPos(Script.GetSpawnPointsMatching("Master_Start")[1].Pos)
      denizensSetup()
    end
    Script.SetLosMode("intruders", "blind")
    Script.SetLosMode("denizens", "blind")

    if IsStoryMode() then
      DoTutorials()
    end

    store.game = Script.SaveGameState()
    Script.EndPlayerInteraction()

    return
  end

  if store.bChoiceMode then
    ZeroMasterAp()
  else
    RespawnDeadPeople(intruders)
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

function OnMove(ent, path)
  return table.getn(path)
end

function OnAction(intruders, round, exec)
  -- Check for players being dead here
  if store.execs == nil then
    store.execs = {}
  end
  store.execs[table.getn(store.execs) + 1] = exec

  if store.bChoiceMode then
    PickEnt(exec.Ent, false)

    --If this was the last choice
    --Assign the remaining monster to the other team.
    --Start the game
    if table.getn(store.DenizenNames) == 3 or table.getn(store.IntruderNames) == 3 then 
      EndChoicePhase()
    end
  else
    --if a minion gets to the center, there is an explosion and the golem moves
    if EntIsMinion(exec.Ent) then
      if GetDistanceBetweenPoints(exec.Ent.Pos, Script.GetSpawnPointsMatching("Objective")[1].Pos) <= 2 then
        StoreDespawn(exec.Ent)
        --Explode()
        PushGolem(exec.Ent.Side.Intruder)
        --If the golem reaches the doors on either side, the opposing side wins.
        Script.SetAp(exec.Ent, 0)
        if store.ScoreCounter >= 20 then
          Script.Sleep(2)
          Script.DialogBox("ui/dialog/Lvl06/Lvl_06_Victory_Intruders.json")
        end
        if store.ScoreCounter <= 0 then
          Script.Sleep(2)
          Script.DialogBox("ui/dialog/Lvl06/Lvl_06_Victory_Denizens.json")
        end
      end
    end
  end



  --after any action, if this ent's Ap is 0, we can select the next ent for them
  if exec.Ent.ApCur == 0 then 
    nextEnt = GetEntityWithMostAP(exec.Ent.Side)
    if nextEnt.ApCur > 0 then
      if exec.Action.Type ~= "Move" then
        Script.Sleep(2)
      end      
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
      Script.DialogBox("ui/dialog/Lvl06/pass_to_denizens.json")
      if store.bChoiceMode then

      else             

      end

    else
      if not bIntruderIntroDone then
        bIntruderIntroDone = true
        Script.DialogBox("ui/dialog/Lvl06/pass_to_intruders.json")
        Script.DialogBox("ui/dialog/Lvl06/Lvl_06_Opening_Intruders.json")
        bSkipOtherChecks = true
      end

      if not bSkipOtherChecks then  --if we haven't showed any of the other start messages, use the generic pass.
        Script.DialogBox("ui/dialog/Lvl06/pass_to_intruders.json")
      end


      if store.bChoiceMode then

      else
        if not store.bDeniSetupDone then
          store.bDeniSetupDone = true
          RespawnDeadPeople(false)
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
      if exec.script_setPos then
        doSetPos(exec)
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
          LastDenizenEnt = GetMasterEnt() --always the master on this board
        end 
      end
    end
    store.execs = {}
  end
end

function EndChoicePhase()
  store.bChoiceMode = false
  if table.getn(store.DenizenNames) < 3 then  --add the last dude to the team with fewer dudes.
    PickEnt(GetEntWithName("Sir Wilhem"), true)
  else
    PickEnt(GetEntWithName("Prof. Keith Evans"), true)
  end
  --Despawn the two dummy dudes.
  StoreDespawn(GetEntWithName("Sir Wilhem"))
  StoreDespawn(GetEntWithName("Prof. Keith Evans"))

  StoreWaypoint("Objective", "intruders", Script.GetSpawnPointsMatching("Objective")[1].Pos, 2, false)
  StoreWaypoint("ObjectiveDen", "denizens", Script.GetSpawnPointsMatching("Objective")[1].Pos, 2, false)

  -- --Spawn the real intruder and master and their teams.
  RespawnDeadPeople(true)
  -- RespawnDeadPeople(false)

  Script.EndPlayerInteraction()

  --Side that didn't just pick will have the first turn.
end

function Explode()
  --!!!!play golem explode animation here 


  for _, ent in pairs(Script.GetAllEnts()) do
    if not EntIsMinion(ent) and not ent.Name == "The Golem" then  --Can remove this if we want the game to be longer...
      if GetDistanceBetweenPoints(ent.Pos, Script.GetSpawnPointsMatching("Objective")[1].Pos) <= 10 then
        if ent.HpCur > 3 then
          --!!!! play got hit animation here
        else
          --!!!! play dead animation here
        end
        Script.SetHp(ent, ent.HpCur - 3)  --  <-- DAMAGE DONE BY EXPLOSION

      end 
    end   
  end
end


function PushGolem(bIntruders)
  golemEnt = GetEntWithName("The Golem")
  newPos = golemEnt.Pos
  newPos.Y = golemEnt.Pos.Y
  if bIntruders then
    newPos.X = golemEnt.Pos.X + 1
    store.ScoreCounter = store.ScoreCounter + 1
  else
    newPos.X = golemEnt.Pos.X - 1
    store.ScoreCounter = store.ScoreCounter - 1
  end
  ClearSpot(newPos)
  StoreSetPos(golemEnt, newPos)
  side3 = {Intruder = bntruders, Denizen = not intruders, Npc = false, Object = false}
  nextEnt = GetEntityWithMostAP(side3)
  Script.SelectEnt(nextEnt)
end

function ZeroMasterAp()
  for _, ent in pairs(Script.GetAllEnts()) do
    if ent.Name ~= "Prof. Keith Evans" and ent.Name ~= "Sir Wilhem" then
      Script.SetAp(ent, 0)
    end  
  end
end

--****END EVENTS****--

--****STORE/DO SECTION****--
function StoreSpawn(entName, spawnPos)
  spawn_exec = {script_spawn=true, name=entName, pos=spawnPos}
  store.execs[table.getn(store.execs) + 1] = spawn_exec
  return doSpawn(spawn_exec)
end

function doSpawn(spawnExec)
   return Script.SpawnEntityAtPosition(spawnExec.name, spawnExec.pos)
end

function StoreSetPos(ent, setPos)
  setpos_exec = {script_setPos=true, entity=ent, pos=setPos}
  store.execs[table.getn(store.execs) + 1] = setpos_exec
  return doSetPos(setpos_exec)
end

function doSetPos(SetPosExec)
   return Script.SetPosition(SetPosExec.entity, SetPosExec.pos)
end

function StoreDespawn(ent)
  despawn_exec = {script_despawn=true, entity=ent}
  store.execs[table.getn(store.execs) + 1] = despawn_exec
  deSpawn(despawn_exec)
end

function deSpawn(despawnExec)
  Script.SetHp(despawnExec.entity, 0)
  Script.PlayAnimations(despawnExec.entity, {"defend", "killed"})
  DeadBodyDump = Script.GetSpawnPointsMatching("Dead_People")
  Script.SetPosition(despawnExec.entity, DeadBodyDump[1].Pos)
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
--****END STORE/DO****--


--****UTILS SECTION ****--
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

function GetEntWithName(name)
  for _, ent in pairs(Script.GetAllEnts()) do
    if ent.Name == name then
      return ent
    end
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

function PickEnt(ent, bForce)
  checkEnt = GetEntWithName("Subject Bosch")
  if not store.sel1 then
    if GetDistanceBetweenEnts(ent, checkEnt) <= 2 or bForce then
      --They selected this ent.
      if ent.Side.Intruder then
        store.IntruderNames[table.getn(store.IntruderNames) + 1] = "Codename: Bosch"
        store.IntruderTypes[table.getn(store.IntruderTypes) + 1] = "Bosch"
        store.IntruderMinions[table.getn(store.IntruderMinions) + 1] = "Wraith for Evans"
      else
        store.DenizenNames[table.getn(store.DenizenNames) + 1] = "Subject Bosch"
        store.DenizenTypes[table.getn(store.DenizenTypes) + 1] = "Bosch"
        store.DenizenMinions[table.getn(store.DenizenMinions) + 1] = "Wraith for Bohn"        
      end
      StoreWaypoint("Bosch", "", "", "", true)
      StoreWaypoint("Bosch2", "", "", "", true)
      StoreDespawn(checkEnt)
      store.sel1 = true
      Script.SetAp(ent, 0)
    end
  end

  checkEnt = GetEntWithName("Subject Orlac")
  if not store.sel2 then
    if GetDistanceBetweenEnts(ent, checkEnt) <= 2 or bForce then
      --They selected this ent.
      if ent.Side.Intruder then
        store.IntruderNames[table.getn(store.IntruderNames) + 1] = "Codename: Orlac"
        store.IntruderTypes[table.getn(store.IntruderTypes) + 1] = "Vampire"
        store.IntruderMinions[table.getn(store.IntruderMinions) + 1] = "Shade for Evans"
      else
        -- NEW THING
        store.DenizenNames[table.getn(store.DenizenNames) + 1] = "Subject Orlac"
        store.DenizenTypes[table.getn(store.DenizenTypes) + 1] = "Vampire"
        store.DenizenMinions[table.getn(store.DenizenMinions) + 1] = "Shade for Bohn"
      end
      StoreWaypoint("Vampire", "", "", "", true)
      StoreWaypoint("Vampire2", "", "", "", true)
      StoreDespawn(checkEnt)
      store.sel2 = true
      Script.SetAp(ent, 0)
    end
  end

  checkEnt = GetEntWithName("Subject Ancient")
  if not store.sel3 then
    if GetDistanceBetweenEnts(ent, checkEnt) <= 2 or bForce then
      --They selected this ent.
      if ent.Side.Intruder then
        store.IntruderNames[table.getn(store.IntruderNames) + 1] = "Codename: Ancient"
        store.IntruderTypes[table.getn(store.IntruderTypes) + 1] = "Ancient"
        store.IntruderMinions[table.getn(store.IntruderMinions) + 1] = "Corpse for Evans"
      else
        store.DenizenNames[table.getn(store.DenizenNames) + 1] = "Subject Ancient"
        store.DenizenTypes[table.getn(store.DenizenTypes) + 1] = "Ancient"
        store.DenizenMinions[table.getn(store.DenizenMinions) + 1] = "Corpse for Bohn"
      end
      StoreWaypoint("Ancient", "", "", "", true)
      StoreWaypoint("Ancient2", "", "", "", true)
      StoreDespawn(checkEnt)
      store.sel3 = true
      Script.SetAp(ent, 0)
    end
  end

  checkEnt = GetEntWithName("Subject Chosen")
  if not store.sel4 then
    if GetDistanceBetweenEnts(ent, checkEnt) <= 2 or bForce then
      --They selected this ent.
      if ent.Side.Intruder then
        store.IntruderNames[table.getn(store.IntruderNames) + 1] = "Codename: Chosen"
        store.IntruderTypes[table.getn(store.IntruderTypes) + 1] = "Cult_Leader"
        store.IntruderMinions[table.getn(store.IntruderMinions) + 1] = "Cultist for Evans"
      else
        store.DenizenNames[table.getn(store.DenizenNames) + 1] = "Subject Chosen"
        store.DenizenTypes[table.getn(store.DenizenTypes) + 1] = "Cult_Leader"
        store.DenizenMinions[table.getn(store.DenizenMinions) + 1] = "Cultist for Bohn"
      end
      StoreWaypoint("Cult Leader", "", "", "", true)
      StoreWaypoint("Cult Leader2", "", "", "", true)
      StoreDespawn(checkEnt)
      store.sel4 = true
      Script.SetAp(ent, 0)
    end
  end  
end

function RespawnDeadPeople(intruders)
  if intruders then
    for i = 1, 3, 1 do
      ent = GetEntWithName(store.IntruderNames[i])
      if not ent then            
        ent = StoreSpawn(store.IntruderNames[i], Script.GetSpawnPointsMatching(store.IntruderTypes[i] .. "_Spawn")[1].Pos)
        Script.SetAp(ent, 0) --death means you lose a turn.
      else
        if ent.HpCur <= 0 then
          ent = StoreSpawn(store.IntruderNames[i], Script.GetSpawnPointsMatching(store.IntruderTypes[i] .. "_Spawn")[1].Pos)
          Script.SetAp(ent, 0) --death means you lose a turn.
        end
      end

    end
  else
    for i = 1, 3, 1 do
      ent = GetEntWithName(store.DenizenNames[i])
      if not ent then  
        ent = StoreSpawn(store.DenizenNames[i], Script.GetSpawnPointsMatching(store.DenizenTypes[i] .. "_Spawn")[1].Pos)
        Script.SetAp(ent, 0)
      else
        if ent.HpCur <= 0 then
          ent = StoreSpawn(store.DenizenNames[i], Script.GetSpawnPointsMatching(store.DenizenTypes[i] .. "_Spawn")[1].Pos)
          Script.SetAp(ent, 0) --death means you lose a turn.
        end
      end
    end
  end
end

function EntIsMinion(ent)
  for i = 1, 3, 1 do
    if ent.Name == store.IntruderNames[i] or ent.Name == store.DenizenNames[i] then
      return false
    end  
  end
  return true
end

function ClearSpot(pos)
  for _, ent in pairs(Script.GetAllEnts()) do
    if ent.Pos.X == pos.X and ent.Pos.Y == pos.Y then
      StoreDespawn(ent)
    end
  end
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

--****END UTILS****--