--robed cultist new

function think()
	conditions = getConditions(me())
	if not conditions ["Focused"] then
		doBasicAttack("Cultic Mantra", me())
	else
	target = pursue()
	if target == nil then
		target = retaliate()
	end
	target = targetAllyTarget()
	if target == nil then
		target = targetLowestStat("curHp")
	end
	moveWithinRangeAndAttack (1, "Sacrificial Blade", target)
end
think()