-- aoe_test
function aoePlaceAndAttack(attack, spec)
	me_stats = getEntityStats(me())
	attack_stats = getAoeAttackStats(me(), attack)
	gz = bestAoeAttackPos (attack, me_stats.apCur - attack_stats.ap, spec)
	dsts = allPathablePoints(pos(me()), gz, 1, attack_stats.range)
	doMove(dsts, 1000)
	if rangedDistBetweenPositions (pos(me()), gz) > attack_stats.range then
		return
	else
		doAoeAttack(attack, gz)
	end
end

function think()
	aoePlaceAndAttack("Abjuration", "enemies only")
	think()
end
think()
