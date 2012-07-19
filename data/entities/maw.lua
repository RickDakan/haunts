--maw
-- 
function Think()
	if Me.ApCur < Me.Actions ["Swipe"].Ap then
		return
	end
	target = nearest()
	if Utils.Exists(target) then
		if Utils.RangedDistBetweenEntities (Me, target) == 1 then
			Do.BasicAttack("Swipe", target)
			Think()
		end
	end
end

