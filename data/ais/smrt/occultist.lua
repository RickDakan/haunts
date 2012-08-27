
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
  leader = GetLeader()
  if Utils.RangedDistBetweenEntities(Me, leader) < 5 then
    while CrushEnemies() do
    end

    while CheckForRelic() do
    end
  end

  if leader.id == Me.id then
    if GoToUnexploredRoom() then
      Think()
    end
  else
    while Follow(leader) do
    end
  end
end

function Follow(leader)
  ps = Utils.AllPathablePoints(Me.Pos, leader.Pos, 1, 3)
  valid, pos = Do.Move(ps, 1000)
  return valid and pos
end

function CrushEnemies()
  enemies = Utils.NearestNEntities(5, "denizen")
  for _, enemy in pairs(enemies) do
    dist = Utils.RangedDistBetweenEntities(Me, enemy)
    if dist > 10 then
      return false
    end
    attack = "Exorcise"
    return Do.BasicAttack(attack, enemy)
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
  valid, pos = Do.Move(ps, 1000)
  if not valid then
    return false
  end
  return Do.InteractWithObject(object)
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
    ps = Utils.DoorPositions(doors[1])
    complete = Do.Move(ps, 1000)
    if not complete then
      return false
    end
    if not Do.DoorToggle(doors[1]) then
      return false
    end
  end

  -- Now that we know the door is open, step into the next room.
  ps = Utils.RoomPositions(target)
  complete, other = Do.Move(ps, 1000)
  return complete
end
