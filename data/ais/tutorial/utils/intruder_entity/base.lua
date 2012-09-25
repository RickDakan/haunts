
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
  print("SCRIPT: TryToClearDoorway - ", Me.Name)
  -- Check if we're on a doorway space, if we are then move to a nearby
  -- space that is in this room but is not a doorway space.
  all_door_ps = {}
  rcm = Utils.RoomContaining(Me)
  rcl = Utils.RoomContaining(Leader())
  if not Utils.RoomsAreEqual(rcm, rcl) then
    -- Don't try to clear a doorway if we're not in the room with the leader
    -- yet, because we might be trying to clear a doorway in a room we were
    -- trying to leave.
    return
  end
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
  print("SCRIPT: In The room, must clear - ", Me.Name)

  -- Make rps a map from position to boolean, true if the position is
  -- in the room
  rps = {}
  for _, p in pairs(Utils.RoomPositions(Utils.RoomContaining(Me))) do
    rps[p.X .. "," .. p.Y] = true
  end

  -- Loop over all nearby positions, if there is one we can move to that
  -- is in this room but is not a door position, then move to it.
  nearby = Utils.AllPathablePoints(Me.Pos, Me.Pos, 1, 3)
  closest = 100000
  closest_pos = nil
  for _, p in pairs(nearby) do
    if rps[p.X .. "," .. p.Y] and not all_door_ps[p.X .. "," .. p.Y] then
      dist = Utils.RangedDistBetweenPositions(Me.Pos, p)
      if dist and dist < closest then
        closest = dist
        closest_pos = p
      end
    end
  end
  Do.Move({closest_pos}, 10)
end

function CrushEnemies(debuf, cond, melee, ranged, aoe)
  print("SCRIPT: CrushEnemies - ", Me.Name)

  enemies = Utils.NearestNEntities(10, "denizen")
  if table.getn(enemies) == 0 then
    return false
  end

  nearest = enemies[1]
  rce = Utils.RoomContaining(nearest)  
  rcm = Utils.RoomContaining(Me)
  if Utils.RoomsAreEqual(rce, rcm) then
    -- Don't crush enemies unless we've tried to clear the doorway first
    TryToClearDoorway()
  end

  print("SCRIPT: Aoe:", aoe)
  if aoe and Me.Actions[aoe].Ap > Me.ApCur then
    aoe_dist = Me.Actions[aoe].Range
    print("SCRIPT: ", aoe_dist)
    pos, ents = Utils.BestAoeAttackPos(aoe, 1, "enemies only")
    print("SCRIPT: Ps:", pos.X, pos.Y)
    -- We can hit more than one entity so we'll go ahead and use our aoe
    if table.getn(ents) > 1 then
      ps = Utils.AllPathablePoints(Me.Pos, pos, 1, aoe_dist)
      Do.Move(ps, 1000)
      Do.AoeAttack(aoe, pos)
    end
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
  if not dist then
    return false
  end
  if debuf and cond and not target.Conditions[cond] and dist <= Me.Actions[debuf].Range then
    return Do.BasicAttack(debuf, target)
  end
  attack = ranged
  if dist == 1 then
    attack = melee
  end
  return Do.BasicAttack(attack, target)
end

function ObjectPos()
  print("SCRIPT: ObjectPos1")
  objects = Utils.NearestNEntities(3, "object")
  if table.getn(objects) == 0 then
    return nil
  end
  print("SCRIPT: ObjectPos2")

  for _,obj in pairs(objects) do
    if obj.State == "ready" then
      return obj.Pos
    end
  end
  print("SCRIPT: ObjectPos3")
  return nil
end

function WaypointPos()
  print("SCRIPT: Waypoint")
  for _, wp in pairs(Utils.Waypoints()) do
  print("SCRIPT: Waypoin2")
    return wp.Pos
  end
  print("SCRIPT: Waypoint3")
  return nil
end

function RelicPos()
  return WaypointPos()
  -- pos = ObjectPos()
  -- if not pos then
  --   pos = WaypointPos()
  -- end
  -- return pos
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
  print("SCRIPT: me/leader, ", Me.Name, "/", Me.Master.Leader)
  if Me.Master.Leader == Me.Name then
    -- Don't go off leading somewhere if there is something nearby that needs
    -- to be dealt with
    ent = Utils.NearestNEntities(1, "denizen")[1]
    if ent then
      dist = Utils.RangedDistBetweenEntities(Me, ent)
      if dist and dist < 5 then
        return false
      end
    end

    pos = false
    if pos then
      print("SCRIPT: RelicPos = ", pos.X, pos.Y)
      return HeadTowards(pos)
    end
    print("SCRIPT: NO RelicPos")
    return false
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

function Follow(leader, leash)
  ps = Utils.AllPathablePoints(Me.Pos, leader.Pos, 1, leash)
  Do.Move(ps, 1000)
  dist =  Utils.RangedDistBetweenEntities(Me, leader)
  return dist and dist <= leash
end


