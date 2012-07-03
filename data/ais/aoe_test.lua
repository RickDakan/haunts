-- aoe_test
function aoePlaceAndAttack(attack, spec)
	gz = bestAoeAttackPos (attack, Me.apCur - Me.actions[attack].ap, spec)
	dsts = allPathablePoints(Me.Pos, gz, 1, Me.actions[attack].range)
	doMove(dsts, 1000)
	if rangedDistBetweenPositions (Me.pos, gz) > Me.actions[attack].range then
		return
	else
		doAoeAttack(attack, gz)
	end
end

function Think()
	aoePlaceAndAttack("Abjuration", "enemies only")
	Think()
end
