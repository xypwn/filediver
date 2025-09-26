package prefab

import (
	"fmt"
	"io"

	"github.com/qmuntal/gltf"
	"github.com/xypwn/filediver/extractor"
	"github.com/xypwn/filediver/extractor/blend_helper"
	extr_material "github.com/xypwn/filediver/extractor/material"
	extr_unit "github.com/xypwn/filediver/extractor/unit"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/prefab"
)

func ConvertOpts(ctx *extractor.Context, gltfDoc *gltf.Document) error {
	fMain, err := ctx.Open(ctx.FileID(), stingray.DataMain)
	if err != nil {
		return err
	}
	prefabData, err := prefab.Load(fMain)
	if err != nil {
		return err
	}

	var doc *gltf.Document = gltfDoc
	if doc == nil {
		doc = gltf.NewDocument()
		doc.Asset.Generator = "https://github.com/xypwn/filediver"
		doc.Samplers = append(doc.Samplers, &gltf.Sampler{
			MagFilter: gltf.MagLinear,
			MinFilter: gltf.MinLinear,
			WrapS:     gltf.WrapRepeat,
			WrapT:     gltf.WrapRepeat,
		})
	}

	imgOpts, err := extr_material.GetImageOpts(ctx)
	if err != nil {
		return err
	}

	for _, object := range prefabData.Objects {
		unitID := stingray.NewFileID(object.UnitHash, stingray.Sum("unit"))
		extras, ok := doc.Extras.(map[string]any)
		if !ok {
			extras = make(map[string]any)
		}

		// Convert to glTF coords
		p := object.Position
		p[1], p[2] = p[2], -p[1]
		object.Position = p

		r := object.Rotation
		r[1], r[2] = r[2], -r[1]
		object.Rotation = r

		if unitMetadataIface, contains := extras[object.UnitHash.String()]; !contains {
			// We have not already loaded this unit, load it now
			fMain, err := ctx.Open(unitID, stingray.DataMain)
			if err != nil {
				return err
			}

			var fGPU io.ReadSeeker
			if ctx.Exists(unitID, stingray.DataGPU) {
				fGPU, err = ctx.Open(unitID, stingray.DataGPU)
				if err != nil {
					return err
				}
			}

			err = extr_unit.ConvertBuffer(fMain, fGPU, object.UnitHash, ctx, imgOpts, doc)
			if err != nil {
				return nil
			}

			extras, ok := doc.Extras.(map[string]any)
			if !ok {
				return fmt.Errorf("could not resolve %s.unit gltf metadata? (this should not happen)", object.UnitHash.String())
			}
			unitMetadataIface, ok := extras[object.UnitHash.String()]
			if !ok {
				return fmt.Errorf("could not resolve %s.unit gltf metadata? (this should not happen)", object.UnitHash.String())
			}
			unitMetadata, ok := unitMetadataIface.(map[string]any)
			if !ok {
				return fmt.Errorf("could not cast %s.unit gltf metadata? (this should not happen)", object.UnitHash.String())
			}
			parent, ok := unitMetadata["parent"].(uint32)
			if !ok {
				return fmt.Errorf("%s.unit did not have a parent set? (this should not happen)", object.UnitHash.String())
			}
			doc.Nodes[parent].Translation = object.Position
			doc.Nodes[parent].Rotation = object.Rotation
			doc.Nodes[parent].Scale = object.Scale
		} else {
			unitMetadata, ok := unitMetadataIface.(map[string]any)
			if !ok {
				return fmt.Errorf("could not cast %s.unit gltf metadata? (this should not happen)", object.UnitHash.String())
			}
			// We have loaded this unit, just copy relevant nodes
			parentIface, contains := unitMetadata["parent"]
			if !contains {
				return fmt.Errorf("%s.unit did not have a parent set? (this should not happen)", object.UnitHash.String())
			}

			parent, ok := parentIface.(uint32)
			if !ok {
				return fmt.Errorf("%s.unit parent cast failed? (this should not happen)", object.UnitHash.String())
			}
			skinIface, containsSkin := unitMetadata["skin"]
			if !containsSkin {
				// No skin, just need to copy parent and its children
				newParent := uint32(len(doc.Nodes))
				doc.Scenes[0].Nodes = append(doc.Scenes[0].Nodes, newParent)
				doc.Nodes = append(doc.Nodes, &gltf.Node{
					Name:        doc.Nodes[parent].Name,
					Translation: object.Position,
					Rotation:    object.Rotation,
					Scale:       object.Scale,
				})
				for _, index := range doc.Nodes[parent].Children {
					doc.Nodes[newParent].Children = append(doc.Nodes[newParent].Children, uint32(len(doc.Nodes)))
					doc.Nodes = append(doc.Nodes, &gltf.Node{
						Name: doc.Nodes[index].Name,
						Mesh: doc.Nodes[index].Mesh,
					})
				}
			} else {
				// Have a skin, so we need to copy the skin, the objects that use it, and the skeleton node
				skin, ok := skinIface.(uint32)
				if !ok {
					return fmt.Errorf("%s.unit skin cast failed? (this should not happen)", object.UnitHash.String())
				}
				objectsIface, contains := unitMetadata["objects"]
				if !contains {
					return fmt.Errorf("%s.unit did not have objects set? (this should not happen)", object.UnitHash.String())
				}
				objects, ok := objectsIface.([]uint32)
				if !ok {
					return fmt.Errorf("%s.unit objects cast failed? (this should not happen)", object.UnitHash.String())
				}
				newSkin := uint32(len(doc.Skins))
				doc.Skins = append(doc.Skins, &gltf.Skin{
					Name:                doc.Skins[skin].Name,
					InverseBindMatrices: doc.Skins[skin].InverseBindMatrices,
				})
				root := doc.Skins[skin].Joints[0]
				var cloneSkeleton func(curr uint32, parent *uint32)
				cloneSkeleton = func(curr uint32, parent *uint32) {
					newCurr := uint32(len(doc.Nodes))
					doc.Skins[newSkin].Joints = append(doc.Skins[newSkin].Joints, newCurr)
					doc.Nodes = append(doc.Nodes, &gltf.Node{
						Name:        doc.Nodes[curr].Name,
						Rotation:    doc.Nodes[curr].Rotation,
						Translation: doc.Nodes[curr].Translation,
						Scale:       doc.Nodes[curr].Scale,
						Extras:      doc.Nodes[curr].Extras,
					})
					if parent != nil {
						doc.Nodes[*parent].Children = append(doc.Nodes[*parent].Children, newCurr)
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
					Name:        doc.Nodes[parent].Name,
					Translation: object.Position,
					Rotation:    object.Rotation,
					Scale:       object.Scale,
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
			}
		}
	}

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
