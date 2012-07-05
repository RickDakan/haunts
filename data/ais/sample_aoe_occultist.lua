function Think()
  -- This is just to check that there is someone around that we can see
  denizens = Utils.NearestNEntities (1, "denizen")
  if table.getn(denizens) == 0 then
    return
  end

  -- gz, or 'ground zero', is where we're going to center our aoe
  gz = Utils.BestAoeAttackPos("Abjuration", Me.ApCur - Me.Actions["Abjuration"].Ap, "enemies only")

  -- find all positions from which we could center our aoe on gz
  dsts = Utils.AllPathablePoints(Me.Pos, gz, 1, Me.Actions["Abjuration"].Range)

  -- move to any one of the closest positions in dsts
  Do.Move(dsts, 1000)

  -- if we're still out of range then we'll just have to try again next turn
  if Utils.RangedDistBetweenPositions(Me.Pos, gz) > Me.Actions["Abjuration"].Range then
    return
  else
    Do.AoeAttack("Abjuration", gz)
  end

  -- More attacks if possible
  Think()
end
