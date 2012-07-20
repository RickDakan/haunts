--master AI
--spec
--"allies ok", "minions ok", "enemies only"

--no enemies in sight?
	-- look at shades and wisps,
	-- whichever I see fewer of, summon one.

--needs a target?
function aoePlaceAndAttack(attack, spec)
	gz = Utils.BestAoeAttackPos (attack, Me.ApCur - Me.Actions[attack].Ap, spec)
	dsts = Utils.AllPathablePoints(Me.Pos, gz, 1, Me.Actions[attack].Range)
	Do.Move(dsts, 1000)
	if Utils.RangedDistBetweenPositions (Me.Pos, gz) > Me.Actions[attack].Range then
		return
	else
		Do.AoeAttack(attack, gz)
	end
end

function Think()
	aoePlaceAndAttack("Poltergeist Blast", "enemies only")
	Think()
end

