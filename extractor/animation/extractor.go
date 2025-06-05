package animation

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/qmuntal/gltf"
	"github.com/qmuntal/gltf/modeler"
	"github.com/xypwn/filediver/extractor"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/animation"
	"github.com/xypwn/filediver/stingray/bones"
	"github.com/xypwn/filediver/stingray/state_machine"
)

func ExtractAnimationJson(ctx extractor.Context) error {
	if !ctx.File().Exists(stingray.DataMain) {
		return errors.New("no main data")
	}
	r, err := ctx.File().Open(ctx.Ctx(), stingray.DataMain)
	if err != nil {
		return err
	}
	defer r.Close()

	anim, err := animation.LoadAnimation(r)
	if err != nil {
		return fmt.Errorf("extract animation json: loading animation failed: %v", err)
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

type positionKeyframe struct {
	Time     float32
	Position mgl32.Vec3
}

type rotationKeyframe struct {
	Time     float32
	Rotation mgl32.Quat
}

type scaleKeyframe struct {
	Time  float32
	Scale mgl32.Vec3
}

func AddAnimation(ctx extractor.Context, doc *gltf.Document, boneInfo *bones.Info, anim state_machine.Animation) error {
	//boneIndexMap := make(map[stingray.ThinHash]uint32)
	var makeMapRecursive func(*gltf.Document, uint32, map[stingray.ThinHash]uint32)
	makeMapRecursive = func(doc *gltf.Document, nodeIdx uint32, boneMap map[stingray.ThinHash]uint32) {
		boneMap[stingray.Sum64([]byte(doc.Nodes[nodeIdx].Name)).Thin()] = nodeIdx
		for childIdx := range doc.Nodes[nodeIdx].Children {
			makeMapRecursive(doc, uint32(childIdx), boneMap)
		}
	}

	for _, path := range anim.AnimationHashes {
		animationFile, ok := ctx.GetResource(path, stingray.Sum64([]byte("animation")))
		if !ok {
			return fmt.Errorf("could not find animation %v", path.String())
		}
		mainR, err := animationFile.Open(ctx.Ctx(), stingray.DataMain)
		if err != nil {
			return fmt.Errorf("could not open animation file %v: %v", path.String(), err)
		}

		animInfo, err := animation.LoadAnimation(mainR)
		if err != nil {
			return fmt.Errorf("could not parse animation file %v: %v", path.String(), err)
		}

		bonePositions := make([][]positionKeyframe, animInfo.Header.BoneCount)
		boneRotations := make([][]rotationKeyframe, animInfo.Header.BoneCount)
		boneScales := make([][]scaleKeyframe, animInfo.Header.BoneCount)
		additive := make([]bool, animInfo.Header.BoneCount)

		for i, initialTransform := range animInfo.Header.InitialTransforms {
			bonePositions[i] = make([]positionKeyframe, 0)
			bonePositions[i] = append(bonePositions[i], positionKeyframe{
				Time:     0.0,
				Position: initialTransform.Position(),
			})

			boneRotations[i] = make([]rotationKeyframe, 0)
			boneRotations[i] = append(boneRotations[i], rotationKeyframe{
				Time:     0.0,
				Rotation: initialTransform.Rotation(),
			})

			boneScales[i] = make([]scaleKeyframe, 0)
			boneScales[i] = append(boneScales[i], scaleKeyframe{
				Time:  0.0,
				Scale: initialTransform.Scale(),
			})

			additive[i] = initialTransform.IsAdditive()
		}

		for i, entry := range animInfo.Entries {
			if uint32(entry.Header.Bone()) >= animInfo.Header.BoneCount {
				return fmt.Errorf("entry %v in animation %v had bone index %v exceeding bone count %v", i, path.String(), entry.Header.Bone(), animInfo.Header.BoneCount)
			}
			switch entry.Header.Type() {
			case animation.EntryTypePosition:
				value, err := entry.Position()
				if err != nil {
					return fmt.Errorf("adding entry %v to animation %v: %v", i, path.String(), err)
				}
				bonePositions[entry.Header.Bone()] = append(bonePositions[entry.Header.Bone()], positionKeyframe{
					Time:     float32(entry.Header.TimeMS()) / 1000.0,
					Position: value,
				})
			case animation.EntryTypeRotation:
				value, err := entry.Rotation()
				if err != nil {
					return fmt.Errorf("adding entry %v to animation %v: %v", i, path.String(), err)
				}
				boneRotations[entry.Header.Bone()] = append(boneRotations[entry.Header.Bone()], rotationKeyframe{
					Time:     float32(entry.Header.TimeMS()) / 1000.0,
					Rotation: value,
				})
			case animation.EntryTypeScale:
				value, err := entry.Scale()
				if err != nil {
					return fmt.Errorf("adding entry %v to animation %v: %v", i, path.String(), err)
				}
				boneScales[entry.Header.Bone()] = append(boneScales[entry.Header.Bone()], scaleKeyframe{
					Time:  float32(entry.Header.TimeMS()) / 1000.0,
					Scale: value,
				})
			case animation.EntryTypeExtended:
				if entry.Header.Subtype() == animation.EntrySubtypePosition {
					value, err := entry.Position()
					if err != nil {
						return fmt.Errorf("adding entry %v to animation %v: %v", i, path.String(), err)
					}
					bonePositions[entry.Header.Bone()] = append(bonePositions[entry.Header.Bone()], positionKeyframe{
						Time:     float32(entry.Header.TimeMS()) / 1000.0,
						Position: value,
					})
				} else if entry.Header.Subtype() == animation.EntrySubtypeRotation {
					value, err := entry.Rotation()
					if err != nil {
						return fmt.Errorf("adding entry %v to animation %v: %v", i, path.String(), err)
					}
					boneRotations[entry.Header.Bone()] = append(boneRotations[entry.Header.Bone()], rotationKeyframe{
						Time:     float32(entry.Header.TimeMS()) / 1000.0,
						Rotation: value,
					})
				} else if entry.Header.Subtype() == animation.EntrySubtypeScale {
					value, err := entry.Scale()
					if err != nil {
						return fmt.Errorf("adding entry %v to animation %v: %v", i, path.String(), err)
					}
					boneScales[entry.Header.Bone()] = append(boneScales[entry.Header.Bone()], scaleKeyframe{
						Time:  float32(entry.Header.TimeMS()) / 1000.0,
						Scale: value,
					})
				}
			default:
				return fmt.Errorf("adding entry %v to animation %v: unimplemented entry type %v", i, path.String(), entry.Header.Type().String())
			}
		}

		samplers := make([]*gltf.AnimationSampler, 0)
		channels := make([]*gltf.Channel, 0)
		for boneIdx := uint32(0); boneIdx < animInfo.Header.BoneCount; boneIdx += 1 {
			var targetNode uint32 = 0xffffffff
			for nodeIdx := range doc.Nodes {
				if doc.Nodes[nodeIdx].Name == boneInfo.NameMap[boneInfo.Hashes[boneIdx]] {
					targetNode = uint32(nodeIdx)
					break
				}
			}
			if targetNode == 0xffffffff {
				return fmt.Errorf("writing gltf animation %v: could not find bone %v in document", path.String(), boneInfo.NameMap[boneInfo.Hashes[boneIdx]])
			}

			times := make([]float32, 0)
			positions := make([][3]float32, 0)
			for _, position := range bonePositions[boneIdx] {
				times = append(times, position.Time)
				translation := position.Position
				if additive[boneIdx] {
					// This doesn't *really* work unfortunately - diver/the .cast blender plugin modifies the
					// basis matrix of the bone rather than modifying the translation, but I don't know if that's
					// feasible with GLTF - maybe theres an extension for additive animations?
					translation = translation.Add(doc.Nodes[targetNode].Translation)
				}
				positions = append(positions, translation)
			}

			// TODO: Hermite/catmull-rom curve interpolation, sampled at a consistent framerate
			// rather than just dumping the animation in as is (though tbh that pretty much works
			// thanks to how hermite curves pass through each control point)
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
			for _, rotation := range boneRotations[boneIdx] {
				times = append(times, rotation.Time)
				data := rotation.Rotation.V.Vec4(rotation.Rotation.W)
				if additive[boneIdx] {
					// Same comment as for translation above - this doesn't seem to quite work, though the animation
					// at least looks sensible rather than just a pile of body parts writhing around, so I'll take
					// the wins where I can get them lol
					vec := mgl32.Vec4(doc.Nodes[targetNode].Rotation)
					addRot := vec.Quat().Mul(rotation.Rotation)
					data = addRot.V.Vec4(addRot.W)
				}
				rotations = append(rotations, data)
			}

			// TODO: Hermite curve interpolation
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
			for _, scale := range boneScales[boneIdx] {
				times = append(times, scale.Time)
				scales = append(scales, scale.Scale)
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
		animationName := ctx.LookupHash(anim.Name)
		pathName := ctx.LookupHash(path)
		if strings.Contains(animationName, "/") {
			animationName = filepath.Base(animationName)
		}
		if strings.Contains(pathName, "/") {
			pathName = filepath.Base(pathName)
		}
		doc.Animations = append(doc.Animations, &gltf.Animation{
			Name:     fmt.Sprintf("%v %v", animationName, pathName),
			Samplers: samplers,
			Channels: channels,
		})
	}

	return nil
}
