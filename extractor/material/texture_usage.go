package material

type TextureUsage uint32

const (
	Albedo                                     TextureUsage = 0xac652e43
	AlbedoArray                                TextureUsage = 0x1db1694e
	AlbedoBlend                                TextureUsage = 0xf672f340
	AlbedoBlendTex                             TextureUsage = 0xdc88db85
	AlbedoColor                                TextureUsage = 0x80205cca
	AlbedoEmissive                             TextureUsage = 0xe67ac0c7
	AlbedoIridescence                          TextureUsage = 0xff2c91cc
	Albedoopacity01                            TextureUsage = 0x65cf5185
	AlbedoTex                                  TextureUsage = 0xfaee8cb2
	AlbedoWear                                 TextureUsage = 0x04974645
	AnteriorChamberHeightmap                   TextureUsage = 0xdf49f462
	AnteriorChamberNormal                      TextureUsage = 0x5f7c3c06
	AoHeightmap                                TextureUsage = 0x525c46f4
	AoMap                                      TextureUsage = 0xffd0dae1
	AtlasTex                                   TextureUsage = 0x15d5813b
	AtmosphericScatteringColor                 TextureUsage = 0x5e058eef
	AtmosphericScatteringTransmittance         TextureUsage = 0xea1a2316
	BackgroundAlpha                            TextureUsage = 0xd8822282
	BackgroundTexture                          TextureUsage = 0xeafae108
	BadgeFlipbook                              TextureUsage = 0xf487b955
	BakerMaterialAtlas                         TextureUsage = 0xf5417edc
	BaseColor                                  TextureUsage = 0xcb577b8f
	BaseColorMetalMap                          TextureUsage = 0x848ba63b
	BaseColorScrolled                          TextureUsage = 0xb6859653
	BaseData                                   TextureUsage = 0xc2eb8d6e
	BaseMap                                    TextureUsage = 0x265b6040
	BasemapHighlands                           TextureUsage = 0xe4bc4cad
	BasemapLowlands                            TextureUsage = 0xd5e20d7d
	BaseMask                                   TextureUsage = 0xe97a4617
	BaseNormalAo                               TextureUsage = 0x0a775b4c
	BaseNormalAoDirt                           TextureUsage = 0xc5590d97
	BaseNormalAoSubsurface                     TextureUsage = 0xa56b0a85
	BcaTex                                     TextureUsage = 0x2084bf32
	BeamTex01                                  TextureUsage = 0x869f2d3c
	Bgtexture                                  TextureUsage = 0x6379d4f6
	BhNoise                                    TextureUsage = 0xc2bc73f7
	BiomeA66E90B33                             TextureUsage = 0x1a853b82
	BiomeB66E90B33                             TextureUsage = 0xaa4da2bc
	BlockerMaskTarget                          TextureUsage = 0x36c09bf1
	BloodLUT                                   TextureUsage = 0xc7f7c50a
	BloodOverlayNormalGrayscale                TextureUsage = 0xa856db0d
	BloodScalarField                           TextureUsage = 0x2a12ffe5
	BloodSplatterTiler                         TextureUsage = 0x30e2d136
	BloodTiler                                 TextureUsage = 0xa72cb013
	BluenoiseTexture                           TextureUsage = 0x730dc38f
	BrdfLUT                                    TextureUsage = 0x0eaf6cdb
	BrushTexture                               TextureUsage = 0xf106e0e7
	BubbleTex                                  TextureUsage = 0xd9a03f7a
	BubbleTex02                                TextureUsage = 0x152a6b79
	BugSplatterTiler                           TextureUsage = 0x37831285
	CapeGradient                               TextureUsage = 0x8d28873d
	CapeLUT                                    TextureUsage = 0x0e494183
	CapeScalarFields                           TextureUsage = 0x11761abe
	CapeTear                                   TextureUsage = 0xe8953219
	CctvFeed                                   TextureUsage = 0x8642e73f
	ClearcoatNormXyIntensityRoughnessMap       TextureUsage = 0x50563d5a
	ClearcoatRoughnessAoSpec                   TextureUsage = 0x7ca0d044
	CliffsTarget                               TextureUsage = 0x3f006233
	CloakingNoise                              TextureUsage = 0x84da474d
	ClosestReflectionMap                       TextureUsage = 0x741e1cd6
	ClothDetailNormal                          TextureUsage = 0x7bc04cca
	CloudMask                                  TextureUsage = 0x36b33f40
	ColorLUT                                   TextureUsage = 0x6bde7b27
	ColorMap                                   TextureUsage = 0xe1d19c8f
	ColorRoughness                             TextureUsage = 0x8a013406
	ColorRoughnessLUT                          TextureUsage = 0x1a1e55c4
	ColorSpecularB                             TextureUsage = 0x828a53ad
	ColorTintLookup                            TextureUsage = 0x278c1034
	ColorTintMaskMap                           TextureUsage = 0x85c8629f
	ColorVariationMask                         TextureUsage = 0x204eb619
	ColorVariationMra                          TextureUsage = 0xf79fe864
	ColorVariationNormal                       TextureUsage = 0xaf1dc221
	CombinedFogVolumesColorDensity             TextureUsage = 0x278fafaf
	CombinedFogVolumesParameters               TextureUsage = 0x3e2c2780
	CompositeArray                             TextureUsage = 0xa17b45a8
	ConcreteSampler                            TextureUsage = 0x4157f1fc
	ConcreteSurfaceData                        TextureUsage = 0x8d69d2ee
	ContinentsLUT                              TextureUsage = 0x7eceb667
	ContinentsTextureMap                       TextureUsage = 0x9ab8c20d
	CorporateColorRoughnessLUT                 TextureUsage = 0x8a55a8df
	CosmicDustLUT                              TextureUsage = 0x68e55484
	CoveringAlbedo                             TextureUsage = 0x8261a5a5
	CoveringNormal                             TextureUsage = 0x4c6fc000
	CraterTexture                              TextureUsage = 0xafaf10b7
	Cubemap                                    TextureUsage = 0xa63da7cc
	CurrentExposure                            TextureUsage = 0x1ffa86ed
	CustomizationCamoTilerArray                TextureUsage = 0x0f5ff78d
	CustomizationMaterialDetailTilerArray      TextureUsage = 0xd3a0408e
	DamageTilerData                            TextureUsage = 0x75a31a80
	DamageTilerDerivatives                     TextureUsage = 0x0b9d4d67
	DataMapFad03Be1                            TextureUsage = 0x7ce4cbec
	DataTex                                    TextureUsage = 0x9079d5c5
	DataTex02                                  TextureUsage = 0x223ec49a
	DataTexture                                TextureUsage = 0x1cad8dfc
	DecalSheet                                 TextureUsage = 0x632a8b80
	DeformableTerrainMask                      TextureUsage = 0x4e5aab73
	DepthColorLookup                           TextureUsage = 0xd95e3585
	DepthStencilBuffer                         TextureUsage = 0x7c27c084
	Detail                                     TextureUsage = 0x20977eed
	DetailData                                 TextureUsage = 0x25288cc7
	DetailMask                                 TextureUsage = 0x47c29e38
	DetailNormal1                              TextureUsage = 0xdf3ee984
	DetailNormalLeather                        TextureUsage = 0xe719da57
	DetailNormalPorcelain                      TextureUsage = 0x04fb61ad
	DetailNormals                              TextureUsage = 0xbe22de88
	DetailTex                                  TextureUsage = 0xf98491cc
	DiffuseMap                                 TextureUsage = 0x3aa8b87e
	DirMapFad03Be1                             TextureUsage = 0x0eaed0e2
	DirtMap                                    TextureUsage = 0x38e4b36f
	DisplacementMap                            TextureUsage = 0x134f5d4a
	DisplacementTex                            TextureUsage = 0x09bdff27
	DistortionMap                              TextureUsage = 0x08279894
	DistortionTex                              TextureUsage = 0xcc3c5ea2
	Distress                                   TextureUsage = 0x98104a2d
	DistressTexture                            TextureUsage = 0x6f9d2f7a
	DistTex                                    TextureUsage = 0xd605f7e4
	Emissive                                   TextureUsage = 0x12a0f5c0
	EmissiveColor                              TextureUsage = 0xc985395a
	EmissiveFStop10IntensityMap                TextureUsage = 0xca6f2cf1
	EmissiveMap                                TextureUsage = 0x4dc19f08
	EmissiveMask                               TextureUsage = 0x7afa76c6
	Emissivemask                               TextureUsage = 0xc150bf7e
	EmissiveNebulaLUT                          TextureUsage = 0x4a48d13d
	EmissivePack                               TextureUsage = 0x7c5e8b1d
	EmissivePlanet                             TextureUsage = 0x553044cf
	EmissiveTexture                            TextureUsage = 0x319f552e
	ErodeMap                                   TextureUsage = 0x48c5615b
	ErodeTex                                   TextureUsage = 0x3965d9c5
	ErodeTexture                               TextureUsage = 0x119fbe44
	EyeLUT                                     TextureUsage = 0x56f19bbf
	FarFog                                     TextureUsage = 0x3b2b806b
	FarFogDiv4Clouds                           TextureUsage = 0x0fa78c4e
	FarFogDiv4CloudsHistory                    TextureUsage = 0x58f59da9
	FarFogDiv4History                          TextureUsage = 0xfba0d53c
	FarShadows                                 TextureUsage = 0x54ef0508
	FarShadowsVolume                           TextureUsage = 0x15414178
	FillTexture                                TextureUsage = 0x713f136c
	FlareNoise                                 TextureUsage = 0xe80ca12a
	FlashMask                                  TextureUsage = 0x4787d988
	FlashMask02                                TextureUsage = 0x02fe63c5
	FlatteningLookup                           TextureUsage = 0xdbcd240d
	FlattenTexture                             TextureUsage = 0xab08c251
	FlickerTex                                 TextureUsage = 0x330902ed
	FlowMapFad03Be1                            TextureUsage = 0x49753283
	FogVolumeBlobTexture                       TextureUsage = 0xc0eee114
	FogVolumeParticleTexture                   TextureUsage = 0xaa65e64d
	FootstepTiler                              TextureUsage = 0xb60322c5
	GalaxyDust                                 TextureUsage = 0x47430bfb
	GalaxyShape                                TextureUsage = 0x51f8b637
	GalaxyStars                                TextureUsage = 0x7b0d9106
	GasGiantLookup                             TextureUsage = 0x7ce9f561
	Gbuffer0                                   TextureUsage = 0xc82bd6d9
	Gbuffer1                                   TextureUsage = 0xc7dfc461
	Gbuffer1Copy                               TextureUsage = 0xa7559b00
	Gbuffer2                                   TextureUsage = 0xdf52eab1
	Gbuffer3                                   TextureUsage = 0x577c03fc
	GbufferEmissive                            TextureUsage = 0x93be548c
	GeneratedGlobalShaderInput                 TextureUsage = 0xb2817fc4
	GeneratedHeightmap                         TextureUsage = 0x2a9e3232
	GeneratedHeightmapF                        TextureUsage = 0xa2003184
	GeneratedHeightmapOffset                   TextureUsage = 0x4758a80a
	GeneratedHeightmapProperties               TextureUsage = 0xebd2b190
	GeneratedHeightmapPropertiesHeight         TextureUsage = 0x6af0465e
	GeneratedHeightmapSlope                    TextureUsage = 0x26c6b85a
	GeneratedHeightmapSource                   TextureUsage = 0xfdf5405b
	GeneratedHeightmapTmpDiv4                  TextureUsage = 0xfdbff6f7
	GeneratedHeightmapUnormDiv2                TextureUsage = 0x9cf5a972
	GeneratedMaterials                         TextureUsage = 0x95e7e3ac
	GeneratedMinimap                           TextureUsage = 0x2c7a7ec8
	GeneratedMinimapSlot                       TextureUsage = 0x0176b77e
	GeneratedPerZoneShaderInput                TextureUsage = 0x4ca53952
	GeneratedRouteDistance                     TextureUsage = 0xf094187b
	GeneratedTerrainAlbedo                     TextureUsage = 0x057000d9
	GeneratedWaterReplaceLookup                TextureUsage = 0xa85e5c9f
	GlassData                                  TextureUsage = 0xc0dedd9d
	GlintSample                                TextureUsage = 0x213aa1e1
	GlintSample0Dfb75Cf                        TextureUsage = 0x9101f99b
	GlintSample61C13523                        TextureUsage = 0xacf10dd1
	GlobalDiffuseMap                           TextureUsage = 0xdb7e5380
	GlobalLensDirtMap                          TextureUsage = 0xe6ca4466
	GlobalRadianceMap                          TextureUsage = 0x9dbf3864
	GlobalSpecularMap                          TextureUsage = 0x45eb46ca
	Gradient                                   TextureUsage = 0xdae1feb6
	GradientMap                                TextureUsage = 0x153ac5b1
	GradientTexture                            TextureUsage = 0x608855e6
	GraphIntegralTexture                       TextureUsage = 0xc97cf284
	GrayscaleSkin                              TextureUsage = 0x73d40a95
	GridTextureMap                             TextureUsage = 0xf7e8333d
	GroundAlbedo                               TextureUsage = 0x4c446f35
	GroundNar                                  TextureUsage = 0x63066ce9
	GrungeMask                                 TextureUsage = 0x9ded506a
	Hdr0                                       TextureUsage = 0x18c071b4
	Hdr0Div4Fullres                            TextureUsage = 0xf0e28536
	Hdr1                                       TextureUsage = 0xf288cc33
	Hdr2                                       TextureUsage = 0x6a2d571b
	HdrSsr                                     TextureUsage = 0x0c708121
	HeathazeTexture                            TextureUsage = 0xabb548ce
	HeightmapHighlands                         TextureUsage = 0x8b8135ea
	HeightmapLowlands                          TextureUsage = 0x6c61b2bc
	HeightNoise                                TextureUsage = 0xb7af487e
	HeightSample247Fa7Cb                       TextureUsage = 0x85cf08c7
	HighAltitudeCloudsColor                    TextureUsage = 0x68fcaaf7
	HighAltitudeCloudsColorProbe               TextureUsage = 0xa6776d55
	Hmap                                       TextureUsage = 0x65c04d06
	HologramCylinderTextureMap                 TextureUsage = 0x91a9d7fb
	IdMasksArray                               TextureUsage = 0xb281e5f2
	IesLookup                                  TextureUsage = 0x2d03d53b
	IlluminateData                             TextureUsage = 0x90b84a53
	IlluminateDotsTexture                      TextureUsage = 0xb267ec71
	InputImage                                 TextureUsage = 0xf7aafe73
	InputTexture                               TextureUsage = 0x50f4dfaf
	InputTexture0                              TextureUsage = 0x317a6c3b
	InputTexture1                              TextureUsage = 0x621239ee
	InputTexture2                              TextureUsage = 0x510fee22
	InputTexture3                              TextureUsage = 0xb8213704
	InputTexture4                              TextureUsage = 0x310d6e5d
	InputTexture5                              TextureUsage = 0x127e846c
	InputTexture6                              TextureUsage = 0x5e15ee49
	InputTexture7                              TextureUsage = 0x3509ba01
	InputTexture8                              TextureUsage = 0x9622568b
	IrisTiler                                  TextureUsage = 0xbb76289c
	LensCutoutTexture                          TextureUsage = 0x89bbcec2
	LensEmissiveTexture                        TextureUsage = 0x1c121028
	LensOcclusionTexture                       TextureUsage = 0x1c8c3930
	LightBleedMap                              TextureUsage = 0x826c239a
	LightProbeSpaceSpecular                    TextureUsage = 0x8f444022
	LinearDepth                                TextureUsage = 0x9b8038e0
	LinearDepthMip6                            TextureUsage = 0xa1bd1332
	LocalLightsShadowAtlas                     TextureUsage = 0x48c88f1d
	LUTEmissive                                TextureUsage = 0xb70d0e9e
	Mask                                       TextureUsage = 0xc2048121
	MaskAtlasTex                               TextureUsage = 0x2d83e8c7
	MaskTex                                    TextureUsage = 0xe58ff005
	Masktexture                                TextureUsage = 0x41e6d8a5
	MaterialLUT                                TextureUsage = 0x7e662968
	MaterialMap                                TextureUsage = 0xa3e48458
	MaterialTiler                              TextureUsage = 0xb0ac108a
	MetallicMap                                TextureUsage = 0x3be74960
	MetalSurfaceData                           TextureUsage = 0xe32e3fa5
	MindScrambleTexture                        TextureUsage = 0x63c89170
	MinimapBaseColor                           TextureUsage = 0x8ebbf7d9
	MinimapBlockerTarget0                      TextureUsage = 0x253621ed
	MinimapBlockerTarget1                      TextureUsage = 0x95b6c270
	MinimapHeightmap                           TextureUsage = 0x5e62940d
	MinimapLUT                                 TextureUsage = 0x83568228
	MinimapMetallicSubsurfaceRoughnessSpecular TextureUsage = 0x60370cf1
	MinimapNormalWetness                       TextureUsage = 0x3597c022
	MinimapRoutes                              TextureUsage = 0x5fce7f1b
	MinimapSubsurfaceColor                     TextureUsage = 0x53f40e98
	MoonLUT                                    TextureUsage = 0x7dca7a0a
	MotionVectors                              TextureUsage = 0x963c740e
	MRA                                        TextureUsage = 0x756f6fa6
	Mrae                                       TextureUsage = 0x1e17b834
	Mre                                        TextureUsage = 0x282f653d
	MsdfTexture                                TextureUsage = 0x88bac99b
	MudNormalsGrayscale                        TextureUsage = 0x5c774481
	Nac                                        TextureUsage = 0x15f155a7
	NAC                                        TextureUsage = 0x1290c14e
	NAR                                        TextureUsage = 0x4c567810
	Nar                                        TextureUsage = 0x07b6ad0d
	NarTex                                     TextureUsage = 0x1d6fe607
	NarTexture                                 TextureUsage = 0xda759e68
	NmsTex                                     TextureUsage = 0x28f6f19d
	Noise01                                    TextureUsage = 0x2fc55200
	Noise01Texture                             TextureUsage = 0x889e281e
	Noise02                                    TextureUsage = 0x5a3439d0
	Noise02Tex                                 TextureUsage = 0x000024c7
	NoiseArray                                 TextureUsage = 0x44f1ac4d
	NoiseMap01                                 TextureUsage = 0xfbef0214
	NoiseMap02                                 TextureUsage = 0x55c87149
	NoiseNormal                                TextureUsage = 0xd84b77a9
	NoisePack                                  TextureUsage = 0xff1abc29
	NoisePack01                                TextureUsage = 0x55114bec
	NoisePack02                                TextureUsage = 0x7f518c82
	NoiseTex                                   TextureUsage = 0x2a6b6861
	NoiseTex01                                 TextureUsage = 0x48097668
	NoiseTex02                                 TextureUsage = 0x1055283c
	NoiseTexture                               TextureUsage = 0xbc75a128
	Normal                                     TextureUsage = 0xcaed6cd6
	NormalAoRoughness                          TextureUsage = 0x7b4ace3b
	NormalArray                                TextureUsage = 0x495c3189
	NormalMap                                  TextureUsage = 0xf5c97d31
	NormalMap01                                TextureUsage = 0x2b33d35f
	NormalMap02                                TextureUsage = 0x5a3bc7c0
	NormalMapWithAlpha                         TextureUsage = 0x5c3f41f5
	NormalOpacity                              TextureUsage = 0xc572595b
	Normals                                    TextureUsage = 0x080b4d6f
	NormalSpecularAO                           TextureUsage = 0xe64c5236
	NormalXyAoRoughMap                         TextureUsage = 0x1d57dcf3
	NormalXyRoughnessOpacity                   TextureUsage = 0xdab4d0ee
	Nrm01                                      TextureUsage = 0x68188042
	Nrm02                                      TextureUsage = 0x36e05ffa
	Offset01                                   TextureUsage = 0xd77e0552
	Offset02                                   TextureUsage = 0xc6d1a992
	OffsetNoise                                TextureUsage = 0xe35d63cc
	OffsetTexture                              TextureUsage = 0x419f9b2a
	OpacityClipMap                             TextureUsage = 0xcbde381b
	OpacityMap                                 TextureUsage = 0x080375a3
	OutsideMapVistaHeightmap                   TextureUsage = 0xcbf3be84
	OutsideMapVistaHeightmapFrequencyMap       TextureUsage = 0x7b91d320
	OverlayTexture                             TextureUsage = 0xbdfb877a
	OverlayTextureMap                          TextureUsage = 0x0bd341d2
	PackMap                                    TextureUsage = 0x314dd02a
	PaletteLUT                                 TextureUsage = 0x518c858d
	ParallaxMap                                TextureUsage = 0x13ce14bb
	PatternData                                TextureUsage = 0x680e08a4
	PatternLUT                                 TextureUsage = 0x81d4c49d
	PatternMasksArray                          TextureUsage = 0x05a27dd5
	PerlinNoise                                TextureUsage = 0x31ba5da5
	PlanetNoise                                TextureUsage = 0xa8730762
	PrevLinearDepthMip6                        TextureUsage = 0x11e9171e
	PrimaryColorVariationNormalMr              TextureUsage = 0x4086a2bc
	PrimaryMaterialAlbedo                      TextureUsage = 0xa6cb0512
	PrimaryMaterialMask                        TextureUsage = 0xfb19c3bd
	PupilHeightmap                             TextureUsage = 0x6ccd49b6
	PupilNormal                                TextureUsage = 0x31944716
	ReticleTexture                             TextureUsage = 0xbdc07044
	RipplesTiler                               TextureUsage = 0x39b5f733
	RoadDataStrip                              TextureUsage = 0xb9031348
	RoadDirectionTarget                        TextureUsage = 0x7f8fdda9
	RoughnessMap                               TextureUsage = 0xc567338d
	ScatterAlbedoOpacity                       TextureUsage = 0xef092502
	ScatterAlbedoOpacityArray                  TextureUsage = 0xd162d49f
	ScatterComparisionDensity                  TextureUsage = 0x9c2b46ef
	ScatterDensity                             TextureUsage = 0x3aeac078
	ScatterDensityMap                          TextureUsage = 0x00cf75dc
	ScatterLookup                              TextureUsage = 0xb23c7aba
	ScatterNormalArray                         TextureUsage = 0xedc73fa2
	ScatterRshArray                            TextureUsage = 0xd38eb0d7
	ScatterSubsurfaceArray                     TextureUsage = 0x0342a232
	ScenarioOverlayTexture                     TextureUsage = 0x18daef8a
	SclarFieldOpacity                          TextureUsage = 0x479fb1ef
	ScorchMarks                                TextureUsage = 0x5637e5d3
	ScreenEffectDataTexture                    TextureUsage = 0xb2321d1c
	ScreenVideoTextureB                        TextureUsage = 0x0766d158
	ScreenVideoTextureR                        TextureUsage = 0x54f02599
	ScreenVideoTextureY                        TextureUsage = 0xedbefdfd
	ScrollNoise                                TextureUsage = 0x0d4b6759
	Sdf                                        TextureUsage = 0xe9e10861
	SecondaryMap                               TextureUsage = 0xc8bf3b14
	SecondaryMaterialMask                      TextureUsage = 0xe3b84ebc
	ShadowMinimapHeightmap                     TextureUsage = 0xb9fc3452
	ShadowOpacity                              TextureUsage = 0x0d1f1734
	ShipHubSpecularArray                       TextureUsage = 0xf64292e6
	ShipHubSpecularLerpFromArray               TextureUsage = 0x103348d4
	ShipHubSpecularLerpToArray                 TextureUsage = 0x2d95b3ff
	SkyboxEffectTexture                        TextureUsage = 0xa2654500
	SkydomeMap                                 TextureUsage = 0x16f1fadc
	Slot0                                      TextureUsage = 0xd008f798
	SmokePack                                  TextureUsage = 0x73812940
	SnowGlintsTiler                            TextureUsage = 0xc16ed639
	SnowMaskTexture                            TextureUsage = 0x1e706dd3
	SnowPnrbArray                              TextureUsage = 0xbf3f9452
	SpaceProbeBackdrop                         TextureUsage = 0x9b2fc0c2
	SpaceStarLUT                               TextureUsage = 0xda2108c8
	SpaceStarLUTTmp                            TextureUsage = 0x1e4dd1bc
	SpecIriIntensityIriThicknessMap            TextureUsage = 0x42daf1df
	SpecularBrdfLUT                            TextureUsage = 0x6a52658f
	SporeNoise                                 TextureUsage = 0xe740c00c
	SsaoBuffer                                 TextureUsage = 0xcde13465
	SssLUT                                     TextureUsage = 0x776bb418
	SubsurfaceOpacity                          TextureUsage = 0xe7bd9019
	SunFlareImage                              TextureUsage = 0x48557460
	SunFlareVisibilityLookup                   TextureUsage = 0x17de6b6c
	SunFlareVisibilityLookupSum                TextureUsage = 0x03e30890
	SunShadowMap                               TextureUsage = 0xb85584a2
	SurfaceData                                TextureUsage = 0x4309d318
	SurveyQrCode                               TextureUsage = 0xe0ef39f7
	SweepTex                                   TextureUsage = 0x8e0e6a64
	SweepTexture                               TextureUsage = 0x4abe3792
	CoalNt                                     TextureUsage = 0x0c4de4b8
	Coh9                                       TextureUsage = 0xec6ce310
	Coi3                                       TextureUsage = 0x8e726dd7
	TearMap                                    TextureUsage = 0xdc560750
	TearNormalsGrayscale                       TextureUsage = 0xfaecf369
	TerrainBlendingData                        TextureUsage = 0xea6d5de8
	TerrainDisplacementMap                     TextureUsage = 0x08a7b585
	TerrainHeightMask                          TextureUsage = 0x2aa699cb
	TerrainTrample                             TextureUsage = 0x2b04cba9
	TerrainTransparent                         TextureUsage = 0x6f7c7526
	Tex                                        TextureUsage = 0x1070f86a
	Tex0                                       TextureUsage = 0x62225332
	Tex01                                      TextureUsage = 0xcb9cbad2
	Tex02                                      TextureUsage = 0xfe156860
	Tex1                                       TextureUsage = 0xb12f9e15
	Tex2                                       TextureUsage = 0x4c188cf2
	Tex3                                       TextureUsage = 0x44ceab49
	Tex4                                       TextureUsage = 0x52b3389a
	TexMask                                    TextureUsage = 0x49352c01
	TexNormals                                 TextureUsage = 0x16f10ad6
	TextureData                                TextureUsage = 0x91074c9e
	TextureGrayscaleMap                        TextureUsage = 0xd6d31981
	TextureLUT                                 TextureUsage = 0xdbd93d8b
	TextureMap                                 TextureUsage = 0xe503152c
	TextureMap01                               TextureUsage = 0x8a20ca12
	TextureMap02                               TextureUsage = 0xd47db28b
	TextureMap03                               TextureUsage = 0x160e944e
	TextureMap04                               TextureUsage = 0xaac103c0
	TextureMap05                               TextureUsage = 0x5fecd6e5
	TextureMap06                               TextureUsage = 0x4076ac3d
	TextureMap0B1B5Dad                         TextureUsage = 0x9c9baa62
	TextureMap166Acea5                         TextureUsage = 0x599a088b
	TextureMap2252C011                         TextureUsage = 0xe75fa972
	TextureMap2Cfaf55E                         TextureUsage = 0x7e97a216
	TextureMap319D3Bb5                         TextureUsage = 0x9d6c5c34
	TextureMap39265Ef9                         TextureUsage = 0x0c75318b
	TextureMap58F9C7E1                         TextureUsage = 0x4d7509a6
	TextureMap5Bd36001                         TextureUsage = 0x5dbcf8d8
	TextureMap6535D7Ee                         TextureUsage = 0x159caa5f
	TextureMap6A0D408C                         TextureUsage = 0x3b21ae98
	TextureMap7936F626                         TextureUsage = 0x406154f0
	TextureMapAc2Cb2Ee                         TextureUsage = 0xa85f86ee
	TextureMapC8D18519                         TextureUsage = 0x7e501b3b
	TextureMapDb014A68                         TextureUsage = 0x4698846d
	TextureMapE814F87A                         TextureUsage = 0xcbe9cfda
	TextureMapE835D457                         TextureUsage = 0x20eb3b80
	TextureMapF11A51E5                         TextureUsage = 0x2fce962d
	TexturePack                                TextureUsage = 0xd35a9ef3
	TextureRgbaMap                             TextureUsage = 0xa64341a1
	ThumbnailImageToCopy                       TextureUsage = 0x29ed4bea
	TilerArray                                 TextureUsage = 0x82e1eff4
	TrimDecal                                  TextureUsage = 0x6cf6dd98
	TrimSheetHeightC982Efd3                    TextureUsage = 0x5fb4c7cd
	TriplanarDetailAlbedo                      TextureUsage = 0x3bd755ec
	TriplanarDetailData                        TextureUsage = 0x3acb9476
	Ui3DShadows                                TextureUsage = 0x23573aaf
	UiDiffuseCubemap                           TextureUsage = 0xf58c49d8
	UiSpecularCubemap                          TextureUsage = 0xd65153d6
	UiTexture                                  TextureUsage = 0x226c8c28
	UiVideoTextureB                            TextureUsage = 0x0292dde3
	UiVideoTextureR                            TextureUsage = 0xdc50b7ba
	UiVideoTextureY                            TextureUsage = 0x65fca664
	UvNoiseTex                                 TextureUsage = 0x5a576a85
	VegetationBending                          TextureUsage = 0xb1322513
	VeinsData                                  TextureUsage = 0x92923438
	VertexNormalMap                            TextureUsage = 0x2590fd37
	VistaCloudAtlas                            TextureUsage = 0x0a97707c
	VistaCloudSubsurfaceAtlas                  TextureUsage = 0xceb099ff
	VistaDetailAlbedo                          TextureUsage = 0x765c2c9f
	VistaDetailNar                             TextureUsage = 0x018e8164
	VolumetricCloudCurrentWeatherMap           TextureUsage = 0xf8b2f787
	VolumetricCloudDetailNoiseCombined         TextureUsage = 0xa68b2b72
	VolumetricCloudNoiseCombined               TextureUsage = 0x68777a90
	VolumetricCloudsColor                      TextureUsage = 0xa982cdae
	VolumetricCloudsColorProbe                 TextureUsage = 0x147a0440
	VolumetricCloudsDepth                      TextureUsage = 0xbbd0a078
	VolumetricCloudsDepthProbe                 TextureUsage = 0xa689c06b
	VolumetricCloudsPrev                       TextureUsage = 0xb7c1297f
	VolumetricCloudsShadowsFinal               TextureUsage = 0x26afa44b
	VolumetricCloudsWeatherCurrent             TextureUsage = 0x823f1397
	VolumetricCloudWeatherMap0                 TextureUsage = 0x6b835004
	VolumetricCloudWeatherMap1                 TextureUsage = 0xa069e924
	VolumetricCurrentHighAltitudeClouds        TextureUsage = 0x5743f2f8
	VolumetricCurrentHighAltitudeWeatherMap    TextureUsage = 0xb636cbad
	VolumetricFog3DImage                       TextureUsage = 0x8ad188a6
	VolumetricFog3DImageHistory                TextureUsage = 0xbf6abe56
	VolumetricHighAltitudeClouds0              TextureUsage = 0x9595ec11
	VolumetricHighAltitudeClouds1              TextureUsage = 0x5cc4d310
	VolumetricHighAltitudeWeatherMap0          TextureUsage = 0xf4b3cb63
	VolumetricHighAltitudeWeatherMap1          TextureUsage = 0x725f700f
	WaterCaustics                              TextureUsage = 0x4e265efc
	WaterHeight                                TextureUsage = 0x4a673160
	WaterPatch                                 TextureUsage = 0x28442ea9
	WaterRt                                    TextureUsage = 0x71b1e9c9
	WaterTarget                                TextureUsage = 0xd9bb11af
	Wear                                       TextureUsage = 0xefa24853
	WearMask                                   TextureUsage = 0xdb786be1
	WearMra                                    TextureUsage = 0x17310795
	WeatheringAlbedo                           TextureUsage = 0x40796c1b
	WeatheringDataMask                         TextureUsage = 0xb4dcc2c1
	WeatheringDirt                             TextureUsage = 0x6834aa9b
	WeatheringNar                              TextureUsage = 0x0f807c55
	WeatheringSpecial                          TextureUsage = 0xd2f99d38
	WindVectorField                            TextureUsage = 0xc24f6afb
	Worldmask                                  TextureUsage = 0xff58ddae
	WorldOverlayTexture                        TextureUsage = 0x493966e5
	WoundData                                  TextureUsage = 0xf8e31d7b
	WoundDerivative                            TextureUsage = 0xa59f5e11
	WoundLUTToAdd                              TextureUsage = 0xbf479c05
	WoundNormal                                TextureUsage = 0x736a0029
	Wounds256                                  TextureUsage = 0xa52f1caa
	Wounds512                                  TextureUsage = 0x75d9cea2
)

func (usage TextureUsage) String() string {
	switch usage {
	case Albedo:
		return "albedo"
	case AlbedoArray:
		return "albedo_array"
	case AlbedoBlend:
		return "albedo_blend"
	case AlbedoBlendTex:
		return "albedo_blend_tex"
	case AlbedoColor:
		return "albedo_color"
	case AlbedoEmissive:
		return "albedo_emissive"
	case AlbedoIridescence:
		return "albedo_iridescence"
	case Albedoopacity01:
		return "albedoopacity_01"
	case AlbedoTex:
		return "albedo_tex"
	case AlbedoWear:
		return "albedo_wear"
	case AnteriorChamberHeightmap:
		return "anterior_chamber_heightmap"
	case AnteriorChamberNormal:
		return "anterior_chamber_normal"
	case AoHeightmap:
		return "ao_heightmap"
	case AoMap:
		return "ao_map"
	case AtlasTex:
		return "atlas_tex"
	case AtmosphericScatteringColor:
		return "atmospheric_scattering_color"
	case AtmosphericScatteringTransmittance:
		return "atmospheric_scattering_transmittance"
	case BackgroundAlpha:
		return "background_alpha"
	case BackgroundTexture:
		return "background_texture"
	case BadgeFlipbook:
		return "badge_flipbook"
	case BakerMaterialAtlas:
		return "baker_material_atlas"
	case BaseColor:
		return "base_color"
	case BaseColorMetalMap:
		return "base_color_metal_map"
	case BaseColorScrolled:
		return "base_color_scrolled"
	case BaseData:
		return "base_data"
	case BaseMap:
		return "base_map"
	case BasemapHighlands:
		return "basemap_highlands"
	case BasemapLowlands:
		return "basemap_lowlands"
	case BaseMask:
		return "base_mask"
	case BaseNormalAo:
		return "base_normal_ao"
	case BaseNormalAoDirt:
		return "base_normal_ao_dirt"
	case BaseNormalAoSubsurface:
		return "base_normal_ao_subsurface"
	case BcaTex:
		return "bca_tex"
	case BeamTex01:
		return "beam_tex_01"
	case Bgtexture:
		return "bgtexture"
	case BhNoise:
		return "bh_noise"
	case BiomeA66E90B33:
		return "biome_a_66e90b33"
	case BiomeB66E90B33:
		return "biome_b_66e90b33"
	case BlockerMaskTarget:
		return "blocker_mask_target"
	case BloodLUT:
		return "blood_lut"
	case BloodOverlayNormalGrayscale:
		return "blood_overlay_normal_grayscale"
	case BloodScalarField:
		return "blood_scalar_field"
	case BloodSplatterTiler:
		return "blood_splatter_tiler"
	case BloodTiler:
		return "blood_tiler"
	case BluenoiseTexture:
		return "bluenoise_texture"
	case BrdfLUT:
		return "brdf_lut"
	case BrushTexture:
		return "brush_texture"
	case BubbleTex:
		return "bubble_tex"
	case BubbleTex02:
		return "bubble_tex_02"
	case BugSplatterTiler:
		return "bug_splatter_tiler"
	case CapeGradient:
		return "cape_gradient"
	case CapeLUT:
		return "cape_lut"
	case CapeScalarFields:
		return "cape_scalar_fields"
	case CapeTear:
		return "cape_tear"
	case CctvFeed:
		return "cctv_feed"
	case ClearcoatNormXyIntensityRoughnessMap:
		return "clearcoat_norm_xy_intensity_roughness_map"
	case ClearcoatRoughnessAoSpec:
		return "clearcoat_roughness_ao_spec"
	case CliffsTarget:
		return "cliffs_target"
	case CloakingNoise:
		return "cloaking_noise"
	case ClosestReflectionMap:
		return "closest_reflection_map"
	case ClothDetailNormal:
		return "cloth_detail_normal"
	case CloudMask:
		return "cloud_mask"
	case ColorLUT:
		return "color_lut"
	case ColorMap:
		return "color_map"
	case ColorRoughness:
		return "color_roughness"
	case ColorRoughnessLUT:
		return "color_roughness_lut"
	case ColorSpecularB:
		return "color_specular_b"
	case ColorTintLookup:
		return "color_tint_lookup"
	case ColorTintMaskMap:
		return "color_tint_mask_map"
	case ColorVariationMask:
		return "color_variation_mask"
	case ColorVariationMra:
		return "color_variation_mra"
	case ColorVariationNormal:
		return "color_variation_normal"
	case CombinedFogVolumesColorDensity:
		return "combined_fog_volumes_color_density"
	case CombinedFogVolumesParameters:
		return "combined_fog_volumes_parameters"
	case CompositeArray:
		return "composite_array"
	case ConcreteSampler:
		return "concrete_sampler"
	case ConcreteSurfaceData:
		return "concrete_surface_data"
	case ContinentsLUT:
		return "continents_LUT"
	case ContinentsTextureMap:
		return "continents_texture_map"
	case CorporateColorRoughnessLUT:
		return "corporate_color_roughness_lut"
	case CosmicDustLUT:
		return "cosmic_dust_lut"
	case CoveringAlbedo:
		return "covering_albedo"
	case CoveringNormal:
		return "covering_normal"
	case CraterTexture:
		return "crater_texture"
	case Cubemap:
		return "cubemap"
	case CurrentExposure:
		return "current_exposure"
	case CustomizationCamoTilerArray:
		return "customization_camo_tiler_array"
	case CustomizationMaterialDetailTilerArray:
		return "customization_material_detail_tiler_array"
	case DamageTilerData:
		return "damage_tiler_data"
	case DamageTilerDerivatives:
		return "damage_tiler_derivatives"
	case DataMapFad03Be1:
		return "data_map_fad03be1"
	case DataTex:
		return "data_tex"
	case DataTex02:
		return "data_tex_02"
	case DataTexture:
		return "data_texture"
	case DecalSheet:
		return "decal_sheet"
	case DeformableTerrainMask:
		return "deformable_terrain_mask"
	case DepthColorLookup:
		return "depth_color_lookup"
	case DepthStencilBuffer:
		return "depth_stencil_buffer"
	case Detail:
		return "detail"
	case DetailData:
		return "Detail_Data"
	case DetailMask:
		return "detail_mask_"
	case DetailNormal1:
		return "detail_normal_1"
	case DetailNormalLeather:
		return "detail_normal_leather"
	case DetailNormalPorcelain:
		return "detail_normal_porcelain"
	case DetailNormals:
		return "detail_normals"
	case DetailTex:
		return "detail_tex"
	case DiffuseMap:
		return "diffuse_map"
	case DirMapFad03Be1:
		return "dir_map_fad03be1"
	case DirtMap:
		return "dirt_map"
	case DisplacementMap:
		return "displacement_map"
	case DisplacementTex:
		return "displacement_tex"
	case DistortionMap:
		return "distortion_map"
	case DistortionTex:
		return "distortion_tex"
	case Distress:
		return "distress"
	case DistressTexture:
		return "distress_texture"
	case DistTex:
		return "dist_tex"
	case Emissive:
		return "emissive"
	case EmissiveColor:
		return "emissive_color"
	case EmissiveFStop10IntensityMap:
		return "emissive_f_stop_10_intensity_map"
	case EmissiveMap:
		return "emissive_map"
	case EmissiveMask:
		return "emissive_mask"
	case Emissivemask:
		return "emissivemask"
	case EmissiveNebulaLUT:
		return "emissive_nebula_lut"
	case EmissivePack:
		return "emissive_pack"
	case EmissivePlanet:
		return "emissive_planet"
	case EmissiveTexture:
		return "emissive_texture"
	case ErodeMap:
		return "erode_map"
	case ErodeTex:
		return "erode_tex"
	case ErodeTexture:
		return "erode_texture"
	case EyeLUT:
		return "eye_lut"
	case FarFog:
		return "far_fog"
	case FarFogDiv4Clouds:
		return "far_fog_div4_clouds"
	case FarFogDiv4CloudsHistory:
		return "far_fog_div4_clouds_history"
	case FarFogDiv4History:
		return "far_fog_div4_history"
	case FarShadows:
		return "far_shadows"
	case FarShadowsVolume:
		return "far_shadows_volume"
	case FillTexture:
		return "fill_texture"
	case FlareNoise:
		return "flare_noise"
	case FlashMask:
		return "flash_mask"
	case FlashMask02:
		return "flash_mask_02"
	case FlatteningLookup:
		return "flattening_lookup"
	case FlattenTexture:
		return "flatten_texture"
	case FlickerTex:
		return "flicker_tex"
	case FlowMapFad03Be1:
		return "flow_map_fad03be1"
	case FogVolumeBlobTexture:
		return "fog_volume_blob_texture"
	case FogVolumeParticleTexture:
		return "fog_volume_particle_texture"
	case FootstepTiler:
		return "footstep_tiler"
	case GalaxyDust:
		return "galaxy_dust"
	case GalaxyShape:
		return "galaxy_shape"
	case GalaxyStars:
		return "galaxy_stars"
	case GasGiantLookup:
		return "gas_giant_lookup"
	case Gbuffer0:
		return "gbuffer0"
	case Gbuffer1:
		return "gbuffer1"
	case Gbuffer1Copy:
		return "gbuffer1_copy"
	case Gbuffer2:
		return "gbuffer2"
	case Gbuffer3:
		return "gbuffer3"
	case GbufferEmissive:
		return "gbuffer_emissive"
	case GeneratedGlobalShaderInput:
		return "generated_global_shader_input"
	case GeneratedHeightmap:
		return "generated_heightmap"
	case GeneratedHeightmapF:
		return "generated_heightmap_f"
	case GeneratedHeightmapOffset:
		return "generated_heightmap_offset"
	case GeneratedHeightmapProperties:
		return "generated_heightmap_properties"
	case GeneratedHeightmapPropertiesHeight:
		return "generated_heightmap_properties_height"
	case GeneratedHeightmapSlope:
		return "generated_heightmap_slope"
	case GeneratedHeightmapSource:
		return "generated_heightmap_source"
	case GeneratedHeightmapTmpDiv4:
		return "generated_heightmap_tmp_div4"
	case GeneratedHeightmapUnormDiv2:
		return "generated_heightmap_unorm_div2"
	case GeneratedMaterials:
		return "generated_materials"
	case GeneratedMinimap:
		return "generated_minimap"
	case GeneratedMinimapSlot:
		return "generated_minimap_slot"
	case GeneratedPerZoneShaderInput:
		return "generated_per_zone_shader_input"
	case GeneratedRouteDistance:
		return "generated_route_distance"
	case GeneratedTerrainAlbedo:
		return "generated_terrain_albedo"
	case GeneratedWaterReplaceLookup:
		return "generated_water_replace_lookup"
	case GlassData:
		return "glass_data"
	case GlintSample:
		return "glint_sample"
	case GlintSample0Dfb75Cf:
		return "glint_sample_0dfb75cf"
	case GlintSample61C13523:
		return "glint_sample_61c13523"
	case GlobalDiffuseMap:
		return "global_diffuse_map"
	case GlobalLensDirtMap:
		return "global_lens_dirt_map"
	case GlobalRadianceMap:
		return "global_radiance_map"
	case GlobalSpecularMap:
		return "global_specular_map"
	case Gradient:
		return "gradient"
	case GradientMap:
		return "gradient_map"
	case GradientTexture:
		return "gradient_texture"
	case GraphIntegralTexture:
		return "graph_integral_texture"
	case GrayscaleSkin:
		return "grayscale_skin"
	case GridTextureMap:
		return "grid_texture_map"
	case GroundAlbedo:
		return "ground_albedo"
	case GroundNar:
		return "ground_nar"
	case GrungeMask:
		return "grunge_mask"
	case Hdr0:
		return "hdr0"
	case Hdr0Div4Fullres:
		return "hdr0_div4_fullres"
	case Hdr1:
		return "hdr1"
	case Hdr2:
		return "hdr2"
	case HdrSsr:
		return "hdr_ssr"
	case HeathazeTexture:
		return "heathaze_texture"
	case HeightmapHighlands:
		return "heightmap_highlands"
	case HeightmapLowlands:
		return "heightmap_lowlands"
	case HeightNoise:
		return "height_noise"
	case HeightSample247Fa7Cb:
		return "height_sample_247fa7cb"
	case HighAltitudeCloudsColor:
		return "high_altitude_clouds_color"
	case HighAltitudeCloudsColorProbe:
		return "high_altitude_clouds_color_probe"
	case Hmap:
		return "hmap"
	case HologramCylinderTextureMap:
		return "hologram_cylinder_texture_map"
	case IdMasksArray:
		return "id_masks_array"
	case IesLookup:
		return "ies_lookup"
	case IlluminateData:
		return "illuminate_data"
	case IlluminateDotsTexture:
		return "illuminate_dots_texture"
	case InputImage:
		return "input_image"
	case InputTexture:
		return "input_texture"
	case InputTexture0:
		return "input_texture0"
	case InputTexture1:
		return "input_texture1"
	case InputTexture2:
		return "input_texture2"
	case InputTexture3:
		return "input_texture3"
	case InputTexture4:
		return "input_texture4"
	case InputTexture5:
		return "input_texture5"
	case InputTexture6:
		return "input_texture6"
	case InputTexture7:
		return "input_texture7"
	case InputTexture8:
		return "input_texture8"
	case IrisTiler:
		return "iris_tiler"
	case LensCutoutTexture:
		return "lens_cutout_texture"
	case LensEmissiveTexture:
		return "lens_emissive_texture"
	case LensOcclusionTexture:
		return "lens_occlusion_texture"
	case LightBleedMap:
		return "light_bleed_map"
	case LightProbeSpaceSpecular:
		return "light_probe_space_specular"
	case LinearDepth:
		return "linear_depth"
	case LinearDepthMip6:
		return "linear_depth_mip6"
	case LocalLightsShadowAtlas:
		return "local_lights_shadow_atlas"
	case LUTEmissive:
		return "lut_emissive"
	case Mask:
		return "mask"
	case MaskAtlasTex:
		return "mask_atlas_tex"
	case MaskTex:
		return "mask_tex"
	case Masktexture:
		return "masktexture"
	case MaterialLUT:
		return "material_lut"
	case MaterialMap:
		return "material_map"
	case MaterialTiler:
		return "material_tiler"
	case MetallicMap:
		return "metallic_map"
	case MetalSurfaceData:
		return "metal_surface_data"
	case MindScrambleTexture:
		return "mind_scramble_texture"
	case MinimapBaseColor:
		return "minimap_base_color"
	case MinimapBlockerTarget0:
		return "minimap_blocker_target0"
	case MinimapBlockerTarget1:
		return "minimap_blocker_target1"
	case MinimapHeightmap:
		return "minimap_heightmap"
	case MinimapLUT:
		return "minimap_lut"
	case MinimapMetallicSubsurfaceRoughnessSpecular:
		return "minimap_metallic_subsurface_roughness_specular"
	case MinimapNormalWetness:
		return "minimap_normal_wetness"
	case MinimapRoutes:
		return "minimap_routes"
	case MinimapSubsurfaceColor:
		return "minimap_subsurface_color"
	case MoonLUT:
		return "moon_lut"
	case MotionVectors:
		return "motion_vectors"
	case MRA:
		return "mra"
	case Mrae:
		return "mrae"
	case Mre:
		return "mre"
	case MsdfTexture:
		return "msdf_texture"
	case MudNormalsGrayscale:
		return "mud_normals_grayscale"
	case Nac:
		return "nac"
	case NAC:
		return "NAC"
	case Nar:
		return "nar"
	case NAR:
		return "NAR"
	case NarTex:
		return "nar_tex"
	case NarTexture:
		return "nar_texture"
	case NmsTex:
		return "nms_tex"
	case Noise01:
		return "noise_01"
	case Noise01Texture:
		return "noise_01_texture"
	case Noise02:
		return "noise_02"
	case Noise02Tex:
		return "noise_02_tex"
	case NoiseArray:
		return "noise_array"
	case NoiseMap01:
		return "noise_map_01"
	case NoiseMap02:
		return "noise_map_02"
	case NoiseNormal:
		return "noise_normal"
	case NoisePack:
		return "noise_pack"
	case NoisePack01:
		return "noise_pack_01"
	case NoisePack02:
		return "noise_pack_02"
	case NoiseTex:
		return "noise_tex"
	case NoiseTex01:
		return "noise_tex_01"
	case NoiseTex02:
		return "noise_tex_02"
	case NoiseTexture:
		return "noise_texture"
	case Normal:
		return "normal"
	case NormalAoRoughness:
		return "normal_ao_roughness"
	case NormalArray:
		return "normal_array"
	case NormalMap:
		return "normal_map"
	case NormalMap01:
		return "normal_map_01"
	case NormalMap02:
		return "normal_map_02"
	case NormalMapWithAlpha:
		return "normal_map_with_alpha"
	case NormalOpacity:
		return "normal_opacity"
	case Normals:
		return "normals"
	case NormalSpecularAO:
		return "normal_specular_ao"
	case NormalXyAoRoughMap:
		return "normal_xy_ao_rough_map"
	case NormalXyRoughnessOpacity:
		return "normal_xy_roughness_opacity"
	case Nrm01:
		return "nrm_01"
	case Nrm02:
		return "nrm_02"
	case Offset01:
		return "offset_01"
	case Offset02:
		return "offset_02"
	case OffsetNoise:
		return "offset_noise"
	case OffsetTexture:
		return "offset_texture"
	case OpacityClipMap:
		return "opacity_clip_map"
	case OpacityMap:
		return "opacity_map"
	case OutsideMapVistaHeightmap:
		return "outside_map_vista_heightmap"
	case OutsideMapVistaHeightmapFrequencyMap:
		return "outside_map_vista_heightmap_frequency_map"
	case OverlayTexture:
		return "overlay_texture"
	case OverlayTextureMap:
		return "overlay_texture_map"
	case PackMap:
		return "pack_map"
	case PaletteLUT:
		return "palette_lut"
	case ParallaxMap:
		return "parallax_map"
	case PatternData:
		return "pattern_data"
	case PatternLUT:
		return "pattern_lut"
	case PatternMasksArray:
		return "pattern_masks_array"
	case PerlinNoise:
		return "perlin_noise"
	case PlanetNoise:
		return "planet_noise"
	case PrevLinearDepthMip6:
		return "prev_linear_depth_mip6"
	case PrimaryColorVariationNormalMr:
		return "primary_color_variation_normal_mr"
	case PrimaryMaterialAlbedo:
		return "primary_material_albedo"
	case PrimaryMaterialMask:
		return "primary_material_mask"
	case PupilHeightmap:
		return "pupil_heightmap"
	case PupilNormal:
		return "pupil_normal"
	case ReticleTexture:
		return "reticle_texture"
	case RipplesTiler:
		return "ripples_tiler"
	case RoadDataStrip:
		return "road_data_strip"
	case RoadDirectionTarget:
		return "road_direction_target"
	case RoughnessMap:
		return "roughness_map"
	case ScatterAlbedoOpacity:
		return "scatter_albedo_opacity"
	case ScatterAlbedoOpacityArray:
		return "scatter_albedo_opacity_array"
	case ScatterComparisionDensity:
		return "scatter_comparision_density"
	case ScatterDensity:
		return "scatter_density"
	case ScatterDensityMap:
		return "scatter_density_map"
	case ScatterLookup:
		return "scatter_lookup"
	case ScatterNormalArray:
		return "scatter_normal_array"
	case ScatterRshArray:
		return "scatter_rsh_array"
	case ScatterSubsurfaceArray:
		return "scatter_subsurface_array"
	case ScenarioOverlayTexture:
		return "scenario_overlay_texture"
	case SclarFieldOpacity:
		return "sclar_field_opacity"
	case ScorchMarks:
		return "scorch_marks"
	case ScreenEffectDataTexture:
		return "screen_effect_data_texture"
	case ScreenVideoTextureB:
		return "screen_video_texture_b"
	case ScreenVideoTextureR:
		return "screen_video_texture_r"
	case ScreenVideoTextureY:
		return "screen_video_texture_y"
	case ScrollNoise:
		return "scroll_noise"
	case Sdf:
		return "sdf"
	case SecondaryMap:
		return "secondary_map"
	case SecondaryMaterialMask:
		return "secondary_material_mask"
	case ShadowMinimapHeightmap:
		return "shadow_minimap_heightmap"
	case ShadowOpacity:
		return "shadow_opacity"
	case ShipHubSpecularArray:
		return "ship_hub_specular_array"
	case ShipHubSpecularLerpFromArray:
		return "ship_hub_specular_lerp_from_array"
	case ShipHubSpecularLerpToArray:
		return "ship_hub_specular_lerp_to_array"
	case SkyboxEffectTexture:
		return "skybox_effect_texture"
	case SkydomeMap:
		return "skydome_map"
	case Slot0:
		return "slot_0"
	case SmokePack:
		return "smoke_pack"
	case SnowGlintsTiler:
		return "snow_glints_tiler"
	case SnowMaskTexture:
		return "snow_mask_texture"
	case SnowPnrbArray:
		return "snow_pnrb_array"
	case SpaceProbeBackdrop:
		return "space_probe_backdrop"
	case SpaceStarLUT:
		return "space_star_lut"
	case SpaceStarLUTTmp:
		return "space_star_lut_tmp"
	case SpecIriIntensityIriThicknessMap:
		return "spec_iri_intensity_iri_thickness_map"
	case SpecularBrdfLUT:
		return "specular_brdf_lut"
	case SporeNoise:
		return "spore_noise"
	case SsaoBuffer:
		return "ssao_buffer"
	case SssLUT:
		return "sss_lut"
	case SubsurfaceOpacity:
		return "subsurface_opacity"
	case SunFlareImage:
		return "sun_flare_image"
	case SunFlareVisibilityLookup:
		return "sun_flare_visibility_lookup"
	case SunFlareVisibilityLookupSum:
		return "sun_flare_visibility_lookup_sum"
	case SunShadowMap:
		return "sun_shadow_map"
	case SurfaceData:
		return "surface_data"
	case SurveyQrCode:
		return "survey_qr_code"
	case SweepTex:
		return "sweep_tex"
	case SweepTexture:
		return "sweep_texture"
	case CoalNt:
		return "coal_nt"
	case Coh9:
		return "coH9"
	case Coi3:
		return "coI3"
	case TearMap:
		return "tear_map"
	case TearNormalsGrayscale:
		return "tear_normals_grayscale"
	case TerrainBlendingData:
		return "terrain_blending_data"
	case TerrainDisplacementMap:
		return "terrain_displacement_map"
	case TerrainHeightMask:
		return "terrain_height_mask"
	case TerrainTrample:
		return "terrain_trample"
	case TerrainTransparent:
		return "terrain_transparent"
	case Tex:
		return "tex"
	case Tex0:
		return "tex0"
	case Tex01:
		return "tex_01"
	case Tex02:
		return "tex_02"
	case Tex1:
		return "tex1"
	case Tex2:
		return "tex2"
	case Tex3:
		return "tex3"
	case Tex4:
		return "tex4"
	case TexMask:
		return "tex_mask"
	case TexNormals:
		return "tex_normals"
	case TextureData:
		return "texture_data"
	case TextureGrayscaleMap:
		return "texture_grayscale_map"
	case TextureLUT:
		return "texture_lut"
	case TextureMap:
		return "texture_map"
	case TextureMap01:
		return "texture_map_01"
	case TextureMap02:
		return "texture_map_02"
	case TextureMap03:
		return "texture_map_03"
	case TextureMap04:
		return "texture_map_04"
	case TextureMap05:
		return "texture_map_05"
	case TextureMap06:
		return "texture_map_06"
	case TextureMap0B1B5Dad:
		return "texture_map_0b1b5dad"
	case TextureMap166Acea5:
		return "texture_map_166acea5"
	case TextureMap2252C011:
		return "texture_map_2252c011"
	case TextureMap2Cfaf55E:
		return "texture_map_2cfaf55e"
	case TextureMap319D3Bb5:
		return "texture_map_319d3bb5"
	case TextureMap39265Ef9:
		return "texture_map_39265ef9"
	case TextureMap58F9C7E1:
		return "texture_map_58f9c7e1"
	case TextureMap5Bd36001:
		return "texture_map_5bd36001"
	case TextureMap6535D7Ee:
		return "texture_map_6535d7ee"
	case TextureMap6A0D408C:
		return "texture_map_6a0d408c"
	case TextureMap7936F626:
		return "texture_map_7936f626"
	case TextureMapAc2Cb2Ee:
		return "texture_map_ac2cb2ee"
	case TextureMapC8D18519:
		return "texture_map_c8d18519"
	case TextureMapDb014A68:
		return "texture_map_db014a68"
	case TextureMapE814F87A:
		return "texture_map_e814f87a"
	case TextureMapE835D457:
		return "texture_map_e835d457"
	case TextureMapF11A51E5:
		return "texture_map_f11a51e5"
	case TexturePack:
		return "texture_pack"
	case TextureRgbaMap:
		return "texture_rgba_map"
	case ThumbnailImageToCopy:
		return "thumbnail_image_to_copy"
	case TilerArray:
		return "tiler_array"
	case TrimDecal:
		return "trim_decal"
	case TrimSheetHeightC982Efd3:
		return "trim_sheet_height_c982efd3"
	case TriplanarDetailAlbedo:
		return "triplanar_detail_albedo"
	case TriplanarDetailData:
		return "triplanar_detail_data"
	case Ui3DShadows:
		return "ui_3d_shadows"
	case UiDiffuseCubemap:
		return "ui_diffuse_cubemap"
	case UiSpecularCubemap:
		return "ui_specular_cubemap"
	case UiTexture:
		return "ui_texture"
	case UiVideoTextureB:
		return "ui_video_texture_b"
	case UiVideoTextureR:
		return "ui_video_texture_r"
	case UiVideoTextureY:
		return "ui_video_texture_y"
	case UvNoiseTex:
		return "uv_noise_tex"
	case VegetationBending:
		return "vegetation_bending"
	case VeinsData:
		return "veins_data"
	case VertexNormalMap:
		return "vertex_normal_map"
	case VistaCloudAtlas:
		return "vista_cloud_atlas"
	case VistaCloudSubsurfaceAtlas:
		return "vista_cloud_subsurface_atlas"
	case VistaDetailAlbedo:
		return "vista_detail_albedo"
	case VistaDetailNar:
		return "vista_detail_nar"
	case VolumetricCloudCurrentWeatherMap:
		return "volumetric_cloud_current_weather_map"
	case VolumetricCloudDetailNoiseCombined:
		return "volumetric_cloud_detail_noise_combined"
	case VolumetricCloudNoiseCombined:
		return "volumetric_cloud_noise_combined"
	case VolumetricCloudsColor:
		return "volumetric_clouds_color"
	case VolumetricCloudsColorProbe:
		return "volumetric_clouds_color_probe"
	case VolumetricCloudsDepth:
		return "volumetric_clouds_depth"
	case VolumetricCloudsDepthProbe:
		return "volumetric_clouds_depth_probe"
	case VolumetricCloudsPrev:
		return "volumetric_clouds_prev"
	case VolumetricCloudsShadowsFinal:
		return "volumetric_clouds_shadows_final"
	case VolumetricCloudsWeatherCurrent:
		return "volumetric_clouds_weather_current"
	case VolumetricCloudWeatherMap0:
		return "volumetric_cloud_weather_map0"
	case VolumetricCloudWeatherMap1:
		return "volumetric_cloud_weather_map1"
	case VolumetricCurrentHighAltitudeClouds:
		return "volumetric_current_high_altitude_clouds"
	case VolumetricCurrentHighAltitudeWeatherMap:
		return "volumetric_current_high_altitude_weather_map"
	case VolumetricFog3DImage:
		return "volumetric_fog_3d_image"
	case VolumetricFog3DImageHistory:
		return "volumetric_fog_3d_image_history"
	case VolumetricHighAltitudeClouds0:
		return "volumetric_high_altitude_clouds0"
	case VolumetricHighAltitudeClouds1:
		return "volumetric_high_altitude_clouds1"
	case VolumetricHighAltitudeWeatherMap0:
		return "volumetric_high_altitude_weather_map0"
	case VolumetricHighAltitudeWeatherMap1:
		return "volumetric_high_altitude_weather_map1"
	case WaterCaustics:
		return "water_caustics"
	case WaterHeight:
		return "water_height"
	case WaterPatch:
		return "water_patch"
	case WaterRt:
		return "water_rt"
	case WaterTarget:
		return "water_target"
	case Wear:
		return "wear"
	case WearMask:
		return "wear_mask"
	case WearMra:
		return "wear_mra"
	case WeatheringAlbedo:
		return "weathering_albedo"
	case WeatheringDataMask:
		return "weathering_data_mask"
	case WeatheringDirt:
		return "weathering_dirt"
	case WeatheringNar:
		return "weathering_nar"
	case WeatheringSpecial:
		return "weathering_special"
	case WindVectorField:
		return "wind_vector_field"
	case Worldmask:
		return "worldmask"
	case WorldOverlayTexture:
		return "world_overlay_texture"
	case WoundData:
		return "wound_data"
	case WoundDerivative:
		return "wound_derivative"
	case WoundLUTToAdd:
		return "wound_lut_to_add"
	case WoundNormal:
		return "wound_normal"
	case Wounds256:
		return "wounds_256"
	case Wounds512:
		return "wounds_512"
	default:
		return "unknown texture usage!"
	}
}
