function IsStoryMode()
  return true
end

function DoTutorials()
  --We should totally do some tutorials here.
  --It would be super cool.
end

function Init(data)
  side_choices = Script.ChooserFromFile("ui/start/versus/side.json")

  Script.LoadHouse("Lvl_09_Gantlet")

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

  --spawn denizens
  Script.SetVisibility("denizens")

  spawn = Script.GetSpawnPointsMatching("Denizen_Start")
  ServitorEnts = {} 
  ServitorEnts[1] = "Major Torgo Klortho"
  ServitorEnts[2] = "Dr. Beatrice Klortho"
  Script.SpawnEntitySomewhereInSpawnPoints(ServitorEnts[1], spawn, false)
  Script.SpawnEntitySomewhereInSpawnPoints(ServitorEnts[2], spawn, false)
  
  intruder_names = {"Genevieve Torres", "Nikola Fury", "Doc Pulver"}
  intruder_spawn = Script.GetSpawnPointsMatching("Intruders_Start")
  for _, name in pairs(intruder_names) do
    ent = Script.SpawnEntitySomewhereInSpawnPoints(name, intruder_spawn, false)
  end
  spawn = Script.GetSpawnPointsMatching("Device_Spawn")
  ent = Script.SpawnEntitySomewhereInSpawnPoints("Belascic Purgation Engine", spawn, false)
  store.DeviceName = "Belascic Purgation Engine"
  Script.SetWaypoint("Device", "denizens", ent.Pos, 2, false)
  Script.SetWaypoint("Device", "intruders", ent.Pos, 2, false)

  Script.SaveStore()

  Script.DialogBox("ui/dialog/Lvl09/Lvl_09_Opening_Denizens.json")
  Script.FocusPos(MasterEnt().Pos)   
  SelectNewOpPoint()
  store.execs = {}

  store.TurnsRemaining = 30
end

function RoundStart(intruders, round)
  if store.execs == nil then
    store.execs = {}
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
  
  if exec.Ent.Side.Denizen then
    --if a deni ent attacks the Device, kill that ent and aoe the room.
    if GetDistanceBetweenEnts(exec.Ent, DeviceEnt()) <= 2 then
      StoreDespawn(exec.Ent)
      DeviceAoe()
    end

    --If a bane attacks a turret, kill it and damage everything around it.
    if exec.Ent.Name == "Sacrificial Orb" then
      if IsBaneNearIntruder(exec.Ent) then
        BaneAoe(exec.Ent)
        StoreDespawn(exec.Ent)
      end
    end
  end

  if DeviceEnt().HpCur <= 0 then
    --game over, the denizens win.
    Script.Sleep(2)
    Script.DialogBox("ui/dialog/Lvl09/Lvl_09_Victory_Denizens.json")
  end

  --so does killing all the denizens
  if not AnyDenizensAlive() then
    --game over, the denizens win.
    Script.Sleep(2)
    Script.DialogBox("ui/dialog/Lvl09/Lvl_09_Victory_Intruders.json")
  end  

  --if they just summoned, we need to find the thing they summoned, bind it's ai and kill it's ap.
  if exec.Action.Type == "Summon" and exec.Ent.Side.Denizen then
    --get all minions, bind their ai's and set their ap to 0
    for _, ent in pairs(Script.GetAllEnts()) do
      if ent.Name == "Swarming Shade" then
        Script.BindAi(ent, "ch09/shade.lua")
        Script.SetAp(ent, 0)
      end
      if ent.Name == "Golem Prototype" then
        Script.BindAi(ent, "ch09/golem.lua")
        Script.SetAp(ent, 0)
      end
      if ent.Name == "Sacrificial Orb" then
        Script.BindAi(ent, "ch09/orb.lua")
        Script.SetAp(ent, 0)
      end        
      if ent.Name == "Serum Fiend" then
        Script.BindAi(ent, "ch09/fiend.lua")
        Script.SetAp(ent, 0)
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

  bSkipOtherChecks = false  --Resets this every round

  if store.side == "Humans" then
    Script.ShowMainBar(false)
    Script.SetLosMode("intruders", "entities")
    Script.SetLosMode("denizens", "entities")
    if intruders then
      Script.SetVisibility("denizens")
    else
      Script.SetVisibility("intruders")
    end

    if intruders then
      Script.DialogBox("ui/dialog/Lvl09/pass_to_denizens.json")
    else
      if not bIntruderIntroDone then
        bIntruderIntroDone = true
        Script.DialogBox("ui/dialog/Lvl09/pass_to_intruders.json")
        Script.DialogBox("ui/dialog/Lvl09/Lvl_09_Opening_Intruders.json")
        bSkipOtherChecks = true
      end

      if not bSkipOtherChecks then  --if we haven't showed any of the other start messages, use the generic pass.
        Script.DialogBox("ui/dialog/Lvl09/pass_to_intruders.json")
      end          
    end

    if intruders then
      store.TurnsRemaining = store.TurnsRemaining - 1
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
      if exec.script_setpos then
        doSetPos(exec)
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

function DeviceEnt()
  for _, ent in pairs(Script.GetAllEnts()) do
    if ent.Name == store.DeviceName then
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

function AnyDenizensAlive()
  for _, ent in pairs(Script.GetAllEnts()) do
    if ent.Side.Denizen and ent.HpCur > 0 then
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
  Script.SetHp(despawnExec.entity, 0)
  Script.PlayAnimations(despawnExec.entity, {"defend", "killed"})
  DeadBodyDump = Script.GetSpawnPointsMatching("Dead_People")
  Script.SetPosition(despawnExec.entity, DeadBodyDump[1].Pos)
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

function StoreDamage(ent, amnt)
  if ent.HpCur - amnt <= 0 then
    StoreDespawn(ent)
    Script.SetPosition(ent, Script.GetSpawnPointsMatching("Dead_People")[1].Pos) --!!!!stop gap.  The despawn wasn't working on the turn that we did damage.
    return
  end
  damage_exec = {script_damage=true, amount=amnt, entity=ent, hpcur=ent.HpCur}
  store.execs[table.getn(store.execs) + 1] = damage_exec
  doDamage(damage_exec)
end

function doDamage(damageExec)
  Script.SetHp(damageExec.entity, damageExec.hpcur - damageExec.amount)
end

function StoreSetPos(ent, pos)
  setPos_exec = {script_setpos=true, entity=ent, newPos=pos}
  store.execs[table.getn(store.execs) + 1] = setPos_exec
  doSetPos(setPos_exec)
end

function doSetPos(setPosExec)
  Script.SetPosition(setPosExec.entity, setPosExec.newPos)
end

function GetEntWithName(name)
  for _, ent in pairs(Script.GetAllEnts()) do
    if ent.Name == name then
      return ent
    end
  end
end

function DeviceAoe()
  device = DeviceEnt()
  for _, ent in pairs(Script.GetAllEnts()) do
    if GetDistanceBetweenEnts(ent, device) <= 6 then
      if ent.Name == store.DeviceName then
        StoreDamage(ent, 50)
      else
        StoreDamage(ent, 4)
      end
    end
  end
end

function IsBaneNearIntruder(baneEnt)
  for _, ent in pairs(Script.GetAllEnts()) do
    if ent.Side.Intruder then
      if GetDistanceBetweenEnts(baneEnt, ent) <= 1 then
        return true
      end
    end
  end
end

function BaneAoe(baneEnt)
  for _, ent in pairs(Script.GetAllEnts()) do
    if GetDistanceBetweenEnts(ent, baneEnt) <= 4 then
      StoreDamage(ent, 4)
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