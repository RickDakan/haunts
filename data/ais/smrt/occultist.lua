
function GetLeader()
  if Me.Master.Leader then
    return Me
  end
  allies = Utils.NearestNEntities(3, "intruder")
  if table.getn(allies) == 0 then
    return Me
  end
  for _, ally in pairs(allies) do
    if ally.Master.Leader then
      return ally
    end
  end
  return Me
end

-- First make sure allies are ok
-- Then deal with enemies
-- Then check for an untouched relic in the room you're in
-- Last look for a nearby unexplored room and head in that direction
function Think()
  print("OCC:")
  leader = GetLeader()
  print("OCC:", leader.Name)
  if Utils.RangedDistBetweenEntities(Me, leader) < 5 then
    while CrushEnemies() do
    end

    while CheckForRelic() do
    end
  end

  print("OCC: Here1")
  if leader.id == Me.id then
  print("OCC: Here2")
    while GoToUnexploredRoom() do
    end
  else
  print("OCC: Here3")
    while Follow(leader) do
    end
  end
end

function Follow(leader)
  print("OCC: Following", leader.Name)
  ps = Utils.AllPathablePoints(Me.Pos, leader.Pos, 1, 3)
  pos = Me.Pos
  new = Do.Move(ps, 1000)
  return pos.X ~= new.X or pos.Y ~= new.Y
end

function SupportAllies()
  -- First make sure I'm always buffed
  if not Me.Conditions["Psychic Shroud"] then
    Do.BasicAttack("Psychic Shroud", Me)
    return true
  end

  -- Now make sure my teammates are buffed if they are in trouble
  allies = Utils.NearestNEntities(3, "intruder")
  for _, ally in pairs(allies) do
    if ally.HpCur <= 5 and not ally.Conditions["Psychic Shroud"] then
      Do.BasicAttack("Psychic Shroud", ally)
      return true
    end
  end

  return false
end

function CrushEnemies()
  enemies = Utils.NearestNEntities(5, "denizen")
  for _, enemy in pairs(enemies) do
    dist = Utils.RangedDistBetweenEntities(Me, enemy)
    if dist > 10 then
      return false
    end
    attack = "Exorcise"
    Do.BasicAttack(attack, enemy)
    return true
  end
  return false
end

function CheckForRelic()
  objects = Utils.NearestNEntities(3, "object")
  if table.getn(objects) == 0 then
    return false
  end

  object = nil
  for i,obj in pairs(objects) do
    if obj.State == "ready" then
      object = obj
      break
    end
  end
  if object == nil then
    return false
  end

  ps = Utils.AllPathablePoints(Me.Pos, object.Pos, 1, Me.Actions.Interact.Range)
  Do.Move(ps, 1000)
  Do.InteractWithObject(object)
  return true
end

-- Moves towards an unexplored room, returns false if it couldn't move towards
-- it, or if there were no more unexplored rooms to go to, true otherwise.
function GoToUnexploredRoom()
  unexplored = Utils.NearbyUnexploredRoom()
  if not unexplored then
    return false  -- No more rooms to explore
  end

  current = Utils.RoomContaining(Me)
  path = Utils.RoomPath(current, unexplored)
  if table.getn(path) == 0 then
    return false  -- No room path to the unexplored room - shouldn't happen
  end

  target = path[1]
  doors = Utils.AllDoorsBetween(current, target)
  if table.getn(doors) == 0 then
    return false   -- No doors to the next room, also shouldn't happen
  end

  -- If the door is closed then go to it and open it.
  if not Utils.DoorIsOpen(doors[1]) then
    cpos = Me.Pos
    ps = Utils.DoorPositions(doors[1])
    Do.Move(ps, 1000)
    if cpos.X == Me.Pos.X and cpos.Y == Me.Pos.Y then
      Do.DoorToggle(doors[1])
    end
    return true
  end

  -- Now that we know the door is open, step into the next room.
  ps = Utils.RoomPositions(target)
  res = Do.Move(ps, 1000)

  return true
end
