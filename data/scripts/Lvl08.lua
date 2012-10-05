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
  Script.LoadHouse("Lvl_08_Nursery")

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

  spawn = Script.GetSpawnPointsMatching("Child_Start")
  ent = Script.SpawnEntitySomewhereInSpawnPoints("Chaos Child", spawn, false)
  filename = "ch08/" .. "Chaos Child" .. ".lua"
  Script.BindAi(ent, filename)
  store.MasterName = "Chaos Child"  --littler weird, since he's not actually a master.  But he's the one we care about.
  ServitorEnts = {} 
  ServitorEnts[1] = "Care Provider"
  ServitorEnts[2] = "Tutor"
  ServitorEnts[3] = "Foster Father"
  

  store.execs = {} 
  for index = 1, 6, 1 do
    SpawnNearChild(RandomServitor())
  end
  SpawnNearChild(ServitorEnts[3])  --spawn a big daddy

  --spawn intruders
  if IsStoryMode() then
    intruder_names = {"Detective", "Occultist", "Ghost Hunter"}
    intruder_spawn = Script.GetSpawnPointsMatching("Intruders_Start")
  end 

  for _, name in pairs(intruder_names) do
    ent = Script.SpawnEntitySomewhereInSpawnPoints(name, intruder_spawn, false)
    Script.SetGear(ent, "Pylons")
  end

  Script.SaveStore()

  Script.DialogBox("ui/dialog/Lvl08/Lvl_08_Opening_Denizens.json")
  Script.FocusPos(MasterEnt().Pos)   
  SelectNewOpPoint()
  store.execs = {} 
end

function RoundStart(intruders, round)
  if store.execs == nil then
    store.execs = {}
  end

  if not intruders then
    MawTrigger()
    MawAttack()
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

  --put a waypoint around the kid to show his escort
  if exec.Ent.Name == store.MasterName then
    StoreWaypoint("Kid", "intruders", MasterEnt().Pos, 1, false)
    StoreWaypoint("Kid2", "denizens", MasterEnt().Pos, 1, false)
  end

  --has the kid reached an op point?
  if exec.Ent.Name == store.MasterName then
    if GetDistanceBetweenEnts(exec.Ent, GetEntWithName("Toy")) <= 3 then
      --kid's gonna flip out!
      choice = Script.DialogBox("ui/dialog/Lvl08/Lvl_08_Waypoint_Choice_Denizens.json")
      if choice[1] == "Reinforce" then
        for index = 1, 4, 1 do        
          SpawnNearChild(RandomServitor())
          store.bReinforce = true
        end 
      end
      if choice[1] == "Maws" then
        for i = 1, 3, 1 do       
          SpawnRandomMaw()
          store.bMaws = true
        end
      end
      if choice[1] == "Turrets" then
        for i = 1, 4, 1 do        
          KillTurret()
          store.bTurrets = true
        end
      end

      SelectNewOpPoint()
    end
  end

  --are the intruders trying to win?
  if exec.Ent.Side.Intruder then
    if GetDistanceBetweenEnts(exec.Ent, MasterEnt()) <= 2 then
      --Is the kid vulnerable
      if not DeniEntNearKid() then
        --they did it.
        Script.Sleep(2)
        Script.DialogBox("ui/dialog/Lvl08/Lvl_08_Victory_Intruders.json")   
      end
    end
  end

  --warn the denizens if there aren't any near the kid.
  if not store.WarnedDenizensAboutVulnerableKid then
    if exec.Ent.Side.Denizen then
      if not DeniEntNearKid() then
        store.WarnedDenizensAboutVulnerableKid = true
        --!!!!!!!Dialogue to warn the denizens
      end
    end
  end


  if not AnyIntrudersAlive() then
    --game over, the denizens win.
    Script.Sleep(2)
    Script.DialogBox("ui/dialog/Lvl08/Lvl_08_Victory_Denizens.json")
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
      Script.DialogBox("ui/dialog/Lvl08/pass_to_denizens.json")
    else
      if not bIntruderIntroDone then
        bIntruderIntroDone = true
        Script.DialogBox("ui/dialog/Lvl08/pass_to_intruders.json")
        Script.DialogBox("ui/dialog/Lvl08/Lvl_08_Opening_Intruders.json")
        bSkipOtherChecks = true
      end

      if not bSkipOtherChecks then  --if we haven't showed any of the other start messages, use the generic pass.
        Script.DialogBox("ui/dialog/Lvl08/pass_to_intruders.json")
      end

      if store.bReinforce then
        Script.DialogBox("ui/dialog/Lvl08/Lvl_08_Choice_Reinforce_Intruders.json")
        store.bReinforce = false
      end
      if store.bMaws then
        Script.DialogBox("ui/dialog/Lvl08/Lvl_08_Choice_Maws_Intruders.json")
        store.bMaws = false
      end
      if store.bTurrets then
        Script.DialogBox("ui/dialog/Lvl08/Lvl_08_Choice_Turrets_Intruders.json")
        store.bTurrets = false
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

function DeniEntNearKid()
  master = MasterEnt()
  for _, ent in pairs(Script.GetAllEnts()) do
    if ent.Side.Denizen and ent.Name ~= store.MasterName then
      if GetDistanceBetweenEnts(ent, MasterEnt()) <= 5 then
        return true
      end
    end
  end
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


function SpawnNearChild(sNameOfThingToSpawn)
  childEnt = MasterEnt()

  i = 1
  while i <= 1 do
    --Summon a new minion
    omgCounter = 1
    nRandomNumberOfAwesomenoess = Script.Rand(200)
    nRandomCounter = 1
    for _, PossibleSpawn in pairs(Script.GetLos(childEnt)) do
      if nRandomCounter == nRandomNumberOfAwesomenoess then
        ent = StoreSpawn(sNameOfThingToSpawn, PossibleSpawn) 
        if ent then  --we succeeded at spawning a dude
          --Script.SetAp(ent, 0)
          i = i + 1
          break
        end
      else
        nRandomCounter = nRandomCounter + 1
      end
    end
    omgCounter = omgCounter + 1
    if omgCounter >= 200 then
      break
    end
  end
end

function RandomServitor()
  if Script.Rand(4) > 2 then
    return ServitorEnts[1]
  else
    return ServitorEnts[2]
  end
end

function SelectNewOpPoint()
  ent = GetEntWithName("Toy")
  if ent then
    startPos = ent.Pos
  else
    spawn = Script.GetSpawnPointsMatching("Toy_Start")[1]
    ent = StoreSpawn("Toy", spawn.Pos)
    startPos = spawn.Pos
  end
  newPos = startPos

  while newPos.X == startPos.X and newPos.Y == startPos.Y do
    spawns = Script.GetSpawnPointsMatching("Op_Point")
    newPos = spawns[Script.Rand(table.getn(spawns))].Pos
    StoreSetPos(ent, newPos)
    StoreWaypoint("OpPoint", "denizens", newPos, 1, false)
  end
end

function GetEntWithName(name)
  for _, ent in pairs(Script.GetAllEnts()) do
    if ent.Name == name then
      return ent
    end
  end
end

function MawTrigger()
  for _, spawn in pairs(Script.GetSpawnPointsMatching("Maw_Spawn")) do
    --if this spawn is near 3 intruder entities, we're mawing it up.
    intruderCount = 0
    for _, ent in pairs(Script.GetAllEnts()) do
      if ent.Side.Intruder then
        if GetDistanceBetweenPoints(ent.Pos, spawn.Pos) <= 7 then
          intruderCount = intruderCount + 1
        end
      end
    end

    if intruderCount >= 3 then  --flip out
      StoreSpawn("Maelstrom", spawn.Pos)
    end
  end
end

function MawAttack()
  for _, mawEnt in pairs(Script.GetAllEnts()) do
    if mawEnt.Name == "Maelstrom" then
      for _, ent in pairs(Script.GetAllEnts()) do
        if ent.Side.Intruder then
          if GetDistanceBetweenEnts(mawEnt, ent) <= 7 then
            StoreDamage(ent, 4)
          end
        end
      end
    end
  end
end

function SpawnRandomMaw()
  i = 1
  while i <= 1 do
    --Summon a new minion
    MawSpawns = Script.GetSpawnPointsMatching("Maw_Spawn")
    ent = StoreSpawn("Maelstrom", MawSpawns[Script.Rand(table.getn(MawSpawns))].Pos)
    if ent then --succeeded.  We can bail.
      i = i + 1
    end
  end
end

function KillTurret()
  for _, ent in pairs(Script.GetAllEnts()) do
    if ent.Name == "Pylon" then
      StoreDespawn(ent)
    end
  end
end
