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
  Script.LoadHouse("Lvl_07_Waxworks")  

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

  wax_intruder_spawns = Script.GetSpawnPointsMatching("Wax_Intruder")
  store.ActiveWaypoints = {}
  i = 1
  for _, spawn in pairs(wax_intruder_spawns) do
    if i < 10 then
      str = "Wax_Intruder0"
    else
      str = "Wax_Intruder"
    end
    Script.SpawnEntityAtPosition(str .. i, spawn.Pos)
    Script.SetWaypoint("Waypoint" .. i , "intruders", spawn.Pos, 1)
    store.ActiveWaypoints[i] = i
    i = i + 1
    if i > 18 then 
      i = 1
    end
  end
  store.WaypointCount = i

  --we will incorporate some randomness here.
  store.MaxWaxIntruders = WaxIntruderCount()
end

function intrudersSetup()

  intruder_names = {"Reporter", "Teen"}
  intruder_spawn = Script.GetSpawnPointsMatching("Intruders_Start") 

  for _, name in pairs(intruder_names) do
    ent = Script.SpawnEntitySomewhereInSpawnPoints(name, intruder_spawn, false)
  end

  Script.SaveStore()
end

function denizensSetup()
  Script.SetVisibility("denizens")

  master_spawn = Script.GetSpawnPointsMatching("Escape")
  ent = Script.SpawnEntitySomewhereInSpawnPoints("The Maestro", master_spawn, false)
  Script.FocusPos(ent.Pos)

  store.MasterName = "The Maestro"
  ServitorEnts = 
  {
    {"Mimic", 1},
  }  

  setLosModeToRoomsWithSpawnsMatching("denizens", "Wax_Denizen_1")
  placed = Script.PlaceEntities("Wax_Denizen_1", ServitorEnts, 0,3)
  setLosModeToRoomsWithSpawnsMatching("denizens", "Wax_Denizen_2")
  placed = Script.PlaceEntities("Wax_Denizen_2", ServitorEnts, 0,3)
  setLosModeToRoomsWithSpawnsMatching("denizens", "Wax_Denizen_3")
  placed = Script.PlaceEntities("Wax_Denizen_3", ServitorEnts, 0,4)

  SaveDeniPositions()
  --put wax dudes in the rest of the deni spawnpoints
  i = 1
  storagePos = Script.GetSpawnPointsMatching("Wax_Storage")[1].Pos
  for _, spawn in pairs(Script.GetSpawnPointsMatching("Wax_Denizen.*")) do
    if i < 10 then
      str = "Wax_Denizen0"
    else
      str = "Wax_Denizen"
    end
    if not GetEntAtPos(spawn.Pos) then
      Script.SpawnEntityAtPosition(str .. i, spawn.Pos)
    else
      --There's a mimic here.  Store the wax version
      Script.SpawnEntityAtPosition(str .. i, storagePos)
      storagePos.X = storagePos.X + 1
    end
    i = i + 1
    if i > 20 then 
      i = 1
    end
  end
end

function RoundStart(intruders, round)
  if store.execs == nil then
    store.execs = {}
  end
  side = {Intruder = intruders, Denizen = not intruders, Npc = false, Object = false}

  if round == 1 then
    if intruders then
      intrudersSetup()     
    else
      Script.DialogBox("ui/dialog/Lvl07/Lvl_07_Opening_Denizens.json")
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

  if not intruders then
    masterEnt = GetEntWithName(store.MasterName)
    --any mimics that start out near the maestro have extra AP
    for _, ent in pairs(Script.GetAllEnts()) do
      if ent.Name == "Mimic" then
        if GetDistanceBetweenEnts(ent, masterEnt) <= 10 then
          Script.SetAp(ent, 12)
        end
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

function OnMove(ent, path)
  return table.getn(path)
end

function OnAction(intruders, round, exec)
  -- Check for players being dead here
  if store.execs == nil then
    store.execs = {}
  end
  store.execs[table.getn(store.execs) + 1] = exec

  --if the intruders got near a wax intruder, check and see if this is real intruder.
  if intruders and not store.bVictimFound then
    ent = FreeingEnt(exec.Ent)
    if ent then
      index = tonumber(string.sub(ent.Name, -2))
      if store.ActiveWaypoints[index] > 0 then
          --remove this waypoint
          StoreWaypoint("Waypoint" .. index, "", "", "", true)
          store.ActiveWaypoints[index] = 0
        if IsThisTheIntruder() then
          store.bVictimFound = true
          Script.DialogBox("ui/dialog/Lvl07/Lvl_07_Intruders_Found_Victim_Intruders.json")
          --remove all the waypoints.  select an exit.  Mark it.
          for i = 1, store.MaxWaxIntruders, 1 do
            if store.ActiveWaypoints[i] > 0 then
              StoreWaypoint("Waypoint" .. i, "", "", "", true)    
            end
          end
          i = Script.Rand(3)
          spawns = Script.GetSpawnPointsMatching("Escape")
          store.EscapePoint = spawns[i].Pos
          StoreWaypoint("Escape", "intruders", store.EscapePoint, 3, false)
          --replace the wax intruder with the real one.
          pos = ent.Pos
          StoreDespawn(ent)
          StoreSpawn("Collector", pos)
        end
      end
    end
  end

  --The big check: Have the intruders won.
  if store.bVictimFound then
    if intruders then
      if GetDistanceBetweenPoints(exec.Ent.Pos, store.EscapePoint) <= 3 then
        --The intruders got to the exit.  Game over.
        Script.Sleep(2)
        Script.DialogBox("ui/dialog/Lvl07/Lvl_07_Victory_Intruders.json")    
      end
    end
  end  

  --deni's win when all intruders dead.
  if not AnyIntrudersAlive() then
    Script.Sleep(2)
    Script.DialogBox("ui/dialog/Lvl07/Lvl_07_Victory_Denizens.json")
  end 

  --after any action, if this ent's Ap is 0, we can select the next ent for them
  if exec.Ent.ApCur == 0 then
    nextEnt = GetEntityWithMostAP(exec.Ent.Side)
    if nextEnt.ApCur > 0 then
      if exec.Action.Type ~= "Move" then
        Script.Sleep(2)
      end
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
      Script.DialogBox("ui/dialog/Lvl07/pass_to_denizens.json")

      if store.bVictimFound and not store.bToldDenizensAboutVictim then
        store.bToldDenizensAboutVictim = true
        Script.DialogBox("ui/dialog/Lvl07/Lvl_07_Intruders_Found_Victim_Denizens.json")
      end 
    else
      Script.DialogBox("ui/dialog/Lvl07/pass_to_intruders.json")
      if not store.bDoneIntruderIntro then
        store.bDoneIntruderIntro = true
        Script.DialogBox("ui/dialog/Lvl07/Lvl_07_Opening_Intruders.json")
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
      if exec.script_setPos then
        doSetPos(exec)
        bDone = true
      end         
      if exec.script_waypoint then
        doWaypoint(exec)
        bDone = true
      end      
      if not bDone then
        Script.DoExec(exec)

        --a mimic attacking will prevent it from switching back to wax.
        if exec.Ent.Name == "Mimic" then
          if exec.Action.Type == "Basic Attack" then
            Script.SetCondition(exec.Ent, "Revealed", true)
          end
        end

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

  store.execs = {}
  SwapDenizens(intruders)
  store.execs = {}
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

function StoreDespawn(ent)
  despawn_exec = {script_despawn=true, entity=ent}
  store.execs[table.getn(store.execs) + 1] = despawn_exec
  deSpawn(despawn_exec)
end

function deSpawn(despawnExec)
  if despawnExec.entity.Name == "Mimic" then
    Script.SetHp(despawnExec.entity, 0)
    Script.PlayAnimations(despawnExec.entity, {"defend", "killed"})
    Script.SetPosition(despawnExec.entity, Script.GetSpawnPointsMatching("Dead_People")[1].Pos)
  else
    if string.find(despawnExec.entity.Name, "Wax_Intruder") then  --only happens when the intruders find the victim
      temp = Script.GetSpawnPointsMatching("Dead_People")[1]
      temp.Pos.X = temp.Pos.X + 1
      Script.SetPosition(despawnExec.entity, temp.Pos)
    else  --storing a wax denizen
      --Despawning a wax figure.  these are objects and need to be stored.
      storage = Script.GetSpawnPointsMatching("Wax_Storage")[1]
      storage.Pos.X = storage.Pos.X + store.DeathCounter
      store.DeathCounter = store.DeathCounter + 1
      Script.SetPosition(despawnExec.entity, storage.Pos)
    end
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

function AnyIntrudersAlive()
  for _, ent in pairs(Script.GetAllEnts()) do
    if ent.Side.Intruder and ent.HpCur > 0 then
      return true
    end
  end
  return false  
end

function StoreWaypoint(wpname, wpside, wppos, wpradius, wpremove)
  waypoint_exec = {script_waypoint=true, name=wpname, side=wpside, pos=wppos, radius=wpradius, remove=wpremove}
  store.execs[table.getn(store.execs) + 1] = waypoint_exec
  doWaypoint(waypoint_exec)
end

function doWaypoint(waypointExec)
  if waypointExec.remove then
    print("happy")
    return Script.RemoveWaypoint(waypointExec.name)
  else
    return Script.SetWaypoint(waypointExec.name, waypointExec.side, waypointExec.pos, waypointExec.radius)
  end
end

function StoreSetPos(ent, setPos)
  setpos_exec = {script_setPos=true, entity=ent, pos=setPos}
  store.execs[table.getn(store.execs) + 1] = setpos_exec
  return doSetPos(setpos_exec)
end

function doSetPos(SetPosExec)
   return Script.SetPosition(SetPosExec.entity, SetPosExec.pos)
end

function IsThisTheIntruder()
  if not store.nCount then
    store.nCount = WaxIntruderCount()
  else
    store.nCount = store.nCount - 1
  end
  nMax = store.MaxWaxIntruders
  nWinner = Script.Rand(nMax)
  --last wax intruder.  Has to be winner.
  if store.nCount == 1 then
    return true
  end
  if store.nCount == nMax then
    return false  --first one can't be the victim
  end
  for i = 1, nMax - store.nCount, 1 do 
    if i == nWinner then
      return true
    end
    i = i + 1
  end
  return false
end

function FreeingEnt(intruderEnt)
  for _, ent in pairs(Script.GetAllEnts()) do
    if string.find(ent.Name, "Wax_Intruder") then
      if GetDistanceBetweenEnts(intruderEnt, ent) <= 3 then
        return ent
      end
    end
  end
end

function WaxIntruderCount()
  i = 0
  for _, ent in pairs(Script.GetAllEnts()) do
    if string.find(ent.Name, "Wax_Intruder") then
      i = i + 1
    end 
  end

  return i
end

function SaveDeniPositions()
  store.DeniPositions = {}
  store.MimicCount = 0
  --Now we have to get creative.
  for _, ent in pairs(Script.GetAllEnts()) do
    if ent.Name == "Mimic" then
      store.DeniPositions[table.getn(store.DeniPositions) + 1] = ent.Pos
      store.MimicCount = store.MimicCount + 1 
    end
  end
end

function SwapDenizens(bToEnts)
  if bToEnts then  --move wax into storage.  Spawn mimics
    store.DeathCounter = 0
    for _, pos in pairs(store.DeniPositions) do 
      if GetEntAtPos(pos).Name ~= "Mimic" then   
        StoreDespawn(GetEntAtPos(pos))
        StoreSpawn("Mimic", pos)
      else
        store.DeathCounter = store.DeathCounter + 1
      end
    end
  else --despawn mimics.  Get out wax
    SaveDeniPositions()
    StoragePos = Script.GetSpawnPointsMatching("Wax_Storage")[1].Pos
    for _, ent in pairs(Script.GetAllEnts()) do
      if ent.Name == "Mimic" then
        if not ent.Conditions["Revealed"] then
          pos = ent.Pos
          StoreDespawn(ent)
          StoreSetPos(GetEntAtPos(StoragePos), pos)
        end
        StoragePos.X = StoragePos.X + 1
      end
    end
  end
end

function GetEntAtPos(pos)
  for _, ent in pairs(Script.GetAllEnts()) do
    if ent.Pos.X == pos.X and ent.Pos.Y == pos.Y then
      return ent 
    end
  end
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
