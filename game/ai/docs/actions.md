Action Objects
--------------

Action objects contain useful stats about an action.  Each type of action exports a different set of stats.

Movement

    act.Type
    -- "Move"


Interact

    act.Type
    -- "Interact"

    act.Ap
    act.Range
    -- Typical stats


Basic Attacks

    act.Type
    -- "Basic Attack"

    act.Name
    -- The name of this specific action.

    act.Ap
    act.Damage
    act.Strength
    act.Range
    -- Typical stats

    act.Ammo
    -- For actions with unlimited ammo this will be a large number (1000)


Aoe Attacks

    act.Type
    -- "Aoe Attack"

    act.Name
    -- The name of this specific action.

    act.Ap
    act.Damage
    act.Strength
    act.Range
    -- Typical stats

    act.Ammo
    -- For actions with unlimited ammo this will be a large number (1000)

    act.Diameter
    -- Diameter of the area affected by the aoe


Summons

    act.Type
    -- "Summon"

    act.Name
    -- The name of this specific action.

    act.Ap
    act.Range
    -- Typical stats

    act.Ammo
    -- For actions with unlimited ammo this will be a large number (1000)

    act.Entity
    -- Name of the entity that this ability summons

    act.Los
    -- Whether or not this ability requires that its user has LoS to the its target, or if it is
    -- sufficient for a teammate to have LoS.
