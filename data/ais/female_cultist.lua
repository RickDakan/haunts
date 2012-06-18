--female cultist
function think()
	target = pursue()
	if not target then
		target = retaliate()
	end
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
	if targetHasCondition(true, "Blindness") then
		moveAndAttack ("Envenomed Blade", target)
	end
	if targetHasCondition(false, "Blindness") then
		moveAndAttack ("Parasitic Gift", target)
	end
end
think()
		
		
		
		