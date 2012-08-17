-- First check for an untouched relic in the room you're in
-- Then look for a nearby unexplored room and head in that direction

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

function Think()
  while CheckForRelic() do
  end
  
  while GoToUnexploredRoom() do
  end
end
