package npc

type ID int

const (
	Skeleton                 ID = 0
	Returned                 ID = 1
	BoneWarrior              ID = 2
	BurningDead              ID = 3
	Horror                   ID = 4
	Zombie                   ID = 5
	HungryDead               ID = 6
	Ghoul                    ID = 7
	DrownedCarcass           ID = 8
	PlagueBearer             ID = 9
	Afflicted                ID = 10
	Tainted                  ID = 11
	Misshapen                ID = 12
	Disfigured               ID = 13
	Damned                   ID = 14
	FoulCrow                 ID = 15
	BloodHawk                ID = 16
	BlackRaptor              ID = 17
	CloudStalker             ID = 18
	Fallen                   ID = 19
	Carver                   ID = 20
	Devilkin                 ID = 21
	DarkOne                  ID = 22
	WarpedFallen             ID = 23
	Brute                    ID = 24
	Yeti                     ID = 25
	Crusher                  ID = 26
	WailingBeast             ID = 27
	GargantuanBeast          ID = 28
	SandRaider               ID = 29
	Marauder                 ID = 30
	Invader                  ID = 31
	Infidel                  ID = 32
	Assailant                ID = 33
	Gorgon                   ID = 34 // Unused
	Gorgon2                  ID = 35 // Unused
	Gorgon3                  ID = 36 // Unused
	Gorgon4                  ID = 37 // Unused
	Ghost                    ID = 38
	Wraith                   ID = 39
	Specter                  ID = 40
	Apparition               ID = 41
	DarkShape                ID = 42
	DarkHunter               ID = 43
	VileHunter               ID = 44
	DarkStalker              ID = 45
	BlackRogue               ID = 46
	FleshHunter              ID = 47
	DuneBeast                ID = 48
	RockDweller              ID = 49
	JungleHunter             ID = 50
	DoomApe                  ID = 51
	TempleGuard              ID = 52
	MoonClan                 ID = 53
	NightClan                ID = 54
	BloodClan                ID = 55
	HellClan                 ID = 56
	DeathClan                ID = 57
	FallenShaman             ID = 58
	CarverShaman             ID = 59
	DevilkinShaman           ID = 60
	DarkShaman               ID = 61
	WarpedShaman             ID = 62
	QuillRat                 ID = 63
	SpikeFiend               ID = 64
	ThornBeast               ID = 65
	RazorSpine               ID = 66
	JungleUrchin             ID = 67
	SandMaggot               ID = 68
	RockWorm                 ID = 69
	Devourer                 ID = 70
	GiantLamprey             ID = 71
	WorldKiller              ID = 72
	TombViper                ID = 73
	ClawViper                ID = 74
	Salamander               ID = 75
	PitViper                 ID = 76
	SerpentMagus             ID = 77
	SandLeaper               ID = 78
	CaveLeaper               ID = 79
	TombCreeper              ID = 80
	TreeLurker               ID = 81
	RazorPitDemon            ID = 82
	Huntress                 ID = 83
	SaberCat                 ID = 84
	NightTiger               ID = 85
	HellCat                  ID = 86
	Itchies                  ID = 87
	BlackLocusts             ID = 88
	PlagueBugs               ID = 89
	HellSwarm                ID = 90
	DungSoldier              ID = 91
	SandWarrior              ID = 92
	Scarab                   ID = 93
	SteelWeevil              ID = 94
	AlbinoRoach              ID = 95
	DriedCorpse              ID = 96
	Decayed                  ID = 97
	Embalmed                 ID = 98
	PreservedDead            ID = 99
	Cadaver                  ID = 100
	HollowOne                ID = 101
	Guardian                 ID = 102
	Unraveler                ID = 103
	HoradrimAncient          ID = 104
	BaalSubjectMummy         ID = 105
	ChaosHorde               ID = 106 // Unused
	ChaosHorde2              ID = 107 // Unused
	ChaosHorde3              ID = 108 // Unused
	ChaosHorde4              ID = 109 // Unused
	CarrionBird              ID = 110
	UndeadScavenger          ID = 111
	HellBuzzard              ID = 112
	WingedNightmare          ID = 113
	Sucker                   ID = 114
	Feeder                   ID = 115
	BloodHook                ID = 116
	BloodWing                ID = 117
	Gloam                    ID = 118
	SwampGhost               ID = 119
	BurningSoul              ID = 120
	BlackSoul                ID = 121
	Arach                    ID = 122
	SandFisher               ID = 123
	PoisonSpinner            ID = 124
	FlameSpider              ID = 125
	SpiderMagus              ID = 126
	ThornedHulk              ID = 127
	BrambleHulk              ID = 128
	Thrasher                 ID = 129
	Spikefist                ID = 130
	GhoulLord                ID = 131
	NightLord                ID = 132
	DarkLord                 ID = 133
	BloodLord                ID = 134
	Banished                 ID = 135
	DesertWing               ID = 136
	Fiend                    ID = 137
	Gloombat                 ID = 138
	BloodDiver               ID = 139
	DarkFamiliar             ID = 140
	RatMan                   ID = 141
	Fetish                   ID = 142
	Flayer                   ID = 143
	SoulKiller               ID = 144
	StygianDoll              ID = 145
	DeckardCain              ID = 146
	Gheed                    ID = 147
	Akara                    ID = 148
	Chicken                  ID = 149 // Dummy
	Kashya                   ID = 150
	Rat                      ID = 151 // Dummy
	Rogue                    ID = 152 // Dummy
	HellMeteor               ID = 153 // Dummy
	Charsi                   ID = 154
	Warriv                   ID = 155
	Andariel                 ID = 156
	Bird                     ID = 157 // Dummy
	Bird2                    ID = 158 // Dummy
	Bat                      ID = 159 // Dummy
	DarkRanger               ID = 160
	VileArcher               ID = 161
	DarkArcher               ID = 162
	BlackArcher              ID = 163
	FleshArcher              ID = 164
	DarkSpearwoman           ID = 165
	VileLancer               ID = 166
	DarkLancer               ID = 167
	BlackLancer              ID = 168
	FleshLancer              ID = 169
	SkeletonArcher           ID = 170
	ReturnedArcher           ID = 171
	BoneArcher               ID = 172
	BurningDeadArcher        ID = 173
	HorrorArcher             ID = 174
	Warriv2                  ID = 175
	Atma                     ID = 176
	Drognan                  ID = 177
	Fara                     ID = 178
	Cow                      ID = 179 // Dummy
	SandMaggotYoung          ID = 180
	RockWormYoung            ID = 181
	DevourerYoung            ID = 182
	GiantLampreyYoung        ID = 183
	WorldKillerYoung         ID = 184
	Camel                    ID = 185 // Dummy
	Blunderbore              ID = 186
	Gorbelly                 ID = 187
	Mauler                   ID = 188
	Urdar                    ID = 189
	SandMaggotEgg            ID = 190
	RockWormEgg              ID = 191
	DevourerEgg              ID = 192
	GiantLampreyEgg          ID = 193
	WorldKillerEgg           ID = 194
	Act2Male                 ID = 195 // Dummy
	Act2Female               ID = 196 // Dummy
	Act2Child                ID = 197 // Dummy
	Greiz                    ID = 198
	Elzix                    ID = 199
	Geglash                  ID = 200
	Jerhyn                   ID = 201
	Lysander                 ID = 202
	Act2Guard                ID = 203 // Dummy
	Act2Vendor               ID = 204 // Dummy
	Act2Vendor2              ID = 205 // Dummy
	FoulCrowNest             ID = 206
	BloodHawkNest            ID = 207
	BlackVultureNest         ID = 208
	CloudStalkerNest         ID = 209
	Meshif                   ID = 210
	Duriel                   ID = 211
	UndeadRatMan             ID = 212 //Unused???
	UndeadFetish             ID = 213 //Unused???
	UndeadFlayer             ID = 214 //Unused???
	UndeadSoulKiller         ID = 215 //Unused???
	UndeadStygianDoll        ID = 216 //Unused???
	DarkGuard                ID = 217 // Unused
	DarkGuard2               ID = 218 // Unused
	DarkGuard3               ID = 219 // Unused
	DarkGuard4               ID = 220 // Unused
	DarkGuard5               ID = 221 // Unused
	BloodMage                ID = 222 // Unused
	BloodMage2               ID = 223 // Unused
	BloodMage3               ID = 224 // Unused
	BloodMage4               ID = 225 // Unused
	BloodMage5               ID = 226 // Unused
	Maggot                   ID = 227
	MummyGenerator           ID = 228 // TEST: Sarcophagus
	Radament                 ID = 229
	FireBeast                ID = 230 // Unused
	IceGlobe                 ID = 231 // Unused
	LightningBeast           ID = 232 // Unused
	PoisonOrb                ID = 233 // Unused
	FlyingScimitar           ID = 234
	Zakarumite               ID = 235
	Faithful                 ID = 236
	Zealot                   ID = 237
	Sexton                   ID = 238
	Cantor                   ID = 239
	Heirophant               ID = 240
	Heirophant2              ID = 241
	Mephisto                 ID = 242
	Diablo                   ID = 243
	DeckardCain2             ID = 244
	DeckardCain3             ID = 245
	DeckardCain4             ID = 246
	SwampDweller             ID = 247
	BogCreature              ID = 248
	SlimePrince              ID = 249
	Summoner                 ID = 250
	Tyrael                   ID = 251
	Asheara                  ID = 252
	Hratli                   ID = 253
	Alkor                    ID = 254
	Ormus                    ID = 255
	Izual                    ID = 256
	Halbu                    ID = 257
	WaterWatcherLimb         ID = 258
	RiverStalkerLimb         ID = 259
	StygianWatcherLimb       ID = 260
	WaterWatcherHead         ID = 261
	RiverStalkerHead         ID = 262
	StygianWatcherHead       ID = 263
	Meshif2                  ID = 264
	DeckardCain5             ID = 265
	Navi                     ID = 266
	BloodRaven               ID = 267
	Bug                      ID = 268 // Dummy
	Scorpion                 ID = 269 // Dummy
	RogueScout               ID = 270
	Rogue2                   ID = 271 // Dummy
	Rogue3                   ID = 272 // Dummy
	GargoyleTrap             ID = 273
	ReturnedMage             ID = 274
	BoneMage                 ID = 275
	BurningDeadMage          ID = 276
	HorrorMage               ID = 277
	RatManShaman             ID = 278
	FetishShaman             ID = 279
	FlayerShaman             ID = 280
	SoulKillerShaman         ID = 281
	StygianDollShaman        ID = 282
	Larva                    ID = 283
	SandMaggotQueen          ID = 284
	RockWormQueen            ID = 285
	DevourerQueen            ID = 286
	GiantLampreyQueen        ID = 287
	WorldKillerQueen         ID = 288
	ClayGolem                ID = 289
	BloodGolem               ID = 290
	IronGolem                ID = 291
	FireGolem                ID = 292
	Familiar                 ID = 293 // Dummy
	Act3Male                 ID = 294 // Dummy
	NightMarauder            ID = 295
	Act3Female               ID = 296 // Dummy
	Natalya                  ID = 297
	FleshSpawner             ID = 298
	StygianHag               ID = 299
	Grotesque                ID = 300
	FleshBeast               ID = 301
	StygianDog               ID = 302
	GrotesqueWyrm            ID = 303
	Groper                   ID = 304
	Strangler                ID = 305
	StormCaster              ID = 306
	Corpulent                ID = 307
	CorpseSpitter            ID = 308
	MawFiend                 ID = 309
	DoomKnight               ID = 310
	AbyssKnight              ID = 311
	OblivionKnight           ID = 312
	QuillBear                ID = 313
	SpikeGiant               ID = 314
	ThornBrute               ID = 315
	RazorBeast               ID = 316
	GiantUrchin              ID = 317
	Snake                    ID = 318 // Dummy
	Parrot                   ID = 319 // Dummy
	Fish                     ID = 320 // Dummy
	EvilHole                 ID = 321 // Dummy
	EvilHole2                ID = 322 // Dummy
	EvilHole3                ID = 323 // Dummy
	EvilHole4                ID = 324 // Dummy
	EvilHole5                ID = 325 // Dummy
	FireboltTrap             ID = 326 // A trap
	HorzMissileTrap          ID = 327 // A trap
	VertMissileTrap          ID = 328 // A trap
	PoisonCloudTrap          ID = 329 // A trap
	LightningTrap            ID = 330 // A trap
	Kaelan                   ID = 331 // Act2Guard2
	InvisoSpawner            ID = 332 // Dummy
	DiabloClone              ID = 333 // Unused???
	SuckerNest               ID = 334
	FeederNest               ID = 335
	BloodHookNest            ID = 336
	BloodWingNest            ID = 337
	Guard                    ID = 338 // Act2Hire
	MiniSpider               ID = 339 // Dummy
	BonePrison               ID = 340 // Unused???
	BonePrison2              ID = 341 // Unused???
	BonePrison3              ID = 342 // Unused???
	BonePrison4              ID = 343 // Unused???
	BoneWall                 ID = 344 // Dummy
	CouncilMember            ID = 345
	CouncilMember2           ID = 346
	CouncilMember3           ID = 347
	Turret                   ID = 348
	Turret2                  ID = 349
	Turret3                  ID = 350
	Hydra                    ID = 351
	Hydra2                   ID = 352
	Hydra3                   ID = 353
	MeleeTrap                ID = 354 // A trap
	SevenTombs               ID = 355 // Dummy
	Decoy                    ID = 356
	Valkyrie                 ID = 357
	Act2Guard3               ID = 358 // Unused???
	IronWolf                 ID = 359 // Act3Hire
	Balrog                   ID = 360
	PitLord                  ID = 361
	VenomLord                ID = 362
	NecroSkeleton            ID = 363
	NecroMage                ID = 364
	Griswold                 ID = 365
	CompellingOrbNpc         ID = 366
	Tyrael2                  ID = 367
	DarkWanderer             ID = 368
	NovaTrap                 ID = 369
	SpiritMummy              ID = 370 // Dummy
	LightningSpire           ID = 371
	FireTower                ID = 372
	Slinger                  ID = 373
	SpearCat                 ID = 374
	NightSlinger             ID = 375
	HellSlinger              ID = 376
	Act2Guard4               ID = 377 // Dummy
	Act2Guard5               ID = 378 // Dummy
	ReturnedMage2            ID = 379
	BoneMage2                ID = 380
	BaalColdMage             ID = 381
	HorrorMage2              ID = 382
	ReturnedMage3            ID = 383
	BoneMage3                ID = 384
	BurningDeadMage2         ID = 385
	HorrorMage3              ID = 386
	ReturnedMage4            ID = 387
	BoneMage4                ID = 388
	BurningDeadMage3         ID = 389
	HorrorMage4              ID = 390
	HellBovine               ID = 391
	Window                   ID = 392 // Dummy
	Window2                  ID = 393 // Dummy
	SpearCat2                ID = 394
	NightSlinger2            ID = 395
	RatMan2                  ID = 396
	Fetish2                  ID = 397
	Flayer2                  ID = 398
	SoulKiller2              ID = 399
	StygianDoll2             ID = 400
	MephistoSpirit           ID = 401 // Dummy
	TheSmith                 ID = 402
	TrappedSoul              ID = 403
	TrappedSoul2             ID = 404
	Jamella                  ID = 405
	Izual2                   ID = 406
	RatMan3                  ID = 407
	Malachai                 ID = 408
	Hephasto                 ID = 409 // The Feature Creep ?!?
	WakeOfDestruction        ID = 410 // Expansion (Are We missing something here?  D2BS has a 410 that we DONT have)
	ChargedBoltSentry        ID = 411
	LightningSentry          ID = 412
	BladeCreeper             ID = 413
	InvisiblePet             ID = 414 // Dummy ? Unused ?
	InfernoSentry            ID = 415
	DeathSentry              ID = 416
	ShadowWarrior            ID = 417
	ShadowMaster             ID = 418
	DruidHawk                ID = 419
	DruidSpiritWolf          ID = 420
	DruidFenris              ID = 421
	SpiritOfBarbs            ID = 422
	HeartOfWolverine         ID = 423
	OakSage                  ID = 424
	DruidPlaguePoppy         ID = 425
	DruidCycleOfLife         ID = 426
	VineCreature             ID = 427
	DruidBear                ID = 428
	Eagle                    ID = 429
	Wolf                     ID = 430
	Bear                     ID = 431
	BarricadeDoor            ID = 432
	BarricadeDoor2           ID = 433
	PrisonDoor               ID = 434
	BarricadeTower           ID = 435
	RotWalker                ID = 436
	ReanimatedHorde          ID = 437
	ProwlingDead             ID = 438
	UnholyCorpse             ID = 439
	DefiledWarrior           ID = 440
	SiegeBeast               ID = 441
	CrushBiest               ID = 442
	BloodBringer             ID = 443
	GoreBearer               ID = 444
	DeamonSteed              ID = 445
	SnowYeti                 ID = 446
	SnowYeti2                ID = 447
	SnowYeti3                ID = 448
	SnowYeti4                ID = 449
	WolfRider                ID = 450
	WolfRider2               ID = 451
	WolfRider3               ID = 452
	MinionExp                ID = 453 // ??
	SlayerExp                ID = 454 // ??
	IceBoar                  ID = 455
	FireBoar                 ID = 456
	HellSpawn                ID = 457
	IceSpawn                 ID = 458
	GreaterHellSpawn         ID = 459
	GreaterIceSpawn          ID = 460
	FanaticMinion            ID = 461
	BerserkSlayer            ID = 462
	ConsumedIceBoar          ID = 463
	ConsumedFireBoar         ID = 464
	FrenziedHellSpawn        ID = 465
	FrenziedIceSpawn         ID = 466
	InsaneHellSpawn          ID = 467
	InsaneIceSpawn           ID = 468
	SuccubusExp              ID = 469 // just Succubus ??
	VileTemptress            ID = 470
	StygianHarlot            ID = 471
	HellTemptress            ID = 472
	BloodTemptress           ID = 473
	Dominus                  ID = 474
	VileWitch                ID = 475
	StygianFury              ID = 476
	BloodWitch               ID = 477
	HellWitch                ID = 478
	OverSeer                 ID = 479
	Lasher                   ID = 480
	OverLord                 ID = 481
	BloodBoss                ID = 482
	HellWhip                 ID = 483
	MinionSpawner            ID = 484
	MinionSlayerSpawner      ID = 485
	MinionBoarSpawner        ID = 486
	MinionBoarSpawner2       ID = 487
	MinionSpawnSpawner       ID = 488
	MinionBoarSpawner3       ID = 489
	MinionBoarSpawner4       ID = 490
	MinionSpawnSpawner2      ID = 491
	Imp                      ID = 492
	Imp2                     ID = 493
	Imp3                     ID = 494
	Imp4                     ID = 495
	Imp5                     ID = 496
	CatapultS                ID = 497
	CatapultE                ID = 498
	CatapultSiege            ID = 499
	CatapultW                ID = 500
	FrozenHorror             ID = 501
	FrozenHorror2            ID = 502
	FrozenHorror3            ID = 503
	FrozenHorror4            ID = 504
	FrozenHorror5            ID = 505
	BloodLord2               ID = 506
	BloodLord3               ID = 507
	BloodLord4               ID = 508
	BloodLord5               ID = 509
	BloodLord6               ID = 510
	Larzuk                   ID = 511
	Drehya                   ID = 512
	Malah                    ID = 513
	NihlathakTown            ID = 514
	QualKehk                 ID = 515
	CatapultSpotterS         ID = 516
	CatapultSpotterE         ID = 517
	CatapultSpotterSiegeName    = 518
	CatapultSpotterW         ID = 519
	DeckardCain6             ID = 520
	Tyrael3                  ID = 521
	Act5Combatant            ID = 522
	Act5Combatant2           ID = 523
	BarricadeWallRight       ID = 524
	BarricadeWallLeft        ID = 525
	Nihlathak                ID = 526
	Drehya2                  ID = 527
	EvilHut                  ID = 528
	DeathMauler              ID = 529
	DeathMauler2             ID = 530
	DeathMauler3             ID = 531
	DeathMauler4             ID = 532
	DeathMauler5             ID = 533
	POW                      ID = 534
	Act5Townguard            ID = 535
	Act5Townguard2           ID = 536
	AncientStatue            ID = 537
	AncientStatueNpc2        ID = 538
	AncientStatueNpc3        ID = 539
	AncientBarbarian         ID = 540
	AncientBarbarian2        ID = 541
	AncientBarbarian3        ID = 542
	BaalThrone               ID = 543
	BaalCrab                 ID = 544
	BaalTaunt                ID = 545
	PutridDefiler            ID = 546
	PutridDefiler2           ID = 547
	PutridDefiler3           ID = 548
	PutridDefiler4           ID = 549
	PutridDefiler5           ID = 550
	PainWorm                 ID = 551
	PainWorm2                ID = 552
	PainWorm3                ID = 553
	PainWorm4                ID = 554
	PainWorm5                ID = 555
	Bunny                    ID = 556
	CouncilMemberBall        ID = 557
	VenomLord2               ID = 558
	BaalCrabToStairs         ID = 559
	Act5Hireling1Hand        ID = 560
	Act5Hireling2Hand        ID = 561
	BaalTentacle             ID = 562
	BaalTentacle2            ID = 563
	BaalTentacle3            ID = 564
	BaalTentacle4            ID = 565
	BaalTentacle5            ID = 566
	InjuredBarbarian         ID = 567
	InjuredBarbarian2        ID = 568
	InjuredBarbarian3        ID = 569
	BaalCrabClone            ID = 570
	BaalsMinion              ID = 571
	BaalsMinion2             ID = 572
	BaalsMinion3             ID = 573
	WorldstoneEffect         ID = 574 // D2BS stops here???
	BurningDeadArcher2       ID = 575
	BoneArcher2              ID = 576
	BurningDeadArcher3       ID = 577
	ReturnedArcher2          ID = 578
	HorrorArcher2            ID = 579
	Afflicted2               ID = 580
	Tainted2                 ID = 581
	Misshapen2               ID = 582
	Disfigured2              ID = 583
	Damned2                  ID = 584
	MoonClan2                ID = 585
	NightClan2               ID = 586
	HellClan2                ID = 587
	BloodClan2               ID = 588
	DeathClan2               ID = 589
	FoulCrow2                ID = 590
	BloodHawk2               ID = 591
	BlackRaptor2             ID = 592
	CloudStalker2            ID = 593
	ClawViper2               ID = 594
	PitViper2                ID = 595
	Salamander2              ID = 596
	TombViper2               ID = 597
	SerpentMagus2            ID = 598
	Marauder2                ID = 599
	Infidel2                 ID = 600
	SandRaider2              ID = 601
	Invader2                 ID = 602
	Assailant2               ID = 603
	DeathMauler6             ID = 604
	QuillRat2                ID = 605
	SpikeFiend2              ID = 606
	RazorSpine2              ID = 607
	CarrionBird2             ID = 608
	ThornedHulk2             ID = 609
	Slinger2                 ID = 610
	Slinger3                 ID = 611
	Slinger4                 ID = 612
	VileArcher2              ID = 613
	DarkArcher2              ID = 614
	VileLancer2              ID = 615
	DarkLancer2              ID = 616
	BlackLancer2             ID = 617
	Blunderbore2             ID = 618
	Mauler2                  ID = 619
	ReturnedMage5            ID = 620
	BurningDeadMage4         ID = 621
	ReturnedMage6            ID = 622
	HorrorMage5              ID = 623
	BoneMage5                ID = 624
	HorrorMage6              ID = 625
	HorrorMage7              ID = 626
	Huntress2                ID = 627
	SaberCat2                ID = 628
	CaveLeaper2              ID = 629
	TombCreeper2             ID = 630
	Ghost2                   ID = 631
	Wraith2                  ID = 632
	Specter2                 ID = 633
	SuccubusExp2             ID = 634
	HellTemptress2           ID = 635
	Dominus2                 ID = 636
	HellWitch2               ID = 637
	VileWitch2               ID = 638
	Gloam2                   ID = 639
	BlackSoul2               ID = 640
	BurningSoul2             ID = 641
	Carver2                  ID = 642
	Devilkin2                ID = 643
	DarkOne2                 ID = 644
	CarverShaman2            ID = 645
	DevilkinShaman2          ID = 646
	DarkShaman2              ID = 647
	BoneWarrior2             ID = 648
	Returned2                ID = 649
	Gloombat2                ID = 650
	Fiend2                   ID = 651
	BloodLord7               ID = 652
	BloodLord8               ID = 653
	Scarab2                  ID = 654
	SteelWeevil2             ID = 655
	Flayer3                  ID = 656
	StygianDoll3             ID = 657
	SoulKiller3              ID = 658
	Flayer4                  ID = 659
	StygianDoll4             ID = 660
	SoulKiller4              ID = 661
	FlayerShaman2            ID = 662
	StygianDollShaman2       ID = 663
	SoulKillerShaman2        ID = 664
	TempleGuard2             ID = 665
	TempleGuard3             ID = 666
	Guardian2                ID = 667
	Unraveler2               ID = 668
	HoradrimAncient2         ID = 669
	HoradrimAncient3         ID = 670
	Zealot2                  ID = 671
	Zealot3                  ID = 672
	Heirophant3              ID = 673
	Heirophant4              ID = 674
	Grotesque2               ID = 675
	FleshSpawner2            ID = 676
	GrotesqueWyrm2           ID = 677
	FleshBeast2              ID = 678
	WorldKiller2             ID = 679
	WorldKillerYoung2        ID = 680
	WorldKillerEgg2          ID = 681
	SlayerExp2               ID = 682
	HellSpawn2               ID = 683
	GreaterHellSpawn2        ID = 684
	Arach2                   ID = 685
	Balrog2                  ID = 686
	PitLord2                 ID = 687
	Imp6                     ID = 688
	Imp7                     ID = 689
	UndeadStygianDoll2       ID = 690
	UndeadSoulKiller2        ID = 691
	Strangler2               ID = 692
	StormCaster2             ID = 693
	MawFiend2                ID = 694
	BloodLord9               ID = 695
	GhoulLord2               ID = 696
	DarkLord2                ID = 697
	UnholyCorpse2            ID = 698
	DoomKnight2              ID = 699
	DoomKnight3              ID = 700
	OblivionKnight2          ID = 701
	OblivionKnight3          ID = 702
	Cadaver2                 ID = 703
	UberMephisto             ID = 704
	UberDiablo               ID = 705
	UberIzual                ID = 706
	Lilith                   ID = 707
	UberDuriel               ID = 708
	UberBaal                 ID = 709
	EvilHut2                 ID = 710
	DemonHole                ID = 711 // Dummy
	PitLord3                 ID = 712
	OblivionKnight4          ID = 713
	Imp8                     ID = 714
	HellSwarm2               ID = 715
	WorldKiller3             ID = 716
	Arach3                   ID = 717
	SteelWeevil3             ID = 718
	HellTemptress3           ID = 719
	VileWitch3               ID = 720
	FleshHunter2             ID = 721
	DarkArcher3              ID = 722
	BlackLancer3             ID = 723
	HellWhip2                ID = 724
	Returned3                ID = 725
	HorrorArcher3            ID = 726
	BurningDeadMage5         ID = 727
	HorrorMage8              ID = 728
	BoneMage6                ID = 729
	HorrorMage9              ID = 730
	DarkLord3                ID = 731
	Specter3                 ID = 732
	BurningSoul3             ID = 733
)
