--new robed cultist

-- am I buffed?

function Think()
		if not Me.Conditions ["Focused"] then 
			Do.BasicAttack ("Cultic Mantra", Me)
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
		target = targetLowestStat(HpCur)
	end
	moveAndAttack("Sacrificial Blade", target)
	end
end
