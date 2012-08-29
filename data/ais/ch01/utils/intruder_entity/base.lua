
function SupportAllies(buf, cond)
  -- First make sure I'm always buffed
  if not Me.Conditions[cond] then
    return Do.BasicAttack(buf, Me)
  end

  -- Now make sure my teammates are buffed if they are in trouble
  allies = Utils.NearestNEntities(3, "intruder")
  for _, ally in pairs(allies) do
    if ally.HpCur <= 5 and not ally.Conditions[cond] then
      return Do.BasicAttack(buf, ally)
    end
  end

  return false
end


-- This will attempt to move Me one space such that it stays in the same
-- room, but is no longer standing in a doorway.  If it is not currently
-- in a doorway right now this function does nothing.
function TryToClearDoorway()
  -- Check if we're on a doorway space, if we are then move to a nearby
  -- space that is in this room but is not a doorway space.
  all_door_ps = {}
  rcm = Utils.RoomContaining(Me)
  ado = Utils.AllDoorsOn(rcm)
  for i, door in pairs(ado) do
    ps = Utils.DoorPositions(door)
    for _, p in pairs(ps) do
      all_door_ps[p.X .. "," .. p.Y] = true
    end
  end

  -- If we're not next to a door then, whatever, don't bother moving.
  if not all_door_ps[Me.Pos.X .. "," .. Me.Pos.Y] then
    return
  end

  -- Make rps a map from position to boolean, true if the position is
  -- in the room
  rps = {}
  for _, p in pairs(Utils.RoomPositions(Utils.RoomContaining(Me))) do
    rps[p.X .. "," .. p.Y] = true
  end

  -- Loop over all nearby positions, if there is one we can move to that
  -- is in this room but is not a door position, then move to it.
  nearby = Utils.AllPathablePoints(Me.Pos, Me.Pos, 1, 1)
  for _, p in pairs(nearby) do
    if rps[p.X .. "," .. p.Y] and not all_door_ps[p.X .. "," .. p.Y] then
      Do.Move({p}, 10)
    end
  end
end

function CrushEnemies(debuf, cond, melee, ranged)
  print("SCRIPT: CrushEnemies")
  -- Don't crush enemies unless we've tried to clear the doorway first
  TryToClearDoorway()
  print("SCRIPT: NOW crushing")

  enemies = Utils.NearestNEntities(10, "denizen")
  if table.getn(enemies) == 0 then
    return false
  end
  max_dist = Me.Actions[ranged].Range
  lowest_hp = 10000
  lowest_ent = nil
  for i, enemy in pairs(enemies) do
    dist = Utils.RangedDistBetweenEntities(Me, enemy)
    if dist and dist <= max_dist and enemy.HpCur < lowest_hp then
      lowest_hp = enemy.HpCur
      lowest_ent = enemy
    end
  end
  if lowest_ent == nil then
    return false
  end
  target = lowest_ent
  dist = Utils.RangedDistBetweenEntities(Me, target)
  if cond and not target.Conditions[cond] and dist <= Me.Actions[debuf].Range then
    return Do.BasicAttack(debuf, target)
  end
  attack = ranged
  if dist == 1 then
    attack = melee
  end
  return Do.BasicAttack(attack, target)
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

function Leader()
  intruders = Utils.NearestNEntities(3, "intruder")
  for _, ent in pairs(intruders) do
    if ent.Name == Me.Master.Leader then
      return ent
    end
  end
  return nil
end

-- assumes that Me is not the leader, returns the intruder that is neither
-- Me or the leader.
function OtherGuy()
  intruders = Utils.NearestNEntities(3, "intruder")
  for _, ent in pairs(intruders) do
    if ent.Name ~= Me.Name and ent.Name ~= Me.Master.Leader then
      return ent
    end
  end
  return nil
end

function LeadOrFollow()
  if Me.Master.Leader == Me.Name then
    -- Don't go off leading somewhere if there is something nearby that needs
    -- to be dealt with
    ent = Utils.NearestNEntities(1, "denizen")[1]
    if ent then
      dist = Utils.RangedDistBetweenEntities(Me, ent)
      if dist < 5 then
        return false
      end
    end

    objects = Utils.NearestNEntities(3, "object")
    if table.getn(objects) == 0 then
      return false
    end
    -- Should only ever be one at a time, so just take the first one
    object = objects[1]
    return HeadTowards(object.Pos)
  else
    leader = Leader()
    if leader then
      if Follow(leader, 2) then
        return false
      end
    end
    other = OtherGuy()
    if other then
      Follow(other, 1)
      return false
    end
    return false
  end
end

function getCenter(thing)
  x = thing.Pos.X + thing.Dims.Dx / 2
  y = thing.Pos.Y + thing.Dims.Dy / 2
  return {X=x, Y=y}
end

function distance(a, b)
  v1 = a.X - b.X
  if v1 < 0 then
    v1 = 0-v1
  end
  v2 = a.Y - b.Y
  if v2 < 0 then
    v2 = 0-v2
  end
  return v1 + v2
end

function FindUnexploredRoomNear(target)
  unexplored = Utils.NearbyUnexploredRooms()
  if table.getn(unexplored) == 0 then
    return nil  -- No more rooms to explore
  end

  -- find the room whose center has the shortest manhattan distance to our
  -- target.
  dist = 0-1
  target_room = nil
  for _, room in pairs(unexplored) do
    c = getCenter(room)
    d = distance(target, getCenter(room))
    if dist == 0-1 or d < dist then
      dist = d
      target_room = room
    end
  end

  return target_room
end

-- pos.X and pos.Y are the position in question
-- region.Pos.X, region.Pos.Y, region.Dims.Dx and region.Dims.Dy define
-- the region that we will check for pos in.
function posIsInRegion(pos, region)
  if pos.X < region.Pos.X then
    return false
  end
  if pos.Y < region.Pos.Y then
    return false
  end
  if pos.X >= region.Pos.X + region.Dims.Dx then
    return false
  end
  if pos.Y >= region.Pos.Y + region.Dims.Dy then
    return false
  end
  return true
end

function HeadTowards(target)
  current = Utils.RoomContaining(Me)
  if posIsInRegion(target, current) then
    ps = Utils.AllPathablePoints(Me.Pos, target, 1, 1)
    valid, pos = Do.Move(ps, 1000)
    return valid and pos
  end

  target_room = FindUnexploredRoomNear(target)
  if target_room == nil then
    return false
  end
  path = Utils.RoomPath(current, target_room)
  if table.getn(path) == 0 then
    return false  -- No room path to the unexplored room - shouldn't happen
  end
  target_room = path[1]

  doors = Utils.AllDoorsBetween(current, target_room)
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

    -- Only open it if we have at least half of our max ap and our peeps
    -- are nearby
    if Me.ApCur < Me.ApMax / 2 then
      return false
    end
    intruders = Utils.NearestNEntities(3, "intruder")
    if table.getn(intruders) > 0 then
      if Utils.RangedDistBetweenEntities(Me, intruders[table.getn(intruders)]) > 5 then
        return false
      end
    end

    if not Do.DoorToggle(doors[1]) then
      return false
    end
  end

  -- Now that we know the door is open, step into the next room.
  ps = Utils.RoomPositions(target_room)
  complete, other = Do.Move(ps, 1000)
  return complete
end

function Follow(leader, leash)
  ps = Utils.AllPathablePoints(Me.Pos, leader.Pos, 1, leash)
  Do.Move(ps, 1000)
  dist =  Utils.RangedDistBetweenEntities(Me, leader)
  return dist and dist <= leash
end


