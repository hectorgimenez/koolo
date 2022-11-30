package stat

type Stat int16

const (
	Strength Stat = iota
	Energy
	Dexterity
	Vitality
	StatPoints
	SkillPoints
	Life
	MaxLife
	Mana
	MaxMana
	Stamina
	MaxStamina
	Level
	Experience
	Gold
	StashGold
	EnhancedDefense
	EnhancedDamageMax
	EnhancedDamage
	AttackRating
	ChanceToBlock
	MinDamage
	MaxDamage
	TwoHandedMinDamage
	TwoHandedMaxDamage
	DamagePercent
	ManaRecovery
	ManaRecoveryBonus
	StaminaRecoveryBonus
	LastExp
	NextExp
	Defense
	DefenseVsMissiles
	DefenseVsHth
	NormalDamageReduction
	MagicDamageReduction
	DamageReduced
	MagicResist
	MaxMagicResist
	FireResist
	MaxFireResist
	LightningResist
	MaxLightningResist
	ColdResist
	MaxColdResist
	PoisonResist
	MaxPoisonResist
	DamageAura
	FireMinDamage
	FireMaxDamage
	LightningMinDamage
	LightningMaxDamage
	MagicMinDamage
	MagicMaxDamage
	ColdMinDamage
	ColdMaxDamage
	ColdLength
	PoisonMinDamage
	PoisonMaxDamage
	PoisonLength
	LifeSteal
	LifeStealMax
	ManaSteal
	ManaStealMax
	StaminaDrainMinDamage
	StaminaDrainMaxDamage
	StunLength
	VelocityPercent
	AttackRate
	OtherAnimRate
	Quantity
	Value
	Durability
	MaxDurability
	ReplenishLife
	MaxDurabilityPercent
	MaxLifePercent
	MaxManaPercent
	AttackerTakesDamage
	GoldFind
	MagicFind
	Knockback
	TimeDuration
	AddClassSkills
	Unused84
	AddExperience
	LifeAfterEachKill
	ReducePrices
	DoubleHerbDuration
	LightRadius
	LightColor
	Requirements
	LevelRequire
	IncreasedAttackSpeed
	LevelRequirePercent
	LastBlockFrame
	FasterRunWalk
	NonClassSkill
	State
	FasterHitRecovery
	PlayerCount
	PoisonOverrideLength
	FasterBlockRate
	BypassUndead
	BypassDemons
	FasterCastRate
	BypassBeasts
	SingleSkill
	SlainMonstersRestInPeace
	CurseResistance
	PoisonLengthReduced
	NormalDamage
	HitCausesMonsterToFlee
	HitBlindsTarget
	DamageTakenGoesToMana
	IgnoreTargetsDefense
	TargetDefense
	PreventMonsterHeal
	HalfFreezeDuration
	AttackRatingPercent
	MonsterDefensePerHit
	DemonDamagePercent
	UndeadDamagePercent
	DemonAttackRating
	UndeadAttackRating
	Throwable
	FireSkills
	AllSkills
	AttackerTakesLightDamage
	IronMaidenLevel
	LifeTapLevel
	ThornsPercent
	BoneArmor
	BoneArmorMax
	FreezesTarget
	OpenWounds
	CrushingBlow
	KickDamage
	ManaAfterKill
	HealAfterDemonKill
	ExtraBlood
	DeadlyStrike
	AbsorbFirePercent
	AbsorbFire
	AbsorbLightningPercent
	AbsorbLightning
	AbsorbMagicPercent
	AbsorbMagic
	AbsorbColdPercent
	AbsorbCold
	SlowsTarget
	Aura
	Indestructible
	CannotBeFrozen
	SlowerStaminaDrain
	Reanimate
	Pierce
	MagicArrow
	ExplosiveArrow
	ThrowMinDamage
	ThrowMaxDamage
	SkillHandofAthena
	SkillStaminaPercent
	SkillPassiveStaminaPercent
	SkillConcentration
	SkillEnchant
	SkillPierce
	SkillConviction
	SkillChillingArmor
	SkillFrenzy
	SkillDecrepify
	SkillArmorPercent
	Alignment
	Target0
	Target1
	GoldLost
	ConverisonLevel
	ConverisonMaxHP
	UnitDooverlay
	AttackVsMonType
	DamageVsMonType
	Fade
	ArmorOverridePercent
	Unused183
	Unused184
	Unused185
	Unused186
	Unused187
	AddSkillTab
	Unused189
	Unused190
	Unused191
	Unused192
	Unused193
	NumSockets
	SkillOnAttack
	SkillOnKill
	SkillOnDeath
	SkillOnHit
	SkillOnLevelUp
	Unused200
	SkillOnGetHit
	Unused202
	Unused203
	ItemChargedSkill
	Unused205
	Unused206
	Unused207
	Unused208
	Unused209
	Unused210
	Unused211
	Unused212
	Unused213
	DefensePerLevel
	ArmorPercentPerLevel
	LifePerLevel
	ManaPerLevel
	MaxDamagePerLevel
	MaxDamagePercentPerLevel
	StrengthPerLevel
	DexterityPerLevel
	EnergyPerLevel
	VitalityPerLevel
	AttackRatingPerLevel
	AttackRatingPercentPerLevel
	ColdDamageMaxPerLevel
	FireDamageMaxPerLevel
	LightningDamageMaxPerLevel
	PoisonDamageMaxPerLevel
	ResistColdPerLevel
	ResistFirePerLevel
	ResistLightningPerLevel
	ResistPoisonPerLevel
	AbsorbColdPerLevel
	AbsorbFirePerLevel
	AbsorbLightningPerLevel
	AbsorbPoisonPerLevel
	ThornsPerLevel
	ExtraGoldPerLevel
	MagicFindPerLevel
	RegenStaminaPerLevel
	StaminaPerLevel
	DamageDemonPerLevel
	DamageUndeadPerLevel
	AttackRatingDemonPerLevel
	AttackRatingUndeadPerLevel
	CrushingBlowPerLevel
	OpenWoundsPerLevel
	KickDamagePerLevel
	DeadlyStrikePerLevel
	FindGemsPerLevel
	ReplenishDurability
	ReplenishQuantity
	ExtraStack
	FindItem
	SlashDamage
	SlashDamagePercent
	CrushDamage
	CrushDamagePercent
	ThrustDamage
	ThrustDamagePercent
	AbsorbSlash
	AbsorbCrush
	AbsorbThrust
	AbsorbSlashPercent
	AbsorbCrushPercent
	AbsorbThrustPercent
	ArmorByTime
	ArmorPercentByTime
	LifeByTime
	ManaByTime
	MaxDamageByTime
	MaxDamagePercentByTime
	StrengthByTime
	DexterityByTime
	EnergyByTime
	VitalityByTime
	AttackRatingByTime
	AttackRatingPercentByTime
	ColdDamageMaxByTime
	FireDamageMaxByTime
	LightningDamageMaxByTime
	PoisonDamageMaxByTime
	ResistColdByTime
	ResistFireByTime
	ResistLightningByTime
	ResistPoisonByTime
	AbsorbColdByTime
	AbsorbFireByTime
	AbsorbLightningByTime
	AbsorbPoisonByTime
	FindGoldByTime
	MagicFindByTime
	RegenStaminaByTime
	StaminaByTime
	DamageDemonByTime
	DamageUndeadByTime
	AttackRatingDemonByTime
	AttackRatingUndeadByTime
	CrushingBlowByTime
	OpenWoundsByTime
	KickDamageByTime
	DeadlyStrikeByTime
	FindGemsByTime
	PierceCold
	PierceFire
	PierceLightning
	PiercePoison
	DamageVsMonster
	DamagePercentVsMonster
	AttackRatingVsMonster
	AttackRatingPercentVsMonster
	AcVsMonster
	AcPercentVsMonster
	FireLength
	BurningMin
	BurningMax
	ProgressiveDamage
	ProgressiveSteal
	ProgressiveOther
	ProgressiveFire
	ProgressiveCold
	ProgressiveLightning
	ExtraCharges
	ProgressiveAttackRating
	PoisonCount
	DamageFrameRate
	PierceIdx
	FireSkillDamage
	LightningSkillDamage
	ColdSkillDamage
	PoisonSkillDamage
	EnemyFireResist
	EnemyLightningResist
	EnemyColdResist
	EnemyPoisonResist
	PassiveCriticalStrike
	PassiveDodge
	PassiveAvoid
	PassiveEvade
	PassiveWarmth
	PassiveMasteryMeleeAttackRating
	PassiveMasteryMeleeDamage
	PassiveMasteryMeleeCritical
	PassiveMasteryThrowAttackRating
	PassiveMasteryThrowDamage
	PassiveMasteryThrowCritical
	PassiveWeaponBlock
	SummonResist
	ModifierListSkill
	ModifierListLevel
	LastSentHPPercent
	SourceUnitType
	SourceUnitID
	ShortParam1
	QuestItemDifficulty
	PassiveMagicMastery
	PassiveMagicPierce
)

func (s Stat) String() string {
	switch s {
	case Strength:
		return "strength"
	case Energy:
		return "energy"
	case Dexterity:
		return "dexterity"
	case Vitality:
		return "vitality"
	case StatPoints:
		return "statpoints"
	case SkillPoints:
		return "skillpoints"
	case Life:
		return "life"
	case MaxLife:
		return "maxlife"
	case Mana:
		return "mana"
	case MaxMana:
		return "maxmana"
	case Stamina:
		return "stamina"
	case MaxStamina:
		return "maxstamina"
	case Level:
		return "level"
	case Experience:
		return "experience"
	case Gold:
		return "gold"
	case StashGold:
		return "stashgold"
	case EnhancedDefense:
		return "enhanceddefense"
	case EnhancedDamageMax:
		return "enhanceddamagemax"
	case EnhancedDamage:
		return "enhanceddamage"
	case AttackRating:
		return "attackrating"
	case ChanceToBlock:
		return "chancetoblock"
	case MinDamage:
		return "mindamage"
	case MaxDamage:
		return "maxdamage"
	case TwoHandedMinDamage:
		return "twohandedmindamage"
	case TwoHandedMaxDamage:
		return "twohandedmaxdamage"
	case DamagePercent:
		return "damagepercent"
	case ManaRecovery:
		return "manarecovery"
	case ManaRecoveryBonus:
		return "manarecoverybonus"
	case StaminaRecoveryBonus:
		return "staminarecoverybonus"
	case LastExp:
		return "lastexp"
	case NextExp:
		return "nextexp"
	case Defense:
		return "defense"
	case DefenseVsMissiles:
		return "defensevsmissiles"
	case DefenseVsHth:
		return "defensevshth"
	case NormalDamageReduction:
		return "normaldamagereduction"
	case MagicDamageReduction:
		return "magicdamagereduction"
	case DamageReduced:
		return "damagereduced"
	case MagicResist:
		return "magicresist"
	case MaxMagicResist:
		return "maxmagicresist"
	case FireResist:
		return "fireresist"
	case MaxFireResist:
		return "maxfireresist"
	case LightningResist:
		return "lightningresist"
	case MaxLightningResist:
		return "maxlightningresist"
	case ColdResist:
		return "coldresist"
	case MaxColdResist:
		return "maxcoldresist"
	case PoisonResist:
		return "poisonresist"
	case MaxPoisonResist:
		return "maxpoisonresist"
	case DamageAura:
		return "damageaura"
	case FireMinDamage:
		return "firemindamage"
	case FireMaxDamage:
		return "firemaxdamage"
	case LightningMinDamage:
		return "lightningmindamage"
	case LightningMaxDamage:
		return "lightningmaxdamage"
	case MagicMinDamage:
		return "magicmindamage"
	case MagicMaxDamage:
		return "magicmaxdamage"
	case ColdMinDamage:
		return "coldmindamage"
	case ColdMaxDamage:
		return "coldmaxdamage"
	case ColdLength:
		return "coldlength"
	case PoisonMinDamage:
		return "poisonmindamage"
	case PoisonMaxDamage:
		return "poisonmaxdamage"
	case PoisonLength:
		return "poisonlength"
	case LifeSteal:
		return "lifesteal"
	case LifeStealMax:
		return "lifestealmax"
	case ManaSteal:
		return "manasteal"
	case ManaStealMax:
		return "manastealmax"
	case StaminaDrainMinDamage:
		return "staminadrainmindamage"
	case StaminaDrainMaxDamage:
		return "staminadrainmaxdamage"
	case StunLength:
		return "stunlength"
	case VelocityPercent:
		return "velocitypercent"
	case AttackRate:
		return "attackrate"
	case OtherAnimRate:
		return "otheranimrate"
	case Quantity:
		return "quantity"
	case Value:
		return "value"
	case Durability:
		return "durability"
	case MaxDurability:
		return "maxdurability"
	case ReplenishLife:
		return "replenishlife"
	case MaxDurabilityPercent:
		return "maxdurabilitypercent"
	case MaxLifePercent:
		return "maxlifepercent"
	case MaxManaPercent:
		return "maxmanapercent"
	case AttackerTakesDamage:
		return "attackertakesdamage"
	case GoldFind:
		return "goldfind"
	case MagicFind:
		return "magicfind"
	case Knockback:
		return "knockback"
	case TimeDuration:
		return "timeduration"
	case AddClassSkills:
		return "addclassskills"
	case Unused84:
		return "unused84"
	case AddExperience:
		return "addexperience"
	case LifeAfterEachKill:
		return "lifeaftereachkill"
	case ReducePrices:
		return "reduceprices"
	case DoubleHerbDuration:
		return "doubleherbduration"
	case LightRadius:
		return "lightradius"
	case LightColor:
		return "lightcolor"
	case Requirements:
		return "requirements"
	case LevelRequire:
		return "levelrequire"
	case IncreasedAttackSpeed:
		return "increasedattackspeed"
	case LevelRequirePercent:
		return "levelrequirepercent"
	case LastBlockFrame:
		return "lastblockframe"
	case FasterRunWalk:
		return "fasterrunwalk"
	case NonClassSkill:
		return "nonclassskill"
	case State:
		return "state"
	case FasterHitRecovery:
		return "fasterhitrecovery"
	case PlayerCount:
		return "playercount"
	case PoisonOverrideLength:
		return "poisonoverridelength"
	case FasterBlockRate:
		return "fasterblockrate"
	case BypassUndead:
		return "bypassundead"
	case BypassDemons:
		return "bypassdemons"
	case FasterCastRate:
		return "fastercastrate"
	case BypassBeasts:
		return "bypassbeasts"
	case SingleSkill:
		return "singleskill"
	case SlainMonstersRestInPeace:
		return "slainmonstersrestinpeace"
	case CurseResistance:
		return "curseresistance"
	case PoisonLengthReduced:
		return "poisonlengthreduced"
	case NormalDamage:
		return "normaldamage"
	case HitCausesMonsterToFlee:
		return "hitcausesmonstertoflee"
	case HitBlindsTarget:
		return "hitblindstarget"
	case DamageTakenGoesToMana:
		return "damagetakengoestomana"
	case IgnoreTargetsDefense:
		return "ignoretargetsdefense"
	case TargetDefense:
		return "targetdefense"
	case PreventMonsterHeal:
		return "preventmonsterheal"
	case HalfFreezeDuration:
		return "halffreezeduration"
	case AttackRatingPercent:
		return "attackratingpercent"
	case MonsterDefensePerHit:
		return "monsterdefenseperhit"
	case DemonDamagePercent:
		return "demondamagepercent"
	case UndeadDamagePercent:
		return "undeaddamagepercent"
	case DemonAttackRating:
		return "demonattackrating"
	case UndeadAttackRating:
		return "undeadattackrating"
	case Throwable:
		return "throwable"
	case FireSkills:
		return "fireskills"
	case AllSkills:
		return "allskills"
	case AttackerTakesLightDamage:
		return "attackertakeslightdamage"
	case IronMaidenLevel:
		return "ironmaidenlevel"
	case LifeTapLevel:
		return "lifetaplevel"
	case ThornsPercent:
		return "thornspercent"
	case BoneArmor:
		return "bonearmor"
	case BoneArmorMax:
		return "bonearmormax"
	case FreezesTarget:
		return "freezestarget"
	case OpenWounds:
		return "openwounds"
	case CrushingBlow:
		return "crushingblow"
	case KickDamage:
		return "kickdamage"
	case ManaAfterKill:
		return "manaafterkill"
	case HealAfterDemonKill:
		return "healafterdemonkill"
	case ExtraBlood:
		return "extrablood"
	case DeadlyStrike:
		return "deadlystrike"
	case AbsorbFirePercent:
		return "absorbfirepercent"
	case AbsorbFire:
		return "absorbfire"
	case AbsorbLightningPercent:
		return "absorblightningpercent"
	case AbsorbLightning:
		return "absorblightning"
	case AbsorbMagicPercent:
		return "absorbmagicpercent"
	case AbsorbMagic:
		return "absorbmagic"
	case AbsorbColdPercent:
		return "absorbcoldpercent"
	case AbsorbCold:
		return "absorbcold"
	case SlowsTarget:
		return "slowstarget"
	case Aura:
		return "aura"
	case Indestructible:
		return "indestructible"
	case CannotBeFrozen:
		return "cannotbefrozen"
	case SlowerStaminaDrain:
		return "slowerstaminadrain"
	case Reanimate:
		return "reanimate"
	case Pierce:
		return "pierce"
	case MagicArrow:
		return "magicarrow"
	case ExplosiveArrow:
		return "explosivearrow"
	case ThrowMinDamage:
		return "throwmindamage"
	case ThrowMaxDamage:
		return "throwmaxdamage"
	case SkillHandofAthena:
		return "skillhandofathena"
	case SkillStaminaPercent:
		return "skillstaminapercent"
	case SkillPassiveStaminaPercent:
		return "skillpassivestaminapercent"
	case SkillConcentration:
		return "skillconcentration"
	case SkillEnchant:
		return "skillenchant"
	case SkillPierce:
		return "skillpierce"
	case SkillConviction:
		return "skillconviction"
	case SkillChillingArmor:
		return "skillchillingarmor"
	case SkillFrenzy:
		return "skillfrenzy"
	case SkillDecrepify:
		return "skilldecrepify"
	case SkillArmorPercent:
		return "skillarmorpercent"
	case Alignment:
		return "alignment"
	case Target0:
		return "target0"
	case Target1:
		return "target1"
	case GoldLost:
		return "goldlost"
	case ConverisonLevel:
		return "converisonlevel"
	case ConverisonMaxHP:
		return "converisonmaxhp"
	case UnitDooverlay:
		return "unitdooverlay"
	case AttackVsMonType:
		return "attackvsmontype"
	case DamageVsMonType:
		return "damagevsmontype"
	case Fade:
		return "fade"
	case ArmorOverridePercent:
		return "armoroverridepercent"
	case Unused183:
		return "unused183"
	case Unused184:
		return "unused184"
	case Unused185:
		return "unused185"
	case Unused186:
		return "unused186"
	case Unused187:
		return "unused187"
	case AddSkillTab:
		return "addskilltab"
	case Unused189:
		return "unused189"
	case Unused190:
		return "unused190"
	case Unused191:
		return "unused191"
	case Unused192:
		return "unused192"
	case Unused193:
		return "unused193"
	case NumSockets:
		return "numsockets"
	case SkillOnAttack:
		return "skillonattack"
	case SkillOnKill:
		return "skillonkill"
	case SkillOnDeath:
		return "skillondeath"
	case SkillOnHit:
		return "skillonhit"
	case SkillOnLevelUp:
		return "skillonlevelup"
	case Unused200:
		return "unused200"
	case SkillOnGetHit:
		return "skillongethit"
	case Unused202:
		return "unused202"
	case Unused203:
		return "unused203"
	case ItemChargedSkill:
		return "itemchargedskill"
	case Unused205:
		return "unused205"
	case Unused206:
		return "unused206"
	case Unused207:
		return "unused207"
	case Unused208:
		return "unused208"
	case Unused209:
		return "unused209"
	case Unused210:
		return "unused210"
	case Unused211:
		return "unused211"
	case Unused212:
		return "unused212"
	case Unused213:
		return "unused213"
	case DefensePerLevel:
		return "defenseperlevel"
	case ArmorPercentPerLevel:
		return "armorpercentperlevel"
	case LifePerLevel:
		return "lifeperlevel"
	case ManaPerLevel:
		return "manaperlevel"
	case MaxDamagePerLevel:
		return "maxdamageperlevel"
	case MaxDamagePercentPerLevel:
		return "maxdamagepercentperlevel"
	case StrengthPerLevel:
		return "strengthperlevel"
	case DexterityPerLevel:
		return "dexterityperlevel"
	case EnergyPerLevel:
		return "energyperlevel"
	case VitalityPerLevel:
		return "vitalityperlevel"
	case AttackRatingPerLevel:
		return "attackratingperlevel"
	case AttackRatingPercentPerLevel:
		return "attackratingpercentperlevel"
	case ColdDamageMaxPerLevel:
		return "colddamagemaxperlevel"
	case FireDamageMaxPerLevel:
		return "firedamagemaxperlevel"
	case LightningDamageMaxPerLevel:
		return "lightningdamagemaxperlevel"
	case PoisonDamageMaxPerLevel:
		return "poisondamagemaxperlevel"
	case ResistColdPerLevel:
		return "resistcoldperlevel"
	case ResistFirePerLevel:
		return "resistfireperlevel"
	case ResistLightningPerLevel:
		return "resistlightningperlevel"
	case ResistPoisonPerLevel:
		return "resistpoisonperlevel"
	case AbsorbColdPerLevel:
		return "absorbcoldperlevel"
	case AbsorbFirePerLevel:
		return "absorbfireperlevel"
	case AbsorbLightningPerLevel:
		return "absorblightningperlevel"
	case AbsorbPoisonPerLevel:
		return "absorbpoisonperlevel"
	case ThornsPerLevel:
		return "thornsperlevel"
	case ExtraGoldPerLevel:
		return "extragoldperlevel"
	case MagicFindPerLevel:
		return "magicfindperlevel"
	case RegenStaminaPerLevel:
		return "regenstaminaperlevel"
	case StaminaPerLevel:
		return "staminaperlevel"
	case DamageDemonPerLevel:
		return "damagedemonperlevel"
	case DamageUndeadPerLevel:
		return "damageundeadperlevel"
	case AttackRatingDemonPerLevel:
		return "attackratingdemonperlevel"
	case AttackRatingUndeadPerLevel:
		return "attackratingundeadperlevel"
	case CrushingBlowPerLevel:
		return "crushingblowperlevel"
	case OpenWoundsPerLevel:
		return "openwoundsperlevel"
	case KickDamagePerLevel:
		return "kickdamageperlevel"
	case DeadlyStrikePerLevel:
		return "deadlystrikeperlevel"
	case FindGemsPerLevel:
		return "findgemsperlevel"
	case ReplenishDurability:
		return "replenishdurability"
	case ReplenishQuantity:
		return "replenishquantity"
	case ExtraStack:
		return "extrastack"
	case FindItem:
		return "finditem"
	case SlashDamage:
		return "slashdamage"
	case SlashDamagePercent:
		return "slashdamagepercent"
	case CrushDamage:
		return "crushdamage"
	case CrushDamagePercent:
		return "crushdamagepercent"
	case ThrustDamage:
		return "thrustdamage"
	case ThrustDamagePercent:
		return "thrustdamagepercent"
	case AbsorbSlash:
		return "absorbslash"
	case AbsorbCrush:
		return "absorbcrush"
	case AbsorbThrust:
		return "absorbthrust"
	case AbsorbSlashPercent:
		return "absorbslashpercent"
	case AbsorbCrushPercent:
		return "absorbcrushpercent"
	case AbsorbThrustPercent:
		return "absorbthrustpercent"
	case ArmorByTime:
		return "armorbytime"
	case ArmorPercentByTime:
		return "armorpercentbytime"
	case LifeByTime:
		return "lifebytime"
	case ManaByTime:
		return "manabytime"
	case MaxDamageByTime:
		return "maxdamagebytime"
	case MaxDamagePercentByTime:
		return "maxdamagepercentbytime"
	case StrengthByTime:
		return "strengthbytime"
	case DexterityByTime:
		return "dexteritybytime"
	case EnergyByTime:
		return "energybytime"
	case VitalityByTime:
		return "vitalitybytime"
	case AttackRatingByTime:
		return "attackratingbytime"
	case AttackRatingPercentByTime:
		return "attackratingpercentbytime"
	case ColdDamageMaxByTime:
		return "colddamagemaxbytime"
	case FireDamageMaxByTime:
		return "firedamagemaxbytime"
	case LightningDamageMaxByTime:
		return "lightningdamagemaxbytime"
	case PoisonDamageMaxByTime:
		return "poisondamagemaxbytime"
	case ResistColdByTime:
		return "resistcoldbytime"
	case ResistFireByTime:
		return "resistfirebytime"
	case ResistLightningByTime:
		return "resistlightningbytime"
	case ResistPoisonByTime:
		return "resistpoisonbytime"
	case AbsorbColdByTime:
		return "absorbcoldbytime"
	case AbsorbFireByTime:
		return "absorbfirebytime"
	case AbsorbLightningByTime:
		return "absorblightningbytime"
	case AbsorbPoisonByTime:
		return "absorbpoisonbytime"
	case FindGoldByTime:
		return "findgoldbytime"
	case MagicFindByTime:
		return "magicfindbytime"
	case RegenStaminaByTime:
		return "regenstaminabytime"
	case StaminaByTime:
		return "staminabytime"
	case DamageDemonByTime:
		return "damagedemonbytime"
	case DamageUndeadByTime:
		return "damageundeadbytime"
	case AttackRatingDemonByTime:
		return "attackratingdemonbytime"
	case AttackRatingUndeadByTime:
		return "attackratingundeadbytime"
	case CrushingBlowByTime:
		return "crushingblowbytime"
	case OpenWoundsByTime:
		return "openwoundsbytime"
	case KickDamageByTime:
		return "kickdamagebytime"
	case DeadlyStrikeByTime:
		return "deadlystrikebytime"
	case FindGemsByTime:
		return "findgemsbytime"
	case PierceCold:
		return "piercecold"
	case PierceFire:
		return "piercefire"
	case PierceLightning:
		return "piercelightning"
	case PiercePoison:
		return "piercepoison"
	case DamageVsMonster:
		return "damagevsmonster"
	case DamagePercentVsMonster:
		return "damagepercentvsmonster"
	case AttackRatingVsMonster:
		return "attackratingvsmonster"
	case AttackRatingPercentVsMonster:
		return "attackratingpercentvsmonster"
	case AcVsMonster:
		return "acvsmonster"
	case AcPercentVsMonster:
		return "acpercentvsmonster"
	case FireLength:
		return "firelength"
	case BurningMin:
		return "burningmin"
	case BurningMax:
		return "burningmax"
	case ProgressiveDamage:
		return "progressivedamage"
	case ProgressiveSteal:
		return "progressivesteal"
	case ProgressiveOther:
		return "progressiveother"
	case ProgressiveFire:
		return "progressivefire"
	case ProgressiveCold:
		return "progressivecold"
	case ProgressiveLightning:
		return "progressivelightning"
	case ExtraCharges:
		return "extracharges"
	case ProgressiveAttackRating:
		return "progressiveattackrating"
	case PoisonCount:
		return "poisoncount"
	case DamageFrameRate:
		return "damageframerate"
	case PierceIdx:
		return "pierceidx"
	case FireSkillDamage:
		return "fireskilldamage"
	case LightningSkillDamage:
		return "lightningskilldamage"
	case ColdSkillDamage:
		return "coldskilldamage"
	case PoisonSkillDamage:
		return "poisonskilldamage"
	case EnemyFireResist:
		return "enemyfireresist"
	case EnemyLightningResist:
		return "enemylightningresist"
	case EnemyColdResist:
		return "enemycoldresist"
	case EnemyPoisonResist:
		return "enemypoisonresist"
	case PassiveCriticalStrike:
		return "passivecriticalstrike"
	case PassiveDodge:
		return "passivedodge"
	case PassiveAvoid:
		return "passiveavoid"
	case PassiveEvade:
		return "passiveevade"
	case PassiveWarmth:
		return "passivewarmth"
	case PassiveMasteryMeleeAttackRating:
		return "passivemasterymeleeattackrating"
	case PassiveMasteryMeleeDamage:
		return "passivemasterymeleedamage"
	case PassiveMasteryMeleeCritical:
		return "passivemasterymeleecritical"
	case PassiveMasteryThrowAttackRating:
		return "passivemasterythrowattackrating"
	case PassiveMasteryThrowDamage:
		return "passivemasterythrowdamage"
	case PassiveMasteryThrowCritical:
		return "passivemasterythrowcritical"
	case PassiveWeaponBlock:
		return "passiveweaponblock"
	case SummonResist:
		return "summonresist"
	case ModifierListSkill:
		return "modifierlistskill"
	case ModifierListLevel:
		return "modifierlistlevel"
	case LastSentHPPercent:
		return "lastsenthppercent"
	case SourceUnitType:
		return "sourceunittype"
	case SourceUnitID:
		return "sourceunitid"
	case ShortParam1:
		return "shortparam1"
	case QuestItemDifficulty:
		return "questitemdifficulty"
	case PassiveMagicMastery:
		return "passivemagicmastery"
	case PassiveMagicPierce:
		return "passivemagicpierce"
	}

	return ""
}
