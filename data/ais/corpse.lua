--intruders = nearestNEntities(3, "intruder")
--mypos = pos(me())

--intruder = intruders[1]
function think()
	target = pursue()
	if target == nil then
		target = retaliate()
	end
	if target == nil then
		target = targetAllyTarget()
	end
	if target == nil then
		target = targetLowestStat("hpCur")
	end
	if target == nil then
		target = nearest()
	end	
	if target == nil then
		return
	end
	if getConditions(target)["Poison"] then
		moveWithinRangeAndAttack (1, "Pummel", target)
	else
		moveWithinRangeAndAttack (1, "Diseased Kiss", target)
	end
end
think()