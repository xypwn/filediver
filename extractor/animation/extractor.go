package animation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/qmuntal/gltf"
	"github.com/qmuntal/gltf/modeler"
	"github.com/xypwn/filediver/extractor"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/animation"
	"github.com/xypwn/filediver/stingray/bones"
)

func ExtractAnimationJson(ctx *extractor.Context) error {
	r, err := ctx.Open(ctx.FileID(), stingray.DataMain)
	if err != nil {
		return err
	}

	anim, err := animation.LoadAnimation(r)
	if err != nil {
		return fmt.Errorf("extract animation json: loading animation failed: %v", err)
	}
	anim.Header.ResolvedHashes = make([]string, 0)
	for _, hash := range anim.Header.Hashes {
		anim.Header.ResolvedHashes = append(anim.Header.ResolvedHashes, ctx.LookupHash(hash))
	}

	anim.Header.ResolvedHashes2 = make([]string, 0)
	for _, hash := range anim.Header.Hashes2 {
		anim.Header.ResolvedHashes2 = append(anim.Header.ResolvedHashes2, ctx.LookupHash(hash))
	}

	text, err := json.Marshal(anim)
	if err != nil {
		return err
	}
	var txtBuf bytes.Buffer
	err = json.Indent(&txtBuf, text, "", "    ")
	if err != nil {
		return err
	}

	out, err := ctx.CreateFile(".animation.json")
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = out.Write(txtBuf.Bytes())
	return err
}

func getTargetNode(doc *gltf.Document, boneInfo *bones.Info, boneIdx uint32) (uint32, error) {
	for nodeIdx := range doc.Nodes {
		if doc.Nodes[nodeIdx].Name == boneInfo.NameMap[boneInfo.Hashes[boneIdx]] {
			return uint32(nodeIdx), nil
		}
	}
	return 0, fmt.Errorf("could not find bone %v in document", boneInfo.NameMap[boneInfo.Hashes[boneIdx]])
}

func AddAnimation(ctx *extractor.Context, doc *gltf.Document, boneInfo *bones.Info, path stingray.Hash) (uint32, error) {
	cfg := ctx.Config()

	mainR, err := ctx.Open(stingray.NewFileID(path, stingray.Sum("animation")), stingray.DataMain)
	if err == stingray.ErrFileNotExist {
		return 0, fmt.Errorf("could not find animation %v", path.String())
	}
	if err != nil {
		return 0, fmt.Errorf("could not open animation file %v: %v", path.String(), err)
	}

	animInfo, err := animation.LoadAnimation(mainR)
	if err != nil {
		return 0, fmt.Errorf("could not parse animation file %v: %v", path.String(), err)
	}

	bonePositions := make([]VectorCurve, animInfo.Header.BoneCount)
	boneRotations := make([]QuaternionCurve, animInfo.Header.BoneCount)
	boneScales := make([]VectorCurve, animInfo.Header.BoneCount)
	additive := make([]bool, animInfo.Header.BoneCount)

	for i, initialTransform := range animInfo.Header.InitialTransforms {
		bonePositions[i].Duration = animInfo.Header.AnimationLength
		bonePositions[i].Keyframes = make([]VectorKeyframe, 0)
		bonePositions[i].Keyframes = append(bonePositions[i].Keyframes, VectorKeyframe{
			Time:   0.0,
			Vector: initialTransform.Position(),
		})

		boneRotations[i].Duration = animInfo.Header.AnimationLength
		boneRotations[i].Keyframes = make([]QuaternionKeyframe, 0)
		boneRotations[i].Keyframes = append(boneRotations[i].Keyframes, QuaternionKeyframe{
			Time:       0.0,
			Quaternion: initialTransform.Rotation(),
		})

		boneScales[i].Duration = animInfo.Header.AnimationLength
		boneScales[i].Keyframes = make([]VectorKeyframe, 0)
		boneScales[i].Keyframes = append(boneScales[i].Keyframes, VectorKeyframe{
			Time:   0.0,
			Vector: initialTransform.Scale(),
		})

		additive[i] = initialTransform.IsAdditive()
	}

	for i, entry := range animInfo.Entries {
		if uint32(entry.Header.Bone()) >= animInfo.Header.BoneCount {
			return 0, fmt.Errorf("entry %v in animation %v had bone index %v exceeding bone count %v", i, path.String(), entry.Header.Bone(), animInfo.Header.BoneCount)
		}
		switch entry.Header.Type() {
		case animation.EntryTypePosition:
			value, err := entry.Position()
			if err != nil {
				return 0, fmt.Errorf("adding entry %v to animation %v: %v", i, path.String(), err)
			}
			bonePositions[entry.Header.Bone()].Keyframes = append(bonePositions[entry.Header.Bone()].Keyframes, VectorKeyframe{
				Time:   float32(entry.Header.TimeMS()) / 1000.0,
				Vector: value,
			})
		case animation.EntryTypeRotation:
			value, err := entry.Rotation()
			if err != nil {
				return 0, fmt.Errorf("adding entry %v to animation %v: %v", i, path.String(), err)
			}
			boneRotations[entry.Header.Bone()].Keyframes = append(boneRotations[entry.Header.Bone()].Keyframes, QuaternionKeyframe{
				Time:       float32(entry.Header.TimeMS()) / 1000.0,
				Quaternion: value,
			})
		case animation.EntryTypeScale:
			value, err := entry.Scale()
			if err != nil {
				return 0, fmt.Errorf("adding entry %v to animation %v: %v", i, path.String(), err)
			}
			boneScales[entry.Header.Bone()].Keyframes = append(boneScales[entry.Header.Bone()].Keyframes, VectorKeyframe{
				Time:   float32(entry.Header.TimeMS()) / 1000.0,
				Vector: value,
			})
		case animation.EntryTypeExtended:
			if entry.Header.Subtype() == animation.EntrySubtypePosition {
				value, err := entry.Position()
				if err != nil {
					return 0, fmt.Errorf("adding entry %v to animation %v: %v", i, path.String(), err)
				}
				bonePositions[entry.Header.Bone()].Keyframes = append(bonePositions[entry.Header.Bone()].Keyframes, VectorKeyframe{
					Time:   float32(entry.Header.TimeMS()) / 1000.0,
					Vector: value,
				})
			} else if entry.Header.Subtype() == animation.EntrySubtypeRotation {
				value, err := entry.Rotation()
				if err != nil {
					return 0, fmt.Errorf("adding entry %v to animation %v: %v", i, path.String(), err)
				}
				boneRotations[entry.Header.Bone()].Keyframes = append(boneRotations[entry.Header.Bone()].Keyframes, QuaternionKeyframe{
					Time:       float32(entry.Header.TimeMS()) / 1000.0,
					Quaternion: value,
				})
			} else if entry.Header.Subtype() == animation.EntrySubtypeScale {
				value, err := entry.Scale()
				if err != nil {
					return 0, fmt.Errorf("adding entry %v to animation %v: %v", i, path.String(), err)
				}
				boneScales[entry.Header.Bone()].Keyframes = append(boneScales[entry.Header.Bone()].Keyframes, VectorKeyframe{
					Time:   float32(entry.Header.TimeMS()) / 1000.0,
					Vector: value,
				})
			}
		default:
			return 0, fmt.Errorf("adding entry %v to animation %v: unimplemented entry type %v", i, path.String(), entry.Header.Type().String())
		}
	}

	if cfg.Unit.SampleAnimations {
		extras, ok := doc.Extras.(map[string]any)
		if !ok {
			extras = make(map[string]any)
		}
		extras["frameRate"] = cfg.Unit.AnimationSampleRate
		doc.Extras = extras
	}

	samplers := make([]*gltf.AnimationSampler, 0)
	channels := make([]*gltf.Channel, 0)
	gltfConvertQuat := mgl32.QuatRotate(mgl32.DegToRad(-90), mgl32.Vec3([3]float32{1, 0, 0}))
	for boneIdx := uint32(0); boneIdx < animInfo.Header.BoneCount; boneIdx += 1 {
		targetNode, err := getTargetNode(doc, boneInfo, boneIdx)
		if err != nil {
			ctx.Warnf("writing gltf animation %v: %v", path.String(), err)
			continue
		}

		var posKeyframes, scaleKeyframes []VectorKeyframe
		var rotKeyframes []QuaternionKeyframe
		if cfg.Unit.SampleAnimations {
			posKeyframes = bonePositions[boneIdx].Sample(cfg.Unit.AnimationSampleRate)
			rotKeyframes = boneRotations[boneIdx].Sample(cfg.Unit.AnimationSampleRate)
			scaleKeyframes = boneScales[boneIdx].Sample(cfg.Unit.AnimationSampleRate)
		} else {
			posKeyframes = bonePositions[boneIdx].Keyframes
			rotKeyframes = boneRotations[boneIdx].Keyframes
			scaleKeyframes = boneScales[boneIdx].Keyframes
		}

		times := make([]float32, 0)
		positions := make([][3]float32, 0)
		for _, position := range posKeyframes {
			times = append(times, position.Time)
			translation := position.Vector
			if additive[boneIdx] {
				// This doesn't *really* work unfortunately - diver/the .cast blender plugin modifies the
				// basis matrix of the bone rather than modifying the translation, but I don't know if that's
				// feasible with GLTF - maybe theres an extension for additive animations?
				translation = translation.Add(doc.Nodes[targetNode].Translation)
			} else if doc.Nodes[targetNode].Name == "StingrayEntityRoot" {
				translation = gltfConvertQuat.Rotate(translation)
			}
			positions = append(positions, translation)
		}

		positionTimesAccessor := modeler.WriteAccessor(doc, gltf.TargetNone, times)
		doc.Accessors[positionTimesAccessor].Min = []float32{0.0}
		doc.Accessors[positionTimesAccessor].Max = []float32{animInfo.Header.AnimationLength}
		positionsAccessor := modeler.WriteAccessor(doc, gltf.TargetNone, positions)

		positionSampler := &gltf.AnimationSampler{
			Input:  positionTimesAccessor,
			Output: positionsAccessor,
		}
		positionChannel := &gltf.Channel{
			Sampler: gltf.Index(uint32(len(samplers))),
			Target: gltf.ChannelTarget{
				Node: gltf.Index(targetNode),
				Path: gltf.TRSTranslation,
			},
		}
		samplers = append(samplers, positionSampler)
		channels = append(channels, positionChannel)

		times = make([]float32, 0)
		rotations := make([][4]float32, 0)
		for _, rotation := range rotKeyframes {
			times = append(times, rotation.Time)
			data := rotation.Quaternion.V.Vec4(rotation.Quaternion.W)
			if additive[boneIdx] {
				// Same comment as for translation above - this doesn't seem to quite work, though the animation
				// at least looks sensible rather than just a pile of body parts writhing around, so I'll take
				// the wins where I can get them lol
				vec := mgl32.Vec4(doc.Nodes[targetNode].Rotation)
				addRot := vec.Quat().Mul(rotation.Quaternion)
				data = addRot.V.Vec4(addRot.W)
			} else if doc.Nodes[targetNode].Name == "StingrayEntityRoot" {
				gltfConvertedRot := gltfConvertQuat.Mul(rotation.Quaternion)
				data = gltfConvertedRot.V.Vec4(gltfConvertedRot.W)
			}
			rotations = append(rotations, data)
		}

		rotationTimesAccessor := modeler.WriteAccessor(doc, gltf.TargetNone, times)
		doc.Accessors[rotationTimesAccessor].Min = []float32{0.0}
		doc.Accessors[rotationTimesAccessor].Max = []float32{animInfo.Header.AnimationLength}
		rotationsAccessor := modeler.WriteAccessor(doc, gltf.TargetNone, rotations)

		rotationSampler := &gltf.AnimationSampler{
			Input:  rotationTimesAccessor,
			Output: rotationsAccessor,
		}
		rotationChannel := &gltf.Channel{
			Sampler: gltf.Index(uint32(len(samplers))),
			Target: gltf.ChannelTarget{
				Path: gltf.TRSRotation,
				Node: gltf.Index(targetNode),
			},
		}
		samplers = append(samplers, rotationSampler)
		channels = append(channels, rotationChannel)

		times = make([]float32, 0)
		scales := make([][3]float32, 0)
		for _, scale := range scaleKeyframes {
			times = append(times, scale.Time)
			scales = append(scales, scale.Vector)
		}

		scaleTimesAccessor := modeler.WriteAccessor(doc, gltf.TargetNone, times)
		doc.Accessors[scaleTimesAccessor].Min = []float32{0.0}
		doc.Accessors[scaleTimesAccessor].Max = []float32{animInfo.Header.AnimationLength}
		scalesAccessor := modeler.WriteAccessor(doc, gltf.TargetNone, scales)

		scaleSampler := &gltf.AnimationSampler{
			Input:  scaleTimesAccessor,
			Output: scalesAccessor,
		}
		scaleChannel := &gltf.Channel{
			Sampler: gltf.Index(uint32(len(samplers))),
			Target: gltf.ChannelTarget{
				Path: gltf.TRSScale,
				Node: gltf.Index(targetNode),
			},
		}
		samplers = append(samplers, scaleSampler)
		channels = append(channels, scaleChannel)
	}
	animationName := ctx.LookupHash(path)
	if strings.Contains(animationName, "/") {
		pathList := make([]string, 0)
		pathList = append(pathList, filepath.Base(animationName))
		curr := filepath.Dir(animationName)
		for !strings.Contains(filepath.Base(curr), "animation") && len(curr) > len("animation") {
			pathList = append(pathList, filepath.Base(curr))
			curr = filepath.Dir(curr)
		}
		animationParent := filepath.Dir(curr)
		pathList = append(pathList, filepath.Base(curr))
		pathList = append(pathList, filepath.Base(animationParent))
		slices.Reverse(pathList)
		animationName = strings.Join(pathList, "/")
	}
	animationIdx := uint32(len(doc.Animations))
	doc.Animations = append(doc.Animations, &gltf.Animation{
		Name:     animationName,
		Samplers: samplers,
		Channels: channels,
	})
	return animationIdx, nil
}
