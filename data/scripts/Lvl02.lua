function setLosModeToRoomsWithSpawnsMatching(side, pattern)
  sp = Script.GetSpawnPointsMatching(pattern)
  rooms = {}
  for i, spawn in pairs(sp) do
    rooms[i] = Script.RoomAtPos(spawn.Pos)
  end
  Script.SetLosMode(side, rooms)
end

function Side()
  if Net.Active() then
    return Net.Side()
  end
  return store.side
end

function OnStartup()
  Script.PlayMusic("Haunts/Music/Adaptive/Bed 1")
  Script.SetMusicParam("tension_level", store.tension)
  if Net.Active() then
    if Side() == "Denizens" then
      Script.SetVisibility("denizens")
    else
      Script.SetVisibility("intruders")
    end
  end
end

function Init(data)
  if Net.Active() then
    -- The Init() function will only be run by the player starting the game who
    -- is necessarily the Denizens player.
    side_choices = {"Denizens"}
  else
    side_choices = Script.ChooserFromFile("ui/start/versus/side.json")
  end

  -- check data.map == "random" or something else
  Script.LoadHouse("Lvl_02_Basement_Lab")
  Script.PlayMusic("Haunts/Music/Adaptive/Bed 1")
  Script.SetMusicParam("tension_level", 0.1)   
  store.tension = 0.1

  store.side = side_choices[1]
  if Side() == "Humans" or Net.Active() then
    Script.BindAi("denizen", "human")
    Script.BindAi("minions", "minions.lua")
    Script.BindAi("intruder", "human")
  else
    if Side() == "Denizens" then
      Script.BindAi("denizen", "human")
      Script.BindAi("minions", "minions.lua")
      Script.BindAi("intruder", "ch02/intruders.lua")
    end
    if Side() == "Intruders" then
      Script.BindAi("denizen", "ch02/denizens.lua")
      Script.BindAi("minions", "minions.lua")
      Script.BindAi("intruder", "human")
    end
  end

  store.DeathCounter = 0

  relic_spawn = Script.GetSpawnPointsMatching("Relic_Spawn")
  i = 1
  store.RelicPositions = {}
  for _, spawn in pairs(relic_spawn) do
    store.RelicPositions[i] = spawn.Pos
    i = i + 1
    Script.SpawnEntityAtPosition("Rift", spawn.Pos)
  end


  intruder_names = {"Collector", "Occultist", "Reporter"}
  intruder_spawn = Script.GetSpawnPointsMatching("Intruders_Start")
  for _, name in pairs(intruder_names) do
    ent = Script.SpawnEntitySomewhereInSpawnPoints(name, intruder_spawn, false)
  end
  Script.SaveStore()  

  store.execs = {}
  Script.DialogBox("ui/dialog/Lvl02/Lvl_02_Opening_Denizens.json")
  Script.FocusPos(Script.GetSpawnPointsMatching("Master_Start")[1].Pos)

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

  setLosModeToRoomsWithSpawnsMatching("denizens", "Servitors_.*")
  placed = Script.PlaceEntities("Servitors_.*", ServitorEnts, 0, 8)
  MoveWaypoint()
  store.execs = {}  
end


function RoundStart(intruders, round)
  if store.execs == nil then
    store.execs = {}
  end

  if Net.Active() then
    if Side() == "Denizens" then
      Script.SetVisibility("denizens")
      denizensOnRound()
    else
      Script.SetVisibility("intruders")
      intrudersOnRound()
    end
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

  if Net.Active() then
    store.game = Script.SaveGameState()
    print("Update State Round/Intruders: ", round, intruders)
    Net.UpdateState(store.game)
  else
    store.game = Script.SaveGameState()
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
    Script.SetMusicParam("tension_level", 0.6)
    Script.DialogBox("ui/dialog/Lvl02/Lvl_02_Intruder_Reaches_Rift.json")
    StoreDamage(25, MasterEnt())
  end 

  if exec.Ent.Name == store.MasterName and GetDistanceBetweenPoints(exec.Ent.Pos, store.ActivePos) <= 3 then
    store.bHarmedGolem = false --reset this so the intruders can race to the next rift.
    MoveWaypoint()
    Script.SetMusicParam("tension_level", 0.3)
    Script.DialogBox("ui/dialog/Lvl02/Lvl_02_Golem_Reaches_Rift.json")
    SpawnMinions(exec.Ent)
  end 

  if not AnyIntrudersAlive() then
    --game over, the denizens win.
    Script.DialogBox("ui/dialog/Lvl02/Lvl_02_Victory_Denizens.json")
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

function MoveWaypoint()
  if not store.bWPSetupDone then
    store.bWPSetupDone = true
    store.ActivePos = {}
    store.ActivePos.X = 0
    store.ActivePos.Y = 0
  end
  bNewPointFound = false
  while not bNewPointFound do
    indexToUse = Script.Rand(table.getn(store.RelicPositions))  
    if store.RelicPositions[indexToUse].X ~= store.ActivePos.X or store.RelicPositions[indexToUse].Y ~= store.ActivePos.Y then
      bNewPointFound = true
    end
  end
  StoreWaypoint("Relic", "denizens", store.RelicPositions[indexToUse], 3, false) 
  StoreWaypoint("RelicInt", "intruders", store.RelicPositions[indexToUse], 3, false) 
  store.ActivePos = store.RelicPositions[indexToUse]
end

function SpawnMinions(mstrEnt)
  i = 1
  while i <= 6 do
    --Summon a new minion
    omgCounter = 1
    nRandomNumberOfAwesomenoess = Script.Rand(200)
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
  if Net.Active() then
    Net.UpdateExecs(Script.SaveGameState(), store.execs)
    Script.ShowMainBar(false)
    Net.Wait()
    -- cur = Script.SaveGameState()
    state, execs = Net.LatestStateAndExecs()
    DoPlayback(state, execs)
    Script.ShowMainBar(true)
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
      Script.DialogBox("ui/dialog/Lvl02/pass_to_denizens.json")
    else
      if not bIntruderIntroDone then
        bIntruderIntroDone = true
        Script.DialogBox("ui/dialog/Lvl02/pass_to_intruders.json")
        Script.DialogBox("ui/dialog/Lvl02/Lvl_02_Opening_Intruders.json")
        Script.SetMusicParam("tension_level", 0.2)
        store.tension = 0.2
        bSkipOtherChecks = true
      end

      if not bSkipOtherChecks then  --if we haven't showed any of the other start messages, use the generic pass.
        Script.DialogBox("ui/dialog/Lvl02/pass_to_intruders.json")
      end
    end

    Script.SetLosMode("intruders", "entities")
    Script.SetLosMode("denizens", "entities")
    DoPlayback(store.game, store.execs)

    if intruders then
      denizensOnRound()
    else
      intrudersOnRound()
    end    
    store.execs = {}
  end
end

function DoPlayback()

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
end

function intrudersOnRound()
  if not bIntruderIntroDone then
    bIntruderIntroDone = true
    Script.DialogBox("ui/dialog/Lvl02/Lvl_02_Opening_Intruders.json")
    Script.SetMusicParam("tension_level", 0.2)
    store.tension = 0.2
  end
end

function denizensOnRound()
  
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