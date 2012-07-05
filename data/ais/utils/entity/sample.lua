function moveToAndMeleeAttack(attack,target)
  print(Me)
  print(target)
  ps = Utils.AllPathablePoints(Me.Pos, target.Pos, 1, 1)
  res = Do.Move(ps, 1000)
  if res == nil then
    return nil
  end
  return Do.BasicAttack(attack, target)
end
