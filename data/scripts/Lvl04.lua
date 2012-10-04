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
  Script.LoadHouse("Lvl_04_Catacombs")
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

  --spawn an initial beacon
  print("SCRIPT: ", "Beacon")
  Script.SpawnEntitySomewhereInSpawnPoints("Beacon", Script.GetSpawnPointsMatching("Beacon_Start"), false)

  --set these modular variables.
  store.MasterName = nil
  store.IntrudersPlacedBeaconLastTurn = false
  store.BeaconEnt = nil
  store.BeaconCount = 0
  store.nBeaconedRooms = 0
end

function intrudersSetup()
  if IsStoryMode() then
    intruder_names = {"Claire Murray", "Peter Chelios", "Tracy Latona"}
    intruder_spawn = Script.GetSpawnPointsMatching("Intruders_Start")
  -- else
  --   --permit all choices for normal vs play
  end 

  for _, name in pairs(intruder_names) do
    print("SCRIPT:", name)
    ent = Script.SpawnEntitySomewhereInSpawnPoints(name, intruder_spawn, false)
    Script.SetCondition(ent, "Pitch Black", true)
    Script.SetGear(ent, "Beacons")
  end

  -- Choose entry point here.
  Script.SaveStore()
end

function denizensSetup()
  
  Script.SetVisibility("denizens")
  master_spawn = Script.GetSpawnPointsMatching("Master_Start")
  Script.SpawnEntitySomewhereInSpawnPoints("Duchess Orlac", master_spawn, false)
  store.MasterName = "Duchess Orlac"
  Script.SelectEnt(GetMasterEnt())

  ServitorEnts = 
  {
    {"Umbral Fury", 1},
    {"Escaped Experiment", 2},
  }  

  -- Just like before the user gets a ui to place these entities, but this
  -- time they can place more, and this time they go into spawn points that
  -- match anything with the prefix "Servitor_".
  setLosModeToRoomsWithSpawnsMatching("denizens", "Servitors_Start")
  placed = Script.PlaceEntities("Servitors_Start", ServitorEnts, 0, 7)
end

function RoundStart(intruders, round)
  if not store.execs then
    store.execs = {}
  end

  if round == 1 then
    if intruders then
      intrudersSetup()     
    else
      Script.DialogBox("ui/dialog/Lvl04/Lvl_04_Opening_Denizens.json")
      Script.SetMusicParam("tension_level", 0.3)
      denizensSetup()
    end
    Script.SetLosMode("intruders", "blind")
    Script.SetLosMode("denizens", "blind")

    if IsStoryMode() then
      DoTutorials()
    end

    store.game = Script.SaveGameState()
    Script.EndPlayerInteraction()

    print("SCRIPT: End round 1")
    return
  end

  if intruders and round > 1 then
    SetActivatedRooms()
  end

  if store.IntrudersPlacedBeaconLastTurn then
    Script.SetVisibility("denizens")
    setLosModeToRoomsWithSpawnsMatching("denizens", "Servitors_Start")
    placed = Script.PlaceEntities("Servitors_Start", ServitorEnts, 0, ValueForReinforce())
    Script.SetLosMode("intruders", "blind")
    Script.SetLosMode("denizens", "blind")    
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

function BeaconCanSeeEnt(ent)
  return BeaconCanSeePoint(ent.Pos)
end

function BeaconCanSeePoint(pos)
  for _, beaconEnt in pairs(Script.GetAllEnts()) do
    if beaconEnt.Name == "Beacon" then
      for _, pointsBeaconCanSee in pairs(Script.GetLos(beaconEnt)) do
        if (pos.X == pointsBeaconCanSee.X) and (pos.Y == pointsBeaconCanSee.Y) then
          return true
        end 
      end
    end
  end 
  return false
end

function OnAction(intruders, round, exec)
  -- Check for players being dead here
  store.execs[table.getn(store.execs) + 1] = exec

  if exec.Action.Name == "Place Beacon" then
    --An intruder placed a beacon
    store.IntrudersPlacedBeaconLastTurn = true

    IlluminateDenizens()
  end

  if exec.Action.Name == "Hand Beacons" then
    StoreGear("Beacons", exec.Target)
    doGear(gear_exec)    
    store.BeaconEnt = exec.Target
    --remove the carrying beacons gear from the exec ent.    
    StoreGear("", exec.Ent)
    doGear(gear_exec)
  end

  --The big check: Have the intruders won.
  if exec.Action.Name == "Place Beacon" then
    SetActivatedRooms()

    if (store.Room1 or store.Room2 or store.Room3 or store.Room4 or store.Room5) and not store.bTalkedAboutBeaconInFirstRoom then
      store.bTalkedAboutBeaconInFirstRoom = true
      Script.SetMusicParam("tension_level", 0.4)  
      Script.DialogBox("ui/dialog/Lvl04/Lvl_04_First_Beacon_Intruders.json")
    end    

    if store.Room1 and store.Room2 and store.Room3 and store.Room4 and store.Room5 then
      --Intruders win
      Script.Sleep(2)
      Script.DialogBox("ui/dialog/Lvl04/Lvl_04_Victory_Intruders.json")
    end 
  end

  --deni's win when the intruder carrying the beacons is dead.
  if not AnyIntrudersAlive() then
    Script.Sleep(2)
    Script.DialogBox("ui/dialog/Lvl04/Lvl_04_Victory_Denizens.json")
  end 


  --if a denizen other than the master ended up in an illuminated area, taking an action sets their ap to 0
  if exec.Ent.Side.Denizen and not (exec.Ent.Name == store.MasterName) then
    --need to see if this ent moved into an illuminated space
    if BeaconCanSeeEnt(exec.Ent) then
    --yup. Need to give them the illuminated status and reduce their ap to at most 3.
      StoreCondition("Illuminated", exec.Ent, true)
      doCondition(condition_exec)
      Script.SetAp(exec.Ent, 0)
    else
    --nope.  remove illuminated from them
      StoreCondition("Illuminated", ent, false)
      doCondition(condition_exec)        
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

function SetActivatedRooms()
  store.Room1 = false
  store.Room2 = false
  store.Room3 = false
  store.Room4 = false
  store.Room5 = false

  StoreWaypoint("Room1", "intruders", (Script.GetSpawnPointsMatching("Room1_mid")[1].Pos), 5, false)
  StoreWaypoint("Room2", "intruders", (Script.GetSpawnPointsMatching("Room2_mid")[1].Pos), 5, false)
  StoreWaypoint("Room3", "intruders", (Script.GetSpawnPointsMatching("Room3_mid")[1].Pos), 5, false)
  StoreWaypoint("Room4", "intruders", (Script.GetSpawnPointsMatching("Room4_mid")[1].Pos), 5, false)
  StoreWaypoint("Room5", "intruders", (Script.GetSpawnPointsMatching("Room5_mid")[1].Pos), 5, false)

  store.BeaconCount = 0
  store.nBeaconedRooms = 0  --need 2 counters b/c they can put > 1 beacon per room

  for _, beaconEnt in pairs(Script.GetAllEnts()) do
    if beaconEnt.Name == "Beacon" then
      --Bacon ent!!!
      store.BeaconCount = store.BeaconCount + 1
      if pointIsInSpawns(beaconEnt.Pos, "Room1") and not store.Room1 then
        store.Room1 = true
        StoreWaypoint("Room1", "", "", "", true)
        store.nBeaconedRooms = store.nBeaconedRooms + 1
      end
      if pointIsInSpawns(beaconEnt.Pos, "Room2") and not store.Room2 then
        store.Room2 = true
        StoreWaypoint("Room2", "", "", "", true)
        store.nBeaconedRooms = store.nBeaconedRooms + 1
      end      
      if pointIsInSpawns(beaconEnt.Pos, "Room3") and not store.Room3 then
        store.Room3 = true
        StoreWaypoint("Room3", "", "", "", true)
        store.nBeaconedRooms = store.nBeaconedRooms + 1
      end
      if pointIsInSpawns(beaconEnt.Pos, "Room4") and not store.Room4 then
        store.Room4 = true
        StoreWaypoint("Room4", "", "", "", true)
        store.nBeaconedRooms = store.nBeaconedRooms + 1
      end
      if pointIsInSpawns(beaconEnt.Pos, "Room5") and not store.Room5 then
        store.Room5 = true
        StoreWaypoint("Room5", "", "", "", true)
        store.nBeaconedRooms = store.nBeaconedRooms + 1
      end   
    end
  end
end 

function IlluminateDenizens()
  for _, deniEnt in pairs(Script.GetAllEnts()) do
    if deniEnt.Side.Denizen  then
      --is this denizen entity within range of any beacons
      if BeaconCanSeeEnt(deniEnt) then
        --this deni is illuminated
        StoreCondition("Illuminated", deniEnt, true)
        doCondition(condition_exec)
      else
        StoreCondition("Illuminated", deniEnt, false)
        doCondition(condition_exec)
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
      Script.DialogBox("ui/dialog/Lvl04/pass_to_denizens.json", {rooms=(5-store.nBeaconedRooms)})

      if store.bTalkedAboutBeaconInFirstRoom and not store.bToldDenisAboutFirstBeacon then
        store.bToldDenisAboutFirstBeacon = true
        Script.DialogBox("ui/dialog/Lvl04/Lvl_04_First_Beacon_Denizens.json")
      end

    else
      store.IntrudersPlacedBeaconLastTurn = false

      if not store.InitialPassToIntrudersDone then
        store.InitialPassToIntrudersDone = true
        Script.DialogBox("ui/dialog/Lvl04/pass_to_intruders_initial.json")        
      else
        Script.DialogBox("ui/dialog/Lvl04/pass_to_intruders.json", {rooms=(5-store.nBeaconedRooms)})
      end


      if not store.bDoneIntruderIntro then
        store.bDoneIntruderIntro = true
        Script.DialogBox("ui/dialog/Lvl04/Lvl_04_Opening_Intruders.json")
        Script.FocusPos(Script.GetSpawnPointsMatching("Intruders_Start")[1].Pos)
      end
    end

    --if the Duchess Orlac is dead, respawn her
    ent = GetMasterEnt()
    if ent then
      if ent.HpCur <= 0 then
        StoreSpawn(store.MasterName, Script.GetSpawnPointsMatching("Master_Start")[1].Pos)        
      end
    else
      --no ent.  make one.
      StoreSpawn(store.MasterName, Script.GetSpawnPointsMatching("Master_Start")[1].Pos)
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
      if exec.script_gear then
        doGear(exec)
        bDone = true
      end
      if exec.script_condition then
        doCondition(exec)
        bDone = true
      end     
      if exec.script_teleport then
        doTeleport(exec)
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
end

function doSpawn(spawnExec)
  Script.SpawnEntityAtPosition(spawnExec.name, spawnExec.pos)
end

function StoreGear(name, ent)
  gear_exec = {script_gear=true, name=name, entity=ent}
  store.execs[table.getn(store.execs) + 1] = gear_exec
end

function doGear(gearExec)
  Script.SetGear(gearExec.entity, gearExec.name)
end


function StoreTeleport(ent, pos)
  teleport_exec = {script_teleport=true, entity=ent, position=pos}
  store.execs[table.getn(store.execs) + 1] = teleport_exec
end

function doTeleport(teleExec)
  Script.SetPosition(teleExec.entity, teleExec.position)
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
end

function deSpawn(despawnExec)
  if despawnExec.entity.HpMax then  --can only kill things that have hp
    Script.PlayAnimations(despawnExec.entity, {"defend", "killed"})
    Script.SetHp(despawnExec.entity, 0)
  end
  --DeadBodyDump = Script.GetSpawnPointsMatching("Dead_People")
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

function ValueForReinforce()
  --The denizens get to reinforce after each waypoint goes down.
  --They get 7 - (value of units on the board) + (1 for each beacon)

  nTotalValueOnBoard = 0
  for _, ent in pairs(Script.GetAllEnts()) do
    for _, entValue in pairs(ServitorEnts) do
      if ent.Name == entValue[1] then
        nTotalValueOnBoard = nTotalValueOnBoard + entValue[2]
      end 
    end
  end
  nAmountToReturn = (7 - nTotalValueOnBoard) + store.BeaconCount
  if nAmountToReturn <= 0 then
    nAmountToReturn = 0
  end
  return nAmountToReturn
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

function AnyIntrudersAlive()
  for _, ent in pairs(Script.GetAllEnts()) do
    if ent.Side.Intruder and ent.HpCur > 0 then
      return true
    end
  end
  return false  
end

function StoreWaypoint(wpname, wpside, wppos, wpradius, wpremove)
  print("SCRIPT: 1")
  waypoint_exec = {script_waypoint=true, name=wpname, side=wpside, pos=wppos, radius=wpradius, remove=wpremove}
  print("SCRIPT: 2")
  store.execs[table.getn(store.execs) + 1] = waypoint_exec
  print("SCRIPT: 3")
  doWaypoint(waypoint_exec)
  print("SCRIPT: 4")
end

function doWaypoint(waypointExec)
  if waypointExec.remove then
    return Script.RemoveWaypoint(waypointExec.name)
  else
    return Script.SetWaypoint(waypointExec.name, waypointExec.side, waypointExec.pos, waypointExec.radius)
  end
end

function GetMasterEnt()
  for _, ent in pairs(Script.GetAllEnts()) do
    if ent.Name == store.MasterName then
      return ent
    end
  end
end