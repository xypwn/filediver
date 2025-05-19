package material

type SettingsUsage uint32

const (
	SettingAlpha                                SettingsUsage = 0x3f697354
	SettingEnvironmentLuminosity                SettingsUsage = 0xeee59e17
	SettingSineAmplitude01                      SettingsUsage = 0x3e7b097e
	SettingSineFrequency                        SettingsUsage = 0xb7088fa1
	SettingSineSecondaryAmplitude               SettingsUsage = 0xfcae94d5
	SettingSineSecondaryFrequency               SettingsUsage = 0xca7baccd
	SettingSineSecondarySpeed                   SettingsUsage = 0xc487fba8
	SettingSineSpeed                            SettingsUsage = 0x25db2966
	SettingTimeWorldPosOffset                   SettingsUsage = 0x877d9a68
	SettingUseDirectionFromCenter               SettingsUsage = 0x4b9f425e
	SettingUseUvVertexMaskVExp                  SettingsUsage = 0x796f72b1
	SettingUvMask                               SettingsUsage = 0x2a8279db
	SettingUvUMask                              SettingsUsage = 0x7b95f38e
	SettingUvUMaskMax                           SettingsUsage = 0xe67e858c
	SettingUvUMaskMin                           SettingsUsage = 0x5f4dd44b
	SettingUvUMultiplier                        SettingsUsage = 0x45b397bb
	SettingUvVMask                              SettingsUsage = 0x41b2516d
	SettingUvVMaskMax                           SettingsUsage = 0x9e70de9e
	SettingUvVMaskMin                           SettingsUsage = 0xa9221b63
	SettingUvVertexMask                         SettingsUsage = 0x7cbf0b6d
	SettingUvVertexMaskMax                      SettingsUsage = 0x40d08754
	SettingUvVertexMaskMin                      SettingsUsage = 0xf589acc8
	SettingUvVertexMaskVExp                     SettingsUsage = 0x279f6869
	SettingAaMultiplier                         SettingsUsage = 0x90ed0fa2
	SettingAberrateBlue                         SettingsUsage = 0xc329914f
	SettingAberrateGreen                        SettingsUsage = 0xe7e5149d
	SettingAberrateRed                          SettingsUsage = 0x219aa092
	SettingAbsoluteHeight                       SettingsUsage = 0xe0d994fa
	SettingAlbedoIntensity                      SettingsUsage = 0x75296d66
	SettingAlbedoIntensityGround                SettingsUsage = 0xa4136d39
	SettingAlbedoIntensityRock                  SettingsUsage = 0x077e7c10
	SettingAlbedoIntensityVistaDetail           SettingsUsage = 0x2b8c01b5
	SettingAlphaBackground                      SettingsUsage = 0x6be0e412
	SettingAlphaBorder                          SettingsUsage = 0x2161cba5
	SettingAlphaFill                            SettingsUsage = 0xe8cf8673
	SettingAlphaMultiplier                      SettingsUsage = 0x6d8e223b
	SettingAmbientAmoint                        SettingsUsage = 0x8a572aa0
	SettingAmbientAmount                        SettingsUsage = 0x7426175c
	SettingAngleFade                            SettingsUsage = 0xf0099378
	SettingAngleFadeEnd                         SettingsUsage = 0xdf456b7e
	SettingAngleFadeStart                       SettingsUsage = 0x94db445e
	SettingArrayIndex                           SettingsUsage = 0x7f556f56
	SettingAtlasSelect                          SettingsUsage = 0xf14b756c
	SettingAtlasSizeRatio                       SettingsUsage = 0x29d1f98d
	SettingAtmosphereLightColor                 SettingsUsage = 0x3e6361b1
	SettingAtmosphereLightDirection             SettingsUsage = 0x7eacbff2
	SettingAtmosphereRimlightColor              SettingsUsage = 0x8d27e8a7
	SettingAtmosphereRimlightIntensity          SettingsUsage = 0x2e766f48
	SettingAtmosphereSaturation                 SettingsUsage = 0x56814e54
	SettingAtmosphericLookup                    SettingsUsage = 0x23e9ba06
	SettingBackFaceVisibility                   SettingsUsage = 0xbf4f4e8c
	SettingBackgroundTile                       SettingsUsage = 0x6aba6c1f
	SettingBarrelOffset                         SettingsUsage = 0x950b0164
	SettingBaseAoIntensity                      SettingsUsage = 0xa235fb52
	SettingBaseColor                            SettingsUsage = 0xcb577b8f
	SettingBaseColorOpacity                     SettingsUsage = 0xe2721bdb
	SettingBaseNormalIntensity                  SettingsUsage = 0xcff312b4
	SettingBaseOpacity                          SettingsUsage = 0xae7ff202
	SettingBcMultiplier                         SettingsUsage = 0x3266408c
	SettingBiomeATile                           SettingsUsage = 0x66a04f3d
	SettingBiomeBTile                           SettingsUsage = 0x595acdc8
	SettingBlackHoleEnable                      SettingsUsage = 0xec381b0a
	SettingBloodColor                           SettingsUsage = 0x5461f4e2
	SettingBloodColor0                          SettingsUsage = 0x8b09df61
	SettingBloodColor1                          SettingsUsage = 0x8753c2fe
	SettingBloodGunkNormalIntensity             SettingsUsage = 0xd7aaa0c3
	SettingBloodNormalFade                      SettingsUsage = 0x0c4b94a3
	SettingBloodRoughness                       SettingsUsage = 0xc998ac9f
	SettingBloodScale                           SettingsUsage = 0x2b05328b
	SettingBloodSubsurface                      SettingsUsage = 0x834e87ba
	SettingBloodWeightsNegative                 SettingsUsage = 0x26102c1e
	SettingBloodWeightsPositive                 SettingsUsage = 0x0955ef2e
	SettingBombSharpen                          SettingsUsage = 0xdf2483f0
	SettingBool                                 SettingsUsage = 0xe58cf2ef
	SettingBorderFalloff                        SettingsUsage = 0x7f002325
	SettingBorderWidth                          SettingsUsage = 0x371855bc
	SettingBottomColorTint                      SettingsUsage = 0x8b1f2c0b
	SettingBottomDirtValue                      SettingsUsage = 0xaa4d1c48
	SettingBoundingVolume                       SettingsUsage = 0x62c16ee4
	SettingBugGunkMinimum                       SettingsUsage = 0xde01726b
	SettingBugGunkWeightsNegative               SettingsUsage = 0x03718e59
	SettingBugGunkWeightsPositive               SettingsUsage = 0x0471d063
	SettingBurnScorch                           SettingsUsage = 0xcd18a390
	SettingBurnScorchSnowMeltGlobalWetness      SettingsUsage = 0xffb232e9
	SettingCAtmosphereCommon                    SettingsUsage = 0xf0124e19
	SettingCCloudStartStop                      SettingsUsage = 0x74fa9302
	SettingCPerInstance                         SettingsUsage = 0x41d3f841
	SettingCPerObject                           SettingsUsage = 0xb5639618
	SettingCUi3d                                SettingsUsage = 0x3a375a94
	SettingCameraCenterPos                      SettingsUsage = 0x0ac8f192
	SettingCameraFadeDistance                   SettingsUsage = 0xc7ac7d10
	SettingCameraInvProjection                  SettingsUsage = 0xa2c5876b
	SettingCameraInvView                        SettingsUsage = 0xf596e8d8
	SettingCameraLastInvProjection              SettingsUsage = 0x05cc9a47
	SettingCameraLastInvView                    SettingsUsage = 0x32ed1500
	SettingCameraLastProjection                 SettingsUsage = 0x8c3a6f9e
	SettingCameraLastView                       SettingsUsage = 0x5f86293b
	SettingCameraLastViewProjection             SettingsUsage = 0x243aaefd
	SettingCameraNearFar                        SettingsUsage = 0xcdf7c2bf
	SettingCameraPos                            SettingsUsage = 0x75053e0e
	SettingCameraProjection                     SettingsUsage = 0x77143f40
	SettingCameraUnprojection                   SettingsUsage = 0x6bc91d73
	SettingCameraView                           SettingsUsage = 0x6a30b459
	SettingCameraViewProjection                 SettingsUsage = 0x097e710f
	SettingCapeHeightFromGround                 SettingsUsage = 0x646bee65
	SettingCapeHeightMult                       SettingsUsage = 0x361a7618
	SettingCenterLightDistanceFalloff           SettingsUsage = 0x231e5ba4
	SettingCenterLightDistanceMult              SettingsUsage = 0x612ebc88
	SettingCenterLightIntensity                 SettingsUsage = 0x2cb85d64
	SettingChannelSelection                     SettingsUsage = 0xe9b736c5
	SettingCivilizationAmount                   SettingsUsage = 0x147dc9d0
	SettingClearcoatIntensity                   SettingsUsage = 0xb132b95e
	SettingClearcoatNormalMix                   SettingsUsage = 0x92607599
	SettingClearcoatRoughness                   SettingsUsage = 0x868e1f43
	SettingClipBox                              SettingsUsage = 0x76d8fcf4
	SettingClipCenter                           SettingsUsage = 0xfbf5a856
	SettingClipDistance                         SettingsUsage = 0xb1a01016
	SettingClosestReflectionMap                 SettingsUsage = 0x741e1cd6
	SettingCloudAmount                          SettingsUsage = 0x3f90d98e
	SettingCloudColor                           SettingsUsage = 0xe3a46787
	SettingCloudContrast                        SettingsUsage = 0x70bd8205
	SettingCloudHeightMult                      SettingsUsage = 0x379b2799
	SettingCloudOpacity                         SettingsUsage = 0xb1b4526a
	SettingCloudRotate                          SettingsUsage = 0x7157fd74
	SettingCloudSpeedU                          SettingsUsage = 0x18cf9121
	SettingCloudSpeedV                          SettingsUsage = 0x91343b7a
	SettingCloudStartHeight                     SettingsUsage = 0xafc15a88
	SettingClusteredShadingData                 SettingsUsage = 0xd3e92882
	SettingColor                                SettingsUsage = 0x06776dda
	SettingColorBorder                          SettingsUsage = 0x4a299572
	SettingColorEdges                           SettingsUsage = 0xf35d9df7
	SettingColorFill                            SettingsUsage = 0x83d478f9
	SettingColorIntensity                       SettingsUsage = 0xa336f8bd
	SettingColorLean                            SettingsUsage = 0x85abf835
	SettingColorMult                            SettingsUsage = 0x0f3bc3a8
	SettingColorMulti                           SettingsUsage = 0x1876aaee
	SettingColorTint                            SettingsUsage = 0xe7a2c776
	SettingColorVariationHighland               SettingsUsage = 0x2eeab085
	SettingColorVariationLowland                SettingsUsage = 0x3183af69
	SettingConeRadiusAdjust                     SettingsUsage = 0x16ae0544
	SettingContextCamera                        SettingsUsage = 0xea353973
	SettingCoriolisForce                        SettingsUsage = 0x0e27ae9f
	SettingCoriolisOffset                       SettingsUsage = 0xd12eccb6
	SettingCsActive                             SettingsUsage = 0xea7a0a15
	SettingCsCameraViewProj                     SettingsUsage = 0x924447ee
	SettingCsClusterBuffer                      SettingsUsage = 0x557c9255
	SettingCsClusterDataSize                    SettingsUsage = 0xe4aa9d19
	SettingCsClusterMaxDepthInvMaxDepth         SettingsUsage = 0xa2a89e2a
	SettingCsClusterSizeInPixels                SettingsUsage = 0x792124ab
	SettingCsClusterSizes                       SettingsUsage = 0x056a5852
	SettingCsLightDataBuffer                    SettingsUsage = 0x73de4d6a
	SettingCsLightDataSize                      SettingsUsage = 0x16622f2f
	SettingCsLightIndexBuffer                   SettingsUsage = 0x5fd09219
	SettingCsLightIndexDataSize                 SettingsUsage = 0xfcdf2ffd
	SettingCsLightShadowMatricesBuffer          SettingsUsage = 0x0cd40b8b
	SettingCsLightShadowMatricesSize            SettingsUsage = 0x774f9271
	SettingCsShadowAtlasSize                    SettingsUsage = 0xbfd8646e
	SettingCubemapFrameHistoryInvalidation      SettingsUsage = 0xa8ea157d
	SettingDarkestValue                         SettingsUsage = 0x6d9a0b58
	SettingDebugLod                             SettingsUsage = 0xeb4b80a1
	SettingDebugMode                            SettingsUsage = 0x205e89cc
	SettingDebugRendering                       SettingsUsage = 0x217a0094
	SettingDebugShadowLod                       SettingsUsage = 0x303e5b68
	SettingDebugSpace                           SettingsUsage = 0xcc4680f5
	SettingDecalAlphaOffset                     SettingsUsage = 0xf339d7b2
	SettingDecalAlphaSharpness                  SettingsUsage = 0xaba6d287
	SettingDecalFadeExp                         SettingsUsage = 0x3ee74976
	SettingDecalNormalIntensity                 SettingsUsage = 0x32ee23d2
	SettingDecalNormalIntentity                 SettingsUsage = 0x7af1aa8d
	SettingDecalNormalOffset                    SettingsUsage = 0xc29c73f9
	SettingDecalScalarfieldEnd                  SettingsUsage = 0x5d8c6616
	SettingDeepWaterColor                       SettingsUsage = 0xe779fe59
	SettingDeltaTime                            SettingsUsage = 0x53e3339e
	SettingDepth                                SettingsUsage = 0x911ffdcb
	SettingDepthFade                            SettingsUsage = 0x57754749
	SettingDepthFadeDist                        SettingsUsage = 0x12695573
	SettingDepthFadeDistance                    SettingsUsage = 0x38799eb5
	SettingDeriveNormalZ                        SettingsUsage = 0x4b756035
	SettingDesaturation                         SettingsUsage = 0x6daca008
	SettingDetailCurvIntensity1                 SettingsUsage = 0xaa926e5f
	SettingDetailCurvIntensity2                 SettingsUsage = 0x1eacab11
	SettingDetailCurvIntensityLeather           SettingsUsage = 0x6e8e2d58
	SettingDetailCurvIntensityPorcelain         SettingsUsage = 0x02325c30
	SettingDetailMixWeight                      SettingsUsage = 0xab51994f
	SettingDetailNormClearcoatIntensity         SettingsUsage = 0x77ce49fa
	SettingDetailNormClearcoatIntensity1        SettingsUsage = 0xdc96c59b
	SettingDetailNormClearcoatIntensity2        SettingsUsage = 0xe01c40e2
	SettingDetailNormIntensity1                 SettingsUsage = 0x9dfcaa44
	SettingDetailNormIntensity2                 SettingsUsage = 0xe21aff4a
	SettingDetailNormIntensityLeather           SettingsUsage = 0x11855634
	SettingDetailNormIntensityPorcelain         SettingsUsage = 0x4c5d8840
	SettingDetailNormTiler1                     SettingsUsage = 0x08f9c22a
	SettingDetailNormTiler2                     SettingsUsage = 0xf7fc3e88
	SettingDetailNormTilerLeather               SettingsUsage = 0xcff8242b
	SettingDetailNormTilerPorcelain             SettingsUsage = 0x8de83b3e
	SettingDetailNormalIntensity                SettingsUsage = 0x673c6fec
	SettingDetailNormalSize                     SettingsUsage = 0xe7880498
	SettingDetailRoughnessClearcoatIntensity1   SettingsUsage = 0x9a2dbee4
	SettingDetailRoughnessClearcoatIntensity2   SettingsUsage = 0x014cb700
	SettingDetailTileFactorMult                 SettingsUsage = 0xcfa59fe4
	SettingDevSelectionColor                    SettingsUsage = 0x588e5a30
	SettingDevSelectionMask                     SettingsUsage = 0x425efe3d
	SettingDiffuseIntensity                     SettingsUsage = 0x3de54d1b
	SettingDirtAo                               SettingsUsage = 0xe3d41d1c
	SettingDirtAoCoverage                       SettingsUsage = 0x64aab07b
	SettingDirtAoSharpness                      SettingsUsage = 0x811d1cae
	SettingDirtColor                            SettingsUsage = 0xbf23c180
	SettingDirtDetailAo                         SettingsUsage = 0xe3d919ca
	SettingDirtDetailMask                       SettingsUsage = 0xa9ab99c1
	SettingDirtDetailMasking                    SettingsUsage = 0xa83f44cd
	SettingDirtDetailSharpness                  SettingsUsage = 0x46fb4810
	SettingDirtGlobalAmount                     SettingsUsage = 0xa3351311
	SettingDirtGradientMax                      SettingsUsage = 0x6ddbae8f
	SettingDirtGradientMin                      SettingsUsage = 0x6fd0b9e7
	SettingDirtIntensity                        SettingsUsage = 0x729b4cb1
	SettingDirtMetallic                         SettingsUsage = 0xee9b82bf
	SettingDirtRoughness                        SettingsUsage = 0xb3d27aa0
	SettingDirtRoughnessBlend                   SettingsUsage = 0xc0f83284
	SettingDirtSharpness                        SettingsUsage = 0x1b90e5fc
	SettingDisplaceUvs                          SettingsUsage = 0x52e39204
	SettingDisplacementScale                    SettingsUsage = 0xb7ef9ee4
	SettingDistFadeOffset                       SettingsUsage = 0x918b6a8c
	SettingDistSpeed                            SettingsUsage = 0x6fd76644
	SettingDistTile                             SettingsUsage = 0x467dd43d
	SettingDistortion                           SettingsUsage = 0x4562fafc
	SettingDistortionAmount                     SettingsUsage = 0xc9f0407a
	SettingDistressUvInfo                       SettingsUsage = 0xf1c6d779
	SettingDrynessAmount                        SettingsUsage = 0x752d6038
	SettingDustColor                            SettingsUsage = 0xe5fb6c6d
	SettingDustFbm                              SettingsUsage = 0x30ded2fe
	SettingDustOpacity                          SettingsUsage = 0xe8e55d4d
	SettingDustTilingAmount                     SettingsUsage = 0x9adca68c
	SettingDword                                SettingsUsage = 0x72b0a98e
	SettingEdgeFadeOffset                       SettingsUsage = 0x6ce72e4f
	SettingEdgeFadeTint                         SettingsUsage = 0x0963ca93
	SettingEdgeNormalIntensity                  SettingsUsage = 0xf2f0a1a3
	SettingEdgeNormalSharpness                  SettingsUsage = 0x03bbaf53
	SettingEmissiveAnimation                    SettingsUsage = 0x7415ab88
	SettingEmissiveColor                        SettingsUsage = 0xc985395a
	SettingEmissiveColorA                       SettingsUsage = 0xae6e0ea9
	SettingEmissiveInnerExp                     SettingsUsage = 0x80b27021
	SettingEmissiveIntensity                    SettingsUsage = 0x02f6dc5b
	SettingEmissiveMult                         SettingsUsage = 0xde7b71ec
	SettingEmissiveOuterExp                     SettingsUsage = 0x1ae042c5
	SettingEmissiveStrength                     SettingsUsage = 0x7fa622b1
	SettingEmissiveUVDirection                  SettingsUsage = 0x3412359b
	SettingEmissiveWaveGradient                 SettingsUsage = 0x3de979ac
	SettingEmissiveWaveSize                     SettingsUsage = 0xd0a0e0fc
	SettingEmissiveWaveSpeed                    SettingsUsage = 0x14f7ceac
	SettingEndFadeExp                           SettingsUsage = 0x804cf93e
	SettingEndFadeTightness                     SettingsUsage = 0xf3578894
	SettingEndTaper                             SettingsUsage = 0xcc6f86ce
	SettingEnvLumMin                            SettingsUsage = 0xe9374c77
	SettingEnvLumMin01                          SettingsUsage = 0x5015885c
	SettingErodeMult                            SettingsUsage = 0x8d2e1980
	SettingErodeSoftness                        SettingsUsage = 0xa24cabf6
	SettingExposure                             SettingsUsage = 0x130e95bc
	SettingExposure280                          SettingsUsage = 0x0b049d4b
	SettingFadeDepth                            SettingsUsage = 0x2fb59e3e
	SettingFadeInOutType                        SettingsUsage = 0xf2fa4ed9
	SettingFadeYAngle                           SettingsUsage = 0xf5963b7d
	SettingFarScatterDensity                    SettingsUsage = 0x94dc0457
	SettingFarScatterNormalIntensityMult        SettingsUsage = 0xceb3755d
	SettingFlareNoiseSpeed                      SettingsUsage = 0x0b8f99c3
	SettingFlareNoiseSpeed02                    SettingsUsage = 0xee6704ff
	SettingFlareNoiseTile                       SettingsUsage = 0x8c1ac990
	SettingFlareNoiseTile02                     SettingsUsage = 0x58c5be9b
	SettingFlareTexExp                          SettingsUsage = 0x0e5e69a7
	SettingFlickerMin                           SettingsUsage = 0xfa750cb9
	SettingFlickerSpd                           SettingsUsage = 0x8e90c583
	SettingFogAmbientDuringTransitionColorBoost SettingsUsage = 0x9f9696c9
	SettingFogBackscatterLerp                   SettingsUsage = 0x5bcce04d
	SettingFogBackscatterPhase                  SettingsUsage = 0x064b4686
	SettingFogColor                             SettingsUsage = 0x67be31ec
	SettingFogColorHax                          SettingsUsage = 0xc5346f76
	SettingFogDustiness                         SettingsUsage = 0x435c9293
	SettingFogEnabled                           SettingsUsage = 0xe5880e06
	SettingFogForwardscatterPhase               SettingsUsage = 0xf8a6f098
	SettingFogIntesity                          SettingsUsage = 0xc7794dfe
	SettingFogLightAmbientIntensity             SettingsUsage = 0xac90a8a0
	SettingFogLightPollution                    SettingsUsage = 0xd320b152
	SettingFogParameters                        SettingsUsage = 0x07379e86
	SettingFogShadowIntensity                   SettingsUsage = 0x81a36aad
	SettingFogSunIntensity                      SettingsUsage = 0x9073b7d2
	SettingFogVolumeAlbedoIntensity             SettingsUsage = 0xfabdd1ee
	SettingFogVolumeAlbedoLerp                  SettingsUsage = 0x45e848be
	SettingFogVolumeColor                       SettingsUsage = 0xc449e020
	SettingFogVolumeDensity                     SettingsUsage = 0x988411cf
	SettingFogVolumeDustiness                   SettingsUsage = 0x7a2d7658
	SettingFogVolumeFalloffPow                  SettingsUsage = 0x22111de6
	SettingFogVolumeHeight                      SettingsUsage = 0x83495529
	SettingFrameNumber                          SettingsUsage = 0xc67a00e3
	SettingFrames                               SettingsUsage = 0xbb0b6743
	SettingFresnel                              SettingsUsage = 0x8bcc0bf6
	SettingFresnelDistance                      SettingsUsage = 0xabe80b64
	SettingFresnelEdges                         SettingsUsage = 0x92379423
	SettingFresnelEdgesMult                     SettingsUsage = 0xa58de6a7
	SettingFresnelExp                           SettingsUsage = 0xfba06af8
	SettingFresnelExpMax                        SettingsUsage = 0x8c05c79d
	SettingFresnelInterior                      SettingsUsage = 0xba27dc03
	SettingFresnelMin                           SettingsUsage = 0x13eebc56
	SettingFresnelMult                          SettingsUsage = 0x959540a3
	SettingFrostWeight                          SettingsUsage = 0xe4d9b883
	SettingGalaxyScaleAlignment                 SettingsUsage = 0xf54ad626
	SettingGalaxyThickness                      SettingsUsage = 0x637447ff
	SettingGlintAmount                          SettingsUsage = 0x377ff51a
	SettingGlintIntensity                       SettingsUsage = 0xbabb70cf
	SettingGlintRoughness                       SettingsUsage = 0x397a34f5
	SettingGlintSize                            SettingsUsage = 0x4ec3a4d1
	SettingGlitchAmount                         SettingsUsage = 0x4408549c
	SettingGlitchCenterOffset                   SettingsUsage = 0xf85d2474
	SettingGlitchColor                          SettingsUsage = 0x80b10f6d
	SettingGlitchColorPower                     SettingsUsage = 0xa85d996a
	SettingGlitchGridsize                       SettingsUsage = 0x6fd56a28
	SettingGlitchSpeed                          SettingsUsage = 0x88b1a3e2
	SettingGlobalDetailTile                     SettingsUsage = 0x47cc2eff
	SettingGlobalDiffuseMap                     SettingsUsage = 0xdb7e5380
	SettingGlobalSurfaceTile                    SettingsUsage = 0xaa189722
	SettingGlobalViewport                       SettingsUsage = 0x516d5ccd
	SettingGlowContrast                         SettingsUsage = 0x550c15a0
	SettingGlowIntensity                        SettingsUsage = 0xd0140e79
	SettingGlowOffset                           SettingsUsage = 0x657c095b
	SettingGlowTemperature                      SettingsUsage = 0xabf3dfb9
	SettingGradientColor01                      SettingsUsage = 0xd7711784
	SettingGradientColor02                      SettingsUsage = 0x20442c96
	SettingGradientColorExp                     SettingsUsage = 0xc11e1b5e
	SettingGradientColorMult                    SettingsUsage = 0xe4127134
	SettingGradientDirtMaxheight                SettingsUsage = 0x1cf7cbfe
	SettingGradientDirtMinheight                SettingsUsage = 0x2a450f19
	SettingGradientExp                          SettingsUsage = 0xa377ca1e
	SettingGradientMult                         SettingsUsage = 0x17941863
	SettingGradientSubtractExp                  SettingsUsage = 0xb3ca37d0
	SettingGradingGroupId                       SettingsUsage = 0x8499a8cd
	SettingGradingGroupIdGround                 SettingsUsage = 0x41090357
	SettingGradingGroupIdMaskedDetails          SettingsUsage = 0xe954a6e2
	SettingGradingGroupIdRock                   SettingsUsage = 0x2e7c6c09
	SettingGradingGroupIdSecondaryColor         SettingsUsage = 0x160acf12
	SettingGradingGroupIdSecondworld            SettingsUsage = 0x212ff30b
	SettingGradingGroupIdThirdworld             SettingsUsage = 0x87d87e4c
	SettingGradingGroupIdTrunk                  SettingsUsage = 0x4eab5256
	SettingGradingGroupIdVistaDetail            SettingsUsage = 0xb1330d2a
	SettingGradingGroupIdWeathering             SettingsUsage = 0x1b9bb2b7
	SettingGradingSecondaryGroupId              SettingsUsage = 0x5b4b896b
	SettingGrainSpeed                           SettingsUsage = 0x06e7453a
	SettingGreyscale                            SettingsUsage = 0xe1823edc
	SettingGunkNormalFade                       SettingsUsage = 0x7817ece3
	SettingGunkScale                            SettingsUsage = 0xac339e75
	SettingHeightContrast                       SettingsUsage = 0x538d72ec
	SettingHeightWetnessAndWash                 SettingsUsage = 0xa287bad0
	SettingHeightmapNormals                     SettingsUsage = 0x7f802a89
	SettingHitExp                               SettingsUsage = 0x5efce576
	SettingHitRExp                              SettingsUsage = 0x299fac0b
	SettingHitRMult                             SettingsUsage = 0x284335f8
	SettingHmapSize                             SettingsUsage = 0x590c7106
	SettingHologramColor                        SettingsUsage = 0xe461ccf3
	SettingHologramColor01                      SettingsUsage = 0xc5ac0e90
	SettingHologramHideAmount                   SettingsUsage = 0x358de6ea
	SettingHologramPlanetLightColors            SettingsUsage = 0x299158dc
	SettingHologramPlanetPositions              SettingsUsage = 0x010689ba
	SettingHologramPlanetScaleMultiplier        SettingsUsage = 0x4de5849b
	SettingHudCurveAmount                       SettingsUsage = 0x125717dc
	SettingIEnd                                 SettingsUsage = 0x3de4a739
	SettingIIntensity                           SettingsUsage = 0xfa2f1a66
	SettingIStart                               SettingsUsage = 0x18ad986b
	SettingIThickness                           SettingsUsage = 0x7d78bc54
	SettingIceFuzzColor                         SettingsUsage = 0x96e9fc91
	SettingIceFuzzIntensity                     SettingsUsage = 0xb02309c8
	SettingIceSubsurfaceColor                   SettingsUsage = 0x652404cc
	SettingIceSubsurfaceDiffusion               SettingsUsage = 0xb80bcee2
	SettingIceSubsurfaceIntensity               SettingsUsage = 0x9bbd9167
	SettingIceSubsurfaceThickness               SettingsUsage = 0x5aabc186
	SettingIceSubsurfaceWrap                    SettingsUsage = 0x2422694a
	SettingIceTint                              SettingsUsage = 0xad15dc03
	SettingIceWarp                              SettingsUsage = 0x3eceae3c
	SettingIesLookup                            SettingsUsage = 0x2d03d53b
	SettingIgnoreParticlecolor                  SettingsUsage = 0xd1b664db
	SettingImpTransparentOverride               SettingsUsage = 0x795a98fb
	SettingInitialSpawnPos                      SettingsUsage = 0xa3771953
	SettingInstanceSeed                         SettingsUsage = 0xc155bc0d
	SettingInstancingZero                       SettingsUsage = 0x3bf0bf86
	SettingIntensity                            SettingsUsage = 0x32f447e5
	SettingInterserctionBrightness              SettingsUsage = 0x5d38f538
	SettingInterserctionExp                     SettingsUsage = 0xdeb3e257
	SettingInterserctionThickness               SettingsUsage = 0x765398ae
	SettingInvHmapSize                          SettingsUsage = 0x5d3d3914
	SettingInvView                              SettingsUsage = 0x35bfe42f
	SettingInvViewProj                          SettingsUsage = 0x7751b420
	SettingInvWorld                             SettingsUsage = 0xb31b34db
	SettingInvertFresnel                        SettingsUsage = 0x9d3bc7e1
	SettingIoffset                              SettingsUsage = 0x774a112b
	SettingJacobiFalloff                        SettingsUsage = 0x6a9c1c61
	SettingLastWorld                            SettingsUsage = 0xb8dadd64
	SettingLavaContrast                         SettingsUsage = 0xbac2d390
	SettingLavaOffset                           SettingsUsage = 0x7e7efe2f
	SettingLavaTemperature                      SettingsUsage = 0x557ef4c8
	SettingLensColor                            SettingsUsage = 0xa08c4aac
	SettingLensCutoutEnabled                    SettingsUsage = 0x92a26701
	SettingLensEmissiveColor                    SettingsUsage = 0xc92c1af4
	SettingLensEmissiveIntensity                SettingsUsage = 0xf4ed50e6
	SettingLensEmissiveOpacity                  SettingsUsage = 0xba57948e
	SettingLensEmissiveTexture                  SettingsUsage = 0x1c121028
	SettingLensOcclusionEnabled                 SettingsUsage = 0x7b23bc15
	SettingLensOcclusionSize                    SettingsUsage = 0xba2f978f
	SettingLensOcclusionTexture                 SettingsUsage = 0x1c8c3930
	SettingLensOffset                           SettingsUsage = 0x269d4c7e
	SettingLensOpacityMul                       SettingsUsage = 0x9c94aff8
	SettingLensParallaxMult                     SettingsUsage = 0xe51619c9
	SettingLensScale                            SettingsUsage = 0xeece6aa3
	SettingLifetimeDrawMult                     SettingsUsage = 0x3aef5343
	SettingLifetimeExp                          SettingsUsage = 0x062bde33
	SettingLightProbeSpaceSpecular              SettingsUsage = 0x8f444022
	SettingLightingData                         SettingsUsage = 0x7ae8af7d
	SettingLightsourceAngularSize               SettingsUsage = 0x4fe73a30
	SettingLinearFadeOffsets                    SettingsUsage = 0xa66c75e2
	SettingLocalLightsShadowAtlas               SettingsUsage = 0x48c88f1d
	SettingLodCameraPos                         SettingsUsage = 0x862904d2
	SettingLodFadeLevel                         SettingsUsage = 0x8b8e7a4d
	SettingLookupPhaseSpeed                     SettingsUsage = 0xcc7d1e0c
	SettingLookupWeight                         SettingsUsage = 0x73e06d8d
	SettingLumMin                               SettingsUsage = 0xa35bc4f7
	SettingLumMinRemap                          SettingsUsage = 0x0492a749
	SettingLuminocityExp                        SettingsUsage = 0x2ca054d5
	SettingLuminosityOpacity                    SettingsUsage = 0xf7345fe3
	SettingLutContrast                          SettingsUsage = 0xf66babdb
	SettingLutMixBiomeA                         SettingsUsage = 0x3befcb4d
	SettingLutMixBiomeB                         SettingsUsage = 0x60179afe
	SettingLutOffset                            SettingsUsage = 0x8fd1d21a
	SettingMaskScale                            SettingsUsage = 0x65777b10
	SettingMaskSharpnessBiome                   SettingsUsage = 0xb2198f28
	SettingMaskSharpnessLut                     SettingsUsage = 0x42c1f33f
	SettingMaskVariation                        SettingsUsage = 0xeaf7e76b
	SettingMaterial01TileMultiplier             SettingsUsage = 0xc644acfd
	SettingMaterial02TileMultiplier             SettingsUsage = 0x96e6e3da
	SettingMaterial03TileMultiplier             SettingsUsage = 0xc0a64f07
	SettingMaterial04TileMultiplier             SettingsUsage = 0xd3cd4d57
	SettingMaterial05TileMultiplier             SettingsUsage = 0xa0efa5e1
	SettingMaterial06TileMultiplier             SettingsUsage = 0x43dd57e1
	SettingMaterial07TileMultiplier             SettingsUsage = 0x11d6e86d
	SettingMaterial08TileMultiplier             SettingsUsage = 0x71a943cc
	SettingMaterial1Metallic                    SettingsUsage = 0xaf0cc9eb
	SettingMaterial1RoughnessBase               SettingsUsage = 0xdaeb0bb8
	SettingMaterial1RoughnessBuildUp            SettingsUsage = 0x90a9f367
	SettingMaterial1Surface                     SettingsUsage = 0xdf83e2df
	SettingMaterial1SurfaceNormal               SettingsUsage = 0xf3f210cc
	SettingMaterial1SurfaceRoughness            SettingsUsage = 0x711f20c0
	SettingMaterial1SurfaceValue                SettingsUsage = 0x6c572070
	SettingMaterial1WearCavityEdge              SettingsUsage = 0xf20147fa
	SettingMaterial2Metallic                    SettingsUsage = 0xc9a20fea
	SettingMaterial2RoughnessBase               SettingsUsage = 0x643edc95
	SettingMaterial2RoughnessBuildUp            SettingsUsage = 0x2fc67a7f
	SettingMaterial2Surface                     SettingsUsage = 0x700dcf36
	SettingMaterial2SurfaceNormal               SettingsUsage = 0x17d76d59
	SettingMaterial2SurfaceRoughness            SettingsUsage = 0x180c0f64
	SettingMaterial2SurfaceValue                SettingsUsage = 0xe8e2df96
	SettingMaterial2WearCavityEdge              SettingsUsage = 0x4601a106
	SettingMaterial2WearCavityEdge01            SettingsUsage = 0x475b1de3
	SettingMaterial2WearCavityEdge06            SettingsUsage = 0x40c4aa65
	SettingMaterial3Metallic                    SettingsUsage = 0x6faeaab2
	SettingMaterial3RoughnessBase               SettingsUsage = 0x3b497b46
	SettingMaterial3RoughnessBuildUp            SettingsUsage = 0x6b2265c1
	SettingMaterial3Surface                     SettingsUsage = 0xc4c6e576
	SettingMaterial3SurfaceNormal               SettingsUsage = 0xfef01a44
	SettingMaterial3SurfaceRoughness            SettingsUsage = 0x460c3989
	SettingMaterial3SurfaceValue                SettingsUsage = 0x24443f71
	SettingMaterial4Metallic                    SettingsUsage = 0x626a0dea
	SettingMaterial4RoughnessBase               SettingsUsage = 0xc53eefd7
	SettingMaterial4RoughnessBuildUp            SettingsUsage = 0xe8d3eb2b
	SettingMaterial4Surface                     SettingsUsage = 0x5b49d65d
	SettingMaterial4SurfaceNormal               SettingsUsage = 0xd6003096
	SettingMaterial4SurfaceRoughness            SettingsUsage = 0xfbeccc0b
	SettingMaterial4SurfaceValue                SettingsUsage = 0x68df1c10
	SettingMaterial4WearCavityEdge              SettingsUsage = 0xa9fdbe64
	SettingMaterial5Metallic                    SettingsUsage = 0x910e7a65
	SettingMaterial5RoughnessBase               SettingsUsage = 0xca77a10e
	SettingMaterial5RoughnessBuildUp            SettingsUsage = 0x098114da
	SettingMaterial5Surface                     SettingsUsage = 0x26f326b5
	SettingMaterial5SurfaceNormal               SettingsUsage = 0x78780de6
	SettingMaterial5SurfaceRoughness            SettingsUsage = 0x03953944
	SettingMaterial5SurfaceValue                SettingsUsage = 0x72c7e9c3
	SettingMaterial5WearCavityEdge              SettingsUsage = 0xf93a43ad
	SettingMaterial6Metallic                    SettingsUsage = 0xd2f08ded
	SettingMaterial6RoughnessBase               SettingsUsage = 0x5afe729b
	SettingMaterial6RoughnessBuildUp            SettingsUsage = 0x547e1075
	SettingMaterial6Surface                     SettingsUsage = 0xee28bf90
	SettingMaterial6SurfaceNormal               SettingsUsage = 0x4d1c3d67
	SettingMaterial6SurfaceRoughness            SettingsUsage = 0xc99c1acd
	SettingMaterial6SurfaceValue                SettingsUsage = 0x2f209ca7
	SettingMaterial6WearCavityEdge              SettingsUsage = 0xe740c470
	SettingMaterial7Metallic                    SettingsUsage = 0x45ded6a2
	SettingMaterial7RoughnessBase               SettingsUsage = 0x253d7000
	SettingMaterial7RoughnessBuildUp            SettingsUsage = 0x4e1163ac
	SettingMaterial7Surface                     SettingsUsage = 0x79c17d85
	SettingMaterial7SurfaceNormal               SettingsUsage = 0xa070cbbb
	SettingMaterial7SurfaceRoughness            SettingsUsage = 0xc8ae0472
	SettingMaterial7SurfaceValue                SettingsUsage = 0x8f9030d5
	SettingMaterial7WearCavityEdge              SettingsUsage = 0x47330096
	SettingMaterial8Metallic                    SettingsUsage = 0x42af02e2
	SettingMaterial8RoughnessBase               SettingsUsage = 0xc44740ce
	SettingMaterial8RoughnessBuildUp            SettingsUsage = 0xd8e6fd76
	SettingMaterial8Surface                     SettingsUsage = 0xc39b826d
	SettingMaterial8SurfaceNormal               SettingsUsage = 0x1d2ababe
	SettingMaterial8SurfaceRoughness            SettingsUsage = 0x9c15fe93
	SettingMaterial8SurfaceValue                SettingsUsage = 0x7d2a970e
	SettingMaterialIndex                        SettingsUsage = 0x51764e98
	SettingMaterialVariable                     SettingsUsage = 0x70733edd
	SettingMaterialWetness                      SettingsUsage = 0x2f77b77e
	SettingMaxEmissive                          SettingsUsage = 0xc09ffbfb
	SettingMetalic                              SettingsUsage = 0x61553d50
	SettingMetallic01                           SettingsUsage = 0x11941d3a
	SettingMetallicOpacity                      SettingsUsage = 0xcd72af93
	SettingMicroAo                              SettingsUsage = 0xb73551ec
	SettingMicroAoIntensity                     SettingsUsage = 0x6a0b8014
	SettingMieBeta                              SettingsUsage = 0x3e5c7796
	SettingMieHeight                            SettingsUsage = 0x985c4670
	SettingMieTintHax                           SettingsUsage = 0xb013d3df
	SettingMinEmissive                          SettingsUsage = 0x0e938de6
	SettingMultiply                             SettingsUsage = 0x0a32bdf8
	SettingNoise01Channel                       SettingsUsage = 0xb0f6a5c4
	SettingNoise01Exp                           SettingsUsage = 0x69da8139
	SettingNoise01ExpMax                        SettingsUsage = 0xe53aed1d
	SettingNoise01ExpMin                        SettingsUsage = 0x86044187
	SettingNoise01Minmax                        SettingsUsage = 0x764f9cef
	SettingNoise01Speed                         SettingsUsage = 0x3f591185
	SettingNoise01Tile                          SettingsUsage = 0x95da4b25
	SettingNoise02Channel                       SettingsUsage = 0x2ff13575
	SettingNoise02Exp                           SettingsUsage = 0xa3b1ad25
	SettingNoise02Minmax                        SettingsUsage = 0x78c223b4
	SettingNoise02Speed                         SettingsUsage = 0xb961d0c7
	SettingNoise02Tile                          SettingsUsage = 0x0dfd4183
	SettingNoiseChannel                         SettingsUsage = 0x19aa33b6
	SettingNoiseChannel02                       SettingsUsage = 0x9d5fd779
	SettingNoiseExp                             SettingsUsage = 0xdf914ea5
	SettingNoiseMax                             SettingsUsage = 0x9b5dbb91
	SettingNoiseMin                             SettingsUsage = 0x16f25bca
	SettingNoiseMult                            SettingsUsage = 0xc6b79ab1
	SettingNoiseOffset                          SettingsUsage = 0x5727dc8a
	SettingNoiseScale                           SettingsUsage = 0x53d111f6
	SettingNoiseStrength                        SettingsUsage = 0x7bb4b328
	SettingNormalCcnormOpacity                  SettingsUsage = 0x53a2e583
	SettingNormalIntensity                      SettingsUsage = 0xb01fb18d
	SettingNormalIntensityBiomeA                SettingsUsage = 0xfc6ed011
	SettingNormalIntensityBiomeB                SettingsUsage = 0x2b8f220a
	SettingNormalIntensityGround                SettingsUsage = 0x1af055be
	SettingNormalIntensityVistaDetail           SettingsUsage = 0xea3791b9
	SettingNormalMirroring                      SettingsUsage = 0xeb55ac15
	SettingNormalOpacity                        SettingsUsage = 0xc572595b
	SettingNormalOverLife                       SettingsUsage = 0x198c16b4
	SettingNormalPullUp                         SettingsUsage = 0x7dda1c04
	SettingNormals                              SettingsUsage = 0x080b4d6f
	SettingNrmStr                               SettingsUsage = 0xbe8a0ddb
	SettingNumTiles                             SettingsUsage = 0xdc03e5b6
	SettingOffsetExp                            SettingsUsage = 0x7c53ba41
	SettingOffsetFlashExp                       SettingsUsage = 0x3345a139
	SettingOffsetFlashMult                      SettingsUsage = 0x60842ae2
	SettingOffsetMax                            SettingsUsage = 0xdd2d6e43
	SettingOffsetMin                            SettingsUsage = 0x82d956a6
	SettingOffsetMult                           SettingsUsage = 0x21904b30
	SettingOpacityOffset                        SettingsUsage = 0xccddd706
	SettingOpacitySharpness                     SettingsUsage = 0xcb8c5728
	SettingOpacityThreshold                     SettingsUsage = 0x529a4aaf
	SettingOpacityThresholdFar                  SettingsUsage = 0xe8b5732a
	SettingOpacityTreshholdFadeDistanceInv      SettingsUsage = 0x76a39cf8
	SettingOverlayMaskAmount                    SettingsUsage = 0x9b2f7128
	SettingPaletteSlot                          SettingsUsage = 0xf6dc872e
	SettingParallaxBias                         SettingsUsage = 0x9b56d88b
	SettingParallaxIntensity                    SettingsUsage = 0x965ad445
	SettingParallaxIntensityCloud               SettingsUsage = 0x534a9352
	SettingParallaxScale                        SettingsUsage = 0x12e9c61b
	SettingParticleAgeLife                      SettingsUsage = 0x686cdd17
	SettingParticleColorOnly                    SettingsUsage = 0x8bdc17ff
	SettingPlanetRoughnessIntesity              SettingsUsage = 0xbb2198cd
	SettingPlanetScaleMult                      SettingsUsage = 0xb7f77596
	SettingPlanetShadowDistanceFalloff          SettingsUsage = 0x291104de
	SettingPlanetWpPos                          SettingsUsage = 0xc6af80ca
	SettingPlanetWpScale                        SettingsUsage = 0x2dc4052a
	SettingPointTaper                           SettingsUsage = 0xc5b5e1bd
	SettingPostEffectsEnabled                   SettingsUsage = 0x954d97d8
	SettingPower                                SettingsUsage = 0x21e1cde6
	SettingPreventTerrainDeformation            SettingsUsage = 0xc012efe1
	SettingPreventsTerrainDeformation           SettingsUsage = 0xcfe823b0
	SettingProj                                 SettingsUsage = 0xe5366cab
	SettingRampDownExpHacky                     SettingsUsage = 0xe9cdbcc7
	SettingRawNonCheckerboardedTargetSize       SettingsUsage = 0xdc7ec5d4
	SettingRawNonCheckerboardedViewport         SettingsUsage = 0xde860c50
	SettingRayleighBeta                         SettingsUsage = 0xbc7e6abb
	SettingRayleighBetaSpaceplanet              SettingsUsage = 0x9bfa2046
	SettingRegionPositionOffset                 SettingsUsage = 0x549d02af
	SettingRemapMinValue                        SettingsUsage = 0x82aa6dd0
	SettingResolutionSetting                    SettingsUsage = 0x43695f7b
	SettingReticleColor                         SettingsUsage = 0xfa2ceef2
	SettingReticleColorIntensity                SettingsUsage = 0xd177ec1f
	SettingReticleOpacity                       SettingsUsage = 0x21ae1ffc
	SettingReticleTexture                       SettingsUsage = 0xbdc07044
	SettingReticuleScale                        SettingsUsage = 0x8915dbd3
	SettingRiverAmount                          SettingsUsage = 0x4b8e8949
	SettingRiverSmooth                          SettingsUsage = 0xedd856cf
	SettingRoughness                            SettingsUsage = 0xea2298b5
	SettingRoughnessMulti                       SettingsUsage = 0xe3847c92
	SettingRoughnessOpacity                     SettingsUsage = 0x57a7bab2
	SettingSampleTerrainAlbedo                  SettingsUsage = 0xc89c53ec
	SettingScalarFieldCutoff                    SettingsUsage = 0x87a3226d
	SettingScanSpeed                            SettingsUsage = 0xfb60747e
	SettingScanline01Intensity                  SettingsUsage = 0x90ff9fe3
	SettingScanline02Intensity                  SettingsUsage = 0xad71ebb1
	SettingScanlineCount                        SettingsUsage = 0x9701870a
	SettingScanlineCount02                      SettingsUsage = 0xf8469f32
	SettingScanlineDistScale                    SettingsUsage = 0x67043721
	SettingScanlineThickness                    SettingsUsage = 0x27c9cffb
	SettingScanlineThickness02                  SettingsUsage = 0x8bc3a6cb
	SettingSecondOpacityInfluence               SettingsUsage = 0xe24dd6c8
	SettingSecondOpacityMin                     SettingsUsage = 0x7b103e3c
	SettingSecondOpacityStart                   SettingsUsage = 0xc07ff215
	SettingSelectedOffset                       SettingsUsage = 0x85c1a314
	SettingSelectedScale                        SettingsUsage = 0xfe83d8dc
	SettingSelfEmissiveColor                    SettingsUsage = 0xabc2e1d1
	SettingSelfEmissiveIntensity                SettingsUsage = 0xa4ef22dd
	SettingSelfPlanetIndex                      SettingsUsage = 0xb480e8e2
	SettingShadowBiasSlice0                     SettingsUsage = 0x72f4e544
	SettingShadowBiasSlice1                     SettingsUsage = 0xa98b571d
	SettingShadowBiasSlice2                     SettingsUsage = 0xc7f745fd
	SettingShadowBiasSlice3                     SettingsUsage = 0x6b1f0050
	SettingShadowClampToNearPlane               SettingsUsage = 0xf682b26b
	SettingShadowDepthBiasSlice0                SettingsUsage = 0xe6df7d8f
	SettingShadowDepthBiasSlice1                SettingsUsage = 0xd411ac66
	SettingShadowDepthBiasSlice2                SettingsUsage = 0xc657a252
	SettingShadowDepthBiasSlice3                SettingsUsage = 0xaae90095
	SettingShadowIntensity                      SettingsUsage = 0x551edfd1
	SettingShadowRotation                       SettingsUsage = 0xbce384d7
	SettingShadowScaleSlice0                    SettingsUsage = 0x3a07880e
	SettingShadowScaleSlice1                    SettingsUsage = 0x8fc1ae83
	SettingShadowScaleSlice2                    SettingsUsage = 0x6c70465f
	SettingShadowScaleSlice3                    SettingsUsage = 0x65d03e76
	SettingShadowsCasting                       SettingsUsage = 0x4a20cc26
	SettingShallowWaterColor                    SettingsUsage = 0x4ce36e31
	SettingShieldId                             SettingsUsage = 0xce2d8adf
	SettingSkipFacing                           SettingsUsage = 0x4e6ecddb
	SettingSnowAmount                           SettingsUsage = 0xd8686aae
	SettingSnowFromBottom                       SettingsUsage = 0xc840fedf
	SettingSnowFromHeight                       SettingsUsage = 0x9c70de00
	SettingSnowFromNormal                       SettingsUsage = 0xae31a7b7
	SettingSnowFromTop                          SettingsUsage = 0xadba2550
	SettingSnowHardnessBottom                   SettingsUsage = 0xaf445998
	SettingSnowHardnessTop                      SettingsUsage = 0x018a050f
	SettingSnowIndex                            SettingsUsage = 0x016ba55d
	SettingSnowIndex0                           SettingsUsage = 0xc4dd3513
	SettingSnowMask                             SettingsUsage = 0x5dee8603
	SettingSnowMaskDepth                        SettingsUsage = 0xf05c45d6
	SettingSnowMaskHardness                     SettingsUsage = 0xe15c44a6
	SettingSnowMaskTile                         SettingsUsage = 0xbb500d15
	SettingSnowNormalDisplacement               SettingsUsage = 0xb4a7f867
	SettingSnowNormalIntensity                  SettingsUsage = 0xe9b95c8d
	SettingSnowNormalIntensity01                SettingsUsage = 0x8696ab8f
	SettingSnowNormalMask                       SettingsUsage = 0x034f906a
	SettingSnowPerZoneDisabling                 SettingsUsage = 0x0a3201b6
	SettingSnowSsThickness                      SettingsUsage = 0xb66f02b6
	SettingSnowTile                             SettingsUsage = 0xed8bddbe
	SettingSnowTile01                           SettingsUsage = 0x76a3727d
	SettingSnowTrampleNormalBlur                SettingsUsage = 0x408666ca
	SettingSnowTrampleNormalIntensity           SettingsUsage = 0xdb7dbe33
	SettingSnowUpDisplacement                   SettingsUsage = 0x8399b37b
	SettingSpecialGunkColor                     SettingsUsage = 0x300e404d
	SettingSpecular                             SettingsUsage = 0xb043861b
	SettingSpecularBrdfLut                      SettingsUsage = 0x6a52658f
	SettingSpecularCurve                        SettingsUsage = 0xd4f0d374
	SettingSpecularIntensity                    SettingsUsage = 0xb4888905
	SettingSpecularMulti                        SettingsUsage = 0xfc3a0cce
	SettingSpeed                                SettingsUsage = 0x2c1c82c8
	SettingSsDiffusion                          SettingsUsage = 0x68c8ea95
	SettingSsIntensityMult                      SettingsUsage = 0xccd376a1
	SettingSsThickness                          SettingsUsage = 0xeb6d5aa8
	SettingSssIntensity                         SettingsUsage = 0xf5c7ccd2
	SettingSssWrap                              SettingsUsage = 0xba2ffb95
	SettingStarColor                            SettingsUsage = 0x40582607
	SettingStarOpacities                        SettingsUsage = 0x7abc9146
	SettingStarTilingAmount                     SettingsUsage = 0xb8d05848
	SettingSubsurfaceDiff                       SettingsUsage = 0xc32c22cb
	SettingSubsurfaceDiffusion                  SettingsUsage = 0xd43cfc8b
	SettingSubsurfaceFadeCurve                  SettingsUsage = 0x519aa783
	SettingSubsurfaceFadeDistance               SettingsUsage = 0x460fa0f8
	SettingSubsurfaceInt                        SettingsUsage = 0x7ed050fb
	SettingSubsurfaceIntensity                  SettingsUsage = 0x38dae7bd
	SettingSubsurfaceOcclusionIntensity         SettingsUsage = 0xe7d8cbf8
	SettingSubsurfaceOcclusionMaskcurve         SettingsUsage = 0x64ca4a51
	SettingSubsurfaceThickness                  SettingsUsage = 0x15e00526
	SettingSubsurfaceTint                       SettingsUsage = 0x304b6073
	SettingSubsurfaceWrap                       SettingsUsage = 0x2c9d0736
	SettingSunAngularSize                       SettingsUsage = 0xeb1b1369
	SettingSunCol                               SettingsUsage = 0x0a926caa
	SettingSunColor                             SettingsUsage = 0x40027a64
	SettingSunColorOverride                     SettingsUsage = 0x51c4d38c
	SettingSunDir                               SettingsUsage = 0xdf33e43d
	SettingSunDirection                         SettingsUsage = 0x93b47324
	SettingSunDirectionOverride                 SettingsUsage = 0x32829be8
	SettingSunEnabled                           SettingsUsage = 0x004d7226
	SettingSunShadowMap                         SettingsUsage = 0xb85584a2
	SettingSunShadowsEnabled                    SettingsUsage = 0x97c21e94
	SettingTaaEnabled                           SettingsUsage = 0x5923ab38
	SettingTaperExp                             SettingsUsage = 0xa12e9fe8
	SettingTearAmount                           SettingsUsage = 0xf299681b
	SettingTerrainBlendCurve                    SettingsUsage = 0x5920e53e
	SettingTerrainBlendHeight                   SettingsUsage = 0x16180767
	SettingTerrainColorLerp                     SettingsUsage = 0x74e52404
	SettingTerrainPatchSize                     SettingsUsage = 0x725e623e
	SettingTerrainSize                          SettingsUsage = 0xaa464213
	SettingTerrainTiling                        SettingsUsage = 0x2bb8e7a6
	SettingTest01                               SettingsUsage = 0x7a3a6f92
	SettingTest02                               SettingsUsage = 0xd707d37a
	SettingTest03                               SettingsUsage = 0xc167182b
	SettingTexIsGreyscale                       SettingsUsage = 0xa9b3eb2d
	SettingTextureDensityVisualization          SettingsUsage = 0x63629748
	SettingTileFactor                           SettingsUsage = 0x4c8f6ef3
	SettingTileFactorCloudsEdges                SettingsUsage = 0x533d5194
	SettingTileFactorCloudsInside               SettingsUsage = 0x547b5d08
	SettingTileFactorNoise                      SettingsUsage = 0x172ff707
	SettingTileSize                             SettingsUsage = 0x8dd38245
	SettingTilerAoIntensity                     SettingsUsage = 0xd90d2e8c
	SettingTilerNormalIntensity                 SettingsUsage = 0x90dd5645
	SettingTiling                               SettingsUsage = 0xf8df3d62
	SettingTilingFactor                         SettingsUsage = 0xa8f68dd2
	SettingTilingGround                         SettingsUsage = 0xeb121575
	SettingTilingVistaDetail                    SettingsUsage = 0xf5b036a5
	SettingTime                                 SettingsUsage = 0x7b9caf6a
	SettingTimeAdd                              SettingsUsage = 0xd90fd931
	SettingTimeMult                             SettingsUsage = 0x0bb5e2d9
	SettingTimeMultiplier                       SettingsUsage = 0x68dee0fc
	SettingTimeOfDayOverridesEnabled            SettingsUsage = 0x1f3d37a9
	SettingTint                                 SettingsUsage = 0xf0d16ec3
	SettingTint01                               SettingsUsage = 0x38e14032
	SettingTint02                               SettingsUsage = 0x3a89b151
	SettingTintMaskColor                        SettingsUsage = 0xc3befbff
	SettingTipGradientRemap                     SettingsUsage = 0x36727284
	SettingTonemapA                             SettingsUsage = 0x35454225
	SettingTonemapB                             SettingsUsage = 0x9f1df561
	SettingTonemapC                             SettingsUsage = 0x338fb230
	SettingTonemapD                             SettingsUsage = 0xc692eaee
	SettingTonemapE                             SettingsUsage = 0x65cfa98c
	SettingTopColorTint                         SettingsUsage = 0x4f18baf8
	SettingTopDirtValue                         SettingsUsage = 0xc1eae640
	SettingTrampleAmount                        SettingsUsage = 0x0e76b081
	SettingTrampleFootstepMipBlur               SettingsUsage = 0x08fe34f0
	SettingTrampleFootstepNormalWeight          SettingsUsage = 0x2bc0d12c
	SettingTrampleMakeWetWeight                 SettingsUsage = 0xd061d671
	SettingTrampleMiscWeight                    SettingsUsage = 0xa8f87c34
	SettingTrampleNormalWeight                  SettingsUsage = 0x326b7f53
	SettingTramplePuddleWeight                  SettingsUsage = 0x80955bf8
	SettingTrampleSnowIndex                     SettingsUsage = 0x672627e1
	SettingTreeDepthScale                       SettingsUsage = 0x2f9388b5
	SettingTriplanarDetailTiler                 SettingsUsage = 0xc2e279d7
	SettingTriplanarDetailTiler0                SettingsUsage = 0x9358daf0
	SettingTriplanarDetailTiler1                SettingsUsage = 0x59536f91
	SettingTriplanarMaskSharpen                 SettingsUsage = 0x0c9a0b54
	SettingTriplanarNormalIntensity             SettingsUsage = 0xe2256f0a
	SettingTriplanarTiling                      SettingsUsage = 0xbd97146f
	SettingTriplanarVertexNormalLerp            SettingsUsage = 0x182674ab
	SettingTurbulence                           SettingsUsage = 0x2c799306
	SettingUi3dCamera                           SettingsUsage = 0x8ec7dc42
	SettingUi3dCameraDir                        SettingsUsage = 0x0e2d53d5
	SettingUi3dCameraPos                        SettingsUsage = 0x1ce1f2e0
	SettingUi3dLightColor                       SettingsUsage = 0x5c7df6a7
	SettingUi3dLightDirection                   SettingsUsage = 0x6e243f8a
	SettingUi3dLightPosition                    SettingsUsage = 0xe7f347a6
	SettingUi3dRenderRect                       SettingsUsage = 0x64c957ba
	SettingUi3dResolution                       SettingsUsage = 0xb847296e
	SettingUi3dScenarioIndex                    SettingsUsage = 0x3cc53d0c
	SettingUi3dShadowCamera0                    SettingsUsage = 0xde08a591
	SettingUi3dShadowCamera1                    SettingsUsage = 0x283cfe3c
	SettingUi3dShadowCamera2                    SettingsUsage = 0x2dc52675
	SettingUi3dShadowCamera3                    SettingsUsage = 0x6805b0d6
	SettingUi3dShadowmapResolution              SettingsUsage = 0xa4fa2b7a
	SettingUi3dShadows                          SettingsUsage = 0x23573aaf
	SettingUiSpecularCubemap                    SettingsUsage = 0xd65153d6
	SettingUnderlyingNormalBehindDecalOpacity   SettingsUsage = 0x16790495
	SettingUpFadeCurve                          SettingsUsage = 0xaf328b9c
	SettingUpFadeDistance                       SettingsUsage = 0x49a6212e
	SettingUseCcrouSpecIrWet                    SettingsUsage = 0xedb40b00
	SettingUseChannelRForWater                  SettingsUsage = 0xe148526a
	SettingUseErodeMap                          SettingsUsage = 0xd588b138
	SettingUseForwardAxis                       SettingsUsage = 0x322fc2f6
	SettingUseFresnel                           SettingsUsage = 0xec449df3
	SettingUseLargeWoundLookup                  SettingsUsage = 0xee28022c
	SettingUseNormalMap                         SettingsUsage = 0x6d98475a
	SettingUseNormalMapAlpha                    SettingsUsage = 0x4ec7b525
	SettingUseObjectCameraTransform             SettingsUsage = 0x5535c51f
	SettingUseParallax                          SettingsUsage = 0xeeea0640
	SettingUseParticleAlpha                     SettingsUsage = 0x44b30730
	SettingUseSss                               SettingsUsage = 0x9c857d5d
	SettingUseTrample                           SettingsUsage = 0x63e8fb5f
	SettingUseUpVector                          SettingsUsage = 0x41848a5d
	SettingUvDistortion                         SettingsUsage = 0x7e5a29de
	SettingUvExp                                SettingsUsage = 0x14f54d76
	SettingUvMappingTechniques                  SettingsUsage = 0x2aa849e7
	SettingUvOffset                             SettingsUsage = 0x312af125
	SettingUvRect                               SettingsUsage = 0xf8f6cf90
	SettingUvRotate01                           SettingsUsage = 0x851c66a8
	SettingUvRotate02                           SettingsUsage = 0x36bb3e24
	SettingUvScale                              SettingsUsage = 0x42f8951f
	SettingUvScalingMult                        SettingsUsage = 0xeec7e67e
	SettingUvScalingRatio                       SettingsUsage = 0x600047db
	SettingVertexDeformationFlags               SettingsUsage = 0x75375353
	SettingView                                 SettingsUsage = 0x3fce6c34
	SettingViewProj                             SettingsUsage = 0x57cbaec2
	SettingVolumetricCloudsColor                SettingsUsage = 0xa982cdae
	SettingVolumetricCloudsColorProbe           SettingsUsage = 0x147a0440
	SettingVolumetricCloudsShadowsFinal         SettingsUsage = 0x26afa44b
	SettingVolumetricFog3dImage                 SettingsUsage = 0x8ad188a6
	SettingVpMaxSlice0                          SettingsUsage = 0x430850bf
	SettingVpMaxSlice1                          SettingsUsage = 0xf8cdbe7d
	SettingVpMaxSlice2                          SettingsUsage = 0x6c9b51a2
	SettingVpMaxSlice3                          SettingsUsage = 0x4330a0f9
	SettingVpMinSlice0                          SettingsUsage = 0x0f5c7f1a
	SettingVpMinSlice1                          SettingsUsage = 0xce7cb68d
	SettingVpMinSlice2                          SettingsUsage = 0xa1d4675c
	SettingVpMinSlice3                          SettingsUsage = 0x59453b4a
	SettingVpRenderResolution                   SettingsUsage = 0xf0c9cdda
	SettingWaterColor                           SettingsUsage = 0x5febd863
	SettingWaterContrast                        SettingsUsage = 0x1a077cd6
	SettingWaterOpacity                         SettingsUsage = 0x219e6755
	SettingWaterRotate                          SettingsUsage = 0xaa2dd43a
	SettingWeatheringAmount                     SettingsUsage = 0x9134dae5
	SettingWeatheringAoAmount                   SettingsUsage = 0x1b7ecef5
	SettingWeatheringAoCutoff                   SettingsUsage = 0x581549f3
	SettingWeatheringCoverage                   SettingsUsage = 0x4a7cd0ef
	SettingWeatheringCoverageMultiplier         SettingsUsage = 0x846db408
	SettingWeatheringCurve                      SettingsUsage = 0x8f6ce1e4
	SettingWeatheringDirtWeightsNegative        SettingsUsage = 0xed874ebf
	SettingWeatheringDirtWeightsPositive        SettingsUsage = 0xb3f0a394
	SettingWeatheringDynamicTiling              SettingsUsage = 0xbd16a396
	SettingWeatheringHeight                     SettingsUsage = 0x8b37df5a
	SettingWeatheringHeightCurve                SettingsUsage = 0xecc7d525
	SettingWeatheringMasking                    SettingsUsage = 0x9ed04da2
	SettingWeatheringNormalIntensity            SettingsUsage = 0x54d24dfc
	SettingWeatheringNormalMult                 SettingsUsage = 0xa99c8a7e
	SettingWeatheringOnDirtAmount               SettingsUsage = 0x0a34cea3
	SettingWeatheringSpecialCase                SettingsUsage = 0x3c015556
	SettingWeatheringSpecialWeightsNegative     SettingsUsage = 0xf6400b64
	SettingWeatheringSpecialWeightsPositive     SettingsUsage = 0x981aafb0
	SettingWeatheringThickness                  SettingsUsage = 0x1bf10d7f
	SettingWeatheringTileFactor                 SettingsUsage = 0xe8c432eb
	SettingWeatheringTiling                     SettingsUsage = 0x325d6bd5
	SettingWeatheringUpFadeInfluence            SettingsUsage = 0x5f01c74b
	SettingWeatheringVariant                    SettingsUsage = 0x60e7d2a1
	SettingWeatheringWrap                       SettingsUsage = 0xf62ae492
	SettingWetness                              SettingsUsage = 0x43c7218a
	SettingWetnessBottom                        SettingsUsage = 0xa02f567b
	SettingWetnessPuddles                       SettingsUsage = 0xcba02151
	SettingWorld                                SettingsUsage = 0x4d46ae3b
	SettingWorldMaskSize                        SettingsUsage = 0xefcc0fe9
	SettingWoundPaintingEnabled                 SettingsUsage = 0x4485c321
	SettingWoundParallaxScale                   SettingsUsage = 0x613a8a38
	SettingWoundTileScale                       SettingsUsage = 0x2e2fcb48
	SettingWoundableId                          SettingsUsage = 0x6f2bffb2
	SettingZeroingDist                          SettingsUsage = 0xfcf98af9
	SettingZoneWetness                          SettingsUsage = 0x26cbbde4
	SettingCameraFadeDistance01                 SettingsUsage = 0x3e604f2a
	SettingColorDamaged                         SettingsUsage = 0xf7a19ea6
	SettingColorDamagedMult                     SettingsUsage = 0x351553a7
	SettingColorOuter                           SettingsUsage = 0x5a22c0ca
	SettingDistFadeOffset01                     SettingsUsage = 0x4ad6b4b8
	SettingDotsMult                             SettingsUsage = 0x0a32bc84
	SettingEnergyLinesSpeed                     SettingsUsage = 0x60843c26
	SettingFarFadeDistance01                    SettingsUsage = 0x4e3ab5db
	SettingFarFadeOffset01                      SettingsUsage = 0x1e43da40
	SettingFarFadeOpacity01                     SettingsUsage = 0xc19eda3a
	SettingFresnelPower                         SettingsUsage = 0xec00ece8
	SettingHitExp01                             SettingsUsage = 0xcd508b25
	SettingHitIntensity                         SettingsUsage = 0xda09185c
	SettingIlluminateDotsTile                   SettingsUsage = 0xda6187c3
	SettingInteresectDistance                   SettingsUsage = 0xa7a3f6a7
	SettingInteresectDistanceMult               SettingsUsage = 0x7a817610
	SettingInteresectExp                        SettingsUsage = 0x9db9a243
	SettingLowHealth                            SettingsUsage = 0xa99194f9
	SettingLumMin02                             SettingsUsage = 0x95a0c459
	SettingLuminocityExp02                      SettingsUsage = 0x8991fe16
	SettingMaxOpac                              SettingsUsage = 0x9b9b4ba4
	SettingMinOpac                              SettingsUsage = 0x8ac92075
	SettingNoiseFresnelExp                      SettingsUsage = 0xf13a78c8
	SettingNoiseFresnelMin                      SettingsUsage = 0xe0a06d64
	SettingOpacityMult01                        SettingsUsage = 0x0198c936
	SettingPower01                              SettingsUsage = 0x39d2fc5a
	SettingSpeedLinesMult                       SettingsUsage = 0x2b79ec65
	SettingSpeedLinesTile                       SettingsUsage = 0x39b35d0e
	SettingTopGlowMult                          SettingsUsage = 0xc5b203a2
	SettingTopGlowTighten                       SettingsUsage = 0xeb848c1a
	SettingUseVetexColor                        SettingsUsage = 0xbdb70075
	SettingUseHitData                           SettingsUsage = 0xf1ff10ad
	SettingWpoDistance01                        SettingsUsage = 0x7d438d95
)

func (usage *SettingsUsage) String() string {
	switch *usage {
	case SettingAlpha:
		return "Alpha"
	case SettingEnvironmentLuminosity:
		return "EnvironmentLuminosity"
	case SettingSineAmplitude01:
		return "SineAmplitude01"
	case SettingSineFrequency:
		return "SineFrequency"
	case SettingSineSecondaryAmplitude:
		return "SineSecondaryAmplitude"
	case SettingSineSecondaryFrequency:
		return "SineSecondaryFrequency"
	case SettingSineSecondarySpeed:
		return "SineSecondarySpeed"
	case SettingSineSpeed:
		return "SineSpeed"
	case SettingTimeWorldPosOffset:
		return "TimeWorldPosOffset"
	case SettingUseDirectionFromCenter:
		return "UseDirectionFromCenter"
	case SettingUseUvVertexMaskVExp:
		return "UseUvVertexMaskVExp"
	case SettingUvMask:
		return "UvMask"
	case SettingUvUMask:
		return "UvUMask"
	case SettingUvUMaskMax:
		return "UvUMaskMax"
	case SettingUvUMaskMin:
		return "UvUMaskMin"
	case SettingUvUMultiplier:
		return "UvUMultiplier"
	case SettingUvVMask:
		return "UvVMask"
	case SettingUvVMaskMax:
		return "UvVMaskMax"
	case SettingUvVMaskMin:
		return "UvVMaskMin"
	case SettingUvVertexMask:
		return "UvVertexMask"
	case SettingUvVertexMaskMax:
		return "UvVertexMaskMax"
	case SettingUvVertexMaskMin:
		return "UvVertexMaskMin"
	case SettingUvVertexMaskVExp:
		return "UvVertexMaskVExp"
	case SettingAaMultiplier:
		return "AaMultiplier"
	case SettingAberrateBlue:
		return "AberrateBlue"
	case SettingAberrateGreen:
		return "AberrateGreen"
	case SettingAberrateRed:
		return "AberrateRed"
	case SettingAbsoluteHeight:
		return "AbsoluteHeight"
	case SettingAlbedoIntensity:
		return "AlbedoIntensity"
	case SettingAlbedoIntensityGround:
		return "AlbedoIntensityGround"
	case SettingAlbedoIntensityRock:
		return "AlbedoIntensityRock"
	case SettingAlbedoIntensityVistaDetail:
		return "AlbedoIntensityVistaDetail"
	case SettingAlphaBackground:
		return "AlphaBackground"
	case SettingAlphaBorder:
		return "AlphaBorder"
	case SettingAlphaFill:
		return "AlphaFill"
	case SettingAlphaMultiplier:
		return "AlphaMultiplier"
	case SettingAmbientAmoint:
		return "AmbientAmoint"
	case SettingAmbientAmount:
		return "AmbientAmount"
	case SettingAngleFade:
		return "AngleFade"
	case SettingAngleFadeEnd:
		return "AngleFadeEnd"
	case SettingAngleFadeStart:
		return "AngleFadeStart"
	case SettingArrayIndex:
		return "ArrayIndex"
	case SettingAtlasSelect:
		return "AtlasSelect"
	case SettingAtlasSizeRatio:
		return "AtlasSizeRatio"
	case SettingAtmosphereLightColor:
		return "AtmosphereLightColor"
	case SettingAtmosphereLightDirection:
		return "AtmosphereLightDirection"
	case SettingAtmosphereRimlightColor:
		return "AtmosphereRimlightColor"
	case SettingAtmosphereRimlightIntensity:
		return "AtmosphereRimlightIntensity"
	case SettingAtmosphereSaturation:
		return "AtmosphereSaturation"
	case SettingAtmosphericLookup:
		return "AtmosphericLookup"
	case SettingBackFaceVisibility:
		return "BackFaceVisibility"
	case SettingBackgroundTile:
		return "BackgroundTile"
	case SettingBarrelOffset:
		return "BarrelOffset"
	case SettingBaseAoIntensity:
		return "BaseAoIntensity"
	case SettingBaseColor:
		return "BaseColor"
	case SettingBaseColorOpacity:
		return "BaseColorOpacity"
	case SettingBaseNormalIntensity:
		return "BaseNormalIntensity"
	case SettingBaseOpacity:
		return "BaseOpacity"
	case SettingBcMultiplier:
		return "BcMultiplier"
	case SettingBiomeATile:
		return "BiomeATile"
	case SettingBiomeBTile:
		return "BiomeBTile"
	case SettingBlackHoleEnable:
		return "BlackHoleEnable"
	case SettingBloodColor:
		return "BloodColor"
	case SettingBloodColor0:
		return "BloodColor0"
	case SettingBloodColor1:
		return "BloodColor1"
	case SettingBloodGunkNormalIntensity:
		return "BloodGunkNormalIntensity"
	case SettingBloodNormalFade:
		return "BloodNormalFade"
	case SettingBloodRoughness:
		return "BloodRoughness"
	case SettingBloodScale:
		return "BloodScale"
	case SettingBloodSubsurface:
		return "BloodSubsurface"
	case SettingBloodWeightsNegative:
		return "BloodWeightsNegative"
	case SettingBloodWeightsPositive:
		return "BloodWeightsPositive"
	case SettingBombSharpen:
		return "BombSharpen"
	case SettingBool:
		return "Bool"
	case SettingBorderFalloff:
		return "BorderFalloff"
	case SettingBorderWidth:
		return "BorderWidth"
	case SettingBottomColorTint:
		return "BottomColorTint"
	case SettingBottomDirtValue:
		return "BottomDirtValue"
	case SettingBoundingVolume:
		return "BoundingVolume"
	case SettingBugGunkMinimum:
		return "BugGunkMinimum"
	case SettingBugGunkWeightsNegative:
		return "BugGunkWeightsNegative"
	case SettingBugGunkWeightsPositive:
		return "BugGunkWeightsPositive"
	case SettingBurnScorch:
		return "BurnScorch"
	case SettingBurnScorchSnowMeltGlobalWetness:
		return "BurnScorchSnowMeltGlobalWetness"
	case SettingCAtmosphereCommon:
		return "CAtmosphereCommon"
	case SettingCCloudStartStop:
		return "CCloudStartStop"
	case SettingCPerInstance:
		return "CPerInstance"
	case SettingCPerObject:
		return "CPerObject"
	case SettingCUi3d:
		return "CUi3d"
	case SettingCameraCenterPos:
		return "CameraCenterPos"
	case SettingCameraFadeDistance:
		return "CameraFadeDistance"
	case SettingCameraInvProjection:
		return "CameraInvProjection"
	case SettingCameraInvView:
		return "CameraInvView"
	case SettingCameraLastInvProjection:
		return "CameraLastInvProjection"
	case SettingCameraLastInvView:
		return "CameraLastInvView"
	case SettingCameraLastProjection:
		return "CameraLastProjection"
	case SettingCameraLastView:
		return "CameraLastView"
	case SettingCameraLastViewProjection:
		return "CameraLastViewProjection"
	case SettingCameraNearFar:
		return "CameraNearFar"
	case SettingCameraPos:
		return "CameraPos"
	case SettingCameraProjection:
		return "CameraProjection"
	case SettingCameraUnprojection:
		return "CameraUnprojection"
	case SettingCameraView:
		return "CameraView"
	case SettingCameraViewProjection:
		return "CameraViewProjection"
	case SettingCapeHeightFromGround:
		return "CapeHeightFromGround"
	case SettingCapeHeightMult:
		return "CapeHeightMult"
	case SettingCenterLightDistanceFalloff:
		return "CenterLightDistanceFalloff"
	case SettingCenterLightDistanceMult:
		return "CenterLightDistanceMult"
	case SettingCenterLightIntensity:
		return "CenterLightIntensity"
	case SettingChannelSelection:
		return "ChannelSelection"
	case SettingCivilizationAmount:
		return "CivilizationAmount"
	case SettingClearcoatIntensity:
		return "ClearcoatIntensity"
	case SettingClearcoatNormalMix:
		return "ClearcoatNormalMix"
	case SettingClearcoatRoughness:
		return "ClearcoatRoughness"
	case SettingClipBox:
		return "ClipBox"
	case SettingClipCenter:
		return "ClipCenter"
	case SettingClipDistance:
		return "ClipDistance"
	case SettingClosestReflectionMap:
		return "ClosestReflectionMap"
	case SettingCloudAmount:
		return "CloudAmount"
	case SettingCloudColor:
		return "CloudColor"
	case SettingCloudContrast:
		return "CloudContrast"
	case SettingCloudHeightMult:
		return "CloudHeightMult"
	case SettingCloudOpacity:
		return "CloudOpacity"
	case SettingCloudRotate:
		return "CloudRotate"
	case SettingCloudSpeedU:
		return "CloudSpeedU"
	case SettingCloudSpeedV:
		return "CloudSpeedV"
	case SettingCloudStartHeight:
		return "CloudStartHeight"
	case SettingClusteredShadingData:
		return "ClusteredShadingData"
	case SettingColor:
		return "Color"
	case SettingColorBorder:
		return "ColorBorder"
	case SettingColorEdges:
		return "ColorEdges"
	case SettingColorFill:
		return "ColorFill"
	case SettingColorIntensity:
		return "ColorIntensity"
	case SettingColorLean:
		return "ColorLean"
	case SettingColorMult:
		return "ColorMult"
	case SettingColorMulti:
		return "ColorMulti"
	case SettingColorTint:
		return "ColorTint"
	case SettingColorVariationHighland:
		return "ColorVariationHighland"
	case SettingColorVariationLowland:
		return "ColorVariationLowland"
	case SettingConeRadiusAdjust:
		return "ConeRadiusAdjust"
	case SettingContextCamera:
		return "ContextCamera"
	case SettingCoriolisForce:
		return "CoriolisForce"
	case SettingCoriolisOffset:
		return "CoriolisOffset"
	case SettingCsActive:
		return "CsActive"
	case SettingCsCameraViewProj:
		return "CsCameraViewProj"
	case SettingCsClusterBuffer:
		return "CsClusterBuffer"
	case SettingCsClusterDataSize:
		return "CsClusterDataSize"
	case SettingCsClusterMaxDepthInvMaxDepth:
		return "CsClusterMaxDepthInvMaxDepth"
	case SettingCsClusterSizeInPixels:
		return "CsClusterSizeInPixels"
	case SettingCsClusterSizes:
		return "CsClusterSizes"
	case SettingCsLightDataBuffer:
		return "CsLightDataBuffer"
	case SettingCsLightDataSize:
		return "CsLightDataSize"
	case SettingCsLightIndexBuffer:
		return "CsLightIndexBuffer"
	case SettingCsLightIndexDataSize:
		return "CsLightIndexDataSize"
	case SettingCsLightShadowMatricesBuffer:
		return "CsLightShadowMatricesBuffer"
	case SettingCsLightShadowMatricesSize:
		return "CsLightShadowMatricesSize"
	case SettingCsShadowAtlasSize:
		return "CsShadowAtlasSize"
	case SettingCubemapFrameHistoryInvalidation:
		return "CubemapFrameHistoryInvalidation"
	case SettingDarkestValue:
		return "DarkestValue"
	case SettingDebugLod:
		return "DebugLod"
	case SettingDebugMode:
		return "DebugMode"
	case SettingDebugRendering:
		return "DebugRendering"
	case SettingDebugShadowLod:
		return "DebugShadowLod"
	case SettingDebugSpace:
		return "DebugSpace"
	case SettingDecalAlphaOffset:
		return "DecalAlphaOffset"
	case SettingDecalAlphaSharpness:
		return "DecalAlphaSharpness"
	case SettingDecalFadeExp:
		return "DecalFadeExp"
	case SettingDecalNormalIntensity:
		return "DecalNormalIntensity"
	case SettingDecalNormalIntentity:
		return "DecalNormalIntentity"
	case SettingDecalNormalOffset:
		return "DecalNormalOffset"
	case SettingDecalScalarfieldEnd:
		return "DecalScalarfieldEnd"
	case SettingDeepWaterColor:
		return "DeepWaterColor"
	case SettingDeltaTime:
		return "DeltaTime"
	case SettingDepth:
		return "Depth"
	case SettingDepthFade:
		return "DepthFade"
	case SettingDepthFadeDist:
		return "DepthFadeDist"
	case SettingDepthFadeDistance:
		return "DepthFadeDistance"
	case SettingDeriveNormalZ:
		return "DeriveNormalZ"
	case SettingDesaturation:
		return "Desaturation"
	case SettingDetailCurvIntensity1:
		return "DetailCurvIntensity1"
	case SettingDetailCurvIntensity2:
		return "DetailCurvIntensity2"
	case SettingDetailCurvIntensityLeather:
		return "DetailCurvIntensityLeather"
	case SettingDetailCurvIntensityPorcelain:
		return "DetailCurvIntensityPorcelain"
	case SettingDetailMixWeight:
		return "DetailMixWeight"
	case SettingDetailNormClearcoatIntensity:
		return "DetailNormClearcoatIntensity"
	case SettingDetailNormClearcoatIntensity1:
		return "DetailNormClearcoatIntensity1"
	case SettingDetailNormClearcoatIntensity2:
		return "DetailNormClearcoatIntensity2"
	case SettingDetailNormIntensity1:
		return "DetailNormIntensity1"
	case SettingDetailNormIntensity2:
		return "DetailNormIntensity2"
	case SettingDetailNormIntensityLeather:
		return "DetailNormIntensityLeather"
	case SettingDetailNormIntensityPorcelain:
		return "DetailNormIntensityPorcelain"
	case SettingDetailNormTiler1:
		return "DetailNormTiler1"
	case SettingDetailNormTiler2:
		return "DetailNormTiler2"
	case SettingDetailNormTilerLeather:
		return "DetailNormTilerLeather"
	case SettingDetailNormTilerPorcelain:
		return "DetailNormTilerPorcelain"
	case SettingDetailNormalIntensity:
		return "DetailNormalIntensity"
	case SettingDetailNormalSize:
		return "DetailNormalSize"
	case SettingDetailRoughnessClearcoatIntensity1:
		return "DetailRoughnessClearcoatIntensity1"
	case SettingDetailRoughnessClearcoatIntensity2:
		return "DetailRoughnessClearcoatIntensity2"
	case SettingDetailTileFactorMult:
		return "DetailTileFactorMult"
	case SettingDevSelectionColor:
		return "DevSelectionColor"
	case SettingDevSelectionMask:
		return "DevSelectionMask"
	case SettingDiffuseIntensity:
		return "DiffuseIntensity"
	case SettingDirtAo:
		return "DirtAo"
	case SettingDirtAoCoverage:
		return "DirtAoCoverage"
	case SettingDirtAoSharpness:
		return "DirtAoSharpness"
	case SettingDirtColor:
		return "DirtColor"
	case SettingDirtDetailAo:
		return "DirtDetailAo"
	case SettingDirtDetailMask:
		return "DirtDetailMask"
	case SettingDirtDetailMasking:
		return "DirtDetailMasking"
	case SettingDirtDetailSharpness:
		return "DirtDetailSharpness"
	case SettingDirtGlobalAmount:
		return "DirtGlobalAmount"
	case SettingDirtGradientMax:
		return "DirtGradientMax"
	case SettingDirtGradientMin:
		return "DirtGradientMin"
	case SettingDirtIntensity:
		return "DirtIntensity"
	case SettingDirtMetallic:
		return "DirtMetallic"
	case SettingDirtRoughness:
		return "DirtRoughness"
	case SettingDirtRoughnessBlend:
		return "DirtRoughnessBlend"
	case SettingDirtSharpness:
		return "DirtSharpness"
	case SettingDisplaceUvs:
		return "DisplaceUvs"
	case SettingDisplacementScale:
		return "DisplacementScale"
	case SettingDistFadeOffset:
		return "DistFadeOffset"
	case SettingDistSpeed:
		return "DistSpeed"
	case SettingDistTile:
		return "DistTile"
	case SettingDistortion:
		return "Distortion"
	case SettingDistortionAmount:
		return "DistortionAmount"
	case SettingDistressUvInfo:
		return "DistressUvInfo"
	case SettingDrynessAmount:
		return "DrynessAmount"
	case SettingDustColor:
		return "DustColor"
	case SettingDustFbm:
		return "DustFbm"
	case SettingDustOpacity:
		return "DustOpacity"
	case SettingDustTilingAmount:
		return "DustTilingAmount"
	case SettingDword:
		return "Dword"
	case SettingEdgeFadeOffset:
		return "EdgeFadeOffset"
	case SettingEdgeFadeTint:
		return "EdgeFadeTint"
	case SettingEdgeNormalIntensity:
		return "EdgeNormalIntensity"
	case SettingEdgeNormalSharpness:
		return "EdgeNormalSharpness"
	case SettingEmissiveAnimation:
		return "EmissiveAnimation"
	case SettingEmissiveColor:
		return "EmissiveColor"
	case SettingEmissiveColorA:
		return "EmissiveColorA"
	case SettingEmissiveInnerExp:
		return "EmissiveInnerExp"
	case SettingEmissiveIntensity:
		return "EmissiveIntensity"
	case SettingEmissiveMult:
		return "EmissiveMult"
	case SettingEmissiveOuterExp:
		return "EmissiveOuterExp"
	case SettingEmissiveStrength:
		return "EmissiveStrength"
	case SettingEmissiveUVDirection:
		return "EmissiveUVDirection"
	case SettingEmissiveWaveGradient:
		return "EmissiveWaveGradient"
	case SettingEmissiveWaveSize:
		return "EmissiveWaveSize"
	case SettingEmissiveWaveSpeed:
		return "EmissiveWaveSpeed"
	case SettingEndFadeExp:
		return "EndFadeExp"
	case SettingEndFadeTightness:
		return "EndFadeTightness"
	case SettingEndTaper:
		return "EndTaper"
	case SettingEnvLumMin:
		return "EnvLumMin"
	case SettingEnvLumMin01:
		return "EnvLumMin01"
	case SettingErodeMult:
		return "ErodeMult"
	case SettingErodeSoftness:
		return "ErodeSoftness"
	case SettingExposure:
		return "Exposure"
	case SettingExposure280:
		return "Exposure280"
	case SettingFadeDepth:
		return "FadeDepth"
	case SettingFadeInOutType:
		return "FadeInOutType"
	case SettingFadeYAngle:
		return "FadeYAngle"
	case SettingFarScatterDensity:
		return "FarScatterDensity"
	case SettingFarScatterNormalIntensityMult:
		return "FarScatterNormalIntensityMult"
	case SettingFlareNoiseSpeed:
		return "FlareNoiseSpeed"
	case SettingFlareNoiseSpeed02:
		return "FlareNoiseSpeed02"
	case SettingFlareNoiseTile:
		return "FlareNoiseTile"
	case SettingFlareNoiseTile02:
		return "FlareNoiseTile02"
	case SettingFlareTexExp:
		return "FlareTexExp"
	case SettingFlickerMin:
		return "FlickerMin"
	case SettingFlickerSpd:
		return "FlickerSpd"
	case SettingFogAmbientDuringTransitionColorBoost:
		return "FogAmbientDuringTransitionColorBoost"
	case SettingFogBackscatterLerp:
		return "FogBackscatterLerp"
	case SettingFogBackscatterPhase:
		return "FogBackscatterPhase"
	case SettingFogColor:
		return "FogColor"
	case SettingFogColorHax:
		return "FogColorHax"
	case SettingFogDustiness:
		return "FogDustiness"
	case SettingFogEnabled:
		return "FogEnabled"
	case SettingFogForwardscatterPhase:
		return "FogForwardscatterPhase"
	case SettingFogIntesity:
		return "FogIntesity"
	case SettingFogLightAmbientIntensity:
		return "FogLightAmbientIntensity"
	case SettingFogLightPollution:
		return "FogLightPollution"
	case SettingFogParameters:
		return "FogParameters"
	case SettingFogShadowIntensity:
		return "FogShadowIntensity"
	case SettingFogSunIntensity:
		return "FogSunIntensity"
	case SettingFogVolumeAlbedoIntensity:
		return "FogVolumeAlbedoIntensity"
	case SettingFogVolumeAlbedoLerp:
		return "FogVolumeAlbedoLerp"
	case SettingFogVolumeColor:
		return "FogVolumeColor"
	case SettingFogVolumeDensity:
		return "FogVolumeDensity"
	case SettingFogVolumeDustiness:
		return "FogVolumeDustiness"
	case SettingFogVolumeFalloffPow:
		return "FogVolumeFalloffPow"
	case SettingFogVolumeHeight:
		return "FogVolumeHeight"
	case SettingFrameNumber:
		return "FrameNumber"
	case SettingFrames:
		return "Frames"
	case SettingFresnel:
		return "Fresnel"
	case SettingFresnelDistance:
		return "FresnelDistance"
	case SettingFresnelEdges:
		return "FresnelEdges"
	case SettingFresnelEdgesMult:
		return "FresnelEdgesMult"
	case SettingFresnelExp:
		return "FresnelExp"
	case SettingFresnelExpMax:
		return "FresnelExpMax"
	case SettingFresnelInterior:
		return "FresnelInterior"
	case SettingFresnelMin:
		return "FresnelMin"
	case SettingFresnelMult:
		return "FresnelMult"
	case SettingFrostWeight:
		return "FrostWeight"
	case SettingGalaxyScaleAlignment:
		return "GalaxyScaleAlignment"
	case SettingGalaxyThickness:
		return "GalaxyThickness"
	case SettingGlintAmount:
		return "GlintAmount"
	case SettingGlintIntensity:
		return "GlintIntensity"
	case SettingGlintRoughness:
		return "GlintRoughness"
	case SettingGlintSize:
		return "GlintSize"
	case SettingGlitchAmount:
		return "GlitchAmount"
	case SettingGlitchCenterOffset:
		return "GlitchCenterOffset"
	case SettingGlitchColor:
		return "GlitchColor"
	case SettingGlitchColorPower:
		return "GlitchColorPower"
	case SettingGlitchGridsize:
		return "GlitchGridsize"
	case SettingGlitchSpeed:
		return "GlitchSpeed"
	case SettingGlobalDetailTile:
		return "GlobalDetailTile"
	case SettingGlobalDiffuseMap:
		return "GlobalDiffuseMap"
	case SettingGlobalSurfaceTile:
		return "GlobalSurfaceTile"
	case SettingGlobalViewport:
		return "GlobalViewport"
	case SettingGlowContrast:
		return "GlowContrast"
	case SettingGlowIntensity:
		return "GlowIntensity"
	case SettingGlowOffset:
		return "GlowOffset"
	case SettingGlowTemperature:
		return "GlowTemperature"
	case SettingGradientColor01:
		return "GradientColor01"
	case SettingGradientColor02:
		return "GradientColor02"
	case SettingGradientColorExp:
		return "GradientColorExp"
	case SettingGradientColorMult:
		return "GradientColorMult"
	case SettingGradientDirtMaxheight:
		return "GradientDirtMaxheight"
	case SettingGradientDirtMinheight:
		return "GradientDirtMinheight"
	case SettingGradientExp:
		return "GradientExp"
	case SettingGradientMult:
		return "GradientMult"
	case SettingGradientSubtractExp:
		return "GradientSubtractExp"
	case SettingGradingGroupId:
		return "GradingGroupId"
	case SettingGradingGroupIdGround:
		return "GradingGroupIdGround"
	case SettingGradingGroupIdMaskedDetails:
		return "GradingGroupIdMaskedDetails"
	case SettingGradingGroupIdRock:
		return "GradingGroupIdRock"
	case SettingGradingGroupIdSecondaryColor:
		return "GradingGroupIdSecondaryColor"
	case SettingGradingGroupIdSecondworld:
		return "GradingGroupIdSecondworld"
	case SettingGradingGroupIdThirdworld:
		return "GradingGroupIdThirdworld"
	case SettingGradingGroupIdTrunk:
		return "GradingGroupIdTrunk"
	case SettingGradingGroupIdVistaDetail:
		return "GradingGroupIdVistaDetail"
	case SettingGradingGroupIdWeathering:
		return "GradingGroupIdWeathering"
	case SettingGradingSecondaryGroupId:
		return "GradingSecondaryGroupId"
	case SettingGrainSpeed:
		return "GrainSpeed"
	case SettingGreyscale:
		return "Greyscale"
	case SettingGunkNormalFade:
		return "GunkNormalFade"
	case SettingGunkScale:
		return "GunkScale"
	case SettingHeightContrast:
		return "HeightContrast"
	case SettingHeightWetnessAndWash:
		return "HeightWetnessAndWash"
	case SettingHeightmapNormals:
		return "HeightmapNormals"
	case SettingHitExp:
		return "HitExp"
	case SettingHitRExp:
		return "HitRExp"
	case SettingHitRMult:
		return "HitRMult"
	case SettingHmapSize:
		return "HmapSize"
	case SettingHologramColor:
		return "HologramColor"
	case SettingHologramColor01:
		return "HologramColor01"
	case SettingHologramHideAmount:
		return "HologramHideAmount"
	case SettingHologramPlanetLightColors:
		return "HologramPlanetLightColors"
	case SettingHologramPlanetPositions:
		return "HologramPlanetPositions"
	case SettingHologramPlanetScaleMultiplier:
		return "HologramPlanetScaleMultiplier"
	case SettingHudCurveAmount:
		return "HudCurveAmount"
	case SettingIEnd:
		return "IEnd"
	case SettingIIntensity:
		return "IIntensity"
	case SettingIStart:
		return "IStart"
	case SettingIThickness:
		return "IThickness"
	case SettingIceFuzzColor:
		return "IceFuzzColor"
	case SettingIceFuzzIntensity:
		return "IceFuzzIntensity"
	case SettingIceSubsurfaceColor:
		return "IceSubsurfaceColor"
	case SettingIceSubsurfaceDiffusion:
		return "IceSubsurfaceDiffusion"
	case SettingIceSubsurfaceIntensity:
		return "IceSubsurfaceIntensity"
	case SettingIceSubsurfaceThickness:
		return "IceSubsurfaceThickness"
	case SettingIceSubsurfaceWrap:
		return "IceSubsurfaceWrap"
	case SettingIceTint:
		return "IceTint"
	case SettingIceWarp:
		return "IceWarp"
	case SettingIesLookup:
		return "IesLookup"
	case SettingIgnoreParticlecolor:
		return "IgnoreParticlecolor"
	case SettingImpTransparentOverride:
		return "ImpTransparentOverride"
	case SettingInitialSpawnPos:
		return "InitialSpawnPos"
	case SettingInstanceSeed:
		return "InstanceSeed"
	case SettingInstancingZero:
		return "InstancingZero"
	case SettingIntensity:
		return "Intensity"
	case SettingInterserctionBrightness:
		return "InterserctionBrightness"
	case SettingInterserctionExp:
		return "InterserctionExp"
	case SettingInterserctionThickness:
		return "InterserctionThickness"
	case SettingInvHmapSize:
		return "InvHmapSize"
	case SettingInvView:
		return "InvView"
	case SettingInvViewProj:
		return "InvViewProj"
	case SettingInvWorld:
		return "InvWorld"
	case SettingInvertFresnel:
		return "InvertFresnel"
	case SettingIoffset:
		return "Ioffset"
	case SettingJacobiFalloff:
		return "JacobiFalloff"
	case SettingLastWorld:
		return "LastWorld"
	case SettingLavaContrast:
		return "LavaContrast"
	case SettingLavaOffset:
		return "LavaOffset"
	case SettingLavaTemperature:
		return "LavaTemperature"
	case SettingLensColor:
		return "LensColor"
	case SettingLensCutoutEnabled:
		return "LensCutoutEnabled"
	case SettingLensEmissiveColor:
		return "LensEmissiveColor"
	case SettingLensEmissiveIntensity:
		return "LensEmissiveIntensity"
	case SettingLensEmissiveOpacity:
		return "LensEmissiveOpacity"
	case SettingLensEmissiveTexture:
		return "LensEmissiveTexture"
	case SettingLensOcclusionEnabled:
		return "LensOcclusionEnabled"
	case SettingLensOcclusionSize:
		return "LensOcclusionSize"
	case SettingLensOcclusionTexture:
		return "LensOcclusionTexture"
	case SettingLensOffset:
		return "LensOffset"
	case SettingLensOpacityMul:
		return "LensOpacityMul"
	case SettingLensParallaxMult:
		return "LensParallaxMult"
	case SettingLensScale:
		return "LensScale"
	case SettingLifetimeDrawMult:
		return "LifetimeDrawMult"
	case SettingLifetimeExp:
		return "LifetimeExp"
	case SettingLightProbeSpaceSpecular:
		return "LightProbeSpaceSpecular"
	case SettingLightingData:
		return "LightingData"
	case SettingLightsourceAngularSize:
		return "LightsourceAngularSize"
	case SettingLinearFadeOffsets:
		return "LinearFadeOffsets"
	case SettingLocalLightsShadowAtlas:
		return "LocalLightsShadowAtlas"
	case SettingLodCameraPos:
		return "LodCameraPos"
	case SettingLodFadeLevel:
		return "LodFadeLevel"
	case SettingLookupPhaseSpeed:
		return "LookupPhaseSpeed"
	case SettingLookupWeight:
		return "LookupWeight"
	case SettingLumMin:
		return "LumMin"
	case SettingLumMinRemap:
		return "LumMinRemap"
	case SettingLuminocityExp:
		return "LuminocityExp"
	case SettingLuminosityOpacity:
		return "LuminosityOpacity"
	case SettingLutContrast:
		return "LutContrast"
	case SettingLutMixBiomeA:
		return "LutMixBiomeA"
	case SettingLutMixBiomeB:
		return "LutMixBiomeB"
	case SettingLutOffset:
		return "LutOffset"
	case SettingMaskScale:
		return "MaskScale"
	case SettingMaskSharpnessBiome:
		return "MaskSharpnessBiome"
	case SettingMaskSharpnessLut:
		return "MaskSharpnessLut"
	case SettingMaskVariation:
		return "MaskVariation"
	case SettingMaterial01TileMultiplier:
		return "Material01TileMultiplier"
	case SettingMaterial02TileMultiplier:
		return "Material02TileMultiplier"
	case SettingMaterial03TileMultiplier:
		return "Material03TileMultiplier"
	case SettingMaterial04TileMultiplier:
		return "Material04TileMultiplier"
	case SettingMaterial05TileMultiplier:
		return "Material05TileMultiplier"
	case SettingMaterial06TileMultiplier:
		return "Material06TileMultiplier"
	case SettingMaterial07TileMultiplier:
		return "Material07TileMultiplier"
	case SettingMaterial08TileMultiplier:
		return "Material08TileMultiplier"
	case SettingMaterial1Metallic:
		return "Material1Metallic"
	case SettingMaterial1RoughnessBase:
		return "Material1RoughnessBase"
	case SettingMaterial1RoughnessBuildUp:
		return "Material1RoughnessBuildUp"
	case SettingMaterial1Surface:
		return "Material1Surface"
	case SettingMaterial1SurfaceNormal:
		return "Material1SurfaceNormal"
	case SettingMaterial1SurfaceRoughness:
		return "Material1SurfaceRoughness"
	case SettingMaterial1SurfaceValue:
		return "Material1SurfaceValue"
	case SettingMaterial1WearCavityEdge:
		return "Material1WearCavityEdge"
	case SettingMaterial2Metallic:
		return "Material2Metallic"
	case SettingMaterial2RoughnessBase:
		return "Material2RoughnessBase"
	case SettingMaterial2RoughnessBuildUp:
		return "Material2RoughnessBuildUp"
	case SettingMaterial2Surface:
		return "Material2Surface"
	case SettingMaterial2SurfaceNormal:
		return "Material2SurfaceNormal"
	case SettingMaterial2SurfaceRoughness:
		return "Material2SurfaceRoughness"
	case SettingMaterial2SurfaceValue:
		return "Material2SurfaceValue"
	case SettingMaterial2WearCavityEdge:
		return "Material2WearCavityEdge"
	case SettingMaterial2WearCavityEdge01:
		return "Material2WearCavityEdge01"
	case SettingMaterial2WearCavityEdge06:
		return "Material2WearCavityEdge06"
	case SettingMaterial3Metallic:
		return "Material3Metallic"
	case SettingMaterial3RoughnessBase:
		return "Material3RoughnessBase"
	case SettingMaterial3RoughnessBuildUp:
		return "Material3RoughnessBuildUp"
	case SettingMaterial3Surface:
		return "Material3Surface"
	case SettingMaterial3SurfaceNormal:
		return "Material3SurfaceNormal"
	case SettingMaterial3SurfaceRoughness:
		return "Material3SurfaceRoughness"
	case SettingMaterial3SurfaceValue:
		return "Material3SurfaceValue"
	case SettingMaterial4Metallic:
		return "Material4Metallic"
	case SettingMaterial4RoughnessBase:
		return "Material4RoughnessBase"
	case SettingMaterial4RoughnessBuildUp:
		return "Material4RoughnessBuildUp"
	case SettingMaterial4Surface:
		return "Material4Surface"
	case SettingMaterial4SurfaceNormal:
		return "Material4SurfaceNormal"
	case SettingMaterial4SurfaceRoughness:
		return "Material4SurfaceRoughness"
	case SettingMaterial4SurfaceValue:
		return "Material4SurfaceValue"
	case SettingMaterial4WearCavityEdge:
		return "Material4WearCavityEdge"
	case SettingMaterial5Metallic:
		return "Material5Metallic"
	case SettingMaterial5RoughnessBase:
		return "Material5RoughnessBase"
	case SettingMaterial5RoughnessBuildUp:
		return "Material5RoughnessBuildUp"
	case SettingMaterial5Surface:
		return "Material5Surface"
	case SettingMaterial5SurfaceNormal:
		return "Material5SurfaceNormal"
	case SettingMaterial5SurfaceRoughness:
		return "Material5SurfaceRoughness"
	case SettingMaterial5SurfaceValue:
		return "Material5SurfaceValue"
	case SettingMaterial5WearCavityEdge:
		return "Material5WearCavityEdge"
	case SettingMaterial6Metallic:
		return "Material6Metallic"
	case SettingMaterial6RoughnessBase:
		return "Material6RoughnessBase"
	case SettingMaterial6RoughnessBuildUp:
		return "Material6RoughnessBuildUp"
	case SettingMaterial6Surface:
		return "Material6Surface"
	case SettingMaterial6SurfaceNormal:
		return "Material6SurfaceNormal"
	case SettingMaterial6SurfaceRoughness:
		return "Material6SurfaceRoughness"
	case SettingMaterial6SurfaceValue:
		return "Material6SurfaceValue"
	case SettingMaterial6WearCavityEdge:
		return "Material6WearCavityEdge"
	case SettingMaterial7Metallic:
		return "Material7Metallic"
	case SettingMaterial7RoughnessBase:
		return "Material7RoughnessBase"
	case SettingMaterial7RoughnessBuildUp:
		return "Material7RoughnessBuildUp"
	case SettingMaterial7Surface:
		return "Material7Surface"
	case SettingMaterial7SurfaceNormal:
		return "Material7SurfaceNormal"
	case SettingMaterial7SurfaceRoughness:
		return "Material7SurfaceRoughness"
	case SettingMaterial7SurfaceValue:
		return "Material7SurfaceValue"
	case SettingMaterial7WearCavityEdge:
		return "Material7WearCavityEdge"
	case SettingMaterial8Metallic:
		return "Material8Metallic"
	case SettingMaterial8RoughnessBase:
		return "Material8RoughnessBase"
	case SettingMaterial8RoughnessBuildUp:
		return "Material8RoughnessBuildUp"
	case SettingMaterial8Surface:
		return "Material8Surface"
	case SettingMaterial8SurfaceNormal:
		return "Material8SurfaceNormal"
	case SettingMaterial8SurfaceRoughness:
		return "Material8SurfaceRoughness"
	case SettingMaterial8SurfaceValue:
		return "Material8SurfaceValue"
	case SettingMaterialIndex:
		return "MaterialIndex"
	case SettingMaterialVariable:
		return "MaterialVariable"
	case SettingMaterialWetness:
		return "MaterialWetness"
	case SettingMaxEmissive:
		return "MaxEmissive"
	case SettingMetalic:
		return "Metalic"
	case SettingMetallic01:
		return "Metallic01"
	case SettingMetallicOpacity:
		return "MetallicOpacity"
	case SettingMicroAo:
		return "MicroAo"
	case SettingMicroAoIntensity:
		return "MicroAoIntensity"
	case SettingMieBeta:
		return "MieBeta"
	case SettingMieHeight:
		return "MieHeight"
	case SettingMieTintHax:
		return "MieTintHax"
	case SettingMinEmissive:
		return "MinEmissive"
	case SettingMultiply:
		return "Multiply"
	case SettingNoise01Channel:
		return "Noise01Channel"
	case SettingNoise01Exp:
		return "Noise01Exp"
	case SettingNoise01ExpMax:
		return "Noise01ExpMax"
	case SettingNoise01ExpMin:
		return "Noise01ExpMin"
	case SettingNoise01Minmax:
		return "Noise01Minmax"
	case SettingNoise01Speed:
		return "Noise01Speed"
	case SettingNoise01Tile:
		return "Noise01Tile"
	case SettingNoise02Channel:
		return "Noise02Channel"
	case SettingNoise02Exp:
		return "Noise02Exp"
	case SettingNoise02Minmax:
		return "Noise02Minmax"
	case SettingNoise02Speed:
		return "Noise02Speed"
	case SettingNoise02Tile:
		return "Noise02Tile"
	case SettingNoiseChannel:
		return "NoiseChannel"
	case SettingNoiseChannel02:
		return "NoiseChannel02"
	case SettingNoiseExp:
		return "NoiseExp"
	case SettingNoiseMax:
		return "NoiseMax"
	case SettingNoiseMin:
		return "NoiseMin"
	case SettingNoiseMult:
		return "NoiseMult"
	case SettingNoiseOffset:
		return "NoiseOffset"
	case SettingNoiseScale:
		return "NoiseScale"
	case SettingNoiseStrength:
		return "NoiseStrength"
	case SettingNormalCcnormOpacity:
		return "NormalCcnormOpacity"
	case SettingNormalIntensity:
		return "NormalIntensity"
	case SettingNormalIntensityBiomeA:
		return "NormalIntensityBiomeA"
	case SettingNormalIntensityBiomeB:
		return "NormalIntensityBiomeB"
	case SettingNormalIntensityGround:
		return "NormalIntensityGround"
	case SettingNormalIntensityVistaDetail:
		return "NormalIntensityVistaDetail"
	case SettingNormalMirroring:
		return "NormalMirroring"
	case SettingNormalOpacity:
		return "NormalOpacity"
	case SettingNormalOverLife:
		return "NormalOverLife"
	case SettingNormalPullUp:
		return "NormalPullUp"
	case SettingNormals:
		return "Normals"
	case SettingNrmStr:
		return "NrmStr"
	case SettingNumTiles:
		return "NumTiles"
	case SettingOffsetExp:
		return "OffsetExp"
	case SettingOffsetFlashExp:
		return "OffsetFlashExp"
	case SettingOffsetFlashMult:
		return "OffsetFlashMult"
	case SettingOffsetMax:
		return "OffsetMax"
	case SettingOffsetMin:
		return "OffsetMin"
	case SettingOffsetMult:
		return "OffsetMult"
	case SettingOpacityOffset:
		return "OpacityOffset"
	case SettingOpacitySharpness:
		return "OpacitySharpness"
	case SettingOpacityThreshold:
		return "OpacityThreshold"
	case SettingOpacityThresholdFar:
		return "OpacityThresholdFar"
	case SettingOpacityTreshholdFadeDistanceInv:
		return "OpacityTreshholdFadeDistanceInv"
	case SettingOverlayMaskAmount:
		return "OverlayMaskAmount"
	case SettingPaletteSlot:
		return "PaletteSlot"
	case SettingParallaxBias:
		return "ParallaxBias"
	case SettingParallaxIntensity:
		return "ParallaxIntensity"
	case SettingParallaxIntensityCloud:
		return "ParallaxIntensityCloud"
	case SettingParallaxScale:
		return "ParallaxScale"
	case SettingParticleAgeLife:
		return "ParticleAgeLife"
	case SettingParticleColorOnly:
		return "ParticleColorOnly"
	case SettingPlanetRoughnessIntesity:
		return "PlanetRoughnessIntesity"
	case SettingPlanetScaleMult:
		return "PlanetScaleMult"
	case SettingPlanetShadowDistanceFalloff:
		return "PlanetShadowDistanceFalloff"
	case SettingPlanetWpPos:
		return "PlanetWpPos"
	case SettingPlanetWpScale:
		return "PlanetWpScale"
	case SettingPointTaper:
		return "PointTaper"
	case SettingPostEffectsEnabled:
		return "PostEffectsEnabled"
	case SettingPower:
		return "Power"
	case SettingPreventTerrainDeformation:
		return "PreventTerrainDeformation"
	case SettingPreventsTerrainDeformation:
		return "PreventsTerrainDeformation"
	case SettingProj:
		return "Proj"
	case SettingRampDownExpHacky:
		return "RampDownExpHacky"
	case SettingRawNonCheckerboardedTargetSize:
		return "RawNonCheckerboardedTargetSize"
	case SettingRawNonCheckerboardedViewport:
		return "RawNonCheckerboardedViewport"
	case SettingRayleighBeta:
		return "RayleighBeta"
	case SettingRayleighBetaSpaceplanet:
		return "RayleighBetaSpaceplanet"
	case SettingRegionPositionOffset:
		return "RegionPositionOffset"
	case SettingRemapMinValue:
		return "RemapMinValue"
	case SettingResolutionSetting:
		return "ResolutionSetting"
	case SettingReticleColor:
		return "ReticleColor"
	case SettingReticleColorIntensity:
		return "ReticleColorIntensity"
	case SettingReticleOpacity:
		return "ReticleOpacity"
	case SettingReticleTexture:
		return "ReticleTexture"
	case SettingReticuleScale:
		return "ReticuleScale"
	case SettingRiverAmount:
		return "RiverAmount"
	case SettingRiverSmooth:
		return "RiverSmooth"
	case SettingRoughness:
		return "Roughness"
	case SettingRoughnessMulti:
		return "RoughnessMulti"
	case SettingRoughnessOpacity:
		return "RoughnessOpacity"
	case SettingSampleTerrainAlbedo:
		return "SampleTerrainAlbedo"
	case SettingScalarFieldCutoff:
		return "ScalarFieldCutoff"
	case SettingScanSpeed:
		return "ScanSpeed"
	case SettingScanline01Intensity:
		return "Scanline01Intensity"
	case SettingScanline02Intensity:
		return "Scanline02Intensity"
	case SettingScanlineCount:
		return "ScanlineCount"
	case SettingScanlineCount02:
		return "ScanlineCount02"
	case SettingScanlineDistScale:
		return "ScanlineDistScale"
	case SettingScanlineThickness:
		return "ScanlineThickness"
	case SettingScanlineThickness02:
		return "ScanlineThickness02"
	case SettingSecondOpacityInfluence:
		return "SecondOpacityInfluence"
	case SettingSecondOpacityMin:
		return "SecondOpacityMin"
	case SettingSecondOpacityStart:
		return "SecondOpacityStart"
	case SettingSelectedOffset:
		return "SelectedOffset"
	case SettingSelectedScale:
		return "SelectedScale"
	case SettingSelfEmissiveColor:
		return "SelfEmissiveColor"
	case SettingSelfEmissiveIntensity:
		return "SelfEmissiveIntensity"
	case SettingSelfPlanetIndex:
		return "SelfPlanetIndex"
	case SettingShadowBiasSlice0:
		return "ShadowBiasSlice0"
	case SettingShadowBiasSlice1:
		return "ShadowBiasSlice1"
	case SettingShadowBiasSlice2:
		return "ShadowBiasSlice2"
	case SettingShadowBiasSlice3:
		return "ShadowBiasSlice3"
	case SettingShadowClampToNearPlane:
		return "ShadowClampToNearPlane"
	case SettingShadowDepthBiasSlice0:
		return "ShadowDepthBiasSlice0"
	case SettingShadowDepthBiasSlice1:
		return "ShadowDepthBiasSlice1"
	case SettingShadowDepthBiasSlice2:
		return "ShadowDepthBiasSlice2"
	case SettingShadowDepthBiasSlice3:
		return "ShadowDepthBiasSlice3"
	case SettingShadowIntensity:
		return "ShadowIntensity"
	case SettingShadowRotation:
		return "ShadowRotation"
	case SettingShadowScaleSlice0:
		return "ShadowScaleSlice0"
	case SettingShadowScaleSlice1:
		return "ShadowScaleSlice1"
	case SettingShadowScaleSlice2:
		return "ShadowScaleSlice2"
	case SettingShadowScaleSlice3:
		return "ShadowScaleSlice3"
	case SettingShadowsCasting:
		return "ShadowsCasting"
	case SettingShallowWaterColor:
		return "ShallowWaterColor"
	case SettingShieldId:
		return "ShieldId"
	case SettingSkipFacing:
		return "SkipFacing"
	case SettingSnowAmount:
		return "SnowAmount"
	case SettingSnowFromBottom:
		return "SnowFromBottom"
	case SettingSnowFromHeight:
		return "SnowFromHeight"
	case SettingSnowFromNormal:
		return "SnowFromNormal"
	case SettingSnowFromTop:
		return "SnowFromTop"
	case SettingSnowHardnessBottom:
		return "SnowHardnessBottom"
	case SettingSnowHardnessTop:
		return "SnowHardnessTop"
	case SettingSnowIndex:
		return "SnowIndex"
	case SettingSnowIndex0:
		return "SnowIndex0"
	case SettingSnowMask:
		return "SnowMask"
	case SettingSnowMaskDepth:
		return "SnowMaskDepth"
	case SettingSnowMaskHardness:
		return "SnowMaskHardness"
	case SettingSnowMaskTile:
		return "SnowMaskTile"
	case SettingSnowNormalDisplacement:
		return "SnowNormalDisplacement"
	case SettingSnowNormalIntensity:
		return "SnowNormalIntensity"
	case SettingSnowNormalIntensity01:
		return "SnowNormalIntensity01"
	case SettingSnowNormalMask:
		return "SnowNormalMask"
	case SettingSnowPerZoneDisabling:
		return "SnowPerZoneDisabling"
	case SettingSnowSsThickness:
		return "SnowSsThickness"
	case SettingSnowTile:
		return "SnowTile"
	case SettingSnowTile01:
		return "SnowTile01"
	case SettingSnowTrampleNormalBlur:
		return "SnowTrampleNormalBlur"
	case SettingSnowTrampleNormalIntensity:
		return "SnowTrampleNormalIntensity"
	case SettingSnowUpDisplacement:
		return "SnowUpDisplacement"
	case SettingSpecialGunkColor:
		return "SpecialGunkColor"
	case SettingSpecular:
		return "Specular"
	case SettingSpecularBrdfLut:
		return "SpecularBrdfLut"
	case SettingSpecularCurve:
		return "SpecularCurve"
	case SettingSpecularIntensity:
		return "SpecularIntensity"
	case SettingSpecularMulti:
		return "SpecularMulti"
	case SettingSpeed:
		return "Speed"
	case SettingSsDiffusion:
		return "SsDiffusion"
	case SettingSsIntensityMult:
		return "SsIntensityMult"
	case SettingSsThickness:
		return "SsThickness"
	case SettingSssIntensity:
		return "SssIntensity"
	case SettingSssWrap:
		return "SssWrap"
	case SettingStarColor:
		return "StarColor"
	case SettingStarOpacities:
		return "StarOpacities"
	case SettingStarTilingAmount:
		return "StarTilingAmount"
	case SettingSubsurfaceDiff:
		return "SubsurfaceDiff"
	case SettingSubsurfaceDiffusion:
		return "SubsurfaceDiffusion"
	case SettingSubsurfaceFadeCurve:
		return "SubsurfaceFadeCurve"
	case SettingSubsurfaceFadeDistance:
		return "SubsurfaceFadeDistance"
	case SettingSubsurfaceInt:
		return "SubsurfaceInt"
	case SettingSubsurfaceIntensity:
		return "SubsurfaceIntensity"
	case SettingSubsurfaceOcclusionIntensity:
		return "SubsurfaceOcclusionIntensity"
	case SettingSubsurfaceOcclusionMaskcurve:
		return "SubsurfaceOcclusionMaskcurve"
	case SettingSubsurfaceThickness:
		return "SubsurfaceThickness"
	case SettingSubsurfaceTint:
		return "SubsurfaceTint"
	case SettingSubsurfaceWrap:
		return "SubsurfaceWrap"
	case SettingSunAngularSize:
		return "SunAngularSize"
	case SettingSunCol:
		return "SunCol"
	case SettingSunColor:
		return "SunColor"
	case SettingSunColorOverride:
		return "SunColorOverride"
	case SettingSunDir:
		return "SunDir"
	case SettingSunDirection:
		return "SunDirection"
	case SettingSunDirectionOverride:
		return "SunDirectionOverride"
	case SettingSunEnabled:
		return "SunEnabled"
	case SettingSunShadowMap:
		return "SunShadowMap"
	case SettingSunShadowsEnabled:
		return "SunShadowsEnabled"
	case SettingTaaEnabled:
		return "TaaEnabled"
	case SettingTaperExp:
		return "TaperExp"
	case SettingTearAmount:
		return "TearAmount"
	case SettingTerrainBlendCurve:
		return "TerrainBlendCurve"
	case SettingTerrainBlendHeight:
		return "TerrainBlendHeight"
	case SettingTerrainColorLerp:
		return "TerrainColorLerp"
	case SettingTerrainPatchSize:
		return "TerrainPatchSize"
	case SettingTerrainSize:
		return "TerrainSize"
	case SettingTerrainTiling:
		return "TerrainTiling"
	case SettingTest01:
		return "Test01"
	case SettingTest02:
		return "Test02"
	case SettingTest03:
		return "Test03"
	case SettingTexIsGreyscale:
		return "TexIsGreyscale"
	case SettingTextureDensityVisualization:
		return "TextureDensityVisualization"
	case SettingTileFactor:
		return "TileFactor"
	case SettingTileFactorCloudsEdges:
		return "TileFactorCloudsEdges"
	case SettingTileFactorCloudsInside:
		return "TileFactorCloudsInside"
	case SettingTileFactorNoise:
		return "TileFactorNoise"
	case SettingTileSize:
		return "TileSize"
	case SettingTilerAoIntensity:
		return "TilerAoIntensity"
	case SettingTilerNormalIntensity:
		return "TilerNormalIntensity"
	case SettingTiling:
		return "Tiling"
	case SettingTilingFactor:
		return "TilingFactor"
	case SettingTilingGround:
		return "TilingGround"
	case SettingTilingVistaDetail:
		return "TilingVistaDetail"
	case SettingTime:
		return "Time"
	case SettingTimeAdd:
		return "TimeAdd"
	case SettingTimeMult:
		return "TimeMult"
	case SettingTimeMultiplier:
		return "TimeMultiplier"
	case SettingTimeOfDayOverridesEnabled:
		return "TimeOfDayOverridesEnabled"
	case SettingTint:
		return "Tint"
	case SettingTint01:
		return "Tint01"
	case SettingTint02:
		return "Tint02"
	case SettingTintMaskColor:
		return "TintMaskColor"
	case SettingTipGradientRemap:
		return "TipGradientRemap"
	case SettingTonemapA:
		return "TonemapA"
	case SettingTonemapB:
		return "TonemapB"
	case SettingTonemapC:
		return "TonemapC"
	case SettingTonemapD:
		return "TonemapD"
	case SettingTonemapE:
		return "TonemapE"
	case SettingTopColorTint:
		return "TopColorTint"
	case SettingTopDirtValue:
		return "TopDirtValue"
	case SettingTrampleAmount:
		return "TrampleAmount"
	case SettingTrampleFootstepMipBlur:
		return "TrampleFootstepMipBlur"
	case SettingTrampleFootstepNormalWeight:
		return "TrampleFootstepNormalWeight"
	case SettingTrampleMakeWetWeight:
		return "TrampleMakeWetWeight"
	case SettingTrampleMiscWeight:
		return "TrampleMiscWeight"
	case SettingTrampleNormalWeight:
		return "TrampleNormalWeight"
	case SettingTramplePuddleWeight:
		return "TramplePuddleWeight"
	case SettingTrampleSnowIndex:
		return "TrampleSnowIndex"
	case SettingTreeDepthScale:
		return "TreeDepthScale"
	case SettingTriplanarDetailTiler:
		return "TriplanarDetailTiler"
	case SettingTriplanarDetailTiler0:
		return "TriplanarDetailTiler0"
	case SettingTriplanarDetailTiler1:
		return "TriplanarDetailTiler1"
	case SettingTriplanarMaskSharpen:
		return "TriplanarMaskSharpen"
	case SettingTriplanarNormalIntensity:
		return "TriplanarNormalIntensity"
	case SettingTriplanarTiling:
		return "TriplanarTiling"
	case SettingTriplanarVertexNormalLerp:
		return "TriplanarVertexNormalLerp"
	case SettingTurbulence:
		return "Turbulence"
	case SettingUi3dCamera:
		return "Ui3dCamera"
	case SettingUi3dCameraDir:
		return "Ui3dCameraDir"
	case SettingUi3dCameraPos:
		return "Ui3dCameraPos"
	case SettingUi3dLightColor:
		return "Ui3dLightColor"
	case SettingUi3dLightDirection:
		return "Ui3dLightDirection"
	case SettingUi3dLightPosition:
		return "Ui3dLightPosition"
	case SettingUi3dRenderRect:
		return "Ui3dRenderRect"
	case SettingUi3dResolution:
		return "Ui3dResolution"
	case SettingUi3dScenarioIndex:
		return "Ui3dScenarioIndex"
	case SettingUi3dShadowCamera0:
		return "Ui3dShadowCamera0"
	case SettingUi3dShadowCamera1:
		return "Ui3dShadowCamera1"
	case SettingUi3dShadowCamera2:
		return "Ui3dShadowCamera2"
	case SettingUi3dShadowCamera3:
		return "Ui3dShadowCamera3"
	case SettingUi3dShadowmapResolution:
		return "Ui3dShadowmapResolution"
	case SettingUi3dShadows:
		return "Ui3dShadows"
	case SettingUiSpecularCubemap:
		return "UiSpecularCubemap"
	case SettingUnderlyingNormalBehindDecalOpacity:
		return "UnderlyingNormalBehindDecalOpacity"
	case SettingUpFadeCurve:
		return "UpFadeCurve"
	case SettingUpFadeDistance:
		return "UpFadeDistance"
	case SettingUseCcrouSpecIrWet:
		return "UseCcrouSpecIrWet"
	case SettingUseChannelRForWater:
		return "UseChannelRForWater"
	case SettingUseErodeMap:
		return "UseErodeMap"
	case SettingUseForwardAxis:
		return "UseForwardAxis"
	case SettingUseFresnel:
		return "UseFresnel"
	case SettingUseLargeWoundLookup:
		return "UseLargeWoundLookup"
	case SettingUseNormalMap:
		return "UseNormalMap"
	case SettingUseNormalMapAlpha:
		return "UseNormalMapAlpha"
	case SettingUseObjectCameraTransform:
		return "UseObjectCameraTransform"
	case SettingUseParallax:
		return "UseParallax"
	case SettingUseParticleAlpha:
		return "UseParticleAlpha"
	case SettingUseSss:
		return "UseSss"
	case SettingUseTrample:
		return "UseTrample"
	case SettingUseUpVector:
		return "UseUpVector"
	case SettingUvDistortion:
		return "UvDistortion"
	case SettingUvExp:
		return "UvExp"
	case SettingUvMappingTechniques:
		return "UvMappingTechniques"
	case SettingUvOffset:
		return "UvOffset"
	case SettingUvRect:
		return "UvRect"
	case SettingUvRotate01:
		return "UvRotate01"
	case SettingUvRotate02:
		return "UvRotate02"
	case SettingUvScale:
		return "UvScale"
	case SettingUvScalingMult:
		return "UvScalingMult"
	case SettingUvScalingRatio:
		return "UvScalingRatio"
	case SettingVertexDeformationFlags:
		return "VertexDeformationFlags"
	case SettingView:
		return "View"
	case SettingViewProj:
		return "ViewProj"
	case SettingVolumetricCloudsColor:
		return "VolumetricCloudsColor"
	case SettingVolumetricCloudsColorProbe:
		return "VolumetricCloudsColorProbe"
	case SettingVolumetricCloudsShadowsFinal:
		return "VolumetricCloudsShadowsFinal"
	case SettingVolumetricFog3dImage:
		return "VolumetricFog3dImage"
	case SettingVpMaxSlice0:
		return "VpMaxSlice0"
	case SettingVpMaxSlice1:
		return "VpMaxSlice1"
	case SettingVpMaxSlice2:
		return "VpMaxSlice2"
	case SettingVpMaxSlice3:
		return "VpMaxSlice3"
	case SettingVpMinSlice0:
		return "VpMinSlice0"
	case SettingVpMinSlice1:
		return "VpMinSlice1"
	case SettingVpMinSlice2:
		return "VpMinSlice2"
	case SettingVpMinSlice3:
		return "VpMinSlice3"
	case SettingVpRenderResolution:
		return "VpRenderResolution"
	case SettingWaterColor:
		return "WaterColor"
	case SettingWaterContrast:
		return "WaterContrast"
	case SettingWaterOpacity:
		return "WaterOpacity"
	case SettingWaterRotate:
		return "WaterRotate"
	case SettingWeatheringAmount:
		return "WeatheringAmount"
	case SettingWeatheringAoAmount:
		return "WeatheringAoAmount"
	case SettingWeatheringAoCutoff:
		return "WeatheringAoCutoff"
	case SettingWeatheringCoverage:
		return "WeatheringCoverage"
	case SettingWeatheringCoverageMultiplier:
		return "WeatheringCoverageMultiplier"
	case SettingWeatheringCurve:
		return "WeatheringCurve"
	case SettingWeatheringDirtWeightsNegative:
		return "WeatheringDirtWeightsNegative"
	case SettingWeatheringDirtWeightsPositive:
		return "WeatheringDirtWeightsPositive"
	case SettingWeatheringDynamicTiling:
		return "WeatheringDynamicTiling"
	case SettingWeatheringHeight:
		return "WeatheringHeight"
	case SettingWeatheringHeightCurve:
		return "WeatheringHeightCurve"
	case SettingWeatheringMasking:
		return "WeatheringMasking"
	case SettingWeatheringNormalIntensity:
		return "WeatheringNormalIntensity"
	case SettingWeatheringNormalMult:
		return "WeatheringNormalMult"
	case SettingWeatheringOnDirtAmount:
		return "WeatheringOnDirtAmount"
	case SettingWeatheringSpecialCase:
		return "WeatheringSpecialCase"
	case SettingWeatheringSpecialWeightsNegative:
		return "WeatheringSpecialWeightsNegative"
	case SettingWeatheringSpecialWeightsPositive:
		return "WeatheringSpecialWeightsPositive"
	case SettingWeatheringThickness:
		return "WeatheringThickness"
	case SettingWeatheringTileFactor:
		return "WeatheringTileFactor"
	case SettingWeatheringTiling:
		return "WeatheringTiling"
	case SettingWeatheringUpFadeInfluence:
		return "WeatheringUpFadeInfluence"
	case SettingWeatheringVariant:
		return "WeatheringVariant"
	case SettingWeatheringWrap:
		return "WeatheringWrap"
	case SettingWetness:
		return "Wetness"
	case SettingWetnessBottom:
		return "WetnessBottom"
	case SettingWetnessPuddles:
		return "WetnessPuddles"
	case SettingWorld:
		return "World"
	case SettingWorldMaskSize:
		return "WorldMaskSize"
	case SettingWoundPaintingEnabled:
		return "WoundPaintingEnabled"
	case SettingWoundParallaxScale:
		return "WoundParallaxScale"
	case SettingWoundTileScale:
		return "WoundTileScale"
	case SettingWoundableId:
		return "WoundableId"
	case SettingZeroingDist:
		return "ZeroingDist"
	case SettingZoneWetness:
		return "ZoneWetness"
	case SettingCameraFadeDistance01:
		return "CameraFadeDistance01"
	case SettingColorDamaged:
		return "ColorDamaged"
	case SettingColorDamagedMult:
		return "ColorDamagedMult"
	case SettingColorOuter:
		return "ColorOuter"
	case SettingDistFadeOffset01:
		return "DistFadeOffset01"
	case SettingDotsMult:
		return "DotsMult"
	case SettingEnergyLinesSpeed:
		return "EnergyLinesSpeed"
	case SettingFarFadeDistance01:
		return "FarFadeDistance01"
	case SettingFarFadeOffset01:
		return "FarFadeOffset01"
	case SettingFarFadeOpacity01:
		return "FarFadeOpacity01"
	case SettingFresnelPower:
		return "FresnelPower"
	case SettingHitExp01:
		return "HitExp01"
	case SettingHitIntensity:
		return "HitIntensity"
	case SettingIlluminateDotsTile:
		return "IlluminateDotsTile"
	case SettingInteresectDistance:
		return "InteresectDistance"
	case SettingInteresectDistanceMult:
		return "InteresectDistanceMult"
	case SettingInteresectExp:
		return "InteresectExp"
	case SettingLowHealth:
		return "LowHealth"
	case SettingLumMin02:
		return "LumMin02"
	case SettingLuminocityExp02:
		return "LuminocityExp02"
	case SettingMaxOpac:
		return "MaxOpac"
	case SettingMinOpac:
		return "MinOpac"
	case SettingNoiseFresnelExp:
		return "NoiseFresnelExp"
	case SettingNoiseFresnelMin:
		return "NoiseFresnelMin"
	case SettingOpacityMult01:
		return "OpacityMult01"
	case SettingPower01:
		return "Power01"
	case SettingSpeedLinesMult:
		return "SpeedLinesMult"
	case SettingSpeedLinesTile:
		return "SpeedLinesTile"
	case SettingTopGlowMult:
		return "TopGlowMult"
	case SettingTopGlowTighten:
		return "TopGlowTighten"
	case SettingUseVetexColor:
		return "UseVetexColor"
	case SettingUseHitData:
		return "UseHitData"
	case SettingWpoDistance01:
		return "WpoDistance01"
	default:
		return "Unknown setting usage!"
	}
}
