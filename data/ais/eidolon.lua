--eidolon

function think()
	target = pursue()
	if not target then
		target = retaliate()
	end
	if not target then
		target = targetHasCondition(true, "Horrified")
	end
	if not target then
		target = targetHighestStat(ego)
	end
	if targetHasCondition (false, "Horrified") then
		moveAndAttack ("Cosmic Infection", target)
	end
	if targetHasCondition (true, "Horrified") then
		moveAndAttack ("Feast", target)
	end
end
think()

