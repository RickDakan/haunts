--female cultist
function think()
	target = pursue()
	if target == nil then
		target = retaliate()
	end
	print("target, ", target)
	if not target then
		target = targetAllyAttacker()
	end
	if not target then
		target = targetAllyTarget()
	end
	if not target then
		target = targetLowestState(hpCur)
	end
	if not target then
  	target = nearest()
	end
	
	if target == nil then
		return
	end
	print("targeting ", target)
	if getConditions(target)["Blindness"] then
		moveWithinRangeAndAttack (1, "Envenomed Blade", target)
	else
		moveWithinRangeAndAttack (1, "Parasitic Gift", target)
	end
end
think()
		
		
		
