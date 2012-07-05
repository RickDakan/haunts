--robed cultist new

function Think()
	if not Me.Conditions["Focused"] then
		Actions.BasicAttack("Cultic Mantra", Me)
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
