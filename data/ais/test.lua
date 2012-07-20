-- Find a random room, make a path, find the next room in the path, find the
-- door to that room, go to the door, open it, step through it, repeat.

function Think()
  if Me.ApCur == 0 then
    return
  end
  unexplored = Utils.NearbyUnexploredRoom()
  print(unexplored)
  if not unexplored then
    -- We've explored the whole house
    return
  end
  path = Utils.RoomPath(Utils.RoomContaining(Me), unexplored)
  if not path then
    -- Can't get there for some odd reason
    return
  end
  doors = Utils.AllDoorsBetween(Utils.RoomContaining(Me), path[1])
  if table.getn(doors) == 0 then
    -- No doors, shouldn't have made this path
    print("There should be doors between", Utils.RoomContaining(Me), path[1])
    return
  end
  ps = Utils.DoorPositions(doors[1])
  res = Do.Move(ps, 1000)
  if res then
    -- We successfully moved towards the door, so we should rethink
    Think()
    return
  end
  -- If it failed it was probably because we were already on the door, so just
  -- assume that for now and try to open it
  print("Utils.DoorIsOpen(", Utils.DoorIsOpen(doors[1]), ")")
  if not Utils.DoorIsOpen(doors[1]) then
    Do.DoorToggle(doors[1])
  end
  -- Either way the door should be open now and we can try to walk through it
  ps = Utils.RoomPositions(path[1])
  Do.Move(ps, 1000)
  Think()
end
