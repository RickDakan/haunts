Entity objects
--------------

These are always contructed dynamically whenever they are accessed, so the information on them is always up-to-date.  Entity ais can access a global Entity object called Me that represents the entity controlled by the ai.

For an entity called ent the following fields are available:

    ent.Name
    -- The name of the entity, as displayed to the user
    
    ent.Side.Denizen
    ent.Side.Intruder
    ent.Side.Npc
    ent.Side.Object
    -- Each entity has these booleans that identify which side it is on, or if it is an Npc or
    -- Object.  All entities will identify as belonging to exactly one of these groups.

    ent.Conditions
    -- A table mapping conditions names to true boolean value

    ent.GearOptions
    -- For intruders this is a mapping from names of available gear items to the path of the large
    -- icon representing that gear.
    -- For non-intruders this is nil.

    ent.Gear
    -- For intruders this is the name of the gear the intruder is using, or the empty string if it is
    -- not using gear.
    -- For non-intruders this is nil.
    
    ent.Actions
    -- A table mapping action name to an action object.
    
    ent.Pos.X
    ent.Pos.Y
    -- Current coordinates

    ent.Corpus
    ent.Ego
    ent.HpCur
    ent.HpMax
    ent.ApCur
    ent.ApMax
    -- These are stats as affected by any conditions on the entity, they are not necesssarily the
    -- same as the entity's base stats.
