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
  Script.LoadHouse("Lvl_02_Basement_Lab")
  Script.PlayMusic("Haunts/Music/Adaptive/Bed 1")
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

  math.randomseed(os.time())
  store.DeathCounter = 0

  relic_spawn = Script.GetSpawnPointsMatching("Relic_Spawn")
  i = 1
  store.RelicPositions = {}
  for _, spawn in pairs(relic_spawn) do
    store.RelicPositions[i] = spawn.Pos
    i = i + 1
    Script.SpawnEntityAtPosition("Rift", spawn.Pos)
  end

end

function intrudersSetup()

  if IsStoryMode() then
    intruder_names = {"Collector", "Occultist", "Reporter"}
    intruder_spawn = Script.GetSpawnPointsMatching("Intruders_Start")
  end 

  for _, name in pairs(intruder_names) do
    ent = Script.SpawnEntitySomewhereInSpawnPoints(name, intruder_spawn, false)
  end

  -- Choose entry point here.
  Script.SaveStore()
end

function denizensSetup()
  Script.SetVisibility("denizens")

  spawn = Script.GetSpawnPointsMatching("Master_.*")
  Script.SpawnEntitySomewhereInSpawnPoints("Golem", spawn, false)

  store.MasterName = "Golem"
  ServitorEnts = 
  {
    {"Angry Shade", 1},
    {"Lost Soul", 1},
    {"Vengeful Wraith", 2},
  }  


  -- Just like before the user gets a ui to place these entities, but this
  -- time they can place more, and this time they go into spawn points that
  -- match anything with the prefix "Servitor_".
  setLosModeToRoomsWithSpawnsMatching("denizens", "Servitors_.*")
  placed = Script.PlaceEntities("Servitors_.*", ServitorEnts, 0, 8)
  MoveWaypoint()
end

function RoundStart(intruders, round)
  if store.execs == nil then
    store.execs = {}
  end
  if round == 1 then
    if intruders and not store.bDoneIntruderSetup then
      store.bDoneIntruderSetup = true
      intrudersSetup()     
    else
      if not store.bDoneDeniSetup then
        store.bDoneDeniSetup = true
        Script.DialogBox("ui/dialog/Lvl02/Lvl_02_Opening_Denizens.json")
        Script.FocusPos(Script.GetSpawnPointsMatching("Master_Start")[1].Pos)
        denizensSetup()
      end
    end
    Script.SetLosMode("intruders", "blind")
    Script.SetLosMode("denizens", "blind")

    if IsStoryMode() then
      DoTutorials()
    end

    store.execs = {}  
    Script.EndPlayerInteraction()
    return
  end

  if store.bCountdownTriggered and not intruders then
    Script.SetAp(MasterEnt(), 5)
    Script.SetMusicParam("tension_level", 0.7)
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


function OnMove(ent, path)

  return table.getn(path)
end

function OnAction(intruders, round, exec)
  if store.execs == nil then
    store.execs = {}
  end
  store.execs[table.getn(store.execs) + 1] = exec
 
  if exec.Action.Type == "Basic Attack" then
    if exec.Target.Name == store.MasterName and exec.Target.Hp <= 0 then
      --master is dead.  Intruders win.
      Script.DialogBox("ui/dialog/Lvl02/Lvl_02_Victory_Intruders.json")
    end
  end

  if  exec.Ent.Side.Intruder and GetDistanceBetweenPoints(exec.Ent.Pos, store.ActivePos) <= 3 and not store.bHarmedGolem then
    --The intruders got to the relic before the master.  They win.
    store.bHarmedGolem = true
    --Script.SetHp(MasterEnt(), MasterEnt().HpCur - 5)
    Script.DialogBox("ui/dialog/Lvl02/Lvl_02_Intruder_Reaches_Rift.json")
    StoreDamage(5, MasterEnt())
  end 

  if exec.Ent.Name == store.MasterName and GetDistanceBetweenPoints(exec.Ent.Pos, store.ActivePos) <= 3 then
    store.bHarmedGolem = false --reset this so the intruders can race to the next rift.
    MoveWaypoint()
    Script.DialogBox("ui/dialog/Lvl02/Lvl_02_Golem_Reaches_Rift.json")
    SpawnMinions(exec.Ent)
  end 

  --the intruders can only see the objective in LoS
  for _, ent in pairs(Script.GetAllEnts()) do
    if ent.Side.Intruder then   
      --can this intruder see the objective?
      for _, place in pairs(Script.GetLos(ent)) do
        if pointIsInSpawn(place, HighlightSpawn) then
          Script.SetVisibleSpawnPoints("intruders", HighlightSpawn.Name) 
        end
      end
    end
  end 

  if not AnyIntrudersAlive() then
    --game over, the denizens win.
    Script.DialogBox("ui/dialog/Lvl02/Lvl_02_Victory_Denizens.json")
  end

  --after any action, if this ent's Ap is 0, we can select the next ent for them
  -- if exec.Ent.ApCur == 0 then
  --   nextEnt = GetEntityWithMostAP(exec.Ent.Side)
  --   if nextEnt.ApCur > 0 then
  --     Script.SelectEnt(nextEnt)
  --   end
  -- end   
end

function MoveWaypoint()
  if not store.bWPSetupDone then
    store.bWPSetupDone = true
    store.ActivePos = {}
    store.ActivePos.X = 0
    store.ActivePos.Y = 0
  end
  bNewPointFound = false
  while not bNewPointFound do
    indexToUse = math.random(1, table.getn(store.RelicPositions))  
    if store.RelicPositions[indexToUse].X ~= store.ActivePos.X or store.RelicPositions[indexToUse].Y ~= store.ActivePos.Y then
      bNewPointFound = true
    end
  end
  StoreWaypoint("Relic", "denizens", store.RelicPositions[indexToUse], 3, false) 
  StoreWaypoint("RelicInt", "intruders", store.RelicPositions[indexToUse], 3, false) 
  store.ActivePos = store.RelicPositions[indexToUse]
  print("gtheckouttamove")
end

function SpawnMinions(mstrEnt)
  i = 1
  while i <= 6 do
    --Summon a new minion
    omgCounter = 1
    nRandomNumberOfAwesomenoess = math.random(200)
    nRandomCounter = 1
    for _, PossibleSpawn in pairs(Script.GetLos(mstrEnt)) do
      if nRandomCounter == nRandomNumberOfAwesomenoess then
        ent = StoreSpawn("Angry Shade", PossibleSpawn) 
        if ent.HpCur > 0 then  --we succeeded at spawning a dude
          Script.SetAp(ent, 0)
          i = i + 1
          break
        end
      else
        nRandomCounter = nRandomCounter + 1
      end
    end
    omgCounter = omgCounter + 1
    if omgCounter >= 50 then
      break
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
      if store.bCountdownTriggered then
        Script.DialogBox("ui/dialog/Lvl02/Lvl_02_Turns_Remaining_Denizens.json", {turns=store.nCountdown})
      else
        Script.DialogBox("ui/dialog/Lvl02/pass_to_denizens.json")
      end
    else
      if not bIntruderIntroDone then
        bIntruderIntroDone = true
        Script.DialogBox("ui/dialog/Lvl02/pass_to_intruders.json")
        Script.DialogBox("ui/dialog/Lvl02/Lvl_02_Opening_Intruders.json")
        bSkipOtherChecks = true
      end

      if store.bCountdownTriggered and not store.bShowedIntruderTimerMessage and not bSkipOtherChecks then
        store.bShowedIntruderTimerMessage = true
        Script.DialogBox("ui/dialog/Lvl02/Lvl_02_Countdown_Started_Intruders.json", {turns=store.nCountdown})
        bSkipOtherChecks = true
      end

      if store.bCountdownTriggered and not bSkipOtherChecks then  --timer is triggered and we've already intro'd it
        Script.DialogBox("ui/dialog/Lvl02/Lvl_02_Turns_Remaining_Intruders.json", {turns=store.nCountdown})
        bSkipOtherChecks = true
      end

      if not bSkipOtherChecks then  --if we haven't showed any of the other start messages, use the generic pass.
        Script.DialogBox("ui/dialog/Lvl02/pass_to_intruders.json")
      end

      if store.bCountdownTriggered then
        store.nCountdown = store.nCountdown - 1
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
      if exec.script_damage then
        doDamage(exec)
        bDone = true
      end                 
      if not bDone then
        Script.DoExec(exec)

        --will be used at turn start to try to reselect the last thing they acted with.
        if exec.Ent.Side == "intruders" then
          store.LastIntruderEnt = exec.Ent
        end 
        if exec.Ent.Side == "denizens" then
          store.LastDenizenEnt = exec.Ent
        end 
      end
    end
    store.execs = {}
  end
end

function MasterEnt()
  for _, ent in pairs(Script.GetAllEnts()) do
    if ent.Name == store.MasterName then
      return ent
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

function pointIsInSpawn(pos, sp)
  return pos.X >= sp.Pos.X and pos.X < sp.Pos.X + sp.Dims.Dx and pos.Y >= sp.Pos.Y and pos.Y < sp.Pos.Y + sp.Dims.Dy
end

function StoreSpawn(name, spawnPos)
  spawn_exec = {script_spawn=true, name=name, pos=spawnPos}
  store.execs[table.getn(store.execs) + 1] = spawn_exec
  return doSpawn(spawn_exec)
end

function doSpawn(spawnExec)
  return Script.SpawnEntityAtPosition(spawnExec.name, spawnExec.pos)
end

function StoreDespawn(ent)
  despawn_exec = {script_despawn=true, entity=ent}
  store.execs[table.getn(store.execs) + 1] = despawn_exec
end

function deSpawn(despawnExec)
  -- DeadBodyDump = Script.GetSpawnPointsMatching("Dead_People")
  -- DeadBodyDump[1].Pos.X = DeadBodyDump[1].Pos.X + store.DeathCounter
  -- store.DeathCounter = store.DeathCounter + 1
  -- Script.SetPosition(despawnExec.entity, DeadBodyDump[1].Pos)
end

function SelectCharAtTurnStart(side)
  bDone = false
  if store.LastIntruderEnt then
    if side.Intruder then
      Script.SelectEnt(store.LastIntruderEnt)
      bDone = true
    end
  end  
  if store.LastDenizenEnt and not bDone then
    if side.Denizen then    
      Script.SelectEnt(store.LastDenizenEnt)
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

function StoreDamage(amnt, ent)
  damage_exec = {script_damage=true, amount=amnt, entity=ent, hpcur=ent.HpCur}
  store.execs[table.getn(store.execs) + 1] = damage_exec
  doDamage(damage_exec)
end

function doDamage(damageExec)
  Script.SetHp(damageExec.entity, damageExec.hpcur - damageExec.amount)
end