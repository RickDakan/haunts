--eidolon

function Think()
	target = pursue()
	if target == nil then
		target = retaliate()
	end
	if target == nil then
		target = targetAllyTarget()
	end
	if target == nil then
		target = targetLowestStat("HpCur")
	end
	if target == nil then
		target = nearest()
	end	
	if target == nil then
		return
	end
	if target.Conditions["Horrified"] then
		moveWithinRangeAndAttack (1, "Feast", target)
	else
		moveWithinRangeAndAttack (1, "Cosmic Infection", target)
	end
end
