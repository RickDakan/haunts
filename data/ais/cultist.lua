--new robed cultist

-- am I buffed?

function think()
	conditions = getConditions (me())
		if not conditions ["Focused"] then 
			doBasicAttack ("Cultic Mantra", me())
		end
	end
	target = pursue()
	if not target then
		target = retaliate()
	end
	if not target then
		target = targetHasCondition(true, "Agony")
	end
	if not target then
		target = targetHasCondition(true, "Blindness")
	end
	if not target then
		target = targetLowestStat(hpCur)
	end
	moveAndAttack("Sacrificial Blade", target)
	end
end
think()

		
		
		
