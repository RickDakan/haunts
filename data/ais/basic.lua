
function Think()
  intruders = NearestNEntities(3, "intruder")

  range = Me().Actions["Abjuration"].Ap
  if intruders[1] then
    movement = Me().ApCur - Me().Actions["Abjuration"].Ap
    if movement < 0 then
      movement = 0
    end
    target = BestAoeAttackPos("Abjuration", movement, "enemies only")
    if not (target.x == 0 and target.y == 0) then
      ps = AllPathablePoints(Me().Pos, target, 1, range)
      res = DoMove(ps, 1000)
      res = DoAoeAttack("Abjuration", target)
      if res == nil then
        return nil
      end
    else
      ps = AllPathablePoints(Me().Pos, intruders[1].Pos, range, range)
      DoMove(ps, 1000)
      return
    end
  else
    return
  end
end
