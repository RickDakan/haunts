while true do
  intruders = nearestNEntities(3, "intruder")
  if intruders[1] then
    mypos = pos(me())
    hispos = pos(intruders[1])
    ps = allPathablePoints(mypos, hispos, 1, 1)
    if not ps[1] then
      break
    end
    doMove(ps, 1000)
    res = doBasicAttack("Chill Touch", intruders[1])
    if res == nil then
      break
    end
  end
end
