package prefab

import (
	"fmt"
	"slices"
	"strings"

	"github.com/qmuntal/gltf"
	"github.com/xypwn/filediver/extractor"
	"github.com/xypwn/filediver/extractor/blend_helper"
	extr_material "github.com/xypwn/filediver/extractor/material"
	extr_unit "github.com/xypwn/filediver/extractor/unit"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/prefab"
)

func cloneUnitNode(doc *gltf.Document, node uint32) (uint32, error) {
	//fmt.Println(doc.Nodes[node])
	nodeMetadata, ok := doc.Nodes[node].Extras.(map[string]any)
	if !ok {
		return 0, fmt.Errorf("failed to load node extras when cloning unit node")
	}
	nodeHashIface, contains := nodeMetadata["hash"]
	if !contains {
		return 0, fmt.Errorf("failed to find node hash when cloning unit node")
	}
	nodeHash, ok := nodeHashIface.(string)
	if !ok {
		return 0, fmt.Errorf("failed to convert node hash to string when cloning unit node")
	}
	extras, ok := doc.Extras.(map[string]any)
	if !ok {
		return 0, fmt.Errorf("failed to convert doc extras when cloning unit node")
	}
	unitMetadataIface, contains := extras[nodeHash]
	if !contains {
		return 0, fmt.Errorf("failed to find unit metadata when cloning unit node")
	}
	unitMetadata, ok := unitMetadataIface.(map[string]any)
	if !ok {
		return 0, fmt.Errorf("failed to convert unit metadata when cloning unit node")
	}

	skinIface, containsSkin := unitMetadata["skin"]
	var parent uint32
	if !containsSkin {
		// No skin, just need to copy parent and its children
		newParent := uint32(len(doc.Nodes))
		doc.Scenes[0].Nodes = append(doc.Scenes[0].Nodes, newParent)
		doc.Nodes = append(doc.Nodes, &gltf.Node{
			Name:        doc.Nodes[node].Name,
			Translation: doc.Nodes[node].Translation,
			Rotation:    doc.Nodes[node].Rotation,
			Scale:       doc.Nodes[node].Scale,
			Extras:      nodeMetadata,
		})
		for _, index := range doc.Nodes[node].Children {
			doc.Nodes[newParent].Children = append(doc.Nodes[newParent].Children, uint32(len(doc.Nodes)))
			doc.Nodes = append(doc.Nodes, &gltf.Node{
				Name: doc.Nodes[index].Name,
				Mesh: doc.Nodes[index].Mesh,
			})
		}
		parent = newParent
	} else {
		// Have a skin, so we need to copy the skin, the objects that use it, and the skeleton node
		skin, ok := skinIface.(uint32)
		if !ok {
			return 0, fmt.Errorf("%s.unit skin cast failed? (this should not happen)", nodeHash)
		}
		objectsIface, contains := unitMetadata["objects"]
		if !contains {
			return 0, fmt.Errorf("%s.unit did not have objects set? (this should not happen)", nodeHash)
		}
		objects, ok := objectsIface.([]uint32)
		if !ok {
			return 0, fmt.Errorf("%s.unit objects cast failed? (this should not happen)", nodeHash)
		}
		newSkin := uint32(len(doc.Skins))
		doc.Skins = append(doc.Skins, &gltf.Skin{
			Name:                doc.Skins[skin].Name,
			InverseBindMatrices: doc.Skins[skin].InverseBindMatrices,
		})
		root := doc.Skins[skin].Joints[0]
		jointNames := make([]string, 0)
		for _, joint := range doc.Skins[skin].Joints {
			jointNames = append(jointNames, doc.Nodes[joint].Name)
		}
		var cloneSkeleton func(curr uint32, parentPtr *uint32)
		cloneSkeleton = func(curr uint32, parentPtr *uint32) {
			if !slices.Contains(jointNames, doc.Nodes[curr].Name) {
				return
			}
			newCurr := uint32(len(doc.Nodes))
			if len(doc.Nodes[curr].Extensions) == 0 && doc.Nodes[curr].Mesh == nil {
				doc.Skins[newSkin].Joints = append(doc.Skins[newSkin].Joints, newCurr)
			}
			doc.Nodes = append(doc.Nodes, &gltf.Node{
				Name:        doc.Nodes[curr].Name,
				Rotation:    doc.Nodes[curr].Rotation,
				Translation: doc.Nodes[curr].Translation,
				Scale:       doc.Nodes[curr].Scale,
				Extras:      doc.Nodes[curr].Extras,
				Extensions:  doc.Nodes[curr].Extensions,
			})
			if parentPtr != nil {
				doc.Nodes[*parentPtr].Children = append(doc.Nodes[*parentPtr].Children, newCurr)
			}
			for _, child := range doc.Nodes[curr].Children {
				cloneSkeleton(child, &newCurr)
			}
		}
		cloneSkeleton(root, nil)

		newParent := uint32(len(doc.Nodes))
		doc.Scenes[0].Nodes = append(doc.Scenes[0].Nodes, newParent)
		doc.Skins[newSkin].Skeleton = gltf.Index(newParent)
		doc.Nodes = append(doc.Nodes, &gltf.Node{
			Name:        doc.Nodes[node].Name,
			Translation: doc.Nodes[node].Translation,
			Rotation:    doc.Nodes[node].Rotation,
			Scale:       doc.Nodes[node].Scale,
			Extras:      nodeMetadata,
			Children:    []uint32{doc.Skins[newSkin].Joints[0]},
		})
		for _, node := range objects {
			doc.Scenes[0].Nodes = append(doc.Scenes[0].Nodes, uint32(len(doc.Nodes)))
			doc.Nodes = append(doc.Nodes, &gltf.Node{
				Name: doc.Nodes[node].Name,
				Skin: gltf.Index(newSkin),
				Mesh: doc.Nodes[node].Mesh,
			})
		}
		parent = newParent
	}
	return parent, nil
}

func duplicateExistingUnit(ctx *extractor.Context, doc *gltf.Document) (uint32, error) {
	extras, ok := doc.Extras.(map[string]any)
	if !ok {
		return 0, fmt.Errorf("no doc extras when trying to duplicate existing unit?")
	}
	unitMetadataIface, contains := extras[extr_unit.GetUnitExtrasID(ctx.FileID())]
	if !contains {
		return 0, fmt.Errorf("no unit metadata when trying to duplicate existing unit?")
	}
	unitMetadata, ok := unitMetadataIface.(map[string]any)
	if !ok {
		return 0, fmt.Errorf("could not convert unitMetadata when trying to duplicate existing unit?")
	}

	// We have loaded this unit, just copy relevant nodes
	rootIface, contains := unitMetadata["root"]
	if !contains {
		return 0, fmt.Errorf("%s.unit did not have a parent set? (this should not happen)", ctx.FileID().Name.String())
	}

	existingRoot, ok := rootIface.(uint32)
	if !ok {
		return 0, fmt.Errorf("%s.unit parent cast failed? (this should not happen)", ctx.FileID().Name.String())
	}

	return cloneUnitNode(doc, existingRoot)
}

func AddOrDuplicateUnit(ctx *extractor.Context, doc *gltf.Document, imgOpts *extr_material.ImageOptions) (uint32, error) {
	extras, ok := doc.Extras.(map[string]any)
	if !ok {
		extras = make(map[string]any)
	}

	var root uint32
	if _, contains := extras[extr_unit.GetUnitExtrasID(ctx.FileID())]; !contains {
		ctx.Warnf("Adding unit %v", ctx.LookupHash(ctx.FileID().Name))
		// We have not already loaded this unit, load it now
		err := extr_unit.ConvertOpts(ctx, imgOpts, doc)
		if err != nil {
			return 0, err
		}

		extras, ok := doc.Extras.(map[string]any)
		if !ok {
			return 0, fmt.Errorf("could not resolve %s.unit gltf metadata? (this should not happen)", ctx.FileID().Name.String())
		}
		unitMetadataIface, ok := extras[extr_unit.GetUnitExtrasID(ctx.FileID())]
		if !ok {
			return 0, fmt.Errorf("could not resolve %s.unit gltf metadata? (this should not happen)", ctx.FileID().Name.String())
		}
		unitMetadata, ok := unitMetadataIface.(map[string]any)
		if !ok {
			return 0, fmt.Errorf("could not cast %s.unit gltf metadata? (this should not happen)", ctx.FileID().Name.String())
		}
		root, ok = unitMetadata["root"].(uint32)
		if !ok {
			return 0, fmt.Errorf("%s.unit did not have a root set? (this should not happen)", ctx.FileID().Name.String())
		}
		doc.Extras = extras
	} else {
		ctx.Warnf("Duplicating unit %v", ctx.LookupHash(ctx.FileID().Name))
		var err error
		root, err = duplicateExistingUnit(ctx, doc)
		if err != nil {
			return 0, err
		}
	}
	return root, nil
}

func addNewPrefab(ctx *extractor.Context, doc *gltf.Document, prefabData *prefab.Prefab, imgOpts *extr_material.ImageOptions) (uint32, error) {
	extras, ok := doc.Extras.(map[string]any)
	if !ok {
		extras = make(map[string]any)
	}

	prefabRoot := uint32(len(doc.Nodes))
	doc.Scenes[0].Nodes = append(doc.Scenes[0].Nodes, prefabRoot)
	doc.Nodes = append(doc.Nodes, &gltf.Node{
		Name:     ctx.LookupHash(ctx.FileID().Name) + ".prefab",
		Children: make([]uint32, 0),
		Extras: map[string]any{
			"hash": GetPrefabExtrasID(ctx.FileID()),
		},
	})

	extras[GetPrefabExtrasID(ctx.FileID())] = map[string]any{
		"node":   prefabRoot,
		"parent": nil,
	}
	doc.Extras = extras

	for _, object := range prefabData.Units {
		unitId := stingray.NewFileID(object.Unit(), stingray.Sum("unit"))
		root, err := AddOrDuplicateUnit(ctx.WithFileID(unitId), doc, imgOpts)
		if err != nil {
			return 0, err
		}
		doc.Nodes[prefabRoot].Children = append(doc.Nodes[prefabRoot].Children, root)
		position, rotation, scale := object.ToGLTF()
		doc.Nodes[root].Translation = position
		doc.Nodes[root].Rotation = rotation
		doc.Nodes[root].Scale = scale

		extras, ok := doc.Extras.(map[string]any)
		if !ok {
			return 0, fmt.Errorf("could not resolve %s.unit gltf metadata? (this should not happen)", unitId.Name.String())
		}
		unitMetadataIface, ok := extras[extr_unit.GetUnitExtrasID(unitId)]
		if !ok {
			return 0, fmt.Errorf("could not resolve %s.unit gltf metadata? (this should not happen)", unitId.Name.String())
		}
		unitMetadata, ok := unitMetadataIface.(map[string]any)
		if !ok {
			return 0, fmt.Errorf("could not cast %s.unit gltf metadata? (this should not happen)", unitId.Name.String())
		}
		unitMetadata["parent"] = prefabRoot
		extras[extr_unit.GetUnitExtrasID(unitId)] = unitMetadata
		doc.Extras = extras
	}

	for _, nested := range prefabData.NestedPrefabs {
		nestedId := stingray.NewFileID(nested.Path, stingray.Sum("prefab"))
		root, err := AddPrefab(ctx.WithFileID(nestedId), doc, imgOpts)
		if err != nil {
			return 0, fmt.Errorf("extracting prefab %v: %v", ctx.LookupHash(nested.Path), err)
		}
		doc.Nodes[prefabRoot].Children = append(doc.Nodes[prefabRoot].Children, root)
		translation, rotation, scale := nested.ToGLTF()
		doc.Nodes[root].Translation = translation
		doc.Nodes[root].Rotation = rotation
		doc.Nodes[root].Scale = scale

		extras, ok := doc.Extras.(map[string]any)
		if !ok {
			return 0, fmt.Errorf("could not resolve %s.prefab gltf metadata? (this should not happen)", nestedId.Name.String())
		}
		prefabMetadataIface, ok := extras[GetPrefabExtrasID(nestedId)]
		if !ok {
			return 0, fmt.Errorf("could not resolve %s.prefab gltf metadata? (this should not happen)", nestedId.Name.String())
		}
		prefabMetadata, ok := prefabMetadataIface.(map[string]any)
		if !ok {
			return 0, fmt.Errorf("could not cast %s.prefab gltf metadata? (this should not happen)", nestedId.Name.String())
		}
		prefabMetadata["parent"] = prefabRoot
		extras[GetPrefabExtrasID(nestedId)] = prefabMetadata
		doc.Extras = extras
	}

	return prefabRoot, nil
}

func cloneExistingPrefab(ctx *extractor.Context, doc *gltf.Document, prefabMap map[string]any) (uint32, error) {
	existingPrefabNodeIdx, ok := prefabMap["node"].(uint32)
	if !ok {
		return 0, fmt.Errorf("failed to convert prefab node to uint32 (should not happen)")
	}
	existingPrefabNode := doc.Nodes[existingPrefabNodeIdx]
	prefabNode := uint32(len(doc.Nodes))
	doc.Nodes = append(doc.Nodes, &gltf.Node{
		Name:     ctx.LookupHash(ctx.FileID().Name) + ".prefab",
		Children: make([]uint32, 0),
		Extras:   existingPrefabNode.Extras,
	})

	for _, childIdx := range existingPrefabNode.Children {
		child := doc.Nodes[childIdx]
		var newChild uint32
		var err error
		if strings.HasSuffix(child.Name, ".prefab") {
			childPrefabMap := map[string]any{
				"node":   childIdx,
				"parent": existingPrefabNodeIdx,
			}
			newChild, err = cloneExistingPrefab(ctx, doc, childPrefabMap)
		} else {
			newChild, err = cloneUnitNode(doc, childIdx)
			doc.Nodes[newChild].Translation = child.Translation
			doc.Nodes[newChild].Rotation = child.Rotation
			doc.Nodes[newChild].Scale = child.Scale
		}
		if err != nil {
			return 0, err
		}
		doc.Nodes[prefabNode].Children = append(doc.Nodes[prefabNode].Children, newChild)
	}

	return prefabNode, nil
}

func GetPrefabExtrasID(fileId stingray.FileID) string {
	return fileId.Name.String() + ".prefab"
}

func AddPrefab(ctx *extractor.Context, doc *gltf.Document, imgOpts *extr_material.ImageOptions) (uint32, error) {
	fMain, err := ctx.Open(ctx.FileID(), stingray.DataMain)
	if err != nil {
		return 0, err
	}
	prefabData, err := prefab.Load(fMain)
	if err != nil {
		return 0, err
	}

	extras, ok := doc.Extras.(map[string]any)
	if !ok {
		extras = make(map[string]any)
	}
	if prefabIface, contains := extras[GetPrefabExtrasID(ctx.FileID())]; contains {
		prefabMap, ok := prefabIface.(map[string]any)
		if !ok {
			return 0, fmt.Errorf("failed to convert prefab interface to map (should not happen)")
		}
		return cloneExistingPrefab(ctx, doc, prefabMap)
	}

	return addNewPrefab(ctx, doc, prefabData, imgOpts)
}

func ConvertOpts(ctx *extractor.Context, gltfDoc *gltf.Document) error {
	imgOpts, err := extr_material.GetImageOpts(ctx)
	if err != nil {
		return err
	}

	var doc *gltf.Document = gltfDoc
	if doc == nil {
		doc = gltf.NewDocument()
		doc.Asset.Generator = "https://github.com/xypwn/filediver"
		if ctx.BuildInfo() != nil {
			doc.Scenes[0].Extras = map[string]any{"Helldivers 2 Version": ctx.BuildInfo().Version}
		}
		doc.Samplers = append(doc.Samplers, &gltf.Sampler{
			MagFilter: gltf.MagLinear,
			MinFilter: gltf.MinLinear,
			WrapS:     gltf.WrapRepeat,
			WrapT:     gltf.WrapRepeat,
		})
	}

	if _, err := AddPrefab(ctx, doc, imgOpts); err != nil {
		return err
	}

	extractor.ClearChildNodesFromScene(ctx, doc)

	cfg := ctx.Config()
	formatIsBlend := cfg.Model.Format == "blend"
	if gltfDoc == nil && !formatIsBlend {
		out, err := ctx.CreateFile(".prefab.glb")
		if err != nil {
			return err
		}
		enc := gltf.NewEncoder(out)
		if err := enc.Encode(doc); err != nil {
			return err
		}
	} else if gltfDoc == nil && formatIsBlend {
		outPath, err := ctx.AllocateFile(".prefab.blend")
		if err != nil {
			return err
		}
		err = blend_helper.ExportBlend(doc, outPath, ctx.Runner())
		if err != nil {
			return err
		}
	}
	return nil
}

func Convert(currDoc *gltf.Document) func(ctx *extractor.Context) error {
	return func(ctx *extractor.Context) error {
		return ConvertOpts(ctx, currDoc)
	}
}
