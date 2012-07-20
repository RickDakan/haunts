Do Functions
------------

All functions available in this table will attempt to have the current entity execute an action in the game.  The return value for all of these will be nil iff the action was invalid for some reason.


###Do.__Move__(_dsts_, _max_ap_)
_dsts_: Array of acceptable destination positions.  
_max_ap_: Maximum ap to spend doing this move.

The current entity will attempt a Move action from its current location to the nearest position in dsts.  If it cannot reach any position in dsts in less than _max_ap_ Ap it will move as far as it can.  If the move action was valid this function will return the number of Ap spend doing the move.

Example:  

    intruders = Utils.NearestNEntities(1, "intruder")
    if table.getn(intruders) > 0 then
        dsts = Utils.AllPathablePoints(Me.Pos, intruders[1].Pos, 1, 1)
        -- dsts not contains all of the points that this entity can walk to that are directly
        -- adjacent to the nearest intruder.

        res = Do.Move(dsts, 5)
        -- Now do the move, but don't spend more than 5 Ap

        if res == nil then
            -- If the move failed it was probably because there was no path to the intruder
        else
            -- Do some sort of attack here.
        end
    end

------

###Do.__BasicAttack__(_attack_name_, _target_)  
_attack_name_: Name of the attack to use.  
_target_: Entity to target with this attack.

The current entity will attempt to use a Basic Attack with the given name targeting the specified entity.  This will fail if the current entity does not have an action with the specified name, if the specified action is not a Basic Attack, if target is not a valid target, or if the current entity does not have enough ap to use the action.  If the attack was valid the return value will be a table with the following values:

    Hit: True iff the attack hit its target.

Example:

    intruders = Utils.NearestNEntities(1, "intruder")
    if table.getn(intruders) > 0 then
        if Utils.RangedDistBetweenEntities(Me, intruders[1]) < Me.Actions["Kick"].Range then
            res = Do.BasicAttack("Kick", intruders[1])
            if not res then
                -- Maybe we forgot to check that we had enough Ap, maybe we didn't have LoS
            else
                if res.Hit then
                    -- Follow up with another attack here
                else
                    -- Run away
                end
            end
        end
    end

------

###Do.__AoeAttack__(_attack_name_, _center_)  
_attack_name_: Name of the attack to use.  
_center_: Position at which to center the AoE.

The current entity will attempt to use an Aoe Attack with the given name centered around the specified position.  This will fail if the current entity does not have an action with the specified name, if the specified action is not an Aoe Attack, if target is not a valid target, or if the current entity does not have enough ap to use the action.  If the attack was valid the return value will be a true boolean value.

Example:

    gz = Utils.BestAoeAttackPos("Grenade", 0, "enemies only")
    if Do.AoeAttack("Grenade", gz) then
        -- Celebrate your fallen enemies
    else
        -- Consider that you maybe should check that gz isn't nil
    end

------

###Do.__DoorToggle__(_door_)  
_door_: The door to open/close.  

The current entity will attempt to use its Interact action to open or close the specified door.  This will fail if the entity does not have an Interact action, if it does not have sufficient Ap, if the door is out of range or if the specified object is not a door.  If the action is successful this function will return a boolean value indicating if the door is currently open.

Example:

    my_room = Utils.RoomContaining(Me)
    path = Utils.RoomPath(my_room, Utils.NearbyUnexploredRoom())
    -- path is a path of rooms from this room to some other room we haven't been to yet.

    doors = Utils.AllDoorsBetween(my_room, path[1])
    -- doors[1] should be the first door we have to go through to get to the unexplored room.

    if not Utils.DoorIsOpen(doors[1]) then
        res = Do.DoorToggle(doors[1])
        if res == nil then
            -- Maybe there wasn't an unexplored room, or maybe we weren't next to the door.
        end
    end

