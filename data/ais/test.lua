-- Find a random room, make a path, find the next room in the path, find the
-- door to that room, go to the door, open it, step through it, repeat.

function think()
  stats = getEntityStats(me())
  if stats.apCur == 0 then
    return
  end
  unexplored = nearbyUnexploredRoom()
  print(unexplored)
  if not unexplored then
    -- We've explored the whole house
    return
  end
  path = roomPath(roomContaining(me()), unexplored)
  if not path then
    -- Can't get there for some odd reason
    return
  end
  doors = allDoorsBetween(roomContaining(me()), path[1])
  if table.getn(doors) == 0 then
    -- No doors, shouldn't have made this path
    print("There should be doors between", roomContaining(me()), path[1])
    return
  end
  ps = doorPositions(doors[1])
  res = doMove(ps, 1000)
  if res then
    -- We successfully moved towards the door, so we should rethink
    think()
    return
  end
  -- If it failed it was probably because we were already on the door, so just
  -- assume that for now and try to open it
  print("doorIsOpen(", doorIsOpen(doors[1]), ")")
  if not doorIsOpen(doors[1]) then
    doDoorToggle(doors[1])
  end
  -- Either way the door should be open now and we can try to walk through it
  ps = roomPositions(path[1])
  doMove(ps, 1000)
  think()
end





think()
