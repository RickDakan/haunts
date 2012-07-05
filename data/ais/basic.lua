
function Think()
  intruders = Utils.NearestNEntities(3, "intruder")

  range = Me.Actions["Abjuration"].Ap
  if intruders[1] then
    movement = Me.ApCur - Me.Actions["Abjuration"].Ap
    if movement < 0 then
      movement = 0
    end
    target = Utils.BestAoeAttackPos("Abjuration", movement, "enemies only")
    if not (target.X == 0 and target.Y == 0) then
      ps = Utils.AllPathablePoints(Me.Pos, target, 1, range)
      res = Actions.Move(ps, 1000)
      res = Actions.AoeAttack("Abjuration", target)
      if res == nil then
        return nil
      end
    else
      ps = Utils.AllPathablePoints(Me.Pos, intruders[1].Pos, range, range)
      Actions.Move(ps, 1000)
      return
    end
  else
    return
  end
end
