-- First check for an untouched relic in the room you're in
-- Then look for a nearby unexplored room and head in that direction

function CheckForRelic()
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
    return  -- No doors to the next room, also shouldn't happen
  end

  -- If the door is closed then go to it and open it.
  if not Utils.DoorIsOpen(doors[1]) then
    ps = Utils.DoorPositions(doors[1])
    res = Do.Move(ps, 1000)
    Do.DoorToggle(doors[1])
  end

  -- Now that we know the door is open, step into the next room.
  ps = Utils.RoomPositions(target)
  res = Do.Move(ps, 1000)

  return true
end

function Think()
  CheckForRelic()
  
  while GoToUnexploredRoom() do
  end
end
