intruders = Utils.NearestNEntities(3, "intruder")
mypos = Me.Pos

intruder = intruders[1]
if intruder then
  if rangedDistBetween(Me, intruder) == 1 then
    -- If we're right next to someone then we will either try to disease
    -- them, if they aren't already diseased, otherwise we will just
    -- attack them as much as possible
    attack = ""
    if intruder.Conditions["Diseased Kiss"] then
      attack = "Pummel"
    else
      attack = "Diseased Kiss"
    end
    while Utils.Exists(intruder) do
      res = Actions.BasicAttack(attack, intruder)
      if res.hit then
        attack = "Pummel"
      end
    end
  else
    -- Path to a nearby intruder.  We'll pick the closest one if possible,
    -- but if we can't get to that one for some reason then we'll pick a
    -- different one.
    for _, intruder in pairs(intruders) do
      ps = Utils.AllPathablePoints(mypos, intruder.Pos, 1, 1)
      if ps[1] then
        Actions.Move(ps, 1000)
        break
      end
    end
  end
else
  -- If there are no intruders in sight we walk around randomly
  ps = Utils.AllPathablePoints(mypos, mypos, 1, 3)
  r = randN(table.getn(ps))
  target = ps[randN(table.getn(ps))]
  a = {}
  a[1] = target
  res = Actions.Move(a, 1000)
end
