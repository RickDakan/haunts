
--The Female Cultist AI would begin here, assuming all the above is in the Utility folder

function Think()
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
		target = targetLowestState(HpCur)
	end

	if targetHasCondition(true, "Blindness") then
		moveAndAttack ("Envenomed Blade", target)
	end
	if targetHasCondition(false, "Blindness") then
		moveAndAttack ("Parasitic Gift", target)
	end
end
